# Install Fabric and run an autonomous task (Windows)

This guide documents the Fabric v1.0.0 Windows amd64 installation path. The
release acceptance tested this package on Windows; Linux amd64 is not yet
qualified. The Windows package is published in the v1.0.0 GitHub release.
You do not need Go or Rust to install Fabric. Install Git, the
[Codex CLI](https://developers.openai.com/codex/cli), and the toolchain required
by your project's configured verification checks. For a Go project, install
Go 1.27.1 or newer.

## Download and verify the package

After publication, download `fabric_1.0.0_windows_amd64.zip` from the
[Fabric v1.0.0 GitHub release](https://github.com/iancuuandrei/Fabric/releases/tag/v1.0.0).
Check its SHA-256 before extracting:

```powershell
$releaseDir = Join-Path $HOME 'Downloads\Fabric-v1.0.0'
$null = New-Item -ItemType Directory -Path $releaseDir -Force
$archive = Join-Path $releaseDir 'fabric_1.0.0_windows_amd64.zip'
$expected = '7add99a1146bfe8a780dd38238904cf74f9fdb7dffd1a36e5211d045988a21e6'
$actual = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()
if ($actual -ne $expected) { throw "Fabric archive checksum mismatch: $actual" }
```

This is an unsigned release. A matching checksum detects a changed or
incomplete download when compared with this published value; it is not a
cryptographic publisher signature. The archive contains a manifest and
per-component hashes; the release verifier checks those package identities.

Install into a new directory, add it to this PowerShell session's `PATH`, and
confirm the reported build identity:

```powershell
$installDir = Join-Path $HOME 'Tools\Fabric-v1.0.0'
if (Test-Path -LiteralPath $installDir) { throw 'Choose a new install directory' }
$null = New-Item -ItemType Directory -Path (Split-Path $installDir -Parent) -Force
New-Item -ItemType Directory -Path $installDir | Out-Null
Expand-Archive -LiteralPath $archive -DestinationPath $installDir
$env:PATH = "$installDir;$env:PATH"
& (Join-Path $installDir 'fabric.exe') version
```

Keep `fabric.exe` and `engorch-ri.exe` together in the install directory.
See the [release guide](../guides/release.md) for build and package details.

## Sign in and initialize your repository

Sign in with the Codex account that can use the selected model. Authentication
stays in the local Codex profile; Fabric does not copy credentials into your
repository or run journal. Model calls consume account usage.

```powershell
codex login
$codex = (Get-Command codex -CommandType Application).Source
Set-Location 'C:\src\your-project'
git status --short
```

Start from a committed baseline. Review or commit your own pending changes before
running a task. Keep Fabric's configuration and durable workspace out of
version control. Add local-only exclusions (these do not change the shared
`.gitignore`):

```powershell
$exclude = Join-Path (git rev-parse --show-toplevel) '.git\info\exclude'
@'
/.harness/
harness.toml
'@ | Add-Content -LiteralPath $exclude
```

Initialize Fabric with the stock Codex executable and the model used in the
Windows release acceptance:

```powershell
$fabric = Join-Path $installDir 'fabric.exe'
& $fabric init --codex $codex --model gpt-6-luna --effort high
if ($LASTEXITCODE -ne 0) { throw 'Fabric initialization failed' }
& $fabric doctor
if ($LASTEXITCODE -ne 0) { throw 'Fabric configuration is invalid' }
```

`init` creates `harness.toml` and does not dispatch a model request. `doctor`
checks local configuration and repository binding; it does not authenticate a
model turn. Confirm that the configured verification command is appropriate for
this repository before starting a task. For a Go project, a common check is
`go test ./...`; use the project's actual required checks and install their
toolchains.

## Run, inspect and integrate

Write a specific objective with observable behavior and tests. The repair
budget is bounded: `2` is the default, and the CLI accepts values from 0 to 8.

```powershell
$objective = 'Fix the documented parser edge case, add a regression test, and preserve the existing public API.'
& $fabric run --autonomous --max-repairs 2 $objective
if ($LASTEXITCODE -ne 0) { Write-Warning 'The run did not finish READY; inspect its durable state before acting.' }
```

Record the `run_id` shown in the result and inspect that exact candidate:

```powershell
$run = 'PASTE_RUN_ID_HERE'
& $fabric status
& $fabric inspect $run
& $fabric diff $run
& $fabric usage $run
```

`READY` means configured verification passed and the recorded reviewer approved
that candidate. Review the isolated worktree diff and project status before
integrating changes into your branch. Fabric does not commit, push, open a pull
request or publish the result automatically. Inspect and reconcile blocked or
uncertain results before any retry; do not resend work with an unknown effect.

The isolated workspace and local run state remain available for inspection.
Keep `.harness/` and `harness.toml` local; do not commit them. The Git worktree
isolates the task candidate from edits to the original checkout, but it is not
an operating-system security sandbox. Fabric executes the checks configured by
the project. The release acceptance does not guarantee that arbitrary
model-generated tasks will succeed.

## What the v1.0.0 acceptance establishes

The accepted installed run is bound to Windows amd64 package v1.0.0, reached
`READY`, received reviewer approval with zero findings, and passed independent
native and held-out checks. In that run, an explorer made four successful
searches through the installed RI binary; three returned validated nonempty
results. The evidence binds those calls to installed RI SHA-256
`65123be724fa8e0388747897cee5dc3cf889881e1c1535404e36eba3ed88ebe0`.
This is evidence for the recorded task and configuration, not general
qualification of every model role, repository, Linux installation or task.
See the [v1.0.0 acceptance record](../evaluation/v1-release-acceptance.md) for
the source, archive, run and evidence identities.

## Reproduce the installed acceptance

The acceptance driver is in
[`evals/v1/harness/accept-installed-humanize.ps1`](../../evals/v1/harness/accept-installed-humanize.ps1).
Run it with PowerShell 7 (`pwsh`) on Windows. It accepts explicit paths for the installed binaries, a separate clean Fabric
source checkout at the v1.0.0 source commit, a clean pinned
`go-humanize` clone, Codex, Go, the prebuilt candidate-copy helper and a new
external evidence directory. It prepares the exact manifest task, publishes the
immutable lexical base before resuming, then gates the reviewed candidate and
runs native and held-out checks on separate bound copies. The model task may
modify its isolated candidate workspace; acceptance tests run on copies and do
not edit the reviewed candidate. Its preflight mode validates paths and
identities without initializing a run or calling a model.

Use the checkout containing this driver as `$driverCheckout`. Create a second,
clean detached source checkout at the release source commit; the helper,
manifest, and held-out fixture must come from that pinned source checkout. Build
the helper outside both checkouts and the task repository, prepare a fresh clone
at the manifest pin, and run the driver first in preflight mode:

```powershell
$driverCheckout = 'C:\src\Fabric'
$fabricSource = 'C:\src\Fabric-v1.0.0-source'
git clone https://github.com/iancuuandrei/Fabric.git $fabricSource
git -C $fabricSource checkout --detach a4ca05692465bedc6b98d6ffbb8219b6d076f5d1
if ((git -C $fabricSource status --porcelain=v1 --untracked-files=all) -join "`n") { throw 'Fabric source checkout is not clean' }
$go = (Get-Command go -CommandType Application | Select-Object -First 1).Source
$codex = (Get-Command codex -CommandType Application | Select-Object -First 1).Source
$pwsh = (Get-Command pwsh -CommandType Application | Select-Object -First 1).Source
$helper = Join-Path $env:TEMP 'fabric-candidatecopy.exe'
& $go -C $fabricSource build -trimpath -o $helper ./evals/v1/candidatecopy
if ($LASTEXITCODE -ne 0) { throw 'Candidate-copy helper build failed' }

$repo = 'C:\src\go-humanize-v1-acceptance'
git clone --no-checkout --filter=blob:none https://github.com/dustin/go-humanize.git $repo
git -C $repo checkout --detach a1b4e66b9a6d890e9e15e7091cf16c8032367d6e
$evidence = Join-Path $env:TEMP 'fabric-installed-humanize-acceptance'
$driver = Join-Path $driverCheckout 'evals/v1/harness/accept-installed-humanize.ps1'
$arguments = @('-InstalledDirectory', $installDir, '-FabricSourceDirectory', $fabricSource,
  '-RepositoryPath', $repo,
  '-CodexExe', $codex, '-GoExe', $go, '-CandidateCopyExe', $helper,
  '-EvidenceOutputDirectory', $evidence)
& $pwsh -File $driver @arguments -PreflightOnly
if ($LASTEXITCODE -ne 0) { throw 'Installed-acceptance preflight failed' }
```

Only after reviewing the preflight inputs, run the same driver once without
`-PreflightOnly`. That step prepares and resumes a real model task and consumes
Codex usage. It requires a new pinned clone and a new evidence directory; on
failure, inspect and retain the evidence rather than retrying an uncertain run.

The script keeps inspect output and logs in the evidence directory; those local
artifacts may contain private workspace details. Keep that directory outside the
repository and do not publish raw run artifacts. The driver never builds the
helper implicitly, retries an uncertain run, or commits the candidate.
