param(
    [string]$ManifestPath = "./configs/wallets-manifest.challenge3-part2.live.json",
    [string]$SenderStatuses = "challenge3_part2_sender",
    [string]$RequiredTags = "sender,challenge3,part2",
    [string]$LaunchAtLocal = "",
    [string]$LaunchAtUtc = "",
    [string]$GatewayUrl = "https://api.battleofnodes.com",
    [string]$ChainId = "B",
    [int]$ShardFilter = -1,
    [int]$TargetTx = 1000000,
    [int]$DurationSeconds = 1800,
    [int]$BatchSize = 12,
    [int]$MaxNonceLookahead = 15,
    [int]$MaxConcurrentReads = 24,
    [int]$MaxActiveWallets = 0,
    [int]$IdleSleepMs = 300,
    [bool]$WaitConfirm = $false,
    [int]$ConfirmWorkers = 32,
    [int]$PollIntervalSeconds = 3,
    [int]$ConfirmTimeoutSeconds = 240,
    [string]$BulkValue = "10000000000000000",
    [string]$RoutingMode = "cross-shard",
    [int]$GasLimit = 50000,
    [int]$GasPrice = 1000000000,
    [int]$HttpTimeoutSeconds = 12,
    [switch]$DryRun
)

if (-not [string]::IsNullOrWhiteSpace($LaunchAtLocal) -or -not [string]::IsNullOrWhiteSpace($LaunchAtUtc)) {
    if ([string]::IsNullOrWhiteSpace($LaunchAtLocal) -eq [string]::IsNullOrWhiteSpace($LaunchAtUtc)) {
        throw "Set only one of -LaunchAtLocal or -LaunchAtUtc."
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
        Write-Host ("[run-bulk-challenge3-part2] waiting until {0} (sleep {1:n1}s)" -f $targetLabel, $delay.TotalSeconds)
        Start-Sleep -Milliseconds ([int][math]::Floor($delay.TotalMilliseconds))
    }
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
$env:BULK_ROUTING_MODE = $RoutingMode
$env:GAS_LIMIT = "$GasLimit"
$env:GAS_PRICE = "$GasPrice"
$env:HTTP_TIMEOUT_SECONDS = "$HttpTimeoutSeconds"

if ($DryRun) {
    go run ./cmd/bulksprint --dry-run
    exit $LASTEXITCODE
}

go run ./cmd/bulksprint
exit $LASTEXITCODE
