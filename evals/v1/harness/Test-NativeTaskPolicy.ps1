$ErrorActionPreference = 'Stop'
$runner = Join-Path $PSScriptRoot '..\..\..\scripts\evaluate-v1.ps1'
$parseTokens = $null
$parseErrors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile($runner, [ref]$parseTokens, [ref]$parseErrors)
if ($parseErrors.Count -ne 0) { throw 'Evaluation runner does not parse.' }
$builder = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-NativeGoArgs' }, $true)
if ($null -eq $builder) { throw 'Native task policy builder missing.' }
. ([scriptblock]::Create($builder.Extent.Text))
foreach ($taskId in @('go-atomic', 'go-atomic-numeric-text', 'godotenv')) {
    $entry = [pscustomobject]@{id=$taskId;native_argv=@('go','test','-count=1','.')}
    $actual = Get-NativeGoArgs $entry
    $shouldSkip = $env:OS -eq 'Windows_NT' -and $taskId -in @('go-atomic','go-atomic-numeric-text')
    $expected = @($entry.native_argv)
    if ($shouldSkip) { $expected += @('-skip','^TestNocmpIntegration$') }
    if (($actual.Argv | ConvertTo-Json -Compress) -ne ($expected | ConvertTo-Json -Compress)) { throw "Native argv differs for $taskId." }
    $expectedScope = if ($shouldSkip) { 'windows-scoped' } else { 'full' }
    if ($actual.Scope -ne $expectedScope) { throw "Native scope differs for $taskId." }
}
Write-Output 'PASS: atomic task variants share the narrow Windows native exclusion; other tasks retain the full policy. No provider calls.'
