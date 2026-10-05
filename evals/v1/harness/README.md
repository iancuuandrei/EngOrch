# Harness regression fixtures

These fixtures qualify local classification and copy mechanics. Real task
outcomes and token accounting are recorded separately in the
[suite README](../README.md) and
[product ledger](../../../docs/roadmap/v2-product-plan.md); fixture PASS is
not a hosted or complete product acceptance result.

`Test-RepairIntelligenceTreatment.ps1` checks the Native-only structured-repair
treatment: exact run argv, matching Prepare/Evaluate request, observed immutable
policy, rejection of substitution and unchanged legacy omission. Invoke
`scripts/evaluate-v1.ps1 -RepairIntelligence` for both Prepare and Evaluate of
the treatment arm; omit it for the control arm. The product flag also enables
explicit reviewer rechecks, so this compares the bundled treatment. Coverage
spectra are not collected or injected by this option. Trials require fresh
pinned clones, unchanged external acceptance and exact usage/time evidence;
no fixer invocation means NOT_EXERCISED for repair quality.

`Classify-CheckOutput.ps1` (`Get-FabricV1CheckClassification`) exists so the
old `evaluate-v1-baseline.ps1` defect cannot return silently:

- Old defect 1: any nonzero `go test` exit was recorded as
  `FAIL_EXPECTED`. `setup_error_importcycle.sample.log` (self-import cycle)
  and `setup_error_compile.sample.log` (typo / undefined symbol) both exit
  nonzero, but the classifier returns `setup_error`, never
  `targeted_assertion` or `expected_missing_api`. The runner must abort on
  `setup_error`; an import cycle is NEVER an intended FAIL.
- Old defect 2: the afero held-out fixture used `package afero` while
  importing `github.com/spf13/afero`. That is a self-import cycle. The fixed
  fixture is `evals/v1/heldout/afero.heldout_test.go` with
  `package afero_test` (external test package, no self-import).
- Old defect 3: the classifier matched only `undefined:` (colon) syntax and
  fell back to generic `FAIL` -> `targeted_assertion`. The real Go toolchain
  emits `<expr>.<Method> undefined (type <Type> has no field or method
  <Method>)` (space, no colon), and arbitrary `FAIL <pkg>` lines without
  `--- FAIL: TestFabricV1Heldout` must not count as the held-out assertion.
- Positive controls: `targeted_assertion.sample.log` classifies as
  `targeted_assertion`; `expected_missing_api_multierror.sample.log`,
  `expected_missing_api_multierror_actual.sample.log` (exact supervisor form
  `e.ErrorsSnapshot undefined (type *Error has no field or method
  ErrorsSnapshot)`, only task `go-multierror`) and
  `expected_missing_api_atomic.sample.log` (only task `go-atomic`, symbols
  `MarshalText`/`UnmarshalText` for type `*Bool`) classify as
  `expected_missing_api`; `pass.sample.log` (exit 0) classifies as `pass`.
- Negative controls: `setup_error_mixed.sample.log` (one allowed missing-API
  line plus one unexpected `undefined` diagnostic) and
  `setup_error_arbitrary_fail.sample.log` (bare `FAIL <pkg>` with no
  `--- FAIL: TestFabricV1Heldout`) both classify as `setup_error`.

Expected mapping (exit code + task id + log):

| Fixture | Exit | Task | Classification |
| --- | --- | --- | --- |
| targeted_assertion | 1 | afero | targeted_assertion |
| expected_missing_api_multierror | 1 | go-multierror | expected_missing_api |
| expected_missing_api_multierror_actual | 1 | go-multierror | expected_missing_api |
| expected_missing_api_atomic | 1 | go-atomic | expected_missing_api |
| setup_error_importcycle | 1 | afero | setup_error |
| setup_error_compile | 1 | go-humanize | setup_error |
| setup_error_mixed | 1 | go-multierror | setup_error |
| setup_error_arbitrary_fail | 1 | go-humanize | setup_error |
| setup_error_nocmp_env | 1 | go-atomic | setup_error |
| pass | 0 | any | pass |

A missing-API pattern on any other task id (for example `ErrorsSnapshot` on
`go-humanize`) is `setup_error`. Every diagnostic line must be the narrow
allowed shape for the correct task, type (`*Error` / `*Bool`) and method;
mixed diagnostics are `setup_error`. There is no generic `FAIL` fallback.
Environmental upstream failures are also `setup_error`, never
`targeted_assertion`: `setup_error_nocmp_env.sample.log`
(`TestNocmpIntegration` Windows `C:\Windows` access-denied, exit 1, task
`go-atomic`) classifies as `setup_error`. The runner handles that case by a
narrow Windows-only `-skip ^TestNocmpIntegration$` in native preflight and
final native package verification (manifest `native_verification`), not by
reclassification.

## Acceptance byte copy

`Copy-CandidateTree.ps1` (`Copy-CandidateTree`, `Get-DirectoryTreeIdentity`)
is retained only for its standalone fixtures: a real byte copy preserving
changed tracked files, untracked files, and deletions. `git clone HEAD` is
never used for acceptance because it drops dirty/untracked candidate
changes, and there is no fallback to the baseline checkout when the
workspace is missing. Symlinks/reparse points are skipped and recorded,
never followed; `.git`, root `.harness/`, and root `harness.toml` are
excluded and recorded; everything else is copied and hashed, and the
destination identity is recomputed and compared before any test runs.
`Test-CopyFixtures.ps1` exercises this standalone (changed file, untracked
file, nested file, deletion absence, `.harness`/`harness.toml` exclusion,
symlink skip, manifest count, identity recomputation, tamper evidence) and
exits nonzero on any failure.

The evaluation acceptance route is NOT the PS self-only copy (which
verified the copy against itself and could admit later source edits).
`scripts/evaluate-v1.ps1 -Action Evaluate` requires an explicit
`-CandidateCopyExe` Go helper (`evals/v1/candidatecopy`, package main,
imports only existing `internal/control`, `internal/worktree`,
`internal/safepath`, `internal/canonical`). The helper decodes the actual
inspector snapshot (bounded, existing types), requires `Candidate.ID()` to
equal the reviewed digest and `verification.plan.candidate_id`, validates
READY/non-pending evidence and exact workspace binding, holds
`worktree.AcquireRead(snapshot.Workspace.Request)` across Capture, copy,
and final fingerprint, requires actual `Candidate == snapshot.Candidate`
before and after, copies precisely the captured `FileStates` via
`os.Root` + `safepath.CopyRegular` (64 MiB/file, 512 MiB total, 4096
files), preserves executable mode, rejects nesting (both directions) and
symlinked parents, never deletes computed paths (failed copies retained as
BLOCKED), verifies the destination `FilesID` manifest equals the source
`FilesHash` with per-file hash+mode comparison, and emits small JSON
`candidate_id/files_hash/file_count/destination/workspace/worktree_id`.
The runner calls it for BOTH Native and PR5Matched with the reviewed
candidate, treats helper failure as BLOCKED before any tests, reuses the
destination manifest identity, validates the expected ID in `Get-RunGate`
(including native diff agreement), runs native and held-out tests only on
copied bytes (held-out added only to the copy, then removed), and leaves
the original untouched. `go test ./evals/v1/candidatecopy` demonstrates
changed-source/incorrect-candidate reject, dirty/untracked/delete capture,
nesting/symlink reject, writer-lease contention, and wrong-ID/manifest
reject with real worktree fixtures and no providers.
