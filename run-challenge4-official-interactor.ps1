param(
    [string]$ConfigPath,
    [string]$StatePath,
    [string]$WorkspacePath,
    [string[]]$InteractorArgs
)

$repoRoot = $PSScriptRoot

if (-not $WorkspacePath) {
    $WorkspacePath = Join-Path $repoRoot "contracts/forwarder-blind/dex-interactor-official"
}

if (-not $ConfigPath) {
    throw "ConfigPath is required."
}

if (-not $StatePath) {
    $StatePath = Join-Path $repoRoot "state/challenge4/official/deploy.toml"
}

if (-not (Test-Path $ConfigPath)) {
    throw "Challenge 4 official config file not found: $ConfigPath"
}

if (-not (Test-Path $WorkspacePath)) {
    throw "Challenge 4 official interactor workspace not found: $WorkspacePath"
}

$stateDir = Split-Path -Parent $StatePath
if ($stateDir -and -not (Test-Path $stateDir)) {
    New-Item -ItemType Directory -Path $stateDir -Force | Out-Null
}

$resolvedConfigPath = (Resolve-Path $ConfigPath).Path
$resolvedWorkspacePath = (Resolve-Path $WorkspacePath).Path
$resolvedStatePath = [System.IO.Path]::GetFullPath($StatePath)

$env:FORWARDER_BLIND_CONFIG = $resolvedConfigPath
$env:FORWARDER_BLIND_STATE = $resolvedStatePath
$env:FORWARDER_BLIND_WORKSPACE = $resolvedWorkspacePath

if (-not $InteractorArgs -or $InteractorArgs.Count -eq 0) {
    throw "No interactor command provided. Example: -InteractorArgs @('balances')"
}

Push-Location $resolvedWorkspacePath
try {
    $binaryPath = Join-Path $resolvedWorkspacePath "target\debug\forwarder-blind-dex-interactor.exe"
    if (-not (Test-Path $binaryPath)) {
        & cargo build --quiet
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }

    & $binaryPath @InteractorArgs
    exit $LASTEXITCODE
}
finally {
    Pop-Location
}
