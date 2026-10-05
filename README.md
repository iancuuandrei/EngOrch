# Fabric

Fabric is an engineering agent runner. It binds model work to a repository,
candidate identity, configured checks, review and durable run history. Go owns
orchestration; Rust provides repository intelligence.

Fabric v1.0.0 has a reproducible Windows amd64 distribution and a recorded
installed-binary acceptance run. Start with the
[Windows installed-task guide](docs/getting-started/installed-autonomous-task.md)
for download, setup, a bounded autonomous task and result inspection. The
[release guide](docs/guides/release.md) describes the accepted package and its
limits. Linux amd64 remains optional and unqualified.

For the current development checkout, use the
[source-build autonomous guide](docs/guides/autonomous-task.md).
The [capability guide](docs/guides/features.md) describes the current
functionality and its limits. The [v2 product roadmap](docs/roadmap/v2-product-plan.md)
records qualification separately from published release acceptance.

## Current development status

**v1.1.0 adds resilient autonomous execution.** The changes
add bounded semantic corrections and ownership replanning, typed autonomous
outcomes, progress-aware execution, optional context/metadata degradation,
future-worker memory admission and an effective read-only doctor plan.
Writer and fixer prompts also require preservation of existing behavior and
focused regression tests around the behavior being changed.
Completed semantic results are preserved when accounting is unavailable;
required accounting still gates subsequent model calls. See the
[v1.1 implementation and release ledger](docs/roadmap/v1.1-resilience.md)
for development evidence and release checks; published qualification is recorded
separately with the GitHub release. Further comparative performance
benchmarking is deferred so this increment can ship.

The latest clean-source evaluation has **8/8 accepted task coverage**, with one
explicit fresh `logr` successor after a sealed provider capacity refusal. The
original cohort remains 7/8; all eight accepted candidates passed native checks,
independent review and unchanged heldouts. See the
[eight-task closure and diagnostic limits](docs/evaluation/v112-eight-task-closure.md).

**v1.0.29** adds Muse Go execution fixes and graceful optional-capability
fallbacks. The prior qualified capability checkpoint is v1.0.27 (`6205de9`);
v1.0.28 updates documentation only. A separate frozen
v1.0.25 Sol High cohort reached **7/8 PASS**: fixed-six **6/6** plus generated
numeric ownership; godotenv remains BLOCKED after two repairs. The seven PASS rows
have native verification, approved candidate-bound review and unchanged
held-out acceptance. They do not establish v2.0.0 release completion.

The earlier persistent v2 development goal remains paused. Separately authorized
Muse route work completed one local go-humanize coding task using exclusively
Muse Spark 1.3 Contributor on OpenCode Go: native tests, candidate-bound review
and unchanged held-out tests passed. This used a custom development binary and
does not qualify the entire task suite or a new release. See the
[route guide](docs/guides/muse-go-route.md) and
[capability fallback guide](docs/guides/graceful-capabilities.md).
See the
[latest checkpoint](docs/roadmap/v2-product-plan.md) for the preserved run
state, measured token types, cache limitations and remaining release work.

## Run your first real coding task

The installed-task guide uses `fabric run --autonomous` to plan, explore,
implement, verify and review with a bounded repair budget. The result stays in
an isolated Git worktree for operator inspection and integration; Fabric does
not commit or publish it automatically. See the
[v1.0.0 release acceptance record](docs/evaluation/v1-release-acceptance.md)
for the exact distribution and installed-run evidence. The
[task graph guide](docs/guides/graph-autonomous-task.md) documents dependency
decomposition, bounded parallel read-only work and configured model allocation.

The earlier [source-built Windows first-task guide](docs/getting-started/real-task.md)
covers the approval-based workflow accepted at its recorded source identity.
Use the installed-task guide above for the v1.0.0 autonomous workflow.

## Historical bootstrap

- [v0.0.1 trusted bootstrap checkpoint](docs/evaluation/v0.0.1.md): first
  trusted pre-alpha checkpoint (M2at: Luna planning, recursive Muse research,
  Muse writer, deterministic verification, Luna review, local commit).

## Try a local plan

Build from source with Go 1.27.1 and Git installed:

```sh
go build -o bin/ ./cmd/fabric
```

Add the generated `bin` directory to your shell's `PATH` before the commands
below. The executable is `fabric.exe` on Windows and `fabric` on other systems.

Use a repository with an existing commit. `init` writes `harness.toml` without
overwriting an existing file. Edit its repository name and required checks.

```sh
fabric --root PATH_TO_REPOSITORY init
```

```sh
fabric --root PATH_TO_REPOSITORY plan "Add a tested greeting"
```

For an existing goal document, use
`fabric --root PATH_TO_REPOSITORY plan --file goal.md`. Relative goal paths resolve
against the selected repository. The file must contain nonempty UTF-8 text of at
most 256 KiB; its exact text is retained in the run's immutable inputs.

The result is canonical JSON in `AWAITING_APPROVAL`, with exact run and plan IDs.
The fake plan exercises protocol mechanics; it is not an evaluated model answer.
See the [local planning guide](docs/getting-started/local-plan.md) for approval,
replay and failure behavior.

The [source-build autonomous guide](docs/guides/autonomous-task.md) remains
available for development checkouts. The [local Codex integration](integrations/codex/engorch/README.md)
packages a separate workflow as a thin skill; this release acceptance does not
qualify plugin installation or sandbox inheritance.

## Architecture and development

- [Fabric v1 product gate](docs/development/v1-product-gate.md): the Windows
  first-task journey is accepted; broader v1 work remains open. Fabric
  self-hosting is optional dogfooding.

- [Architecture](docs/architecture/system.md) and [ADRs](docs/adr/0001-language-split.md)
- [Journal contract](docs/specifications/run-journal.md)
- [CLI reference](docs/reference/cli.md)
- [All implemented capabilities and qualification](docs/guides/features.md)
- [Engineering ranking and semantic queries](docs/guides/engineering-orientation.md)
- [Go file facts](docs/guides/go-file-facts.md)
- [Resource-bounded isolated writers](docs/guides/isolated-writers.md)
- [Local task schedules](docs/guides/task-schedules.md)
- [Autonomous task graphs](docs/guides/graph-autonomous-task.md): dependency
  waves, bounded context and parallel read tasks; development qualification
  remains in progress.
- [Impact-aware review context](docs/guides/reviewer-impact-context.md)
- [Prompt-cache recipes](docs/guides/prompt-cache-recipes.md)
- [Empirical model calibration and conservative fallback](docs/guides/empirical-model-calibration.md)
- [Native runtime compaction](docs/guides/native-auto-compaction.md)
- [Candidate-bound checkpoints](docs/guides/checkpoints.md)
- [Verified fresh-context rounds](docs/guides/verified-fresh-rounds.md)
- [Deterministic Go-format observations](docs/guides/deterministic-go-format-observation.md)
- [Typed Codex usage observations](docs/guides/codex-native-usage-observations.md)
- [Apply file changes](docs/guides/file-changes.md)
- [Build and verify a local package](docs/guides/local-packaging.md): local directory only; never a signed, tagged, or published release.
- [Reproducible release bundles and installation](docs/guides/release.md):
  Windows build, integrity verification and installation tooling; a generated
  bundle does not by itself establish release qualification.
- [Real-repository evaluation suite](evals/v1/README.md): pinned tasks and
  candidate-bound hidden acceptance checks, including the PR #5 comparison.
- [Optional engineering procedures](docs/guides/procedures.md)
- [Contributing](CONTRIBUTING.md) and [documentation standard](docs/contributing/documentation-standard.md)
- [Research provenance](docs/research/oss-mechanisms.md)
- [Current evidence](docs/evaluation/status.md)
- [Real repository comparison](docs/evaluation/v1-real-repository-comparison.md):
  measured Native/PR #5 outcomes, real concurrency and observed repair limitations.
- [WP05/WP06 program history and completion status](docs/evaluation/wp05-wp06-history.md)
- [v0.0.1 trusted bootstrap checkpoint](docs/evaluation/v0.0.1.md)
