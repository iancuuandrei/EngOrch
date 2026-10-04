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
