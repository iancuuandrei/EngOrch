$ErrorActionPreference = 'Stop'
# Extract the pure argument builder without executing Prepare or Evaluate.
$runner = Join-Path $PSScriptRoot '..\..\..\scripts\evaluate-v1.ps1'
$parseTokens = $null
$parseErrors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile($runner, [ref]$parseTokens, [ref]$parseErrors)
if ($parseErrors.Count -ne 0) { throw 'Evaluation runner does not parse.' }
$builder = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-NativeRunArgs' }, $true)
if ($null -eq $builder) { throw 'Native argument builder missing.' }
$autoCompactBinding = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-AutoCompactRunnerBinding' }, $true)
$autoCompactPrepared = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-AutoCompactPreparedBinding' }, $true)
$autoCompactObserved = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-AutoCompactObserved' }, $true)
if ($null -eq $autoCompactBinding -or $null -eq $autoCompactPrepared -or $null -eq $autoCompactObserved) { throw 'Automatic compaction binding helpers missing.' }
$runnerOptions = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-IsolatedRunnerOptionShape' }, $true)
if ($null -eq $runnerOptions) { throw 'Isolated runner option validator missing.' }
$goMode = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Test-GoSourceContextMode' }, $true)
if ($null -eq $goMode) { throw 'Go source context mode predicate missing.' }
$fileHash = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-FileSha256' }, $true)
$policyBinding = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-IsolationPolicyBinding' }, $true)
$preparedBinding = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-IsolationPolicyPreparedBinding' }, $true)
$observedPolicy = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-IsolationPolicyObserved' }, $true)
$currentBinding = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-CurrentIsolationPolicyBinding' }, $true)
$taskConfig = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Set-TaskVerificationConfig' }, $true)
$nativeGoArgs = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-NativeGoArgs' }, $true)
$fileHash = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-FileSha256' }, $true)
if ($null -eq $fileHash -or $null -eq $policyBinding -or $null -eq $preparedBinding -or $null -eq $observedPolicy -or $null -eq $currentBinding) { throw 'Isolation policy binding helpers missing.' }
if ($null -eq $taskConfig -or $null -eq $nativeGoArgs) { throw 'Task verification configuration helper missing.' }
$contextValidator = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-PlannerContextBindingShape' }, $true)
if ($null -eq $contextValidator) { throw 'Planner context binding validator missing.' }
$reviewImpactBinding = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-ReviewImpactRunnerBindingShape' }, $true)
$reviewImpactPrepared = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-ReviewImpactPreparedBinding' }, $true)
$reviewImpactObserved = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-ReviewImpactObserved' }, $true)
$factsCacheBinding = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-CandidateFactsCacheRunnerBindingShape' }, $true)
$factsCachePrepared = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-CandidateFactsCachePreparedBinding' }, $true)
$factsCacheObserved = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-CandidateFactsCacheObserved' }, $true)
$fixerShape = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-FixerAccessRunnerBinding' }, $true)
$accessBinding = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-AccessConfigBinding' }, $true)
$currentAccessBinding = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-CurrentAccessConfigBinding' }, $true)
$preparedFixerBinding = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-FixerAccessPreparedBinding' }, $true)
$preparedWriterContract = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-WriterContractPreparedBinding' }, $true)
$writerContractRequest = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-WriterContractRequest' }, $true)
$observedWriterContract = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-WriterContractObserved' }, $true)
$observedFixerRoute = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Assert-FixerRouteObserved' }, $true)
$nativeInitArgs = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-NativeInitArgs' }, $true)
if ($null -eq $reviewImpactBinding -or $null -eq $reviewImpactPrepared -or $null -eq $reviewImpactObserved -or
    $null -eq $factsCacheBinding -or $null -eq $factsCachePrepared -or $null -eq $factsCacheObserved) {
    throw 'Review-impact/candidate-facts-cache treatment binding helpers missing.'
}
if ($null -eq $fixerShape -or $null -eq $accessBinding -or $null -eq $currentAccessBinding -or
    $null -eq $preparedFixerBinding -or $null -eq $preparedWriterContract -or $null -eq $writerContractRequest -or
    $null -eq $observedWriterContract -or $null -eq $observedFixerRoute -or $null -eq $nativeInitArgs) {
    throw 'Fixer/access or writer-contract runner binding helpers missing.'
}
. ([scriptblock]::Create($goMode.Extent.Text))
. ([scriptblock]::Create($fileHash.Extent.Text))
. ([scriptblock]::Create($contextValidator.Extent.Text))
. ([scriptblock]::Create($reviewImpactBinding.Extent.Text))
. ([scriptblock]::Create($reviewImpactPrepared.Extent.Text))
. ([scriptblock]::Create($reviewImpactObserved.Extent.Text))
. ([scriptblock]::Create($factsCacheBinding.Extent.Text))
. ([scriptblock]::Create($factsCachePrepared.Extent.Text))
. ([scriptblock]::Create($factsCacheObserved.Extent.Text))
. ([scriptblock]::Create($fixerShape.Extent.Text))
. ([scriptblock]::Create($accessBinding.Extent.Text))
. ([scriptblock]::Create($currentAccessBinding.Extent.Text))
. ([scriptblock]::Create($preparedFixerBinding.Extent.Text))
. ([scriptblock]::Create($preparedWriterContract.Extent.Text))
. ([scriptblock]::Create($writerContractRequest.Extent.Text))
. ([scriptblock]::Create($observedWriterContract.Extent.Text))
. ([scriptblock]::Create($observedFixerRoute.Extent.Text))
. ([scriptblock]::Create($nativeInitArgs.Extent.Text))
. ([scriptblock]::Create($policyBinding.Extent.Text))
. ([scriptblock]::Create($preparedBinding.Extent.Text))
. ([scriptblock]::Create($observedPolicy.Extent.Text))
. ([scriptblock]::Create($currentBinding.Extent.Text))
. ([scriptblock]::Create($taskConfig.Extent.Text))
. ([scriptblock]::Create($nativeGoArgs.Extent.Text))
. ([scriptblock]::Create($runnerOptions.Extent.Text))
. ([scriptblock]::Create($builder.Extent.Text))
. ([scriptblock]::Create($autoCompactBinding.Extent.Text))
. ([scriptblock]::Create($autoCompactPrepared.Extent.Text))
. ([scriptblock]::Create($autoCompactObserved.Extent.Text))
$tomlPolicy = Join-Path $PSScriptRoot 'Toml-ArgvPolicy.ps1'
. $tomlPolicy
if (-not (Test-GoSourceContextMode 'go-source-context-v1') -or
    -not (Test-GoSourceContextMode 'go-source-context-v2') -or
    -not (Test-GoSourceContextMode 'go-contract-context-v1') -or
    -not (Test-GoSourceContextMode 'go-contract-context-v2') -or
    (Test-GoSourceContextMode 'source-bounded-v1') -or
    (Test-GoSourceContextMode '')) {
    throw 'Pinned Go planner context predicate does not distinguish v1, v2, contract v1, and legacy modes.'
}
Assert-IsolatedRunnerOptionShape $false '' $false 'Native' 0
foreach ($invalidOptions in @(
    @{ Enabled=$true; Path=''; Parallel=$false; Mode='Native'; Max=2 },
    @{ Enabled=$false; Path='C:\policy.json'; Parallel=$false; Mode='Native'; Max=2 },
    @{ Enabled=$true; Path='C:\policy.json'; Parallel=$true; Mode='Native'; Max=2 },
    @{ Enabled=$true; Path='C:\policy.json'; Parallel=$false; Mode='PR5Matched'; Max=2 },
    @{ Enabled=$true; Path='C:\policy.json'; Parallel=$false; Mode='Native'; Max=0 }
)) {
    $rejected = $false
    try { Assert-IsolatedRunnerOptionShape $invalidOptions.Enabled $invalidOptions.Path $invalidOptions.Parallel $invalidOptions.Mode $invalidOptions.Max } catch { $rejected = $true }
    if (-not $rejected) { throw 'Invalid isolated runner mode, policy, or scheduler combination was accepted.' }
}

$taskPath = 'D:\task path\repo'
$objective = 'Implement a real task with --literal text'
$legacyInitArgs = @(Get-NativeInitArgs $taskPath 'C:\tools\codex.exe' 'gpt-6-luna' 'high' $false)
$expectedLegacyInitArgs = @('--root', $taskPath, 'init', '--codex', 'C:\tools\codex.exe', '--model', 'gpt-6-luna', '--effort', 'high')
if (($legacyInitArgs | ConvertTo-Json -Compress) -ne ($expectedLegacyInitArgs | ConvertTo-Json -Compress)) { throw 'Default init argv changed.' }
$pilotInitArgs = @(Get-NativeInitArgs $taskPath 'C:\tools\codex.exe' 'gpt-6-luna' 'high' $false 'gpt-6.1-sol' 'high' 'D:\policy\public-access-v1.json' $true)
$expectedPilotInitArgs = @('--root', $taskPath, 'init', '--codex', 'C:\tools\codex.exe', '--model', 'gpt-6-luna', '--effort', 'high', '--fixer-model', 'gpt-6.1-sol', '--fixer-effort', 'high', '--access-config', 'D:\policy\public-access-v1.json', '--strict-writer-edits')
if (($pilotInitArgs | ConvertTo-Json -Compress) -ne ($expectedPilotInitArgs | ConvertTo-Json -Compress)) { throw 'Pilot init argv does not select only the explicit fixer route/access file and strict writer contract.' }
Assert-FixerAccessRunnerBinding 'Native' $true 'gpt-6.1-sol' $true 'high' $true 'D:\policy\access.json'
Assert-FixerAccessRunnerBinding 'Native' $true 'gpt-6.1-sol' $false '' $true 'D:\policy\access.json'
foreach ($badFixerBinding in @(
    @{ Mode='Native'; ModelOn=$true; Model='gpt-6-sol'; EffortOn=$false; Effort=''; AccessOn=$false; Access='' },
    @{ Mode='Native'; ModelOn=$false; Model=''; EffortOn=$false; Effort=''; AccessOn=$true; Access='D:\policy\access.json' },
    @{ Mode='PR5Matched'; ModelOn=$true; Model='gpt-6-sol'; EffortOn=$true; Effort='high'; AccessOn=$true; Access='D:\policy\access.json' },
    @{ Mode='Native'; ModelOn=$true; Model=' '; EffortOn=$false; Effort=''; AccessOn=$true; Access='D:\policy\access.json' }
)) {
    $rejected = $false
    try { Assert-FixerAccessRunnerBinding $badFixerBinding.Mode $badFixerBinding.ModelOn $badFixerBinding.Model $badFixerBinding.EffortOn $badFixerBinding.Effort $badFixerBinding.AccessOn $badFixerBinding.Access } catch { $rejected = $true }
    if (-not $rejected) { throw 'Invalid fixer/access runner combination was accepted.' }
}
if ((Get-WriterContractRequest $false $false 'Native') -cne '' -or
    (Get-WriterContractRequest $true $false 'Native') -cne 'anchored-edits-v2' -or
    (Get-WriterContractRequest $false $true 'Native') -cne 'anchored-edits-v3') { throw 'Writer contract request mapping is incorrect.' }
$rejectedWriterContractConflict = $false
try { Get-WriterContractRequest $true $true 'Native' | Out-Null } catch { $rejectedWriterContractConflict = $true }
if (-not $rejectedWriterContractConflict) { throw 'Conflicting writer contract flags were accepted.' }
$preparedWriterV3 = [pscustomobject]@{ writer_contract_requested = 'anchored-edits-v3' }
Assert-WriterContractPreparedBinding $preparedWriterV3 'anchored-edits-v3'
Assert-WriterContractPreparedBinding ([pscustomobject]@{}) ''
$rejectedWriterContractMutation = $false
try { Assert-WriterContractPreparedBinding $preparedWriterV3 'anchored-edits-v2' } catch { $rejectedWriterContractMutation = $true }
if (-not $rejectedWriterContractMutation) { throw 'Evaluate accepted a changed writer contract.' }
$observedWriterV3 = Assert-WriterContractObserved @{creation=@{config=@{writer_contract='anchored-edits-v3'}}} 'anchored-edits-v3'
if ($observedWriterV3 -cne 'anchored-edits-v3') { throw 'Inspected strict writer contract was not observed.' }
$observedFixer = Assert-FixerRouteObserved @{creation=@{config=@{fixer=@{model='gpt-6.1-sol';effort='high'}}}} $true 'gpt-6.1-sol' 'high'
if ($observedFixer.model -cne 'gpt-6.1-sol' -or $observedFixer.effort -cne 'high') { throw 'Inspected fixer route did not preserve model/effort evidence.' }
$rejectedWrongFixer = $false
try { Assert-FixerRouteObserved @{creation=@{config=@{fixer=@{model='gpt-6-luna';effort='high'}}}} $true 'gpt-6.1-sol' 'high' } catch { $rejectedWrongFixer = $true }
if (-not $rejectedWrongFixer) { throw 'Inspected fixer route mismatch was accepted.' }

$accessTestRoot = Join-Path $env:TEMP ('runner-access-config-' + [Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $accessTestRoot | Out-Null
try {
    $accessPath = Join-Path $accessTestRoot 'public-access-v1.json'
    [IO.File]::WriteAllText($accessPath, '{"class":"PUBLIC"}', (New-Object System.Text.UTF8Encoding($false)))
    $access = Get-AccessConfigBinding $accessPath
    if ($access.Path -cne $accessPath -or $access.Bytes -ne 18 -or $access.Sha256 -notmatch '^[0-9a-f]{64}$') { throw 'Access config path/size/hash binding is incomplete.' }
    Assert-CurrentAccessConfigBinding $access
    $preparedAccess = [pscustomobject]@{
        fixer_model_requested = 'gpt-6.1-sol'; fixer_effort_requested = 'high'
        access_config_path_requested = $access.Path; access_config_sha256_requested = $access.Sha256
        access_config_bytes_requested = $access.Bytes
    }
    Assert-FixerAccessPreparedBinding $preparedAccess $true 'gpt-6.1-sol' $true 'high' $access
    Assert-FixerAccessPreparedBinding ([pscustomobject]@{}) $false '' $false '' $null
    $changedAccess = [pscustomobject]@{
        fixer_model_requested = 'gpt-6.1-sol'; fixer_effort_requested = 'high'
        access_config_path_requested = $access.Path; access_config_sha256_requested = ('0' * 64)
        access_config_bytes_requested = $access.Bytes
    }
    $rejectedAccessMutation = $false
    try { Assert-FixerAccessPreparedBinding $changedAccess $true 'gpt-6.1-sol' $true 'high' $access } catch { $rejectedAccessMutation = $true }
    if (-not $rejectedAccessMutation) { throw 'Evaluate accepted a changed access-config content hash.' }
    $changedAccessPath = [pscustomobject]@{
        fixer_model_requested = 'gpt-6.1-sol'; fixer_effort_requested = 'high'
        access_config_path_requested = (Join-Path $accessTestRoot 'other.json'); access_config_sha256_requested = $access.Sha256
        access_config_bytes_requested = $access.Bytes
    }
    $rejectedAccessPathMutation = $false
    try { Assert-FixerAccessPreparedBinding $changedAccessPath $true 'gpt-6.1-sol' $true 'high' $access } catch { $rejectedAccessPathMutation = $true }
    if (-not $rejectedAccessPathMutation) { throw 'Evaluate accepted an access-config path change with the same recorded bytes/hash.' }
    [IO.File]::WriteAllText($accessPath, '{"class":"PUBLIC","changed":true}', (New-Object System.Text.UTF8Encoding($false)))
    $rejectedCurrentAccessMutation = $false
    try { Assert-CurrentAccessConfigBinding $access } catch { $rejectedCurrentAccessMutation = $true }
    if (-not $rejectedCurrentAccessMutation) { throw 'Changed access-config bytes passed the pre-init binding check.' }
    [IO.File]::WriteAllText($accessPath, ('x' * (32 * 1024 + 1)), (New-Object System.Text.UTF8Encoding($false)))
    $rejectedOversizedAccess = $false
    try { Get-AccessConfigBinding $accessPath | Out-Null } catch { $rejectedOversizedAccess = $true }
    if (-not $rejectedOversizedAccess) { throw 'Oversized access config was accepted.' }
} finally {
    Remove-Item -LiteralPath $accessTestRoot -Recurse -Force -ErrorAction SilentlyContinue
}
Assert-AutoCompactRunnerBinding 'Native' 0 $false
Assert-AutoCompactRunnerBinding 'Native' 64000 $true
foreach ($badAutoCompact in @(
    @{ Mode='Native'; Limit=0; Explicit=$true },
    @{ Mode='Native'; Limit=-1; Explicit=$true },
    @{ Mode='Native'; Limit=10000001; Explicit=$true },
    @{ Mode='PR5Matched'; Limit=64000; Explicit=$true }
)) {
    $rejected = $false
    try { Assert-AutoCompactRunnerBinding $badAutoCompact.Mode $badAutoCompact.Limit $badAutoCompact.Explicit } catch { $rejected = $true }
    if (-not $rejected) { throw 'Invalid or non-Native automatic compaction option was accepted.' }
}
$expectedLegacy = @('--root', $taskPath, 'run', '--autonomous', $objective)
$actualLegacy = @(Get-NativeRunArgs $taskPath $objective $false 0)
if (($actualLegacy | ConvertTo-Json -Compress) -ne ($expectedLegacy | ConvertTo-Json -Compress)) { throw 'Default native invocation changed.' }
$autoCompact = @(Get-NativeRunArgs $taskPath $objective $false 0 '' '' '' '' $false '' 64000)
$expectedAutoCompact = @('--root', $taskPath, 'run', '--autonomous', '--auto-compact-token-limit', '64000', $objective)
if (($autoCompact | ConvertTo-Json -Compress) -ne ($expectedAutoCompact | ConvertTo-Json -Compress)) { throw 'Native automatic compaction argv differs.' }
foreach ($badLimit in @(-1, 10000001)) {
    $rejected = $false
    try { Get-NativeRunArgs $taskPath $objective $false 0 '' '' '' '' $false '' $badLimit | Out-Null } catch { $rejected = $true }
    if (-not $rejected) { throw "Invalid automatic compaction threshold $badLimit was accepted." }
}
$preparedAutoCompact = [pscustomobject]@{ auto_compact_token_limit_requested = 64000 }
Assert-AutoCompactPreparedBinding $preparedAutoCompact 64000
$legacyPrepared = [pscustomobject]@{}
Assert-AutoCompactPreparedBinding $legacyPrepared 0
$rejectedPreparedMismatch = $false
try { Assert-AutoCompactPreparedBinding $preparedAutoCompact 64001 } catch { $rejectedPreparedMismatch = $true }
if (-not $rejectedPreparedMismatch) { throw 'Evaluate accepted a changed prepared automatic compaction threshold.' }
$autoSnapshot = '{"creation":{"execution":{"codex_auto_compact":{"version":1,"token_limit":64000}}}}' | ConvertFrom-Json
if ((Assert-AutoCompactObserved $autoSnapshot 64000).token_limit -ne 64000) { throw 'Inspected Native compaction policy was not returned.' }
$rejectedAutoObservedMismatch = $false
try { Assert-AutoCompactObserved $autoSnapshot 64001 | Out-Null } catch { $rejectedAutoObservedMismatch = $true }
if (-not $rejectedAutoObservedMismatch) { throw 'Mismatched inspected automatic compaction policy was accepted.' }
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
$goSourceV2 = @(Get-NativeRunArgs $taskPath $objective $false 0 'go-source-context-v2' '' $parserPath $parserHash)
$expectedGoSourceV2 = @('--root', $taskPath, 'run', '--autonomous', '--planner-context', 'go-source-context-v2', '--planner-context-ri-executable', $parserPath, '--planner-context-ri-executable-sha256', $parserHash, $objective)
if (($goSourceV2 | ConvertTo-Json -Compress) -ne ($expectedGoSourceV2 | ConvertTo-Json -Compress)) { throw 'go-source-context-v2 argv does not preserve the exact explicit parser binding.' }
$goContractV1 = @(Get-NativeRunArgs $taskPath $objective $false 0 'go-contract-context-v1' '' $parserPath $parserHash)
$expectedGoContractV1 = @('--root', $taskPath, 'run', '--autonomous', '--planner-context', 'go-contract-context-v1', '--planner-context-ri-executable', $parserPath, '--planner-context-ri-executable-sha256', $parserHash, $objective)
if (($goContractV1 | ConvertTo-Json -Compress) -ne ($expectedGoContractV1 | ConvertTo-Json -Compress)) { throw 'go-contract-context-v1 argv does not preserve the exact explicit parser binding.' }
$goContractV2 = @(Get-NativeRunArgs $taskPath $objective $false 0 'go-contract-context-v2' '' $parserPath $parserHash)
$expectedGoContractV2 = @('--root', $taskPath, 'run', '--autonomous', '--planner-context', 'go-contract-context-v2', '--planner-context-ri-executable', $parserPath, '--planner-context-ri-executable-sha256', $parserHash, $objective)
if (($goContractV2 | ConvertTo-Json -Compress) -ne ($expectedGoContractV2 | ConvertTo-Json -Compress)) { throw 'go-contract-context-v2 argv does not preserve the exact explicit parser binding.' }
$reviewImpact = @(Get-NativeRunArgs $taskPath $objective $false 1 'go-contract-context-v1' '' $parserPath $parserHash $false '' 0 $true)
$expectedReviewImpact = @('--root', $taskPath, 'run', '--autonomous', '--planner-context', 'go-contract-context-v1', '--planner-context-ri-executable', $parserPath, '--planner-context-ri-executable-sha256', $parserHash, '--review-impact-context', '--max-parallel', '1', $objective)
if (($reviewImpact | ConvertTo-Json -Compress) -ne ($expectedReviewImpact | ConvertTo-Json -Compress)) { throw 'Review-impact argv does not bind the contract planner and exact pinned parser.' }
$reviewImpactFactsCache = @(Get-NativeRunArgs $taskPath $objective $false 1 'go-contract-context-v1' '' $parserPath $parserHash $false '' 0 $true $true)
$expectedReviewImpactFactsCache = @('--root', $taskPath, 'run', '--autonomous', '--planner-context', 'go-contract-context-v1', '--planner-context-ri-executable', $parserPath, '--planner-context-ri-executable-sha256', $parserHash, '--review-impact-context', '--review-impact-candidate-facts-cache', '--max-parallel', '1', $objective)
if (($reviewImpactFactsCache | ConvertTo-Json -Compress) -ne ($expectedReviewImpactFactsCache | ConvertTo-Json -Compress)) { throw 'Candidate facts cache argv does not add the explicit v1 cache policy to the reviewer-impact treatment.' }
$reviewImpactFactsCacheV2 = @(Get-NativeRunArgs $taskPath $objective $false 1 'go-contract-context-v2' '' $parserPath $parserHash $false '' 0 $true $true)
$expectedReviewImpactFactsCacheV2 = @('--root', $taskPath, 'run', '--autonomous', '--planner-context', 'go-contract-context-v2', '--planner-context-ri-executable', $parserPath, '--planner-context-ri-executable-sha256', $parserHash, '--review-impact-context', '--review-impact-candidate-facts-cache', '--max-parallel', '1', $objective)
if (($reviewImpactFactsCacheV2 | ConvertTo-Json -Compress) -ne ($expectedReviewImpactFactsCacheV2 | ConvertTo-Json -Compress)) { throw 'Candidate facts cache argv does not add the explicit v1 cache policy to the v2 reviewer-impact treatment.' }
Assert-ReviewImpactRunnerBindingShape $false '' 'Native'
Assert-ReviewImpactRunnerBindingShape $true 'go-contract-context-v1' 'Native'
Assert-ReviewImpactRunnerBindingShape $true 'go-contract-context-v2' 'Native'
foreach ($badReviewImpact in @(
    @{ Enabled=$true; Planner=''; Mode='Native' },
    @{ Enabled=$true; Planner='go-source-context-v2'; Mode='Native' },
    @{ Enabled=$true; Planner='go-contract-context-v1'; Mode='PR5Matched' },
    @{ Enabled=$true; Planner='go-contract-context-v2'; Mode='PR5Matched' }
)) {
    $rejected = $false
    try { Assert-ReviewImpactRunnerBindingShape $badReviewImpact.Enabled $badReviewImpact.Planner $badReviewImpact.Mode } catch { $rejected = $true }
    if (-not $rejected) { throw 'Invalid reviewer-impact mode/planner combination was accepted.' }
}
$validFactsCache = @(
    @{ Enabled=$false; ReviewImpact=$false; Planner=''; Mode='Native' },
    @{ Enabled=$true; ReviewImpact=$true; Planner='go-contract-context-v1'; Mode='Native' },
    @{ Enabled=$true; ReviewImpact=$true; Planner='go-contract-context-v2'; Mode='Native' }
)
foreach ($case in $validFactsCache) { Assert-CandidateFactsCacheRunnerBindingShape $case.Enabled $case.ReviewImpact $case.Planner $case.Mode }
foreach ($badFactsCache in @(
    @{ Enabled=$true; ReviewImpact=$false; Planner='go-contract-context-v1'; Mode='Native' },
    @{ Enabled=$true; ReviewImpact=$true; Planner='go-source-context-v2'; Mode='Native' },
    @{ Enabled=$true; ReviewImpact=$true; Planner='go-source-context-v1'; Mode='Native' },
    @{ Enabled=$true; ReviewImpact=$true; Planner='go-contract-context-v1'; Mode='PR5Matched' },
    @{ Enabled=$true; ReviewImpact=$true; Planner='go-contract-context-v2'; Mode='PR5Matched' }
)) {
    $rejected = $false
    try { Assert-CandidateFactsCacheRunnerBindingShape $badFactsCache.Enabled $badFactsCache.ReviewImpact $badFactsCache.Planner $badFactsCache.Mode } catch { $rejected = $true }
    if (-not $rejected) { throw 'Invalid candidate-facts-cache mode/review-context combination was accepted.' }
}
$reviewImpactPrepared = [pscustomobject]@{ review_impact_context_requested = $true }
Assert-ReviewImpactPreparedBinding $reviewImpactPrepared $true
Assert-ReviewImpactPreparedBinding ([pscustomobject]@{}) $false
$rejectedReviewImpactMismatch = $false
try { Assert-ReviewImpactPreparedBinding $reviewImpactPrepared $false } catch { $rejectedReviewImpactMismatch = $true }
if (-not $rejectedReviewImpactMismatch) { throw 'Evaluate accepted a changed reviewer-impact treatment.' }
$factsCachePrepared = [pscustomobject]@{ candidate_facts_cache_version_requested = 1 }
Assert-CandidateFactsCachePreparedBinding $factsCachePrepared $true
Assert-CandidateFactsCachePreparedBinding ([pscustomobject]@{}) $false
$rejectedFactsCacheMismatch = $false
try { Assert-CandidateFactsCachePreparedBinding $factsCachePrepared $false } catch { $rejectedFactsCacheMismatch = $true }
if (-not $rejectedFactsCacheMismatch) { throw 'Evaluate accepted a changed candidate-facts-cache treatment.' }
$rejectedFactsCacheMissing = $false
try { Assert-CandidateFactsCachePreparedBinding ([pscustomobject]@{}) $true } catch { $rejectedFactsCacheMissing = $true }
if (-not $rejectedFactsCacheMissing) { throw 'Evaluate accepted a candidate-facts-cache treatment absent from Prepare.' }
$factsCacheObserved = Assert-CandidateFactsCacheObserved (@{creation=@{execution=@{review_impact_context_version=1;candidate_facts_cache_version=1}}}) $true
if ($factsCacheObserved -ne 1) { throw 'Candidate-facts-cache policy version was not observed.' }
$disabledFactsCacheObserved = Assert-CandidateFactsCacheObserved (@{creation=@{execution=@{}}}) $false
if ($null -ne $disabledFactsCacheObserved) { throw 'Legacy run unexpectedly reports a candidate-facts-cache policy.' }
foreach ($invalidFactsCacheSnapshot in @(
    @{creation=@{execution=@{review_impact_context_version=1;candidate_facts_cache_version=2}}},
    @{creation=@{execution=@{review_impact_context_version=0;candidate_facts_cache_version=1}}},
    @{creation=@{execution=@{review_impact_context_version=1}}}
)) {
    $rejected = $false
    try { Assert-CandidateFactsCacheObserved $invalidFactsCacheSnapshot $true } catch { $rejected = $true }
    if (-not $rejected) { throw 'Invalid or absent observed candidate-facts-cache policy was accepted.' }
}
$preEffectProbeRoot = Join-Path ([IO.Path]::GetTempPath()) ('fabric-v1-cache-invalid-' + [guid]::NewGuid().ToString('N'))
if (Test-Path -LiteralPath $preEffectProbeRoot) { throw 'Pre-effect probe path unexpectedly exists.' }
$rejectedBeforeEffects = $false
try { & $runner -Action Prepare -RunRoot $preEffectProbeRoot -CandidateFactsCache } catch { $rejectedBeforeEffects = $true }
if (-not $rejectedBeforeEffects -or (Test-Path -LiteralPath $preEffectProbeRoot)) {
    throw 'Invalid CandidateFactsCache runner arguments were not rejected before creating the run root.'
}
$fixerPreEffectRoot = Join-Path ([IO.Path]::GetTempPath()) ('fabric-v1-fixer-invalid-' + [guid]::NewGuid().ToString('N'))
$rejectedMissingAccess = $false
try { & $runner -Action Prepare -RunRoot $fixerPreEffectRoot -FixerModel 'gpt-6-sol' } catch { $rejectedMissingAccess = $true }
if (-not $rejectedMissingAccess -or (Test-Path -LiteralPath $fixerPreEffectRoot)) {
    throw 'Missing fixer access policy was not rejected before creating the run root.'
}
$strictConflictRoot = Join-Path ([IO.Path]::GetTempPath()) ('fabric-v1-writer-contract-invalid-' + [guid]::NewGuid().ToString('N'))
$rejectedWriterConflictBeforeEffects = $false
try { & $runner -Action Prepare -RunRoot $strictConflictRoot -ValidateWriterEdits -StrictWriterEdits } catch { $rejectedWriterConflictBeforeEffects = $true }
if (-not $rejectedWriterConflictBeforeEffects -or (Test-Path -LiteralPath $strictConflictRoot)) {
    throw 'Conflicting writer contract flags were not rejected before creating the run root.'
}
$reviewCandidate = 'a' * 64
$reviewImpactSnapshot = @{
    creation = @{ execution = @{ review_impact_context_version = 1 } }
    candidate = @{ files_hash = 'b' * 64 }
    review_impact_contexts = @(@{
        version = 1; candidate_id = $reviewCandidate; candidate_files_hash = 'b' * 64; record_id = 'c' * 64
        changed_path_count = 1; admitted_path_count = 0; deleted_path_count = 0; omitted_count = 0
    })
}
$observedReviewImpact = Assert-ReviewImpactObserved $reviewImpactSnapshot $reviewCandidate $true
if ($observedReviewImpact.version -ne 1 -or $observedReviewImpact.candidate_id -cne $reviewCandidate) { throw 'Observed review-impact evidence was not returned.' }
$rejectedMissingImpact = $false
try { Assert-ReviewImpactObserved (@{ creation=@{ execution=@{ review_impact_context_version=1 } }; candidate=@{ files_hash='b' * 64 }; review_impact_contexts=@() }) $reviewCandidate $true } catch { $rejectedMissingImpact = $true }
if (-not $rejectedMissingImpact) { throw 'Enabled reviewer-impact treatment passed without a candidate-bound record.' }
$legacyImpactObserved = Assert-ReviewImpactObserved (@{ creation=@{ execution=@{} } }) '' $false
if ($null -ne $legacyImpactObserved) { throw 'Legacy mode unexpectedly returned reviewer-impact evidence.' }
$policyRoot = Join-Path $env:TEMP ('isolated-runner-policy-' + [Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $policyRoot | Out-Null
try {
    $policyPath = [IO.Path]::GetFullPath((Join-Path $policyRoot 'policy.json'))
    $policyJson = '{"version":1,"capacity":{"cpu_milli":2000,"memory_mib":2048,"verification_slots":1,"total_runtime_slots":2,"provider_slots":2,"model_slots":2,"runtime_slots":2},"estimate":{"cpu_milli":500,"memory_mib":512,"verification_slots":0,"runtime_slots":1}}'
    [IO.File]::WriteAllText($policyPath, $policyJson, (New-Object System.Text.UTF8Encoding($false)))
    $binding = Get-IsolationPolicyBinding $policyPath
    Assert-CurrentIsolationPolicyBinding $binding
    $isolated = @(Get-NativeRunArgs $taskPath $objective $false 2 '' '' '' '' $true $binding.Path)
    $expectedIsolated = @('--root', $taskPath, 'run', '--autonomous', '--isolated-writers', '--isolation-policy', $policyPath, '--max-parallel', '2', $objective)
    if (($isolated | ConvertTo-Json -Compress) -ne ($expectedIsolated | ConvertTo-Json -Compress)) { throw 'Isolated writer argv does not preserve exact policy path and MaxParallel.' }
    $isolatedReviewImpact = @(Get-NativeRunArgs $taskPath $objective $false 2 'go-contract-context-v1' '' $parserPath $parserHash $true $binding.Path 0 $true)
    $expectedIsolatedReviewImpact = @('--root', $taskPath, 'run', '--autonomous', '--planner-context', 'go-contract-context-v1', '--planner-context-ri-executable', $parserPath, '--planner-context-ri-executable-sha256', $parserHash, '--review-impact-context', '--isolated-writers', '--isolation-policy', $policyPath, '--max-parallel', '2', $objective)
    if (($isolatedReviewImpact | ConvertTo-Json -Compress) -ne ($expectedIsolatedReviewImpact | ConvertTo-Json -Compress)) { throw 'Topology treatment argv does not preserve the pinned contract parser, review flag, resource policy and parallel cap.' }
    $isolatedReviewImpactFactsCache = @(Get-NativeRunArgs $taskPath $objective $false 2 'go-contract-context-v1' '' $parserPath $parserHash $true $binding.Path 0 $true $true)
    $expectedIsolatedReviewImpactFactsCache = @('--root', $taskPath, 'run', '--autonomous', '--planner-context', 'go-contract-context-v1', '--planner-context-ri-executable', $parserPath, '--planner-context-ri-executable-sha256', $parserHash, '--review-impact-context', '--review-impact-candidate-facts-cache', '--isolated-writers', '--isolation-policy', $policyPath, '--max-parallel', '2', $objective)
    if (($isolatedReviewImpactFactsCache | ConvertTo-Json -Compress) -ne ($expectedIsolatedReviewImpactFactsCache | ConvertTo-Json -Compress)) { throw 'Topology treatment argv omits or reorders the candidate facts cache policy.' }

    $priorBinding = [pscustomobject]@{
        isolated_writers_requested = $true
        isolation_policy_path_requested = $binding.Path
        isolation_policy_sha256_requested = $binding.Sha256
        isolation_policy_version_requested = 1
        max_parallel_requested = 2
    }
    Assert-IsolationPolicyPreparedBinding $priorBinding $binding $true 2
    $priorBinding.isolation_policy_sha256_requested = ('0' * 64)
    $rejectedMutation = $false
    try { Assert-IsolationPolicyPreparedBinding $priorBinding $binding $true 2 } catch { $rejectedMutation = $true }
    if (-not $rejectedMutation) { throw 'Changed prepared policy hash was accepted.' }
    $priorBinding.isolation_policy_sha256_requested = $binding.Sha256
    $priorBinding.isolation_policy_path_requested = $binding.Path + '.other'
    $rejectedPathChange = $false
    try { Assert-IsolationPolicyPreparedBinding $priorBinding $binding $true 2 } catch { $rejectedPathChange = $true }
    if (-not $rejectedPathChange) { throw 'Changed prepared policy path was accepted.' }

    $observedJson = '{"creation":{"execution":{"max_parallel":2,"isolated_implementation_version":1,"isolation_capacity":{"cpu_milli":2000,"memory_mib":2048,"verification_slots":1,"total_runtime_slots":2,"provider_slots":[{"provider":"fixture","slots":2}],"model_slots":[{"model":{"provider":"fixture","model":"model-v1"},"slots":2}],"runtime_slots":[{"runtime":{"profile_id":"profile-bound","provider":"fixture","model":"model-v1"},"slots":2}]},"isolation_estimate":{"cpu_milli":500,"memory_mib":512,"verification_slots":0,"runtime_slots":1}},"config":{"writer":{"provider":"fixture","model":"model-v1"}}}}'
    $snapshot = $observedJson | ConvertFrom-Json
    $observed = Assert-IsolationPolicyObserved $snapshot $binding 2
    if ($observed.isolated_implementation_version -ne 1 -or $observed.max_parallel -ne 2) { throw 'Isolated execution policy observation was not retained.' }
    $snapshot.creation.execution.isolation_capacity.memory_mib = 1
    $rejectedObservedCapacity = $false
    try { Assert-IsolationPolicyObserved $snapshot $binding 2 | Out-Null } catch { $rejectedObservedCapacity = $true }
    if (-not $rejectedObservedCapacity) { throw 'Mismatched observed resource capacity was accepted.' }

    foreach ($bad in @(
        @{ Parallel=$true; Isolated=$true; Path=$binding.Path; Max=2 },
        @{ Parallel=$false; Isolated=$true; Path=''; Max=2 },
        @{ Parallel=$false; Isolated=$true; Path=$binding.Path; Max=0 }
    )) {
        $rejected = $false
        try { Get-NativeRunArgs $taskPath $objective $bad.Parallel $bad.Max '' '' '' '' $bad.Isolated $bad.Path | Out-Null } catch { $rejected = $true }
        if (-not $rejected) { throw 'Invalid isolated-writer runner arguments were accepted.' }
    }
    $changedJson = $policyJson.Replace('"total_runtime_slots":2', '"total_runtime_slots":3')
    [IO.File]::WriteAllText($policyPath, $changedJson, (New-Object System.Text.UTF8Encoding($false)))
    $changedBinding = Get-IsolationPolicyBinding $policyPath
    if ($changedBinding.Path -cne $binding.Path -or $changedBinding.Sha256 -ceq $binding.Sha256) { throw 'Policy content mutation did not change the content binding.' }
    $rejectedCurrentMutation = $false
    try { Assert-CurrentIsolationPolicyBinding $binding } catch { $rejectedCurrentMutation = $true }
    if (-not $rejectedCurrentMutation) { throw 'Changed current policy bytes were not detected before Fabric invocation.' }
    $priorBinding.isolation_policy_path_requested = $changedBinding.Path
    $priorBinding.isolation_policy_sha256_requested = $binding.Sha256
    $rejectedChangedContent = $false
    try { Assert-IsolationPolicyPreparedBinding $priorBinding $changedBinding $true 2 } catch { $rejectedChangedContent = $true }
    if (-not $rejectedChangedContent) { throw 'Changed policy bytes were accepted for Evaluate.' }
    [IO.File]::WriteAllText($policyPath, (' ' * (32 * 1024 + 1)), (New-Object System.Text.UTF8Encoding($false)))
    $rejectedOversizedPolicy = $false
    try { Get-IsolationPolicyBinding $policyPath | Out-Null } catch { $rejectedOversizedPolicy = $true }
    if (-not $rejectedOversizedPolicy) { throw 'Oversized isolation policy was accepted.' }
} finally {
    Remove-Item -LiteralPath (Join-Path $policyRoot 'policy.json') -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath $policyRoot -ErrorAction SilentlyContinue
}

$configTestRoot = Join-Path $env:TEMP ('isolated-state-root-' + [Guid]::NewGuid().ToString('N'))
$taskRepo = Join-Path $configTestRoot 'repo'
$taskOutput = Join-Path $configTestRoot 'eval-task'
New-Item -ItemType Directory -Path $taskRepo, $taskOutput | Out-Null
try {
    $configPath = Join-Path $taskRepo 'harness.toml'
    $entry = [pscustomobject]@{ id = 'fixture'; native_argv = @('go', 'test', './...') }
    $baselineConfig = "version = 1`nrepository = 'fixture'`nbase_branch = 'main'`n`n[verification]`nname = 'native'`nargv = ['go', 'test', './...']`ntimeout_seconds = 60`n"
    # `fabric init` currently marshals this default as an empty TOML string.
    $initDefaultConfig = 'controller_state_root = ""' + "`n" + $baselineConfig
    [IO.File]::WriteAllText($configPath, $initDefaultConfig, (New-Object System.Text.UTF8Encoding($false)))
    $externalStateRoot = Join-Path $taskOutput 'controller-state'
    $bound = Set-TaskVerificationConfig $taskRepo $entry $externalStateRoot
    $configured = Get-Content -Raw -LiteralPath $configPath
    $rootCount = ([regex]::Matches($configured, '(?m)^\s*controller_state_root\s*=')).Count
    if ($rootCount -ne 1 -or $configured -notmatch '(?m)^controller_state_root\s*=\s*"') { throw 'Isolated task config omitted or duplicated its external controller state root.' }
    $quotedRoot = [regex]::Match($configured, '(?m)^controller_state_root\s*=\s*(?<value>"(?:[^"\\]|\\.)*")').Groups['value'].Value
    if ([string]::IsNullOrWhiteSpace($quotedRoot) -or (ConvertFrom-Json $quotedRoot) -cne $externalStateRoot) { throw 'TOML controller state root does not bind the exact external path.' }
    if ($bound.ControllerStateRoot -cne $externalStateRoot -or $bound.ConfigSha -cne (Get-FileSha256 $configPath)) { throw 'Task config hash/metadata omitted the external root binding.' }

    [IO.File]::WriteAllText($configPath, $initDefaultConfig, (New-Object System.Text.UTF8Encoding($false)))
    $beforeRejected = Get-FileSha256 $configPath
    $insideRejected = $false
    try { Set-TaskVerificationConfig $taskRepo $entry (Join-Path $taskRepo 'controller-state') | Out-Null } catch { $insideRejected = $true }
    if (-not $insideRejected -or (Get-FileSha256 $configPath) -cne $beforeRejected) { throw 'Inside-checkout state root was not rejected without changing config.' }

    [IO.File]::WriteAllText($configPath, 'controller_state_root = '''' # empty init default' + "`n" + $baselineConfig, (New-Object System.Text.UTF8Encoding($false)))
    $singleQuoted = Set-TaskVerificationConfig $taskRepo $entry $externalStateRoot
    $singleQuotedConfig = Get-Content -Raw -LiteralPath $configPath
    if (([regex]::Matches($singleQuotedConfig, '(?m)^\s*controller_state_root\s*=')).Count -ne 1 -or $singleQuoted.ControllerStateRoot -cne $externalStateRoot) { throw 'Single-quoted empty TOML default was not replaced with one bound external root.' }

    $invalidRootCases = @(
        [pscustomobject]@{ Name = 'nonempty'; Line = 'controller_state_root = "D:\\already-set"' },
        [pscustomobject]@{ Name = 'malformed'; Line = 'controller_state_root = null' },
        [pscustomobject]@{ Name = 'duplicate'; Line = 'controller_state_root = ""' + "`n" + 'controller_state_root = ""' }
    )
    foreach ($invalidRootCase in $invalidRootCases) {
        [IO.File]::WriteAllText($configPath, $invalidRootCase.Line + "`n" + $baselineConfig, (New-Object System.Text.UTF8Encoding($false)))
        $invalidRootHash = Get-FileSha256 $configPath
        $invalidRootRejected = $false
        try { Set-TaskVerificationConfig $taskRepo $entry $externalStateRoot | Out-Null } catch { $invalidRootRejected = $true }
        if (-not $invalidRootRejected -or (Get-FileSha256 $configPath) -cne $invalidRootHash) { throw "Invalid controller state root case '$($invalidRootCase.Name)' did not reject without changing config." }
    }

    [IO.File]::WriteAllText($configPath, $initDefaultConfig, (New-Object System.Text.UTF8Encoding($false)))
    $legacy = Set-TaskVerificationConfig $taskRepo $entry
    $legacyConfig = Get-Content -Raw -LiteralPath $configPath
    if ($legacyConfig -notmatch '(?m)^controller_state_root\s*=\s*""\s*$') { throw 'Default isolated-state field changed on the ordinary evaluation path.' }
    if ($null -ne $legacy.ControllerStateRoot) { throw 'Default task config metadata acquired isolated state provenance.' }
} finally {
    Remove-Item -LiteralPath $configTestRoot -Recurse -Force -ErrorAction SilentlyContinue
}
$invalidBindings = @(
    @{ Mode='go-source-context-v1'; Path=''; Hash='' },
    @{ Mode='go-source-context-v1'; Path=$parserPath; Hash='' },
    @{ Mode='go-source-context-v1'; Path='relative\ri.exe'; Hash=$parserHash },
    @{ Mode='go-source-context-v1'; Path=([IO.Path]::GetDirectoryName($parserPath) + [IO.Path]::DirectorySeparatorChar + '.' + [IO.Path]::DirectorySeparatorChar + [IO.Path]::GetFileName($parserPath)); Hash=$parserHash },
    @{ Mode='go-source-context-v1'; Path=$parserPath; Hash=$parserHash.ToUpperInvariant() },
    @{ Mode='go-source-context-v2'; Path=''; Hash='' },
    @{ Mode='go-source-context-v2'; Path=$parserPath; Hash='' },
    @{ Mode='go-source-context-v2'; Path=$parserPath; Hash=$parserHash.ToUpperInvariant() },
    @{ Mode='go-contract-context-v1'; Path=''; Hash='' },
    @{ Mode='go-contract-context-v1'; Path=$parserPath; Hash='' },
    @{ Mode='go-contract-context-v1'; Path=$parserPath; Hash=$parserHash.ToUpperInvariant() },
    @{ Mode='go-contract-context-v2'; Path=''; Hash='' },
    @{ Mode='go-contract-context-v2'; Path=$parserPath; Hash='' },
    @{ Mode='go-contract-context-v2'; Path=$parserPath; Hash=$parserHash.ToUpperInvariant() },
    @{ Mode='GO-CONTRACT-CONTEXT-V1'; Path=''; Hash='' },
    @{ Mode='GO-CONTRACT-CONTEXT-V2'; Path=''; Hash='' },
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
Write-Output 'PASS: legacy argv is byte-order stable; Go planner context v1/v2 modes bind exact parser provenance; review-impact/cache and fixer/access/writer-contract options match Prepare/Evaluate; access bytes and observed policies are bound; invalid combinations and mutations reject; objectives stay one argument; no provider calls.'
