# Fabric v1 real-repository evaluation: pinned custom tasks

This suite defines eight custom real-repository tasks, each pinned to an
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
independent text-encoding API extensions: Int64 and Uint64. Native baseline
tests pass; the held-out baseline builds and fails runtime interface
assertions for both missing APIs. The [preflight receipt](results/atomic-numeric-preflight-20261003.json)
records zero provider calls. This task is intended for serial/parallel
qualification, separately from the original six-task PR #5 comparison.
Its Windows native check has the same narrow NocmpIntegration exclusion as
go-atomic; its held-out check never excludes a test.

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
  IDs. Hidden acceptance tests run only in a disposable byte-copy of the
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
