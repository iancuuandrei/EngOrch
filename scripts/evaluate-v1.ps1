<#
.SYNOPSIS
Real Fabric v1 evaluation runner: Prepare fresh pinned tasks, then Evaluate
them with an explicit provider-gated action. No meta-controller.
.DESCRIPTION
Prepare creates fresh pinned detached clones under an external run root,
excludes Fabric config (.harness/harness.toml) from the candidate diff, runs
a native `go test` preflight (manifest native_argv policy; Windows-only
`-skip ^TestNocmpIntegration$` for go-atomic; explicit file argv for legacy
go-difflib) plus a disposable-copy held-out discriminator with exact
stdout/stderr capture and targeted_assertion / expected_missing_api /
setup_error / pass classification, and writes a sanitized run.json with
provider_calls=0. It never invokes the provider.

Manifest native_argv keeps the full Fabric verification argv including the
`go` executable; local runs strip and validate it at invocation
(Go-ArgvPolicy.ps1) so GoExe never receives a doubled executable.
Evaluate runs only under -Action Evaluate with explicit FabricExe, GoExe,
CodexExe, Model and Effort (default high, user value persisted) plus an
explicit -CandidateCopyExe observation/copy helper (fingerprinted, never
auto-built or relaunched). Before any run is created, the task
harness.toml required checks are set to the explicit manifest native_argv
policy (init defaults to `go test ./...`) and hashed. Native mode runs
`fabric init --codex EXE --model MODEL --effort EFFORT`, then native
`fabric run --autonomous` objective plus inspect/usage/diff evidence
gathering. PR5Matched mode uses the explicit -PR5BaselineScript with the
supplied baseline exe (init with baseline exe/model first, inspect+usage
with the baseline exe, exact run-identity parsing, diff collected from the
candidate path because the baseline lacks diff). Task PASS requires a
successful run+inspect, READY snapshot, exact current candidate bound by the
helper to the reviewed candidate_id and verification plan candidate_id,
non-pending all-PASS verification, reviewer approve with zero findings and
matching candidate_id+verification_plan_id, plus native upstream
verification AND held-out PASS. Missing evidence is BLOCKED, never rerun.
Hidden acceptance tests run only in the helper's candidate-bound byte-copy
of the finished candidate; the original worktree is never mutated.
UNKNOWN outcomes are BLOCKED, never automatically rerun. receipt_matched
rows are completed runtime invocations, not provider calls: provider_calls
stays null unless an actual count exists, while runtime invocations and
ContextUsage provider_usage tokens are recorded separately. Unknown metrics
stay null, never zero. Evaluation provenance records evaluation-time
product_head/dirty, exact binary bytes plus embedded Go vcs metadata,
all runner/helper source hashes, linked prepare run.json bytes, task pins,
and helper binary provenance separately. No raw transcripts or credentials
are published.
#>
[CmdletBinding()]
param(
    [ValidateSet('Prepare', 'Evaluate')][string]$Action = 'Prepare',
    [string]$FabricExe,
    [string]$GoExe = 'D:\dev\EngOrch-toolchains\go\1.27.1\go\bin\go.exe',
    [string]$CodexExe,
    [string]$Model,
    [ValidateSet('Native', 'PR5Matched')][string]$EvalMode = 'Native',
    [string]$Effort = 'high',
    [string]$PR5BaselineExe,
    [string]$PR5BaselineScript,
    [string]$CandidateCopyExe,
    [string]$NativeBuildReceiptPath,
    [string]$RunRoot = 'D:\dev\Fabric-v1-eval-runs',
    [string[]]$TaskIds,
    [string]$RunId
)

$ErrorActionPreference = 'Stop'
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$manifestPath = Join-Path $repoRoot 'evals\v1\manifest.json'
$heldoutDir = Join-Path $repoRoot 'evals\v1\heldout'
$defaultRunTaskScript = Join-Path $PSScriptRoot 'run-task.ps1'
$manifest = Get-Content -Raw $manifestPath | ConvertFrom-Json
$git = (Get-Command git -ErrorAction Stop).Source
. (Join-Path $repoRoot 'evals\v1\harness\Classify-CheckOutput.ps1')
# Copy-CandidateTree.ps1 is NOT used for acceptance here: the candidate-bound
# Go helper (-CandidateCopyExe) owns observation/copy under a read lease and
# binds Candidate.ID() to the reviewed candidate and verification plan.
# The PS copy remains only for its standalone fixtures, never as an
# acceptance route (self-only copy would verify the copy against itself).
. (Join-Path $repoRoot 'evals\v1\harness\Go-ArgvPolicy.ps1')
. (Join-Path $repoRoot 'evals\v1\harness\Toml-ArgvPolicy.ps1')

$DiffByteLimit = 1048576

function Get-GitText([string]$Path, [string[]]$GitArgs) {
    $value = & $git -C $Path @GitArgs
    if ($LASTEXITCODE -ne 0) { throw "git failed in ${Path}: $($GitArgs -join ' ')" }
    return ($value -join "`n").Trim()
}

function Get-FileSha256([string]$Path) {
    $h = [System.Security.Cryptography.SHA256]::Create()
    try {
        $fs = [System.IO.File]::OpenRead($Path)
        try { return ([BitConverter]::ToString($h.ComputeHash($fs)) -replace '-', '').ToLowerInvariant() }
        finally { $fs.Close() }
    } finally { $h.Dispose() }
}

function Get-HeldoutSource([string]$Check) {
    $map = @{
        'humanize'   = 'humanize.heldout_test.go'
        'afero'      = 'afero.heldout_test.go'
        'multierror' = 'multierror.heldout_test.go'
        'atomic'     = 'atomic.heldout_test.go'
        'difflib'    = 'difflib.heldout_test.go'
        'logr'       = 'logr.heldout_test.go'
    }
    if (-not $map.ContainsKey($Check)) { throw "Unknown held-out check: $Check" }
    $p = Join-Path $heldoutDir $map[$Check]
    if (-not (Test-Path -LiteralPath $p)) { throw "Held-out fixture missing: $p" }
    return Get-Content -Raw -LiteralPath $p
}

function Get-NativeGoArgs([object]$Entry) {
    # Manifest native_argv policy; Windows-only skip of precisely
    # ^TestNocmpIntegration$ for go-atomic (cmd.Env HOME-only -> C:\Windows
    # temp). Linux runs the full native suite. Held-out tests never excluded.
    if ($null -eq $Entry.native_argv -or @($Entry.native_argv).Count -eq 0) { throw "Manifest native_argv policy missing for $($Entry.id)" }
    $testArgs = @($Entry.native_argv)
    $scope = 'full'
    if ($Entry.id -eq 'go-difflib') { $scope = 'legacy-tests-without-vet' }
    if (($env:OS -eq 'Windows_NT') -and ($Entry.id -eq 'go-atomic')) {
        $testArgs += @('-skip', '^TestNocmpIntegration$')
        $scope = 'windows-scoped'
    }
    return [ordered]@{ Argv = $testArgs; Scope = $scope }
}

function Get-HeldoutGoArgs([object]$Entry, [string]$HeldoutFile) {
    # Held-out invocation: same manifest policy base with the Fabric held-out
    # selector. Legacy go-difflib (no go.mod) appends the held-out file to the
    # explicit source-file list; all other repos keep the package pattern.
    if ($null -eq $Entry.native_argv -or @($Entry.native_argv).Count -eq 0) { throw "Manifest native_argv policy missing for $($Entry.id)" }
    if ($Entry.id -eq 'go-difflib') {
        return @(Add-FabricV1HeldoutArguments -FullArgv @($Entry.native_argv) -HeldoutFile $HeldoutFile)
    }
    return @(Add-FabricV1HeldoutArguments -FullArgv @($Entry.native_argv))
}

function Invoke-GoTest([string]$WorkDir, [string[]]$TestArgs) {
    # $TestArgs is the full manifest policy argv starting with `go`; strip
    # and validate the executable here so local runs never become `go go`.
    # Recorded policy argv (logs, run.json, harness.toml) keeps the full form.
    $split = Split-GoInvocation -FullArgv $TestArgs
    $outFile = [System.IO.Path]::GetTempFileName()
    $errFile = [System.IO.Path]::GetTempFileName()
    try {
        Push-Location -LiteralPath $WorkDir
        try {
            & $GoExe @($split.Args) 1> $outFile 2> $errFile
            $exit = $LASTEXITCODE
        } finally { Pop-Location }
        $stdout = Get-Content -Raw -LiteralPath $outFile
        if ($null -eq $stdout) { $stdout = '' }
        $stderr = Get-Content -Raw -LiteralPath $errFile
        if ($null -eq $stderr) { $stderr = '' }
        return [pscustomobject]@{ ExitCode = $exit; Stdout = $stdout; Stderr = $stderr; Combined = ($stdout + "`n" + $stderr) }
    } finally {
        Remove-Item -LiteralPath $outFile, $errFile -Force -ErrorAction SilentlyContinue
    }
}

function Set-TaskVerificationConfig([string]$TaskPath, [object]$Entry) {
    # init defaults required checks to `go test ./...`; the task policy
    # requires the explicit manifest native_argv BEFORE any run is created
    # (Creation.Config binds required checks). Rewrites and hashes harness.toml.
    $native = Get-NativeGoArgs $Entry
    $configPath = Join-Path $TaskPath 'harness.toml'
    if (-not (Test-Path -LiteralPath $configPath)) { throw "harness.toml missing after init for $($Entry.id)" }
    $text = Get-Content -Raw -LiteralPath $configPath
    # init emits single-quoted argv; rewritten policy uses double quotes.
    # Read-VerificationArgv supports both simple quote forms and rejects
    # malformed, extra, or empty material explicitly.
    $currentItems = @(Read-VerificationArgv -Text $text)
    $argvDrift = Compare-Object $currentItems @('go', 'test', './...')
    if ($null -ne $argvDrift) {
        throw "harness.toml verification argv is not the init default for $($entry.id): $($currentItems -join ' ')"
    }
    $argvMatch = [regex]::Match($text, '(?m)^argv\s*=\s*\[[^\]]*\]')
    if (-not $argvMatch.Success) { throw "no verification argv line in harness.toml for $($entry.id); refusing to guess config shape" }
    $tomlArgv = 'argv = [' + (($native.Argv | ForEach-Object { '"' + ($_ -replace '"', '\"') + '"' }) -join ', ') + ']'
    $text = $text.Substring(0, $argvMatch.Index) + $tomlArgv + $text.Substring($argvMatch.Index + $argvMatch.Length)
    Set-Content -NoNewline -Encoding utf8 -LiteralPath $configPath $text
    return [ordered]@{
        Argv        = @($native.Argv)
        ArgvText    = ($native.Argv -join ' ')
        Scope       = $native.Scope
        ConfigSha   = (Get-FileSha256 $configPath).ToLowerInvariant()
    }
}

function Get-UsageMetrics([object]$Usage) {
    # internal/control RunUsage: invocations[].usage.provider_usage.
    # {input_tokens,output_tokens} (runtime.Usage *int64, null when
    # unavailable). receipt_matched rows are completed runtime invocations,
    # NOT a provider call count: provider_calls stays null (no actual count
    # exists in the contracts). Unknown metrics stay null, never zero.
    $inSum = 0; $outSum = 0; $inSeen = $false; $outSeen = $false
    $total = $null; $matched = $null
    if ($null -ne $Usage -and $null -ne $Usage.invocations) {
        $total = 0; $matched = 0
        foreach ($inv in @($Usage.invocations)) {
            $total++
            if ($inv.receipt_matched -eq $true) { $matched++ }
            $pu = $inv.usage.provider_usage
            if ($null -ne $pu) {
                if ($null -ne $pu.input_tokens) { $inSum += [int64]$pu.input_tokens; $inSeen = $true }
                if ($null -ne $pu.output_tokens) { $outSum += [int64]$pu.output_tokens; $outSeen = $true }
            }
        }
    }
    return [ordered]@{
        RuntimeInvocations          = $total
        RuntimeInvocationsCompleted = $matched
        ProviderCalls               = $null
        InputTokens                 = if ($inSeen) { $inSum } else { $null }
        OutputTokens                = if ($outSeen) { $outSum } else { $null }
    }
}

function Get-CandidateDiffMetrics([string]$Workspace, [string]$OutDir) {
    # Safe git diff collection from the candidate path (PR5 baseline lacks
    # `diff`). --no-optional-locks plus -c diff.autoRefreshIndex=false keeps
    # read-only observation from mutating the index hash the candidate
    # binding depends on. Exact limit: 1 MiB total like internal/cli diff.go.
    # Unknown metrics stay null, never zero. Untracked added files counted
    # separately.
    $noLock = @('--no-optional-locks', '-c', 'diff.autoRefreshIndex=false')
    $metrics = [ordered]@{
        diff_files = $null; diff_insertions = $null; diff_deletions = $null
        untracked_added_files = $null; diff_bytes = $null
        diff_byte_limit = $script:DiffByteLimit; diff_truncated = $null; diff_note = $null
        index_lock_avoidance = 'git --no-optional-locks -c diff.autoRefreshIndex=false'
    }
    try {
        $numstat = & $git @noLock -C $Workspace diff --no-ext-diff --numstat HEAD --
        if ($LASTEXITCODE -ne 0) { $metrics.diff_note = 'git diff numstat failed'; return $metrics }
        $files = 0; $ins = 0; $del = 0; $bytes = 0
        foreach ($line in ($numstat -join "`n" -split "`n")) {
            $t = $line.Trim()
            if ($t -eq '') { continue }
            $m = [regex]::Match($t, '^(\d+)\s+(\d+)\s+\S')
            if ($m.Success) { $files++; $ins += [int]$m.Groups[1].Value; $del += [int]$m.Groups[2].Value }
            elseif ($t -match '^-\s+-\s+\S') { $files++ }
            else { $metrics.diff_note = "unparsable numstat line: $t"; return $metrics }
            $bytes += [System.Text.Encoding]::UTF8.GetByteCount($t) + 1
        }
        $untracked = & $git @noLock -C $Workspace ls-files --others --exclude-standard -z
        if ($LASTEXITCODE -ne 0) { $metrics.diff_note = 'git untracked listing failed'; return $metrics }
        $uCount = @($untracked -split "`0" | Where-Object { $_ -ne '' }).Count
        $metrics.diff_files = $files; $metrics.diff_insertions = $ins; $metrics.diff_deletions = $del
        $metrics.untracked_added_files = $uCount; $metrics.diff_bytes = $bytes
        $metrics.diff_truncated = ($bytes -gt $script:DiffByteLimit)
        if ($bytes -gt $script:DiffByteLimit) { $metrics.diff_note = "diff exceeds $($script:DiffByteLimit)-byte limit" }
        Set-Content -NoNewline -Encoding utf8 (Join-Path $OutDir 'candidate-numstat.log') ($numstat -join "`n")
        return $metrics
    } catch {
        $metrics.diff_note = "diff collection error: $($_.Exception.Message)"
        return $metrics
    }
}

function Get-BinaryGoProvenance([string]$BinaryPath) {
    # Exact binary bytes (SHA-256) plus embedded Go build metadata from
    # `GoExe version -m`. Source HEAD (runner tree) is recorded separately;
    # the embedded vcs.revision/vcs.modified is the binary's own build claim,
    # never blindly trusted as the runner source.
    $sha = (Get-FileSha256 $BinaryPath).ToLowerInvariant()
    $raw = (& $GoExe version -m $BinaryPath 2>&1 | Out-String)
    if ($LASTEXITCODE -ne 0) { throw 'Cannot inspect embedded Go build metadata.' }
    $rev = $null; $mod = $null
    $m = [regex]::Match($raw, 'vcs\.revision=([0-9a-f]{40})')
    if ($m.Success) { $rev = $m.Groups[1].Value.ToLowerInvariant() }
    $m2 = [regex]::Match($raw, 'vcs\.modified=(true|false)')
    if ($m2.Success) { $mod = $m2.Groups[1].Value.ToLowerInvariant() }
    return [ordered]@{ Sha256 = $sha; VersionM = $raw.Trim(); VcsRevision = $rev; VcsModified = $mod }
}

function Get-RunnerSourceHashes() {
    # Hash ALL runner/helper source inputs + manifest and heldout assets, not
    # just the primary PS script. Relative repo-rooted paths -> sha256.
    $rel = @(
        'scripts/evaluate-v1.ps1',
        'evals/v1/harness/Classify-CheckOutput.ps1',
        'evals/v1/harness/Copy-CandidateTree.ps1',
        'evals/v1/harness/Go-ArgvPolicy.ps1',
        'evals/v1/harness/Toml-ArgvPolicy.ps1',
        'evals/v1/harness/Test-ArgvPolicy.ps1',
        'evals/v1/harness/Test-CopyFixtures.ps1',
        'evals/v1/harness/Test-TomlArgvPolicy.ps1',
        'evals/v1/manifest.json'
    )
    $out = [ordered]@{}
    foreach ($r in $rel) {
        $p = Join-Path $repoRoot ($r -replace '/', [string][System.IO.Path]::DirectorySeparatorChar)
        if (Test-Path -LiteralPath $p -PathType Leaf) { $out[$r] = (Get-FileSha256 $p).ToLowerInvariant() }
    }
    foreach ($f in @(Get-ChildItem -LiteralPath (Join-Path $repoRoot 'evals\v1\candidatecopy') -Filter '*.go' -File -ErrorAction SilentlyContinue)) {
        $out["evals/v1/candidatecopy/$($f.Name)"] = (Get-FileSha256 $f.FullName).ToLowerInvariant()
    }
    foreach ($f in @(Get-ChildItem -LiteralPath $heldoutDir -Filter '*.go' -File -ErrorAction SilentlyContinue)) {
        $out["evals/v1/heldout/$($f.Name)"] = (Get-FileSha256 $f.FullName).ToLowerInvariant()
    }
    return $out
}

function Get-ReviewCandidateId([object]$Snap) {
    if ($null -eq $Snap -or $null -eq $Snap.review -or $null -eq $Snap.review.result -or [string]::IsNullOrWhiteSpace($Snap.review.result.output)) { return $null }
    try { $v = ($Snap.review.result.output | ConvertFrom-Json) } catch { return $null }
    if ([string]::IsNullOrWhiteSpace($v.candidate_id)) { return $null }
    return [string]$v.candidate_id
}

function Invoke-CandidateCopy([string]$HelperExe, [string]$SnapshotFile, [string]$ExpectedCandidate, [string]$Destination, [string]$LogDir) {
    # Calls the explicit helper exe (fingerprinted by the caller, never
    # auto-built or relaunched). Helper failure is BLOCKED before any
    # acceptance tests, never rerun. A failed destination is retained.
    $outLog = Join-Path $LogDir 'candidatecopy.stdout.log'
    $errLog = Join-Path $LogDir 'candidatecopy.stderr.log'
    & $HelperExe --snapshot $SnapshotFile --expected-candidate $ExpectedCandidate --destination $Destination 1> $outLog 2> $errLog
    $exit = $LASTEXITCODE
    $stdout = ''
    if (Test-Path -LiteralPath $outLog) { $stdout = Get-Content -Raw -LiteralPath $outLog }
    if ($null -eq $stdout) { $stdout = '' }
    if ($exit -ne 0) {
        $stderr = ''
        if (Test-Path -LiteralPath $errLog) { $stderr = Get-Content -Raw -LiteralPath $errLog }
        throw "candidatecopy helper BLOCKED (exit $exit): $($stderr.Trim())"
    }
    if ([string]::IsNullOrWhiteSpace($stdout)) { throw 'candidatecopy helper produced no output' }
    try { $parsed = ($stdout.Trim() | ConvertFrom-Json) } catch { throw "candidatecopy output is not JSON: $($_.Exception.Message)" }
    if ([string]::IsNullOrWhiteSpace($parsed.candidate_id) -or $parsed.candidate_id -ne $ExpectedCandidate) {
        throw 'candidatecopy observation candidate_id does not match expected reviewed candidate'
    }
    return $parsed
}

function Get-RunGate([object]$Snap, [string]$FabricRunId, [string]$DiffCandidateId, [string]$ExpectedCandidateId) {
    # Strict PASS gate from exact Snapshot/Verification/review contracts plus
    # the helper-observed candidate binding. Candidate.ID() itself is computed
    # inside the Go helper (canonical hash); here the expected reviewed digest
    # must appear in verification plan, every observation, and the review
    # verdict (and in the native diff when present). PR5 has no fabric diff
    # (diffCandidateId nil) and relies on the helper binding alone.
    # Returns @{ Pass; FailReason; BlockedReason; ... }. Missing/uncertain
    # evidence is BLOCKED, never rerun. Complete negative evidence is FAIL:
    # FAIL observations / failed review / REPAIRING state are FAIL, while
    # unknown/pending statuses and non-READY incomplete states are BLOCKED.
    $gate = [ordered]@{
        Pass = $false; FailReason = $null; BlockedReason = $null
        State = $null; RepairAttempts = $null
        ReviewDecision = $null; ReviewFindings = $null
        ReviewCandidateId = $null; ReviewPlanId = $null
        VerificationObservations = $null; VerificationChecks = $null
        ExpectedCandidateId = $ExpectedCandidateId
    }
    if ($null -eq $Snap) { $gate.BlockedReason = 'inspect produced no snapshot'; return $gate }
    $gate.State = $Snap.state
    $gate.RepairAttempts = $Snap.repair_attempts
    if ([string]::IsNullOrWhiteSpace($ExpectedCandidateId) -or $ExpectedCandidateId -notmatch '^[0-9a-f]{64}$') { $gate.BlockedReason = 'missing expected reviewed candidate binding'; return $gate }
    if ($Snap.run_id -ne $FabricRunId) { $gate.BlockedReason = 'snapshot run_id binding mismatch'; return $gate }
    if ($null -eq $Snap.candidate) { $gate.BlockedReason = 'missing candidate observation'; return $gate }
    if ($null -eq $Snap.verification) { $gate.BlockedReason = 'missing verification state'; return $gate }
    if ($Snap.verification.pending -eq $true) { $gate.BlockedReason = 'verification pending'; return $gate }
    if ($Snap.verification.plan.candidate_id -ne $ExpectedCandidateId) { $gate.BlockedReason = 'verification plan candidate does not match expected reviewed candidate'; return $gate }
    $obs = @($Snap.verification.observations)
    $inv = @($Snap.verification.plan.invocations)
    $gate.VerificationObservations = $obs.Count
    $gate.VerificationChecks = $inv.Count
    if ($inv.Count -eq 0 -or $obs.Count -ne $inv.Count) { $gate.BlockedReason = 'verification observations do not cover all planned checks'; return $gate }
    foreach ($o in $obs) {
        if ($o.result.candidate_id -ne $ExpectedCandidateId -or $o.result.candidate_id -ne $Snap.verification.plan.candidate_id) { $gate.BlockedReason = 'verification result candidate binding mismatch'; return $gate }
        $st = [string]$o.result.status
        if ($st -eq 'PASS') { continue }
        elseif ($st -eq 'FAIL') { $gate.FailReason = 'verification observation not PASS'; return $gate }
        else { $gate.BlockedReason = "verification observation status is $st, not PASS"; return $gate }
    }
    if ($null -eq $Snap.review) { $gate.BlockedReason = 'missing review record'; return $gate }
    try { $verdict = ($Snap.review.result.output | ConvertFrom-Json) }
    catch { $gate.BlockedReason = "review output is not a JSON verdict: $($_.Exception.Message)"; return $gate }
    $gate.ReviewDecision = $verdict.decision
    $gate.ReviewFindings = @($verdict.findings).Count
    $gate.ReviewCandidateId = $verdict.candidate_id
    $gate.ReviewPlanId = $verdict.verification_plan_id
    if ($verdict.candidate_id -ne $ExpectedCandidateId) { $gate.BlockedReason = 'review candidate_id does not match expected reviewed candidate'; return $gate }
    if ($verdict.candidate_id -ne $Snap.verification.plan.candidate_id) { $gate.BlockedReason = 'review candidate_id does not match verification candidate'; return $gate }
    if ($verdict.verification_plan_id -ne $Snap.verification.plan_id) { $gate.BlockedReason = 'review verification_plan_id does not match verification plan'; return $gate }
    if ($null -ne $DiffCandidateId -and $DiffCandidateId -ne '' -and $DiffCandidateId -ne $verdict.candidate_id) {
        $gate.BlockedReason = 'diff candidate_id does not match reviewed candidate'; return $gate
    }
    if ($verdict.decision -ne 'approve' -or @($verdict.findings).Count -ne 0) { $gate.FailReason = 'reviewer did not approve with zero findings'; return $gate }
    if ($Snap.state -eq 'READY') { $gate.Pass = $true; return $gate }
    elseif ($Snap.state -eq 'REPAIRING') { $gate.FailReason = "snapshot state is REPAIRING after complete negative evidence"; return $gate }
    else { $gate.BlockedReason = "snapshot state is $($Snap.state), not READY"; return $gate }
}

if (-not (Test-Path $GoExe -PathType Leaf)) { throw "Pinned Go executable not found: $GoExe" }
$goVersion = (& $GoExe version).Trim()
if ($LASTEXITCODE -ne 0 -or $goVersion -notmatch 'go1\.27\.1') { throw "Expected Go 1.27.1, got $goVersion" }
$productHead = Get-GitText $repoRoot @('rev-parse', 'HEAD')
$productDirty = [bool]((Get-GitText $repoRoot @('status', '--porcelain')) -ne '')
$runnerSha = (Get-FileSha256 $PSCommandPath).ToLowerInvariant()

$entries = @($manifest.repositories)
if ($TaskIds -and $TaskIds.Count -gt 0) {
    $wanted = @($TaskIds)
    $entries = @($entries | Where-Object { $wanted -contains $_.id })
    if ($entries.Count -ne $wanted.Count) { throw 'Unknown -TaskIds selection.' }
}

if ($Action -eq 'Prepare') {
    # Prepare MUST NOT invoke any provider: no FabricExe/CodexExe run, no model
    # calls, no run-task script. Only git + go test discriminator.
    $now = [DateTime]::UtcNow.ToString('yyyyMMddTHHmmssZ')
    if (-not $RunId) { $RunId = "$now-$([Guid]::NewGuid().ToString('N').Substring(0, 8))" }
    if ($RunId -notmatch '^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$') { throw 'RunId must contain only letters, digits, dot, underscore, or dash.' }
    $runPath = Join-Path (Join-Path $RunRoot $manifest.suite_id) $RunId
    if (Test-Path $runPath) { throw "Run directory already exists (new run nonce required): $runPath" }
    $taskRoot = Join-Path $runPath 'tasks'
    $baselineRoot = Join-Path $runPath 'baseline-checks'
    New-Item -ItemType Directory -Force $taskRoot, $baselineRoot | Out-Null
    $records = @()
    $started = [Diagnostics.Stopwatch]::StartNew()
    foreach ($entry in $entries) {
        $taskPath = Join-Path (Join-Path $taskRoot $entry.id) 'repo'
        $basePath = Join-Path $baselineRoot $entry.id
        New-Item -ItemType Directory -Force (Split-Path $taskPath), $basePath | Out-Null
        & $git clone -q --no-checkout --filter=blob:none $entry.url $taskPath
        if ($LASTEXITCODE -ne 0) { throw "Clone failed for $($entry.id)" }
        & $git -C $taskPath checkout -q --detach $entry.sha
        if ($LASTEXITCODE -ne 0) { throw "Pinned checkout failed for $($entry.id)" }
        $actual = Get-GitText $taskPath @('rev-parse', 'HEAD')
        if ($actual -ne $entry.sha) { throw "SHA mismatch for $($entry.id): $actual" }
        $excludePath = Join-Path $taskPath '.git\info\exclude'
        Add-Content -LiteralPath $excludePath -Value "`n.harness/`nharness.toml`n" -Encoding utf8
        if ((Get-GitText $taskPath @('status', '--porcelain')) -ne '') { throw "Task checkout dirty after exclude for $($entry.id)" }
        & $git clone -q --no-hardlinks --local $taskPath $basePath
        if ($LASTEXITCODE -ne 0) { throw "Baseline copy failed for $($entry.id)" }
        $baseHead = Get-GitText $basePath @('rev-parse', 'HEAD')

        $native = Get-NativeGoArgs $entry
        $preflight = Invoke-GoTest $basePath $native.Argv
        Set-Content -NoNewline -Encoding utf8 (Join-Path $basePath 'preflight.stdout.log') $preflight.Stdout
        Set-Content -NoNewline -Encoding utf8 (Join-Path $basePath 'preflight.stderr.log') $preflight.Stderr
        Set-Content -NoNewline -Encoding utf8 (Join-Path $basePath 'preflight.args.log') ($native.Argv -join ' ')
        Set-Content -NoNewline -Encoding utf8 (Join-Path $basePath 'preflight.scope.log') $native.Scope
        if ($preflight.ExitCode -ne 0) {
            throw "$($entry.id) preflight FAILED on pinned baseline (exit $($preflight.ExitCode), scope $($native.Scope), args: $($native.Argv -join ' ')); see baseline-checks/$($entry.id)/preflight.*.log"
        }

        if ($entry.check -eq 'difflib') {
            $heldoutRel = 'difflib/fabric_v1_heldout_test.go'
            $checkPath = Join-Path $basePath $heldoutRel
            $heldoutArg = $heldoutRel
        } else {
            $heldoutRel = 'fabric_v1_heldout_test.go'
            $checkPath = Join-Path $basePath $heldoutRel
            $heldoutArg = ''
        }
        if (Test-Path -LiteralPath (Join-Path $taskPath $heldoutRel)) { throw "Held-out test leaked into task checkout for $($entry.id)" }
        Set-Content -NoNewline -Encoding utf8 $checkPath (Get-HeldoutSource $entry.check)
        try {
            $heldout = Invoke-GoTest $basePath (Get-HeldoutGoArgs $entry $heldoutArg)
        } finally { Remove-Item -LiteralPath $checkPath -Force }
        Set-Content -NoNewline -Encoding utf8 (Join-Path $basePath 'heldout.stdout.log') $heldout.Stdout
        Set-Content -NoNewline -Encoding utf8 (Join-Path $basePath 'heldout.stderr.log') $heldout.Stderr
        Set-Content -NoNewline -Encoding utf8 (Join-Path $basePath 'heldout.combined.log') $heldout.Combined
        $classification = Get-FabricV1CheckClassification -TaskId $entry.id -ExitCode $heldout.ExitCode -CombinedOutput $heldout.Combined
        if ($classification -notin @('targeted_assertion', 'expected_missing_api')) {
            throw "$($entry.id) held-out baseline classification=$classification (exit $($heldout.ExitCode)); expected targeted_assertion or expected_missing_api."
        }
        if ($classification -eq 'expected_missing_api' -and $entry.id -notin @('go-multierror', 'go-atomic')) {
            throw "$($entry.id) missing-API baseline only expected for go-multierror ErrorsSnapshot and go-atomic MarshalText/UnmarshalText."
        }
        $records += [ordered]@{
            task_id                = $entry.id
            repository_url         = $entry.url
            source_sha             = $actual
            baseline_copy_sha      = $baseHead
            license                = $entry.license
            task_checkout          = $taskPath
            heldout_baseline       = 'FAIL_EXPECTED'
            heldout_classification = $classification
            heldout_exit_code      = $heldout.ExitCode
            preflight_exit_code    = $preflight.ExitCode
            preflight_args         = ($native.Argv -join ' ')
            native_argv_policy     = (@($entry.native_argv) -join ' ')
            native_test_scope      = $native.Scope
            windows_exclusion      = '^TestNocmpIntegration$ (windows-only, go-atomic; rationale in manifest native_verification)'
            task_completion        = 'NOT RUN'
            provider_calls         = 0
            input_tokens           = $null
            output_tokens          = $null
            review_findings        = $null
            repair_attempts        = $null
            diff_files             = $null
            diff_insertions        = $null
            diff_deletions         = $null
            untracked_added_files  = $null
            human_interventions    = $null
            elapsed_ms             = $null
            candidate_sha          = $null
            candidate_tree_sha256  = $null
        }
    }
    $started.Stop()
    $record = [ordered]@{
        schema_version                  = 1
        suite_id                      = $manifest.suite_id
        run_id                        = $RunId
        mode                          = 'prepare'
        created_utc                   = [DateTime]::UtcNow.ToString('o')
        product_head                  = $productHead
        product_tree_dirty_at_prepare = $productDirty
        runner_sha256                 = $runnerSha
        go_version                    = $goVersion
        preparation_elapsed_ms        = $started.ElapsedMilliseconds
        provider_calls                = 0
        heldout_checks                = 'baseline discriminator only; task checkouts never contained held-out tests; exact stdout/stderr under baseline-checks/<id>/*.log'
        native_verification           = $manifest.native_verification
        native_test_scope_note        = 'Windows verification is scoped, not the full upstream suite: Windows-only -skip ^TestNocmpIntegration$ applies to go-atomic preflight and final native package verification; Linux runs the full native suite; legacy go-difflib uses explicit source-file argv; held-out task tests are never excluded.'
        raw_transcripts_retained      = $false
        credentials_retained          = $false
        task_records                  = $records
    }
    $record | ConvertTo-Json -Depth 8 | Set-Content -Encoding utf8 (Join-Path $runPath 'run.json')
    Write-Output "PASS prepared $($records.Count) pinned tasks; provider calls=0"
    Write-Output "Run record: $(Join-Path $runPath 'run.json')"
    # The held-out discriminator intentionally fails on the unmodified base.
    # Do not leak that expected native exit code as the successful runner exit.
    exit 0
}

# ---------------- Evaluate (provider-gated) ----------------
if ([string]::IsNullOrWhiteSpace($FabricExe)) { throw '-FabricExe is required for Evaluate.' }
if ([string]::IsNullOrWhiteSpace($CodexExe)) { throw '-CodexExe is required for Evaluate.' }
if ([string]::IsNullOrWhiteSpace($Model)) { throw '-Model is required for Evaluate.' }
if ([string]::IsNullOrWhiteSpace($Effort)) { throw '-Effort must not be empty; default high, user-requested value accepted and persisted.' }
if (-not $RunId) { throw '-RunId of a prepared run is required for Evaluate.' }
if ($RunId -notmatch '^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$') { throw 'RunId must contain only letters, digits, dot, underscore, or dash (same bound as Prepare).' }
if ([string]::IsNullOrWhiteSpace($CandidateCopyExe)) { throw '-CandidateCopyExe is required for Evaluate (explicit helper path, fingerprinted, never auto-built).' }
$FabricExe = (Resolve-Path -LiteralPath $FabricExe).Path
$CodexExe = (Resolve-Path -LiteralPath $CodexExe).Path
$CandidateCopyExe = (Resolve-Path -LiteralPath $CandidateCopyExe).Path
$fabricProv = Get-BinaryGoProvenance $FabricExe
if ($EvalMode -eq 'Native' -and ($null -eq $fabricProv.VcsRevision -or $fabricProv.VcsModified -ne 'false')) {
    throw 'Native evaluation requires a binary with an embedded commit and clean VCS metadata.'
}
$fabricSha = $fabricProv.Sha256
$copyProv = Get-BinaryGoProvenance $CandidateCopyExe
$copySha = $copyProv.Sha256
$runtimeSha = (Get-FileSha256 $CodexExe).ToLowerInvariant()
if ($EvalMode -eq 'PR5Matched') {
    if ([string]::IsNullOrWhiteSpace($PR5BaselineExe)) { throw '-PR5BaselineExe is required for PR5Matched mode.' }
    $baselineExe = (Resolve-Path -LiteralPath $PR5BaselineExe).Path
    $baselineProv = Get-BinaryGoProvenance $baselineExe
    $baselineSha = $baselineProv.Sha256
    $pr5Script = if ([string]::IsNullOrWhiteSpace($PR5BaselineScript)) { $defaultRunTaskScript } else { (Resolve-Path -LiteralPath $PR5BaselineScript).Path }
    if (-not (Test-Path -LiteralPath $pr5Script)) { throw "PR5 baseline script not found: $pr5Script" }
    $pr5ScriptSha = (Get-FileSha256 $pr5Script).ToLowerInvariant()
}
# Evaluation-time provenance: current runner HEAD/dirty is the runner tree
# (which is dirty due to untracked helper source), NOT the clean binary's
# build source. Binary vcs.revision/vcs.modified above is the binary's own
# embedded claim. Do not claim clean binary source from a dirty runner HEAD.
$evalProductHead = Get-GitText $repoRoot @('rev-parse', 'HEAD')
$evalProductDirty = [bool]((Get-GitText $repoRoot @('status', '--porcelain')) -ne '')
$runnerSourceHashes = Get-RunnerSourceHashes
$runPath = Join-Path (Join-Path $RunRoot $manifest.suite_id) $RunId
$runJsonPath = Join-Path $runPath 'run.json'
if (-not (Test-Path -LiteralPath $runJsonPath)) { throw "Prepared run not found: $runJsonPath. Prepare first with a new run nonce." }
$priorRunJsonSha = (Get-FileSha256 $runJsonPath).ToLowerInvariant()
$prior = Get-Content -Raw -LiteralPath $runJsonPath | ConvertFrom-Json
$taskPins = @($entries | ForEach-Object { [ordered]@{ task_id = $_.id; source_sha = $_.sha; url = $_.url } })
# Optional known external build receipt: validate binary hash when present,
# never blindly trust stamped source.
$buildReceipt = $null
if (-not [string]::IsNullOrWhiteSpace($NativeBuildReceiptPath)) {
    $buildReceiptPath = (Resolve-Path -LiteralPath $NativeBuildReceiptPath -ErrorAction Stop).Path
    try {
        $receiptRaw = Get-Content -Raw -LiteralPath $buildReceiptPath
        $receiptJson = ($receiptRaw | ConvertFrom-Json)
        $receiptHash = $null
        foreach ($k in @('binary_sha256', 'fabric_sha256', 'sha256', 'binary_hash')) {
            if ($null -ne $receiptJson.$k -and $receiptJson.$k -ne '') { $receiptHash = [string]$receiptJson.$k; break }
        }
        $buildReceipt = [ordered]@{
            path = $buildReceiptPath
            receipt_sha256 = (Get-FileSha256 $buildReceiptPath).ToLowerInvariant()
            claimed_binary_sha256 = $receiptHash
            matches_native_binary = if ($null -ne $receiptHash) { ($receiptHash.ToLowerInvariant() -eq $fabricSha) } else { $null }
        }
        if ($null -ne $receiptHash -and $receiptHash.ToLowerInvariant() -ne $fabricSha) {
            throw "known build receipt binary hash does not match Native FabricExe bytes; refusing to trust stamped source"
        }
        if ($EvalMode -eq 'Native' -and ($receiptJson.source_commit -ne $fabricProv.VcsRevision -or
            $receiptJson.source_clean_before -ne $true -or $receiptJson.source_clean_after -ne $true)) {
            throw 'Known native build receipt disagrees with embedded source identity or clean build evidence.'
        }
    } catch {
        throw "build receipt validation failed: $($_.Exception.Message)"
    }
}
$nowEval = [DateTime]::UtcNow.ToString('yyyyMMddTHHmmssZ')
$evalId = "eval-$nowEval-$([Guid]::NewGuid().ToString('N').Substring(0, 8))"
$evalRoot = Join-Path $runPath $evalId
if (Test-Path -LiteralPath $evalRoot) { throw "Eval directory already exists (new eval nonce required): $evalRoot" }
New-Item -ItemType Directory -Path $evalRoot | Out-Null
$results = @()
foreach ($entry in $entries) {
    $taskPath = Join-Path (Join-Path (Join-Path $runPath 'tasks') $entry.id) 'repo'
    $taskOutDir = Join-Path $evalRoot $entry.id
    New-Item -ItemType Directory -Force $taskOutDir | Out-Null
    $sw = [Diagnostics.Stopwatch]::StartNew()
    $result = [ordered]@{
        task_id = $entry.id; eval_mode = $EvalMode; terminal_state = 'BLOCKED'
        blocked_reason = $null; fail_reason = $null
    }
    try {
        if (-not (Test-Path -LiteralPath $taskPath)) { throw "Task checkout missing for $($entry.id): $taskPath" }
        if ((Get-GitText $taskPath @('rev-parse', 'HEAD')) -ne $entry.sha) { throw "Task checkout is not the fresh pinned SHA for $($entry.id); matched comparison requires fresh identical repo/task clones" }
        if (Test-Path -LiteralPath (Join-Path $taskPath 'fabric_v1_heldout_test.go')) { throw "Held-out test present in task clone before candidate completion for $($entry.id)" }
        $result.source_sha = $entry.sha

        if ($EvalMode -eq 'Native') {
            # Native new evaluation: explicit init, task-policy verification
            # config BEFORE run creation, autonomous run, inspect/usage/diff.
            if ((Test-Path -LiteralPath (Join-Path $taskPath 'harness.toml')) -or (Test-Path -LiteralPath (Join-Path $taskPath '.harness'))) {
                throw "Task checkout for $($entry.id) already initialized; Evaluate requires a fresh prepared clone"
            }
            & $FabricExe --root $taskPath init --codex $CodexExe --model $Model --effort $Effort 1> (Join-Path $taskOutDir 'fabric-init.stdout.log') 2> (Join-Path $taskOutDir 'fabric-init.stderr.log')
            if ($LASTEXITCODE -ne 0) { throw "fabric init failed for $($entry.id); see fabric-init.*.log" }
            $result.effort = $Effort
            $verPolicy = Set-TaskVerificationConfig $taskPath $entry
            $result.verification_argv = $verPolicy.ArgvText
            $result.verification_config_sha256 = $verPolicy.ConfigSha
            $result.native_test_scope = $verPolicy.Scope
            Set-Content -NoNewline -Encoding utf8 (Join-Path $taskOutDir 'verification-argv.log') $verPolicy.ArgvText
            & $FabricExe --root $taskPath run --autonomous $entry.task 1> (Join-Path $taskOutDir 'fabric-run.stdout.log') 2> (Join-Path $taskOutDir 'fabric-run.stderr.log')
            if ($LASTEXITCODE -ne 0) { throw "fabric run --autonomous failed for $($entry.id); see fabric-run.*.log" }
            $runOut = Get-Content -Raw -LiteralPath (Join-Path $taskOutDir 'fabric-run.stdout.log')
            if ([string]::IsNullOrWhiteSpace($runOut)) { throw "fabric run produced no output for $($entry.id)" }
            $parsed = ($runOut | ConvertFrom-Json)
            $fabricRunId = $parsed.run_id
            if ([string]::IsNullOrWhiteSpace($fabricRunId) -or $fabricRunId -notmatch '^[0-9a-f]{64}$') { throw "fabric run output has no exact run identity for $($entry.id)" }
            $result.fabric_run_id = $fabricRunId
            $result.fabric_sha256 = $fabricSha
            & $FabricExe --root $taskPath inspect $fabricRunId 1> (Join-Path $taskOutDir 'fabric-inspect.stdout.log') 2> (Join-Path $taskOutDir 'fabric-inspect.stderr.log')
            if ($LASTEXITCODE -ne 0) { throw "fabric inspect failed for $($entry.id); see fabric-inspect.*.log" }
            & $FabricExe --root $taskPath usage $fabricRunId 1> (Join-Path $taskOutDir 'fabric-usage.stdout.log') 2> (Join-Path $taskOutDir 'fabric-usage.stderr.log')
            if ($LASTEXITCODE -ne 0) { throw "fabric usage failed for $($entry.id); see fabric-usage.*.log" }
            & $FabricExe --root $taskPath diff $fabricRunId 1> (Join-Path $taskOutDir 'fabric-diff.stdout.log') 2> (Join-Path $taskOutDir 'fabric-diff.stderr.log')
            if ($LASTEXITCODE -ne 0) { throw "fabric diff failed for $($entry.id); see fabric-diff.*.log" }
            $snap = (Get-Content -Raw -LiteralPath (Join-Path $taskOutDir 'fabric-inspect.stdout.log') | ConvertFrom-Json)
            $usage = (Get-Content -Raw -LiteralPath (Join-Path $taskOutDir 'fabric-usage.stdout.log') | ConvertFrom-Json)
            $diffOut = (Get-Content -Raw -LiteralPath (Join-Path $taskOutDir 'fabric-diff.stdout.log') | ConvertFrom-Json)
            $diffCandidateId = $diffOut.candidate_id
            $result.fabric_diff_candidate_id = $diffCandidateId
            $result.diff_source = 'fabric-diff'
            # Native autonomous journey performs zero human interventions.
            $result.human_interventions = @()
            $result.model = $Model
            $result.runtime_sha256 = $runtimeSha
            $result.config_sha256 = (Get-FileSha256 (Join-Path $taskPath 'harness.toml')).ToLowerInvariant()
            $metrics = Get-UsageMetrics $usage
            $result.runtime_invocations = $metrics.RuntimeInvocations
            $result.runtime_invocations_completed = $metrics.RuntimeInvocationsCompleted
            $result.provider_calls = $metrics.ProviderCalls
            $result.input_tokens = $metrics.InputTokens
            $result.output_tokens = $metrics.OutputTokens
        } else {
            # PR5 matched evaluation: explicit supplied old script; init with
            # the baseline exe/model first, same verification policy and
            # model/runtime; inspect+usage with the baseline executable;
            # exact run identity from the script result (never the first
            # arbitrary hash); diff collected from the candidate path because
            # the baseline lacks a diff command.
            if ((Test-Path -LiteralPath (Join-Path $taskPath 'harness.toml')) -or (Test-Path -LiteralPath (Join-Path $taskPath '.harness'))) {
                throw "Task checkout for $($entry.id) already initialized; Evaluate requires a fresh prepared clone"
            }
            & $baselineExe --root $taskPath init --codex $CodexExe --model $Model --effort $Effort 1> (Join-Path $taskOutDir 'baseline-init.stdout.log') 2> (Join-Path $taskOutDir 'baseline-init.stderr.log')
            if ($LASTEXITCODE -ne 0) { throw "baseline init failed for $($entry.id); see baseline-init.*.log" }
            $result.effort = $Effort
            $verPolicy = Set-TaskVerificationConfig $taskPath $entry
            $result.verification_argv = $verPolicy.ArgvText
            $result.verification_config_sha256 = $verPolicy.ConfigSha
            $result.native_test_scope = $verPolicy.Scope
            Set-Content -NoNewline -Encoding utf8 (Join-Path $taskOutDir 'verification-argv.log') $verPolicy.ArgvText
            # Old run-task.ps1 calls git diff --stat itself; do not edit that
            # baseline script. Scope diff.autoRefreshIndex=false via
            # GIT_CONFIG_COUNT appends so its reads cannot mutate the index
            # hash. Pre-existing entries are preserved and the environment is
            # restored in finally even on failure.
            $savedGitConfigCount = $env:GIT_CONFIG_COUNT
            $appendIndex = 0
            if (-not [string]::IsNullOrEmpty($savedGitConfigCount)) {
                $parsedCount = 0
                if (-not [int]::TryParse($savedGitConfigCount, [ref]$parsedCount)) { throw "existing GIT_CONFIG_COUNT is not an integer: $savedGitConfigCount" }
                $appendIndex = $parsedCount
            }
            $scopedKey = "GIT_CONFIG_KEY_$appendIndex"
            $scopedValue = "GIT_CONFIG_VALUE_$appendIndex"
            if ((Test-Path "env:$scopedKey") -or (Test-Path "env:$scopedValue")) { throw "GIT_CONFIG slot collision at index $appendIndex" }
            try {
                $env:GIT_CONFIG_COUNT = ($appendIndex + 1).ToString()
                Set-Item -Path "env:$scopedKey" -Value 'diff.autoRefreshIndex'
                Set-Item -Path "env:$scopedValue" -Value 'false'
                $taskResult = & $pr5Script -Fabric $baselineExe -Repository $taskPath -Objective $entry.task -ApprovePlan -ApproveChanges -NonInteractive
            } finally {
                Remove-Item -Path "env:$scopedKey" -ErrorAction SilentlyContinue
                Remove-Item -Path "env:$scopedValue" -ErrorAction SilentlyContinue
                if ($null -eq $savedGitConfigCount) { Remove-Item -Path 'env:GIT_CONFIG_COUNT' -ErrorAction SilentlyContinue }
                else { $env:GIT_CONFIG_COUNT = $savedGitConfigCount }
            }
            $result.pr5_git_config_scope = 'appended diff.autoRefreshIndex=false via GIT_CONFIG_COUNT; pre-existing entries preserved and restored in finally'
            $exactId = @($taskResult) | Where-Object { $null -ne $_ -and $_.run_id -match '^[0-9a-f]{64}$' } | Select-Object -Last 1
            if ($null -eq $exactId) {
                throw "PR5 script returned no exact run identity for $($entry.id)"
            }
            $fabricRunId = [string]$exactId.run_id
            $result.fabric_run_id = $fabricRunId
            $result.baseline_sha256 = $baselineSha
            $result.baseline_script = $pr5Script
            Set-Content -NoNewline -Encoding utf8 (Join-Path $taskOutDir 'run-task.result.log') ("run_id=$fabricRunId state=$($taskResult.state)")
            & $baselineExe --root $taskPath inspect $fabricRunId 1> (Join-Path $taskOutDir 'fabric-inspect.stdout.log') 2> (Join-Path $taskOutDir 'fabric-inspect.stderr.log')
            if ($LASTEXITCODE -ne 0) { throw "baseline inspect failed for $($entry.id); see fabric-inspect.*.log" }
            & $baselineExe --root $taskPath usage $fabricRunId 1> (Join-Path $taskOutDir 'fabric-usage.stdout.log') 2> (Join-Path $taskOutDir 'fabric-usage.stderr.log')
            if ($LASTEXITCODE -ne 0) { throw "baseline usage failed for $($entry.id); see fabric-usage.*.log" }
            $snap = (Get-Content -Raw -LiteralPath (Join-Path $taskOutDir 'fabric-inspect.stdout.log') | ConvertFrom-Json)
            $usage = (Get-Content -Raw -LiteralPath (Join-Path $taskOutDir 'fabric-usage.stdout.log') | ConvertFrom-Json)
            $diffCandidateId = $null
            $result.diff_source = 'candidate-path-git'
            # Exactly two required approval gates, preauthorized, no invented
            # human identity, no repair loop started by this runner.
            $result.human_interventions = @(
                [ordered]@{ gate = 'plan'; mechanism = 'run-task.ps1 -ApprovePlan'; actor = 'preauthorized'; human_identity = $null }
                [ordered]@{ gate = 'file'; mechanism = 'run-task.ps1 -ApproveChanges -NonInteractive'; actor = 'preauthorized'; human_identity = $null }
            )
            $result.model = $Model
            $result.runtime_sha256 = $runtimeSha
            $result.config_sha256 = (Get-FileSha256 (Join-Path $taskPath 'harness.toml')).ToLowerInvariant()
            $metrics = Get-UsageMetrics $usage
            $result.runtime_invocations = $metrics.RuntimeInvocations
            $result.runtime_invocations_completed = $metrics.RuntimeInvocationsCompleted
            $result.provider_calls = $metrics.ProviderCalls
            $result.input_tokens = $metrics.InputTokens
            $result.output_tokens = $metrics.OutputTokens
        }

        # Workspace: exact contract path snap.workspace.request.path, bound to
        # this run. No object-string coercion, no task-checkout fallback.
        $wsReq = $snap.workspace.request
        if ($null -eq $wsReq -or [string]::IsNullOrWhiteSpace($wsReq.path)) { throw "snapshot has no isolated workspace for $($entry.id)" }
        $candidateWorkspace = [string]$wsReq.path
        $expectedRoot = Join-Path $taskPath (Join-Path '.harness' 'worktrees')
        if ($wsReq.run_id -ne $fabricRunId) { throw "workspace request is not bound to run $fabricRunId for $($entry.id)" }
        if (-not $candidateWorkspace.StartsWith($expectedRoot + [string][System.IO.Path]::DirectorySeparatorChar)) {
            throw "isolated workspace is outside the task worktree registry for $($entry.id)"
        }
        if (-not (Test-Path -LiteralPath $candidateWorkspace -PathType Container)) { throw "isolated workspace path missing for $($entry.id)" }
        if ($snap.candidate.head -ne $entry.sha) { throw "candidate HEAD is not the pinned source commit for $($entry.id)" }
        $result.candidate_workspace = $candidateWorkspace

        # Reviewed candidate binding: the Go helper owns Candidate.ID()
        # computation (canonical hash). Extract the reviewed digest here and
        # require it in both the helper and Get-RunGate. Native diff digest
        # must still agree when present; PR5 has no fabric diff (nil).
        $expectedCandidateId = Get-ReviewCandidateId $snap
        if ([string]::IsNullOrWhiteSpace($expectedCandidateId)) {
            $result.blocked_reason = 'missing reviewed candidate binding for helper'
            $sw.Stop()
            $result.elapsed_ms = $sw.ElapsedMilliseconds
            $results += $result
            continue
        }
        $result.review_candidate_id_expected = $expectedCandidateId

        # Candidate-bound observation/copy under a read lease. Helper failure
        # is BLOCKED before any acceptance tests, never rerun; a failed
        # destination is retained, never recursively deleted.
        $copyStage = Join-Path $taskOutDir 'acceptance-copy'
        try {
            $copyObs = Invoke-CandidateCopy $CandidateCopyExe (Join-Path $taskOutDir 'fabric-inspect.stdout.log') $expectedCandidateId $copyStage $taskOutDir
        } catch {
            $result.blocked_reason = "candidatecopy helper BLOCKED: $($_.Exception.Message)"
            $sw.Stop()
            $result.elapsed_ms = $sw.ElapsedMilliseconds
            $results += $result
            continue
        }
        $copyPath = [string]$copyObs.destination
        $result.acceptance_copy_path = $copyPath
        $result.acceptance_copy_identity = [string]$copyObs.files_hash
        $result.acceptance_copy_files = [int]$copyObs.file_count
        $result.acceptance_copy_candidate_id = [string]$copyObs.candidate_id
        $result.acceptance_copy_workspace = [string]$copyObs.workspace
        $copyObs | ConvertTo-Json -Depth 6 | Set-Content -Encoding utf8 (Join-Path $taskOutDir 'acceptance-copy.manifest.json')
        if ([string]$copyObs.candidate_id -ne $expectedCandidateId) {
            $result.blocked_reason = 'helper observation candidate_id does not match expected reviewed candidate'
            $sw.Stop()
            $result.elapsed_ms = $sw.ElapsedMilliseconds
            $results += $result
            continue
        }

        $gate = Get-RunGate $snap $fabricRunId $diffCandidateId $expectedCandidateId
        $result.snapshot_state = $gate.State
        $result.repair_attempts = $gate.RepairAttempts
        $result.review_decision = $gate.ReviewDecision
        $result.review_findings = $gate.ReviewFindings
        $result.review_candidate_id = $gate.ReviewCandidateId
        $result.review_verification_plan_id = $gate.ReviewPlanId
        $result.verification_observations = $gate.VerificationObservations
        $result.verification_checks = $gate.VerificationChecks
        if ($null -ne $gate.BlockedReason) {
            # Missing/uncertain evidence: BLOCKED, never automatically rerun.
            $result.blocked_reason = $gate.BlockedReason
            $sw.Stop()
            $result.elapsed_ms = $sw.ElapsedMilliseconds
            $results += $result
            continue
        }

        # Tests run only on the helper's copied bytes; the original worktree
        # is never mutated. The held-out check is added only to the copy and
        # removed afterwards. Destination manifest identity (files_hash) was
        # already verified by the helper against the source Candidate.
        # Final native upstream verification on the copy, then held-out.
        $native = Get-NativeGoArgs $entry
        $nativeVerify = Invoke-GoTest $copyPath $native.Argv
        Set-Content -NoNewline -Encoding utf8 (Join-Path $taskOutDir 'native-verify.stdout.log') $nativeVerify.Stdout
        Set-Content -NoNewline -Encoding utf8 (Join-Path $taskOutDir 'native-verify.stderr.log') $nativeVerify.Stderr
        Set-Content -NoNewline -Encoding utf8 (Join-Path $taskOutDir 'native-verify.args.log') ($native.Argv -join ' ')
        $result.native_verify_exit_code = $nativeVerify.ExitCode
        $result.native_verify_args = ($native.Argv -join ' ')

        if ($entry.check -eq 'difflib') {
            $checkDest = Join-Path (Join-Path $copyPath 'difflib') 'fabric_v1_heldout_test.go'
            $heldoutFileArg = 'difflib/fabric_v1_heldout_test.go'
        } else {
            $checkDest = Join-Path $copyPath 'fabric_v1_heldout_test.go'
            $heldoutFileArg = ''
        }
        Copy-Item -LiteralPath (Join-Path $heldoutDir "$($entry.check).heldout_test.go") -Destination $checkDest -Force
        try {
            $heldout = Invoke-GoTest $copyPath (Get-HeldoutGoArgs $entry $heldoutFileArg)
        } finally { Remove-Item -LiteralPath $checkDest -Force -ErrorAction SilentlyContinue }
        Set-Content -NoNewline -Encoding utf8 (Join-Path $taskOutDir 'acceptance.stdout.log') $heldout.Stdout
        Set-Content -NoNewline -Encoding utf8 (Join-Path $taskOutDir 'acceptance.stderr.log') $heldout.Stderr
        $acceptClass = Get-FabricV1CheckClassification -TaskId $entry.id -ExitCode $heldout.ExitCode -CombinedOutput $heldout.Combined
        $result.acceptance_classification = $acceptClass
        $result.acceptance_exit_code = $heldout.ExitCode
        if ($acceptClass -eq 'setup_error') {
            $result.blocked_reason = 'acceptance setup_error; environment failure is never an intended result'
            $sw.Stop()
            $result.elapsed_ms = $sw.ElapsedMilliseconds
            $results += $result
            continue
        }

        $diffMetrics = Get-CandidateDiffMetrics $candidateWorkspace $taskOutDir
        $result.diff_files = $diffMetrics.diff_files
        $result.diff_insertions = $diffMetrics.diff_insertions
        $result.diff_deletions = $diffMetrics.diff_deletions
        $result.untracked_added_files = $diffMetrics.untracked_added_files
        $result.diff_bytes = $diffMetrics.diff_bytes
        $result.diff_byte_limit = $diffMetrics.diff_byte_limit
        $result.diff_truncated = $diffMetrics.diff_truncated
        $result.diff_note = $diffMetrics.diff_note
        $result.index_lock_avoidance = $diffMetrics.index_lock_avoidance

        # PASS requires the Fabric gate AND native upstream verification AND
        # held-out PASS. Complete-but-negative evidence is FAIL; anything
        # missing/uncertain is BLOCKED above.
        if ($gate.Pass -and $nativeVerify.ExitCode -eq 0 -and $heldout.ExitCode -eq 0 -and $acceptClass -eq 'pass') {
            $result.terminal_state = 'PASS'
        } elseif ($nativeVerify.ExitCode -ne 0 -or $heldout.ExitCode -ne 0 -or $null -ne $gate.FailReason) {
            $result.terminal_state = 'FAIL'
            $result.fail_reason = if ($null -ne $gate.FailReason) { $gate.FailReason } elseif ($nativeVerify.ExitCode -ne 0) { 'native upstream verification failed' } else { 'held-out acceptance failed' }
        } else {
            $result.blocked_reason = 'undetermined outcome; treated as BLOCKED without rerun'
        }
        $sw.Stop()
        $result.elapsed_ms = $sw.ElapsedMilliseconds
        $results += $result
    } catch {
        # Exact error preserved; uncertain provider work is never rerun here.
        $result.blocked_reason = "evaluate error: $($_.Exception.Message)"
        try { $sw.Stop() } catch { }
        try { $result.elapsed_ms = $sw.ElapsedMilliseconds } catch { }
        $results += $result
    }
}
$evalRecord = [ordered]@{
    schema_version           = 1
    suite_id               = $prior.suite_id
    run_id                 = $prior.run_id
    eval_id                = $evalId
    eval_mode              = $EvalMode
    model                  = $Model
    effort                 = $Effort
    fabric_sha256          = if ($EvalMode -eq 'Native') { $fabricSha } else { $null }
    fabric_go_version_m    = if ($EvalMode -eq 'Native') { $fabricProv.VersionM } else { $null }
    fabric_vcs_revision    = if ($EvalMode -eq 'Native') { $fabricProv.VcsRevision } else { $null }
    fabric_vcs_modified    = if ($EvalMode -eq 'Native') { $fabricProv.VcsModified } else { $null }
    baseline_sha256        = if ($EvalMode -eq 'PR5Matched') { $baselineSha } else { $null }
    baseline_go_version_m  = if ($EvalMode -eq 'PR5Matched') { $baselineProv.VersionM } else { $null }
    baseline_vcs_revision  = if ($EvalMode -eq 'PR5Matched') { $baselineProv.VcsRevision } else { $null }
    baseline_vcs_modified  = if ($EvalMode -eq 'PR5Matched') { $baselineProv.VcsModified } else { $null }
    baseline_script        = if ($EvalMode -eq 'PR5Matched') { $pr5Script } else { $null }
    baseline_script_sha256 = if ($EvalMode -eq 'PR5Matched') { $pr5ScriptSha } else { $null }
    candidate_copy_exe     = $CandidateCopyExe
    candidate_copy_sha256  = $copySha
    candidate_copy_go_version_m = $copyProv.VersionM
    candidate_copy_vcs_revision = $copyProv.VcsRevision
    candidate_copy_vcs_modified = $copyProv.VcsModified
    runtime_sha256         = $runtimeSha
    runner_sha256          = $runnerSha
    runner_source_hashes   = $runnerSourceHashes
    product_head_at_evaluate = $evalProductHead
    product_tree_dirty_at_evaluate = $evalProductDirty
    product_head_note      = 'Runner HEAD/dirty is the evaluation-time runner tree (dirty due to untracked helper source); binary vcs.revision/vcs.modified above is the binary build claim, not the runner HEAD. Do not claim clean binary source from a dirty runner HEAD.'
    go_version             = $goVersion
    prior_run_json_sha256  = $priorRunJsonSha
    task_pins              = $taskPins
    build_receipt          = $buildReceipt
    created_utc            = [DateTime]::UtcNow.ToString('o')
    native_verification    = $manifest.native_verification
    raw_transcripts_retained = $false
    credentials_retained   = $false
    results                = $results
}
$evalRecord | ConvertTo-Json -Depth 10 | Set-Content -Encoding utf8 (Join-Path $evalRoot 'eval.json')
Write-Output "Evaluate complete ($EvalMode): $(@($results | Where-Object { $_.terminal_state -eq 'PASS' }).Count)/$($results.Count) PASS. See $evalRoot\eval.json"
