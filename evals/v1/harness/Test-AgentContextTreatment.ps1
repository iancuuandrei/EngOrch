param([string]$Runner = (Join-Path $PSScriptRoot '..\..\..\scripts\evaluate-v1.ps1'))
$ErrorActionPreference = 'Stop'
$tokens = $null; $errors = $null
$ast = [Management.Automation.Language.Parser]::ParseFile($Runner, [ref]$tokens, [ref]$errors)
if ($errors.Count -ne 0) { throw 'Evaluation runner has syntax errors.' }
foreach ($name in @('Assert-AgentContextRunnerBinding', 'Assert-AgentContextPreparedBinding', 'Assert-AgentContextObserved', 'Get-NativeRunArgs', 'Test-GoSourceContextMode', 'Assert-PlannerContextBindingShape', 'Assert-ReviewImpactRunnerBindingShape', 'Assert-CandidateFactsCacheRunnerBindingShape')) {
    $definition = $ast.Find({ param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq $name }, $true)
    if ($null -eq $definition) { throw "Missing evaluation helper: $name" }
    . ([scriptblock]::Create($definition.Extent.Text))
}
function Expect-Rejection([scriptblock]$Action) {
    $rejected = $false
    try { & $Action } catch { $rejected = $true }
    if (-not $rejected) { throw 'Invalid agent-context treatment was accepted.' }
}
Assert-AgentContextRunnerBinding 'Native' 'Enabled'
Assert-AgentContextRunnerBinding 'PR5Matched' 'Default'
Expect-Rejection { Assert-AgentContextRunnerBinding 'PR5Matched' 'Disabled' }
Expect-Rejection { Assert-AgentContextRunnerBinding 'Native' 'invalid' }
Assert-AgentContextPreparedBinding ([pscustomobject]@{}) 'Default'
Assert-AgentContextPreparedBinding ([pscustomobject]@{agent_context_requested='Disabled'}) 'Disabled'
Expect-Rejection { Assert-AgentContextPreparedBinding ([pscustomobject]@{}) 'Enabled' }
Expect-Rejection { Assert-AgentContextPreparedBinding ([pscustomobject]@{agent_context_requested='Enabled'}) 'Disabled' }
$plain = @(Get-NativeRunArgs 'D:\task' 'objective' $false 0)
if (($plain -join '|') -cne '--root|D:\task|run|--autonomous|objective') { throw 'Default invocation changed.' }
foreach ($treatment in @('Disabled', 'Enabled')) {
    $argv = @(Get-NativeRunArgs 'D:\task' 'objective' $false 0 -AgentContext $treatment)
    $expected = '--agent-context=' + ($treatment -eq 'Enabled').ToString().ToLowerInvariant()
    if ($argv.Count -ne 6 -or $argv[4] -cne $expected -or $argv[5] -cne 'objective') { throw 'Explicit treatment argv is not bound.' }
}
Expect-Rejection { Get-NativeRunArgs 'D:\task' 'objective' $false 0 -AgentContext 'invalid' }
$commit = 'a' * 40
$bundle = [pscustomobject]@{version=1;source_id=('b'*64);source_commit=$commit;instructions=@();skills=@()}
$snapshot = [pscustomobject]@{creation=[pscustomobject]@{repository=[pscustomobject]@{commit=$commit};agent_context=$bundle}}
$observed = Assert-AgentContextObserved $snapshot 'Enabled'
if (-not $observed.present -or $observed.source_commit -cne $commit -or $observed.instructions -ne 0 -or $observed.skills -ne 0) { throw 'Empty but valid bundle was not observed.' }
Expect-Rejection { Assert-AgentContextObserved $snapshot 'Disabled' }
$bundle.instructions = $null; $bundle.skills = $null
$observed = Assert-AgentContextObserved $snapshot 'Enabled'
if ($observed.instructions -ne 0 -or $observed.skills -ne 0) { throw 'Null inventory was counted as one document.' }
$bundle.version = 2
Expect-Rejection { Assert-AgentContextObserved $snapshot 'Enabled' }
$bundle.version = 1
$bundle.source_commit = 'c' * 40
Expect-Rejection { Assert-AgentContextObserved $snapshot 'Enabled' }
$bundle.source_commit = $commit
$bundle.source_id = 'not-a-source-id'
Expect-Rejection { Assert-AgentContextObserved $snapshot 'Enabled' }
$legacy = [pscustomobject]@{creation=[pscustomobject]@{repository=[pscustomobject]@{commit=$commit}}}
Expect-Rejection { Assert-AgentContextObserved $legacy 'Enabled' }
if ((Assert-AgentContextObserved $legacy 'Disabled').present) { throw 'Disabled treatment reported a bundle.' }
'PASS: explicit agent-context argv, preparation matching, observed source binding and legacy omission'
