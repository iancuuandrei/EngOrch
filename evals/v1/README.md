# Fabric v1 real-repository evaluation: pinned custom tasks

This suite defines ten custom real-repository tasks, each pinned to an
audited upstream commit. It is not a SWE-bench score and it is not a general
full-v1 proof; it reports per-task behavioral acceptance on the selected
repositories only. `manifest.json` is the authoritative source, commit,
license, task, and package map. Source-task licenses were verified against
each pinned repository's LICENSE file: go-humanize MIT, afero Apache-2.0,
go-multierror MPL-2.0, go-atomic MIT, go-difflib BSD-3-Clause, logr
Apache-2.0, godotenv MIT (`LICENCE`). The original six-task comparison remains
a six-task result; adding a manifest entry does not add a successful run.
The seventh task exercises multiline dotenv parsing, compatibility and
documentation in a previously unfamiliar parser repository. Its pinned
upstream tests pass before the change; its held-out multiline assertions fail
before the feature. A [fresh public-main attempt](results/godotenv-main10-20261003.json)
was blocked before file application because a proposed source anchor did not
match the current file. Actual Fabric completion is still pending.

`go-atomic-numeric-text` uses the same pinned MIT go-atomic source for two
text-encoding API extensions: Int64 and Uint64. Both wrappers share a generator
template; a regeneration-stable implementation requires explicit shared
ownership or dependencies. Their distinct filenames do not establish safe
independence. Native baseline
tests pass; the held-out baseline builds and fails runtime interface
assertions for both missing APIs. The [preflight receipt](results/atomic-numeric-preflight-20261003.json)
records zero provider calls. The earlier parallel attempt used an incorrect
independence assumption and exhausted two repairs with native PASS but review
changes requested for generated-source drift. See
[the retained failure](results/numeric-generated-code-de2f587-20261003.json).
The task now explicitly requires generator consistency; it is not the selected
serial/parallel qualification case.
Its Windows native check has the same narrow NocmpIntegration exclusion as
go-atomic; its held-out check never excludes a test.

`go-humanize-commaf-performance` measures a practical formatting improvement
on the pinned MIT go-humanize repository. The pristine full upstream suite
passes. On Go 1.27.1, ordinary `Commaf` inputs use four allocations per call;
finite extremes use up to eleven. Acceptance preserves exact output on
boundary values and deterministic float bit patterns and requires at most
two allocations per call on the documented finite representatives. Allocation
counts are the performance gate; benchmark timings are informational. The
[preflight receipt](results/commaf-performance-preflight-20261003.json)
records full native PASS, output-equivalence PASS and the targeted allocation
baseline failure, with zero provider calls.
This task adds no successful result to the original six-task comparison.

The [separate Commaf acceptance](results/commaf-performance-accepted-20261003.json)
passed native, review and supplementary candidate-bound held-out checks;
the original evaluation's copy-decoder BLOCKED outcome remains preserved.

`go-humanize-feature-performance` combines the ParseBytes separator feature
with the Commaf allocation task on the same pinned repository. The source/test
ownership groups are `bytes.go`/`bytes_test.go` and `comma.go`/`comma_test.go`.
Both behavioral and performance oracles are mandatory. The planner must still
verify independence; the task does not force unsafe parallel execution. The
[runner-supported topology recipe](topology/humanize-feature-performance/README.md)
uses fresh isolated arms with `MaxParallel` 1 and 2, the same resource policy,
and candidate-bound reviewer impact enabled in both. It preserves the manifest
objective, source, binary, model, runtime, full native check, and both held-out
assertions. The recipe is not yet qualified; adding this task is not accepted
parallel execution or a speedup claim.
The [composed preflight](results/humanize-composed-preflight-20261003.json)
records full native baseline PASS and compiled, targeted baseline failures for
both required improvements, with zero provider calls.
Prepare and acceptance reuse the two unchanged original fixtures through an
exact-selector wrapper that invokes both. The earlier
[standalone fixture preflight](results/humanize-feature-performance-preflight-20261003.json)
is retained as historical evidence for that previous fixture implementation.

The portable [scheduling decision experiment](scheduling-decision-experiment/README.md)
compares the actual nonparked ready-task priority with estimated critical-path
priority and demonstrates an unsafe evidence-stopping counterexample. These
are modeled outcomes, not real runtime gains; neither heuristic is adopted.

Held-out acceptance sources live in `evals/v1/heldout/` and the classifier in
`evals/v1/harness/Classify-CheckOutput.ps1` with regression fixtures under
`evals/v1/harness/fixtures/`. The afero fixture uses external test package
`afero_test` so it can live inside the afero checkout without a self-import
cycle (`package afero` importing `github.com/spf13/afero` is an import cycle).
The nested `heldout/go.mod` keeps these separate upstream-package fixtures
out of Fabric's own Go package discovery; each is tested only after copying
it into its matching upstream repository.

## Runner

The pinned go-difflib test sources contain pre-existing `go vet` failures
(nonconstant `fmt.Printf` and example names) on Go 1.27.1. Its explicit-file
commands use `-vet=off` in both comparison arms, recorded as
`legacy-tests-without-vet`. All upstream behavioral tests and held-out
assertions still run; no static-analysis PASS is claimed for that task.
`-NativeBuildReceiptPath` optionally supplies an exact binary build receipt;
the runner validates it when explicitly supplied and never discovers an old
receipt from a machine-specific artifact directory.

The selected `-GoExe` is also bound to the environment of each Fabric child:
its directory is temporarily placed first in PATH, and `go` must resolve to
that exact selected executable before the command starts. The caller's PATH
is restored after success or failure. This keeps the manifest's verification
argv unchanged while making its `go` executable available to Fabric. Go
identity and child-resolution policy are recorded in evaluation provenance.

Full inspect snapshots remain in the task evidence and supply the evaluation
gate. The acceptance copier receives a separate bounded snapshot projection
with the unchanged run/state/workspace/candidate/verification/review fields;
unrelated host context stays in the full artifact. Both artifacts are hashed.
Canonical event and copier input limits remain enforced.

- `scripts/evaluate-v1.ps1 -Action Prepare` prepares a unique run
  directory with fresh pinned detached clones. Fabric config
  (`.harness/`/`harness.toml`) is excluded from the candidate diff via
  `.git/info/exclude`. For each task it runs a native `go test` preflight
  (must pass on the clean pin), then adds the held-out test only to a
  disposable `baseline-checks/<id>` copy and classifies the result as
  `targeted_assertion`, `expected_missing_api` (only go-multierror
  `ErrorsSnapshot` and go-atomic `MarshalText`/`UnmarshalText`), `setup_error`
  (import cycle, build failure, missing module: NEVER an intended FAIL), or
  `pass`. Exact stdout/stderr per check is captured under
  `baseline-checks/<id>/*.log`. The task checkout at `tasks/<id>/repo` never
  contains the acceptance test during preparation. No model/provider is
  called (`provider_calls=0`). Existing run directories are never
  overwritten. The script writes a sanitized `run.json` with product/runner
  hashes; it does not retain transcripts or credentials and never
  commits/pushes.
- Windows verification is scoped, not the full upstream suite: on Windows
  only (`$env:OS -eq 'Windows_NT'`), native `go test` preflight and final
  native package verification for `go-atomic` pass `-skip
  ^TestNocmpIntegration$` (rationale recorded in
  `manifest.json#native_verification` and per-task run evidence
  `preflight.args.log` / `native-verify.args.log` with `windows-scoped`
  scope). Upstream `nocmp_test.go` sets `cmd.Env` to `HOME` only, so
  `TEMP`/`TMP` are unset and Go temp defaults to `C:\Windows`, causing
  access denied. Precisely `^TestNocmpIntegration$` is skipped; held-out
  task tests (`-run=^TestFabricV1Heldout$`) are never excluded and Linux
  runs the full native suite. No upstream product sources are modified.
- `scripts/evaluate-v1.ps1 -Action Prepare` is the same prepare with explicit
  `FabricExe/GoExe/CodexExe/Model`, external run root, `-TaskIds` filter and
  new run nonce. `-Action Evaluate` is the only provider-gated path. Before
  any run is created, the task `harness.toml` required checks are set to
  the explicit manifest `native_argv` policy (`init` defaults to
  `go test ./...`) and hashed. Native mode runs
  `fabric init --codex EXE --model MODEL`, then native
  `fabric run --autonomous` objective plus `inspect`/`usage`/`diff` evidence
  gathering. PR5Matched mode uses the explicit `-PR5BaselineScript` (default
  `scripts/run-task.ps1`) with the supplied baseline exe: init with the
  baseline exe/model first, same verification policy and model/runtime,
  `inspect`+`usage` with the baseline executable, exact run identity from
  the script result, and diff collected from the candidate path (the
  baseline lacks `diff`); the baseline binary SHA is recorded, not the
  native SHA. It reports the two required plan/file approval interventions
  explicitly (preauthorized, no invented human identity) and no repairs.
  Task PASS requires a successful run+inspect, `READY` snapshot, exact
  current candidate, non-pending all-PASS verification, reviewer `approve`
  with zero findings and matching `candidate_id`+`verification_plan_id`,
  plus native upstream verification AND held-out PASS; missing evidence is
  `BLOCKED`, never rerun. The workspace is exactly
  `snapshot.workspace.request.path`, validated inside the task worktree
  registry and bound to the run; there is no task-checkout fallback.
  Repair accounting uses `repair_attempts` (absent stays null); the review
  `result.output` JSON verdict is parsed for decision/findings with bound
  IDs. Native and hidden gates start from independently captured disposable
  copies. The hidden-gate copy is captured after native tests finish, so native
  test writes cannot become the hidden gate's input. Each helper observation
  must match the same candidate, workspace, file hash and file count; separate
  paths and observations are retained. Hidden acceptance tests run only in a byte-copy of the
  finished candidate by the explicit `-CandidateCopyExe` Go helper under
  a read lease. The helper binds the live candidate to both the review and
  verification plan, captures exact file bytes and modes, and rejects
  unsupported entries rather than silently dropping symlinks. The original
  is never mutated. `Copy-CandidateTree.ps1` is a standalone fixture utility,
  not the acceptance copier. `UNKNOWN` outcomes are recorded `BLOCKED` and never automatically
  rerun. `receipt_matched` rows are completed runtime invocations, not
  provider calls: `provider_calls` stays null unless an actual count
  exists, while runtime invocations and `provider_usage` tokens are recorded
  separately. Diff metrics count untracked added files separately with an
  exact 1 MiB bound. Unknown metrics stay `null`, never zero. No raw
  transcripts or credentials are published.

The checks are deliberately behavioral and run after the relevant change is
made. They are a pre-task baseline discriminator, not evidence of model
quality. Do not expose `heldout` check source or acceptance details in task
prompts. A later execution harness must inject checks only after a task
reaches its stated completion point, run the repository's normal checks, and
append an outcome record bound to the exact task candidate and
Fabric/runtime identities.

## Metrics contract

Each eventual task record must distinguish candidate source SHA and dirty-tree
digest, Fabric source/binary SHA-256, runner/config/runtime identities, terminal
state (`PASS`, `FAIL`, `NOT RUN`, or `BLOCKED`), held-out/normal check counts,
review findings, repair attempts, elapsed time, runtime invocations, provider
calls and token counts, diff size plus untracked added files, and human
interventions. Unknown monetary usage and unavailable provider token counts
stay `null`. Raw transcripts and credentials are excluded. Legacy
`go-difflib` (no `go.mod`) uses explicit source-file argv
(`go test difflib/difflib.go difflib/difflib_test.go`, plus the held-out file
when present) everywhere: preflight, Fabric required checks, and final
native verification.

The PR #5 sequential journey remains the comparison baseline. It requires its
own clean clone and explicit human approval/review accounting; its historical
single-task acceptance is not treated as a six-repository result.

The opt-in `-CandidateFactsCache` runner switch is available only in Native
mode with `-ReviewImpactContext`, `-PlannerContext go-contract-context-v1`, and
the explicit pinned RI executable/hash. Pass it to both Prepare and Evaluate;
the runner rejects mismatches before evaluation effects and checks the
inspected execution policy version. Default runs omit the switch and preserve
their previous command and receipt shape. A recorded cache-policy version is
not a cache-hit measurement: corpus diagnostic statistics are not included in
the evaluation receipt, and no hit-rate or performance claim follows from the
flag.

## Controlled parallel-writer comparison

Native evaluations accept `-ParallelWriters` and `-MaxParallel 1..8`.
`-ValidateWriterEdits` additionally selects the experimental
`anchored-edits-v2` contract during initialization of a fresh Codex task.
It enables read-only anchor validation during the writer/fixer turn; it does
not authorize files or replace final proposal validation. The default init
arguments remain unchanged. PR5Matched rejects these newer policy overrides
before effects. The evaluation and native task row record
`writer_edit_validation_requested`; this records a request, not proof that the
model used the tool or completed the task.
Omitting both preserves the original invocation; these overrides reject in
`PR5Matched` mode before any evaluation effects. The runner records requested
policy separately from observed execution. A requested flag is not evidence
that two writers actually ran.

For a serial/parallel pair, prepare two fresh run directories with identical
`-TaskIds`, source pins and objectives. Evaluate both with the same clean
Fabric binary, runtime, model, effort, helper and build receipt, and enable
`-ParallelWriters` in both. Set `-MaxParallel 1` for the serial arm and
`-MaxParallel 2` for the parallel arm. Do not reuse or resume an uncertain
evaluation as the second arm.

Acceptance still requires the normal candidate-bound native tests, held-out
checks and review. Inspect durable evidence for two genuinely independent
implementation tasks, distinct actual runtime invocations, their start/end
times, and one merged file effect. A single-writer plan can be a valid product
result but cannot establish parallel implementation benefit. Compare elapsed
time and token usage only within the accepted, equivalently scoped pair;
unknown provider request counts or monetary costs remain unknown.

## Planner-context treatment comparison

Native evaluation can opt into the immutable `source-bounded-v1` planner
context with `-PlannerContext source-bounded-v1`. Omitting the parameter keeps
the legacy empty planner-context setting and emits no planner-context run
argument. The evaluator records the requested treatment in each prepared and
evaluation receipt, requires Evaluate to match its prepared run, and records
the value observed in the inspected run creation. `PR5Matched` rejects this
Native-only option.

The separate `go-source-context-v1` treatment also requires an explicitly
selected RI parser: pass `-PlannerContextRIExecutable` with its absolute clean
path and `-PlannerContextRIExecutableSHA256` with the lowercase SHA-256 of the
executable bytes to both Prepare and Evaluate. The runner verifies those bytes
before preparation/evaluation, records the path and digest in receipts, and
requires the inspected run to retain the same binding. It never discovers the
parser from PATH or environment variables. This treatment uses bounded,
partial Go source graph evidence and should be evaluated as a distinct arm;
the `source-bounded-v1` comparison below does not qualify it.

For a matched comparison, prepare fresh runs for both treatments from the same
six manifest tasks: `go-humanize`, `afero`, `go-multierror`, `go-atomic`,
`go-difflib`, and `logr`. Evaluate both with the same clean Fabric binary and
build receipt, candidate-copy helper, runner checkout, Codex executable,
`gpt-6-luna` at `high`, pinned Go 1.27.1 executable, verification policy, and
scheduler/writer flags. Use separate new run IDs; do not resume or reuse an
uncertain run. The only treatment difference is omitting `-PlannerContext` in
the control arm and setting it to `source-bounded-v1` in the treatment arm.

```powershell
$runner = 'scripts/evaluate-v1.ps1'
$tasks = @('go-humanize', 'afero', 'go-multierror', 'go-atomic', 'go-difflib', 'logr')
$runRoot = 'D:\dev\Fabric-v1-eval-runs'
$goExe = 'D:\dev\EngOrch-toolchains\go\1.27.1\go\bin\go.exe'
$prepare = @{ RunRoot = $runRoot; GoExe = $goExe; TaskIds = $tasks }
& $runner -Action Prepare -RunId planner-context-control @prepare
& $runner -Action Prepare -RunId planner-context-source-bounded -PlannerContext source-bounded-v1 @prepare

$evaluate = @{
    RunRoot = $runRoot; GoExe = $goExe; TaskIds = $tasks
    FabricExe = $fabricExe; NativeBuildReceiptPath = $buildReceipt
    CandidateCopyExe = $candidateCopyExe; CodexExe = $codexExe
    Model = 'gpt-6-luna'; Effort = 'high'
}
& $runner -Action Evaluate -RunId planner-context-control @evaluate
& $runner -Action Evaluate -RunId planner-context-source-bounded -PlannerContext source-bounded-v1 @evaluate
```

Set `$fabricExe`, `$buildReceipt`, `$candidateCopyExe`, and `$codexExe` to the
same explicitly approved artifacts for both arms. Keep `-ParallelWriters`,
`-MaxParallel`, and `-ValidateWriterEdits` omitted in both, or supply the same
values to both; do not interpret this comparison as a scheduler treatment.
The treatment is an input change, not proof that selected source context was
useful. Report context coverage and omission evidence alongside graph
structure, exact candidate-bound native/review/held-out outcomes, repairs,
elapsed time, runtime invocations and observed token usage. Unknown
provider-call counts and costs stay unknown.

When present, `graph_writer_results[TASK].dispatch.started_at` and `.ended_at`
record the completed writer's controller wrapper interval, including runtime
setup. They permit direct overlap measurement for new runs; they are not
provider-request timestamps. Older runs and unusable clock observations omit
the fields. Do not reconstruct their overlap from journal ordering or requested
worker limits. Retain end-to-end evaluation time, repairs and token usage
separately when comparing accepted serial and parallel arms.
