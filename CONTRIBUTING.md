# Contributing

Read [AGENTS.md](AGENTS.md) first. Fabric development happens on `dev`; keep its
complete incremental history. Use patch-version commit titles such as
`v1.1.3: describe the change` and the prescribed automation author/committer.
Go 1.27.1 and the repository's pinned Rust tooling are the documented toolchains.

## Source publication

Trusted dev pushes synchronize a rolling major/minor view on `main`, such as
`v1.1`. The workflow rewrites only the current line's snapshot and preserves
its original dates. A newer line adds one checkpoint. Published tags, packages
and evidence remain immutable. See the illustrated
[source-history policy](docs/contributing/source-history.md).

Review changes and complete applicable requested verification before pushing
dev. Direct human/Codex edits are supported; self-hosting is optional dogfooding.
Preserve unrelated changes, immutable policies and external-effect ownership.

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
