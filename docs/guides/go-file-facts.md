# Go file facts

`fabric ri facts EXE EXE_SHA256 PATH [CACHE_DIR]` returns bounded syntax facts
for one regular file at the repository's currently observed commit. The command
reads Git's committed blob; dirty worktree bytes are ignored. `PATH` is a
repository-relative path. Files must be valid UTF-8 and no larger than 1 MiB.

The result includes the repository ID, commit, Git blob ID, byte count and
content SHA-256, plus facts bound to the exact path, content digest and pinned
RI executable digest. The parser can report declarations, import specs,
syntactic calls and generated-code markers. Call resolution is always
`UNRESOLVED`, and coverage is always `PARTIAL`; this is not type checking,
dependency resolution, semantic indexing, or evidence that generated code was
executed.

The optional cache directory is an explicit local derived-data location and
must be an absolute clean path writable only by a trusted user/process. The cache key binds the path,
source digest, parser and producer digest; the body digest detects accidental
corruption but is not a signature or proof of semantic derivation. Treat
cached facts as advisory and never as authorization or verification evidence.
A cache hit does not make coverage complete. Without `CACHE_DIR`, parsing is
uncached and reports `cache: "miss"` and `parse_count: 1`.

Example:

```powershell
fabric --root D:\src\project ri facts D:\tools\engorch-ri.exe <exe-sha256> internal\pkg\file.go D:\cache\engorch\go-facts
```

## Explicit Go graph and task context

`fabric ri graph EXE EXE_SHA256 SPEC_JSON [CACHE_DIR]` reads every listed
regular Go file from the same committed repository snapshot, obtains facts
through one pinned RI stream, and builds a partial graph. The required spec is
strict JSON with this shape. Whitespace and object key order are flexible; duplicate
or unknown members, invalid UTF-8 and malformed values are rejected:

```json
{"files":[{"import_path":"example.invalid/app/pkg","module_path":"example.invalid/app","path":"pkg/api.go"}],"generators":[]}
```

Each file needs an explicit import and module path. External test packages may
also set `test_of_import_path`. Optional generator records contain
`generator_path`, `generated_path`, and the exact source `directive`; both
files and the directive must be present in the bounded corpus. The CLI never
runs generator commands. Specs reject duplicate JSON members, unlisted or
sensitive paths, and invalid package identities before parsing begins. The
corpus is limited to 256 files, 1 MiB per file, and 8 MiB total. The result
includes repository identity, per-file Git blob/content observations, and the
graph with `coverage: PARTIAL` and an empty candidate ID for this committed
base.

`fabric ri context EXE EXE_SHA256 SPEC_JSON OBJECTIVE [CACHE_DIR]` uses the
same committed corpus and graph builder, then runs the existing bounded task
context selector. Its result includes the same repository/source observations
and a context manifest bound to the graph digest. `ri context` has no candidate
ID and therefore accepts no changed-file claim; it uses objective-matched
symbol/path hints from the explicitly listed corpus. Missing facts or omitted
context never prove that a file or dependency is absent. The default task
context selector limits the prompt view to 12 files and 48 KiB.

`CACHE_DIR`, when supplied to any of these commands, has the same trusted-local
derived-cache meaning as above. The cache remains outside graph identity and
does not turn partial syntax into semantic resolution.

## Advisory topology query

`fabric ri topology EXE EXE_SHA256 SPEC_JSON CHANGED_PATHS_JSON MAX_GROUP_FILES [CACHE_DIR]`
reuses the same explicit committed corpus and graph builder, then summarizes
observed package import degrees, graph-reachable impact, explicit generator
coupling, potential test files and bounded review groups. `CHANGED_PATHS_JSON`
is a strict JSON array of 1 to 64 unique eligible repository-relative `.go`
paths; the entries must occur in `SPEC_JSON`. The file must be regular and no
larger than 32 KiB. `MAX_GROUP_FILES` is an integer from 1 to 32. The result
includes repository/source digests and the topology's graph and result digests.

The query is deterministic and read-only. Package degrees count only exact
package bindings already supplied in the graph; unresolved source-local imports
remain unresolved. Review groups preserve component IDs when a large component
must be split. These groups and potential tests are advisory only: they do not
change scheduler policy, create a candidate, authorize edits, or count as
verification evidence. Partial coverage, traversal limits, or omitted graph
inputs must not be interpreted as proof that no other files are affected.

Example:

```powershell
$utf8 = [System.Text.UTF8Encoding]::new($false)
[System.IO.File]::WriteAllText('changed-paths.json', '["internal/ri/engineering_graph.go"]', $utf8)
fabric --root D:\src\project ri topology D:\tools\engorch-ri.exe <exe-sha256> go-graph-spec.json changed-paths.json 8 D:\cache\engorch\go-facts
```
