param(
    [string]$ManifestPath = "./configs/challenge4/wallets-manifest.challenge4-callers.json",
    [string]$WalletsDir = "./wallets/challenge4/callers",
    [string]$Prefix = "c4call",
    [int]$Shard0Count = 33,
    [int]$Shard1Count = 34,
    [int]$Shard2Count = 33,
    [switch]$Force
)

$resolvedManifest = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot $ManifestPath))
$resolvedWalletsDir = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot $WalletsDir))

if ((Test-Path $resolvedManifest) -and -not $Force) {
    throw "Manifest already exists: $resolvedManifest. Use -Force only if you really want to replace it."
}

if ((Test-Path $resolvedWalletsDir) -and -not $Force) {
    throw "Wallet directory already exists: $resolvedWalletsDir. Use -Force only if you really want to replace it."
}

if ($Force) {
    if (Test-Path $resolvedManifest) {
        Remove-Item -LiteralPath $resolvedManifest -Force
    }
    if (Test-Path $resolvedWalletsDir) {
        Remove-Item -LiteralPath $resolvedWalletsDir -Recurse -Force
    }
}

New-Item -ItemType Directory -Path (Split-Path -Parent $resolvedManifest) -Force | Out-Null
New-Item -ItemType Directory -Path $resolvedWalletsDir -Force | Out-Null

go run ./cmd/genwallets `
  -mode challenge3 `
  -manifest $resolvedManifest `
  -wallets-dir $resolvedWalletsDir `
  -prefix $Prefix `
  -target-shard0 $Shard0Count `
  -target-shard1 $Shard1Count `
  -target-shard2-c3 $Shard2Count `
  -status challenge4_caller `
  -extra-tags "caller,challenge4" `
  -notes "Challenge 4 caller wallets."

exit $LASTEXITCODE
