#![allow(missing_docs)]
//! Bounded, content-addressed Go source facts. These are syntax observations,
//! not an index, a semantic resolver, or an authority to run generated code.

use crate::canonical;
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use std::{
    fs,
    io::Read,
    path::Path,
    time::{Duration, Instant, SystemTime},
};
use tree_sitter::{Node, ParseOptions, Parser};

pub const GO_FACT_SCHEMA: &str = "engorch.go-file-facts.v1";

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
        if let Ok(meta) = fs::symlink_metadata(p) {
            if cache_reparse(&meta) {
                return Err("cache directory symlink rejected".into());
            }
            if !meta.is_dir() {
                return Err("cache ancestor is not a directory".into());
            }
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
        validate_cache_dir(dir)?;
        fs::create_dir_all(dir).map_err(|e| e.to_string())?;
        let target = dir.join(format!("{k}.json"));
        let entries = fs::read_dir(dir)
            .map_err(|e| e.to_string())?
            .take(1025)
            .count();
        if target.exists() || entries < 1024 {
            let raw = canonical::encode(&serde_json::to_value(&f).map_err(|e| e.to_string())?)?;
            let nonce = SystemTime::now()
                .duration_since(SystemTime::UNIX_EPOCH)
                .map_err(|e| e.to_string())?
                .as_nanos();
            let tmp = dir.join(format!(".{k}.{}.{}.tmp", std::process::id(), nonce));
            let mut out = fs::OpenOptions::new()
                .write(true)
                .create_new(true)
                .open(&tmp)
                .map_err(|e| e.to_string())?;
            std::io::Write::write_all(&mut out, &raw).map_err(|e| e.to_string())?;
            drop(out);
            fs::rename(tmp, target).map_err(|e| e.to_string())?;
        }
    }
    Ok(f)
}
