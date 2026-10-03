# Standalone argument-policy fixtures for Go-ArgvPolicy.ps1.
# Run: pwsh -File evals/v1/harness/Test-ArgvPolicy.ps1
# Exit 0 when every fixture passes; nonzero with a diagnostic otherwise.
# No git, Go toolchain, network, provider, or credentials required.

$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'Go-ArgvPolicy.ps1')

$failures = New-Object System.Collections.Generic.List[string]
function Assert([bool]$Cond, [string]$Name) {
    if ($Cond) { Write-Output "PASS $Name" }
    else { $failures.Add($Name); Write-Output "FAIL $Name" }
}

# Full manifest policy argv keeps the executable for Fabric verification.
$split = Split-GoInvocation -FullArgv @('go', 'test', '-count=1', '.')
Assert (($split.Exe -ceq 'go') -and (($split.Args -join ' ') -eq 'test -count=1 .')) 'package-policy-strips-go-executable'

# Legacy go-difflib explicit source-file argv.
$split = Split-GoInvocation -FullArgv @('go', 'test', '-count=1', 'difflib/difflib.go', 'difflib/difflib_test.go')
Assert (($split.Args.Count -eq 4) -and ($split.Args[3] -ceq 'difflib/difflib_test.go')) 'file-policy-strips-go-executable'

$split = Split-GoInvocation -FullArgv @('go', 'test', '-vet=off', '-count=1', 'difflib/difflib.go', 'difflib/difflib_test.go')
Assert (($split.Args -join ' ') -ceq 'test -vet=off -count=1 difflib/difflib.go difflib/difflib_test.go') 'legacy-vet-scope-preserved-after-strip'

$heldout = @(Add-FabricV1HeldoutArguments -FullArgv @('go', 'test', '-vet=off', '-count=1', 'difflib/difflib.go', 'difflib/difflib_test.go') -HeldoutFile 'difflib/fabric_v1_heldout_test.go')
Assert (($heldout -join ' ') -ceq 'go test -run=^TestFabricV1Heldout$ -vet=off -count=1 difflib/difflib.go difflib/difflib_test.go difflib/fabric_v1_heldout_test.go') 'heldout-file-compiled-and-selector-before-files'
. (Join-Path $PSScriptRoot 'Classify-CheckOutput.ps1')
Assert ((Get-FabricV1CheckClassification -TaskId 'go-difflib' -ExitCode 0 -CombinedOutput 'ok command-line-arguments 0.1s [no tests to run]') -ceq 'setup_error') 'zero-tests-is-not-heldout-pass'
Assert ((Get-FabricV1CheckClassification -TaskId 'afero' -ExitCode 0 -CombinedOutput "ok example/root 0.1s`nok example/other 0.1s [no tests to run]") -ceq 'pass') 'unrelated-empty-package-does-not-hide-executed-tests'
Assert ((Get-FabricV1CheckClassification -TaskId 'afero' -ExitCode 0 -CombinedOutput '? example/root [no test files]') -ceq 'setup_error') 'missing-test-files-is-not-heldout-pass'

# Held-out argv appends the selector (and the held-out file for difflib).
$split = Split-GoInvocation -FullArgv @('go', 'test', '-count=1', '.', '-run=^TestFabricV1Heldout$')
Assert ($split.Args -contains '-run=^TestFabricV1Heldout$') 'heldout-selector-preserved-after-strip'

# Windows-only upstream skip survives the strip.
$split = Split-GoInvocation -FullArgv @('go', 'test', '-count=1', '.', '-skip', '^TestNocmpIntegration$')
Assert (($split.Args -join ' ') -eq 'test -count=1 . -skip ^TestNocmpIntegration$') 'skip-flag-preserved-after-strip'

# Narrow validation: empty, missing executable, or wrong executable rejected.
$rejected = $false
try { $null = Split-GoInvocation -FullArgv @() } catch { $rejected = $true }
Assert ($rejected) 'empty-argv-rejected'

$rejected = $false
try { $null = Split-GoInvocation -FullArgv @('test', '-count=1', '.') } catch { $rejected = $true }
Assert ($rejected) 'missing-executable-rejected'

$rejected = $false
try { $null = Split-GoInvocation -FullArgv @('gotest', 'test', '.') } catch { $rejected = $true }
Assert ($rejected) 'wrong-executable-rejected'

if ($failures.Count -gt 0) { throw "Argv policy fixtures failed: $($failures -join ', ')" }
Write-Output 'All argv policy fixtures passed.'
