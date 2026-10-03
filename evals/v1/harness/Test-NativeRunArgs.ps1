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
. ([scriptblock]::Create($goMode.Extent.Text))
. ([scriptblock]::Create($fileHash.Extent.Text))
. ([scriptblock]::Create($contextValidator.Extent.Text))
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
    @{ Mode='GO-CONTRACT-CONTEXT-V1'; Path=''; Hash='' },
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
Write-Output 'PASS: legacy argv is byte-order stable; v1/v2/contract-v1 Go planner treatments bind exact parser provenance; isolated mode binds policy and a validated external controller state root into the task config hash; inside-checkout roots reject; invalid combinations and mutations reject; objectives stay one argument; no provider calls.'
