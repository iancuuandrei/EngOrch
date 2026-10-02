# Verification argv parsing for Fabric v1 evaluation.
#
# Actual `fabric init` emits single-quoted TOML
# (`argv = ['go', 'test', './...']`), while rewritten policy uses
# double-quoted basic strings. Read-VerificationArgv supports both simple
# quote forms, rejects malformed or extra unparsed material and empty
# lists explicitly, and never guesses: no argv line, malformed items, or
# an empty list all throw with an exact diagnostic.

function Read-VerificationArgv {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][AllowEmptyString()][string]$Text
    )
    $line = [regex]::Match($Text, '(?m)^argv\s*=\s*\[(?<items>[^\]]*)\]')
    if (-not $line.Success) { throw 'no verification argv line found; refusing to guess config shape' }
    $items = $line.Groups['items'].Value
    if ($items -match '^\s*$') { throw 'verification argv list is empty' }
    $tokenMatches = [regex]::Matches($items, '"(?:[^"\\]|\\.)*"|''[^'']*''')
    $shape = $items
    foreach ($tok in $tokenMatches) {
        $at = $shape.IndexOf($tok.Value, [System.StringComparison]::Ordinal)
        $shape = $shape.Substring(0, $at) + 'T' + $shape.Substring($at + $tok.Value.Length)
    }
    if ($shape -notmatch '^\s*T(\s*,\s*T)*\s*,?\s*$') {
        throw "malformed verification argv material: $($shape.Trim())"
    }
    $parsed = New-Object System.Collections.Generic.List[string]
    foreach ($tok in $tokenMatches) {
        $raw = $tok.Value
        if ($raw.StartsWith("'")) {
            [void]$parsed.Add($raw.Substring(1, $raw.Length - 2))
        } else {
            $inner = $raw.Substring(1, $raw.Length - 2)
            if ($inner -match '\\[^"\\]') { throw "unsupported escape in verification argv token: $raw" }
            [void]$parsed.Add(($inner -replace '\\"', '"' -replace '\\\\', '\'))
        }
    }
    if ($parsed.Count -eq 0) { throw 'verification argv list is empty' }
    return $parsed.ToArray()
}
