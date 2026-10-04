$ErrorActionPreference = 'Stop'
$runner = Join-Path $PSScriptRoot '..\..\..\scripts\evaluate-v1.ps1'
$parseTokens = $null
$parseErrors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile($runner, [ref]$parseTokens, [ref]$parseErrors)
if ($parseErrors.Count -ne 0) { throw 'Evaluation runner does not parse.' }
$builder = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-NativeGoArgs' }, $true)
if ($null -eq $builder) { throw 'Native task policy builder missing.' }
. ([scriptblock]::Create($builder.Extent.Text))
$heldoutBuilder = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-HeldoutSource' }, $true)
if ($null -eq $heldoutBuilder) { throw 'Held-out source mapping function missing.' }
. ([scriptblock]::Create($heldoutBuilder.Extent.Text))
$heldoutSourcesBuilder = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-HeldoutSources' }, $true)
if ($null -eq $heldoutSourcesBuilder) { throw 'Held-out source composition function missing.' }
. ([scriptblock]::Create($heldoutSourcesBuilder.Extent.Text))
$heldoutArgBuilder = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-HeldoutGoArgs' }, $true)
if ($null -eq $heldoutArgBuilder) { throw 'Held-out argv builder missing.' }
. (Join-Path $PSScriptRoot 'Go-ArgvPolicy.ps1')
. ([scriptblock]::Create($heldoutArgBuilder.Extent.Text))
$heldoutDir = (Resolve-Path (Join-Path $PSScriptRoot '..\heldout')).Path
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
$manifestPath = Join-Path $PSScriptRoot '..\manifest.json'
$manifest = Get-Content -Raw -LiteralPath $manifestPath | ConvertFrom-Json
$entry = @($manifest.repositories | Where-Object id -eq 'go-humanize-feature-performance')
if ($entry.Count -ne 1 -or $entry[0].check -ne 'humanize-feature-performance') {
    throw 'Combined humanize manifest entry/check mapping is missing or ambiguous.'
}
$topologyFixturePath = Join-Path $PSScriptRoot '..\topology\humanize-feature-performance\fixture.json'
$topologyFixture = Get-Content -Raw -LiteralPath $topologyFixturePath | ConvertFrom-Json
if ($topologyFixture.manifest_task -ne $entry[0].id -or
    $topologyFixture.review_impact_context_requested -ne $true -or
    $topologyFixture.planner_context -cne 'go-contract-context-v1' -or
    $topologyFixture.review_impact_context_version_required -ne 1 -or
    $topologyFixture.candidate_facts_cache_requested -ne $true -or
    $topologyFixture.candidate_facts_cache_version_required -ne 1 -or
    $topologyFixture.current_product_gate -cne 'not_qualified_until_a_fresh_matched_pair_is_prepared_and_accepted') {
    throw 'Humanize topology fixture is not bound to the supported opt-in reviewer-impact runner treatment.'
}
$sources = @(Get-HeldoutSources $entry[0].check)
if ($sources.Count -ne 3 -or
    ($sources.RelativePath -join ',') -cne 'fabric_v1_heldout_humanize_test.go,fabric_v1_heldout_commaf_test.go,fabric_v1_heldout_test.go') {
    throw 'Combined held-out must compose the two original fixture sources and one exact-selector wrapper as separate files.'
}
if (-not $sources[0].Content.Contains('func runFabricV1HeldoutHumanize(t *testing.T)') -or
    -not $sources[0].Content.Contains('ParseBytes("1_024 B")') -or
    -not $sources[1].Content.Contains('func runFabricV1HeldoutCommaf(t *testing.T)') -or
    -not $sources[1].Content.Contains('checkFabricV1HeldoutCommafOutputEquivalence') -or
    -not $sources[1].Content.Contains('checkFabricV1HeldoutCommafAllocationBudget')) {
    throw 'A composed original held-out oracle is missing its required assertion after test-entry renaming.'
}
if (-not $sources[2].Content.Contains('func TestFabricV1Heldout(t *testing.T)') -or
    -not $sources[2].Content.Contains('t.Run("ParseBytesUnderscores", runFabricV1HeldoutHumanize)') -or
    -not $sources[2].Content.Contains('t.Run("CommafPerformance", runFabricV1HeldoutCommaf)')) {
    throw 'Combined exact-selector wrapper does not require both original held-out tests.'
}
if (([regex]::Matches($sources[2].Content, 'func TestFabricV1Heldout\(t \*testing\.T\)')).Count -ne 1) {
    throw 'Composed fixture must expose exactly one runner-selected wrapper.'
}
# The checked-in source fixtures remain byte-identical and retain their
# individual TestFabricV1Heldout entry points for their existing task IDs.
if (-not ((Get-HeldoutSource 'humanize').Contains('func TestFabricV1Heldout(t *testing.T)')) -or
    -not ((Get-HeldoutSource 'commaf-performance').Contains('func TestFabricV1Heldout(t *testing.T)'))) {
    throw 'Original standalone held-out entry points changed.'
}
foreach ($legacyCheck in @('humanize', 'commaf-performance')) {
    $legacySources = @(Get-HeldoutSources $legacyCheck)
    if ($legacySources.Count -ne 1 -or $legacySources[0].RelativePath -cne 'fabric_v1_heldout_test.go' -or
        $legacySources[0].Content -cne (Get-HeldoutSource $legacyCheck)) {
        throw "Existing individual held-out behavior changed for $legacyCheck."
    }
}
$heldoutArgs = @(Get-HeldoutGoArgs $entry[0] '')
$expectedHeldoutArgs = @('go', 'test', '-run=^TestFabricV1Heldout$', '-count=1', './...')
if (($heldoutArgs | ConvertTo-Json -Compress) -cne ($expectedHeldoutArgs | ConvertTo-Json -Compress)) {
    throw 'Combined humanize held-out argv differs from the exact runner-selected invocation.'
}
Write-Output 'PASS: native argv policy composes both unchanged humanize oracles as separate files under one exact-selector wrapper; legacy task mappings remain intact. No provider calls.'
