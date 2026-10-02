# Fabric

A standalone engineering harness around coding agents: models supply judgment;
deterministic code binds repository context, routing, approvals and outcomes.
Go owns orchestration. Rust owns immutable repository intelligence.

The local implementation includes deterministic planning and effect journals,
isolated writer workspaces, approved file changes, verification execution, a Codex
planning/source-tool adapter and Rust SCIP intelligence. Explicit producer, import
and local publication lifecycles feed immutable runtime queries. Seven optional
engineering procedures accompany the code.

The bounded writer/reviewer model workflow (plan, implement, verify, review,
commit) is qualified end to end at the v0.0.1 trusted bootstrap checkpoint
(see below); the full product v1.0.0 program remains future work. Local
fixtures and bounded authenticated planning tests do not establish model
quality or OS sandboxing.

## Run your first real coding task

The [native autonomous task guide](docs/guides/autonomous-task.md) introduces
`fabric run --autonomous`: plan, explore, implement, verify and review with a
bounded repair budget and durable result inspection. A
[real repository pilot](docs/evaluation/native-autonomous-pilot.md) reached
`READY` and passed independent hidden checks. The
[task graph guide](docs/guides/graph-autonomous-task.md) covers dependency-based
decomposition, bounded parallel read-only agents, scoped implementation and
configured adaptive model allocation. Full v1 release acceptance remains open.

Follow the [Windows first-task guide](docs/getting-started/real-task.md) to build
Fabric, configure a stock Codex model, and run a coding task through planning,
delegated inspection, implementation, verification and a separate review.
The guide shows plan/file approvals and durable result inspection. This path has
passed a real-model acceptance run from a fresh GitHub checkout at the exact
source recorded in the [acceptance evidence](docs/evaluation/v1-product-gate.md).

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

The [local Codex integration](integrations/codex/engorch/README.md) packages this
workflow as a thin skill. Codex is the intended primary interface; the same CLI
remains usable independently. Plugin installation and sandbox inheritance have
not yet been qualified.

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
