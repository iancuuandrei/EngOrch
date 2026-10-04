# Verified fresh-context rounds

**Prepared procedure — NOT EXECUTED.** This guide describes a supported manual
path from an accepted run to a fresh run based on that exact committed
candidate. It does not claim that automatic compaction has triggered or that a
multi-round Humanize evaluation has passed. Use only after the intended frozen
Fabric, runner, Go, Codex, RI, and acceptance inputs are qualified and pinned.

## Why a new source checkout is required

`prepare-commit` captures the READY candidate from its isolated workspace.
After an exact approval, `commit` updates the workspace's
`refs/heads/harness/<RUN>` branch and records a read-back receipt. It does not
move the original configured checkout's `HEAD`. A new run against the same
`--root` therefore starts from that checkout's prior commit. Keep the accepted
run terminal; do not try to resume READY as another implementation round.

## Confirm and commit round one

Before commit, require `checkpoint RUN` to report `accepted_checkpoint: true`
and inspect the READY snapshot. The checkpoint must bind the confirmed
candidate, all required native checks passing on that candidate, and an
approved review of the same candidate and verification plan. Stop if any
external intent is UNKNOWN or active.

Prepare an explicit metadata file. The names, email, Unix timestamps, and
message are operator-chosen; the example values below are placeholders:

```json
{
  "author": {
    "name": "Example Operator",
    "email": "operator@example.invalid",
    "unix_seconds": 1791072000
  },
  "committer": {
    "name": "Example Operator",
    "email": "operator@example.invalid",
    "unix_seconds": 1791072000
  },
  "message": "Add validated ParseBytes separators\n"
}
```

Run the exact local commit flow and retain its output:

```powershell
fabric --root $Round1Root prepare-commit $Round1Run $MetadataJson
# Save its canonical JSON output unchanged at $PreviewPath; copy intent_id to $IntentId.
fabric --root $Round1Root commit $Round1Run $PreviewPath $IntentId $Actor
fabric --root $Round1Root inspect $Round1Run
```

Pass the `intent_id` returned in the exact preview to `commit`; do not edit or
reconstruct the preview. The final inspect must show `state: COMMITTED`,
`commit.outcome: CONFIRMED`, and `commit.intent.commit_id` equal to the new
candidate/workspace source commit. The `commit.observed` receipt is the durable
confirmation. If the result is UNKNOWN, stop and use only the existing
reconciliation path; do not retry the commit.

## Create the round-two source and run

Create a separate detached source checkout at that exact confirmed commit.
Keep round one's checkout, worktree, and journal untouched:

```powershell
git -C $Round1Root worktree add --detach $Round2Root $CommitId
Copy-Item -LiteralPath (Join-Path $Round1Root 'harness.toml') `
  -Destination (Join-Path $Round2Root 'harness.toml')
```

The copied `harness.toml` retains the operator's model, verification, and
runtime policy. Change only `controller_state_root` to a new absolute path
outside the source checkout and distinct from round one's state root. Record
the copied config's SHA-256. Run `fabric --root $Round2Root doctor` and verify
`git -C $Round2Root rev-parse HEAD` equals `$CommitId` before creating a run.
If either identity differs, stop.

Create round two with a new objective and run ID. It has fresh planner/writer/
reviewer threads. For an automatic-compaction observation, pass an explicit
`--auto-compact-token-limit N` selected for the pinned model's effective window;
this binds the requested setting but does not prove it activates. Keep the
frozen repair limit and all invocation/token budgets unchanged, and keep
`--max-parallel 1` so the writer turn remains attributable:

```powershell
fabric --root $Round2Root run --autonomous --max-parallel 1 `
  --auto-compact-token-limit $TokenLimit --file $Round2Objective
```

Use the unchanged public Humanize task inputs in `evals/v1/manifest.json`:

1. Round one: `go-humanize`, which adds correctly validated underscore
   separators to `ParseBytes` inputs.
2. Round two: `go-humanize-commaf-performance`, which reduces `Commaf`
   allocations while preserving its public behavior.

Both rows pin
`github.com/dustin/go-humanize@a1b4e66b9a6d890e9e15e7091cf16c8032367d6e`.
Preserve each row's exact objective and native argv: round one runs
`go test -count=1 .`; round two runs `go test -count=1 ./...`. Retain the
existing candidate-bound held-out assertions from `humanize.heldout_test.go`
and `commaf_performance.heldout_test.go`, respectively, without putting them
in the model-visible checkout. Keep review and each task's frozen repair
budget. Round two must start at round one's confirmed commit, not the original
manifest SHA. Give this chained comparison its own evaluation identity and
record the parent commit; do not reuse or rewrite historical evaluation
journals. The existing fixed-source evaluation runner does not by itself
represent this parent-commit chain.

## Interruption, resume, and compaction evidence

A cooperative stop can use the existing lifecycle commands: request pause,
let admitted work settle, attest quiescence, then reopen and continue the exact
run:

```powershell
fabric --root $Round2Root pause $Round2Run $Actor $PauseNonce
fabric --root $Round2Root settle-lifecycle $Round2Run $Actor `
  'Operator confirmed workloads stopped' workloads-stopped
fabric --root $Round2Root resume $Round2Run $Actor $ResumeNonce
fabric --root $Round2Root resume --autonomous $Round2Run
```

The stop evidence must be true and specific. Inspect before resuming; retain
UNKNOWN outcomes and consumed repairs. Resume only the exact nonterminal run
with its original invocation bindings. This is a cooperative lifecycle
pause/resume, not proof of abrupt-process or in-flight provider recovery. If
the run cannot reach a settled pause without interrupting an uncertain effect,
stop and record that interruption qualification as NOT RUN.

Report compaction as observed only when one invocation/thread/turn has a
matching `contextCompaction` item ID in both `item/started` and
`item/completed` and the turn completes. A configured threshold, a fresh
round, a mock fixture, or a resumed run is not that evidence. Preserve the
same-run usage and lifecycle receipts, then require the unchanged native,
held-out, candidate, and review gates. Never use manual compact RPCs, add
filler turns, weaken checks, or resend an uncertain invocation.
