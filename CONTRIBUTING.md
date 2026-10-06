# Contributing

Read [AGENTS.md](AGENTS.md) for repository navigation, then the
[source-history policy](docs/contributing/source-history.md) before committing
or publishing. It defines branch, version-title and automation identity rules.
Go 1.27.1 and the repository's pinned Rust tooling are the documented toolchains.

## Source publication

Follow the illustrated [source-history policy](docs/contributing/source-history.md)
for development branches, rolling snapshots and immutable releases.

Review changes and complete applicable requested verification before pushing
dev. Direct human/Codex edits are supported; self-hosting is optional dogfooding.
Preserve unrelated changes, immutable policies and external-effect ownership.

## Product focus and agent tooling

Ship useful capabilities and harden demonstrated failures. Self-hosting is
optional dogfooding; direct Codex/human product edits are supported. Bound
infrastructure investigations to data corruption, duplicate external effects
or demonstrated normal-use blockers. Avoid governance-only work packages and
repeated benchmarks. The [roadmap](docs/roadmap/v2-product-plan.md) records the
current product priorities.

Use subagents only when explicitly requested or authorized for delegation.
For implementation, independent review, and repair, use Muse Spark 1.3 Contributor High through OpenCode, exact route opencode-go/muse-spark-1.3-contributor with --variant high; do not use subscription Luna agents for these roles. Give writers explicit ownership
and warn that other agents share the checkout. An uncertain route failure never permits repeating an effect or claiming success. If this route is unavailable, report the blocker rather than silently switching model or repeating an uncertain effect.

## Formatting and verification

Format changed code. When verification is requested, select checks that exercise
the changed behavior and report the exact executed scope. The standard Go
commands are:

```sh
go fmt ./...
```

```sh
go test ./...
```

```sh
go vet ./...
```

```sh
go test -race ./...
```

The race detector needs a supported C toolchain. Report unavailable checks as
NOT_RUN. Test output is evidence of the executed scope, never provider or hosted
qualification. Add fault cases for durable state and authority boundaries.

Follow the [documentation standard](docs/contributing/documentation-standard.md).
Use ADRs for consequential decisions and update specifications with behavior.
Keep each change independently reviewable. Before calling it complete,
challenge ownership, unnecessary abstractions, duplicated authority, explicit
errors, reproducible failure, uncertainty and performance regressions.

Historical Muse recursive qualification uses a separately provisioned pinned
OpenCode binary and S0 checkout. Its offline scaffolding preflight is opt-in with
`ENGORCH_MUSE_RECURSIVE_PREFLIGHT=1`; live qualification has its own existing
`ENGORCH_MUSE_RECURSIVE_LIVE` gate. Default source tests do not establish either
qualification. Product first-run acceptance uses the stock Codex route described
in the [real-task guide](docs/getting-started/real-task.md).

Regenerate [CLI reference](docs/reference/cli.md) from `harness reference`; preserve
UTF-8 and LF. Test fixtures create temporary independent repositories and never
change the developer's checkout. [The documentation hub](docs/README.md)
separates current guides, contracts, decisions and retained evidence.
Security limitations are documented in the
[trust boundary](docs/security/trust-boundary.md).
