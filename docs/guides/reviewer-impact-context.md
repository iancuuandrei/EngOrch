# Candidate-bound reviewer impact context

The reviewer impact context is an opt-in hint for autonomous reviews. It is
available only with `--planner-context go-contract-context-v1` and the same
absolute, SHA-256-pinned RI executable used to admit that planner context:

```powershell
fabric run --autonomous `
  --planner-context go-contract-context-v1 `
  --planner-context-ri-executable 'C:\tools\engorch-ri.exe' `
  --planner-context-ri-executable-sha256 '<lowercase-sha256>' `
  --review-impact-context `
  --max-repairs 2 `
  'Update the Go package behavior and its tests'
```

Before dispatching a reviewer for each verified candidate, Fabric captures the
candidate under the worktree read lease, derives a bounded Go graph overlay
from the already-admitted committed graph, validates the module inventory and
candidate binding, and records the result in the run journal. If that exact
candidate already has an admitted context (including after a pending reviewer
intent), the record is reused without recollecting or dispatching another RI
process. The history is bounded by the initial candidate plus the configured
repair allowance, up to the controller's hard limit of nine records.

The reviewer receives a deterministic topology projection capped at 8 KiB,
plus candidate and graph digests, path-change counts, omissions, and an explicit
coverage marker. Durable source text is not copied into the reviewer prompt.
Each candidate observation is separately bound to the repository source,
candidate file hash, RI producer, committed graph, and committed and candidate
module inventories. If the bounded durable record cannot retain a replayable
corpus, the reviewer receives an explicit unavailable reason and summary.

This is partial advisory evidence. Import and generator relationships are
observed syntax/declarations; call hints are unresolved syntax, not Go name
resolution. Deleted, omitted, unparsed, or otherwise absent paths remain
unknown. The topology does not prove test completeness, behavioral correctness,
or safe task independence, and it grants no write or verification authority.
Use the normal candidate tools and verification results for review evidence.

The flag changes reviewer input only. It does not change planner partitioning,
verification selection, repair budgets, or any legacy invocation when omitted.
