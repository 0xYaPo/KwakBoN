param(
    [Parameter(Mandatory = $true)]
    [string]$ManifestPath,
    [Parameter(Mandatory = $true)]
    [string]$SenderStatuses,
    [string]$RequiredSenderTags = "sender,challenge3",
    [string]$GatewayUrl = "https://api.battleofnodes.com",
    [string]$ChainId = "B",
    [int]$ShardFilter = -1,
    [int]$TargetTx = 100000,
    [int]$DurationSeconds = 0,
    [int]$SustainedTps = 700,
    [int]$Workers = 64,
    [int]$ConfirmWorkers = 32,
    [string]$TxValue = "1",
    [string]$RoutingMode = "cross-shard",
    [bool]$ReuseSendersAsReceivers = $true,
    [bool]$ContinueOnTransientSendError = $true,
    [int]$SuccessThreshold = 0,
    [int]$ConfirmTimeoutSeconds = 240,
    [int]$PollIntervalSeconds = 3,
    [int]$HttpTimeoutSeconds = 25,
    [int]$GasLimit = 50000,
    [int]$GasPrice = 1000000000,
    [switch]$DryRun
)

$env:DEFAULT_NETWORK = "battle"
$env:GATEWAY_URL = $GatewayUrl
$env:CHAIN_ID = $ChainId
$env:TX_VERSION = "2"
$env:WALLETS_MANIFEST = $ManifestPath
$env:SPRINT_SHARD_FILTER = "$ShardFilter"
$env:SPRINT_SENDER_STATUSES = $SenderStatuses
$env:SPRINT_REQUIRED_SENDER_TAGS = $RequiredSenderTags
$env:SPRINT_RECEIVER_STATUSES = ""
$env:SPRINT_REQUIRED_RECEIVER_TAGS = ""
$env:SPRINT_RECEIVER_ADDRESSES = ""
$env:SPRINT_DURATION_SECONDS = "$DurationSeconds"
$env:CONTINUE_ON_TRANSIENT_SEND_ERROR = "$ContinueOnTransientSendError"
$env:SPRINT_TARGET_TX = "$TargetTx"
$env:SUSTAINED_TPS = "$SustainedTps"
$env:SPRINT_WORKERS = "$Workers"
$env:SPRINT_CONFIRM_WORKERS = "$ConfirmWorkers"
$env:SPRINT_TX_VALUE = $TxValue
$env:SPRINT_ROUTING_MODE = $RoutingMode
$env:SPRINT_REUSE_SENDERS_AS_RECEIVERS = "$ReuseSendersAsReceivers"
$env:INCLUDE_TREASURY_RECEIVER = "false"
if ($SuccessThreshold -le 0) {
    $SuccessThreshold = $TargetTx
}
$env:CONFIRM_SUCCESS_TARGET = "$SuccessThreshold"
$env:CONFIRM_TIMEOUT_SECONDS = "$ConfirmTimeoutSeconds"
$env:WAIT_CONFIRM = "true"
$env:POLL_INTERVAL_SECONDS = "$PollIntervalSeconds"
$env:HTTP_TIMEOUT_SECONDS = "$HttpTimeoutSeconds"
$env:GAS_LIMIT = "$GasLimit"
$env:GAS_PRICE = "$GasPrice"

if ($DryRun) {
    go run ./cmd/windowsprint --dry-run
    exit $LASTEXITCODE
}

go run ./cmd/windowsprint
exit $LASTEXITCODE
