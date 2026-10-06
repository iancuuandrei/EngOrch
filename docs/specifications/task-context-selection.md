# Task context selection

Normative selection and binding rules for bounded role task context. The
controller remains authoritative; pure selectors grant no authority and never
acquire new evidence on their own.

## Selector modes

- Empty preserves exact historical/default weighted selection byte-for-byte,
  including `InputHash`, `ManifestID` and wire JSON.
- `rrf-coverage-v1` is the single experimental mode. It requires immutable
  `ExecutionPolicy.Context` `bounded-v1` and immutable optional
  `ExecutionPolicy.ContextSelector` `rrf-coverage-v1` (`omitempty`).
- Unknown selector modes are rejected pre-run and pre-provider; a malformed
  `--context-selector` flag creates no run. There is no default promotion.

## Independent rankers and RRF

`rrf-coverage-v1` fuses only actually admitted ordinal sources with equal
unweighted Reciprocal Rank Fusion:

```text
RRF(d) = sum_r 1 / (60 + rank_r(d))
```

Ranks are 1-based per ranker; absent means no signal and that ranker is
skipped for that file. Rankers are `changed`, `hint`, `path_token`,
`content_lexical` and `anchor` (anchor only when present). No SCIP, PPR, SBFL
or type relations are claimed.

`k=60` is fixed provenance from Cormack, Clarke and Buettcher, SIGIR 2009
(`https://cormack.uwaterloo.ca/cormacksigir09-rrf.pdf`, paper lines 39-40),
not tuned on any Fabric benchmark. The implementation is independent under the
existing Fabric license; no donor code is ported. Scores are exact rationals
internally; no floats appear in canonical JSON. Ties break by path so
equivalent input permutations preserve identity.

Read-corpus prioritization uses the same equal RRF over writer, exploration
(including evidence), lexical and path-token ordinals, with no file reads.

## Bounded coverage

After RRF ordering, the mode selects a complementary excerpt set by marginal
binary coverage per actual returned UTF-8 excerpt byte cost (`gain/cost`),
with RRF as documented secondary ordering and path as final tie-break. This
is the binary specialization of `sum_c[1-product_(i in S)(1-a_ic)]` with
diminishing returns and no redundancy multiplier.

Concepts are evidence-bound and finite: admitted objective terms plus explicit
admitted path/anchor cues (at most 128 concepts). Matches apply to visible
excerpts only, never hidden full-file text. Excerpt pivots reuse the safe
existing bounded rare-match logic. Limits are unchanged: at most 24 reads,
32768 bytes per file, 768 KiB input, and at most 12 files / 48 KiB returned
(8 KiB per file) or lower caller bounds. Empty, non-UTF-8, sensitive and
oversized observations are omitted safely with `sensitive_path` redacted.
No-signal inputs use explicit `deterministic_fallback`; remaining files report
`redundant`, `unranked` or budget reasons.

## Provenance and binding

`Manifest.Selector` records the producing mode. Each selected item may carry
`ranks` (ranker to 1-based ordinal, at most 5 entries) and `covered` (newly
covered binary concepts, at most 16 entries of 2-64 UTF-8 bytes each).
Admitted cues longer than 64 bytes remain valid RRF rank signals but are never
truncated or fabricated into coverage provenance; they are skipped as
concepts. Legacy empty-selector manifests carry neither, preserving exact
hashes. All new fields are `omitempty` under the unchanged
`harness.task-context-manifest.v1` and `harness.task-context-input.v1`
domains.

Admission binds the same source, candidate, exact query and read closure as
legacy; replay rejects selector substitution between run policy and manifest.
Pure intelligence failure falls back to legacy behavior only before admission
with explicit fallback/omission evidence; invalid authority, bindings or
policy never silently downgrade.
