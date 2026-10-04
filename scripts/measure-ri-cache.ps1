param(
    [Parameter(Mandatory = $true)][string]$RiBinary,
    [Parameter(Mandatory = $true)][string]$HumanizeRepository,
    [Parameter(Mandatory = $true)][string]$GoExecutable,
    [Parameter(Mandatory = $true)][string]$OutputPath,
    [ValidateRange(1, 10)][int]$Count = 5,
    [ValidateSet('Corpus', 'Contract')][string]$Mode = 'Corpus'
)

$ErrorActionPreference = 'Stop'
$repoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$riPath = (Resolve-Path -LiteralPath $RiBinary).Path
$goPath = (Resolve-Path -LiteralPath $GoExecutable).Path
$humanizePath = (Resolve-Path -LiteralPath $HumanizeRepository).Path
$outputPath = [IO.Path]::GetFullPath($OutputPath)
$outputParent = Split-Path -Parent $outputPath
if (-not (Test-Path -LiteralPath $outputParent -PathType Container)) { throw 'Output parent directory must already exist.' }

function Get-BenchmarkSourceFingerprint([string]$root) {
    $files = @('go.mod', 'go.sum')
    foreach ($directory in @('internal/ri', 'internal/repository', 'internal/canonical', 'internal/safepath', 'internal/taskcontext', 'internal/gitexec')) {
        $files += Get-ChildItem -LiteralPath (Join-Path $root $directory) -File -Recurse | ForEach-Object {
            [IO.Path]::GetRelativePath($root, $_.FullName).Replace('\', '/')
        }
    }
    if ($Mode -eq 'Contract') {
        $files += Get-ChildItem -LiteralPath (Join-Path $root 'internal') -File -Recurse | ForEach-Object {
            [IO.Path]::GetRelativePath($root, $_.FullName).Replace('\', '/')
        }
    }
    $lines = foreach ($relative in ($files | Sort-Object -Unique)) {
        $fullPath = Join-Path $root $relative
        if (-not (Test-Path -LiteralPath $fullPath -PathType Leaf)) { continue }
        $hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $fullPath).Hash.ToLowerInvariant()
        "$relative=$hash"
    }
    $bytes = [Text.Encoding]::UTF8.GetBytes(($lines -join "`n"))
    [pscustomobject]@{
        Digest = [Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($bytes)).ToLowerInvariant()
        FileCount = @($lines).Count
    }
}

foreach ($path in @($riPath, $goPath)) {
    if (-not (Test-Path -LiteralPath $path -PathType Leaf)) {
        throw "Required executable is missing: $path"
    }
}
if (-not (Test-Path -LiteralPath $humanizePath -PathType Container)) {
    throw 'Pinned go-humanize checkout is missing.'
}
$humanizeCommit = (& git -C $humanizePath rev-parse --verify 'HEAD^{commit}').Trim()
if ($LASTEXITCODE -ne 0 -or $humanizeCommit -ne 'a1b4e66b9a6d890e9e15e7091cf16c8032367d6e') {
    throw 'Benchmark checkout does not match the pinned go-humanize commit.'
}
$humanizeTree = (& git -C $humanizePath rev-parse --verify "$humanizeCommit^{tree}").Trim()
if ($LASTEXITCODE -ne 0) { throw 'Could not bind the pinned go-humanize tree.' }
$sourceCommitBefore = (& git -C $repoRoot rev-parse --verify 'HEAD^{commit}').Trim()
if ($LASTEXITCODE -ne 0) { throw 'Could not bind the benchmark source commit.' }
$sourceTreeBefore = (& git -C $repoRoot rev-parse --verify "$sourceCommitBefore^{tree}").Trim()
if ($LASTEXITCODE -ne 0) { throw 'Could not bind the benchmark source tree.' }
$goVersion = (& $goPath version).Trim()
if ($LASTEXITCODE -ne 0) { throw 'Could not read the Go toolchain version.' }
$goHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $goPath).Hash.ToLowerInvariant()
$riHashBefore = (Get-FileHash -Algorithm SHA256 -LiteralPath $riPath).Hash.ToLowerInvariant()
$sourceFingerprintBefore = Get-BenchmarkSourceFingerprint $repoRoot

$oldRi = $env:ENGORCH_RI_BINARY
$oldHumanize = $env:ENGORCH_RI_BENCH_REPOSITORY
$oldHumanizePin = $env:ENGORCH_RI_BENCH_HUMANIZE_COMMIT
$oldRiPin = $env:ENGORCH_RI_BENCH_RI_SHA256
$oldPath = $env:PATH
try {
    $env:ENGORCH_RI_BINARY = $riPath
    $env:ENGORCH_RI_BENCH_REPOSITORY = $humanizePath
    $env:ENGORCH_RI_BENCH_HUMANIZE_COMMIT = $humanizeCommit
    $env:ENGORCH_RI_BENCH_RI_SHA256 = $riHashBefore
    $env:PATH = (Split-Path -Parent $goPath) + [IO.Path]::PathSeparator + $oldPath

    $start = [Diagnostics.ProcessStartInfo]::new()
    $start.FileName = $goPath
    $start.WorkingDirectory = $repoRoot
    $start.UseShellExecute = $false
    $start.RedirectStandardOutput = $true
    $start.RedirectStandardError = $true
    $benchmarkPackage = if ($Mode -eq 'Contract') { './internal/control' } else { './internal/ri' }
    $benchmarkName = if ($Mode -eq 'Contract') { '^BenchmarkPlannerGoContractResource$' } else { '^BenchmarkGoCorpusResource/(synthetic-24x64KiB|pinned-go-humanize)$' }
    foreach ($argument in @(
        'test', $benchmarkPackage, '-run', '^$', '-v',
        '-bench', $benchmarkName,
        '-benchmem', '-benchtime=1x', "-count=$Count"
    )) {
        $start.ArgumentList.Add($argument)
    }
    $process = [Diagnostics.Process]::new()
    $process.StartInfo = $start
    if (-not $process.Start()) { throw 'Go benchmark process failed to start.' }
    $stdoutTask = $process.StandardOutput.ReadToEndAsync()
    $stderrTask = $process.StandardError.ReadToEndAsync()
    $peakWorkingSet = [int64]0
    $peakObservedCpuMs = [double]0
    while (-not $process.HasExited) {
        $entries = @(Get-CimInstance Win32_Process -ErrorAction Stop)
        $ids = [Collections.Generic.HashSet[int]]::new()
        [void]$ids.Add($process.Id)
        do {
            $changed = $false
            foreach ($entry in $entries) {
                if ($ids.Contains([int]$entry.ParentProcessId) -and $ids.Add([int]$entry.ProcessId)) {
                    $changed = $true
                }
            }
        } while ($changed)
        $workingSet = [int64]0
        $cpuMs = [double]0
        foreach ($id in $ids) {
            $child = Get-Process -Id $id -ErrorAction SilentlyContinue
            if ($null -ne $child) {
                $workingSet += [int64]$child.WorkingSet64
                if ($null -ne $child.CPU) { $cpuMs += [double]$child.CPU * 1000 }
            }
        }
        if ($workingSet -gt $peakWorkingSet) { $peakWorkingSet = $workingSet }
        if ($cpuMs -gt $peakObservedCpuMs) { $peakObservedCpuMs = $cpuMs }
        Start-Sleep -Milliseconds 100
    }
    $process.WaitForExit()
    $stdout = $stdoutTask.GetAwaiter().GetResult()
    $stderr = $stderrTask.GetAwaiter().GetResult()
    if ($process.ExitCode -ne 0) {
        [IO.File]::WriteAllText("$outputPath.failure.log", "$stderr`n$stdout", [Text.UTF8Encoding]::new($false))
        throw "Go benchmark failed with exit $($process.ExitCode); diagnostics retained in $outputPath.failure.log"
    }
    $sourceCommitAfter = (& git -C $repoRoot rev-parse --verify 'HEAD^{commit}').Trim()
    if ($LASTEXITCODE -ne 0 -or $sourceCommitAfter -ne $sourceCommitBefore) {
        throw 'Benchmark source commit changed while the measurement was running.'
    }
    $sourceFingerprintAfter = Get-BenchmarkSourceFingerprint $repoRoot
    if ($sourceFingerprintBefore.Digest -ne $sourceFingerprintAfter.Digest) {
        throw 'Benchmark source changed while the measurement was running.'
    }
    $riHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $riPath).Hash.ToLowerInvariant()
    if ($riHash -cne $riHashBefore -or (Get-FileHash -Algorithm SHA256 -LiteralPath $goPath).Hash.ToLowerInvariant() -cne $goHash) {
        throw 'Benchmark executable bytes changed while the measurement was running.'
    }
    $humanizeCommitAfter = (& git -C $humanizePath rev-parse --verify 'HEAD^{commit}').Trim()
    if ($LASTEXITCODE -ne 0 -or $humanizeCommitAfter -cne $humanizeCommit) {
        throw 'Benchmark repository HEAD changed while the measurement was running.'
    }
    $result = [ordered]@{
        schema = 'engorch.ri.parse-cache-resource-measurement.v1'
        timestamp_utc = [DateTime]::UtcNow.ToString('o')
        source_commit = $sourceCommitBefore
        source_tree = $sourceTreeBefore
        source_fingerprint_sha256 = $sourceFingerprintBefore.Digest
        source_file_count = $sourceFingerprintBefore.FileCount
        go_version = $goVersion
        go_executable_sha256 = $goHash
        ri_executable_sha256 = $riHash
        humanize_commit = $humanizeCommit
        humanize_tree = $humanizeTree
        count = $Count
        command = "go test $benchmarkPackage -run ^$ -v -bench $benchmarkName -benchmem -benchtime=1x"
        measurement_mode = $Mode
        peak_sampled_process_tree_working_set_bytes = $peakWorkingSet
        max_sampled_process_tree_cpu_ms = [math]::Round($peakObservedCpuMs, 1)
        benchmark_output = $stdout.Trim()
        benchmark_stderr = $stderr.Trim()
        limitation = 'Sampling covers the whole go test process tree, including fixture setup, build activity, cache priming, and RI subprocesses; it is not per-case attribution. The 100 ms sample interval can miss brief peaks and may perturb runtime.'
    }
    [IO.File]::WriteAllText($outputPath, ($result | ConvertTo-Json -Depth 8), [Text.UTF8Encoding]::new($false))
    Write-Output $outputPath
} finally {
    $env:ENGORCH_RI_BINARY = $oldRi
    $env:ENGORCH_RI_BENCH_REPOSITORY = $oldHumanize
    $env:ENGORCH_RI_BENCH_HUMANIZE_COMMIT = $oldHumanizePin
    $env:ENGORCH_RI_BENCH_RI_SHA256 = $oldRiPin
    $env:PATH = $oldPath
}
