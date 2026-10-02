# Fabric v1 product gate

Status: planned; the complete fresh-user journey has not passed acceptance.

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

## Current gaps

- The root README's starter path uses a deterministic fake planner and stops
  before implementation. That path remains useful as an offline smoke check.
- Portable local packaging exists, but configuring real roles and completing
  a task are separate from installation. A source build is sufficient initially.
- The CLI already exposes planning, exploration, writer proposals, file approval,
  verification, review and inspection. The user must currently compose these
  operations and exact JSON identities themselves.
- `provider-api` is executable only for planner in the current implementation.
  It cannot be advertised as a complete task route. Evaluate the existing
  Codex app-server and OpenCode routes for all required roles before selecting
  the happy path. Do not add a new transport merely to finish this milestone.
- Existing model guides contain historical qualification details. The happy
  path must work with publicly obtainable tools and user-owned authentication.

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
| Clone/build | Fresh public checkout and working executable | PASS local build; independent first-user repeat NOT RUN |
| Configure/init | Documented real role configuration and successful doctor | NOT RUN |
| Plan/delegate | Real plan and a read-only explorer result | NOT RUN |
| Execute | Useful writer change visible in an isolated workspace | NOT RUN |
| Verify/review | Project check receipts and candidate-bound review verdict | NOT RUN |
| Inspect/restart | Diff and durable evidence remain readable after restart | NOT RUN |

Fixture tests validate local behavior; they do not fill the real-model acceptance
rows. A failed or uncertain invocation remains recorded with its actual outcome.

On 2026-10-02, a new clone at public main
`72b0f43fae532524bee39e8aa07467cf967552fd` built successfully using local
Go 1.27.1 on Windows amd64. The output was outside the repository and `help`
executed successfully. Binary SHA-256:
`127ccc8006f8833e81e3112e1d9ad7a9106256cedba81463d2c2001dffd377af`.
The documentation check and `git diff --check` passed. No product source changed;
this observation does not establish publicly obtainable toolchain setup, a real
model task, or the complete fresh-user gate. Full product tests were not run for
this documentation-only change.

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
