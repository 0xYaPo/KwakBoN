param(
    [Parameter(Mandatory = $true)]
    [string]$TreasuryPemPath,
    [string]$ManifestPath = "./configs/challenge4/wallets-manifest.challenge4-operators.json",
    [string]$IncludeStatuses = "challenge4_operator",
    [string]$TreasuryAddress = "erd1n28dse0sej7m2rz02ceftr30596arzx2a0trcegl9p3ztdagjahqa9szcx",
    [string]$RequiredTags = "operator,challenge4",
    [string]$TargetEgld = "10.0",
    [string]$GatewayUrl = "https://api.battleofnodes.com",
    [string]$ChainId = "B",
    [int]$ShardFilter = -1,
    [int]$GasLimit = 50000,
    [int]$GasPrice = 1000000000,
    [int]$PollIntervalSeconds = 3,
    [int]$ConfirmTimeoutSeconds = 180,
    [int]$HttpTimeoutSeconds = 25,
    [switch]$DryRun
)

$env:GATEWAY_URL = $GatewayUrl
$env:CHAIN_ID = $ChainId
$env:TX_VERSION = "2"
$env:WALLETS_MANIFEST = $ManifestPath
$env:TREASURY_ADDRESS = $TreasuryAddress
$env:TREASURY_PEM_PATH = $TreasuryPemPath
$env:FUND_SHARD_FILTER = "$ShardFilter"
$env:FUND_INCLUDE_STATUSES = $IncludeStatuses
$env:FUND_EXCLUDE_STATUSES = "treasury,receiver"
$env:FUND_REQUIRED_TAGS = $RequiredTags
$env:FUND_TARGET_CHALLENGE4_OPERATOR_EGLD = $TargetEgld
$env:FUND_GAS_LIMIT = "$GasLimit"
$env:GAS_PRICE = "$GasPrice"
$env:WAIT_CONFIRM = "true"
$env:POLL_INTERVAL_SECONDS = "$PollIntervalSeconds"
$env:FUND_CONFIRM_TIMEOUT_SECONDS = "$ConfirmTimeoutSeconds"
$env:HTTP_TIMEOUT_SECONDS = "$HttpTimeoutSeconds"

if ($DryRun) {
    go run ./cmd/fundwallets --dry-run
    exit $LASTEXITCODE
}

go run ./cmd/fundwallets
exit $LASTEXITCODE
