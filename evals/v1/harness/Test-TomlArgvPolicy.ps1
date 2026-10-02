# Standalone TOML argv-policy fixtures for Toml-ArgvPolicy.ps1.
# Run: pwsh -File evals/v1/harness/Test-TomlArgvPolicy.ps1
# Exit 0 when every fixture passes; nonzero with a diagnostic otherwise.
# No git, Go toolchain, network, provider, or credentials required.
# Fixtures run before any model calls: parsing is pure local validation.

$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'Toml-ArgvPolicy.ps1')

$failures = New-Object System.Collections.Generic.List[string]
function Assert([bool]$Cond, [string]$Name) {
    if ($Cond) { Write-Output "PASS $Name" }
    else { $failures.Add($Name); Write-Output "FAIL $Name" }
}

function Assert-Throws([scriptblock]$Code, [string]$Pattern, [string]$Name) {
    try { $null = & $Code; $failures.Add($Name); Write-Output "FAIL $Name (no throw)" }
    catch {
        if ($_.Exception.Message -match $Pattern) { Write-Output "PASS $Name" }
        else { $failures.Add($Name); Write-Output "FAIL $Name (wrong message: $($_.Exception.Message))" }
    }
}

# Actual init-shaped TOML (single quotes) parses to the exact init default.
$fixture = Get-Content -Raw -LiteralPath (Join-Path $PSScriptRoot 'fixtures/init-harness.sample.toml')
$parsed = @(Read-VerificationArgv -Text $fixture)
Assert ((($parsed -join ' ') -ceq 'go test ./...')) 'init-shaped-single-quoted-default'

# Double-quoted form (rewritten policy) parses identically.
$parsed = @(Read-VerificationArgv -Text "argv = [`"go`", `"test`", `"./...`"]")
Assert ((($parsed -join ' ') -ceq 'go test ./...')) 'double-quoted-default'

# Mixed quote forms parse.
$parsed = @(Read-VerificationArgv -Text "argv = ['go', `"test`", './...']")
Assert ((($parsed -join ' ') -ceq 'go test ./...')) 'mixed-quote-forms'

# Rewritten task policy round-trips.
$parsed = @(Read-VerificationArgv -Text 'argv = ["go", "test", "-count=1", "."]')
Assert ((($parsed -join ' ') -ceq 'go test -count=1 .')) 'rewritten-policy-roundtrip'

# Empty list rejected explicitly.
Assert-Throws { Read-VerificationArgv -Text 'argv = []' } 'empty' 'empty-list-rejected'
Assert-Throws { Read-VerificationArgv -Text 'argv = [   ]' } 'empty' 'blank-list-rejected'

# Malformed and extra unparsed material rejected.
Assert-Throws { Read-VerificationArgv -Text 'argv = [go, test]' } 'malformed' 'unquoted-tokens-rejected'
Assert-Throws { Read-VerificationArgv -Text "argv = ['go' 'test']" } 'malformed' 'missing-comma-rejected'
Assert-Throws { Read-VerificationArgv -Text "argv = ['go'; 'test']" } 'malformed' 'semicolon-rejected'
Assert-Throws { Read-VerificationArgv -Text "argv = ['go', 42]" } 'malformed' 'nonstring-token-rejected'

# Missing argv line rejected without guessing.
Assert-Throws { Read-VerificationArgv -Text "version = 1`nname = 'unit'" } 'no verification argv' 'missing-argv-rejected'

if ($failures.Count -gt 0) { throw "TOML argv policy fixtures failed: $($failures -join ', ')" }
Write-Output 'All TOML argv policy fixtures passed.'
