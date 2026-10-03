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

`fabric ri modules` emits a bounded `GoModuleInventory` for the configured
repository's exact committed tree. It records parsed declarations and safe
omissions for committed `go.mod`, `go.work`, and `vendor/modules.txt` files,
including the repository ID, commit, tree and inventory digest. It does not
read dirty manifest contents, run Go, select a workspace or vendor mode, or
resolve dependencies. `coverage: partial` and per-file omission statuses must
remain visible; a partial inventory does not establish that an unobserved
module declaration is absent.

The output can be embedded as `module_inventory` in a graph spec. The inventory
is bound to its source ID, commit and tree; a spec from another committed
source is rejected. When supplied, graph package identities are derived from
the committed module declarations and each Go file's package clause. Explicit
`import_path` and `module_path` fields may be omitted in that mode; if supplied,
they must agree with the declaration. The resulting graph is still a committed
base graph with an empty candidate ID. Candidate overlays require their own
manifest closure and are not created by this command.

For example, PowerShell can embed the exact command output without manually
copying its source identity or digest:

```powershell
$inventoryJson = (& fabric --root D:\src\project ri modules) -join "`n"
[System.IO.File]::WriteAllText('modules.json', $inventoryJson, [System.Text.UTF8Encoding]::new($false))
$modules = Get-Content modules.json -Raw | ConvertFrom-Json
$spec = @{
  files = @(@{ path = 'internal\pkg\api.go' })
  generators = @()
  module_inventory = $modules
}
[System.IO.File]::WriteAllText('graph-spec.json', ($spec | ConvertTo-Json -Depth 32), [System.Text.UTF8Encoding]::new($false))
fabric --root D:\src\project ri graph D:\tools\engorch-ri.exe <exe-sha256> graph-spec.json
```

`fabric ri graph EXE EXE_SHA256 SPEC_JSON [CACHE_DIR]` reads every listed
regular Go file from the same committed repository snapshot, obtains facts
through one pinned RI stream, and builds a partial graph. The required spec is
strict JSON with this shape. Whitespace and object key order are flexible; duplicate
or unknown members, invalid UTF-8 and malformed values are rejected. The strict
graph spec, including an embedded module inventory, is limited to 1 MiB:

```json
{"files":[{"import_path":"example.invalid/app/pkg","module_path":"example.invalid/app","path":"pkg/api.go"}],"generators":[]}
```

Without `module_inventory`, each file needs an explicit import and module path.
External test packages may also set `test_of_import_path`. Optional generator records contain
`generator_path`, `generated_path`, and the exact source `directive`; both
files and the directive must be present in the bounded corpus. The CLI never
runs generator commands. Specs reject duplicate JSON members, unlisted or
sensitive paths, and invalid package identities before parsing begins. The
corpus is limited to 256 files, 1 MiB per file, and 8 MiB total. The result
includes repository identity, per-file Git blob/content observations, and the
graph with `coverage: PARTIAL` and an empty candidate ID for this committed
base. With `module_inventory`, the import and module paths may be omitted and
are instead derived from the package clause plus the unique deepest committed
module declaration.

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

## Bounded graph query

`fabric ri query EXE EXE_SHA256 SPEC_JSON QUERY_JSON [CACHE_DIR]` builds the
same committed graph, then runs one strict JSON query against its observed
facts. `QUERY_JSON` must be a regular UTF-8 JSON file no larger than 32 KiB;
duplicate and unknown members are rejected. Every query requires an integer
`limit` from 1 to 1000. The result wraps the repository identity and source
observations with a `result` bound to the source ID, graph digest, and pinned
parser digest. Coverage remains `PARTIAL`, even for an empty result.

Supported `vocabulary` values are `symbol`, `imports`, `calls`, `tests`,
`generators`, `module`, `path`, and `impact`. A query uses the matching selector
fields: `path` for one exact Go file, `name_prefix` and/or `kind` for symbols,
`import_path` for an exact relation target, or `paths` plus optional
`max_depth` for impact. Impact accepts at most 64 exact paths and depth at most
16. Calls always carry `resolution: UNRESOLVED`; module results represent
declared ownership only. `references` and `implementations` are unsupported by
this graph-only query. Use the existing SCIP-backed `ri definition` and
`ri references` commands when querying a semantic snapshot.

Example query file and invocation:

```json
{"vocabulary":"calls","path":"internal/cli/ri_graph.go","limit":50}
```

```powershell
fabric --root D:\src\project ri query D:\tools\engorch-ri.exe <exe-sha256> go-graph-spec.json semantic-query.json
```

This command is read-only and deterministic for the committed source, graph
spec, parser binary, and query. Partial graph input or omitted items do not
prove that declarations, references, callers, or impacts are absent.

## Candidate graph query

`fabric ri candidate-query RUN EXE EXE_SHA256 BASE_SPEC_JSON QUERY_JSON [CACHE_DIR]`
applies a bounded Go-source overlay from one exact confirmed controller run
candidate to a committed base graph, then runs the same semantic query. `RUN`
must name a replay-valid journal whose confirmed workspace and candidate are
bound to the currently configured repository identity. A stale source, missing
candidate, uncertain workspace/file outcome, changed candidate, or candidate
from another source is rejected. This command does not advance the run or
change its journal.

`BASE_SPEC_JSON` has the `ri graph` format and must include a committed v1
`module_inventory`. The inventory is revalidated against the configured Git
source. Candidate module inventory v2 is collected separately from the exact
candidate and bound to that base inventory and candidate ID. The candidate
collector compares the candidate with the bounded base graph, parses at most
eight eligible changed/new Go files (up to 1 MiB each), applies deletions, and
retains only unchanged declared generator links. `changed_path_count` covers
changed paths from the base graph; `admitted_path_count` covers successfully
parsed candidate Go paths outside that base graph. Omissions appear in the
bounded report. A changed base file omitted from parsing is removed from the
candidate graph, so old declarations cannot appear as current candidate facts.
If no admitted Go facts remain, the query rejects without a result instead of
returning an old graph or claiming complete absence.
New or changed generator directives are not inferred. The
candidate graph uses the pinned parser digest from `EXE_SHA256`; `CACHE_DIR`, if
supplied, applies only to committed base parsing, while candidate parsing is
uncached.

The output identifies the run, repository, candidate, base and candidate graph
digests, both module-inventory digests, candidate-manifest coverage and
omissions, changed/admitted/deleted counts, bounded safe Go-source omissions
and the semantic `result`. It does not include source bytes or the
full candidate graph. Coverage remains `PARTIAL`; a bounded omission, absent
item, unresolved call, or unsupported semantic vocabulary is not proof of
absence. `references` and `implementations` remain unsupported by this
graph-only query. The command is read-only and makes no provider calls.

Example:

```powershell
fabric --root D:\src\project ri candidate-query <run-id> D:\tools\engorch-ri.exe <exe-sha256> go-graph-spec.json candidate-query.json
```

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
