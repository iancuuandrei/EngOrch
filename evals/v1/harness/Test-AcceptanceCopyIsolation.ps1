param([string]$Runner = (Join-Path $PSScriptRoot '..\..\..\scripts\evaluate-v1.ps1'))

$source = [System.IO.File]::ReadAllText([System.IO.Path]::GetFullPath($Runner))
$tokens = $null
$parseErrors = $null
[System.Management.Automation.Language.Parser]::ParseInput($source, [ref]$tokens, [ref]$parseErrors) | Out-Null
if ($parseErrors.Count) { throw "Runner parse failed: $($parseErrors[0].Message)" }

# Pin the acceptance flow: candidate-controlled native tests get a disposable
# copy, then held-out tests get a fresh copy of the same candidate binding.
$checks = @(
    'function Invoke-CandidateCopy',
    '$result.acceptance_copy_path = $copyPath',
    '$result.native_copy_path = $copyPath',
    '$nativeVerify = Invoke-GoTest $copyPath $native.Argv',
    '$heldoutCopyObs = Invoke-CandidateCopy $CandidateCopyExe $copySnapshotPath $expectedCandidateId $heldoutCopyStage $taskOutDir ''heldout-candidatecopy''',
    '$writtenHeldoutSources = Write-HeldoutSources $heldoutCopyPath $heldoutSources',
    '$heldout = Invoke-GoTest $heldoutCopyPath (Get-HeldoutGoArgs $entry $heldoutFileArg)'
)
$positions = foreach ($needle in $checks) {
    $i = $source.IndexOf($needle, [StringComparison]::Ordinal)
    if ($i -lt 0) { throw "Required acceptance-isolation contract missing: $needle" }
    $i
}
for ($i = 1; $i -lt $positions.Count; $i++) {
    if ($positions[$i] -le $positions[$i - 1]) {
        throw "Acceptance isolation ordering invalid near '$($checks[$i])'"
    }
}

$requiredBindings = @(
    '[string]$heldoutCopyObs.candidate_id -ne $expectedCandidateId',
    '[string]$heldoutCopyObs.files_hash -ne [string]$copyObs.files_hash',
    '[string]$heldoutCopyObs.workspace -ne [string]$copyObs.workspace',
    '[int]$heldoutCopyObs.file_count -ne [int]$copyObs.file_count',
    'heldout-acceptance-copy.manifest.json'
)
foreach ($binding in $requiredBindings) {
    if (-not $source.Contains($binding)) { throw "Required held-out copy evidence/binding missing: $binding" }
}

'PASS: runner parses; native tests use the first copy; held-out source/tests use a later fresh copy; copy bindings and per-gate manifests are present.'
