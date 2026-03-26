param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("0", "1", "2")]
    [string]$Shard,
    [string]$ConfigPath,
    [string]$StatePath,
    [string]$WorkspacePath,
    [string[]]$InteractorArgs
)

$repoRoot = $PSScriptRoot

if (-not $ConfigPath) {
    $ConfigPath = Join-Path $repoRoot "configs/challenge4/shard$Shard.toml"
}

if (-not $StatePath) {
    $StatePath = Join-Path $repoRoot "state/challenge4/forwarder-blind-shard$Shard.toml"
}

if (-not $WorkspacePath) {
    $WorkspacePath = Join-Path $repoRoot "contracts/forwarder-blind/dex-interactor"
}

if (-not (Test-Path $ConfigPath)) {
    throw "Challenge 4 config file not found: $ConfigPath"
}

if (-not (Test-Path $WorkspacePath)) {
    throw "Challenge 4 interactor workspace not found: $WorkspacePath"
}

$stateDir = Split-Path -Parent $StatePath
if ($stateDir -and -not (Test-Path $stateDir)) {
    New-Item -ItemType Directory -Path $stateDir -Force | Out-Null
}

$resolvedConfigPath = (Resolve-Path $ConfigPath).Path
$resolvedWorkspacePath = (Resolve-Path $WorkspacePath).Path
$resolvedStatePath = [System.IO.Path]::GetFullPath($StatePath)

$configContent = Get-Content $resolvedConfigPath -Raw
$walletPemMatch = [regex]::Match($configContent, "(?m)^\s*wallet_pem\s*=\s*'([^']+)'\s*$")
if (-not $walletPemMatch.Success) {
    throw "wallet_pem not found in Challenge 4 config: $resolvedConfigPath"
}

$walletPemPath = $walletPemMatch.Groups[1].Value
if (-not (Test-Path $walletPemPath)) {
    throw "wallet_pem path from config does not exist: $walletPemPath"
}

$env:FORWARDER_BLIND_CONFIG = $resolvedConfigPath
$env:FORWARDER_BLIND_STATE = $resolvedStatePath
$env:FORWARDER_BLIND_WORKSPACE = $resolvedWorkspacePath

if (-not $InteractorArgs -or $InteractorArgs.Count -eq 0) {
    throw "No interactor command provided. Example: -InteractorArgs @('deploy')"
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
