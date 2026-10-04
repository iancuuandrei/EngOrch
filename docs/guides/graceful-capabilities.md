# Capability fallback and acceptance

Fabric keeps identity, ownership, safe-path, credential, journal and uncertain
effect checks strict. A fallback never authorizes a write, changes the selected
model, grants a worker credentials or turns an unverified candidate into READY.

For new autonomous runs, optional preferences can degrade before dispatch:

| Preference unavailable | Selected behavior |
| --- | --- |
| Pinned RI executable absent | Bounded committed-source planner context; dependent topology/cache options disabled |
| Parallel writer runtime unsupported | Serial writer on the configured model |
| External controller state absent for isolated writers | Serial writer |
| Per-writer estimate exceeds declared isolation capacity | Serial writer |
| Any configured role lacks Codex compaction support | Compaction disabled |
| Optional syntax-cache directory permission or storage failure | Recompute without that cache |

The run's immutable execution policy records pre-dispatch fallback choices as
`capability_fallbacks`, including capability, reason, disposition and selection.
Existing run inputs are never rewritten. An existing parser's hash mismatch,
unsafe state path, malformed policy or corrupt cache identity remains an error.
The fallback baseline still requires its own normal admission checks.

OpenCode context-tool bursts use a bounded serial FIFO queue for planner,
explorer, writer, fixer and reviewer. The queue preserves the existing catalog
and ownership checks; it does not create extra model dispatches. Requests beyond
the finite queue bound and cancellation remain observable failures.

Autonomous failure output retains the durable `run_id`, controller `state`,
`phase` and `blocked_reason`, and adds `status`, `disposition`, `next_action` and
`candidate_retained`. Repair-budget exhaustion produces `NEEDS_ATTENTION` with
the candidate and failed-gate evidence retained. An unresolved effect produces
`UNKNOWN` / `DENY`, requiring reconciliation without resend. Other failures
request operator inspection. Exit codes still indicate that acceptance was not
achieved; NEEDS_ATTENTION is never READY and does not grant a new repair budget.

Scope expansion is not automatic. A confirmed proposal outside declared
ownership produces `NEEDS_REPLAN`, preserving the proposal and candidate while
requesting a candidate-bound scope replan under the governing authority. No file
effect is authorized by that status. Invalid proposals do not
fall back to whole-file replacement. Live resource scheduling and memory
pressure adaptation are not established by these pre-dispatch preferences.
