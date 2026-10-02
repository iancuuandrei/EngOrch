# Go argv policy for Fabric v1 evaluation.
#
# Manifest `native_argv` deliberately keeps the FULL Fabric verification
# argv, executable included (`["go", "test", ...]`), because that exact
# array is written into harness.toml required checks. Local `go test`
# invocations run the pinned `$GoExe` binary directly, so the leading `go`
# executable element must be stripped at helper invocation time -- never
# from the recorded policy. Passing the full array to `$GoExe` would run
# `go.exe go test ...` and fail with `unknown command go`.
#
# Split-GoInvocation validates that the policy array starts with exactly
# `go` and returns the executable name plus the remaining arguments.

function Split-GoInvocation {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][AllowEmptyCollection()][string[]]$FullArgv
    )
    if ($FullArgv.Count -eq 0) { throw 'Go argv policy is empty; expected full argv starting with go' }
    if ($FullArgv[0] -cne 'go') { throw "Go argv policy must start with the go executable, got: $($FullArgv[0])" }
    return [ordered]@{
        Exe  = 'go'
        Args = @($FullArgv | Select-Object -Skip 1)
    }
}

# Go stops parsing its own flags when it reaches explicit source files.
# Keep the held-out selector before every package/file argument so the test
# source is compiled instead of becoming an ignored test-binary argument.
function Add-FabricV1HeldoutArguments {
    param([Parameter(Mandatory)][string[]]$FullArgv, [string]$HeldoutFile)
    $null = Split-GoInvocation -FullArgv $FullArgv
    if ($FullArgv.Count -lt 3 -or $FullArgv[1] -cne 'test') { throw 'Held-out policy requires go test with a package or source file' }
    $result = @($FullArgv[0..1] + @('-run=^TestFabricV1Heldout$') + $FullArgv[2..($FullArgv.Count - 1)])
    if (-not [string]::IsNullOrWhiteSpace($HeldoutFile)) { $result += $HeldoutFile }
    return $result
}
