$ErrorActionPreference = 'Stop'
$runner = Join-Path $PSScriptRoot '..\..\..\scripts\evaluate-v1.ps1'
$parseTokens = $null; $parseErrors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile($runner, [ref]$parseTokens, [ref]$parseErrors)
if ($parseErrors.Count -ne 0) { throw 'Runner does not parse.' }
$builder = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Get-UsageMetrics' }, $true)
. ([scriptblock]::Create($builder.Extent.Text))
function New-UsageRow([int64]$InputTotal, [int64]$Cached, [int64]$OutputTotal, [int64]$Reasoning) {
    [pscustomobject]@{ receipt_matched = $true; usage = @{ provider_usage = @{
        input_tokens = $InputTotal; output_tokens = $OutputTotal
        accounting = @{ coverage = 'OBSERVED'; delta = @{ inputTokens = $InputTotal; cachedInputTokens = $Cached; outputTokens = $OutputTotal; reasoningOutputTokens = $Reasoning } }
    } } }
}
$usage = @{ invocations = @((New-UsageRow 100 75 20 8), (New-UsageRow 200 150 30 10)) }
$metrics = Get-UsageMetrics $usage
if ($metrics.InputTokens -ne 300 -or $metrics.CachedInputTokens -ne 225 -or $metrics.UncachedInputTokens -ne 75 -or
    $metrics.OutputTokens -ne 50 -or $metrics.ReasoningOutputTokens -ne 18 -or $metrics.TokenTypeCoverage -ne 'observed' -or $null -ne $metrics.ProviderCalls) {
    throw 'Token subsets were double counted or observations changed.'
}
$unknown = Get-UsageMetrics $null
if ($null -ne $unknown.InputTokens -or $null -ne $unknown.CachedInputTokens -or $unknown.TokenTypeCoverage -ne 'unknown') { throw 'Unknown became zero.' }
$partial = Get-UsageMetrics @{ invocations = @((New-UsageRow 100 75 20 8), @{ receipt_matched = $false; usage = $null }) }
if ($null -ne $partial.CachedInputTokens -or $null -ne $partial.UncachedInputTokens -or $partial.TokenTypeCoverage -ne 'partial') { throw 'Incomplete token types presented as complete.' }
$unmatchedRow = New-UsageRow 200 150 30 10
$unmatchedRow.receipt_matched = $false
$unmatched = Get-UsageMetrics @{ invocations = @((New-UsageRow 100 75 20 8), $unmatchedRow) }
if ($null -ne $unmatched.CachedInputTokens -or $null -ne $unmatched.ReasoningOutputTokens -or $unmatched.TokenTypeCoverage -ne 'partial') { throw 'Unmatched usage presented as complete.' }
$invalid = Get-UsageMetrics @{ invocations = @((New-UsageRow 100 101 20 8)) }
if ($null -ne $invalid.CachedInputTokens -or $invalid.TokenTypeCoverage -ne 'unknown') { throw 'Invalid cached subset accepted.' }
Write-Output 'PASS: token types stay separate; unknown/partial preserved; no provider calls.'
