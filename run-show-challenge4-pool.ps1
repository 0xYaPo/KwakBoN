param(
    [string]$AllocationPath = "./configs/challenge4/wallet-allocation.challenge4-callers.json",
    [ValidateSet("shard1_sync", "shard1_te_swap1_bias", "shard1_te_swap2_bias", "shard0_async1", "shard2_async2", "all")]
    [string]$Pool = "all",
    [int]$Limit = 10
)

$resolvedAllocation = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot $AllocationPath))
if (-not (Test-Path $resolvedAllocation)) {
    throw "Allocation file not found: $resolvedAllocation"
}

$allocation = Get-Content $resolvedAllocation -Raw | ConvertFrom-Json
$poolNames = @($allocation.pools.PSObject.Properties.Name)

if ($Pool -eq "all") {
    foreach ($name in $poolNames) {
        $items = @($allocation.pools.$name)
        Write-Host ("pool={0} count={1}" -f $name, $items.Count)
    }
    exit 0
}

$items = @($allocation.pools.$Pool)
Write-Host ("pool={0} count={1}" -f $Pool, $items.Count)
$items | Select-Object -First $Limit | ForEach-Object {
    "walletId={0} shard={1} address={2} pem={3}" -f $_.walletId, $_.shard, $_.address, $_.pemPath
}
