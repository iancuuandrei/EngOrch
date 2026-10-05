# Explicit reviewer rechecks — 2026-10-05

Development increment: v1.1.11, based on
`b76310a050250e756e198281ceaa35c62798a2be`.

New repair-intelligence runs with configured structured review freeze a separate
immutable recheck policy. Existing runs retain their previous writer/reviewer
invocations. History derives from ordinary admitted reviews without new events.

## Focused executed observations

- PASS: explicit source-bound original concern in the real Git repair fixture;
  generic approval is rejected without a journal append, then explicit recheck
  produces accepted READY and `reviewer_recheck_closed` for the original ID.
- PASS: missing, foreign, duplicate, unresolved-approval and invalid/oversized
  rationale cases cannot satisfy the requested rechecks.
- PASS: unresolved feedback must use the original path and exact rationale in
  current ordinary findings, preserving the existing repair feedback route.
- PASS: bounded history omission and oversized optional context preserve the
  ordinary output contract without inventing concern closure.
- PASS: actual reviewer input exceeding its byte limit only because of optional
  rechecks falls back to the ordinary contract with mandatory plan data intact.
- PASS: both finite legacy/recheck schemas reach the actual Codex wire fixture;
  schema substitution is rejected before RPC. Codex RPC 0.466 s, runtime 0.340 s.

Initial recheck fixture suite passed in 47.123 s; expanded input-fallback fixture
passed in 23.040 s. These use actual local Git/native processes and deterministic
role runtimes, not live model judgment.

The final scoped regression run passed: control 69.852 s, CLI 1.831 s,
runtime 0.307 s and Codex RPC 0.333 s. Global API documentation inspection
initially failed on three existing undocumented methods; this increment adds
their comments. The subsequent documentation/link checks passed in 1.947 s,
with the separate history count/byte fallback regression passing in 0.510 s.
Independent review approved the behavior and authority boundaries.

Live repair benefit, token/cost savings, installed
package and final v2 qualification remain NOT RUN.

History represents admitted review occurrences, including unresolved feedback
echoes with their new review evidence IDs. It does not guess semantic duplicates.
Repeated unresolved rounds can therefore reach the explicit count/byte fallback.
