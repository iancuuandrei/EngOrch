# Contract-context evaluation runner

The Native evaluation runner accepts `go-contract-context-v1` as an explicit
planner-context treatment. It requires the same pinned local RI executable and
lowercase SHA-256 binding as `go-source-context-v1` and
`go-source-context-v2`. The runner verifies the executable bytes before
preparation and records the requested and inspected mode, executable path, and
digest. `PR5Matched` remains unchanged and does not accept planner-context
overrides.

This runner support is not a qualification result. The compiler and controller
mode still require a fresh matched evaluation with the existing native and
held-out gates. Do not resume or append tasks to a prior run.

## Matched comparison

For a context-only comparison, use one frozen product binary and runner for
both arms. Prepare fresh serial Native runs with identical task pins, objective
text, model, effort, Go executable, Codex runtime, RI executable and digest,
prompt recipe, writer-edit validation, and held-out checks. The sole treatment
difference is:

- control: `-PlannerContext go-source-context-v2`
- treatment: `-PlannerContext go-contract-context-v1`

Pass the same explicit values to both arms:

```powershell
$common = @{
    Action = 'Prepare'
    EvalMode = 'Native'
    FabricExe = 'C:\tools\fabric-v1.0.10.exe'
    GoExe = 'C:\tools\go\bin\go.exe'
    CodexExe = 'C:\tools\codex.exe'
    CandidateCopyExe = 'C:\tools\candidatecopy.exe'
    Model = 'gpt-6-luna'
    Effort = 'high'
    PlannerContextRIExecutable = 'C:\tools\engorch-ri.exe'
    PlannerContextRIExecutableSHA256 = '<64 lowercase hex characters>'
    PromptRecipe = 'cache-prefix-v1'
    ValidateWriterEdits = $true
    MaxParallel = 1
    TaskIds = @(
        'go-humanize', 'afero', 'go-multierror', 'go-atomic',
        'go-difflib', 'logr', 'godotenv', 'go-atomic-numeric-text'
    )
}

$control = $common.Clone()
$control.PlannerContext = 'go-source-context-v2'
$control.RunRoot = 'D:\eval\context-v2'
& .\scripts\evaluate-v1.ps1 @control

$treatment = $common.Clone()
$treatment.PlannerContext = 'go-contract-context-v1'
$treatment.RunRoot = 'D:\eval\contract-context-v1'
& .\scripts\evaluate-v1.ps1 @treatment
```

The example's task selection is the fixed six repository tasks plus the
`godotenv` and generated numeric-text tasks. Use identical task selection and
objective-v2 manifest text in both arms. Each root must be new and external to
the source checkout. Prepare must record zero provider calls. Evaluate the
prepared arms sequentially only after the product, runner, and evaluation gates
are qualified and authorized.

Comparing a v1.0.9 v2 run to a v1.0.10 contract run changes both product version
and planner mode. Such a comparison can describe the two releases, but cannot
isolate the effect of contract context. The existing v1.0.9 prepared runs have
different task coverage and are not reusable as either arm in this cohort.

## Outcomes and measurements

Per-task acceptance remains the primary result: the run must complete, inspect
as READY, bind the current candidate, pass native verification and held-out
checks, and receive reviewer approval with no findings. Keep blocked, failed,
and unresolved rows in the report; do not treat missing outcomes as zero or
successful evidence.

The runner records elapsed time, repairs, runtime invocations, and typed usage
when every invocation has matching receipts and complete typed fields. Available
usage fields include input, cached input, uncached input, output, and reasoning
output tokens. Missing or inconsistent measurements stay null. Provider-call
count is not inferred from invocation count.

Contract admission evidence is partial by design. For any prompt-size analysis,
retain the bounded admission-record digest, excerpt and omission counts, and
record size as separate evidence. Do not infer semantic completeness from a
smaller context or claim quality improvement from token changes alone.
