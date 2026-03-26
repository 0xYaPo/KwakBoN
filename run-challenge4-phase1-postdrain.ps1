param(
    [string]$Swap1Amount = "10000000000000000",
    [string]$Swap2Amount = "40000",
    [string]$MinAmount = "1",
    [int]$LimitSync = 0,
    [int]$LimitTeSwap1 = 0,
    [int]$LimitTeSwap2 = 0,
    [int]$LimitAsync1 = 0,
    [int]$LimitAsync2 = 0,
    [int]$PoolMaxParallel = 4,
    [switch]$ParallelPools,
    [switch]$WhatIf
)

$runner = Join-Path $PSScriptRoot "run-challenge4-phase2.ps1"
& $runner `
    -Swap1Amount $Swap1Amount `
    -Swap2Amount $Swap2Amount `
    -MinAmount $MinAmount `
    -LimitSync $LimitSync `
    -LimitTeSwap1 $LimitTeSwap1 `
    -LimitTeSwap2 $LimitTeSwap2 `
    -LimitAsync1 $LimitAsync1 `
    -LimitAsync2 $LimitAsync2 `
    -PoolMaxParallel $PoolMaxParallel `
    -ParallelPools:$ParallelPools `
    -WhatIf:$WhatIf

exit $LASTEXITCODE
