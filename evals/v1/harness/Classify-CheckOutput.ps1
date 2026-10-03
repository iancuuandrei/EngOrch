# Deterministic classification for Fabric v1 held-out check output.
#
# The old evaluate-v1-baseline.ps1 treated ANY nonzero `go test` exit as an
# expected behavioral failure (FAIL_EXPECTED). That is wrong: compiler errors,
# missing modules, and self-import cycles also exit nonzero but are never an
# intended FAIL. Use Get-FabricV1CheckClassification instead.
#
# Categories:
#   pass                 - exit 0, checks passed.
#   targeted_assertion   - the test binary built and ran, and
#                          --- FAIL: TestFabricV1Heldout is present with no
#                          compiler diagnostics. The only expected baseline FAIL.
#   expected_missing_api - every compiler diagnostic line is the narrow known
#                          not-yet-implemented API shape for the correct task:
#                          go-multierror: `<expr>.ErrorsSnapshot undefined
#                          (type *Error has no field or method ErrorsSnapshot)`;
#                          go-atomic: `<expr>.MarshalText/.UnmarshalText
#                          undefined (type *Bool has no field or method ...)`
#                          or `(*Bool does not implement
#                          encoding.TextMarshaler/TextUnmarshaler
#                          (missing method ...))`. Any mixed or off-task
#                          diagnostic is setup_error.
#   setup_error          - anything else (import cycle, build failed, vet
#                          failure, missing module, arbitrary FAIL without the
#                          held-out assertion). NEVER an intended FAIL; the
#                          runner must surface it and stop, not record
#                          FAIL_EXPECTED.

function Get-FabricV1CheckClassification {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$TaskId,
        [Parameter(Mandatory)][int]$ExitCode,
        [Parameter(Mandatory)][AllowEmptyString()][string]$CombinedOutput
    )
    if ($ExitCode -eq 0) {
        $packageResults = @($CombinedOutput -split "`r?`n" | Where-Object { $_ -match '^ok\s+\S+\s+' })
        $executedPackages = @($packageResults | Where-Object { $_ -notmatch '\[no tests to run\]' })
        if ($executedPackages.Count -eq 0) { return 'setup_error' }
        return 'pass'
    }
    if ($CombinedOutput -match 'import cycle') { return 'setup_error' }

    # Compiler diagnostic lines carry line AND column (`file.go:11:16:`).
    # Runtime assertion output (`file.go:12: message`) has only a line number
    # and must not be treated as a compiler diagnostic.
    $diagLines = @($CombinedOutput -split "`r?`n" | Where-Object { $_ -match '\.go:\d+:\d+:' })
    if ($diagLines.Count -gt 0) {
        foreach ($line in $diagLines) {
            $allowed = $false
            if ($TaskId -eq 'go-multierror') {
                if ($line -match '\.ErrorsSnapshot undefined' -and
                    $line -match '\*Error' -and
                    $line -match 'has no field or method ErrorsSnapshot') {
                    $allowed = $true
                }
            } elseif ($TaskId -eq 'go-atomic') {
                $isMarshalLine = $line -match '\.MarshalText undefined' -or $line -match '\.UnmarshalText undefined'
                $isImplLine = $line -match 'does not implement encoding\.Text(Marshaler|Unmarshaler)' -and $line -match 'missing method (MarshalText|UnmarshalText)'
                if (($isMarshalLine -and $line -match '\*Bool' -and $line -match 'has no field or method (MarshalText|UnmarshalText)') -or $isImplLine) {
                    $allowed = $true
                }
            }
            if (-not $allowed) { return 'setup_error' }
        }
        return 'expected_missing_api'
    }

    if ($CombinedOutput -match '--- FAIL:\s*TestFabricV1Heldout') { return 'targeted_assertion' }
    return 'setup_error'
}
