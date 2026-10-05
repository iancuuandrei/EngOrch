# Structured fixer context — 2026-10-05

Development increment: v1.1.9, based on
`ba3539e38112e274eaffac691044b4587e5ea985`.

New autonomous runs can opt into `--repair-intelligence`. The integration uses
existing admitted task-context records and failed gate evidence to supply a
bounded structured repair object to the writer. The immutable opt-in is absent
from legacy creations. No new journal event or workspace observation is added.

## Executed evidence

- PASS: complete recorded preimages yield candidate-bound anchors; partial,
  altered, hash-mismatched and foreign-candidate contexts do not validate them.
- PASS: relocation uses the same source and original candidate admission;
  foreign-source context does not relocate stale findings.
- PASS: diagnostic journal-head changes do not change derived prompt data.
  Compact anchor identity matches the actual record with code omitted.
- PASS: a 128-finding oversized projection becomes a bounded omission record;
  changed omitted evidence changes its retained full-projection hash.
- PASS: CLI stores version 1 only with the explicit new-run flag. Unsupported
  policy versions and missing graph/context/repair prerequisites are refused.

- PASS: the actual writer drops optional intelligence when it alone takes input
  over 256 KiB; the ordinary plan remains byte-for-byte intact in decoded input.
- PASS: replay-valid real Git design → writer → fresh native checks/review
  fixture reaches READY. Review uses a deterministic provider fixture.

Final targeted suite: control 47.186 s, CLI 1.641 s, documentation 0.490 s,
exit 0. Scope includes repair intelligence, diagnosis/anchors, the complete
repair fixture, legacy policy identity, generated CLI reference and local links.
`go vet` on control/CLI and `git diff --check` also passed. No live model
acceptance run is implied.

Independent GPT 6 Luna High review: APPROVE for the scoped integration, based
on static review and reported executed checks; reviewer did not run tests.
After helper extraction, the focused intelligence/policy cases passed again
in 0.156 s. Live repair quality, cached/input/
output token benefit, installed package acceptance and final v2 qualification
remain NOT RUN. Deterministic transforms, automatic strategy routing and exact
historical finding closure remain pending.
