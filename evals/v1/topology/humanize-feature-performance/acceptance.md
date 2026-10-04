# Matched serial and isolated acceptance recipe

## Freeze the treatment

Use the unchanged manifest task and objective-v2 task text in both arms. Pin
the same source commit, Fabric binary, runner, Go executable, Codex runtime,
model and effort, RI executable and SHA-256, prompt recipe, edit validation,
repair budget, native check, and held-out suite. Use fresh detached source
copies and distinct external controller-state roots. Prepare both arms before
either evaluation; Prepare must make zero provider calls. Evaluate serially,
one arm at a time. Do not resume an earlier Humanize run.

Use isolated-writer execution with the same resource policy in both arms; only
the scheduler cap differs:

- Serial control: `MaxParallel = 1`, `IsolatedWriters = true`.
- Two-writer treatment: `MaxParallel = 2`, `IsolatedWriters = true`.

Enable `ReviewImpactContext = $true` in the common hashtable for both arms.
Use `PlannerContext = 'go-contract-context-v1'` with the same explicit pinned
RI parser path and SHA-256 in both. The runner rejects Prepare/Evaluate mode
mismatches, requires this exact planner mode for the feature, records the
requested flag, and verifies that the inspected run reports version 1 plus one
durable context record for the exact reviewed candidate. Do not infer context
usefulness from the requested flag alone; retain candidate binding, partial
coverage, omission counts, and unavailable reasons from the observed record.
Both arms use the pinned CLI default repair allowance of two.

Create the policy at an absolute path outside both source checkouts and pin its
content hash through the runner:

```json
{
  "version": 1,
  "capacity": {
    "cpu_milli": 2000,
    "memory_mib": 4096,
    "verification_slots": 1,
    "total_runtime_slots": 2,
    "provider_slots": 2,
    "model_slots": 2,
    "runtime_slots": 2
  },
  "estimate": {
    "cpu_milli": 1000,
    "memory_mib": 1024,
    "verification_slots": 0,
    "runtime_slots": 1
  }
}
```

Use the existing runner, selecting only the exact manifest task:

```powershell
$common = @{
    Action = 'Prepare'
    EvalMode = 'Native'
    FabricExe = 'C:\tools\fabric.exe'
    GoExe = 'C:\tools\go\bin\go.exe'
    CodexExe = 'C:\tools\codex.exe'
    CandidateCopyExe = 'C:\tools\candidatecopy.exe'
    Model = 'gpt-6-luna'
    Effort = 'high'
    PlannerContext = 'go-contract-context-v1'
    PlannerContextRIExecutable = 'C:\tools\engorch-ri.exe'
    PlannerContextRIExecutableSHA256 = '<verified lowercase SHA-256>'
    ReviewImpactContext = $true
    PromptRecipe = 'cache-prefix-v1'
    ValidateWriterEdits = $true
    TaskIds = @('go-humanize-feature-performance')
}

$serial = $common.Clone()
$serial.RunRoot = 'D:\eval\humanize-serial'
$serial.MaxParallel = 1
$serial.IsolatedWriters = $true
$serial.IsolationPolicyPath = 'D:\eval\policy\humanize-two-writers.json'
& .\scripts\evaluate-v1.ps1 @serial

$parallel = $common.Clone()
$parallel.RunRoot = 'D:\eval\humanize-isolated'
$parallel.MaxParallel = 2
$parallel.IsolatedWriters = $true
$parallel.IsolationPolicyPath = 'D:\eval\policy\humanize-two-writers.json'
& .\scripts\evaluate-v1.ps1 @parallel
```

Before Evaluate, clone each Prepare hashtable, set `Action = 'Evaluate'`, and
add its exact `RunId` from that arm's `run.json`; keep the other bindings
identical. Evaluate the serial run first, then the isolated run. Do not invoke
Evaluate until the exact product binary, runner, source pin, policy bytes,
and acceptance gates are frozen.

Require equal task,
objective, source, model/runtime, Go, RI, prompt, edit-validation, and repair
bindings in the two retained run receipts; reject the pair if any differ.

## Evidence for topology

The cap-2 isolated treatment qualifies the two-leaf topology only if its
validated graph and journal prove all of these:

1. A read-only design/research task completed before implementation and caused
   no file effect.
2. The two initial ready implementation tasks have disjoint write sets with
   the specified bytes/comma source-test boundaries. Their IDs need not be
   prescribed.
3. Each task has a distinct confirmed child-worktree/candidate binding rooted
   in the same pristine parent candidate, and a task-bound writer receipt.
4. Both writer dispatch intervals overlap by a positive duration. These are
   controller task-dispatch intervals, not provider-request overlap or a
   provider-call count.
5. The parent has one deterministic aggregate for exactly those two results,
   confirmed parent file effects, and fresh native verification, both
   unchanged held-out assertions, and review approve with zero findings on the
   combined candidate. There are no unresolved external intents.

The cap-1 serial control must finish with the same acceptance gates and have
no overlapping writer-dispatch intervals. Since the cap is also a planner
input, it may choose one cohesive implementation task; record that actual
plan and do not require it to bind the same two implementation contracts as
the cap-2 treatment. If the cap-2 treatment emits one implementation task,
lacks the read-only hub, or proposes overlapping writes, record `NOT
QUALIFIED` for the two-leaf topology. Do not edit or replan either graph to
manufacture the desired cohort.

The runner's existing combined held-out check composes the unchanged
`humanize.heldout_test.go` and `commaf_performance.heldout_test.go`; retain its
exact PASS result. Native acceptance remains the manifest's full
`go test -count=1 ./...` check. Report each task's READY state, candidate and
review identities, repair use, native/held-out outcomes, dispatch intervals,
and typed usage only where complete. Unknown token fields and provider-call
counts stay unknown. A single matched pair demonstrates this topology once;
it does not establish general speedup.
