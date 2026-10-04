# Imported repair spectra — 2026-10-05

Development increment: v1.1.16, based on
`fb1aa20188a3e16e564b5b94a30e0abed2d6cd52`.
This is a read-only advisory product capability, not authenticated per-test
receipt ingestion or live repair-benefit qualification.

## Executed evidence

| Check | Result and scope |
|---|---|
| Native individual Go profile fixture | PASS: an actual failing test and passing test produce compatible coverage; the defective branch ranks first with exact Ochiai squared score 1 |
| Exact ranking regressions | PASS: raw per-test counts, deterministic path/range ties and profile order independence; visit counts are not additional tests |
| Import regressions | PASS: UNKNOWN, duplicate test/block identities, numeric coordinate aliases, missing blocks, statement/mode mismatches, foreign/protected paths and encoding/bounds are rejected |
| Direct API allocation preflight | PASS: oversized profile/outcome/source fields reject before canonical serialization; escaped encoded-size overflow rejects |
| Candidate source observation | PASS: actual Git worktree read lease and complete source hashes; foreign run/candidate rejection; source hash/content/range/partial-file failures withhold ranked blocks |
| Drift and ownership | PASS: unadmitted candidate drift preserves base diagnosis; actual read-guard mutation causes close failure and withholds spectrum ranking |
| CLI lifecycle fixture | PASS: `diagnose --spectrum` imports a private artifact, observes exact candidate bytes, exports a byte-identical journal, and leaves ordinary diagnosis without a spectrum |

Final focused suite: faultlocalization 6.954 s, control 11.749 s, CLI 0.412 s,
exit 0. Scope: `TestSpectrum`, `TestRepairSpectrum`, `TestRepairAnchorClose`,
`TestDiagnoseArguments`, `TestReferenceIsCurrent`. The complete CLI lifecycle
fixture separately passed on the final implementation in 30.333 s (earlier
draft: 58.904 s). Documentation/link/ADR checks passed in 0.465 s. Vet on the
three affected packages and diff whitespace checks passed.
The generated CLI reference also escapes argument-alternative separators so
its Markdown table retains three columns. The reference regressions passed
in 0.156 s; final documentation checks passed in 0.377 s.
GitHub's Markdown API rendered the generated diagnosis row with three cells
and literal argument pipes, without displayed escape characters. This checks
the intended [GFM table rendering](https://github.github.com/gfm/#tables-extension-),
not merely the raw Markdown regression.

Independent review identified pre-serialization profile and outcome bounds
that the CLI reader alone could not guarantee for direct Go callers. Both
were corrected before publication and covered by explicit regressions.
Final independent GPT 6 Luna High review: APPROVE, including the source/guard
boundaries and actual GFM render receipt. Reviewer used static review and the
reported executed checks; no tests or providers were run by the reviewer.

The native profile fixture is an intentionally faulty tiny local Go program;
the controller/CLI source-binding fixture supplies explicitly untrusted spectra.
Neither is a real model repair or evidence of downstream quality/efficiency.
No provider was dispatched and no frozen v2 hidden material was accessed.

## Boundaries and remaining work

The importer validates bytes and range bounds, not that a test really ran or
that a source mapping represents its instrumentation. Test outcome assertions
cannot settle UNKNOWN, grant effects/ownership or close a finding. Original
findings/specifications remain intact and all existing final gates still apply.

Automatic coverage production/admission, fixer consumption, independent ranking
fusion, JEV acquisition selection and the required matched repair experiment
remain unfinished. No live model benefit, full-suite/source release, installed
package or v2 completion is established by this increment.

See the [guide](../guides/repair-localization.md),
[contract](../specifications/repair-diagnosis.md) and
[architectural decision](../adr/0008-repair-spectra.md).
