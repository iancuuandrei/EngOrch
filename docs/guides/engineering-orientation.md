# Engineering orientation queries

These read-only commands reuse Fabric's existing RI engine. They do not change
planner selection, write permissions or verification policy. Pin the RI executable
and its SHA-256; run from the configured, committed task repository.

## Rank observed Go files

Create an explicit corpus specification. For example, replace this path and
package/module metadata with declarations from your committed project:

```json
{"files":[{"path":"api.go","import_path":"example.test/project","module_path":"example.test/project"}],"generators":[]}
```

The corpus is partial and declaration-based; it does not establish active build
resolution. Save it as `corpus.json`, then create `rank-query.json`:

```json
{"objective":"improve ParseBytes handling","top_n":8}
```

```powershell
fabric ri rank C:\tools\engorch-ri.exe RI_SHA256 corpus.json rank-query.json
```

The bounded result binds the repository, source observations, graph, parser
producer and exact query. Ranking uses exact objective-token matches in
observed declarations/paths, then distinct local-package import degree, then
path order. It is advisory, with PARTIAL coverage. Its component identifiers
cover the selected files and observed coupling/generator closure; they are not
a global community partition or evidence that tasks are independent.

Queries support at most 32 results, a 2-KiB UTF-8 objective and 64 terms.
The result states truncation and omissions. It neither resolves calls nor proves
which tests execute. Optional cache arguments retain the graph command's
trusted-local, derived-fact semantics.

## Read semantic snapshot evidence

For an already admitted SCIP-backed snapshot, create a query for implementations:

```json
{"vocabulary":"implementations","node":"EXACT_NODE_ID","producer":"EXACT_PRODUCER_ID","direction":"INCOMING","limit":16}
```

```powershell
fabric ri semantic C:\tools\engorch-ri.exe RI_SHA256 snapshot.json SNAPSHOT_ID semantic-query.json
```

`INCOMING` lists declared implementations of the selected node; `OUTGOING`
lists what that node explicitly implements. Only observed `IMPLEMENTS` edges
are returned. The exact producer and direction retain their coverage and
absence declaration. UNKNOWN or PARTIAL evidence cannot prove an absence.

The same command supports direct references:

```json
{"vocabulary":"references","symbol":"EXACT_SCIP_SYMBOL","producer":"EXACT_PRODUCER_ID","limit":16}
```

Reference results retain PARTIAL coverage without absence inference. Existing
`fabric ri references` remains available. A returned `next_after` cursor can
be supplied as the final argument; its snapshot and query binding are checked.
Limits are 1–128. Query files must be regular strict JSON objects, at most
32 KiB, with no duplicate or unknown fields. The graph-only `ri query` adapter
continues to reject semantic references and implementations when their producer
evidence is unavailable.

## Index a TypeScript fixture with the explicit 0.4.0 profile (opt-in, scoped local PASS 2026-10-08)

The explicit `scip_typescript040` import policy admits only
`scip-typescript` version `0.4.0` with the pinned UTF-16 position policy
`engorch.scip-typescript.0.4.0.positions.v2:utf16-code-units+omit-invalid-synthetic-file-enclosing`;
generic strict
imports still reject an omitted position encoding. The pinned profile drops
only the advisory enclosing of the exact synthetic file-module marker (zero
`[0,0,0]` definition with matching file descriptor and `SymbolInformation`
provenance) when its well-formed enclosing does not contain the anchor; the
raw index hash is retained and all other enclosing checks are unchanged. The opt-in
`TestActualScipTypeScriptCLI` fixture commits a dependency-free `package.json`,
`tsconfig.json` and `library.ts` (LF, UTF-8) and runs the producer as
`node ENTRY index --output INDEX.scip --no-progress-bar` through the existing
`ri prepare-producer`, `produce`, `bind-import`, `import`, `publish` and
`runtime-binding` commands, with the project-root URI taken from Node's
`url.pathToFileURL` formatting. Manifest inputs bind the indexed source, the
committed configs, the entry script hash and the exact policy hash. Locate,
definition and reference queries check exact source bytes and provenance with
PARTIAL coverage and no absence claim. Scoped local PASS on uncommitted
candidate BASE `40d1ba5` (future v1.1.50; no publish or final SHA claimed):
pinned Rust 19 tests PASS, real-producer `TestActualScipTypeScriptCLI` PASS
with Node 22.23.3 / scip-typescript 0.4.0, plus Go RI/doc/vet PASS; see
[status](../evaluation/status.md). Initial local FAILs retained without
pooling; full v2, live benefit and acceptance/release remain NOT QUALIFIED,
review PENDING. See the [RI occurrences contract](../specifications/ri-occurrences.md)
for the admitted behavior.
