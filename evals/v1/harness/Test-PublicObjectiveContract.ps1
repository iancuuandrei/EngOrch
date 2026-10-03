$ErrorActionPreference = 'Stop'
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..\..')).Path
$manifest = Get-Content -LiteralPath (Join-Path $repoRoot 'evals/v1/manifest.json') -Raw | ConvertFrom-Json
$entry = @($manifest.repositories | Where-Object id -eq 'go-difflib')
if ($entry.Count -ne 1) { throw 'Exactly one difflib task is required.' }
$taskDoc = Get-Content -LiteralPath (Join-Path $repoRoot 'evals/v1/tasks/go-difflib.md') -Raw
$publicContract = 'Nonempty text without a trailing newline must retain the current convention of returning its last line with a newline; existing newlines must be preserved.'
if (-not $taskDoc.Contains($publicContract) -or -not $entry[0].task.Contains($publicContract)) {
    throw 'Planner objective must retain the pre-existing public final-line convention.'
}
if ($manifest.suite_id -ne 'fabric-v1-real-repository-2026-10-03-objective-v2') {
    throw 'The clarified objective must not reuse the old frozen suite identity.'
}
if ($entry[0].sha -ne '5d4384ee4fb2527b0a1256a821ebfc92f91efefc' -or $entry[0].check -ne 'difflib') {
    throw 'Objective clarification must not change the repository pin or acceptance selector.'
}
$expectedArgv = 'go|test|-vet=off|-count=1|difflib/difflib.go|difflib/difflib_test.go'
if (($entry[0].native_argv -join '|') -ne $expectedArgv) {
    throw 'Objective clarification must not narrow native verification.'
}
Write-Output 'PASS: objective-v2 names the existing public contract; source pin, native argv and acceptance selector are unchanged. Historical frozen outcomes are not reclassified.'
