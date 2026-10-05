# Historical native recheck evidence — 2026-10-05

Development increment: v1.1.10, based on
`d21b1b747196cc05a6fcb1f110fe61d29c78c41a`.

The read-only `diagnose RUN --closure` command projects historical findings
and later exact native check receipts from one fully validated journal prefix.
It performs no provider or workspace effects and adds no journal schema.

## Executed focused observations

- PASS: original complete-state finding IDs and evidence hashes are preserved.
- PASS: all invocation inputs except candidate identity participate in the
  stable native oracle definition, including arguments, limits, executable
  path/hash and complete selected environment.
- PASS: incomplete final acceptance, missing review, pending verification,
  unstarted receipts and source observation errors cannot close findings.
- PASS: unavailable checks remain nonsemantic; a nonlater attempt cannot close
  its own finding. Generic reviewer approval remains `recheck_required`.
- PASS: actual journal reads reject foreign run/repository requests and
  integrity-valid but semantically invalid history, without journal mutation.
- PASS: 129 retained-history inputs produce 128 entries, one omission and a
  deterministic omission hash that changes when omitted evidence changes.
- PASS: actual Git `grep` initially fails, an explicitly authorized fixture
  file effect changes the candidate, the same native check passes, and a
  deterministic independent-role review reaches READY. Before review the
  finding remains open; afterward the report binds the original identity to
  the exact matching native pass. Initial execution: 4.985 s.
- PASS: existing graph design → writer → fresh native checks/review fixture
  preserves the original reviewer finding identity and reports recheck required
  after READY. Initial expanded scoped suite: control 20.139 s, CLI 0.139 s.

Final focused suite: control 38.421 s, CLI 0.399 s, documentation 0.828 s,
exit 0. `go vet` for control/CLI and `git diff --check` passed. Independent
The actual CLI planning/resume fixture with `--closure` and unchanged export
passed separately in 27.020 s; no invented finding or acceptance is returned.
GPT 6 Luna High static code review: APPROVE; reviewer did not run tests.
The omission hash binds original IDs, oracle definitions and sequence, not
derived closure statuses for omitted entries.

These are deterministic fixtures and actual local native processes;
no live model repair quality, token benefit, installed package or final v2
qualification is established. Explicit reviewer recheck remains pending.
