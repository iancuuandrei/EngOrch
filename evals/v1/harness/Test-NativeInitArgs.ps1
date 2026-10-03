$ErrorActionPreference = 'Stop'
# Extract only the pure builder; never initialize a checkout or call a provider.
$runner = Join-Path $PSScriptRoot '..\..\..\scripts\evaluate-v1.ps1'
$parseTokens = $null
$parseErrors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile($runner, [ref]$parseTokens, [ref]$parseErrors)
if ($parseErrors.Count -ne 0) { throw 'Evaluation runner does not parse.' }
$builder = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-NativeInitArgs' }, $true)
if ($null -eq $builder) { throw 'Native init argument builder missing.' }
. ([scriptblock]::Create($builder.Extent.Text))
$taskPath = 'D:\task path\repo'
$runtimePath = 'D:\runtime path\codex.exe'
$model = 'gpt-6-luna'
$effort = 'high'
$expected = @('--root', $taskPath, 'init', '--codex', $runtimePath, '--model', $model, '--effort', $effort)
$actual = @(Get-NativeInitArgs $taskPath $runtimePath $model $effort $false)
if (($actual | ConvertTo-Json -Compress) -ne ($expected | ConvertTo-Json -Compress)) { throw 'Default native init invocation changed.' }
$validated = @(Get-NativeInitArgs $taskPath $runtimePath $model $effort $true)
$expectedValidated = $expected + '--validate-writer-edits'
if (($validated | ConvertTo-Json -Compress) -ne ($expectedValidated | ConvertTo-Json -Compress)) { throw 'Validated writer init invocation differs.' }
Write-Output 'PASS: default init argv preserved; validation opt-in exact; paths remain single arguments. No provider calls.'
