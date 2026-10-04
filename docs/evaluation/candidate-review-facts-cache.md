# Candidate review facts cache

`--review-impact-candidate-facts-cache` opts an autonomous run into version 1
local reuse of candidate Go syntax facts during review-impact collection. It
requires both `--review-impact-context` and a contract planner context
(`go-contract-context-v1/v2/v3`) with the same pinned RI executable.

The cache contains only per-file parser observations keyed by the RI runtime on
normalized path, exact source SHA-256, parser schema/version, and producer
SHA-256. It is private to one checkout and producer under the user cache
root. Different checkout roots do not share entries. The cache has no durable
record field and no model-visible prompt field.

Every review collection still captures and verifies the current candidate,
reads and hashes its current file bytes, recomputes candidate module ownership,
rebuilds the candidate graph and overlay, and validates generator retention.
A fact hit cannot reuse a candidate corpus, overlay, `READY` state, verifier
outcome, review decision, or external effect. Cache hit/miss diagnostics are
excluded from graph, record, prompt, and semantic identity.

The RI cache implementation used for this capability enforces at most 1,024
published entries and 64 MiB per cache leaf, with bounded scans and oldest-mtime
then filename eviction. Those storage guarantees require the cache-capable RI
binary; pinning an older RI binary does not silently acquire them. A missing,
corrupt, stale, or incompatible entry is a parser miss; it does not establish
that a candidate is unchanged or semantically valid.

Focused checks:

```powershell
& "$goDir\go.exe" test ./internal/ri -run 'TestCollectCandidateGoCorpusCandidateFactsCacheRevalidatesCandidateBindings|TestCandidateFactsCacheDoesNotReuseModuleOwnership' -count=1
& "$goDir\go.exe" test ./internal/control -run 'TestReviewImpactContext' -count=1
& "$goDir\go.exe" test ./internal/cli -run 'TestAutonomousReviewImpactContextRequiresAndBindsPinnedContractMode' -count=1
```

The focused cache checks were rerun against the final symlink-hardened RI
artifact at `D:\fabric-ci2-tools\ri-facts-cache-v13-final2-target\release\engorch-ri.exe`
with SHA-256 `b1894e16caf3b73b069722a5dd12829ce0f28fc4ded4e488b411b98f734d5603`.
Earlier receipts using `45a2bb47d34a00ae6fd7ccf76bc0fea8b4c176ff5c4c79cb525aa67449370d38`
remain attributable to that earlier artifact and are not evidence for the
final cache implementation.
