param(
    [string]$GatewayUrl = "https://api.battleofnodes.com",
    [string]$ChainId = "B",
    [string]$ManifestPath = "./configs/wallets-manifest.json",
    [string]$ReceiverAddresses = "",
    [int]$ShardFilter = 2,
    [int]$TargetTx = 1100000,
    [int]$SustainedTps = 700,
    [int]$Workers = 64,
    [int]$ConfirmWorkers = 32,
    [string]$TxValue = "1",
    [int]$GasLimit = 50000,
    [int]$GasPrice = 1000000000,
    [bool]$IncludeTreasuryReceiver = $true,
    [int]$ReceiverWeight = 4,
    [int]$TreasuryReceiverWeight = 1,
    [int]$SuccessThreshold = 1100000,
    [int]$ConfirmTimeoutSeconds = 0,
    [int]$PollIntervalSeconds = 3,
    [int]$HttpTimeoutSeconds = 25,
    [switch]$DryRun
)

$env:DEFAULT_NETWORK = "battle"
$env:GATEWAY_URL = $GatewayUrl
$env:CHAIN_ID = $ChainId
$env:TX_VERSION = "2"
$env:WALLETS_MANIFEST = $ManifestPath
$env:SPRINT_RECEIVER_ADDRESSES = $ReceiverAddresses
$env:SPRINT_SHARD_FILTER = "$ShardFilter"
$env:SPRINT_SENDER_STATUSES = "active_candidate"
$env:SPRINT_RECEIVER_STATUSES = "receiver"
$env:SPRINT_REQUIRED_SENDER_TAGS = "window_a,sender"
$env:SPRINT_REQUIRED_RECEIVER_TAGS = "window_a,sink"
$env:INCLUDE_TREASURY_RECEIVER = "$IncludeTreasuryReceiver"
$env:RECEIVER_WEIGHT = "$ReceiverWeight"
$env:TREASURY_RECEIVER_WEIGHT = "$TreasuryReceiverWeight"
$env:SPRINT_TARGET_TX = "$TargetTx"
$env:SUSTAINED_TPS = "$SustainedTps"
$env:SPRINT_WORKERS = "$Workers"
$env:SPRINT_CONFIRM_WORKERS = "$ConfirmWorkers"
$env:SPRINT_TX_VALUE = $TxValue
$env:GAS_LIMIT = "$GasLimit"
$env:GAS_PRICE = "$GasPrice"
$env:WAIT_CONFIRM = "true"
$env:CONFIRM_SUCCESS_TARGET = "$SuccessThreshold"
$env:CONFIRM_TIMEOUT_SECONDS = "$ConfirmTimeoutSeconds"
$env:POLL_INTERVAL_SECONDS = "$PollIntervalSeconds"
$env:HTTP_TIMEOUT_SECONDS = "$HttpTimeoutSeconds"

if ($DryRun) {
    go run ./cmd/windowsprint --dry-run
    exit $LASTEXITCODE
}

go run ./cmd/windowsprint
exit $LASTEXITCODE
