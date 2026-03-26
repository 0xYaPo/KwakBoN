param(
    [string]$Swap1Amount = "10000000000000000",
    [string]$Swap2Amount = "10000",
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

$repoRoot = $PSScriptRoot
$poolRunner = Join-Path $repoRoot "run-challenge4-pool.ps1"
$drainRunner = Join-Path $repoRoot "run-challenge4-drain.ps1"

function Invoke-Pool {
    param(
        [string]$Pool,
        [string]$ForwarderShard,
        [string]$Action,
        [string]$Method,
        [string]$Amount,
        [int]$Limit
    )

    $args = @{
        Pool = $Pool
        ForwarderShard = $ForwarderShard
        Action = $Action
        Method = $Method
        Amount = $Amount
        MinAmount = $MinAmount
        MaxParallel = $PoolMaxParallel
    }
    if ($Limit -gt 0) {
        $args["Limit"] = $Limit
    }
    if ($WhatIf) {
        $args["WhatIf"] = $true
    }

    & $poolRunner @args
    if (-not $WhatIf -and $LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

Write-Host "Phase 2: scale stable lanes and opportunistic TE"
if ($ParallelPools -and -not $WhatIf) {
    $poolJobs = [System.Collections.ArrayList]::new()
    try {
        $poolSpecs = @(
            @{ Pool = "shard1_sync"; ForwarderShard = "1"; Action = "swap1"; Method = "sync"; Amount = $Swap1Amount; Limit = $LimitSync },
            @{ Pool = "shard0_async1"; ForwarderShard = "0"; Action = "swap1"; Method = "async1"; Amount = $Swap1Amount; Limit = $LimitAsync1 },
            @{ Pool = "shard2_async2"; ForwarderShard = "2"; Action = "swap1"; Method = "async2"; Amount = $Swap1Amount; Limit = $LimitAsync2 }
        )
        if (($LimitTeSwap1 -gt 0) -or ($LimitTeSwap2 -gt 0) -or (-not $WhatIf)) {
            $poolSpecs += @{ Pool = "shard1_te_swap1_bias"; ForwarderShard = "1"; Action = "swap1"; Method = "te"; Amount = $Swap1Amount; Limit = $LimitTeSwap1 }
            $poolSpecs += @{ Pool = "shard1_te_swap2_bias"; ForwarderShard = "1"; Action = "swap2"; Method = "te"; Amount = $Swap2Amount; Limit = $LimitTeSwap2 }
        }

        foreach ($spec in $poolSpecs) {
            $job = Start-Job -Name $spec.Pool -ScriptBlock {
                param($scriptPath, $argsHash)
                & $scriptPath @argsHash
                exit $LASTEXITCODE
            } -ArgumentList $poolRunner, @{
                Pool = $spec.Pool
                ForwarderShard = $spec.ForwarderShard
                Action = $spec.Action
                Method = $spec.Method
                Amount = $spec.Amount
                MinAmount = $MinAmount
                MaxParallel = $PoolMaxParallel
                Limit = $spec.Limit
            }
            [void]$poolJobs.Add($job)
        }

        foreach ($job in @($poolJobs)) {
            Wait-Job -Job $job | Out-Null
            Receive-Job -Job $job | Write-Host
            if ($job.State -ne "Completed" -or $job.ChildJobs[0].JobStateInfo.Reason) {
                $reason = $job.ChildJobs[0].JobStateInfo.Reason
                throw "Phase 2 pool failed: $($job.Name): $reason"
            }
            if ($job.ChildJobs[0].Output.Count -gt 0) {
                $lastOutput = $job.ChildJobs[0].Output[-1]
                if ($lastOutput -is [pscustomobject] -and $lastOutput.PSObject.Properties.Name -contains "ExitCode" -and $lastOutput.ExitCode -ne 0) {
                    throw "Phase 2 pool failed: $($job.Name) exitCode=$lastOutput"
                }
            }
            Remove-Job -Job $job -Force
        }
    }
    finally {
        foreach ($job in @($poolJobs)) {
            Remove-Job -Job $job -Force -ErrorAction SilentlyContinue
        }
    }
} else {
    Invoke-Pool -Pool "shard1_sync" -ForwarderShard "1" -Action "swap1" -Method "sync" -Amount $Swap1Amount -Limit $LimitSync
    Invoke-Pool -Pool "shard0_async1" -ForwarderShard "0" -Action "swap1" -Method "async1" -Amount $Swap1Amount -Limit $LimitAsync1
    Invoke-Pool -Pool "shard2_async2" -ForwarderShard "2" -Action "swap1" -Method "async2" -Amount $Swap1Amount -Limit $LimitAsync2

    if (($LimitTeSwap1 -gt 0) -or ($LimitTeSwap2 -gt 0) -or (-not $WhatIf)) {
        Invoke-Pool -Pool "shard1_te_swap1_bias" -ForwarderShard "1" -Action "swap1" -Method "te" -Amount $Swap1Amount -Limit $LimitTeSwap1
        Invoke-Pool -Pool "shard1_te_swap2_bias" -ForwarderShard "1" -Action "swap2" -Method "te" -Amount $Swap2Amount -Limit $LimitTeSwap2
    }
}

if ($WhatIf) {
    Write-Host "WhatIf mode: skipping drain commands"
    exit 0
}

Write-Host "Phase 2 drain cycle"
& $drainRunner -Shard 0
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
& $drainRunner -Shard 1
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
& $drainRunner -Shard 2
exit $LASTEXITCODE
