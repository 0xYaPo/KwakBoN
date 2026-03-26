param(
    [string]$LaunchAtLocal = "",
    [string]$LaunchAtUtc = "",
    [string]$GatewayUrl = "https://api.battleofnodes.com",
    [string]$ChainId = "B",
    [string]$ManifestPath = "./configs/wallets-manifest.supernova-balanced-250.json",
    [string]$SenderStatuses = "supernova_balanced_active",
    [string]$RequiredTags = "sender",
    [int]$ShardFilter = -1,
    [int]$TargetTx = 9999999,
    [int]$DurationSeconds = 1800,
    [int]$BatchSize = 25,
    [int]$MaxNonceLookahead = 50,
    [int]$MaxConcurrentReads = 24,
    [int]$MaxActiveWallets = 0,
    [int]$IdleSleepMs = 300,
    [bool]$WaitConfirm = $false,
    [int]$ConfirmWorkers = 32,
    [int]$PollIntervalSeconds = 3,
    [int]$ConfirmTimeoutSeconds = 240,
    [string]$BulkValue = "50000000000000",
    [int]$GasLimit = 50000,
    [int]$GasPrice = 1000000000,
    [int]$HttpTimeoutSeconds = 12,
    [switch]$DryRun
)

if ([string]::IsNullOrWhiteSpace($LaunchAtLocal) -and [string]::IsNullOrWhiteSpace($LaunchAtUtc)) {
    throw "Set either -LaunchAtLocal 'YYYY-MM-DD HH:MM:SS' or -LaunchAtUtc 'YYYY-MM-DDTHH:MM:SSZ'."
}

if (-not [string]::IsNullOrWhiteSpace($LaunchAtLocal) -and -not [string]::IsNullOrWhiteSpace($LaunchAtUtc)) {
    throw "Use only one of -LaunchAtLocal or -LaunchAtUtc."
}

if (-not [string]::IsNullOrWhiteSpace($LaunchAtLocal)) {
    $target = Get-Date $LaunchAtLocal
    $targetLabel = $target.ToString("yyyy-MM-dd HH:mm:ss")
}
else {
    $target = [datetime]::Parse($LaunchAtUtc).ToUniversalTime().ToLocalTime()
    $targetLabel = "$LaunchAtUtc UTC"
}

$delay = $target - (Get-Date)
if ($delay.TotalMilliseconds -gt 0) {
    Write-Host ("[run-bulk-at-time] waiting until {0} (sleep {1:n1}s)" -f $targetLabel, $delay.TotalSeconds)
    Start-Sleep -Milliseconds ([int][math]::Floor($delay.TotalMilliseconds))
}
else {
    Write-Host ("[run-bulk-at-time] target time already passed by {0:n1}s, starting immediately" -f (-1 * $delay.TotalSeconds))
}

$env:DEFAULT_NETWORK = "battle"
$env:GATEWAY_URL = $GatewayUrl
$env:CHAIN_ID = $ChainId
$env:TX_VERSION = "2"
$env:WALLETS_MANIFEST = $ManifestPath
$env:BULK_SENDER_STATUSES = $SenderStatuses
$env:BULK_REQUIRED_TAGS = $RequiredTags
$env:BULK_SHARD_FILTER = "$ShardFilter"
$env:BULK_TARGET_TX = "$TargetTx"
$env:BULK_DURATION_SECONDS = "$DurationSeconds"
$env:BULK_BATCH_SIZE = "$BatchSize"
$env:BULK_MAX_NONCE_LOOKAHEAD = "$MaxNonceLookahead"
$env:BULK_MAX_CONCURRENT_READS = "$MaxConcurrentReads"
$env:BULK_MAX_ACTIVE_WALLETS = "$MaxActiveWallets"
$env:BULK_IDLE_SLEEP_MS = "$IdleSleepMs"
$env:BULK_WAIT_CONFIRM = "$WaitConfirm"
$env:BULK_CONFIRM_WORKERS = "$ConfirmWorkers"
$env:BULK_POLL_INTERVAL_SECONDS = "$PollIntervalSeconds"
$env:BULK_CONFIRM_TIMEOUT_SECONDS = "$ConfirmTimeoutSeconds"
$env:BULK_VALUE = $BulkValue
$env:GAS_LIMIT = "$GasLimit"
$env:GAS_PRICE = "$GasPrice"
$env:HTTP_TIMEOUT_SECONDS = "$HttpTimeoutSeconds"

if ($DryRun) {
    go run ./cmd/bulksprint --dry-run
    exit $LASTEXITCODE
}

go run ./cmd/bulksprint
exit $LASTEXITCODE
