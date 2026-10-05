# Candidate-bound repair anchors — 2026-10-05

Development increment: v1.1.8, based on
`23a8101cf363c6aaf646abb6dead54fc69dcbde0`.

The [diagnosis contract](../specifications/repair-diagnosis.md) now supports
explicit leased current-file observations and unique complete-line relocation
of prior report hints. The default report remains byte-compatible and no model
invocation input or persisted controller schema changes.

## Executed focused evidence

- PASS: current file/range hashes and original/current candidate separation.
- PASS: unique exact line relocation and C0 → C1 → C2 report chaining preserve
  the original finding and its source candidate.
- PASS: repeated, missing, substring-only, foreign-path and incomplete-file
  evidence do not relocate. Altered and duplicate hint identities are rejected.
- PASS: empty, oversized, non-UTF-8 and incomplete lines, and invalid columns
  remain explicitly unavailable.
- PASS: strategy suggestion requires every finding's current anchor to lie
  inside the ready task's existing write paths; similar filename prefixes do
  not widen ownership.
- PASS: real Git candidate drift preserves the base diagnosis without verified
  anchors. Real read-lease guard mutation produces a close ownership failure;
  failed close cannot publish anchored evidence.
- PASS: existing replay-valid design → writer → fresh verification/review
  fixture observes a reviewer file while retaining a different exact repair
  write scope; no localized strategy is inferred for the unrelated path.
- PASS: CLI rejects foreign, oversized, header/candidate-substituted and unknown-field previous reports and
  invalid flags; a missing workspace preserves the base report without events.

Checks are targeted `go test` executions over `internal/control` and
`internal/cli`, with `go vet` on those packages. These observations establish
the bounded read-only interface. No live model, repair-efficiency, installed
package or final v2 qualification is implied. Strategy execution, deterministic
transformations and exact finding closure remain pending.

The review identified a missing final read-lease close check. It was corrected
and covered by the actual guard-mutation regression before publication.
Previous report hashes provide content integrity, not authenticated authority;
the new current-file observations validate bytes, not persistence of a defect.
Independent GPT 6 Luna High review: APPROVE after the close and chained-report
corrections and completed test/vet receipts; reviewer did not run tests.
