param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("shard1_sync", "shard1_te_swap1_bias", "shard1_te_swap2_bias", "shard0_async1", "shard2_async2")]
    [string]$Pool,
    [Parameter(Mandatory = $true)]
    [string]$Amount,
    [string]$AllocationPath = "./configs/challenge4/wallet-allocation.challenge4-callers.json",
    [int]$StartIndex = 0,
    [int]$Limit = 0,
    [switch]$WhatIf
)

$repoRoot = $PSScriptRoot
$resolvedAllocation = [System.IO.Path]::GetFullPath((Join-Path $repoRoot $AllocationPath))
if (-not (Test-Path $resolvedAllocation)) {
    throw "Allocation file not found: $resolvedAllocation"
}

$allocation = Get-Content $resolvedAllocation -Raw | ConvertFrom-Json
$items = @($allocation.pools.$Pool)
if ($StartIndex -lt 0) {
    throw "StartIndex must be >= 0"
}
if ($StartIndex -ge $items.Count) {
    throw "StartIndex $StartIndex is outside pool $Pool size $($items.Count)"
}

$slice = @($items | Select-Object -Skip $StartIndex)
if ($Limit -gt 0) {
    $slice = @($slice | Select-Object -First $Limit)
}

$wrapScript = Join-Path $repoRoot "run-challenge4-wrap-caller.ps1"

foreach ($item in $slice) {
    $cmd = ".\\run-challenge4-wrap-caller.ps1 -WalletId $($item.walletId) -Amount $Amount"
    Write-Host ("walletId={0} wrapAmount={1}" -f $item.walletId, $Amount)
    if ($WhatIf) {
        Write-Host $cmd
        continue
    }

    & $wrapScript -WalletId $item.walletId -Amount $Amount
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

exit 0
