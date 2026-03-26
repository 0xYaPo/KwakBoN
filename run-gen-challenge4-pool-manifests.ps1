param(
    [string]$AllocationPath = "./configs/challenge4/wallet-allocation.challenge4-callers.json",
    [string]$SourceManifestPath = "./configs/challenge4/wallets-manifest.challenge4-callers.json",
    [string]$OutputDir = "./configs/challenge4/pools"
)

$resolvedAllocation = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot $AllocationPath))
$resolvedSourceManifest = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot $SourceManifestPath))
$resolvedOutputDir = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot $OutputDir))

if (-not (Test-Path $resolvedAllocation)) {
    throw "Allocation file not found: $resolvedAllocation"
}
if (-not (Test-Path $resolvedSourceManifest)) {
    throw "Source manifest not found: $resolvedSourceManifest"
}

New-Item -ItemType Directory -Path $resolvedOutputDir -Force | Out-Null

$allocation = Get-Content $resolvedAllocation -Raw | ConvertFrom-Json
$sourceManifest = Get-Content $resolvedSourceManifest -Raw | ConvertFrom-Json
$walletIndex = @{}
foreach ($wallet in $sourceManifest.wallets) {
    $walletIndex[$wallet.walletId] = $wallet
}

foreach ($poolProp in $allocation.pools.PSObject.Properties) {
    $poolName = $poolProp.Name
    $poolItems = @($poolProp.Value)
    $wallets = foreach ($item in $poolItems) {
        if (-not $walletIndex.ContainsKey($item.walletId)) {
            throw "Wallet $($item.walletId) from allocation not found in source manifest"
        }
        $walletIndex[$item.walletId]
    }

    $poolManifest = [ordered]@{
        version = 1
        generatedAt = (Get-Date).ToUniversalTime().ToString("o")
        wallets = $wallets
    }

    $outPath = Join-Path $resolvedOutputDir ("wallets-manifest.challenge4.{0}.json" -f $poolName)
    $json = $poolManifest | ConvertTo-Json -Depth 6
    Set-Content -LiteralPath $outPath -Value $json
    Write-Host ("written={0} count={1}" -f $outPath, $wallets.Count)
}
