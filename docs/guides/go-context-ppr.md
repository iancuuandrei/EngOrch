# Bounded PPR planner context

Opt `go-source-context-v2` planning into bounded Personalized PageRank file
hints over the already admitted engineering graph. The treatment reorders
which tied files become visible; it adds no evidence, authority or default
behavior change. See the [PPR specification](../specifications/go-context-ppr.md)
for the normative contract.

## When to use it

Use `--planner-ppr` when several planner-visible files tie on existing
selection scores and graph topology (imports, tests, package co-membership,
generation) is a relevant tie-break for the objective. It is advisory only:
lexical and path evidence still decide admission, and excerpt bytes,
symbol/span visibility and all file/byte bounds are unchanged.

Do not expect measured quality or latency gains: none are claimed. Later
optional-capability source changes require their own verification.

## Usage

```sh
fabric run --autonomous \
  --planner-context go-source-context-v2 \
  --planner-context-ri-executable PATH \
  --planner-context-ri-executable-sha256 SHA256 \
  --planner-ppr \
  --prepare-only --max-parallel 1 \
  "Correct SplitLines semantics and document the example contract"
```

Requirements:

- `--planner-context go-source-context-v2` with its absolute pinned RI parser
  path and lowercase SHA-256. Any other context with `--planner-ppr` is
  rejected pre-run and creates no run.
- `--inspect-plan` reports the effective binding as `"planner_ppr":1`
  (legacy default `0`). Inspecting never creates a run or dispatches.
- If the parser is unavailable, capability fallback clears the opt-in along
  with the parser-bound context before any dispatch.

## Limits

- Fixed treatment only: damping 17/20, restart 3/20, integer mass scale
  1000000, at most 50 diffusion iterations, L1 stop threshold 0.1% of mass,
  20M work-unit budget (checked after adjacency allocation), at most 128
  eligible projection files (checked before any adjacency allocation), at most
  64 stored ranks. No parameters are tunable.
- Planner corpus admission caps at 24 files; stored ranks beyond the corpus
  cannot occur through the planner. Package groups above 32 files project
  sparsely (no intra-package edges) and never affect planner runs; planner
  runs never reach the 128-file projection cap either.
- Manifest output stays within 128 KiB; the journal record within its bound.
  Graphs above the projection file cap degrade to the current selector with
  explicit `graph_size_exhausted` no-signal provenance (distinct
  size-exhausted projection hash, zero ranks, no truncated fabricated
  projection), never to silent substitution. Walk-budget exhaustion degrades
  separately as `work_exhausted`; large-corpora exhaustion is not claimed from
  the walk budget alone.

## Advisory authority and fallback

Ranks order already admitted files; they cannot acquire evidence, permit
writes or effects, claim absence from missing/PARTIAL/UNRESOLVED
observations, or resolve calls. Missing seeds (`empty_seeds`), an exhausted
walk budget (`work_exhausted`) or an oversized projection
(`graph_size_exhausted`) keep the current selector with explicit no-signal
provenance instead of aborting the run.

## Source bindings

Seeds are observed inputs: changed paths present in the graph, exact
case-sensitive identifier matches and exact path cues from the objective.
Planner admission re-derives them from the bound graph and query and rejects
any substitution, including a honestly recomputed ranking from different
seeds.

## Math parameters and approximation

Scores are finite fixed-point approximations at the reported iteration count
with exact per-iteration mass conservation (scores sum to scale). No
per-score error bound against the exact stationary distribution is claimed;
the converged flag reports a finite L1 threshold outcome, not a stationary
proof. The single-step analytical reference and mass assertions live in the
focused `TestPPR*` suite.

## Provenance

The admitted record carries fixed parameters, exact seeds, iteration outcome,
projection hash and rank hash in both the record-level and manifest-level
PPR copies (always equal). Replay recomputes the walk from the bound inputs
and rejects nested-only, record-only, mismatched or forged copies. The
model-visible prompt includes the full rank provenance.
