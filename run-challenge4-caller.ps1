param(
    [Parameter(Mandatory = $true)]
    [string]$WalletId,
    [Parameter(Mandatory = $true)]
    [ValidateSet("0", "1", "2")]
    [string]$ForwarderShard,
    [Parameter(Mandatory = $true)]
    [ValidateSet("swap1", "swap2")]
    [string]$Action,
    [Parameter(Mandatory = $true)]
    [ValidateSet("direct", "sync", "async1", "async2", "te")]
    [string]$Method,
    [string]$Amount,
    [string]$MinAmount = "1",
    [string]$ManifestPath = "./configs/challenge4/wallets-manifest.challenge4-callers.json",
    [string]$TempDir = "./tmp/challenge4-caller-configs"
)

$repoRoot = $PSScriptRoot
$resolvedManifest = [System.IO.Path]::GetFullPath((Join-Path $repoRoot $ManifestPath))
if (-not (Test-Path $resolvedManifest)) {
    throw "Caller manifest not found: $resolvedManifest"
}

$manifest = Get-Content $resolvedManifest -Raw | ConvertFrom-Json
$wallet = @($manifest.wallets | Where-Object { $_.walletId -eq $WalletId -and $_.enabled }) | Select-Object -First 1
if (-not $wallet) {
    throw "Enabled caller wallet not found for walletId=$WalletId"
}

$baseConfigPath = Join-Path $repoRoot ("configs/challenge4/shard{0}.toml" -f $ForwarderShard)
if (-not (Test-Path $baseConfigPath)) {
    throw "Base shard config not found: $baseConfigPath"
}

$statePath = Join-Path $repoRoot ("state/challenge4/forwarder-blind-shard{0}.toml" -f $ForwarderShard)
if (-not (Test-Path $statePath)) {
    throw "Forwarder state file not found: $statePath"
}

$resolvedTempDir = [System.IO.Path]::GetFullPath((Join-Path $repoRoot $TempDir))
New-Item -ItemType Directory -Path $resolvedTempDir -Force | Out-Null

$tempConfigPath = Join-Path $resolvedTempDir ("{0}.forwarder{1}.toml" -f $WalletId, $ForwarderShard)
$walletPemPath = [System.IO.Path]::GetFullPath($wallet.pemPath)
$normalizedWalletPem = $walletPemPath -replace '\\', '/'
$lines = Get-Content $baseConfigPath
$updatedLines = foreach ($line in $lines) {
    if ($line -match "^\s*wallet_pem\s*=") {
        "wallet_pem = '$normalizedWalletPem'"
    } else {
        $line
    }
}
Set-Content -LiteralPath $tempConfigPath -Value $updatedLines

if (-not $Amount) {
    $Amount = if ($Action -eq "swap1") { "100000000000000000" } else { "10000" }
}

$runner = Join-Path $repoRoot "run-challenge4-interactor.ps1"

if ($Action -eq "swap1") {
    & $runner `
        -Shard $ForwarderShard `
        -ConfigPath $tempConfigPath `
        -StatePath $statePath `
        -InteractorArgs @("swap1", $Method, "-a", $Amount, "-m", $MinAmount)
    exit $LASTEXITCODE
}

& $runner `
    -Shard $ForwarderShard `
    -ConfigPath $tempConfigPath `
    -StatePath $statePath `
    -InteractorArgs @("swap2", $Method, "-a", $Amount, "-m", $MinAmount)
exit $LASTEXITCODE
