param(
    [string]$GoExe = 'D:\dev\EngOrch-toolchains\go\1.27.1\go\bin\go.exe'
)

$ErrorActionPreference = 'Stop'
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..\..')).Path
$runner = Join-Path $repoRoot 'scripts\evaluate-v1.ps1'
$parseTokens = $null
$parseErrors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile($runner, [ref]$parseTokens, [ref]$parseErrors)
if ($parseErrors.Count -ne 0) { throw 'Evaluation runner does not parse.' }

foreach ($name in @('Get-FileSha256', 'Get-PinnedGoEnvironmentBinding', 'Invoke-WithPinnedGo', 'Write-CandidateCopySnapshotProjection')) {
    $functionAst = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq $name }, $true)
    if ($null -eq $functionAst) { throw "Runner helper missing: $name" }
    . ([scriptblock]::Create($functionAst.Extent.Text))
}

$resolvedGo = (Resolve-Path -LiteralPath $GoExe -ErrorAction Stop).Path
$binding = Get-PinnedGoEnvironmentBinding $resolvedGo
$expectedHash = (Get-FileSha256 $resolvedGo).ToLowerInvariant()
if ($binding.selected_go_sha256 -ne $expectedHash) { throw 'Pinned Go provenance hash differs from selected executable.' }
if (-not (Test-Path -LiteralPath $resolvedGo -PathType Leaf)) { throw 'Selected Go executable is missing.' }

$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("fabric-pinned-go-test-" + [Guid]::NewGuid().ToString('N'))
$emptyPath = Join-Path $tempRoot 'empty-path'
New-Item -ItemType Directory -Path $emptyPath -Force | Out-Null
$callerPath = $env:PATH
try {
    $env:PATH = $emptyPath
    if (Get-Command go -CommandType Application -ErrorAction SilentlyContinue) {
        throw 'Test PATH unexpectedly resolves go before the wrapper runs.'
    }

    $expectedArgs = @('env', 'GOVERSION')
    $probe = Invoke-WithPinnedGo $resolvedGo {
        $resolvedChild = (Get-Command go -CommandType Application -ErrorAction Stop).Source
        $output = & go @expectedArgs
        [pscustomobject]@{
            Resolved = $resolvedChild
            Output = ($output -join "`n").Trim()
            ExitCode = $LASTEXITCODE
            Args = @($expectedArgs)
        }
    }
    if ($probe.ExitCode -ne 0 -or $probe.Output -ne 'go1.27.1') { throw 'Child could not invoke the selected Go executable.' }
    if (-not [string]::Equals((Resolve-Path -LiteralPath $probe.Resolved).Path, $resolvedGo, [StringComparison]::OrdinalIgnoreCase)) {
        throw 'Child command resolution differs from selected Go executable.'
    }
    if (($probe.Args | ConvertTo-Json -Compress) -ne ($expectedArgs | ConvertTo-Json -Compress)) { throw 'Wrapped invocation changed child argv.' }
    if ($env:PATH -cne $emptyPath) { throw 'Caller PATH was not restored after a successful child.' }

    $nonzeroArgs = @('__fabric_invalid_subcommand__')
    $nonzero = Invoke-WithPinnedGo $resolvedGo {
        & go @nonzeroArgs 1> $null 2> $null
        [pscustomobject]@{ ExitCode = $LASTEXITCODE; Args = @($nonzeroArgs) }
    }
    if ($nonzero.ExitCode -eq 0) { throw 'Nonzero-exit fixture unexpectedly succeeded.' }
    if (($nonzero.Args | ConvertTo-Json -Compress) -ne ($nonzeroArgs | ConvertTo-Json -Compress)) { throw 'Wrapped failing invocation changed child argv.' }
    if ($env:PATH -cne $emptyPath) { throw 'Caller PATH was not restored after a nonzero child.' }

    $actionFailed = $false
    try { Invoke-WithPinnedGo $resolvedGo { throw 'expected fixture failure' } | Out-Null }
    catch { $actionFailed = $true }
    if (-not $actionFailed) { throw 'Expected wrapper action failure was not propagated.' }
    if ($env:PATH -cne $emptyPath) { throw 'Caller PATH was not restored after a thrown action.' }

    $invalidGoPath = Join-Path $tempRoot 'not-go.exe'
    [System.IO.File]::WriteAllBytes($invalidGoPath, [byte[]]@())
    $resolutionFailed = $false
    try { Invoke-WithPinnedGo $invalidGoPath { 'unreachable' } | Out-Null }
    catch { $resolutionFailed = $true }
    if (-not $resolutionFailed) { throw 'Wrapper admitted a directory that cannot resolve the selected Go executable.' }
    if ($env:PATH -cne $emptyPath) { throw 'Caller PATH was not restored after command-resolution failure.' }
} finally {
    if ($null -eq $callerPath) { Remove-Item -Path 'env:PATH' -ErrorAction SilentlyContinue }
    else { $env:PATH = $callerPath }
}

$snapshotRoot = Join-Path $tempRoot 'snapshot-projection'
New-Item -ItemType Directory -Path $snapshotRoot -Force | Out-Null
$fullPath = Join-Path $snapshotRoot 'full-inspect.json'
$projectionPath = Join-Path $snapshotRoot 'candidatecopy-snapshot.json'
$largeHosts = 'x' * (1100 * 1024)
$fixture = [ordered]@{
    run_id = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'
    state = 'READY'
    workspace = [ordered]@{ request = [ordered]@{ run_id = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'; path = 'D:\repo\.harness\worktrees\run'; candidate_identity = 'semantic-index-v2' }; binding = $null }
    candidate = [ordered]@{ version = 2; worktree_id = ('a' * 64); head = ('b' * 40); index_hash = ('c' * 64); files_hash = ('d' * 64); file_count = 2 }
    verification = [ordered]@{ pending = $false; plan_id = 'plan-1'; plan = [ordered]@{ candidate_id = ('e' * 64); invocations = @(@{ check_id = 'check-1' }); unused = $null }; observations = @(@{ result = @{ candidate_id = ('e' * 64); status = 'PASS' } }) }
    review = $null
    hosts = $largeHosts
}
$fullJson = ConvertTo-Json -InputObject $fixture -Depth 100 -Compress
$encoding = New-Object System.Text.UTF8Encoding($false)
[System.IO.File]::WriteAllText($fullPath, $fullJson, $encoding)
$snapshot = Get-Content -Raw -LiteralPath $fullPath | ConvertFrom-Json
$projectionEvidence = Write-CandidateCopySnapshotProjection $snapshot $fullPath $projectionPath
$projection = Get-Content -Raw -LiteralPath $projectionPath | ConvertFrom-Json
$requiredFields = @('run_id', 'state', 'workspace', 'candidate', 'verification', 'review')
$projectionFields = @($projection.PSObject.Properties | ForEach-Object { $_.Name })
if ($projectionFields.Count -ne $requiredFields.Count) { throw 'Candidate-copy projection includes missing or unrelated top-level fields.' }
foreach ($name in $requiredFields) {
    $expectedJson = ConvertTo-Json -InputObject $snapshot.$name -Depth 100 -Compress
    $actualJson = ConvertTo-Json -InputObject $projection.$name -Depth 100 -Compress
    if ($expectedJson -cne $actualJson) { throw "Candidate-copy projection changed required field $name." }
}
if (-not ($projection.PSObject.Properties.Name -contains 'review') -or $null -ne $projection.review) { throw 'Null review was not preserved in candidate-copy projection.' }
if ($projection.PSObject.Properties.Name -contains 'hosts') { throw 'Unrelated host data leaked into candidate-copy input.' }
if ((Get-Item -LiteralPath $fullPath).Length -le 1MB -or $projectionEvidence.ProjectionBytes -ge 1MB) {
    throw 'Large full inspect input was not reduced below the strict helper decoder bound.'
}
if ($projectionEvidence.FullSnapshotSha256 -ne (Get-FileSha256 $fullPath).ToLowerInvariant() -or
    $projectionEvidence.ProjectionSha256 -ne (Get-FileSha256 $projectionPath).ToLowerInvariant()) {
    throw 'Snapshot projection provenance hashes do not match written files.'
}

Remove-Item -LiteralPath $tempRoot -Recurse -Force
Write-Output 'PASS: pinned Go is child-resolvable with PATH cleared; argv is preserved; PATH restores after success, nonzero exit, and thrown failure; candidate-copy projection preserves six required fields and nulls while excluding large host data.'
