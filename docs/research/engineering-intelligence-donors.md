# Engineering-intelligence donor research

**Research date:** 2026-10-03  
**Fabric source inspected:** `cd7afb0f13f28e088e61c31d5679f0dfc310d8f3`  
**Method:** read-only, detached clones under `D:\fabric-donors`; no donor
binary, installer, indexer, or agent was executed. A donor's README or
benchmark is not product qualification.

## Fabric starting point

Fabric already has the evidence boundary that an adoption must preserve:

- `crates/ri` admits immutable, content-addressed snapshots bound to an exact
  repository source, producer declarations, input hashes, graph records and
  occurrences. Its finite graph, occurrence and location pages bind their
  continuation to the exact snapshot and query. Empty results do not establish
  absence; coverage stays a producer declaration.
- Rust lexical indexing builds independently verified base shards from a full
  manifest. A candidate overlay is explicitly bound to the candidate and may
  use at most 64 MiB of resident changed-source bytes. The stdio-stream process
  retains only one verified base index and re-verifies its manifest and shard
  bytes on every reuse.
- Runtime roles receive only a journal-selected, executable-hash-pinned RI
  binding. The existing agent-facing surface is `ri_status`, `ri_locate`,
  `ri_definition`, `ri_references`, and, when a lexical route is bound,
  `ri_search`. The control layer distinguishes base commit and candidate
  overlay context.

This leaves useful product gaps: a bounded, task-oriented *selection* over
already admitted source evidence; compact symbol/file summaries for a selected
area; and a language/producer coverage matrix that says what navigation is
available. It does not justify a mutable global index, relevance-derived
authority, or a live editor server in the v1 path.

## Donor findings

| Donor and pinned source | Directly observed mechanism | Decision | License and copying requirement |
| --- | --- | --- | --- |
| [SylphxAI/repomap `749c8de`](https://github.com/SylphxAI/repomap/tree/749c8de53e86853542d255a175314805d4d6bc99) | Rust core walks a working tree, parses tree-sitter languages, derives import/call/file graphs, PageRank and Louvain communities, and searches AST chunks with BM25; optional dense search exists. Per-file facts are cached at a path-derived cache root and considered fresh from mtime and size. | **Adapt mechanisms selectively, not the system.** A future bounded `map` response may measure centrality, community and hybrid ranking over Fabric's admitted snapshots. Do not import the donor cache, working-tree scan, MCP server, or mutable “current repo” identity. Each mechanism needs a bounded Fabric implementation and held-out measurement before adoption. | MIT. A copied substantial implementation needs the MIT notice and copyright. A clean-room implementation of bounded Fabric query data needs no code copy, but record the source inspiration. |
| [sourcegraph/scip `f07c097`](https://github.com/sourcegraph/scip/tree/f07c097d9b5d952c50eecde328200e454937fa09) | Protocol and bindings describe language-independent documents, occurrences, symbols, relationships, indexer metadata and document position encodings. Its Go canonicalizer shows deterministic document/symbol/relationship ordering. | **Already adopted protocol; no duplicate feature.** Fabric vendors a prior pinned `scip.proto`, generates Rust bindings, and adds its own source/producer admission. Track the new upstream revision only as a deliberate compatibility update with a corpus and producer test; do not treat protocol decoding as graph completeness. | Apache-2.0. Existing vendored protocol retains its license. A copied or updated upstream file must retain Apache notice/LICENSE and state modifications where applicable. |
| [oraios/serena `d0f7f92`](https://github.com/oraios/serena/tree/d0f7f92631c23dc4c5ed0b5ccd35bc623b19a809) | Symbol tools offer overview, name-path matching, body inclusion, kind filters, depth and character limits. They operate through project language servers. The repository changelog documents index waits, stale cache fixes, language-specific lifecycle and external server failures. | **Reject Serena application/runtime for v1.** Its useful product lesson is a small discoverable API: `symbols_overview(path, depth, limit)` and `find_symbol(name, path?, depth?, limit)` should be implemented only over Fabric's immutable admitted SCIP/structural snapshot. A live LSP server is a different mutable effect and cannot supply replayable evidence without a separate product need. | Serena application is GPL-3.0-or-later: do not copy application code. The `src/solidlsp` component is separately MIT according to its LICENSE; any future isolated reuse requires a file-level provenance and notice review. |
| [Aider-AI/aider `5dc9490`](https://github.com/Aider-AI/aider/tree/5dc9490bb35f9729ef2c95d00a19ccd30c26339c) | `RepoMap` extracts tree-sitter definition/reference tags, builds a NetworkX `MultiDiGraph`, uses personalized PageRank (chat files, mentioned paths/identifiers), renders selected lines, then binary-searches toward a token budget. Tag and map caches are keyed by filesystem path/mtime and the map may be reused according to refresh policy. | **Adapt one experiment first.** Evaluate task path/identifier seeds, centrality and hybrid ranking under a fixed byte/file budget against Fabric's selector. Do not adopt its filesystem state, mtime caches, unbounded tag walk, NetworkX runtime, approximate token counter, or cache refresh semantics. PageRank is a selection mechanism, never authority or a semantic result. | Apache-2.0. Retain license/NOTICE if copying code. Prefer a new Rust implementation after the experiment so no donor code is needed. |

## Concrete v2 sequence

1. **Snapshot map, before ranking.** Add a read-only query over an admitted
   snapshot that returns deterministic direct file/symbol summaries, producer
   IDs, declared coverage and explicit truncation. No new indexer, cache,
   controller state, or effect. This fills the Serena-style discoverability gap
   using existing RI facts.
2. **Selector experiment.** For a fixed, source-pinned task corpus, compare
   the existing bounded lexical/path selector with a deterministic
   path-and-identifier seeded structural/SCIP selector. Keep the same selected
   byte/file budget and record selected source IDs, coverage, latency and
   downstream outcome. Only add a graph-centrality variant if it improves the
   held-out task measure within the same bounds.
3. **Language coverage, not universal semantic claims.** Expand SCIP producer
   qualification one language at a time. Each producer must bind indexer
   binary/version/arguments, committed input hashes, source position policy and
   a real query fixture. Structural extraction remains an explicitly weaker
   producer where semantic output is absent.
4. **Candidate overlay ergonomics.** Keep lexical overlays candidate-bound.
   If symbol-level candidate queries are requested, design a distinct,
   candidate-bound producer/result protocol; never silently merge working-tree
   data into a base snapshot.

## Executable test ideas

- **Map determinism:** construct one admitted SCIP snapshot with two producers
  and an incomplete structural declaration; permute input order and assert
  identical canonical map bytes, bounded page/cursor behavior and no
  `absence_proven` field becoming true.
- **Map authority:** reject a runtime map request if the snapshot ID,
  executable hash, producer, cursor, or candidate overlay identity differs
  from the recorded binding. Verify an empty symbol page reports declared
  coverage rather than absence.
- **Selector A/B:** run the identical fixed task corpus twice with a source
  snapshot fixed by commit and manifest. Assert each arm's selected file and
  byte ceilings, preserve the selection manifest, and report success, native
  verification, review, elapsed time and typed token counters separately.
  Reject the decision if the treatment changes source identity or has only
  synthetic/fixture improvement.
- **Cache rejection:** mutate a same-mtime source file and show that no
  commit/source-hash-bound Fabric query can reuse it. This is the specific
  incompatibility with the donors' mtime/path cache model.
- **Producer coverage:** for each claimed language, feed a false-absence
  fixture and require either an observed occurrence or an explicit incomplete
  coverage declaration. A producer failure leaves the result unavailable or
  UNKNOWN; it must not be converted to an empty semantic answer.

## Evidence limits

The upstream mechanisms above are source observations at the pinned revisions.
Their repository benchmarks, supported-language lists, and product statements
were not executed here and are **reported claims, not Fabric measurements**.
Fabric's existing Rust unit/integration tests and previous installed-RI
receipts establish only the scopes stated in their own evidence; they do not
measure donor ranking quality, cold/warm index cost, broad language coverage,
or end-user coding-task improvement.


## Go module ownership follow-up

The committed corpus currently emits source-local package identities. Those
identities do not resolve import paths. Automatic module ownership must read
the complete bounded inventory of committed `go.mod` files, choose the deepest
containing module and account for nested modules before deriving package paths.
The [Go module reference](https://go.dev/ref/mod#modules) defines module roots
and package paths; the official [modfile parser](https://pkg.go.dev/golang.org/x/mod/modfile)
is the candidate parser rather than a guessed text regular expression. Workspaces,
replacements, vendor metadata and build constraints require explicit provenance;
an observed path match alone does not establish active dependency resolution or
Go type/call resolution. This is an outstanding v1.1 requirement, not a claim
that the existing source-local graph already implements it.
