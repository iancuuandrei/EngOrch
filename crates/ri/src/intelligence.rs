#![allow(missing_docs)]
//! Bounded, content-addressed Go source facts. These are syntax observations,
//! not an index, a semantic resolver, or an authority to run generated code.

use crate::canonical;
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use std::{
    fs,
    io::Read,
    path::{Path, PathBuf},
    time::{Duration, Instant, SystemTime},
};
use tree_sitter::{Node, ParseOptions, Parser};

pub const GO_FACT_SCHEMA: &str = "engorch.go-file-facts.v1";

const GO_FACTS_CACHE_MAX_ENTRIES: usize = 1024;
const GO_FACTS_CACHE_MAX_BYTES: u64 = 64 << 20;
const GO_FACTS_CACHE_MAX_SCAN_ENTRIES: usize = 4096;
const GO_FACTS_CACHE_LOCK: &str = ".cache-write.lock";
const GO_FACTS_CACHE_POLICY: &str = ".cache-policy-v1";
const GO_FACTS_CACHE_POLICY_BYTES: &[u8] = b"engorch-go-facts-cache-v1\n";

#[derive(Clone, Copy)]
struct CacheLimits {
    entries: usize,
    bytes: u64,
    scan_entries: usize,
}

const GO_FACTS_CACHE_LIMITS: CacheLimits = CacheLimits {
    entries: GO_FACTS_CACHE_MAX_ENTRIES,
    bytes: GO_FACTS_CACHE_MAX_BYTES,
    scan_entries: GO_FACTS_CACHE_MAX_SCAN_ENTRIES,
};

struct CacheEntry {
    path: PathBuf,
    name: String,
    bytes: u64,
    modified: SystemTime,
    protected: bool,
}

struct CacheWriteLock(PathBuf);

impl Drop for CacheWriteLock {
    fn drop(&mut self) {
        let _ = fs::remove_file(&self.0);
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Range {
    pub start_byte: usize,
    pub end_byte: usize,
}
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Symbol {
    pub name: String,
    pub kind: String,
    pub range: Range,
    pub test: bool,
}
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Import {
    pub path: String,
    pub alias: Option<String>,
    pub range: Range,
}
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Call {
    pub spelling: String,
    pub resolution: String,
    pub range: Range,
}
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct GoFileFacts {
    pub schema: String,
    pub language: String,
    pub parser_version: String,
    pub path: String,
    pub source_sha256: String,
    pub producer_sha256: String,
    pub cache_key: String,
    pub body_sha256: String,
    pub syntax_errors: bool,
    pub coverage: String,
    pub declarations: Vec<Symbol>,
    pub imports: Vec<Import>,
    pub calls: Vec<Call>,
    pub generated_markers: Vec<String>,
    pub cache: String,
    pub parse_count: u8,
}

fn digest(bytes: &[u8]) -> String {
    Sha256::digest(bytes)
        .iter()
        .map(|b| format!("{b:02x}"))
        .collect()
}
fn valid_digest(s: &str) -> bool {
    s.len() == 64
        && s.bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}
fn valid_path(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 4096
        && !s.contains(['\\', ':'])
        && !s.split('/').any(|p| p.is_empty() || p == "." || p == "..")
}
fn text(node: Node<'_>, src: &[u8]) -> Result<String, String> {
    node.utf8_text(src)
        .map(str::to_owned)
        .map_err(|e| e.to_string())
}
fn range(n: Node<'_>) -> Range {
    Range {
        start_byte: n.start_byte(),
        end_byte: n.end_byte(),
    }
}
fn key(path: &str, source: &str, producer: &str) -> Result<String, String> {
    Ok(canonical::hash(
        "harness.ri.go-file-facts.v1",
        &canonical::encode(
            &serde_json::json!({"schema":GO_FACT_SCHEMA,"language":"go","parser":"tree-sitter-go-0.25.0","path":path,"source_sha256":source,"producer_sha256":producer}),
        )?,
    ))
}
fn body(mut value: GoFileFacts) -> Result<String, String> {
    value.body_sha256.clear();
    value.cache.clear();
    value.parse_count = 0;
    Ok(digest(&canonical::encode(
        &serde_json::to_value(value).map_err(|e| e.to_string())?,
    )?))
}
fn go_string(q: &str) -> Result<String, String> {
    if q.starts_with('`') && q.ends_with('`') {
        return Ok(q[1..q.len() - 1].replace('\r', ""));
    }
    let b = q.as_bytes();
    if b.len() < 2 || b[0] != b'"' || *b.last().unwrap() != b'"' {
        return Err("invalid Go import string".into());
    }
    let mut out = Vec::new();
    let mut i = 1;
    while i + 1 < b.len() {
        if b[i] != b'\\' {
            let end = b[i..b.len() - 1]
                .iter()
                .position(|byte| *byte == b'\\')
                .map_or(b.len() - 1, |offset| i + offset);
            out.extend_from_slice(&b[i..end]);
            i = end;
            continue;
        }
        i += 1;
        if i >= b.len() - 1 {
            return Err("invalid Go import escape".into());
        }
        let c = b[i];
        i += 1;
        match c {
            b'a' => out.push(7),
            b'b' => out.push(8),
            b'f' => out.push(12),
            b'n' => out.push(b'\n'),
            b't' => out.push(b'\t'),
            b'r' => out.push(b'\r'),
            b'v' => out.push(11),
            b'\\' => out.push(b'\\'),
            b'"' => out.push(b'"'),
            b'x' => {
                if i + 2 > b.len() - 1 {
                    return Err("invalid Go hex escape".into());
                }
                let v = u8::from_str_radix(
                    std::str::from_utf8(&b[i..i + 2]).map_err(|_| "invalid Go hex escape")?,
                    16,
                )
                .map_err(|_| "invalid Go hex escape")?;
                out.push(v);
                i += 2
            }
            b'0'..=b'7' => {
                if i + 2 > b.len() - 1 {
                    return Err("invalid Go octal escape".into());
                }
                let mut s = String::new();
                s.push(c as char);
                s.push(b[i] as char);
                s.push(b[i + 1] as char);
                let v = u8::from_str_radix(&s, 8).map_err(|_| "invalid Go octal escape")?;
                out.push(v);
                i += 2
            }
            b'u' | b'U' => {
                let n = if c == b'u' { 4 } else { 8 };
                if i + n > b.len() - 1 {
                    return Err("invalid Go unicode escape".into());
                }
                let v = u32::from_str_radix(
                    std::str::from_utf8(&b[i..i + n]).map_err(|_| "invalid Go unicode escape")?,
                    16,
                )
                .map_err(|_| "invalid Go unicode escape")?;
                let mut encoded = [0u8; 4];
                out.extend_from_slice(
                    char::from_u32(v)
                        .ok_or("invalid Go unicode escape")?
                        .encode_utf8(&mut encoded)
                        .as_bytes(),
                );
                i += n
            }
            _ => return Err("invalid Go import escape".into()),
        }
    }
    String::from_utf8(out).map_err(|_| "Go import path is not UTF-8".into())
}
fn valid_fact(
    f: &GoFileFacts,
    path: &str,
    source: &str,
    producer: &str,
    key: &str,
    bytes: &[u8],
) -> bool {
    let normalized_source = std::str::from_utf8(bytes).unwrap_or("").replace('\r', "");
    let exact = |r: &Range, value: &str| {
        r.start_byte < r.end_byte
            && r.end_byte <= bytes.len()
            && &bytes[r.start_byte..r.end_byte] == value.as_bytes()
    };
    f.schema == GO_FACT_SCHEMA
        && f.language == "go"
        && f.parser_version == "tree-sitter-go-0.25.0"
        && f.path == path
        && f.source_sha256 == source
        && f.producer_sha256 == producer
        && f.cache_key == key
        && f.coverage == "PARTIAL"
        && f.declarations
            .len()
            .saturating_add(f.imports.len())
            .saturating_add(f.calls.len())
            .saturating_add(f.generated_markers.len())
            <= 100_000
        && f.declarations.iter().all(|x| {
            matches!(
                x.kind.as_str(),
                "function_declaration" | "method_declaration" | "type_spec" | "type_alias"
            ) && exact(&x.range, &x.name)
        })
        && f.declarations
            .windows(2)
            .all(|w| (&w[0].range.start_byte, &w[0].name) <= (&w[1].range.start_byte, &w[1].name))
        && f.imports.iter().all(|x| {
            x.path.len() <= 4096
                && x.range.start_byte < x.range.end_byte
                && x.range.end_byte <= bytes.len()
        })
        && f.imports
            .windows(2)
            .all(|w| (&w[0].range.start_byte, &w[0].path) <= (&w[1].range.start_byte, &w[1].path))
        && f.calls
            .iter()
            .all(|x| x.resolution == "UNRESOLVED" && exact(&x.range, &x.spelling))
        && f.calls.windows(2).all(|w| {
            (&w[0].range.start_byte, &w[0].spelling) <= (&w[1].range.start_byte, &w[1].spelling)
        })
        && f.generated_markers.windows(2).all(|w| w[0] <= w[1])
        && f.generated_markers
            .iter()
            .all(|marker| !marker.is_empty() && normalized_source.contains(marker))
}
fn validate_cache_dir(dir: &Path) -> Result<(), String> {
    if !dir.is_absolute() {
        return Err("cache directory must be absolute".into());
    }
    let mut cur = Some(dir);
    while let Some(p) = cur {
        match fs::symlink_metadata(p) {
            Ok(meta) => {
                if cache_reparse(&meta) {
                    return Err("cache directory symlink rejected".into());
                }
                if !meta.is_dir() {
                    return Err("cache ancestor is not a directory".into());
                }
            }
            Err(error) if error.kind() == std::io::ErrorKind::NotFound => {}
            Err(error) => return Err(error.to_string()),
        }
        cur = p.parent()
    }
    Ok(())
}

fn cache_reparse(meta: &fs::Metadata) -> bool {
    #[cfg(windows)]
    {
        use std::os::windows::fs::MetadataExt;
        meta.file_attributes() & 0x400 != 0
    }
    #[cfg(not(windows))]
    {
        meta.file_type().is_symlink()
    }
}

fn acquire_cache_write_lock(dir: &Path) -> Result<Option<CacheWriteLock>, String> {
    let path = dir.join(GO_FACTS_CACHE_LOCK);
    match fs::OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&path)
    {
        Ok(_) => Ok(Some(CacheWriteLock(path))),
        Err(error) if error.kind() == std::io::ErrorKind::AlreadyExists => Ok(None),
        Err(error) => Err(error.to_string()),
    }
}

fn cache_policy_matches(dir: &Path) -> Result<bool, String> {
    let path = dir.join(GO_FACTS_CACHE_POLICY);
    let metadata = match fs::symlink_metadata(&path) {
        Ok(metadata) => metadata,
        Err(error) if error.kind() == std::io::ErrorKind::NotFound => return Ok(false),
        Err(error) => return Err(error.to_string()),
    };
    if cache_reparse(&metadata)
        || !metadata.is_file()
        || metadata.len() > GO_FACTS_CACHE_POLICY_BYTES.len() as u64
    {
        return Err("Go facts cache policy marker is not a bounded regular file".into());
    }
    let bytes = fs::read(path).map_err(|error| error.to_string())?;
    if bytes == GO_FACTS_CACHE_POLICY_BYTES {
        Ok(true)
    } else {
        Err("Go facts cache policy marker is malformed".into())
    }
}

fn write_cache_policy_marker(dir: &Path) -> Result<(), String> {
    let key = digest(GO_FACTS_CACHE_POLICY_BYTES);
    let nonce = SystemTime::now()
        .duration_since(SystemTime::UNIX_EPOCH)
        .map_err(|error| error.to_string())?
        .as_nanos();
    let temp = dir.join(format!(".{key}.{}.{}.tmp", std::process::id(), nonce));
    let target = dir.join(GO_FACTS_CACHE_POLICY);
    let mut output = fs::OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&temp)
        .map_err(|error| error.to_string())?;
    std::io::Write::write_all(&mut output, GO_FACTS_CACHE_POLICY_BYTES)
        .map_err(|error| error.to_string())?;
    drop(output);
    fs::rename(temp, target).map_err(|error| error.to_string())
}

fn initialize_cache_policy_locked(dir: &Path, limits: CacheLimits) -> Result<(), String> {
    if cache_policy_matches(dir)? {
        return Ok(());
    }
    if GO_FACTS_CACHE_POLICY_BYTES.len() as u64 > limits.bytes {
        return Err("Go facts cache byte bound cannot hold its policy marker".into());
    }
    let mut entries = scan_cache_entries(dir, limits.scan_entries)?;
    loop {
        let total = entries.iter().try_fold(0u64, |sum, entry| {
            sum.checked_add(entry.bytes)
                .ok_or_else(|| "Go facts cache byte total overflow".to_string())
        })?;
        let facts = entries.iter().filter(|entry| !entry.protected).count();
        if facts <= limits.entries
            && total <= limits.bytes - GO_FACTS_CACHE_POLICY_BYTES.len() as u64
        {
            break;
        }
        let Some((oldest_index, _)) = entries
            .iter()
            .enumerate()
            .filter(|(_, entry)| !entry.protected)
            .min_by(|(_, a), (_, b)| {
                a.modified
                    .cmp(&b.modified)
                    .then_with(|| a.name.cmp(&b.name))
            })
        else {
            return Err("Go facts cache cannot migrate within resident bounds".into());
        };
        let oldest = entries.remove(oldest_index);
        fs::remove_file(oldest.path).map_err(|error| error.to_string())?;
    }
    write_cache_policy_marker(dir)
}

fn ensure_cache_policy(dir: &Path, limits: CacheLimits) -> Result<bool, String> {
    validate_cache_dir(dir)?;
    fs::create_dir_all(dir).map_err(|error| error.to_string())?;
    validate_cache_dir(dir)?;
    if cache_policy_matches(dir)? {
        return Ok(true);
    }
    let Some(_lock) = acquire_cache_write_lock(dir)? else {
        return Ok(false);
    };
    initialize_cache_policy_locked(dir, limits)?;
    Ok(true)
}

fn valid_cache_entry_name(name: &str) -> bool {
    name.len() == 69 && name.is_ascii() && name.ends_with(".json") && valid_digest(&name[..64])
}

fn valid_cache_temp_name(name: &str) -> bool {
    let Some(body) = name.strip_prefix('.') else {
        return false;
    };
    let Some(body) = body.strip_suffix(".tmp") else {
        return false;
    };
    let parts: Vec<_> = body.split('.').collect();
    parts.len() == 3
        && valid_digest(parts[0])
        && !parts[1].is_empty()
        && parts[1].bytes().all(|b| b.is_ascii_digit())
        && !parts[2].is_empty()
        && parts[2].bytes().all(|b| b.is_ascii_digit())
}

fn scan_cache_entries(dir: &Path, max_scan_entries: usize) -> Result<Vec<CacheEntry>, String> {
    let mut entries = Vec::new();
    let mut stale_temps = Vec::new();
    for (index, item) in fs::read_dir(dir)
        .map_err(|error| error.to_string())?
        .enumerate()
    {
        if index >= max_scan_entries {
            return Err("Go facts cache scan bound exceeded".into());
        }
        let item = item.map_err(|error| error.to_string())?;
        let path = item.path();
        let name = item.file_name().to_string_lossy().into_owned();
        let metadata = fs::symlink_metadata(&path).map_err(|error| error.to_string())?;
        if cache_reparse(&metadata) || !metadata.is_file() {
            return Err("Go facts cache contains a non-regular entry".into());
        }
        if name == GO_FACTS_CACHE_LOCK {
            if metadata.len() != 0 {
                return Err("Go facts cache lock has unexpected contents".into());
            }
            continue;
        }
        if name == GO_FACTS_CACHE_POLICY {
            if !cache_policy_matches(dir)? {
                return Err("Go facts cache policy marker is malformed".into());
            }
            let modified = metadata.modified().map_err(|error| error.to_string())?;
            entries.push(CacheEntry {
                path,
                name,
                bytes: metadata.len(),
                modified,
                protected: true,
            });
            continue;
        }
        if valid_cache_temp_name(&name) {
            stale_temps.push(path);
            continue;
        }
        if !valid_cache_entry_name(&name) {
            return Err("Go facts cache contains a malformed entry name".into());
        }
        let modified = metadata.modified().map_err(|error| error.to_string())?;
        entries.push(CacheEntry {
            path,
            name,
            bytes: metadata.len(),
            modified,
            protected: false,
        });
    }
    // Crash-left temporary files are unpublished data and can be removed once
    // the exclusive cache lock is held. Their content is never decoded.
    for path in stale_temps {
        fs::remove_file(path).map_err(|error| error.to_string())?;
    }
    entries.sort_by(|a, b| a.name.cmp(&b.name));
    Ok(entries)
}

fn make_cache_room(
    entries: &mut Vec<CacheEntry>,
    target_name: &str,
    new_bytes: u64,
    limits: CacheLimits,
) -> Result<(), String> {
    let marker_bytes = GO_FACTS_CACHE_POLICY_BYTES.len() as u64;
    if new_bytes > limits.bytes.saturating_sub(marker_bytes) {
        return Err("Go facts cache entry exceeds resident byte bound".into());
    }
    if let Some(index) = entries.iter().position(|entry| entry.name == target_name) {
        let replaced = entries.remove(index);
        fs::remove_file(replaced.path).map_err(|error| error.to_string())?;
    }
    loop {
        let total = entries.iter().try_fold(0u64, |sum, entry| {
            sum.checked_add(entry.bytes)
                .ok_or_else(|| "Go facts cache byte total overflow".to_string())
        })?;
        let facts = entries.iter().filter(|entry| !entry.protected).count();
        if facts < limits.entries && total <= limits.bytes.saturating_sub(new_bytes) {
            break;
        }
        let Some((oldest_index, _)) = entries
            .iter()
            .enumerate()
            .filter(|(_, entry)| !entry.protected)
            .min_by(|(_, a), (_, b)| {
                a.modified
                    .cmp(&b.modified)
                    .then_with(|| a.name.cmp(&b.name))
            })
        else {
            return Err("Go facts cache cannot fit entry within resident bounds".into());
        };
        let oldest = entries.remove(oldest_index);
        fs::remove_file(oldest.path).map_err(|error| error.to_string())?;
    }
    Ok(())
}

fn write_cache_entry_with_limits(
    dir: &Path,
    key: &str,
    raw: &[u8],
    limits: CacheLimits,
) -> Result<(), String> {
    if !valid_digest(key) || raw.len() > canonical::MAX_BYTES {
        return Err("invalid Go facts cache write".into());
    }
    validate_cache_dir(dir)?;
    fs::create_dir_all(dir).map_err(|error| error.to_string())?;
    validate_cache_dir(dir)?;
    let Some(_lock) = acquire_cache_write_lock(dir)? else {
        // Cache contention is a miss for storage purposes, never a parser
        // failure or a reason to repeat/skip a repository effect.
        return Ok(());
    };
    initialize_cache_policy_locked(dir, limits)?;
    let mut entries = scan_cache_entries(dir, limits.scan_entries)?;
    let target_name = format!("{key}.json");
    make_cache_room(&mut entries, &target_name, raw.len() as u64, limits)?;
    let nonce = SystemTime::now()
        .duration_since(SystemTime::UNIX_EPOCH)
        .map_err(|error| error.to_string())?
        .as_nanos();
    let temp = dir.join(format!(".{key}.{}.{}.tmp", std::process::id(), nonce));
    let target = dir.join(target_name);
    let mut output = fs::OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&temp)
        .map_err(|error| error.to_string())?;
    std::io::Write::write_all(&mut output, raw).map_err(|error| error.to_string())?;
    drop(output);
    fs::rename(&temp, &target).map_err(|error| error.to_string())?;
    Ok(())
}

fn write_cache_entry(dir: &Path, key: &str, raw: &[u8]) -> Result<(), String> {
    write_cache_entry_with_limits(dir, key, raw, GO_FACTS_CACHE_LIMITS)
}

pub fn go_file_facts(
    path: &str,
    source: &[u8],
    expected: &str,
    producer: &str,
    cache_dir: Option<&Path>,
) -> Result<GoFileFacts, String> {
    if !valid_path(path)
        || source.len() > 1 << 20
        || std::str::from_utf8(source).is_err()
        || !valid_digest(expected)
        || !valid_digest(producer)
        || digest(source) != expected
    {
        return Err("invalid Go facts source binding".into());
    }
    let k = key(path, expected, producer)?;
    if let Some(dir) = cache_dir {
        validate_cache_dir(dir)?;
        let cache_ready = ensure_cache_policy(dir, GO_FACTS_CACHE_LIMITS).unwrap_or(false);
        if cache_ready {
            let p = dir.join(format!("{k}.json"));
            if let Ok(meta) = fs::symlink_metadata(&p) {
                if cache_reparse(&meta) || !meta.is_file() {
                    return Err("cache entry must be a regular file".into());
                }
                if meta.file_type().is_file() {
                    if let Ok(file) = fs::File::open(&p) {
                        let mut raw = Vec::new();
                        if file
                            .take(canonical::MAX_BYTES as u64 + 1)
                            .read_to_end(&mut raw)
                            .is_ok()
                            && raw.len() <= canonical::MAX_BYTES
                        {
                            if let Ok(mut fact) = canonical::decode::<GoFileFacts>(&raw) {
                                let b = body(fact.clone())?;
                                if valid_fact(&fact, path, expected, producer, &k, source)
                                    && fact.body_sha256 == b
                                {
                                    fact.cache = "hit".into();
                                    fact.parse_count = 0;
                                    return Ok(fact);
                                }
                            }
                        }
                    }
                }
            }
        }
    }
    let deadline = Instant::now() + Duration::from_secs(5);
    let mut parser = Parser::new();
    parser
        .set_language(&tree_sitter_go::LANGUAGE.into())
        .map_err(|e| e.to_string())?;
    let mut progress = |_: &tree_sitter::ParseState| {
        if Instant::now() > deadline {
            std::ops::ControlFlow::Break(())
        } else {
            std::ops::ControlFlow::Continue(())
        }
    };
    let tree = parser
        .parse_with_options(
            &mut |o, _| &source[o.min(source.len())..],
            None,
            Some(ParseOptions::new().progress_callback(&mut progress)),
        )
        .ok_or("Go parser cancelled")?;
    let mut f = GoFileFacts {
        schema: GO_FACT_SCHEMA.into(),
        language: "go".into(),
        parser_version: "tree-sitter-go-0.25.0".into(),
        path: path.into(),
        source_sha256: expected.into(),
        producer_sha256: producer.into(),
        cache_key: k.clone(),
        body_sha256: String::new(),
        syntax_errors: tree.root_node().has_error(),
        coverage: "PARTIAL".into(),
        declarations: vec![],
        imports: vec![],
        calls: vec![],
        generated_markers: vec![],
        cache: "miss".into(),
        parse_count: 1,
    };
    let mut stack = vec![tree.root_node()];
    let mut seen = 0usize;
    while let Some(n) = stack.pop() {
        seen += 1;
        if seen > 100_000 || Instant::now() > deadline {
            return Err("Go facts traversal budget exceeded".into());
        };
        match n.kind() {
            "function_declaration" | "method_declaration" | "type_spec" | "type_alias" => {
                if let Some(name) = n.child_by_field_name("name") {
                    let x = text(name, source)?;
                    f.declarations.push(Symbol {
                        test: x.starts_with("Test")
                            || x.starts_with("Benchmark")
                            || path.ends_with("_test.go"),
                        name: x,
                        kind: n.kind().into(),
                        range: range(name),
                    })
                }
            }
            "import_spec" => {
                let mut c = n.walk();
                let ch: Vec<_> = n.named_children(&mut c).collect();
                if let Some(p) = ch.iter().find(|x| {
                    matches!(
                        x.kind(),
                        "interpreted_string_literal" | "raw_string_literal"
                    )
                }) {
                    let q = text(*p, source)?;
                    f.imports.push(Import {
                        path: go_string(&q)?,
                        alias: ch
                            .iter()
                            .find(|x| {
                                matches!(
                                    x.kind(),
                                    "package_identifier" | "dot" | "blank_identifier"
                                )
                            })
                            .map(|x| text(*x, source))
                            .transpose()?,
                        range: range(n),
                    })
                }
            }
            "call_expression" => {
                if let Some(fun) = n.child_by_field_name("function") {
                    f.calls.push(Call {
                        spelling: text(fun, source)?,
                        resolution: "UNRESOLVED".into(),
                        range: range(fun),
                    })
                }
            }
            "comment" => {
                // Go's scanner removes carriage returns from comment text.
                let x = text(n, source)?.replace('\r', "");
                if x.contains("//go:generate") || x.contains("Code generated") {
                    f.generated_markers.push(x)
                }
            }
            _ => {}
        }
        let mut c = n.walk();
        let children: Vec<_> = n.named_children(&mut c).collect();
        stack.extend(children.into_iter().rev());
    }
    f.declarations
        .sort_by(|a, b| (&a.range.start_byte, &a.name).cmp(&(&b.range.start_byte, &b.name)));
    f.imports
        .sort_by(|a, b| (&a.range.start_byte, &a.path).cmp(&(&b.range.start_byte, &b.path)));
    f.calls.sort_by(|a, b| {
        (&a.range.start_byte, &a.spelling).cmp(&(&b.range.start_byte, &b.spelling))
    });
    f.generated_markers.sort();
    f.body_sha256 = body(f.clone())?;
    if let Some(dir) = cache_dir {
        // Persistence is best-effort: filesystem contention, malformed cache
        // contents or a full cache cannot change the parsed facts result.
        if let Ok(value) = serde_json::to_value(&f) {
            if let Ok(raw) = canonical::encode(&value) {
                let _ = write_cache_entry(dir, &k, &raw);
            }
        }
    }
    Ok(f)
}

#[cfg(test)]
mod cache_storage_tests {
    use super::*;

    fn limits(entries: usize, bytes: u64, scan_entries: usize) -> CacheLimits {
        CacheLimits {
            entries,
            bytes,
            scan_entries,
        }
    }

    fn entry_path(dir: &Path, key: char) -> PathBuf {
        dir.join(format!("{}.json", key.to_string().repeat(64)))
    }

    fn write_entry(dir: &Path, key: char, bytes: usize, modified_seconds: u64) -> PathBuf {
        let path = entry_path(dir, key);
        fs::write(&path, vec![b'x'; bytes]).unwrap();
        let file = fs::OpenOptions::new().write(true).open(&path).unwrap();
        file.set_times(
            fs::FileTimes::new()
                .set_modified(SystemTime::UNIX_EPOCH + Duration::from_secs(modified_seconds)),
        )
        .unwrap();
        path
    }

    #[test]
    fn eviction_obeys_byte_and_entry_bounds_with_oldest_then_name_order() {
        let directory = tempfile::tempdir().unwrap();
        let oldest = write_entry(directory.path(), 'a', 6, 1);
        let newer = write_entry(directory.path(), 'b', 4, 2);
        let new_key = "c".repeat(64);
        write_cache_entry_with_limits(
            directory.path(),
            &new_key,
            b"12345",
            limits(2, GO_FACTS_CACHE_POLICY_BYTES.len() as u64 + 9, 8),
        )
        .unwrap();

        assert!(!oldest.exists(), "oldest cache entry was not evicted");
        assert!(
            newer.exists(),
            "newer cache entry was unnecessarily evicted"
        );
        assert!(directory.path().join(format!("{new_key}.json")).exists());
        let entries = scan_cache_entries(directory.path(), 8).unwrap();
        assert_eq!(entries.iter().filter(|entry| !entry.protected).count(), 2);
        assert_eq!(
            entries.iter().map(|entry| entry.bytes).sum::<u64>(),
            GO_FACTS_CACHE_POLICY_BYTES.len() as u64 + 9
        );
    }

    #[test]
    fn entry_limit_uses_filename_tie_break_for_equal_mtime() {
        let directory = tempfile::tempdir().unwrap();
        let first = write_entry(directory.path(), 'a', 6, 1);
        let second = write_entry(directory.path(), 'b', 4, 1);
        let new_key = "c".repeat(64);
        write_cache_entry_with_limits(directory.path(), &new_key, b"12345", limits(2, 128, 8))
            .unwrap();

        assert!(!first.exists(), "filename tie-break did not evict key a");
        assert!(second.exists(), "filename tie-break evicted key b");
        assert!(directory.path().join(format!("{new_key}.json")).exists());
        let entries = scan_cache_entries(directory.path(), 8).unwrap();
        assert_eq!(entries.iter().filter(|entry| !entry.protected).count(), 2);
    }

    #[test]
    fn cache_lock_contention_skips_storage_without_failing() {
        let directory = tempfile::tempdir().unwrap();
        fs::write(directory.path().join(GO_FACTS_CACHE_LOCK), []).unwrap();
        assert!(
            write_cache_entry_with_limits(
                directory.path(),
                &"a".repeat(64),
                b"{}",
                limits(2, 10, 8),
            )
            .is_ok()
        );
        assert!(!entry_path(directory.path(), 'a').exists());
    }

    #[test]
    fn malformed_cache_entry_names_fail_closed_for_writes() {
        let directory = tempfile::tempdir().unwrap();
        fs::write(directory.path().join("unexpected.json"), b"{}").unwrap();
        assert!(
            write_cache_entry_with_limits(
                directory.path(),
                &"a".repeat(64),
                b"{}",
                limits(2, 10, 8),
            )
            .is_err()
        );
        assert!(directory.path().join("unexpected.json").exists());
        assert!(!entry_path(directory.path(), 'a').exists());
    }

    #[test]
    fn evicted_fact_is_reparsed_with_the_same_body() {
        let directory = tempfile::tempdir().unwrap();
        let source = b"package p\nfunc F() {}\n";
        let producer = "a".repeat(64);
        let cold = go_file_facts(
            "a.go",
            source,
            &digest(source),
            &producer,
            Some(directory.path()),
        )
        .unwrap();
        assert_eq!(cold.parse_count, 1);
        let mut entries = scan_cache_entries(directory.path(), 8).unwrap();
        make_cache_room(&mut entries, "b".repeat(64).as_str(), 0, limits(1, 64, 8)).unwrap();
        assert!(
            !directory
                .path()
                .join(format!("{}.json", cold.cache_key))
                .exists()
        );

        let reparsed = go_file_facts(
            "a.go",
            source,
            &digest(source),
            &producer,
            Some(directory.path()),
        )
        .unwrap();
        assert_eq!(reparsed.parse_count, 1);
        assert_eq!(reparsed.body_sha256, cold.body_sha256);
    }

    #[cfg(unix)]
    #[test]
    fn cache_scan_rejects_symlink_entries() {
        use std::os::unix::fs::symlink;
        let directory = tempfile::tempdir().unwrap();
        let outside = directory.path().join("outside");
        fs::write(&outside, b"{}").unwrap();
        symlink(&outside, entry_path(directory.path(), 'a')).unwrap();
        assert!(scan_cache_entries(directory.path(), 8).is_err());
    }

    #[cfg(unix)]
    #[test]
    fn cache_creation_rejects_symlink_ancestors_before_creating_target() {
        use std::os::unix::fs::symlink;
        let directory = tempfile::tempdir().unwrap();
        let outside = directory.path().join("outside");
        let link = directory.path().join("cache-link");
        fs::create_dir(&outside).unwrap();
        symlink(&outside, &link).unwrap();
        let target = link.join("nested").join("cache");

        assert!(ensure_cache_policy(&target, GO_FACTS_CACHE_LIMITS).is_err());
        assert!(!outside.join("nested").exists());
        assert!(
            write_cache_entry_with_limits(&target, &"a".repeat(64), b"{}", GO_FACTS_CACHE_LIMITS)
                .is_err()
        );
        assert!(!outside.join("nested").exists());
    }
}
