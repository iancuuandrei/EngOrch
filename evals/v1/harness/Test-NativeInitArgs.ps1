$ErrorActionPreference = 'Stop'
# Extract only the pure builder; never initialize a checkout or call a provider.
$runner = Join-Path $PSScriptRoot '..\..\..\scripts\evaluate-v1.ps1'
$parseTokens = $null
$parseErrors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile($runner, [ref]$parseTokens, [ref]$parseErrors)
if ($parseErrors.Count -ne 0) { throw 'Evaluation runner does not parse.' }
$builder = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-NativeInitArgs' }, $true)
if ($null -eq $builder) { throw 'Native init argument builder missing.' }
$fixerBinding = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-FixerAccessRunnerBinding' }, $true)
if ($null -eq $fixerBinding) { throw 'Fixer/access init binding helper missing.' }
$modelPolicyBinding = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-ModelPolicyRunnerBinding' }, $true)
if ($null -eq $modelPolicyBinding) { throw 'Model policy init binding helper missing.' }
. ([scriptblock]::Create($fixerBinding.Extent.Text))
. ([scriptblock]::Create($modelPolicyBinding.Extent.Text))
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
$strict = @(Get-NativeInitArgs $taskPath $runtimePath $model $effort $false '' '' '' $true)
$expectedStrict = $expected + '--strict-writer-edits'
if (($strict | ConvertTo-Json -Compress) -ne ($expectedStrict | ConvertTo-Json -Compress)) { throw 'Strict writer init invocation differs.' }
$fixerModel = 'gpt-6.1-sol'
$accessPath = 'D:\access policy\public access.json'
$fixer = @(Get-NativeInitArgs $taskPath $runtimePath $model $effort $false $fixerModel 'high' $accessPath)
$expectedFixer = $expected + @('--fixer-model', $fixerModel, '--fixer-effort', 'high', '--access-config', $accessPath)
if (($fixer | ConvertTo-Json -Compress) -ne ($expectedFixer | ConvertTo-Json -Compress)) { throw 'Explicit fixer/access init invocation differs.' }
$rejectedMissingAccess = $false
try { Get-NativeInitArgs $taskPath $runtimePath $model $effort $false $fixerModel 'high' '' | Out-Null } catch { $rejectedMissingAccess = $true }
if (-not $rejectedMissingAccess) { throw 'Fixer init was accepted without explicit access config.' }
$rejectedConflictingWriterContracts = $false
try { Get-NativeInitArgs $taskPath $runtimePath $model $effort $true '' '' '' $true | Out-Null } catch { $rejectedConflictingWriterContracts = $true }
if (-not $rejectedConflictingWriterContracts) { throw 'Conflicting writer-contract init flags were accepted.' }
Write-Output 'PASS: default init argv preserved; validation, strict-writer, and explicit fixer/access argv exact; invalid combinations rejected; paths remain single arguments. No provider calls.'
