# Autonomous task run (native guide)

Scope: run one bounded autonomous coding task with the native `fabric` CLI.
Task graphs, bounded parallel exploration and configured adaptive routing are
available. This guide does not claim full v1 completion or release qualification.

New real-model initialization selects the `anchored-edits-v1` writer contract:
proposals replace exact, unique text fragments in existing files. Fabric checks
the original file hash and preserves bytes outside the selected fragments.
Overlapping, absent or ambiguous anchors are rejected. New files require
explicit content; this contract does not implicitly delete files. Graph writers also
receive the currently ready implementation task and its declared write paths;
repair instructions cannot grant additional file ownership. The output schema
binds the exact candidate ID. Existing `utf8-v2`, `utf8-replace-v3` and
`utf8-scoped-v4` runs keep their frozen instructions. Existing file and proposal
size limits still apply. Native checks and review remain required; valid edits
alone do not prove that the task was solved.

New initialization also selects the `json-v2` explorer contract. Its schema
restricts `candidate_id` to the exact invocation-bound ID, and replay checks the
same identity independently. Existing `json-v1` runs keep their original schema;
a rejected or uncertain old result is never normalized into a successful one.

## 0. Prerequisites

- A Codex account with local authentication: run `codex login` first so the
  auth file exists (default `~/.codex/auth.json`, or `$CODEX_HOME/auth.json`).
  `fabric init` only binds the auth path; it never reads credential contents.
- A Go toolchain, and a committed Git repository for the task (real-model
  `init` requires one).

Fabric Git operations resolve Git from conventional system installation paths instead
of searching `PATH`. For a nonstandard installation, set
`FABRIC_GIT_EXECUTABLE` to the absolute path of the Git executable you trust.
An invalid override fails explicitly. The shared resolver applies to diff,
repository, worktree, local-store and push operations.

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

Optional `init` flags `--writer-model`, `--writer-effort`, `--reviewer-model`
and `--reviewer-effort` configure those roles independently. Omitted values
inherit `--model` and `--effort`; planner and explorer keep the base profile.
This uses explicit model choices and does not claim an automatic cost or quality
improvement. Model availability is determined by your authenticated runtime.

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

### Inspect options before dispatch (v1.1 development)

`doctor` reports configured roles, runtime support and verification executable
readiness. To inspect the effective autonomous options, use the same run flags
with `--inspect-plan` and omit the objective:

```sh
fabric run --autonomous --inspect-plan --max-parallel 3
fabric run --autonomous --inspect-plan --parallel-writers --auto-compact-token-limit 32000
```

This reads configuration, repository identity and executable pins; it creates
no run and makes no provider calls. The report includes selected context,
writer topology, cache and compaction settings, resource estimates and admitted
capability fallbacks. `AVAILABLE_NOT_RUN` means the executable was resolved and
hashed, not that verification passed. Provider qualification remains `NOT_RUN`.
Memory observations report `OBSERVED` or `UNAVAILABLE`; unavailable observations
do not mean zero free memory. With isolated-writer options, the report includes
an `ESTIMATED_NOT_ADMITTED` future-worker preview using the supplied per-worker
estimate. It is neither a worker memory measurement nor a reservation. Actual
future claims recheck pressure; an observation does not cancel admitted work.

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

### v1.1 development behavior

For configuration v1, a Codex read-only explorer rejected with a sealed
`serverOverloaded` failure can continue on `resume --autonomous RUN`. Fabric
retains that failed invocation and admits at most one new invocation for the
same graph task. Missing terminal evidence remains UNKNOWN and is never resent.
Failed invocation usage remains incomplete; it is not counted as successful
exploration. Configuration v2 retains its model-access settlement rules.

New autonomous runs admit at most two separate output-correction turns per
role. Each correction has a new invocation identity and retains the exact
completed predecessor receipt and rejected output. An unapplied invalid anchor
can be corrected only against captured, hash-verified candidate source; a wrong
candidate or preimage hash remains an authority failure. Output correction does
not consume the code-repair allowance.

When explorer and writer roles are configured, new runs admit at most two
candidate-bound ownership replans. Serial and parallel/isolated runs use their
respective scope policies. A read-only design must approve the exact additional
paths within the original immutable scope. Completed proposals may be retained;
integration freshly prepares a parent file effect and keeps its normal
authorization and verification requirements. Existing runs retain the policy
recorded at creation.

`NEEDS_ATTENTION` retains useful work when the current budget, accounting or
environment prevents progress. Required usage that is missing after a known
completed turn preserves the semantic result and its unsettled reservation;
new provider calls wait for usage reconciliation. Missing usage and cost remain
unknown. `NEEDS_REPLAN` retains the candidate when a legal ownership revision
has not completed. Neither outcome means acceptance. `UNSAFE` identifies an
authority, identity or history failure; `PAUSED` requires a settled pause.

### Explained runtime stops

Failure JSON retains `run_id`, `phase`, `blocked_reason` and `next_action`.
For synchronous OpenCode failures, `runtime_diagnostic` adds a bounded stage
and code: for example `response_decode` / `validation_or_runtime_failure`.
The runtime's first diagnostic is also retained beside its journal as
`<runtime journal>.failure.json`. These labels contain no provider response,
prompt or credentials and never authorize a retry or validate an output.

An exactly sealed initial Codex planner `serverOverloaded` refusal is reported
as `planner_capacity_refused` / `NEEDS_ATTENTION`. It contains no accepted plan;
wait for capacity before explicitly starting a new run. The failed invocation
is preserved. Incomplete terminal evidence or genuinely uncertain provider
effects remain UNKNOWN and must be reconciled without resending.

These development behaviors require final release qualification; the published
release package is qualified separately from later development commits.
