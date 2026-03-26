param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("shard1_sync", "shard1_te_swap1_bias", "shard1_te_swap2_bias", "shard0_async1", "shard2_async2")]
    [string]$Pool,
    [Parameter(Mandatory = $true)]
    [ValidateSet("0", "1", "2")]
    [string]$ForwarderShard,
    [Parameter(Mandatory = $true)]
    [ValidateSet("swap1", "swap2")]
    [string]$Action,
    [Parameter(Mandatory = $true)]
    [ValidateSet("direct", "sync", "async1", "async2", "te")]
    [string]$Method,
    [string]$Amount,
    [string]$MinAmount = "1",
    [string]$AllocationPath = "./configs/challenge4/wallet-allocation.challenge4-callers.json",
    [int]$StartIndex = 0,
    [int]$Limit = 0,
    [int]$MaxParallel = 1,
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
if ($MaxParallel -lt 1) {
    throw "MaxParallel must be >= 1"
}

$slice = @($items | Select-Object -Skip $StartIndex)
if ($Limit -gt 0) {
    $slice = @($slice | Select-Object -First $Limit)
}

$callerScript = Join-Path $repoRoot "run-challenge4-caller.ps1"

function Wait-ForJobSlot {
    param(
        [System.Collections.ArrayList]$Jobs,
        [int]$MaxParallelJobs
    )

    while ($Jobs.Count -ge $MaxParallelJobs) {
        $completed = Wait-Job -Job $Jobs -Any
        foreach ($job in @($completed)) {
            Receive-Job -Job $job | Write-Host
            if ($job.State -ne "Completed" -or $job.ChildJobs[0].JobStateInfo.Reason) {
                $reason = $job.ChildJobs[0].JobStateInfo.Reason
                throw "Pool job failed for walletId=$($job.Name): $reason"
            }
            if ($job.ChildJobs[0].Output.Count -gt 0) {
                $lastOutput = $job.ChildJobs[0].Output[-1]
                if ($lastOutput -is [pscustomobject] -and $lastOutput.PSObject.Properties.Name -contains "ExitCode" -and $lastOutput.ExitCode -ne 0) {
                    throw "Pool job failed for walletId=$($job.Name) exitCode=$lastOutput"
                }
            }
            [void]$Jobs.Remove($job)
            Remove-Job -Job $job -Force
        }
    }
}

foreach ($item in $slice) {
    $cmd = ".\\run-challenge4-caller.ps1 -WalletId $($item.walletId) -ForwarderShard $ForwarderShard -Action $Action -Method $Method -Amount $Amount -MinAmount $MinAmount"
    $pairShard = 1
    $lane = if ([int]$ForwarderShard -eq $pairShard) { "forwarder-same-shard" } else { "forwarder-cross-shard" }
    Write-Host ("walletId={0} callerShard={1} forwarderShard={2} pairShard={3} lane={4} action={5} method={6}" -f $item.walletId, $item.shard, $ForwarderShard, $pairShard, $lane, $Action, $Method)
    if ($WhatIf) {
        Write-Host $cmd
        continue
    }
}

if ($WhatIf) {
    exit 0
}

if ($MaxParallel -eq 1) {
    foreach ($item in $slice) {
        & $callerScript `
            -WalletId $item.walletId `
            -ForwarderShard $ForwarderShard `
            -Action $Action `
            -Method $Method `
            -Amount $Amount `
            -MinAmount $MinAmount

        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }
    exit 0
}

$jobs = [System.Collections.ArrayList]::new()
try {
    foreach ($item in $slice) {
        Wait-ForJobSlot -Jobs $jobs -MaxParallelJobs $MaxParallel
        $job = Start-Job -Name $item.walletId -ScriptBlock {
            param($scriptPath, $walletId, $forwarderShard, $action, $method, $amount, $minAmount)
            & $scriptPath `
                -WalletId $walletId `
                -ForwarderShard $forwarderShard `
                -Action $action `
                -Method $method `
                -Amount $amount `
                -MinAmount $minAmount
            [pscustomobject]@{ ExitCode = $LASTEXITCODE }
        } -ArgumentList $callerScript, $item.walletId, $ForwarderShard, $Action, $Method, $Amount, $MinAmount
        [void]$jobs.Add($job)
    }

    while ($jobs.Count -gt 0) {
        Wait-ForJobSlot -Jobs $jobs -MaxParallelJobs 1
    }
}
finally {
    foreach ($job in @($jobs)) {
        Remove-Job -Job $job -Force -ErrorAction SilentlyContinue
    }
}

exit 0
