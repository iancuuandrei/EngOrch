# Fabric v1.0.0 Windows distribution acceptance

The immutable distribution was built from public source
`a4ca05692465bedc6b98d6ffbb8219b6d076f5d1`, tree
`ce581f6f38fa28f3e42cd6bdd3f37acf7ccc64e4`. Technical distribution and installed
task checks passed. The fresh matched Humanize comparison also passed:
484.962 seconds serial versus 304.622 seconds parallel. The Windows archive
is [published](https://github.com/iancuuandrei/Fabric/releases/tag/v1.0.0).
Independent review approved this evidence. Distribution acceptance is PASS
for Windows amd64; final documentation integration is recorded in
[PR #12](https://github.com/iancuuandrei/Fabric/pull/12).

## Distribution

| Check | Observed result |
|---|---|
| Declared supported platform | Windows amd64 |
| Independent clean builds | Two; identical archive bytes |
| Archive | `fabric_1.0.0_windows_amd64.zip`, 36,499,100 bytes |
| SHA-256 | `7add99a1146bfe8a780dd38238904cf74f9fdb7dffd1a36e5211d045988a21e6` |
| Build epoch | `1791012510` |
| Toolchains | Go 1.27.1; Cargo and Rust 1.94.1 |
| Package integrity and fresh installation | PASS |

Reproduce from the source commit above using the procedure in the
[release guide](../guides/release.md), version `v1.0.0`, the recorded epoch and
`-Platforms windows-amd64`. Linux is an optional build target and is not
qualified by this acceptance. Use a fresh detached checkout so later release
documentation commits do not change the build identity:

```powershell
git clone https://github.com/iancuuandrei/Fabric.git Fabric-v1-reproduce
Set-Location Fabric-v1-reproduce
git checkout --detach a4ca05692465bedc6b98d6ffbb8219b6d076f5d1
pwsh -File scripts/release-build.ps1 -Version v1.0.0 `
  -SourceDateEpoch 1791012510 -Platforms windows-amd64 `
  -OutputDirectory C:\tmp\fabric-v1-reproduce-a
pwsh -File scripts/release-build.ps1 -Version v1.0.0 `
  -SourceDateEpoch 1791012510 -Platforms windows-amd64 `
  -OutputDirectory C:\tmp\fabric-v1-reproduce-b
pwsh -File scripts/release-verify.ps1 -ReleaseDirectory C:\tmp\fabric-v1-reproduce-a
pwsh -File scripts/release-verify.ps1 -ReleaseDirectory C:\tmp\fabric-v1-reproduce-b
Get-FileHash C:\tmp\fabric-v1-reproduce-a\fabric_1.0.0_windows_amd64.zip
Get-FileHash C:\tmp\fabric-v1-reproduce-b\fabric_1.0.0_windows_amd64.zip
```

Both output directories must be new. Select the recorded toolchains explicitly
if multiple Go/Rust versions are installed. Packages are unsigned. Generated manifests
retain `release_qualified: false`; a build manifest does not encode a subsequent
independent acceptance or publication decision.

## Actual installed task

The installed executable ran a real ParseBytes feature task in `dustin/go-humanize`
at pinned commit `a1b4e66b9a6d890e9e15e7091cf16c8032367d6e`, with this exact objective:

> Use the configured repository intelligence (ri_search) to locate the byte
> parser before implementing the following objective. Add correctly validated
> underscore separators to ParseBytes integer inputs without weakening overflow
> detection.

It used stock Codex,
`gpt-6-luna`, high effort and writer edit validation. It reached READY and received
review approval with zero findings. Independent native and held-out checks
passed on separately captured copies of the accepted candidate. The original
prepared run was resumed after publishing its immutable lexical base; no
provider effects were repeated by the acceptance checks. The executed sequence
was `init --validate-writer-edits`, `run --autonomous --prepare-only`,
`ri prepare-lexical`, `ri lexical` with its unchanged preview and exact intent,
`ri lexical-ref`, a bounded operator `ri search`, then
`resume --autonomous` of that same run. The operator query was not counted as
agent RI use; the separate journal proof observes the explorer's actual calls.

| Binding | Identity |
|---|---|
| Run | `7465e3271f8f13d754edaa0fe41f00866963997f02984274bf1ff274a67d7a86` |
| Candidate | `086ce91b9f3ec6be952da71b6ce4f7f3a732a5fa6b3fce1d659d78ca01383a12` |
| Verification and review plan | `ddcdded95e4f278595a4a6c959aad5dbc614cced0692081ffef65e6e2b1c50fd` |
| Installed Fabric SHA-256 | `564c31d9531420fdf87aa9356d3078643678638437786130cfeb81a33501be63` |
| Installed RI SHA-256 | `65123be724fa8e0388747897cee5dc3cf889881e1c1535404e36eba3ed88ebe0` |

The explorer made four successful `ri_search` calls through the installed RI
executable. Three returned nonempty matches validated against the published
manifest. Controller and runtime journal chains, executable identity, completed
runtime result and controller receipt agree; read-only controller semantic
replay passed. This proves local integrity-checked linkage for the recorded
run. It does not prove external notarization, candidate overlays, semantic SCIP
queries or RI use by every role.

The [installed-task reproduction guide](../getting-started/installed-autonomous-task.md#reproduce-the-installed-acceptance)
uses the public installed-acceptance driver. After that driver reaches READY
and its native/held-out gates pass, reproduce the separate agent RI proof with
Python 3 and the included read-only collector. Use the driver checkout,
installed directory, task repository and evidence directory from that guide:

```powershell
$observed = Get-Content (Join-Path $evidence 'inspect-ready.stdout.json') -Raw | ConvertFrom-Json
$collector = Join-Path $driverCheckout 'evals\v1\harness\collect-installed-ri-agent-proof.ps1'
& $pwsh -NoProfile -File $collector `
  -InstalledRI (Join-Path $installDir 'engorch-ri.exe') `
  -FabricCLI (Join-Path $installDir 'fabric.exe') `
  -RepositoryRoot $repo -RunId $observed.run_id `
  -RuntimeStateRoot $observed.creation.config.codex.state_root `
  > (Join-Path $evidence 'agent-ri-proof.json')
if ($LASTEXITCODE -ne 0) { throw 'Agent RI use is not proven; inspect the proof result.' }
```

Do not pass the evidence directory as `-Acceptance`: that optional argument
accepts a repository/run-root directory or a receipt with repository metadata.
The command above supplies the actual repository and run explicitly. A new
model run may choose different searches; an operator query alone never supplies
this proof. The collector must report `PROVEN_JOURNAL_BOUND`, valid controller
chain and semantic replay, and manifest-validated nonempty agent results.

The in-repository collector was run read-only against the retained installed
run and reproduced that proof. It added no model calls. It uses canonical LF
line endings and adds strict run ID validation to the
independently reviewed collector before invoking the Fabric CLI.
Python source SHA-256:
`8ef4d09d1a1ab4d07c27f32218dad9ff33f0c7dcc611217aae550e788025163d`,
PowerShell SHA-256
`cbecf95d7ee7b08c299385dc6091aec7956fcb9a55de11cf429fa51ed2fec98d`.

## Integration evidence and limits

The complete Go suite passed on `58c6c115e9647f21e9caf9a8a16166c9e52bc5e6`.
Released source `a4ca056` differs only in documentation and three result records;
production, build and test source is identical. The suite was not rerun on
`a4ca056`. Exact-source main Sonar passed with zero issues. The
[integration receipt](../../evals/v1/results/pr11-integration-58c6c11-20261003.json)
records the actual test source and outcome.

The [machine-readable acceptance](../../evals/v1/results/release-v1.0.0-acceptance-20261003.json)
records technical gates and qualification status separately. Native and held-out
acceptance added no provider calls and did not edit the live candidate workspace.
An earlier preparation receipt marked held-out checks NOT RUN; the later
candidate-acceptance receipt supplies their actual PASS evidence.

This task and the wider real-repository evaluation demonstrate bounded useful
capabilities. They do not establish unrestricted success, a general success
rate, an OS security sandbox, Linux support or signed publisher authenticity.

Start with the [installed Windows guide](../getting-started/installed-autonomous-task.md).
