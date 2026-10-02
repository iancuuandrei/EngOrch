# Standalone copy-behavior fixtures for Copy-CandidateTree.ps1.
# Run: pwsh -File evals/v1/harness/Test-CopyFixtures.ps1
# Exit 0 when every fixture passes; nonzero with a diagnostic otherwise.
# No git, network, provider, or credentials required.

$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'Copy-CandidateTree.ps1')

$failures = New-Object System.Collections.Generic.List[string]
function Assert([bool]$Cond, [string]$Name) {
    if ($Cond) { Write-Output "PASS $Name" }
    else { $failures.Add($Name); Write-Output "FAIL $Name" }
}

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ('fabric-copy-fixture-' + [Guid]::NewGuid().ToString('N'))
$src = Join-Path $tmp 'src'
$dst = Join-Path $tmp 'dst'
New-Item -ItemType Directory -Force (Join-Path $src 'sub') | Out-Null
New-Item -ItemType Directory -Force (Join-Path $src '.harness') | Out-Null

# Fixture tree: changed tracked-style file, untracked file, nested file,
# deletion witness (file absent), runtime secret, held-out-shaped test file.
[System.IO.File]::WriteAllText((Join-Path $src 'base.txt'), "base-v2`n", [System.Text.UTF8Encoding]::new($false))
[System.IO.File]::WriteAllText((Join-Path $src 'added.txt'), "untracked candidate change`n", [System.Text.UTF8Encoding]::new($false))
[System.IO.File]::WriteAllText((Join-Path $src 'sub/nested.txt'), "nested`n", [System.Text.UTF8Encoding]::new($false))
[System.IO.File]::WriteAllText((Join-Path $src '.harness/secret.json'), '{"token":"must-never-copy"}', [System.Text.UTF8Encoding]::new($false))
[System.IO.File]::WriteAllText((Join-Path $src 'harness.toml'), 'secret config', [System.Text.UTF8Encoding]::new($false))
# 'deleted.txt' is deliberately absent: deletions stay absent in a fresh copy.

$linkOk = $true
try { $null = New-Item -ItemType SymbolicLink -Path (Join-Path $src 'link.txt') -Target (Join-Path $src 'base.txt') -ErrorAction Stop }
catch { $linkOk = $false }

$result = Copy-CandidateTree -Source $src -Destination $dst

Assert ((Test-Path -LiteralPath (Join-Path $dst 'base.txt')) -and ([System.IO.File]::ReadAllText((Join-Path $dst 'base.txt')) -eq "base-v2`n")) 'changed-tracked-file-preserved'
Assert ((Test-Path -LiteralPath (Join-Path $dst 'added.txt'))) 'untracked-file-preserved'
Assert ((Test-Path -LiteralPath (Join-Path $dst 'sub/nested.txt'))) 'nested-file-preserved'
Assert (-not (Test-Path -LiteralPath (Join-Path $dst 'deleted.txt'))) 'deletion-stays-absent'
Assert (-not (Test-Path -LiteralPath (Join-Path $dst '.harness'))) '.harness-never-copied'
Assert (-not (Test-Path -LiteralPath (Join-Path $dst 'harness.toml'))) 'harness.toml-never-copied'
if ($linkOk) {
    Assert (-not (Test-Path -LiteralPath (Join-Path $dst 'link.txt')) -and ($result.skipped_reparse -contains 'link.txt')) 'symlink-skipped-and-recorded'
} else {
    Write-Output 'SKIP symlink-skipped-and-recorded (fixture link creation unavailable)'
}
Assert ($result.files.Count -eq $result.file_count -and $result.file_count -eq 3) 'manifest-count-exact'
Assert ($result.excluded_paths -contains '.harness/') 'exclusions-recorded'

# Deterministic identity: recompute over destination, must match exactly.
$recomputed = Get-DirectoryTreeIdentity -Root $dst
Assert ($recomputed -eq $result.identity) 'identity-recomputation-matches'

# Tamper evidence: mutating the copy changes its identity.
[System.IO.File]::AppendAllText((Join-Path $dst 'added.txt'), "tamper`n")
Assert ((Get-DirectoryTreeIdentity -Root $dst) -ne $result.identity) 'tamper-changes-identity'

# Guarded cleanup: only delete when the resolved absolute target sits
# directly under the exact GetTempPath parent with a GUID-prefixed
# fabric-copy-fixture leaf (Windows filesystem rules: never delete an
# unresolved or out-of-tree path).
$resolvedTmp = [System.IO.Path]::GetFullPath($tmp)
$tempParent = [System.IO.Path]::GetFullPath([System.IO.Path]::GetTempPath()).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
$tmpLeaf = Split-Path -Leaf $resolvedTmp
$tmpParent = Split-Path -Parent $resolvedTmp
if (($tmpLeaf -notmatch '^fabric-copy-fixture-[0-9a-f]{32}$') -or ($tmpParent -ine $tempParent)) {
    throw "Refusing to delete unexpected fixture path: $resolvedTmp"
}
Remove-Item -LiteralPath $resolvedTmp -Recurse -Force -ErrorAction SilentlyContinue
if ($failures.Count -gt 0) { throw "Copy fixtures failed: $($failures -join ', ')" }
Write-Output "All copy fixtures passed."
