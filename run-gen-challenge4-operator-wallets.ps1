param(
    [string]$ManifestPath = "./configs/challenge4/wallets-manifest.challenge4-operators.json",
    [string]$WalletsDir = "./wallets/challenge4/operators",
    [string]$Prefix = "c4op",
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
  -target-shard0 1 `
  -target-shard1 1 `
  -target-shard2-c3 1 `
  -status challenge4_operator `
  -extra-tags "operator,challenge4" `
  -notes "Challenge 4 operator wallets: one per shard."

if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

$manifest = Get-Content $resolvedManifest -Raw | ConvertFrom-Json
$wallets = @($manifest.wallets | Where-Object { $_.enabled })

foreach ($wallet in $wallets) {
    $configPath = Join-Path $PSScriptRoot ("configs/challenge4/shard{0}.toml" -f $wallet.shard)
    if (-not (Test-Path $configPath)) {
        throw "Missing shard config file: $configPath"
    }

    $walletPemPath = [System.IO.Path]::GetFullPath($wallet.pemPath)
    $updated = (Get-Content $configPath -Raw) -replace "(?m)^wallet_pem\s*=\s*'.*'$", ("wallet_pem = '{0}'" -f ($walletPemPath -replace '\\', '/'))
    Set-Content -LiteralPath $configPath -Value $updated -NoNewline
}

$wallets | Sort-Object shard | ForEach-Object {
    "shard={0} address={1} pem={2}" -f $_.shard, $_.address, $_.pemPath
}
