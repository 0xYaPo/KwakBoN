param(
    [string]$TreasuryAddress = "erd1n28dse0sej7m2rz02ceftr30596arzx2a0trcegl9p3ztdagjahqa9szcx",
    [string]$GatewayUrl = "https://api.battleofnodes.com",
    [string]$ChainId = "B",
    [string]$ManifestPath = "./configs/wallets-manifest.json",
    [int]$ShardFilter = -1,
    [string]$IncludeStatuses = "active_candidate,warm_reserve,window_b_reserve",
    [string]$ExcludeStatuses = "treasury,receiver",
    [string]$RequiredTags = "",
    [string]$MinRemainEgld = "0",
    [int]$GasLimit = 50000,
    [int]$GasPrice = 1000000000,
    [int]$PollIntervalSeconds = 3,
    [int]$HttpTimeoutSeconds = 25,
    [switch]$DryRun
)

$env:GATEWAY_URL = $GatewayUrl
$env:CHAIN_ID = $ChainId
$env:TX_VERSION = "2"
$env:WALLETS_MANIFEST = $ManifestPath
$env:TREASURY_ADDRESS = $TreasuryAddress
$env:SWEEP_SHARD_FILTER = "$ShardFilter"
$env:SWEEP_INCLUDE_STATUSES = $IncludeStatuses
$env:SWEEP_EXCLUDE_STATUSES = $ExcludeStatuses
$env:SWEEP_REQUIRED_TAGS = $RequiredTags
$env:SWEEP_MIN_REMAIN_EGLD = $MinRemainEgld
$env:SWEEP_GAS_LIMIT = "$GasLimit"
$env:GAS_PRICE = "$GasPrice"
$env:WAIT_CONFIRM = "true"
$env:POLL_INTERVAL_SECONDS = "$PollIntervalSeconds"
$env:HTTP_TIMEOUT_SECONDS = "$HttpTimeoutSeconds"

if ($DryRun) {
    go run ./cmd/sweepwallets --dry-run
    exit $LASTEXITCODE
}

go run ./cmd/sweepwallets
exit $LASTEXITCODE
