param(
    [Parameter(Mandatory = $true)]
    [string]$TreasuryPemPath,
    [string]$TreasuryAddress = "erd1n28dse0sej7m2rz02ceftr30596arzx2a0trcegl9p3ztdagjahqa9szcx",
    [string]$AllocationPath = "./configs/challenge4/wallet-allocation.challenge4-callers.json",
    [string]$TokenId = "WEGLD-bd4d79",
    [string]$SyncPerWallet = "150000000000000000",
    [string]$TeSwap1PerWallet = "220000000000000000",
    [string]$TeSwap2PerWallet = "0",
    [string]$Async1PerWallet = "100000000000000000",
    [string]$Async2PerWallet = "100000000000000000",
    [string]$GatewayUrl = "https://api.battleofnodes.com",
    [string]$ChainId = "B",
    [int]$GasLimit = 500000,
    [int]$GasPrice = 1000000000,
    [int]$BatchSize = 25,
    [int]$PollIntervalSeconds = 3,
    [int]$ConfirmTimeoutSeconds = 180,
    [int]$HttpTimeoutSeconds = 25,
    [switch]$DryRun
)

$repoRoot = $PSScriptRoot
$resolvedAllocation = [System.IO.Path]::GetFullPath((Join-Path $repoRoot $AllocationPath))
if (-not (Test-Path $resolvedAllocation)) {
    throw "Allocation file not found: $resolvedAllocation"
}

$prepStartedAt = Get-Date

function Format-Elapsed {
    param(
        [TimeSpan]$Elapsed
    )

    return $Elapsed.ToString("hh\:mm\:ss")
}

function Write-StepBanner {
    param(
        [string]$Message
    )

    $elapsed = (Get-Date) - $prepStartedAt
    Write-Host ("[{0:o}] {1} (elapsed={2})" -f (Get-Date).ToUniversalTime(), $Message, (Format-Elapsed -Elapsed $elapsed))
}

function Invoke-PoolDistribution {
    param(
        [string]$Pool,
        [string]$AmountBase
    )

    if ([decimal]$AmountBase -le 0) {
        Write-Host ("skip pool={0} amount={1}" -f $Pool, $AmountBase)
        return
    }

    Write-StepBanner -Message ("START pool={0} amountPerWalletBase={1}" -f $Pool, $AmountBase)
    $poolStartedAt = Get-Date

    $args = @(
        "./cmd/distributeesdt",
        "-allocation", $resolvedAllocation,
        "-pool", $Pool,
        "-token-id", $TokenId,
        "-amount-base", $AmountBase,
        "-treasury-address", $TreasuryAddress,
        "-treasury-pem", $TreasuryPemPath,
        "-gateway", $GatewayUrl,
        "-chain-id", $ChainId,
        "-gas-limit", "$GasLimit",
        "-gas-price", "$GasPrice",
        "-batch-size", "$BatchSize",
        "-poll-interval-seconds", "$PollIntervalSeconds",
        "-confirm-timeout-seconds", "$ConfirmTimeoutSeconds",
        "-http-timeout-seconds", "$HttpTimeoutSeconds"
    )
    if ($DryRun) {
        $args += "-dry-run"
    }

    go run @args
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }

    $poolElapsed = (Get-Date) - $poolStartedAt
    Write-StepBanner -Message ("DONE  pool={0} in {1}" -f $Pool, (Format-Elapsed -Elapsed $poolElapsed))
}

Write-StepBanner -Message "Challenge 4 treasury WEGLD prep"
Invoke-PoolDistribution -Pool "shard1_sync" -AmountBase $SyncPerWallet
Invoke-PoolDistribution -Pool "shard1_te_swap1_bias" -AmountBase $TeSwap1PerWallet
Invoke-PoolDistribution -Pool "shard1_te_swap2_bias" -AmountBase $TeSwap2PerWallet
Invoke-PoolDistribution -Pool "shard0_async1" -AmountBase $Async1PerWallet
Invoke-PoolDistribution -Pool "shard2_async2" -AmountBase $Async2PerWallet

exit 0
