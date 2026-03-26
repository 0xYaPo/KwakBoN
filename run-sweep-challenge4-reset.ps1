param(
    [string]$TreasuryAddress = "erd1n28dse0sej7m2rz02ceftr30596arzx2a0trcegl9p3ztdagjahqa9szcx",
    [string]$GatewayUrl = "https://api.battleofnodes.com",
    [string]$ChainId = "B",
    [string]$CallersManifestPath = "./configs/challenge4/wallets-manifest.challenge4-callers.json",
    [string]$OperatorsManifestPath = "./configs/challenge4/wallets-manifest.challenge4-operators.json",
    [string]$WegldTokenId = "WEGLD-bd4d79",
    [string]$UsdcTokenId = "USDC-c76f1f",
    [int]$ShardFilter = -1,
    [int]$SweepGasLimit = 50000,
    [int]$EsdtGasLimit = 500000,
    [int]$GasPrice = 1000000000,
    [int]$PollIntervalSeconds = 3,
    [int]$ConfirmTimeoutSeconds = 180,
    [int]$ConfirmWorkers = 32,
    [int]$HttpTimeoutSeconds = 25,
    [string]$MinRemainEgld = "0",
    [switch]$SkipDrain,
    [switch]$DryRun
)

$repoRoot = $PSScriptRoot

function Invoke-Go {
    param(
        [string[]]$CommandArgs
    )

    & go @CommandArgs
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

function Invoke-EsdtSweep {
    param(
        [string]$ManifestPath,
        [string]$IncludeStatuses,
        [string]$RequiredTags,
        [string]$TokenId
    )

    $cmdArgs = @(
        "run", "./cmd/sweepesdtwallets",
        "-manifest", $ManifestPath,
        "-treasury-address", $TreasuryAddress,
        "-token-id", $TokenId,
        "-gateway", $GatewayUrl,
        "-chain-id", $ChainId,
        "-gas-limit", "$EsdtGasLimit",
        "-gas-price", "$GasPrice",
        "-poll-interval-seconds", "$PollIntervalSeconds",
        "-confirm-timeout-seconds", "$ConfirmTimeoutSeconds",
        "-confirm-workers", "$ConfirmWorkers",
        "-http-timeout-seconds", "$HttpTimeoutSeconds",
        "-shard-filter", "$ShardFilter",
        "-include-statuses", $IncludeStatuses,
        "-exclude-statuses", "treasury,receiver",
        "-required-tags", $RequiredTags
    )
    if ($DryRun) {
        $cmdArgs += "--dry-run"
    }
    Invoke-Go -CommandArgs $cmdArgs
}

function Invoke-EgldSweep {
    param(
        [string]$ManifestPath,
        [string]$IncludeStatuses,
        [string]$RequiredTags
    )

    $env:GATEWAY_URL = $GatewayUrl
    $env:CHAIN_ID = $ChainId
    $env:TX_VERSION = "2"
    $env:WALLETS_MANIFEST = $ManifestPath
    $env:TREASURY_ADDRESS = $TreasuryAddress
    $env:SWEEP_SHARD_FILTER = "$ShardFilter"
    $env:SWEEP_INCLUDE_STATUSES = $IncludeStatuses
    $env:SWEEP_EXCLUDE_STATUSES = "treasury,receiver"
    $env:SWEEP_REQUIRED_TAGS = $RequiredTags
    $env:SWEEP_MIN_REMAIN_EGLD = $MinRemainEgld
    $env:SWEEP_GAS_LIMIT = "$SweepGasLimit"
    $env:GAS_PRICE = "$GasPrice"
    $env:WAIT_CONFIRM = "true"
    $env:POLL_INTERVAL_SECONDS = "$PollIntervalSeconds"
    $env:SWEEP_CONFIRM_TIMEOUT_SECONDS = "$ConfirmTimeoutSeconds"
    $env:SWEEP_CONFIRM_WORKERS = "$ConfirmWorkers"
    $env:HTTP_TIMEOUT_SECONDS = "$HttpTimeoutSeconds"

    $cmdArgs = @("run", "./cmd/sweepwallets")
    if ($DryRun) {
        $cmdArgs += "--dry-run"
    }
    Invoke-Go -CommandArgs $cmdArgs
}

if (-not $SkipDrain) {
    Write-Host "Challenge 4 reset: drain forwarders"
    & (Join-Path $repoRoot "run-challenge4-drain.ps1") -Shard 0
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    & (Join-Path $repoRoot "run-challenge4-drain.ps1") -Shard 1
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    & (Join-Path $repoRoot "run-challenge4-drain.ps1") -Shard 2
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

Write-Host "Challenge 4 reset: sweep WEGLD from callers"
Invoke-EsdtSweep -ManifestPath $CallersManifestPath -IncludeStatuses "challenge4_caller" -RequiredTags "caller,challenge4" -TokenId $WegldTokenId

Write-Host "Challenge 4 reset: sweep USDC from callers"
Invoke-EsdtSweep -ManifestPath $CallersManifestPath -IncludeStatuses "challenge4_caller" -RequiredTags "caller,challenge4" -TokenId $UsdcTokenId

Write-Host "Challenge 4 reset: sweep WEGLD from operators"
Invoke-EsdtSweep -ManifestPath $OperatorsManifestPath -IncludeStatuses "challenge4_operator" -RequiredTags "operator,challenge4" -TokenId $WegldTokenId

Write-Host "Challenge 4 reset: sweep USDC from operators"
Invoke-EsdtSweep -ManifestPath $OperatorsManifestPath -IncludeStatuses "challenge4_operator" -RequiredTags "operator,challenge4" -TokenId $UsdcTokenId

Write-Host "Challenge 4 reset: sweep EGLD from callers"
Invoke-EgldSweep -ManifestPath $CallersManifestPath -IncludeStatuses "challenge4_caller" -RequiredTags "caller,challenge4"

Write-Host "Challenge 4 reset: sweep EGLD from operators"
Invoke-EgldSweep -ManifestPath $OperatorsManifestPath -IncludeStatuses "challenge4_operator" -RequiredTags "operator,challenge4"
