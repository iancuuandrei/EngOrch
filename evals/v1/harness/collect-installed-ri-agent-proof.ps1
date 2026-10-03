[CmdletBinding()]
param(
    [Parameter(Mandatory=$true)][string]$InstalledRI,
    [string]$Acceptance,
    [string]$RunId,
    [string]$RepositoryRoot,
    [string]$ControllerJournal,
    [string]$RuntimeStateRoot,
    [string[]]$RuntimeJournal = @(),
    [string]$FabricCLI
)
$ErrorActionPreference = 'Stop'
$collector = Join-Path $PSScriptRoot 'collect-installed-ri-agent-proof.py'
if (-not (Test-Path -LiteralPath $collector -PathType Leaf)) { throw 'Proof collector Python source is missing.' }
$python = Get-Command python -ErrorAction SilentlyContinue
if (-not $python) { $python = Get-Command py -ErrorAction SilentlyContinue }
if (-not $python) { throw 'Python 3 is required for read-only SQLite journal inspection.' }
$pythonArgs = @()
if ($python.Name -eq 'py.exe' -or $python.Name -eq 'py') { $pythonArgs += '-3' }
$pythonArgs += $collector
$pythonArgs += @('--installed-ri', $InstalledRI)
if ($Acceptance) { $pythonArgs += @('--acceptance', $Acceptance) }
if ($RunId) { $pythonArgs += @('--run-id', $RunId) }
if ($RepositoryRoot) { $pythonArgs += @('--repository-root', $RepositoryRoot) }
if ($ControllerJournal) { $pythonArgs += @('--controller-journal', $ControllerJournal) }
if ($RuntimeStateRoot) { $pythonArgs += @('--runtime-state-root', $RuntimeStateRoot) }
if ($FabricCLI) { $pythonArgs += @('--fabric-cli', $FabricCLI) }
foreach ($path in $RuntimeJournal) { $pythonArgs += @('--runtime-journal', $path) }
& $python.Source @pythonArgs
exit $LASTEXITCODE
