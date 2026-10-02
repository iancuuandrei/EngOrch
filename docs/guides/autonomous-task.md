# Autonomous task run (native guide)

Scope: run one bounded autonomous coding task with the native `fabric` CLI.
Graph, parallel and adaptive routing are not integrated yet. This guide does
not claim full v1 completion, and no test run is claimed here.

## 0. Prerequisites

- A Codex account with local authentication: run `codex login` first so the
  auth file exists (default `~/.codex/auth.json`, or `$CODEX_HOME/auth.json`).
  `fabric init` only binds the auth path; it never reads credential contents.
- A Go toolchain, and a committed Git repository for the task (real-model
  `init` requires one).

`fabric diff` resolves Git from conventional system installation paths instead
of searching `PATH`. For a nonstandard installation, set
`FABRIC_GIT_EXECUTABLE` to the absolute path of the Git executable you trust.
An invalid override fails explicitly. Other Git operations retain their existing
executable lookup behavior; this setting currently applies to diff observations.

## Windows PowerShell happy path

With Go 1.27.1, Git and a stock Codex executable installed:

```powershell
git clone https://github.com/iancuuandrei/Fabric.git Fabric
Set-Location Fabric
go build -o fabric.exe ./cmd/fabric
$fabric = (Resolve-Path ./fabric.exe).Path
$codex = 'C:\path\to\codex.exe' # Set the installed executable path.
& $codex login
Set-Location 'C:\path\to\your\committed-project'
Add-Content .git/info/exclude "`n.harness/`nharness.toml`n"
& $fabric init --codex $codex --model gpt-6-luna --effort high
& $fabric doctor
# Set harness.toml verification argv to this project's required checks.
& $fabric run --autonomous 'Fix the parser bug and add regression tests.'
& $fabric status
& $fabric inspect
& $fabric diff
```

Inspect reports the run ID and isolated workspace path. Use that explicit ID
with `usage`, `diff`, or `resume --autonomous`; copying or committing the result
is an operator action. The example repository and Codex paths are local choices.

## 1. Clone and build

```sh
git clone https://github.com/iancuuandrei/Fabric.git fabric-v1-product
cd fabric-v1-product
go build ./cmd/fabric
export PATH="$PWD:$PATH"
```

This produces the `fabric` binary (`fabric.exe` on Windows). All commands below
run as `fabric ...` from the task repository root.

## 2. Configure the project

```sh
cd <your committed repository>
fabric init --codex /absolute/path/to/codex --model <model>
fabric doctor
```

Useful flags: `--effort medium|high`, `--auth-source PATH` (non-default
auth file), `--state-root PATH` (private runtime state; must be outside the
repository). `init` never overwrites an existing `harness.toml`.

## 3. Ignored outputs and local state

Keep machine-local state out of commits with the explicitly local
`.git/info/exclude` (never committed, unlike `.gitignore`):

```sh
printf '/.harness/\nharness.toml\n' >> .git/info/exclude
```

`.harness/` holds run journals, isolated worktrees and lease state. Do not
commit it. `fabric diff` likewise excludes ignored files without reading them.

## 4. Verification project policy

Edit the `[[verification]]` checks in `harness.toml` to the project's real
checks (the default is `go test ./...`):

```toml
[[verification]]
name = "unit"
argv = ["go", "test", "./..."]
timeout_seconds = 120
```

Verification admits READY only when every required check passes on the
unchanged candidate.

## 5. Run a real objective

```sh
fabric run --autonomous "Fix the off-by-one in the pager offset and cover it with a unit check."
fabric run --autonomous --file goal.md
fabric run --autonomous --max-repairs 3 "Fix the off-by-one in the pager offset and cover it with a unit check."
```

`--max-repairs` is 0-8 (default 2). Objectives and goal files must be
nonempty UTF-8 of at most 256 KiB; invalid input is rejected before any
durable run is created.

## 6. Resume, bounded repairs and status

```sh
fabric resume --autonomous [RUN]
fabric status
fabric inspect RUN
fabric usage RUN
fabric diff [RUN]
```

An omitted RUN on `resume --autonomous` selects the latest validated
autonomous run by modification timestamp. That is a documented convenience,
not a framework: on timestamp ties, or in scripts, pass RUN explicitly.
`diff` shows the tracked HEAD diff plus non-ignored untracked files with
coverage metadata. Results land in the isolated workspace at
`.harness/worktrees/<RUN>/` (default layout).

## 7. Outcomes

- READY means the local candidate passed verification (and review when a
  reviewer is configured). Nothing is published: commit and push are separate
  explicit commands.
- A blocked run prints a coarse JSON summary on stdout and a sanitized
  boundary error on stderr; provider detail stays out of logs.
- UNKNOWN workspace, file or commit outcomes mean stop: run
  `fabric reconcile RUN` to observe them without retrying writes, and never
  resend uncertain work.
