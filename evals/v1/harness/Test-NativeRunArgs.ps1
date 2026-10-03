$ErrorActionPreference = 'Stop'
# Extract the pure argument builder without executing Prepare or Evaluate.
$runner = Join-Path $PSScriptRoot '..\..\..\scripts\evaluate-v1.ps1'
$parseTokens = $null
$parseErrors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile($runner, [ref]$parseTokens, [ref]$parseErrors)
if ($parseErrors.Count -ne 0) { throw 'Evaluation runner does not parse.' }
$builder = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-NativeRunArgs' }, $true)
if ($null -eq $builder) { throw 'Native argument builder missing.' }
. ([scriptblock]::Create($builder.Extent.Text))

$taskPath = 'D:\task path\repo'
$objective = 'Implement a real task with --literal text'
$expectedLegacy = @('--root', $taskPath, 'run', '--autonomous', $objective)
$actualLegacy = @(Get-NativeRunArgs $taskPath $objective $false 0)
if (($actualLegacy | ConvertTo-Json -Compress) -ne ($expectedLegacy | ConvertTo-Json -Compress)) { throw 'Default native invocation changed.' }
foreach ($limit in @(1, 2, 8)) {
    $expected = @('--root', $taskPath, 'run', '--autonomous', '--parallel-writers', '--max-parallel', [string]$limit, $objective)
    $actual = @(Get-NativeRunArgs $taskPath $objective $true $limit)
    if (($actual | ConvertTo-Json -Compress) -ne ($expected | ConvertTo-Json -Compress)) { throw "Parallel invocation differs for limit $limit." }
}
$serial = @(Get-NativeRunArgs $taskPath $objective $false 1)
if ($serial -contains '--parallel-writers' -or $serial[-1] -ne $objective -or $serial[-2] -ne '1') { throw 'Serial override unexpectedly enables parallel writers or splits the objective.' }
foreach ($invalid in @(-1, 9)) {
    $rejected = $false
    try { Get-NativeRunArgs $taskPath $objective $true $invalid | Out-Null } catch { $rejected = $true }
    if (-not $rejected) { throw "Invalid limit $invalid was admitted." }
}
Write-Output 'PASS: default argv preserved; parallel/serial overrides exact; objectives stay one argument; invalid limits reject. No provider calls.'
