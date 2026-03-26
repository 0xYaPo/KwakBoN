param(
    [Parameter(Mandatory = $true)]
    [ValidateSet(0, 1, 2)]
    [int]$Shard,
    [Parameter(Mandatory = $true)]
    [string]$TreasuryPemPath,
    [string]$TargetEgld = "0.75",
    [string]$TreasuryAddress = "erd1n28dse0sej7m2rz02ceftr30596arzx2a0trcegl9p3ztdagjahqa9szcx",
    [string]$GatewayUrl = "https://api.battleofnodes.com",
    [string]$ChainId = "B",
    [string]$ManifestPath = "./configs/wallets-manifest.challenge3-part2.live.json",
    [int]$GasLimit = 50000,
    [int]$GasPrice = 1000000000,
    [int]$PollIntervalSeconds = 3,
    [int]$ConfirmTimeoutSeconds = 120,
    [int]$HttpTimeoutSeconds = 20,
    [switch]$DryRun
)

$env:GATEWAY_URL = $GatewayUrl
$env:CHAIN_ID = $ChainId
$env:TX_VERSION = "2"
$env:WALLETS_MANIFEST = $ManifestPath
$env:TREASURY_ADDRESS = $TreasuryAddress
$env:TREASURY_PEM_PATH = $TreasuryPemPath
$env:FUND_INCLUDE_STATUSES = "challenge3_part2_sender"
$env:FUND_EXCLUDE_STATUSES = "treasury,receiver"
$env:FUND_REQUIRED_TAGS = "sender,challenge3,part2"
$env:FUND_SHARD_FILTER = "$Shard"
$env:FUND_TARGET_ACTIVE_EGLD = $TargetEgld
$env:FUND_TARGET_RESERVE_EGLD = $TargetEgld
$env:FUND_TARGET_WINDOW_B_RESERVE_EGLD = $TargetEgld
$env:FUND_TARGET_CHALLENGE3_PART2_EGLD = $TargetEgld
$env:FUND_GAS_LIMIT = "$GasLimit"
$env:GAS_PRICE = "$GasPrice"
$env:WAIT_CONFIRM = "true"
$env:POLL_INTERVAL_SECONDS = "$PollIntervalSeconds"
$env:FUND_CONFIRM_TIMEOUT_SECONDS = "$ConfirmTimeoutSeconds"
$env:HTTP_TIMEOUT_SECONDS = "$HttpTimeoutSeconds"

Write-Host ("[challenge3-part2-rebalance] shard={0} target={1} EGLD manifest={2}" -f $Shard, $TargetEgld, $ManifestPath)

if ($DryRun) {
    go run ./cmd/fundwallets --dry-run
    exit $LASTEXITCODE
}

go run ./cmd/fundwallets
exit $LASTEXITCODE
