# Fabric v1 product gate

Status: Windows happy path accepted on 2026-10-02 from a fresh GitHub checkout
at the exact source recorded in the
[executed acceptance evidence](../evaluation/v1-product-gate.md).

Fabric exists to build useful agent infrastructure. Product source may be edited
directly by contributors; using Fabric to author its own changes is optional
dogfooding. Preserve runtime effect ownership and durable evidence.

## User outcome

One Windows PowerShell guide takes a new user through:

`clone -> build/install -> configure a real model -> init -> task -> plan/delegate/execute/verify/review -> inspect result and evidence`

The demonstration uses a small committed project and a real coding task: add
a behavior with a meaningful regression test, delegate repository inspection to
a read-only explorer, produce a writer change, run the project's checks, and
obtain a reviewer verdict. The user can inspect the diff, model operations,
verification receipts and verdict, then find the same evidence after restarting.

## Starting gaps and delivered changes

- The former main starter path used a deterministic fake planner and stopped
  before implementation. A real-task guide is now the README's primary entry
  point; the fake path remains an offline smoke check.
- Portable local packaging exists, but configuring real roles and completing
  a task are separate from installation. A source build is sufficient initially.
- The task script now composes planning, exploration, writer proposals, file
  approval, verification, review and inspection. It displays approvals without
  requiring manual JSON plumbing and stops on failures or uncertain attempts.
- `provider-api` is executable only for planner in the current implementation.
  It cannot be advertised as a complete task route. Evaluate the existing
  Codex app-server and OpenCode routes for all required roles before selecting
  the happy path. Do not add a new transport merely to finish this milestone.
- Real-role initialization uses a stock Codex binary and existing user-owned
  authentication. Typed explorer output fixes a demonstrated first-run defect.

## Delivery order

1. **Usable setup.** Build from public main, choose one complete stock runtime,
   add a checked example configuration and setup commands, and make `doctor`
   identify a missing tool, credential reference or role setting clearly.
2. **Usable coding task.** Complete one real task through existing controller
   operations. Add a thin CLI/script entry point only where the task otherwise
   requires manual JSON plumbing. Show approvals and the resulting diff. Fix
   demonstrated defects directly, with regression tests appropriate to the defect.
3. **Repeatable first run.** Publish one guide and repeat it in a new checkout
   with new state. Capture the task result, check outcomes, reviewer verdict
   and evidence locations. The guide becomes the root README's main entry point
   only after this acceptance run succeeds.

No private G0 runner, historical journal, patched local tool directory or internal
qualification manifest may be a prerequisite. Rust semantic indexing is optional
for this first task. Signed releases, hosted publication, recursive execution,
multiple providers and additional platforms are subsequent milestones.

## Acceptance evidence

| Step | Required observable evidence | Current status |
| --- | --- | --- |
| Clone/build | Fresh public checkout and working executable | PASS |
| Configure/init | Documented real role configuration and successful doctor | PASS |
| Plan/delegate | Real plan and a read-only explorer result | PASS |
| Execute | Useful writer change visible in an isolated workspace | PASS |
| Verify/review | Project check receipts and candidate-bound review verdict | PASS |
| Inspect/restart | Diff and durable evidence remain readable after restart | PASS |

These rows are supported by the real-model run identified in the acceptance
evidence. Fixture tests support local behavior separately. Failed or uncertain
invocations retain their actual outcomes. Release publication remains separate
from this accepted journey.

An initial baseline build passed before implementation. The real-model acceptance
record above supersedes that earlier build-only smoke result.

## Development limits

Freeze controller architecture except for demonstrated corruption, duplicated
effects, or blockers to normal use. Give an infrastructure defect one focused
session (default two hours) and at most two implementation attempts. Record an
unresolved defect and use an existing supported route where possible. Never
repeat an uncertain external effect. Keep work packages tied to an observable
product improvement; bookkeeping alone is not a deliverable.

After this gate, prioritize scheduling/parallel work, context routing, adaptive
model allocation, codebase intelligence and verification. Adopt JEV or mathematical
orchestration when an executable experiment shows an improvement to decisions.
