param(
    [string]$SyncPerWallet = "150000000000000000",
    [string]$TeSwap1PerWallet = "220000000000000000",
    [string]$TeSwap2PerWallet = "0",
    [string]$Async1PerWallet = "100000000000000000",
    [string]$Async2PerWallet = "100000000000000000",
    [switch]$WhatIf
)

$repoRoot = $PSScriptRoot
$allocationPath = Join-Path $repoRoot "configs/challenge4/wallet-allocation.challenge4-callers.json"
if (-not (Test-Path $allocationPath)) {
    throw "Allocation file not found: $allocationPath"
}

$allocation = Get-Content $allocationPath -Raw | ConvertFrom-Json
$wrapPool = Join-Path $repoRoot "run-challenge4-wrap-pool.ps1"

function Invoke-WrapPool {
    param(
        [string]$Pool,
        [string]$Amount
    )

    if ([decimal]$Amount -le 0) {
        Write-Host ("skip pool={0} amount={1}" -f $Pool, $Amount)
        return
    }

    $args = @{
        Pool = $Pool
        Amount = $Amount
    }
    if ($WhatIf) {
        $args["WhatIf"] = $true
    }

    & $wrapPool @args
    if (-not $WhatIf -and $LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

function Show-Total {
    param(
        [string]$Pool,
        [string]$PerWallet
    )

    $count = @($allocation.pools.$Pool).Count
    $total = [System.Numerics.BigInteger]::Parse($PerWallet) * $count
    Write-Host ("pool={0} wallets={1} perWallet={2} totalBase={3}" -f $Pool, $count, $PerWallet, $total)
}

Write-Host "Challenge 4 wrap prep plan"
Show-Total -Pool "shard1_sync" -PerWallet $SyncPerWallet
Show-Total -Pool "shard1_te_swap1_bias" -PerWallet $TeSwap1PerWallet
Show-Total -Pool "shard1_te_swap2_bias" -PerWallet $TeSwap2PerWallet
Show-Total -Pool "shard0_async1" -PerWallet $Async1PerWallet
Show-Total -Pool "shard2_async2" -PerWallet $Async2PerWallet

Write-Host "Applying wraps"
Invoke-WrapPool -Pool "shard1_sync" -Amount $SyncPerWallet
Invoke-WrapPool -Pool "shard1_te_swap1_bias" -Amount $TeSwap1PerWallet
Invoke-WrapPool -Pool "shard1_te_swap2_bias" -Amount $TeSwap2PerWallet
Invoke-WrapPool -Pool "shard0_async1" -Amount $Async1PerWallet
Invoke-WrapPool -Pool "shard2_async2" -Amount $Async2PerWallet

exit 0
