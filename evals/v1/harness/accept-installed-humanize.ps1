<#
.SYNOPSIS
Reproduce the installed Windows v1.0.0 Humanize task and candidate acceptance.
.DESCRIPTION
Requires the public driver checkout plus a clean Fabric source checkout, an exact pinned go-humanize clone,
the installed v1.0.0 binaries, stock Codex, Go 1.27.1, and an explicit
candidatecopy helper built from that clean Fabric checkout. It never builds a
helper or automatically retries an uncertain run. The model task may modify its
isolated candidate workspace; native and held-out acceptance checks run on copies.
Use -PreflightOnly to validate inputs without creating a run or invoking a model.
#>
[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)][string]$InstalledDirectory,
    [Parameter(Mandatory = $true)][string]$FabricSourceDirectory,
    [Parameter(Mandatory = $true)][string]$RepositoryPath,
    [Parameter(Mandatory = $true)][string]$CodexExe,
    [Parameter(Mandatory = $true)][string]$GoExe,
    [Parameter(Mandatory = $true)][string]$CandidateCopyExe,
    [Parameter(Mandatory = $true)][string]$EvidenceOutputDirectory,
    [switch]$PreflightOnly
)

$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $false
$DriverRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..\..\..'))
$FabricSourceRoot = [System.IO.Path]::GetFullPath((Resolve-Path -LiteralPath $FabricSourceDirectory -ErrorAction Stop).Path)
$ManifestPath = Join-Path $FabricSourceRoot 'evals\v1\manifest.json'
$HeldoutFixture = Join-Path $FabricSourceRoot 'evals\v1\heldout\humanize.heldout_test.go'
$ExpectedSourceCommit = 'a4ca05692465bedc6b98d6ffbb8219b6d076f5d1'
$ExpectedRepoCommit = 'a1b4e66b9a6d890e9e15e7091cf16c8032367d6e'
$ExpectedFabricSha256 = '564c31d9531420fdf87aa9356d3078643678638437786130cfeb81a33501be63'
$ExpectedRISha256 = '65123be724fa8e0388747897cee5dc3cf889881e1c1535404e36eba3ed88ebe0'
$ExpectedArchiveSha256 = '7add99a1146bfe8a780dd38238904cf74f9fdb7dffd1a36e5211d045988a21e6'

. (Join-Path $FabricSourceRoot 'evals\v1\harness\Go-ArgvPolicy.ps1')
. (Join-Path $FabricSourceRoot 'evals\v1\harness\Toml-ArgvPolicy.ps1')

function Get-FileSha256([string]$Path) {
    $stream = [System.IO.File]::OpenRead($Path)
    $sha = [System.Security.Cryptography.SHA256]::Create()
    try { return ([BitConverter]::ToString($sha.ComputeHash($stream)) -replace '-', '').ToLowerInvariant() }
    finally { $stream.Dispose(); $sha.Dispose() }
}

function Resolve-RequiredFile([string]$Path, [string]$Name) {
    $item = Get-Item -LiteralPath $Path -ErrorAction Stop
    if ($item.PSIsContainer) { throw "$Name must be a file." }
    return $item.FullName
}

function Invoke-GitTextAt([string]$Path, [string[]]$Arguments) {
    $value = & $script:GitExe --no-optional-locks -c diff.autoRefreshIndex=false -C $Path @Arguments
    if ($LASTEXITCODE -ne 0) { throw 'Pinned repository identity check failed.' }
    return ($value -join "`n").Trim()
}

function Invoke-GitText([string[]]$Arguments) {
    return Invoke-GitTextAt $script:RepositoryRoot $Arguments
}

function Test-PathContains([string]$Parent, [string]$Child) {
    $parentFull = [System.IO.Path]::GetFullPath($Parent).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
    $childFull = [System.IO.Path]::GetFullPath($Child).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
    return $childFull.Equals($parentFull, [StringComparison]::OrdinalIgnoreCase) -or
        $childFull.StartsWith($parentFull + [System.IO.Path]::DirectorySeparatorChar, [StringComparison]::OrdinalIgnoreCase)
}

function Invoke-WithPinnedGo([scriptblock]$Action) {
    $selected = $script:GoExecutable
    $selectedDir = [System.IO.Path]::GetDirectoryName($selected)
    $previousPath = $env:PATH
    try {
        $env:PATH = if ([string]::IsNullOrEmpty($previousPath)) { $selectedDir } else { $selectedDir + [System.IO.Path]::PathSeparator + $previousPath }
        $visibleGo = Get-Command go -CommandType Application -ErrorAction Stop | Select-Object -First 1
        $visiblePath = [System.IO.Path]::GetFullPath($visibleGo.Source)
        if (-not $visiblePath.Equals($selected, [StringComparison]::OrdinalIgnoreCase)) {
            throw 'Child PATH does not resolve go to the selected executable.'
        }
        & $Action
    } finally {
        if ($null -eq $previousPath) { Remove-Item -Path 'env:PATH' -ErrorAction SilentlyContinue }
        else { $env:PATH = $previousPath }
    }
}

function Invoke-FabricJson([string]$Label, [string[]]$Arguments, [switch]$NoRepositoryRoot) {
    $stdoutPath = Join-Path $script:EvidenceRoot ($Label + '.stdout.json')
    $stderrPath = Join-Path $script:EvidenceRoot ($Label + '.stderr.log')
    Invoke-WithPinnedGo {
        if ($NoRepositoryRoot) { & $script:FabricExe @Arguments 1> $stdoutPath 2> $stderrPath }
        else { & $script:FabricExe --root $script:RepositoryRoot @Arguments 1> $stdoutPath 2> $stderrPath }
        $script:LastFabricExitCode = $LASTEXITCODE
    }
    if ($script:LastFabricExitCode -ne 0) {
        throw "Fabric step '$Label' failed; inspect retained evidence and do not automatically retry."
    }
    $raw = Get-Content -LiteralPath $stdoutPath -Raw
    if ([string]::IsNullOrWhiteSpace($raw)) { throw "Fabric step '$Label' produced no JSON output." }
    try { return ($raw | ConvertFrom-Json -Depth 100) }
    catch { throw "Fabric step '$Label' output was not valid JSON; evidence is retained." }
}

function Invoke-GoTest([string]$WorkDirectory, [string[]]$FullArgv, [string]$Label) {
    $split = Split-GoInvocation -FullArgv $FullArgv
    $stdoutPath = Join-Path $script:EvidenceRoot ($Label + '.stdout.log')
    $stderrPath = Join-Path $script:EvidenceRoot ($Label + '.stderr.log')
    Push-Location -LiteralPath $WorkDirectory
    try {
        & $script:GoExecutable @($split.Args) 1> $stdoutPath 2> $stderrPath
        $exitCode = $LASTEXITCODE
    } finally { Pop-Location }
    $stdout = if (Test-Path -LiteralPath $stdoutPath) { Get-Content -LiteralPath $stdoutPath -Raw } else { '' }
    $stderr = if (Test-Path -LiteralPath $stderrPath) { Get-Content -LiteralPath $stderrPath -Raw } else { '' }
    return [pscustomobject]@{ ExitCode = $exitCode; Stdout = [string]$stdout; Stderr = [string]$stderr }
}

function Set-HumanizeVerificationConfig([object]$Entry) {
    $configPath = Join-Path $script:RepositoryRoot 'harness.toml'
    if (-not (Test-Path -LiteralPath $configPath -PathType Leaf)) { throw 'Fabric init did not create harness.toml.' }
    $text = Get-Content -LiteralPath $configPath -Raw
    $current = @(Read-VerificationArgv -Text $text)
    if (($current -join "`0") -cne (@('go', 'test', './...') -join "`0")) {
        throw 'harness.toml verification argv differs from the expected init default.'
    }
    $native = @($Entry.native_argv)
    $match = [regex]::Match($text, '(?m)^argv\s*=\s*\[[^\]]*\]')
    if (-not $match.Success -or $native.Count -lt 3 -or $native[0] -cne 'go' -or $native[1] -cne 'test') {
        throw 'Manifest verification argv or initialized TOML shape is invalid.'
    }
    $replacement = 'argv = [' + (($native | ForEach-Object { '"' + ($_ -replace '"', '\"') + '"' }) -join ', ') + ']'
    $text = $text.Substring(0, $match.Index) + $replacement + $text.Substring($match.Index + $match.Length)
    [System.IO.File]::WriteAllText($configPath, $text, (New-Object System.Text.UTF8Encoding($false)))
    return [ordered]@{ argv = $native; sha256 = (Get-FileSha256 $configPath) }
}

function Add-LocalExclusions {
    $gitPath = Invoke-GitText @('rev-parse', '--git-path', 'info/exclude')
    $excludePath = if ([System.IO.Path]::IsPathRooted($gitPath)) { $gitPath } else { Join-Path $script:RepositoryRoot $gitPath }
    if (-not (Test-Path -LiteralPath $excludePath -PathType Leaf)) { throw 'Repository-local Git exclude file is missing.' }
    $text = Get-Content -LiteralPath $excludePath -Raw
    foreach ($pattern in @('.harness/', 'harness.toml')) {
        if (-not [regex]::IsMatch($text, '(?m)^/?' + [regex]::Escape($pattern) + '$')) {
            if (-not $text.EndsWith("`n")) { $text += "`n" }
            $text += $pattern + "`n"
        }
    }
    [System.IO.File]::WriteAllText($excludePath, $text, (New-Object System.Text.UTF8Encoding($false)))
}

function Invoke-CandidateCopy([string]$SnapshotPath, [string]$ExpectedCandidate, [string]$Destination, [string]$Label) {
    $stdoutPath = Join-Path $script:EvidenceRoot ($Label + '.stdout.json')
    $stderrPath = Join-Path $script:EvidenceRoot ($Label + '.stderr.log')
    & $script:CandidateCopyExecutable --snapshot $SnapshotPath --expected-candidate $ExpectedCandidate --destination $Destination 1> $stdoutPath 2> $stderrPath
    if ($LASTEXITCODE -ne 0) { throw "Candidate copy '$Label' was blocked; preserve its destination and logs." }
    $raw = Get-Content -LiteralPath $stdoutPath -Raw
    if ([string]::IsNullOrWhiteSpace($raw)) { throw "Candidate copy '$Label' produced no observation." }
    try { $observation = $raw | ConvertFrom-Json } catch { throw "Candidate copy '$Label' output was not valid JSON." }
    if ($observation.candidate_id -cne $ExpectedCandidate) { throw "Candidate copy '$Label' returned a different candidate." }
    return $observation
}

# Resolve all explicit inputs before creating evidence or changing the task clone.
if ($PSVersionTable.PSVersion.Major -lt 7) { throw 'This installed acceptance driver requires PowerShell 7 or newer.' }
if ($env:OS -ne 'Windows_NT') { throw 'This installed acceptance driver is qualified for Windows only.' }
if (-not (Test-Path -LiteralPath $ManifestPath -PathType Leaf) -or -not (Test-Path -LiteralPath $HeldoutFixture -PathType Leaf)) {
    throw 'Public manifest or Humanize held-out fixture is missing.'
}
$InstallRoot = (Resolve-Path -LiteralPath $InstalledDirectory -ErrorAction Stop).Path
$script:FabricExe = Resolve-RequiredFile (Join-Path $InstallRoot 'fabric.exe') 'Installed Fabric executable'
$riExecutable = Resolve-RequiredFile (Join-Path $InstallRoot 'engorch-ri.exe') 'Installed RI executable'
$script:RepositoryRoot = (Resolve-Path -LiteralPath $RepositoryPath -ErrorAction Stop).Path
$script:CodexExecutable = Resolve-RequiredFile $CodexExe 'Codex executable'
$script:GoExecutable = Resolve-RequiredFile $GoExe 'Go executable'
$script:CandidateCopyExecutable = Resolve-RequiredFile $CandidateCopyExe 'Candidate-copy helper'
$script:GitExe = (Get-Command git -CommandType Application -ErrorAction Stop | Select-Object -First 1).Source
$outputParent = [System.IO.Path]::GetDirectoryName([System.IO.Path]::GetFullPath($EvidenceOutputDirectory))
if ([string]::IsNullOrWhiteSpace($outputParent) -or -not (Test-Path -LiteralPath $outputParent -PathType Container)) {
    throw 'Evidence output parent must already exist.'
}
$outputFull = [System.IO.Path]::GetFullPath($EvidenceOutputDirectory)
if (Test-Path -LiteralPath $outputFull) { throw 'Evidence output directory must be new; retain and inspect prior evidence.' }
foreach ($root in @($script:RepositoryRoot, $FabricSourceRoot, $DriverRoot)) {
    if ((Test-PathContains $root $outputFull) -or (Test-PathContains $outputFull $root)) {
        throw 'Evidence output must be separate from the task repository and Fabric source checkout.'
    }
}

$manifest = Get-Content -LiteralPath $ManifestPath -Raw | ConvertFrom-Json
$entries = @($manifest.repositories | Where-Object { $_.id -eq 'go-humanize' })
if ($entries.Count -ne 1 -or $entries[0].sha -cne $ExpectedRepoCommit -or $entries[0].check -cne 'humanize') {
    throw 'Pinned Humanize manifest entry does not match the acceptance contract.'
}
$entry = $entries[0]
$repoRoot = Invoke-GitText @('rev-parse', '--show-toplevel')
if (-not [System.IO.Path]::GetFullPath($repoRoot).Equals([System.IO.Path]::GetFullPath($script:RepositoryRoot), [StringComparison]::OrdinalIgnoreCase)) {
    throw 'RepositoryPath must name the pinned clone root.'
}
$repoHead = Invoke-GitText @('rev-parse', 'HEAD')
if ($repoHead -cne $ExpectedRepoCommit -or -not [string]::IsNullOrWhiteSpace((Invoke-GitText @('status', '--porcelain=v1', '--untracked-files=all')))) {
    throw 'Task repository must be clean at the exact manifest commit.'
}
if ((Test-Path -LiteralPath (Join-Path $script:RepositoryRoot 'harness.toml')) -or
    (Test-Path -LiteralPath (Join-Path $script:RepositoryRoot '.harness'))) {
    throw 'Task clone is already initialized; use a fresh pinned clone and preserve existing runs.'
}

$script:FabricSha256 = Get-FileSha256 $script:FabricExe
$script:RISha256 = Get-FileSha256 $riExecutable
if ($script:FabricSha256 -cne $ExpectedFabricSha256 -or $script:RISha256 -cne $ExpectedRISha256) {
    throw 'Installed binary hashes do not match the accepted v1.0.0 Windows bundle.'
}
$versionText = (& $script:FabricExe version | Out-String).Trim()
if ($LASTEXITCODE -ne 0) { throw 'Installed Fabric version command failed.' }
try { $version = $versionText | ConvertFrom-Json } catch { throw 'Installed Fabric version output is invalid JSON.' }
if ($version.version -cne 'v1.0.0' -or $version.commit -cne $ExpectedSourceCommit) { throw 'Installed Fabric version identity mismatch.' }
$goVersion = (& $script:GoExecutable version | Out-String).Trim()
if ($LASTEXITCODE -ne 0 -or $goVersion -cne 'go version go1.27.1 windows/amd64') { throw 'Acceptance requires the recorded Go 1.27.1 Windows amd64 toolchain.' }
$script:GoSha256 = Get-FileSha256 $script:GoExecutable
$script:CodexSha256 = Get-FileSha256 $script:CodexExecutable
$script:CandidateCopySha256 = Get-FileSha256 $script:CandidateCopyExecutable

$sourceGitRoot = Invoke-GitTextAt $FabricSourceRoot @('rev-parse', '--show-toplevel')
if (-not [System.IO.Path]::GetFullPath($sourceGitRoot).Equals([System.IO.Path]::GetFullPath($FabricSourceRoot), [StringComparison]::OrdinalIgnoreCase)) {
    throw 'FabricSourceDirectory must name the clean Fabric source checkout root.'
}
$sourceHead = Invoke-GitTextAt $FabricSourceRoot @('rev-parse', 'HEAD')
if ($sourceHead -cne $ExpectedSourceCommit -or -not [string]::IsNullOrWhiteSpace((Invoke-GitTextAt $FabricSourceRoot @('status', '--porcelain=v1', '--untracked-files=all')))) {
    throw 'FabricSourceDirectory must be a clean checkout at the accepted v1.0.0 source commit.'
}
$helperBuildInfo = (& $script:GoExecutable version -m $script:CandidateCopyExecutable 2>&1 | Out-String)
if ($LASTEXITCODE -ne 0) { throw 'Could not inspect candidate-copy helper build identity.' }
$helperRevision = [regex]::Match($helperBuildInfo, '(?m)^\s*build\s+vcs.revision=([0-9a-f]{40})\r?$')
$helperModified = [regex]::Match($helperBuildInfo, '(?m)^\s*build\s+vcs.modified=(true|false)\r?$')
if (-not $helperRevision.Success -or $helperRevision.Groups[1].Value -cne $sourceHead -or
    -not $helperModified.Success -or $helperModified.Groups[1].Value -cne 'false') {
    throw 'Candidate-copy helper must be built from the clean accepted v1.0.0 Fabric source checkout.'
}

if ($PreflightOnly) {
    Write-Output 'PASS input preflight; no repository files changed, no run created, no model invoked.'
    return
}

$script:EvidenceRoot = $outputFull
New-Item -ItemType Directory -Path $script:EvidenceRoot | Out-Null
$nativeFullArgv = @($entry.native_argv)
if ($nativeFullArgv.Count -lt 3 -or $nativeFullArgv[0] -cne 'go' -or $nativeFullArgv[1] -cne 'test') {
    throw 'Manifest native verification argv is not a full Go test command.'
}
$nativeFixturePath = 'fabric_v1_heldout_test.go'
$objective = 'Use the configured repository intelligence (ri_search) to locate the byte parser before implementing the following objective. ' + [string]$entry.task

Add-LocalExclusions
if (-not [string]::IsNullOrWhiteSpace((Invoke-GitText @('status', '--porcelain=v1', '--untracked-files=all')))) {
    throw 'Pinned repository became dirty before initialization.'
}

$versionReceipt = [ordered]@{
    fabric_source_commit = $sourceHead
    acceptance_driver_sha256 = Get-FileSha256 $PSCommandPath
    task_repository = [ordered]@{ id = $entry.id; commit = $repoHead; task = [string]$entry.task }
    installed_fabric_sha256 = $script:FabricSha256
    installed_ri_sha256 = $script:RISha256
    codex_sha256 = $script:CodexSha256
    go = $goVersion
    go_sha256 = $script:GoSha256
    candidatecopy_sha256 = $script:CandidateCopySha256
    release_archive_sha256 = $ExpectedArchiveSha256
    created_utc = [DateTime]::UtcNow.ToString('o')
    additional_provider_calls_by_acceptance = 0
    raw_transcripts_published = $false
}
$versionReceipt | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath (Join-Path $script:EvidenceRoot 'inputs.json') -Encoding utf8

$init = Invoke-FabricJson 'init' @('init', '--codex', $script:CodexExecutable, '--model', 'gpt-6-luna', '--effort', 'high', '--validate-writer-edits')
if ($init.status -cne 'CREATED' -or $init.configuration -cne 'harness.toml' -or $init.runtime -cne 'codex-app-server') {
    throw 'Installed Fabric initialization contract mismatch.'
}
$verification = Set-HumanizeVerificationConfig $entry
Invoke-FabricJson 'doctor' @('doctor') | Out-Null

$prepared = Invoke-FabricJson 'prepared' @('run', '--autonomous', '--prepare-only', $objective)
if ($prepared.status -cne 'PREPARED' -or $prepared.state -cne 'IMPLEMENTING' -or $prepared.run_id -notmatch '^[0-9a-f]{64}$') {
    throw 'Exact prepared autonomous run was not created; retain its log and do not retry.'
}
$runId = [string]$prepared.run_id

$baseStage = Join-Path $script:EvidenceRoot 'lexical-base'
$basePreview = Invoke-FabricJson 'base-preview' @('ri', 'prepare-lexical', $runId, $riExecutable, $script:RISha256, $baseStage, '67108864', '10000')
if ([string]::IsNullOrWhiteSpace($basePreview.intent_id)) { throw 'Lexical base preview omitted its intent.' }
$base = Invoke-FabricJson 'base-published' @('ri', 'lexical', $runId, (Join-Path $script:EvidenceRoot 'base-preview.stdout.json'), [string]$basePreview.intent_id, 'automation:fabric-v1-acceptance')
if ($base.ri_lexical.outcome -cne 'CONFIRMED') { throw 'Immutable lexical base publication was not confirmed; do not resend.' }
$reference = Invoke-FabricJson 'base-reference' @('ri', 'lexical-ref', $runId)
if ($reference.manifest.id -cne $base.ri_lexical.observation.build.manifest_id) { throw 'Published lexical base reference mismatch.' }
$search = Invoke-FabricJson 'base-search' @('ri', 'search', $riExecutable, $script:RISha256, (Join-Path $script:EvidenceRoot 'base-reference.stdout.json'), '--fixed', '--limit', '8', 'ParseBytes')
if ($search.manifest_id -cne $reference.manifest.id -or @($search.matches).Count -lt 1 -or @($search.matches).Count -gt 8) {
    throw 'Installed RI did not return a bounded ParseBytes result from the published base.'
}

Invoke-FabricJson 'resumed' @('resume', '--autonomous', $runId) | Out-Null
$snapshotPath = Join-Path $script:EvidenceRoot 'inspect-ready.stdout.json'
$snapshot = Invoke-FabricJson 'inspect-ready' @('inspect', $runId)
$usage = Invoke-FabricJson 'usage' @('usage', $runId)
$diff = Invoke-FabricJson 'diff' @('diff', $runId)
if ($snapshot.run_id -cne $runId -or $snapshot.state -cne 'READY') { throw 'Run snapshot is not READY for the exact prepared run.' }
if ($null -eq $snapshot.verification -or $snapshot.verification.pending -eq $true) { throw 'Snapshot verification is missing or pending.' }
$invocations = @($snapshot.verification.plan.invocations)
$observations = @($snapshot.verification.observations)
if ($snapshot.verification.plan.candidate_id -notmatch '^[0-9a-f]{64}$' -or
    $invocations.Count -eq 0 -or $observations.Count -ne $invocations.Count) { throw 'Verification plan and observations are incomplete.' }
$candidateId = [string]$snapshot.verification.plan.candidate_id
foreach ($observation in $observations) {
    if ($observation.result.candidate_id -cne $candidateId -or $observation.result.status -cne 'PASS') {
        throw 'Verification observation is not a PASS bound to the candidate.'
    }
}
if ($null -eq $snapshot.review -or [string]::IsNullOrWhiteSpace($snapshot.review.result.output)) { throw 'Reviewer result is missing.' }
try { $verdict = $snapshot.review.result.output | ConvertFrom-Json } catch { throw 'Reviewer result is not valid JSON.' }
if ($verdict.candidate_id -cne $candidateId -or $verdict.verification_plan_id -cne $snapshot.verification.plan_id -or
    $verdict.decision -cne 'approve' -or @($verdict.findings).Count -ne 0 -or $diff.candidate_id -cne $candidateId) {
    throw 'Review, verification and diff do not agree on one approved candidate.'
}
$workspace = [string]$snapshot.workspace.request.path
if ([string]::IsNullOrWhiteSpace($workspace) -or $snapshot.workspace.request.run_id -cne $runId) {
    throw 'Candidate workspace is not bound to the exact run.'
}

# candidatecopy intentionally accepts only this complete six-field projection.
$projection = [ordered]@{}
foreach ($name in @('run_id', 'state', 'workspace', 'candidate', 'verification', 'review')) {
    $property = $snapshot.PSObject.Properties[$name]
    if ($null -eq $property) { throw "Inspect snapshot is missing required field '$name'." }
    $projection[$name] = $property.Value
}
$projectionPath = Join-Path $script:EvidenceRoot 'candidatecopy-snapshot.json'
$projectionJson = ConvertTo-Json -InputObject $projection -Depth 100 -Compress
[System.IO.File]::WriteAllText($projectionPath, $projectionJson, (New-Object System.Text.UTF8Encoding($false)))

$nativeCopyPath = Join-Path $script:EvidenceRoot 'native-candidate-copy'
$nativeCopy = Invoke-CandidateCopy $projectionPath $candidateId $nativeCopyPath 'native-copy'
$nativeManifest = $nativeCopy | ConvertTo-Json -Depth 8
Set-Content -LiteralPath (Join-Path $script:EvidenceRoot 'native-copy-manifest.json') -Value $nativeManifest -Encoding utf8
$nativeResult = Invoke-GoTest ([string]$nativeCopy.destination) $nativeFullArgv 'native-verification'

# Capture the held-out candidate separately, after candidate-controlled native
# tests finish. Both copies must match the same reviewed source identity.
$heldoutCopyPath = Join-Path $script:EvidenceRoot 'heldout-candidate-copy'
$heldoutCopy = Invoke-CandidateCopy $projectionPath $candidateId $heldoutCopyPath 'heldout-copy'
$heldoutManifest = $heldoutCopy | ConvertTo-Json -Depth 8
Set-Content -LiteralPath (Join-Path $script:EvidenceRoot 'heldout-copy-manifest.json') -Value $heldoutManifest -Encoding utf8
if ($heldoutCopy.candidate_id -cne $candidateId -or $heldoutCopy.files_hash -cne $nativeCopy.files_hash -or
    $heldoutCopy.workspace -cne $nativeCopy.workspace -or [int]$heldoutCopy.file_count -ne [int]$nativeCopy.file_count) {
    throw 'Independent acceptance copies do not bind identical candidate files and workspace.'
}
$fixtureHash = Get-FileSha256 $HeldoutFixture
$candidateFixturePath = Join-Path ([string]$heldoutCopy.destination) $nativeFixturePath
if (Test-Path -LiteralPath $candidateFixturePath) { throw 'Held-out test destination already exists in the fresh copy.' }
Copy-Item -LiteralPath $HeldoutFixture -Destination $candidateFixturePath
try {
    $heldoutFullArgv = @(Add-FabricV1HeldoutArguments -FullArgv $nativeFullArgv)
    if ($heldoutFullArgv.Count -lt 4) { throw 'Held-out Go argv construction failed.' }
    $heldoutFullArgv = @($heldoutFullArgv[0..2] + @('-v') + $heldoutFullArgv[3..($heldoutFullArgv.Count - 1)])
    $heldoutResult = Invoke-GoTest ([string]$heldoutCopy.destination) $heldoutFullArgv 'heldout-acceptance'
} finally {
    if (Test-Path -LiteralPath $candidateFixturePath -PathType Leaf) { Remove-Item -LiteralPath $candidateFixturePath -Force }
}
$heldoutOutput = $heldoutResult.Stdout + "`n" + $heldoutResult.Stderr
$heldoutRan = $heldoutOutput -match '(?m)^=== RUN\s+TestFabricV1Heldout\s*$'
$heldoutPassed = $heldoutOutput -match '(?m)^--- PASS:\s+TestFabricV1Heldout(?:\s|$)'
$acceptance = if ($nativeResult.ExitCode -eq 0 -and $heldoutResult.ExitCode -eq 0 -and $heldoutRan -and $heldoutPassed) { 'PASS' } else { 'FAIL' }
$receipt = [ordered]@{
    schema = 'fabric.v1.installed_candidate_acceptance.v1'
    acceptance = $acceptance
    release_qualified = $false
    source_commit = $ExpectedSourceCommit
    fabric_source_commit = $sourceHead
    acceptance_driver_sha256 = Get-FileSha256 $PSCommandPath
    repository_id = [string]$entry.id
    repository_commit = $repoHead
    task = [string]$entry.task
    run_id = $runId
    terminal_state = [string]$snapshot.state
    candidate_id = $candidateId
    verification_plan_id = [string]$snapshot.verification.plan_id
    review_decision = [string]$verdict.decision
    review_findings = @($verdict.findings).Count
    installed_fabric_sha256 = $script:FabricSha256
    installed_ri_sha256 = $script:RISha256
    codex_sha256 = $script:CodexSha256
    go_version = $goVersion
    go_sha256 = $script:GoSha256
    candidatecopy_sha256 = $script:CandidateCopySha256
    manifest_sha256 = Get-FileSha256 $ManifestPath
    heldout_fixture_sha256 = $fixtureHash
    native_argv = $nativeFullArgv
    native_exit_code = $nativeResult.ExitCode
    heldout_argv = $heldoutFullArgv
    heldout_exit_code = $heldoutResult.ExitCode
    heldout_test_observed = [bool]($heldoutRan -and $heldoutPassed)
    native_copy_candidate_id = [string]$nativeCopy.candidate_id
    heldout_copy_candidate_id = [string]$heldoutCopy.candidate_id
    native_copy_files_hash = [string]$nativeCopy.files_hash
    heldout_copy_files_hash = [string]$heldoutCopy.files_hash
    gate_copies = 'independent; held-out copy captured after native tests finish'
    acceptance_added_provider_calls = 0
    acceptance_gate_candidate_workspace_edits = $false
    archive_sha256 = $ExpectedArchiveSha256
    published_lexical_base_id = [string]$reference.manifest.id
    installed_base_search_match_count = @($search.matches).Count
    usage_observation_retained = $true
    inspect_sha256 = Get-FileSha256 $snapshotPath
    projection_sha256 = Get-FileSha256 $projectionPath
    prepared_config_sha256 = [string]$verification.sha256
    created_utc = [DateTime]::UtcNow.ToString('o')
}
$receipt | ConvertTo-Json -Depth 12 | Set-Content -LiteralPath (Join-Path $script:EvidenceRoot 'acceptance-receipt.json') -Encoding utf8
if ($acceptance -ne 'PASS') { throw 'Installed candidate acceptance failed; retain evidence and do not automatically rerun.' }
Write-Output "PASS installed acceptance; run_id=$runId candidate_id=$candidateId"
