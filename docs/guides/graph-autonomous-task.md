# Autonomous hierarchical task graph with bounded parallel explorers

Scope: genuine autonomous graph execution in this checkout. New CLI autonomous
runs use graph execution and bounded task context by default; legacy journals
replay identically and old nil/empty policies remain sequential. A real-model
development trial on pinned go-humanize completed planning, exploration,
writing, native verification and approving review; its exact reviewed candidate
also passed independent held-out checks. An installed Afero development task
also passed acceptance after two useful investigations overlapped by 21.890
seconds. A matched speed improvement and final release acceptance remain
unproven; see the real-repository evaluation report.

## Defaults

- `fabric run --autonomous` creates
  `ExecutionPolicy{Mode: autonomous-v1, MaxRepairs, Context: bounded-v1,
  GraphVersion: 1, RepairPlanningVersion: 1, MaxParallel}`.
- `--max-parallel` is 1..8, default 3; `1` is the sequential override
  (graph still applies, but without overlap). `--max-repairs` remains 0..8.
- `--parallel-writers` is an experimental opt-in, off by default; see its
  separate contract and pending acceptance below.
- `--prepare-only` accepts the graph, confirms the isolated workspace and
  pristine candidate, and returns `status: PREPARED` before explorer or writer
  dispatch. The durable run remains `IMPLEMENTING`; continue that exact run
  with `resume --autonomous RUN` after optional RI lexical staging.
- Empty `Context` and `GraphVersion 0 / MaxParallel 0` preserve the legacy
  sequential workflow byte-for-byte; `canonical.Hash("harness.run.v1")` and
  `canonical.Hash("harness.execution-policy.v1")` are unchanged for old runs.
- `Config.PlannerContract` may be `""`, `plan-v1` (legacy) or `plan-graph-v1` / `plan-graph-v2` / `plan-graph-v3` / `plan-graph-v4`
  (strict autonomous graph). Empty retains historical raw planner input
  identity.

## Planner contract

- Default new autonomous CLI runs freeze `plan-graph-v3`; the parallel-writer opt-in freezes `plan-graph-v4`. Both include the exact JSON schema in the invocation and Codex request. Legacy graph invocation recipes are unchanged. Evidence entries require `kind` and `description`; root scope `.` is allowed, but writes require concrete paths. Strict wire objects include every property; empty parent IDs and empty dependency/write arrays represent absent values. Initial implementation scope also bounds subsequent repair designs.

- `plan-graph-v1` instructs planners to return only the engineeringplan v1
  strict JSON schema (version, mode, summary, tasks). `completed` and
  `attempts` are absent from the wire schema; any planner-supplied completion
  is rejected as fictional.
- `ParsePlannerGraph` rejects `Completed`/`Attempts`; `ValidateAutonomousGraph`
  then enforces parent/dependency cycles, write ownership, exactly one initial
  implementation with concrete `WritePaths` within scope, research/design
  dependencies, and native verification/review gates depending on the
  implementation. `ModeDirect` holds exactly one implementation with no
  dependencies, avoiding graph branches for simple tasks.
- The existing `PlanID = Hash("harness.plan.v1", result)` already binds the
  graph bytes; no redundant machine-approval authority was added. Machine
  approval still binds the exact `PlanID` and immutable policy only.

## Durable graph progress

- Events (controller journal, replayed by `control.Replay`):
  `graph.recorded` (revision 1, digest of accepted plan output),
  `graph.progress` (one attempt with actual evidence IDs),
  `graph.revised` (validated evolution, revision+1).
- Keys: existing `PlanID` plus `Digest`/`Revision`. The accepted `Plan`
  remains the evidence; `GraphState.Graph.Completed/Attempts` plus
  `GraphState.Evidence` are runner-owned.
- `graph.recorded` requires `IMPLEMENTING`, matches `PlanID`, and its digest
  equals `Digest(ParsePlannerGraph(Plan.Output))`; substitution rejects.
- `graph.progress` requires dependency completion (actual gating), appends
  attempt history, and validates evidence by kind:
  research/design need an exact recorded `ExplorerRecord` whose question binds
  the graph task ID; implementation needs the actual candidate plus writer
  invocation and scoped paths; verification needs the actual
  `VerificationPlanID` with PASS observations for `completed`; review needs
  the actual review invocation (approve -> completed, else failed).
  Model claims alone never count.
- `graph.revised` uses `ValidateAutonomousRevision`: never-started nodes may
  change/remove; started/completed/UNKNOWN nodes keep identity and history;
  completed results preserved; UNKNOWN blocks retry and revision.
- Repair/review failure in historical graph policies appends the original
  scoped implementation/verification/review extension byte-for-byte. New CLI
  runs bind `RepairPlanningVersion: 1`, with v3 by default or v4 for the parallel
  writer opt-in: each repair slot
  first adds a read-only design task bound to the exact failed gate and current
  candidate. Its recorded `Exploration.Paths` may fill only that never-started
  repair implementation's empty `WritePaths`, and every path must remain
  within the initial implementation's immutable `ScopePaths`. The reviewer
  never grants paths. The resulting graph revision is replay-validated from
  the exact design record; it cannot alter other tasks or exceed the existing
  `MaxRepairs` bound. Unknown/pending effects are never resent.

## Writer scope and READY

- Before file authorization/application, native writer proposal paths must be
  within the active implementation `WritePaths` (union fallback across repair
  implementations for scoped fixes); out-of-scope writes reject before any
  effect.
- `READY` requires every mandatory graph task `Completed` with matching
  evidence plus native verification/review evidence; otherwise it stays
  blocked.

## Experimental parallel implementation opt-in

`fabric run --autonomous --parallel-writers "OBJECTIVE"` selects the new
`plan-graph-v4` recipe and immutable `ParallelImplementationVersion: 1`.
The default remains the v3 singleton writer path. The opt-in requires bounded
context, graph execution, repair planning, the anchored-edits writer contract,
and the JSON-v2 explorer contract. Unsupported writer routes reject before run
creation; the first production route is Codex.

The planner may choose one initial implementation task, or two independent
tasks with disjoint concrete write paths. A small objective does not need a
second writer. `--max-parallel 1` remains valid with the opt-in, allowing the
same two-task shape to execute serially for comparison.

Parallel implementation acceptance is pending. CLI and planner regression
checks pass, but they do not prove two model invocations overlap, safe merged
application, or improved real-task latency. Those require the cohort
integration tests and a candidate-bound real-repository trial.

## Parallel seams

- `taskscheduler.Bind/Tick/Pump` plus `ScheduledDispatchAdapter` carry durable
  static claims. `scheduledInvocation` still requires v2 except for the narrow
  static V1 `OperationExplorer` case: graph-enabled, `IMPLEMENTING` (or the
  exact repair-design cohort in `REPAIRING` with repair planning enabled) with
  resolved candidate/workspace/plan/graph, no `TaskPool`, no dynamic turn.
  Default V1 writer/reviewer/planner and dynamic `agent_turn` retain their
  existing gates. The parallel-writer opt-in adds a separately bound static
  writer cohort; it does not enable dynamic turns.
  V1 `TaskPool` remains forbidden; `PumpOptions.Workers` is
  `min(MaxParallel, readyCount, remaining MaxExplorationRecords)` (1..8).
- `ExplorerHost` singleton is retained for legacy compatibility; graph runs
  keep per-invocation state in `Snapshot.ExplorerRuns` keyed by exact
  invocation ID. Host directories were already invocation-specific
  (`StateRoot/RunID/explorer-<invocation.ID>`); every map lookup and receipt
  binds the exact invocation, never a sibling. `model_access_runtime`,
  `scheduled_dispatch` runtime journals, `usage`, and lifecycle unresolved
  sets all consult the map.
- Frozen batch (`buildExplorerBatch`): for each exact explorer question, admit
  `task_context` (`bounded-v1`) and freeze/confirm `roleRI`/`roleLexical`
  build+overlay selection. `explorerInvocation` binds
  candidate/PlanID/objective/question/RI/lexical/context, not prior
  explorations/head. `TaskSpec.InvocationID` must remain exact at
  dispatch/replay. Questions embed bounded task ID/title/scope/evidence and
  are <=4096 bytes. Completion requires the exact matching `ExplorerRecord`
  and scheduler terminal receipt. Dependent tasks run after dependencies;
  the default single writer runs after all required research/design. The
  opt-in permits at most two independent implementations with disjoint paths.
- Head exception (`ExecuteScheduledClaim` checks `ControllerHead` twice):
  legacy/exclusive gates remain. The explorer exception is static graph-bound
  explorers in the same frozen batch: recomputed InvocationID/RunID/PlanID/
  candidate/lifecycle/host must be unchanged, and every journal event since
  the bound head must be attributable to exact cohort explorer invocations
  and read-only kinds (`explorer.*`, `task.context-admitted`,
  `graph.progress` for research). Writer/lifecycle/mutation/unrelated kinds
  and UNKNOWN reject. Candidate fingerprint is rechecked under a shared read
  lease before/after runtime; no concurrent writer is allowed. Reconciliation
  observes only an existing runtime attempt and never starts a missing call.
  `Pump` stops when the finite batch is terminal (observations + cancellation
  + join), never forever.

## Files and choices

- Owned: `internal/control/*.go` (notably `autonomous_graph.go`,
  `explorer_host.go`, `explorer_run.go`, `scheduled_dispatch.go`,
  `control.go`, `autonomous.go`, `lifecycle.go`, `model_access_runtime.go`,
  `usage.go`, `planner_contract.go`), `internal/config/*.go`,
  `internal/engineeringplan/*.go`, `internal/cli/autonomous.go` and
  `autonomous_cli_test.go`, plus this guide.
- Preserved: native increment, Sonar fixes, bounded-v1 context module and
  builder/direct admission; other CLI files, `taskcontext`/`modelpolicy`
  packages, and evaluation/release files untouched. Existing
  controller/effect/runtime and `taskscheduler` reused; no generic recovery
  or G0 governance added.
- Fake runtime explorers synthesize a deterministic bounded advisory result
  (question as summary, empty paths) under a shared read lease for fixtures;
  codex/opencode paths retain full host, RI/lexical, usage, and model-access
  checks.

## Limits and pending acceptance

For a prepared run with an operator-supplied Rust RI reader, stage the committed
base index and candidate overlay before continuing the graph. Use the exact
reader path and SHA-256 supplied by the operator, save each preview unchanged,
and authorize its displayed intent ID:

```text
harness run --autonomous --prepare-only "OBJECTIVE"
harness ri prepare-lexical RUN EXE EXE_SHA256 STAGE_ROOT 67108864 10000
harness ri lexical RUN BASE_PREVIEW_JSON BASE_INTENT_ID ACTOR
harness ri lexical-ref RUN
harness ri prepare-overlay RUN OVERLAY_STAGE_ROOT
harness ri overlay RUN OVERLAY_PREVIEW_JSON OVERLAY_INTENT_ID ACTOR
harness ri overlay-ref RUN
harness resume --autonomous RUN
```

Both lexical effects are bound to this run; the overlay requires its confirmed
workspace candidate. Runtime explorer, writer and reviewer selection then uses
the confirmed artifacts automatically. A `PREPARED` response is not task
acceptance or `READY`.

- Fixture coverage focuses on legacy identity/replay, strict schema and no
  planner completion, parent cycles, digest/PlanID substitution, dependency
  gating, out-of-scope writer rejection, direct/sequential mode, overlapping
  V1 explorers with separate identities and unchanged candidate, sibling vs
  unrelated/mutation/lifecycle deltas, frozen drift rejection, UNKNOWN
  blocking, and repair evolution preserving evidence with fresh
  verification/review.
- Actual development trials and the matched six-task evaluation are recorded
  in [the evaluation report](../evaluation/v1-real-repository-comparison.md).
  Explorer overlap has been observed; a controlled wall-clock improvement
  and parallel implementation acceptance remain unproven.
