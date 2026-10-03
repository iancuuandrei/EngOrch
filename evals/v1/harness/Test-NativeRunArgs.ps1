$ErrorActionPreference = 'Stop'
# Extract the pure argument builder without executing Prepare or Evaluate.
$runner = Join-Path $PSScriptRoot '..\..\..\scripts\evaluate-v1.ps1'
$parseTokens = $null
$parseErrors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile($runner, [ref]$parseTokens, [ref]$parseErrors)
if ($parseErrors.Count -ne 0) { throw 'Evaluation runner does not parse.' }
$builder = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-NativeRunArgs' }, $true)
if ($null -eq $builder) { throw 'Native argument builder missing.' }
$contextValidator = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-PlannerContextBindingShape' }, $true)
if ($null -eq $contextValidator) { throw 'Planner context binding validator missing.' }
. ([scriptblock]::Create($contextValidator.Extent.Text))
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
$recipe = @(Get-NativeRunArgs $taskPath $objective $false 0 '' 'cache-prefix-v1')
$expectedRecipe = @('--root', $taskPath, 'run', '--autonomous', '--prompt-recipe', 'cache-prefix-v1', $objective)
if (($recipe | ConvertTo-Json -Compress) -ne ($expectedRecipe | ConvertTo-Json -Compress)) { throw 'Prompt recipe argument differs.' }
$legacyPlanner = @(Get-NativeRunArgs $taskPath $objective $false 0 'source-bounded-v1' '')
$expectedLegacyPlanner = @('--root', $taskPath, 'run', '--autonomous', '--planner-context', 'source-bounded-v1', $objective)
if (($legacyPlanner | ConvertTo-Json -Compress) -ne ($expectedLegacyPlanner | ConvertTo-Json -Compress)) { throw 'Existing source-bounded planner argv changed.' }
$parserPath = [IO.Path]::GetFullPath((Join-Path $env:TEMP 'pinned-ri.exe'))
$parserHash = 'a' * 64
$goSource = @(Get-NativeRunArgs $taskPath $objective $false 0 'go-source-context-v1' '' $parserPath $parserHash)
$expectedGoSource = @('--root', $taskPath, 'run', '--autonomous', '--planner-context', 'go-source-context-v1', '--planner-context-ri-executable', $parserPath, '--planner-context-ri-executable-sha256', $parserHash, $objective)
if (($goSource | ConvertTo-Json -Compress) -ne ($expectedGoSource | ConvertTo-Json -Compress)) { throw 'go-source-context-v1 argv does not preserve the exact explicit parser binding.' }
$invalidBindings = @(
    @{ Mode='go-source-context-v1'; Path=''; Hash='' },
    @{ Mode='go-source-context-v1'; Path=$parserPath; Hash='' },
    @{ Mode='go-source-context-v1'; Path='relative\ri.exe'; Hash=$parserHash },
    @{ Mode='go-source-context-v1'; Path=([IO.Path]::GetDirectoryName($parserPath) + [IO.Path]::DirectorySeparatorChar + '.' + [IO.Path]::DirectorySeparatorChar + [IO.Path]::GetFileName($parserPath)); Hash=$parserHash },
    @{ Mode='go-source-context-v1'; Path=$parserPath; Hash=$parserHash.ToUpperInvariant() },
    @{ Mode='source-bounded-v1'; Path=$parserPath; Hash=$parserHash }
)
foreach ($binding in $invalidBindings) {
    $rejected = $false
    try { Assert-PlannerContextBindingShape $binding.Mode $binding.Path $binding.Hash } catch { $rejected = $true }
    if (-not $rejected) { throw "Invalid planner parser binding was accepted: $($binding.Mode) $($binding.Path)" }
}
$rejectedRecipe = $false
try { Get-NativeRunArgs $taskPath $objective $false 0 '' 'unknown' | Out-Null } catch { $rejectedRecipe = $true }
if (-not $rejectedRecipe) { throw 'Unknown prompt recipe admitted.' }
foreach ($invalid in @(-1, 9)) {
    $rejected = $false
    try { Get-NativeRunArgs $taskPath $objective $true $invalid | Out-Null } catch { $rejected = $true }
    if (-not $rejected) { throw "Invalid limit $invalid was admitted." }
}
Write-Output 'PASS: legacy argv is byte-order stable; planner treatments bind exact parser provenance; invalid bindings reject; objectives stay one argument; no provider calls.'
