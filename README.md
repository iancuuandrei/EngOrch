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

## Checkpoint

- [v0.0.1 trusted bootstrap checkpoint](docs/evaluation/v0.0.1.md): first
  trusted pre-alpha checkpoint (M2at: Luna planning, recursive Muse research,
  Muse writer, deterministic verification, Luna review, local commit).

## Try a local plan

Build from source with Go 1.27.1 and Git installed:

```sh
go build -o bin/harness ./cmd/harness
```

Use a repository with an existing commit. `init` writes `harness.toml` without
overwriting an existing file. Edit its repository name and required checks.

```sh
harness --root PATH_TO_REPOSITORY init
```

```sh
harness --root PATH_TO_REPOSITORY plan "Add a tested greeting"
```

For an existing goal document, use
`harness --root PATH_TO_REPOSITORY plan --file goal.md`. Relative goal paths resolve
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
- [Local task schedules](docs/guides/task-schedules.md)
- [Autonomous task graphs](docs/guides/graph-autonomous-task.md): dependency
  waves, bounded context and parallel read tasks; development qualification
  remains in progress.
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
