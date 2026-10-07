# Go planner PPR context

Normative treatment rules for the opt-in bounded Personalized PageRank file
ordering over an already admitted Go engineering graph. The controller remains
authoritative; ranks grant no authority and never acquire new evidence.

## Treatment identity

- Version 1 is the single admitted treatment. It requires immutable
  `ExecutionPolicy.PlannerContext` `go-source-context-v2`,
  `ExecutionPolicy.PlannerPPRVersion` `1` (`omitempty`) and the pinned Go RI
  parser binding already required by that context.
- A nil treatment preserves every earlier record, manifest, selection and hash
  representation byte-for-byte. There is no default promotion: unknown flag
  combinations and unknown versions are rejected pre-run with no run created.
- Parser-unavailable fallback clears the opt-in along with the context that
  requires the parser.

## Fixed parameters

Only the documented constants are admitted; any deviation is rejected, so
there is no tuning surface:

| Parameter | Value | Meaning |
| --- | --- | --- |
| `alpha` | 17/20 | Lazy-walk diffusion factor (restart 3/20) |
| `scale` | 1000000 | Fixed-point probability mass unit |
| `max_iterations` | 50 | Hard diffusion bound |
| `convergence_epsilon` | scale/1000 | L1 early-stop threshold in mass units |
| `max_work_units` | 20000000 | `iterations * (files + projected edges)` budget, evaluated after adjacency allocation |
| `max_projection_files` | 128 | Eligible projection file cap, checked before any adjacency allocation |
| `max_ranks` | 64 | Stored file ranks in durable provenance |
| `clique_limit` | 32 | Package co-membership clique bound (larger groups project sparsely) |

The walk shape follows the lazy undirected walk with a personalized restart
vector (principles only from Andersen/Chung/Lang local partitioning and the
Brin/Page damping factor; no donor code, no imported guarantees). The graph is
explicitly PARTIAL with UNRESOLVED calls, so a rank is advisory file ordering
only.

## Seeds

- Seeds derive from observed inputs only: changed paths (each present in the
  graph), full-identifier query matches (case-sensitive exact declaration
  labels, never substring or prefix) and exact path cues (the objective
  contains a graph file's exact path string).
- Planner admission compiles with no changed paths, so stored seeds and the
  truncation flag MUST exactly equal `DeriveGoPPRSeeds` of the bound graph,
  bound query and nil changed paths. A self-consistent recomputation from a
  different valid seed set is still a substitution at admission and MUST be
  rejected. Seeds are observed inputs, not free knobs attached to the query
  hash.
- Empty seeds are not an error: they yield explicit `empty_seeds` no-signal
  provenance and the current selector runs unchanged.

## Projection

- The walk runs on a deterministic file-level UNDIRECTED adjacency built only
  from admitted relations. Direction is not otherwise modeled and no directed
  call resolution is performed.
- Projected relations: `IN_PACKAGE`/`DECLARED_OWNERSHIP` co-membership
  (clique up to the clique limit), `IMPORTS` (importer adjacent to every
  eligible local file of the imported package), `TESTS` (test file adjacent to
  other eligible files of the tested package) and `GENERATED_BY` (generated
  file adjacent to its generator file, explicit source-bound relations only).
- `CALLS_UNRESOLVED` containment and `CALL_NAME_CANDIDATE` spelling hints
  MUST NOT create cross-file edges. Only task-context-eligible paths
  participate; sensitive paths are omitted without any absence claim.
- Package groups above the clique limit project NO intra-package edges (sparse
  projection, no synthetic hub, no fabricated equivalence).
- The projection admits at most `max_projection_files` (128) eligible files,
  checked before any neighbor map is built. `IMPORTS`/`TESTS` expansion links
  one importer to every eligible file of the imported package, so without this
  cap a large admitted graph (up to 4096 files/300k edges) could allocate
  millions of neighbor entries before the post-hoc work budget applies. Graphs
  above the cap return explicit `graph_size_exhausted` no-signal provenance
  with a distinct size-exhausted projection hash (never a truncated fabricated
  projection) instead of building adjacency. Planner corpus admission caps at
  24 files, so planner projections never reach either the file cap or the
  clique bound; the file cap guards only direct primitive use on larger
  graphs. Invalid graphs still reject and never return success.

## Walk and scores

- Iteration: `pi_{k+1} = (1-alpha)*s + alpha*P^T*pi_k` with fixed-point
  integer arithmetic: floor plus deterministic remainder distribution by exact
  path order, so total mass is conserved exactly (scores sum to scale) every
  iteration.
- Stored scores are finite fixed-point approximations at the reported
  iteration count. No per-score error bound against the exact stationary
  distribution is claimed. `Converged` reports that L1 movement fell at or
  below the finite epsilon threshold when the walk stopped; it is a stopping
  outcome, not a proof of exact stationarity.
- Work exhaustion degrades before admission with explicit `work_exhausted`
  provenance instead of aborting a valid run. Oversized graphs degrade
  separately with explicit `graph_size_exhausted` provenance before adjacency
  allocation; large-corpora exhaustion is not claimed from the walk budget
  alone.

## Selection integration

- Only positive-score ranks become an explicit advisory ordering for the SAME
  bounded excerpt selector, file/byte limits and symbol/span narrowing. The
  ordering breaks legacy score ties before the historical path tie-break; it
  is hashed into the task-context input identity when present.
- Zero-score and unreached paths MUST NOT become positive evidence: score-0
  files stay omitted as `unranked`. Excerpt pivots, reasons and byte costs
  follow the existing safe logic; the treatment changes which tied file is
  visible, never the excerpt mechanics.
- The ordering is incompatible with the experimental `rrf-coverage-v1`
  selector; mixing is rejected. Rank provenance for the treatment lives in
  the manifest-level PPR provenance, not in per-file selector fields.

## Provenance and record binding

- Provenance binds graph digest, query hash, seeds, truncation, fixed
  parameters, iteration outcome, projection hash and rank hash. Validation
  recomputes from the bound inputs and rejects any stale, foreign or tampered
  copy without silent downgrade.
- The v3 planner record carries the treatment in BOTH `PPR` and
  `Context.PPR`, always deeply equal. Nested-only, record-only and
  treatment-mismatch shapes MUST be rejected, as MUST any treatment under a
  legacy (opt-out) policy or any legacy record under an opt-in policy.
- Unavailable records carry no treatment. Manifest digests cover the
  treatment; the model-visible prompt projects the full manifest including
  its rank provenance.
- The manifest output budget (128 KiB) and journal record budget apply
  unchanged.

## Authority and fallback

- Ranks are advisory file ordering only. They cannot acquire evidence, permit
  writes or effects, turn missing/PARTIAL/UNRESOLVED observations into
  absence, or resolve calls.
- No-signal, work-exhausted and graph-size-exhausted outcomes keep the current
  selector with explicit provenance (`FallbackReason`, `NoSignal`,
  `WorkExhausted`, `HintsTruncated` when seeds truncate). `FallbackReason` is
  `empty_seeds`, `work_exhausted` or `graph_size_exhausted`; only the size
  reason reports the pre-adjacency file cap.
