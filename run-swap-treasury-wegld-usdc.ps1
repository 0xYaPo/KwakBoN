param(
    [Parameter(Mandatory = $true)]
    [string]$TreasuryPemPath,
    [string]$TreasuryAddress = "erd1n28dse0sej7m2rz02ceftr30596arzx2a0trcegl9p3ztdagjahqa9szcx",
    [ValidateSet(0, 1, 2)]
    [int]$TreasuryShard = 2,
    [string]$ValidatorRepoPath = "C:\Users\y.pochon\Dev\Mvx\KwakBoN-Validators",
    [string]$GatewayUrl = "https://api.battleofnodes.com",
    [string]$ChainId = "B",
    [string]$PoolAddress = "erd1qqqqqqqqqqqqqpgqeel2kumf0r8ffyhth7pqdujjat9nx0862jpsg2pqaq",
    [string]$TokenInId = "WEGLD-bd4d79",
    [int]$TokenInDecimals = 18,
    [string]$TokenInAmount = "50",
    [string]$TokenOutId = "USDC-c76f1f",
    [string]$MinOutBase = "1",
    [int]$GasLimit = 15000000,
    [int]$GasPrice = 1000000000,
    [int]$ConfirmTimeoutSeconds = 240,
    [int]$PollIntervalSeconds = 3,
    [int]$HttpTimeoutSeconds = 12,
    [switch]$DryRun
)

if (-not (Test-Path $ValidatorRepoPath)) {
    throw "Validator repo path not found: $ValidatorRepoPath"
}

$tmpDir = Join-Path (Resolve-Path .).Path "tmp-test"
New-Item -ItemType Directory -Force -Path $tmpDir | Out-Null
$manifestPath = Join-Path $tmpDir "treasury-single-swap-manifest.json"

$manifest = [PSCustomObject]@{
    version = 1
    generatedAt = (Get-Date).ToUniversalTime().ToString("s") + "Z"
    wallets = @(
        [PSCustomObject]@{
            walletId = "treasury-swap"
            address = $TreasuryAddress
            shard = $TreasuryShard
            pemPath = $TreasuryPemPath
            status = "validator_active"
            tags = @("sender", "validator_track", "challenge6")
            enabled = $true
            notes = "Temporary one-sender manifest for treasury WEGLD->USDC swap."
        }
    )
}

$manifest | ConvertTo-Json -Depth 6 | Set-Content -Path $manifestPath

$args = @(
    "-ExecutionPolicy", "Bypass",
    "-File", (Join-Path $ValidatorRepoPath "run-vladatos-window-b.ps1"),
    "-GatewayUrl", $GatewayUrl,
    "-ChainId", $ChainId,
    "-ManifestPath", $manifestPath,
    "-SenderStatuses", "validator_active",
    "-RequiredSenderTags", "sender,validator_track,challenge6",
    "-PoolAddress", $PoolAddress,
    "-TokenInId", $TokenInId,
    "-TokenInDecimals", "$TokenInDecimals",
    "-TokenInAmount", $TokenInAmount,
    "-TokenOutId", $TokenOutId,
    "-MinOutBase", $MinOutBase,
    "-TargetTx", "1",
    "-SustainedTps", "1",
    "-GasLimit", "$GasLimit",
    "-GasPrice", "$GasPrice",
    "-WaitConfirm", "true",
    "-ConfirmSuccessTarget", "1",
    "-ConfirmTimeoutSeconds", "$ConfirmTimeoutSeconds",
    "-PollIntervalSeconds", "$PollIntervalSeconds",
    "-HttpTimeoutSeconds", "$HttpTimeoutSeconds"
)

if ($DryRun) {
    $args += "-DryRun"
}

Push-Location $ValidatorRepoPath
try {
    & powershell @args
    exit $LASTEXITCODE
}
finally {
    Pop-Location
}
