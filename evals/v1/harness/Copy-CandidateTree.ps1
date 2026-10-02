# Byte-exact candidate copy for Fabric v1 acceptance.
#
# Copy-CandidateTree performs a real byte copy of the finished candidate
# source: changed tracked files, untracked files, and deletions (absent
# files are absent in a fresh destination) are all preserved, unlike
# `git clone HEAD` which drops dirty/untracked candidate changes.
#
# Rules:
# - Filesystem walk only; symlinks/reparse points are never followed and
#   never copied (recorded in skipped_reparse).
# - `.git`, root `.harness/`, and root `harness.toml` (secrets/runtime
#   state) are never copied (recorded in excluded_paths).
# - Everything else, including git-ignored build outputs, IS copied; the
#   manifest lists every copied file, and the identity covers exactly the
#   files tested. The only ignored-source ambiguity is the recorded
#   exclusion list, tracked honestly in the manifest.
# - Identity is deterministic: sorted relative paths (forward slashes),
#   hashed as "path:<p>\nbytes:<n>\n" + raw bytes.
# - The caller must recompute the destination identity with
#   Get-DirectoryTreeIdentity and compare before running tests; any
#   mismatch is BLOCKED, never silently accepted.

function Get-TreeIdentityFromBytes {
    [CmdletBinding()]
    param([Parameter(Mandatory)][AllowEmptyCollection()][object[]]$Entries)
    # Each entry: @{ path = 'a/b'; bytes = [byte[]] }.
    $h = [System.Security.Cryptography.SHA256]::Create()
    try {
        $ms = New-Object System.IO.MemoryStream
        try {
            foreach ($e in ($Entries | Sort-Object { $_.path })) {
                $head = [System.Text.Encoding]::UTF8.GetBytes("path:" + $e.path + "`nbytes:$($e.bytes.Length)`n")
                $ms.Write($head, 0, $head.Length)
                if ($e.bytes.Length -gt 0) { $ms.Write($e.bytes, 0, $e.bytes.Length) }
            }
            $ms.Position = 0
            return ([BitConverter]::ToString($h.ComputeHash($ms)) -replace '-', '').ToLowerInvariant()
        } finally { $ms.Close() }
    } finally { $h.Dispose() }
}

function Copy-CandidateTree {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$Source,
        [Parameter(Mandatory)][string]$Destination
    )
    $Source = [System.IO.Path]::GetFullPath($Source)
    $Destination = [System.IO.Path]::GetFullPath($Destination)
    if (-not (Test-Path -LiteralPath $Source -PathType Container)) { throw "Copy source missing: $Source" }
    if (Test-Path -LiteralPath $Destination) { throw "Copy destination already exists: $Destination" }
    New-Item -ItemType Directory -Force $Destination | Out-Null

    $sep = [string][System.IO.Path]::DirectorySeparatorChar
    $manifest = New-Object System.Collections.Generic.List[object]
    $hashInputs = New-Object System.Collections.Generic.List[object]
    $skippedReparse = New-Object System.Collections.Generic.List[string]
    $stack = New-Object System.Collections.Stack
    $stack.Push('')
    while ($stack.Count -gt 0) {
        $rel = $stack.Pop()
        $srcDir = if ($rel -eq '') { $Source } else { Join-Path $Source ($rel -replace '/', $sep) }
        foreach ($item in Get-ChildItem -LiteralPath $srcDir -Force) {
            $itemRel = if ($rel -eq '') { $item.Name } else { "$rel/$($item.Name)" }
            if ($rel -eq '' -and $item.Name -eq '.git') { continue }
            if ($rel -eq '' -and ($item.Name -eq '.harness' -or $item.Name -eq 'harness.toml')) { continue }
            if (($item.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
                [void]$skippedReparse.Add($itemRel)
                continue
            }
            if ($item.PSIsContainer) {
                $null = New-Item -ItemType Directory -Force (Join-Path $Destination ($itemRel -replace '/', $sep))
                [void]$stack.Push($itemRel)
            } else {
                $srcBytes = [System.IO.File]::ReadAllBytes($item.FullName)
                [System.IO.File]::WriteAllBytes((Join-Path $Destination ($itemRel -replace '/', $sep)), $srcBytes)
                $h = [System.Security.Cryptography.SHA256]::Create()
                try { $hash = ([BitConverter]::ToString($h.ComputeHash($srcBytes)) -replace '-', '').ToLowerInvariant() }
                finally { $h.Dispose() }
                [void]$manifest.Add([pscustomobject]@{ path = $itemRel; sha256 = $hash; bytes = $srcBytes.Length })
                [void]$hashInputs.Add([pscustomobject]@{ path = $itemRel; bytes = $srcBytes })
            }
        }
    }
    $sorted = @($manifest.ToArray() | Sort-Object path)
    return [ordered]@{
        source            = $Source
        destination       = $Destination
        files             = $sorted
        file_count        = $sorted.Count
        identity          = (Get-TreeIdentityFromBytes -Entries $hashInputs.ToArray())
        skipped_reparse   = $skippedReparse.ToArray()
        excluded_paths    = @('.git', '.harness/', 'harness.toml')
        ignored_ambiguity = 'Only .git, root .harness/ and root harness.toml plus reparse points are excluded; all other files including git-ignored outputs are copied and hashed. Acceptance ran on exactly the manifested files.'
    }
}

function Get-DirectoryTreeIdentity {
    [CmdletBinding()]
    param([Parameter(Mandatory)][string]$Root)
    $Root = [System.IO.Path]::GetFullPath($Root)
    $sep = [string][System.IO.Path]::DirectorySeparatorChar
    $entries = New-Object System.Collections.Generic.List[object]
    $stack = New-Object System.Collections.Stack
    $stack.Push('')
    while ($stack.Count -gt 0) {
        $rel = $stack.Pop()
        $dir = if ($rel -eq '') { $Root } else { Join-Path $Root ($rel -replace '/', $sep) }
        foreach ($item in Get-ChildItem -LiteralPath $dir -Force) {
            $itemRel = if ($rel -eq '') { $item.Name } else { "$rel/$($item.Name)" }
            if ($rel -eq '' -and $item.Name -eq '.git') { continue }
            if ($rel -eq '' -and ($item.Name -eq '.harness' -or $item.Name -eq 'harness.toml')) { continue }
            if (($item.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) { continue }
            if ($item.PSIsContainer) { [void]$stack.Push($itemRel); continue }
            [void]$entries.Add([pscustomobject]@{ path = $itemRel; bytes = [System.IO.File]::ReadAllBytes($item.FullName) })
        }
    }
    return (Get-TreeIdentityFromBytes -Entries $entries.ToArray())
}
