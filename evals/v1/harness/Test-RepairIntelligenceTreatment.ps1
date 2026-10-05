param([string]$Runner = (Join-Path $PSScriptRoot '..\..\..\scripts\evaluate-v1.ps1'))
$ErrorActionPreference = 'Stop'
$tokens=$null; $errors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile($Runner,[ref]$tokens,[ref]$errors)
if ($errors.Count -ne 0) { throw 'Evaluation runner has syntax errors.' }
foreach ($name in @('Assert-RepairIntelligenceRunnerBinding','Assert-RepairIntelligencePreparedBinding','Assert-RepairIntelligenceObserved','Get-NativeRunArgs','Test-GoSourceContextMode','Assert-PlannerContextBindingShape','Assert-ReviewImpactRunnerBindingShape','Assert-CandidateFactsCacheRunnerBindingShape')) {
    $definition=$ast.Find({param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq $name},$true)
    if ($null -eq $definition) { throw "Missing helper: $name" }
    . ([scriptblock]::Create($definition.Extent.Text))
}
function Expect-Rejection([scriptblock]$Action) {
    $rejected=$false
    try { & $Action } catch { $rejected=$true }
    if (-not $rejected) { throw 'Invalid repair treatment accepted.' }
}
Assert-RepairIntelligenceRunnerBinding 'Native' $true
Assert-RepairIntelligenceRunnerBinding 'PR5Matched' $false
Expect-Rejection { Assert-RepairIntelligenceRunnerBinding 'PR5Matched' $true }
Assert-RepairIntelligencePreparedBinding ([pscustomobject]@{}) $false
Assert-RepairIntelligencePreparedBinding ([pscustomobject]@{repair_intelligence_version_requested=1}) $true
foreach ($version in @(0,2,'1',$true,$null,1.0)) {
    Expect-Rejection { Assert-RepairIntelligencePreparedBinding ([pscustomobject]@{repair_intelligence_version_requested=$version}) $true }
}
Expect-Rejection { Assert-RepairIntelligencePreparedBinding ([pscustomobject]@{}) $true }
Expect-Rejection { Assert-RepairIntelligencePreparedBinding ([pscustomobject]@{repair_intelligence_version_requested=1}) $false }
$plain=@(Get-NativeRunArgs 'D:\task' 'objective' $false 0)
if (($plain -join '|') -cne '--root|D:\task|run|--autonomous|objective') { throw 'Default invocation changed.' }
$treatment=@(Get-NativeRunArgs 'D:\task' 'objective' $false 1 -EnableRepairIntelligence $true)
if (($treatment -join '|') -cne '--root|D:\task|run|--autonomous|--repair-intelligence|--max-parallel|1|objective') { throw 'Treatment flag omitted or duplicated.' }
$execution=[pscustomobject]@{repair_intelligence_version=1;graph_version=1;repair_planning_version=1;context='bounded-v1'}
$snapshot=[pscustomobject]@{creation=[pscustomobject]@{execution=$execution}}
if ((Assert-RepairIntelligenceObserved $snapshot $true) -ne 1) { throw 'Observed version missing.' }
Expect-Rejection { Assert-RepairIntelligenceObserved $snapshot $false }
foreach ($version in @(0,2,'1',$true)) {
    $execution.repair_intelligence_version=$version
    Expect-Rejection { Assert-RepairIntelligenceObserved $snapshot $true }
}
$execution.repair_intelligence_version=1
$snapshot.creation | Add-Member config ([pscustomobject]@{reviewer_contract='json-v1';reviewer=[pscustomobject]@{}})
$execution | Add-Member review_recheck_version 1
foreach ($version in @(0,2,'1',$true,$null,1.0)) {
    $execution.review_recheck_version=$version
    Expect-Rejection { Assert-RepairIntelligenceObserved $snapshot $true }
}
$execution.review_recheck_version=1
if ((Assert-RepairIntelligenceObserved $snapshot $true) -ne 1) { throw 'Valid bundled reviewer policy rejected.' }
foreach ($field in @('graph_version','repair_planning_version')) {
    foreach ($version in @(0,2,'1',$true,$null,1.0)) {
        $execution.$field=$version
        Expect-Rejection { Assert-RepairIntelligenceObserved $snapshot $true }
    }
    $execution.$field=1
}
$execution.context=''
Expect-Rejection { Assert-RepairIntelligenceObserved $snapshot $true }
$execution.context=@('bounded-v1')
Expect-Rejection { Assert-RepairIntelligenceObserved $snapshot $true }
$execution.context=$null
Expect-Rejection { Assert-RepairIntelligenceObserved $snapshot $true }
$legacy=[pscustomobject]@{creation=[pscustomobject]@{execution=[pscustomobject]@{}}}
if ((Assert-RepairIntelligenceObserved $legacy $false) -ne 0) { throw 'Legacy policy changed.' }
Expect-Rejection { Assert-RepairIntelligenceObserved $legacy $true }
'PASS: frozen repair-intelligence treatment argv, preparation and observed policy'
