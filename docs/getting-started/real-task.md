# Run your first real coding task

Use Windows PowerShell 7, Git, Go 1.27.1 or newer, and a stock
[Codex CLI](https://developers.openai.com/codex/cli). Put Git and Go on PATH.
Sign in with `codex login` using an account that can access the chosen model.
Fabric uses that existing authentication file; credentials are not copied into
the project or its journals. A model task consumes your account's usage.

## Build Fabric and create a project

```powershell
git clone https://github.com/iancuuandrei/Fabric.git
Set-Location Fabric
go build -o bin/engorch.exe ./cmd/harness
if ($LASTEXITCODE -ne 0) { throw 'Fabric build failed' }
$fabricRoot = (Get-Location).Path
$fabric = Join-Path $fabricRoot 'bin/engorch.exe'

$project = Join-Path (Split-Path $fabricRoot) 'fabric-first-task'
if (Test-Path $project) { throw 'Choose a new project directory' }
New-Item -ItemType Directory -Path $project | Out-Null
Copy-Item "$fabricRoot/examples/first-task/*" $project -Force
Copy-Item "$fabricRoot/examples/first-task/.gitignore" "$project/.gitignore" -Force
git -C $project init
git -C $project add .
git -C $project commit -m 'Initial greeting project'
```

Git needs your normal author name/email configured for that initial commit.
The sample starts with a greeting function and one test. Its ignore file excludes
Fabric configuration and runtime state. In your own project, commit existing work
first and ignore `.harness/` and `harness.toml` before creating a task.

## Configure a real model

Select a model available to your Codex account. The Windows acceptance run uses
`gpt-6.1-sol` with low effort for separate planner, explorer, writer and reviewer
invocations; select another supported identifier explicitly if needed.

```powershell
$codex = (Get-Command codex -CommandType Application).Source
& $fabric --root $project init --codex $codex --model gpt-6.1-sol --effort low
if ($LASTEXITCODE -ne 0) { throw 'Fabric initialization failed' }
& $fabric --root $project doctor
if ($LASTEXITCODE -ne 0) { throw 'Fabric configuration is invalid' }
```

This creates all four role settings and pins the executable's hash. Runtime state
defaults to your local application cache outside the project. Authentication is
located through `CODEX_HOME`, or your user `.codex/auth.json`. Use
`--auth-source PATH` or `--state-root PATH` to select another location. `init`
never overwrites an existing configuration and does not send a model request.
`doctor` validates local configuration/repository identity; it does not prove
model availability or authenticate a model turn.

The generated verification check is `go test ./...`; edit it to match your own
project before planning. Build commands should explicitly put their outputs
outside source paths if those outputs would otherwise change the candidate.
Changing configuration does not change an already created run's settings.

## Submit the task and inspect approvals

```powershell
$result = & "$fabricRoot/scripts/run-task.ps1" `
  -Fabric $fabric -Repository $project `
  -Objective 'Update Hello(name) to trim leading and trailing whitespace, use world when the name is empty or whitespace-only, and add table-driven regression tests for normal, padded, empty, and whitespace-only inputs. Preserve the existing greeting format. Keep the implementation small and standard-library-only.'
```

Read the plan and type `yes` to approve it. Fabric creates an isolated Git worktree,
delegates source inspection to an explorer, and asks the writer for file changes.
Read the displayed file paths, previous hashes and proposed contents, then type
`yes` to apply them. The script executes the configured project checks and invokes
a separate reviewer bound to the checked candidate. The original project checkout
is unchanged; the result remains in the worktree shown by `$result.workspace`.

If you decline approval, the script returns the run ID and the required action.
Continue that same run with `-RunId $result.run_id` instead of `-Objective`.
For unattended, preauthorized tasks, `-ApprovePlan -ApproveChanges -NonInteractive`
explicitly accepts the displayed plan and changes. Without those approval switches,
`-NonInteractive` stops at the corresponding gate.

The script stops on failed checks, requested changes, interrupted attempts and
uncertain effects. It does not start a repair loop or retry model requests.
Inspect the exact run before deciding how to proceed.

## Inspect the result and durable evidence

```powershell
git -C $result.workspace diff
git -C $result.workspace status --short
& $fabric --root $project inspect $result.run_id
& $fabric --root $project usage $result.run_id
& $fabric --root $project inspect $result.run_id --export-jsonl
```

`inspect` shows the plan, explorer/writer/reviewer results, candidate identity,
file-effect outcome and verification observations. `usage` reports available
runtime accounting; missing monetary cost remains unknown. `--export-jsonl`
replays and exports the validated event history. Run IDs and evidence are durable:
after reopening PowerShell, use the same binary, project path and run ID to read
them again. Default controller journals are under `.harness/runs/`; Codex runtime
journals are under the state root recorded in `harness.toml`.

The isolated result can be reviewed and integrated using your normal Git workflow.
The task script does not commit, push, open a PR or publish a release. A `READY`
result establishes the executed checks and recorded model review; it does not
establish general correctness or an operating-system security sandbox.
