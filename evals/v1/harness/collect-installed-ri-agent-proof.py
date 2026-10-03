#!/usr/bin/env python3
"""Read-only, metadata-only collector for journal-bound installed RI tool use."""
from __future__ import annotations
import argparse
import hashlib
import json
import os
import re
from pathlib import Path
import sqlite3
import sys
from urllib.parse import quote

ZERO = "0" * 64
MAX_SAFE = 9007199254740991

class EvidenceError(Exception):
    pass


def validated_run_id(run_id):
    if not isinstance(run_id, str) or re.fullmatch(r"[0-9a-f]{64}", run_id) is None:
        raise EvidenceError("run id must be exactly 64 lowercase hexadecimal characters")
    return run_id

def reject_float(_s):
    raise EvidenceError("floating-point JSON is not canonical")

def reject_constant(_s):
    raise EvidenceError("non-finite JSON is not canonical")

def object_pairs(pairs):
    out = {}
    for key, value in pairs:
        if not key.isascii() or key in out:
            raise EvidenceError("duplicate or non-ASCII JSON object key")
        out[key] = value
    return out

def parse(raw):
    if isinstance(raw, bytes):
        raw = raw.decode("utf-8", "strict")
    return json.loads(raw, parse_float=reject_float, parse_constant=reject_constant,
                      object_pairs_hook=object_pairs)

def canon(value):
    if value is None:
        return "null"
    if value is True:
        return "true"
    if value is False:
        return "false"
    if isinstance(value, int):
        if abs(value) > MAX_SAFE:
            raise EvidenceError("integer outside canonical range")
        return str(value)
    if isinstance(value, float):
        raise EvidenceError("floating-point JSON is not canonical")
    if isinstance(value, str):
        return json.dumps(value, ensure_ascii=False, separators=(",", ":"))
    if isinstance(value, list):
        return "[" + ",".join(canon(item) for item in value) + "]"
    if isinstance(value, dict):
        if any(not k.isascii() for k in value):
            raise EvidenceError("non-ASCII JSON object key")
        return "{" + ",".join(canon(k) + ":" + canon(value[k]) for k in sorted(value)) + "}"
    raise EvidenceError("unsupported canonical JSON value")

def domain_hash(domain, value):
    return hashlib.sha256((domain + "\n").encode("ascii") + canon(value).encode("utf-8")).hexdigest()

def file_sha256(path):
    digest = hashlib.sha256()
    with Path(path).open("rb") as stream:
        for chunk in iter(lambda: stream.read(1 << 20), b""):
            digest.update(chunk)
    return digest.hexdigest()

def open_events(path):
    path = Path(path).resolve(strict=True)
    uri = "file:" + quote(str(path).replace("\\", "/"), safe="/:_") + "?mode=ro"
    db = sqlite3.connect(uri, uri=True)
    db.execute("PRAGMA query_only=ON")
    try:
        meta = dict(db.execute("SELECT key, value FROM journal_metadata"))
        if meta.get("schema_version") != "1":
            raise EvidenceError("unsupported journal schema")
        rows = db.execute("SELECT sequence, version, previous, kind, payload, hash FROM events ORDER BY sequence").fetchall()
    except Exception:
        db.close()
        raise
    events = []
    previous = ZERO
    total = 0
    for index, row in enumerate(rows, 1):
        sequence, version, prev, kind, payload_raw, stored = row
        if sequence != index or version != 1 or prev != previous or not isinstance(kind, str) or not kind:
            db.close()
            raise EvidenceError("journal sequence or chain continuity invalid")
        raw_bytes = bytes(payload_raw) if not isinstance(payload_raw, str) else payload_raw.encode("utf-8")
        total += len(raw_bytes) + 1
        if len(raw_bytes) > (1 << 20) or total > (64 << 20):
            db.close()
            raise EvidenceError("journal payload or total size exceeds Fabric bounds")
        payload = parse(raw_bytes)
        if canon(payload).encode("utf-8") != raw_bytes:
            db.close()
            raise EvidenceError("SQLite payload bytes are not exact canonical JSON")
        if not isinstance(payload, dict):
            db.close()
            raise EvidenceError("journal payload is not an object")
        envelope = {"version": version, "sequence": sequence, "previous": prev,
                    "kind": kind, "payload": payload}
        actual = domain_hash("harness.event.v1", envelope)
        if actual != stored:
            db.close()
            raise EvidenceError("journal event hash mismatch")
        events.append({"sequence": sequence, "kind": kind, "payload": payload, "hash": stored})
        previous = stored
    db.close()
    return events

def get_acceptance(path, run_id):
    p = Path(path).resolve(strict=True)
    receipt = None
    if p.is_file() and p.suffix.lower() == ".json":
        receipt = parse(p.read_bytes())
        if not isinstance(receipt, dict):
            raise EvidenceError("acceptance receipt must be a JSON object")
        run_id = run_id or receipt.get("run_id")
        creation = receipt.get("creation", {})
        repo = creation.get("repository", {}).get("root") if isinstance(creation, dict) else None
        codex = creation.get("config", {}).get("codex") if isinstance(creation, dict) else None
        state_root = codex.get("state_root") if isinstance(codex, dict) else None
        return run_id, repo, state_root
    if not p.is_dir():
        raise EvidenceError("acceptance input must be a directory or JSON receipt")
    if not run_id:
        candidates = list((p / ".harness" / "runs").glob("*.jsonl")) if (p / ".harness" / "runs").is_dir() else []
        if len(candidates) == 1:
            run_id = candidates[0].stem
        else:
            raise EvidenceError("run id is required when directory does not contain exactly one controller journal")
    return run_id, str(p), None

def controller_path(run_id, repo, explicit):
    if explicit:
        return Path(explicit).resolve(strict=True)
    if not repo:
        raise EvidenceError("controller journal or source repository root is required")
    root = Path(repo)
    for candidate in (root / ".harness" / "runs" / (run_id + ".jsonl"), root / (run_id + ".jsonl")):
        if candidate.is_file():
            return candidate.resolve(strict=True)
    raise EvidenceError("controller journal not found")

def runtime_paths(run_id, state_root, explicit):
    found = [Path(p).resolve(strict=True) for p in (explicit or [])]
    if state_root:
        base = Path(state_root) / run_id
        if base.is_dir():
            found.extend(p.resolve(strict=True) for p in base.rglob("*.jsonl") if p.is_file())
    unique = {}
    for p in found:
        unique[str(p).casefold()] = p
    return list(unique.values())

def receipt_rows(events):
    rows = []
    kinds = {"writer.runtime-observed": "writer", "review.runtime-observed": "reviewer",
             "explorer.runtime-observed": "explorer", "graph.writer.runtime-observed": "writer"}
    for event in events:
        role = kinds.get(event["kind"])
        if not role:
            continue
        p = event["payload"]
        receipt = p.get("runtime_receipt") if event["kind"] == "graph.writer.runtime-observed" else p
        if isinstance(receipt, dict):
            rows.append({"role": role, "receipt": receipt, "event_sequence": event["sequence"]})
    return rows

def inspect_with_fabric(executable, repo, run_id, expected_head):
    try:
        run_id = validated_run_id(run_id)
        import subprocess
        binary = str(Path(executable).resolve(strict=True))
        root = str(Path(repo).resolve(strict=True))
        clean_env = {k: v for k, v in os.environ.items() if k != "ENGORCH_OTLP_TRACES_ENDPOINT"}
        export = subprocess.run([binary, "--root", root, "inspect", run_id, "--export-jsonl"],
                    capture_output=True, timeout=30, check=False, env=clean_env)
        if export.returncode != 0 or len(export.stdout) > (64 << 20) or not export.stdout.endswith(b"\n"):
            return None
        previous = ZERO
        last_hash = None
        lines = export.stdout[:-1].split(b"\n")
        if not lines:
            return None
        for index, line in enumerate(lines, 1):
            event = parse(line)
            if canon(event).encode("utf-8") != line or not isinstance(event, dict) or set(event) != {"version", "sequence", "previous", "kind", "payload", "hash"}:
                return None
            if (event.get("version") != 1 or event.get("sequence") != index or event.get("previous") != previous or
                    not isinstance(event.get("payload"), dict) or event.get("kind") == ""):
                return None
            digest = domain_hash("harness.event.v1", {k: event[k] for k in ("version", "sequence", "previous", "kind", "payload")})
            if digest != event.get("hash"):
                return None
            previous = last_hash = digest
        if last_hash != expected_head:
            return None
        proc = subprocess.run([binary, "--root", root, "inspect", run_id],
                    capture_output=True, timeout=30, check=False, env=clean_env)
        if proc.returncode != 0 or len(proc.stdout) > (64 << 20):
            return None
        snapshot = parse(proc.stdout)
        if not isinstance(snapshot, dict) or snapshot.get("run_id") != run_id:
            return None
        return snapshot
    except Exception:
        return None

def lexical_matches_controller(snapshot, record, installed):
    try:
        if not isinstance(snapshot, dict) or not isinstance(record, dict):
            return False
        lexical = snapshot.get("ri_lexical")
        if not isinstance(lexical, dict) or lexical.get("outcome") != "CONFIRMED":
            return False
        intent = lexical.get("intent", {}).get("plan")
        observed = lexical.get("observation", {}).get("build")
        base = record.get("base")
        if not isinstance(intent, dict) or not isinstance(observed, dict) or not isinstance(base, dict):
            return False
        manifest = base.get("manifest")
        if not isinstance(manifest, dict) or manifest.get("id") != intent.get("manifest_id"):
            return False
        if base.get("build") != observed:
            return False
        if record.get("executable_sha256") != intent.get("executable_sha256"):
            return False
        if os.path.normcase(os.path.realpath(record.get("executable", ""))) != os.path.normcase(os.path.realpath(intent.get("executable", ""))):
            return False
        if os.path.normcase(os.path.realpath(intent.get("executable", ""))) != os.path.normcase(os.path.realpath(installed)):
            return False
        repository = intent.get("repository")
        source = manifest.get("source")
        if not isinstance(repository, dict) or not isinstance(source, dict):
            return False
        repo_id = domain_hash("harness.repository.v1", repository)
        expected_source = {"repository_id": repo_id, "object_format": repository.get("object_format"),
                           "commit": repository.get("commit"), "tree": repository.get("tree")}
        return source == expected_source
    except Exception:
        return False

def load_lexical_manifest(ref):
    path = Path(ref["path"]).resolve(strict=True)
    if not path.is_file() or path.stat().st_size > (512 << 20):
        raise EvidenceError("lexical manifest unavailable or exceeds bound")
    expected_id = ref.get("id")
    expected_source = ref.get("source")
    expected_files = ref.get("files")
    if not isinstance(expected_id, str) or not isinstance(expected_source, dict):
        raise EvidenceError("lexical manifest reference malformed")
    files = []
    digest = hashlib.sha256()
    digest.update(b"harness.ri.lexical-manifest.v1\x00")
    source_seen = False
    total = 0
    with path.open("rb") as stream:
        for rawline in stream:
            total += len(rawline)
            if total > (512 << 20) or len(rawline) > ((1 << 20) + 1) or not rawline.endswith(b"\n"):
                raise EvidenceError("lexical manifest record boundary or size invalid")
            raw = rawline[:-1]
            entry = parse(raw)
            if canon(entry).encode("utf-8") != raw:
                raise EvidenceError("lexical manifest record is not canonical")
            if not isinstance(entry, dict) or set(entry) != {"kind", "value"}:
                raise EvidenceError("lexical manifest record shape invalid")
            if entry["kind"] == "source":
                if source_seen or files or entry["value"] != expected_source:
                    raise EvidenceError("lexical manifest source mismatch")
                source_seen = True
                digest.update(canon(expected_source).encode("utf-8") + b"\n")
            elif entry["kind"] == "file":
                value = entry["value"]
                if not source_seen or not isinstance(value, dict) or set(value) != {"path", "blob", "sha256", "bytes"}:
                    raise EvidenceError("lexical manifest file record malformed")
                name, blob, file_hash, size = value.get("path"), value.get("blob"), value.get("sha256"), value.get("bytes")
                if (not isinstance(name, str) or not name or name.startswith("/") or "\\" in name or
                        any(x in ("", ".", "..") for x in name.split("/")) or len(name) > 4096 or
                        (files and files[-1]["path"] >= name) or not isinstance(blob, str) or
                        len(blob) not in (40, 64) or any(c not in "0123456789abcdef" for c in blob) or
                        not isinstance(file_hash, str) or len(file_hash) != 64 or any(c not in "0123456789abcdef" for c in file_hash) or
                        not isinstance(size, int) or size < 0 or size > MAX_SAFE):
                    raise EvidenceError("lexical manifest file bounds invalid")
                files.append(value)
                digest.update(canon(value).encode("utf-8") + b"\n")
            else:
                raise EvidenceError("unknown lexical manifest record kind")
    if not source_seen or len(files) != expected_files or digest.hexdigest() != expected_id:
        raise EvidenceError("lexical manifest identity or file count mismatch")
    return files

def query_id(args):
    if not isinstance(args, dict):
        return None
    allowed = {"pattern", "fixed", "case_insensitive", "limit", "after", "path", "type"}
    if not set(args).issubset(allowed) or not {"pattern", "fixed", "case_insensitive", "limit", "after"}.issubset(args):
        return None
    pattern, fixed, insensitive, limit = args["pattern"], args["fixed"], args["case_insensitive"], args["limit"]
    if not isinstance(pattern, str) or len(pattern.encode("utf-8")) > 16 << 10 or not isinstance(fixed, bool) or not isinstance(insensitive, bool) or not isinstance(limit, int) or not 1 <= limit <= 100:
        return None
    path_filter = args.get("path", "")
    type_filter = args.get("type", "")
    if not isinstance(path_filter, str) or not isinstance(type_filter, str):
        return None
    if path_filter:
        if len(path_filter) > 4096 or any(ord(c) < 32 for c in path_filter) or any(c in path_filter for c in "\\:") or any(x in ("", ".", "..") for x in path_filter.split("/")):
            return None
    if type_filter:
        if len(type_filter) > 32 or not type_filter.isascii() or any(not (c.isalnum() or (i > 0 and c in "_+-")) for i, c in enumerate(type_filter)):
            return None
    after = args.get("after")
    if after is not None and (not isinstance(after, dict) or set(after) != {"manifest_id", "query_id", "path", "range"}):
        return None
    if path_filter == "" and type_filter == "":
        value = {"pattern": pattern, "fixed": fixed, "case_insensitive": insensitive,
                 "profile": "rust-regex-bytes-v1-tgrep-e2007b52"}
        return domain_hash("harness.ri.lexical-query.v1", value)
    value = {"pattern": pattern, "fixed": fixed, "case_insensitive": insensitive,
             "path": path_filter or None, "type": type_filter or None,
             "profile": "rust-regex-bytes-v1-tgrep-e2007b52-path-type-v1"}
    return domain_hash("harness.ri.lexical-query.v2", value)

def validate_search_response(record, content, request):
    try:
        if not isinstance(record, dict) or record.get("overlay") is not None or not isinstance(content, str) or not isinstance(request, dict):
            return False
        raw = content.encode("utf-8")
        if len(raw) > (768 << 10):
            return False
        result = parse(raw)
        if canon(result).encode("utf-8") != raw or not isinstance(result, dict):
            return False
        args = request.get("arguments")
        qid = query_id(args)
        base = record.get("base", {})
        manifest_ref = base.get("manifest", {})
        manifest_id = manifest_ref.get("id")
        if qid is None or result.get("query_id") != qid or result.get("manifest_id") != manifest_id:
            return False
        limit = args.get("limit")
        hits = result.get("matches")
        if (set(result) != {"query_id", "next", "manifest_id", "full_scan", "candidate_files", "searched_files", "truncated", "matches"} or
                type(result.get("full_scan")) is not bool or type(result.get("truncated")) is not bool or
                not isinstance(hits, list) or not hits or len(hits) > limit or
                type(result.get("candidate_files")) is not int or type(result.get("searched_files")) is not int or
                result["candidate_files"] < 1 or result["searched_files"] < 1 or
                result["searched_files"] > result["candidate_files"] or
                result.get("truncated") != (result.get("next") is not None) or
                (not result.get("truncated") and result["searched_files"] != result["candidate_files"])):
            return False
        files = load_lexical_manifest(manifest_ref)
        if result["candidate_files"] > len(files) or (result.get("full_scan") is True and result["candidate_files"] != len(files)):
            return False
        by_path = {f["path"]: f for f in files}
        prior = args.get("after")
        if prior is not None:
            if (not isinstance(prior, dict) or set(prior) != {"manifest_id", "query_id", "path", "range"} or
                    prior.get("manifest_id") != manifest_id or prior.get("query_id") != qid or
                    not isinstance(prior.get("path"), str) or prior.get("path") not in by_path or
                    not isinstance(prior.get("range"), list) or len(prior["range"]) != 2 or
                    any(type(x) is not int or x < 0 for x in prior["range"]) or
                    prior["range"][0] > prior["range"][1] or prior["range"][1] > by_path[prior["path"]]["bytes"]):
                return False
            prior_order = (prior["path"], prior["range"][0], prior["range"][1])
        else:
            prior_order = None
        last_order = None
        has_nonempty_hit = False
        source_format = manifest_ref.get("source", {}).get("object_format")
        blob_len = 64 if source_format == "sha256" else 40
        for hit in hits:
            if not isinstance(hit, dict) or set(hit) != {"path", "blob", "range"}:
                return False
            f = by_path.get(hit.get("path"))
            interval = hit.get("range")
            if (f is None or hit.get("blob") != f.get("blob") or not isinstance(interval, list) or len(interval) != 2 or
                    any(type(x) is not int or x < 0 for x in interval) or interval[0] > interval[1] or interval[1] > f.get("bytes", -1) or
                    not isinstance(hit.get("blob"), str) or len(hit["blob"]) != blob_len):
                return False
            order = (hit["path"], interval[0], interval[1])
            if prior_order is not None and not (prior_order < order):
                return False
            if last_order is not None and not (last_order < order):
                return False
            prior_order = last_order = order
            has_nonempty_hit = has_nonempty_hit or interval[1] > interval[0]
        if not has_nonempty_hit:
            return False
        if result.get("truncated"):
            nxt = result.get("next")
            last = hits[-1]
            if (len(hits) != limit or not isinstance(nxt, dict) or nxt.get("manifest_id") != manifest_id or nxt.get("query_id") != qid or
                    nxt.get("path") != last.get("path") or nxt.get("range") != last.get("range")):
                return False
        return True
    except Exception:
        return False

def inspect_runtime(path):
    events = open_events(path)
    intent = next((e["payload"] for e in events if e["kind"] == "runtime.intent"), None)
    if not isinstance(intent, dict) or not isinstance(intent.get("invocation"), dict):
        return None
    inv = intent["invocation"]
    lexical = next((e["payload"] for e in events if e["kind"] == "runtime.lexical-record"), None)
    thread = next((e["payload"] for e in events if e["kind"] == "runtime.thread"), None)
    result_event = next((e for e in reversed(events) if e["kind"] == "runtime.result"), None)
    turn_events = [(e["sequence"], e["payload"].get("id")) for e in events if e["kind"] == "runtime.turn" and isinstance(e["payload"], dict)]
    status_events = [(e["sequence"], e["payload"]) for e in events if e["kind"] == "runtime.turn-status" and isinstance(e["payload"], dict)]
    terminal_events = [(seq, status) for seq, status in status_events if status.get("status") == "completed"]
    runtime_shape_valid = (bool(events) and events[0]["kind"] == "runtime.intent" and
        sum(1 for e in events if e["kind"] == "runtime.intent") == 1 and
        sum(1 for e in events if e["kind"] == "runtime.thread") == 1 and
        sum(1 for e in events if e["kind"] == "runtime.lexical-record") == 1 and
        len(turn_events) == 1 and len(terminal_events) >= 1 and
        result_event is not None and events[-1]["kind"] == "runtime.result" and
        thread is not None and lexical is not None and
        next((e["sequence"] for e in events if e["kind"] == "runtime.lexical-record"), 1 << 60) <
        next((e["sequence"] for e in events if e["kind"] == "runtime.thread"), 0) < turn_events[0][0] < terminal_events[0][0] < result_event["sequence"] and
        all(status.get("id") == turn_events[0][1] and status.get("status") in ("inProgress", "completed") for _, status in status_events) and
        all(status.get("id") == turn_events[0][1] for _, status in terminal_events) and
        isinstance(result_event["payload"], dict) and result_event["payload"].get("turn_id") == turn_events[0][1])
    requests = {}
    successes = []
    failures = []
    response_ids = set()
    thread_id = thread.get("thread_id") if isinstance(thread, dict) else None
    result_turn_id = result_event["payload"].get("turn_id") if result_event and isinstance(result_event["payload"], dict) else None
    lexical_seq = next((e["sequence"] for e in events if e["kind"] == "runtime.lexical-record"), None)
    result_seq = result_event["sequence"] if result_event else None
    for e in events:
        if e["kind"] == "runtime.tool-request":
            request = e["payload"]
            call_id = request.get("callId")
            if call_id in requests:
                raise EvidenceError("duplicate tool call identity in runtime journal")
            requests[call_id] = (e["sequence"], request)
        elif e["kind"] == "runtime.tool-response":
            response = e["payload"]
            call_id = response.get("call_id")
            if call_id in response_ids:
                raise EvidenceError("duplicate tool response identity in runtime journal")
            response_ids.add(call_id)
            matched = requests.get(call_id)
            if matched and matched[1].get("tool") == "ri_search":
                request_seq, request = matched
                ordered = (runtime_shape_valid and request_seq < e["sequence"] and result_seq is not None and
                           turn_events and turn_events[0][0] < request_seq and terminal_events and
                           e["sequence"] < terminal_events[0][0] < result_seq)
                exact_turn = request.get("threadId") == thread_id and request.get("turnId") == result_turn_id
                exact_call = isinstance(call_id, str) and 0 < len(call_id) <= 256 and request.get("namespace") is None
                after_binding = lexical_seq is not None and lexical_seq < request_seq
                row = {"success": response.get("success") is True and ordered and exact_turn and exact_call and after_binding,
                       "call_id": call_id, "thread_id": request.get("threadId"),
                       "turn_id": request.get("turnId"), "request": request,
                       "content": response.get("content"), "request_sequence": request_seq,
                       "response_sequence": e["sequence"]}
                (successes if row["success"] else failures).append(row)
    return {"path": path, "events": events, "invocation": inv,
            "invocation_id": inv.get("id"), "role": inv.get("profile", {}).get("role"),
            "lexical": lexical, "thread": thread, "result_event": result_event,
            "ri_search_successes": successes, "ri_search_failures": failures,
            "runtime_shape_valid": runtime_shape_valid,
            "head": events[-1]["hash"] if events else None}

def main():
    ap = argparse.ArgumentParser(description="Collect read-only proof of successful installed RI agent-tool use.")
    ap.add_argument("--acceptance", help="acceptance JSON receipt or repository/run-root directory")
    ap.add_argument("--run-id")
    ap.add_argument("--controller-journal")
    ap.add_argument("--runtime-state-root", help="Codex state root; searches only its <run-id> subtree")
    ap.add_argument("--runtime-journal", action="append", default=[], help="explicit runtime SQLite journal; repeatable")
    ap.add_argument("--fabric-cli", help="trusted Fabric CLI executable used for authoritative read-only `inspect RUN` semantic replay")
    ap.add_argument("--repository-root", help="repository root when using --fabric-cli without an acceptance receipt")
    ap.add_argument("--installed-ri", required=True, help="installed Rust RI executable whose bytes are to be verified")
    args = ap.parse_args()
    out = {"proof": "NOT_PROVEN", "run_id": args.run_id, "installed_ri_sha256": None,
           "controller_chain": "NOT_CHECKED", "runtime_chain_count": 0,
           "runtime_invocations": [], "successful_installed_ri_searches": [], "reasons": []}
    try:
        if args.acceptance:
            run_id, repo, state_root = get_acceptance(args.acceptance, args.run_id)
        else:
            run_id, repo, state_root = args.run_id, args.repository_root, None
        if not run_id:
            raise EvidenceError("run id required")
        run_id = validated_run_id(run_id)
        out["run_id"] = run_id
        installed = Path(args.installed_ri).resolve(strict=True)
        if not installed.is_file():
            raise EvidenceError("installed RI path is not a file")
        installed_hash = file_sha256(installed)
        out["installed_ri_sha256"] = installed_hash
        controller = controller_path(run_id, repo, args.controller_journal)
        cevents = open_events(controller)
        out["controller_chain"] = "VALID"
        receipts = [r for r in receipt_rows(cevents) if r["receipt"].get("invocation_id")]
        semantic_snapshot = None
        if args.fabric_cli and repo:
            semantic_snapshot = inspect_with_fabric(args.fabric_cli, repo, run_id, cevents[-1]["hash"] if cevents else None)
            if semantic_snapshot:
                try:
                    rechecked = open_events(controller)
                    if not rechecked or rechecked[-1]["hash"] != cevents[-1]["hash"]:
                        semantic_snapshot = None
                except Exception:
                    semantic_snapshot = None
            out["controller_semantic_replay"] = "VALID" if semantic_snapshot else "NOT_VALIDATED"
        else:
            out["controller_semantic_replay"] = "NOT_VALIDATED"
        search_roots = args.runtime_state_root or state_root
        paths = runtime_paths(run_id, search_roots, args.runtime_journal)
        if not paths:
            out["reasons"].append("no runtime journals supplied or found")
        runtime_records = []
        for path in paths:
            try:
                runtime_record = inspect_runtime(str(path))
                if runtime_record:
                    runtime_records.append(runtime_record)
            except Exception:
                # Unrelated/stale journals are not allowed to abort scanning, but never count as proof.
                continue
        out["runtime_chain_count"] = len(runtime_records)
        used_receipts = set()
        for rt in runtime_records:
            inv_id = rt["invocation_id"]
            matching = [r for r in receipts if r["receipt"].get("invocation_id") == inv_id]
            expected_receipt_role = "writer" if rt["role"] == "fixer" else rt["role"]
            lexical = rt["lexical"] if isinstance(rt["lexical"], dict) else {}
            executable = lexical.get("executable")
            lexical_hash = lexical.get("executable_sha256")
            path_match = False
            if executable:
                try:
                    path_match = os.path.normcase(os.path.realpath(executable)) == os.path.normcase(os.path.realpath(installed))
                except Exception:
                    path_match = False
            lexical_controller_match = lexical_matches_controller(semantic_snapshot, lexical, installed) if semantic_snapshot else False
            report = {"role": rt["role"], "invocation_id": inv_id,
                      "controller_receipt_matches": len(matching),
                      "runtime_head_matches_receipt": False, "runtime_result_matches_receipt": False,
                      "lexical_binding_present": bool(lexical), "lexical_executable_hash_matches_installed": lexical_hash == installed_hash,
                      "lexical_executable_path_matches_installed": path_match,
                      "lexical_base_matches_confirmed_controller_build": lexical_controller_match,
                      "validated_nonempty_ri_search_count": 0,
                      "successful_ri_search_count": len(rt["ri_search_successes"]),
                      "failed_ri_search_count": len(rt["ri_search_failures"])}
            role_domains = {"writer": "harness.writer-result.v1", "fixer": "harness.writer-result.v1",
                            "reviewer": "harness.review-result.v1", "explorer": "harness.explorer-result.v1"}
            if len(matching) == 1 and matching[0]["role"] == expected_receipt_role:
                rec = matching[0]["receipt"]
                result_event = rt["result_event"]
                report["runtime_head_matches_receipt"] = rec.get("journal_head") == rt["head"]
                if result_event and isinstance(result_event["payload"], dict):
                    result = result_event["payload"].get("result")
                    domain = role_domains.get(rt["role"])
                    if result is not None and domain:
                        report["runtime_result_matches_receipt"] = domain_hash(domain, result) == rec.get("result_hash")
                thread_id = (rt["thread"] or {}).get("thread_id") if isinstance(rt["thread"], dict) else None
                turn_id = rt["result_event"]["payload"].get("turn_id") if rt["result_event"] else None
                report["receipt_thread_matches"] = rec.get("thread_id") == thread_id
                report["receipt_turn_matches"] = rec.get("turn_id") == turn_id
                used_receipts.add((matching[0]["role"], matching[0]["event_sequence"]))
            else:
                report["receipt_thread_matches"] = False
                report["receipt_turn_matches"] = False
                if matching and len(matching) == 1:
                    report["controller_receipt_matches"] = 0
            out["runtime_invocations"].append(report)
            for call in rt["ri_search_successes"]:
                call = dict(call)
                call["role"] = rt["role"]
                call["invocation_id"] = inv_id
                call["tied_to_controller_receipt"] = (len(matching) == 1 and report["runtime_head_matches_receipt"] and
                    report["runtime_result_matches_receipt"] and report["receipt_thread_matches"] and report["receipt_turn_matches"])
                call["installed_executable_match"] = (report["lexical_executable_hash_matches_installed"] and report["lexical_executable_path_matches_installed"])
                content = call.pop("content", None)
                if (semantic_snapshot and call["tied_to_controller_receipt"] and call["installed_executable_match"]
                        and lexical_controller_match and validate_search_response(rt["lexical"], content, call.get("request"))):
                    call.pop("request", None)
                    call.pop("request_sequence", None)
                    call.pop("response_sequence", None)
                    report["validated_nonempty_ri_search_count"] += 1
                    out["successful_installed_ri_searches"].append(call)
        if out["successful_installed_ri_searches"]:
            out["proof"] = "PROVEN_JOURNAL_BOUND"
            out["reasons"].append("successful ri_search request/response is bound to the installed executable and a controller-observed runtime result")
        elif not any(rt.get("successful_ri_search_count", 0) for rt in out["runtime_invocations"]):
            out["reasons"].append("no successful ri_search request/response was found")
        elif not semantic_snapshot:
            out["reasons"].append("tool call exists, but authoritative Fabric inspect semantic replay was not supplied or did not validate")
        else:
            out["reasons"].append("tool call exists, but lexical publication, query, nonempty hit, or installed executable bindings did not all validate")
    except EvidenceError as exc:
        out["proof"] = "NOT_PROVEN"
        out["reasons"].append(str(exc))
    except Exception:
        out["proof"] = "NOT_PROVEN"
        out["reasons"].append("input unavailable or journal could not be validated")
    out["evidence_scope"] = "local integrity-checked journal linkage; not external notarization"
    print(json.dumps(out, ensure_ascii=False, indent=2, sort_keys=True))
    return 0 if out["proof"] == "PROVEN_JOURNAL_BOUND" else 2

if __name__ == "__main__":
    sys.exit(main())
