param(
    [Parameter(Mandatory = $true)]
    [string]$TreasuryPemPath,
    [string]$TreasuryAddress = "erd1n28dse0sej7m2rz02ceftr30596arzx2a0trcegl9p3ztdagjahqa9szcx",
    [string]$WegldWrapperAddress = "erd1qqqqqqqqqqqqqpgqmuk0q2saj0mgutxm4teywre6dl8wqf58xamqdrukln",
    [string]$WegldTokenId = "WEGLD-bd4d79",
    [string]$UsdcTokenId = "USDC-c76f1f",
    [string]$SyncWegldPerWallet = "150000000000000000",
    [string]$TeSwap1WegldPerWallet = "200000000000000000",
    [string]$Async1WegldPerWallet = "100000000000000000",
    [string]$Async2WegldPerWallet = "100000000000000000",
    [string]$TeSwap2UsdcPerWallet = "3000000",
    [string]$TeSwap2StarterWegldAmount = "5000000000000000000",
    [string]$TeSwap2StarterUsdcMin = "1",
    [string]$StarterUsdcSourcePemPath,
    [string]$StarterUsdcSourceAddress,
    [string]$GatewayUrl = "https://api.battleofnodes.com",
    [string]$ChainId = "B",
    [int]$GasLimit = 500000,
    [int]$GasPrice = 1000000000,
    [int]$BatchSize = 25,
    [int]$PollIntervalSeconds = 3,
    [int]$ConfirmTimeoutSeconds = 180,
    [int]$HttpTimeoutSeconds = 25,
    [string]$CallerTargetEgld = "0.5",
    [string]$OperatorTargetEgld = "5.0",
    [switch]$AutoWrapWegld,
    [switch]$SkipWegldBalanceCheck,
    [switch]$SkipCallerFunding,
    [switch]$SkipOperatorFunding,
    [switch]$DryRun
)

$repoRoot = $PSScriptRoot
$wegldPrep = Join-Path $repoRoot "run-challenge4-prep-wegld-treasury.ps1"
$fundCallers = Join-Path $repoRoot "run-fund-challenge4-callers.ps1"
$fundOperators = Join-Path $repoRoot "run-fund-challenge4-operators.ps1"
$wrapScript = Join-Path $repoRoot "run-challenge4-wrap.ps1"
$swap1Script = Join-Path $repoRoot "run-challenge4-swap1.ps1"
$allocationPath = Join-Path $repoRoot "configs/challenge4/wallet-allocation.challenge4-callers.json"
$operatorManifestPath = Join-Path $repoRoot "configs/challenge4/wallets-manifest.challenge4-operators.json"
$allocation = Get-Content $allocationPath -Raw | ConvertFrom-Json
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

function Invoke-LoggedStep {
    param(
        [string]$Name,
        [scriptblock]$Script
    )

    Write-StepBanner -Message ("START {0}" -f $Name)
    $stepStartedAt = Get-Date
    & $Script
    $stepElapsed = (Get-Date) - $stepStartedAt
    Write-StepBanner -Message ("DONE  {0} in {1}" -f $Name, (Format-Elapsed -Elapsed $stepElapsed))
}

function Get-TokenBalance {
    param(
        [string]$Address,
        [string]$TokenIdentifier
    )

    $tokens = Invoke-RestMethod -Uri "$GatewayUrl/accounts/$Address/tokens" -Method Get
    $entry = $tokens | Where-Object { $_.identifier -eq $TokenIdentifier } | Select-Object -First 1
    if (-not $entry) {
        return [bigint]0
    }
    return [bigint]$entry.balance
}

function Wait-TokenBalanceAtLeast {
    param(
        [string]$Address,
        [string]$TokenIdentifier,
        [bigint]$MinimumBalance,
        [int]$TimeoutSeconds = 180
    )

    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    $waitStartedAt = Get-Date
    while ((Get-Date) -lt $deadline) {
        $balance = Get-TokenBalance -Address $Address -TokenIdentifier $TokenIdentifier
        $elapsed = (Get-Date) - $waitStartedAt
        Write-Host ("[{0:o}] WAIT token={1} balance={2} target={3} elapsed={4}" -f (Get-Date).ToUniversalTime(), $TokenIdentifier, $balance, $MinimumBalance, (Format-Elapsed -Elapsed $elapsed))
        if ($balance -ge $MinimumBalance) {
            return
        }
        Start-Sleep -Seconds 3
    }
    throw "Timeout waiting for $TokenIdentifier treasury balance to reach $MinimumBalance"
}

function Invoke-WegldWrap {
    param(
        [bigint]$AmountBase
    )

    if ($AmountBase -le 0) {
        return
    }

    Write-Host ("Wrapping treasury WEGLD amount={0}" -f $AmountBase)
    $wrapOutput = & mxpy tx new `
        --pem $TreasuryPemPath `
        --receiver $WegldWrapperAddress `
        --value $AmountBase.ToString() `
        --gas-limit 6000000 `
        --data "wrapEgld" `
        --chain $ChainId `
        --proxy $GatewayUrl `
        --send
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
    Write-Host $wrapOutput
}

function Get-PoolTotal {
    param(
        [string]$Pool,
        [string]$PerWalletBase
    )

    $count = @($allocation.pools.$Pool).Count
    return ([bigint]$PerWalletBase) * $count
}

function Get-Shard1Operator {
    if (-not (Test-Path $operatorManifestPath)) {
        throw "Operator manifest not found: $operatorManifestPath"
    }

    $manifest = Get-Content $operatorManifestPath -Raw | ConvertFrom-Json
    $operator = @($manifest.wallets | Where-Object { $_.enabled -and $_.shard -eq 1 }) | Select-Object -First 1
    if (-not $operator) {
        throw "No enabled shard 1 operator found in operator manifest."
    }

    return $operator
}

$requiredWegld =
    (Get-PoolTotal -Pool "shard1_sync" -PerWalletBase $SyncWegldPerWallet) +
    (Get-PoolTotal -Pool "shard1_te_swap1_bias" -PerWalletBase $TeSwap1WegldPerWallet) +
    (Get-PoolTotal -Pool "shard0_async1" -PerWalletBase $Async1WegldPerWallet) +
    (Get-PoolTotal -Pool "shard2_async2" -PerWalletBase $Async2WegldPerWallet)

$requiredUsdc = Get-PoolTotal -Pool "shard1_te_swap2_bias" -PerWalletBase $TeSwap2UsdcPerWallet

Write-Host ("Sustained inventory totals: WEGLD={0} USDC={1}" -f $requiredWegld, $requiredUsdc)

if (-not $SkipCallerFunding) {
    Invoke-LoggedStep -Name ("caller EGLD funding target={0}" -f $CallerTargetEgld) -Script {
        & $fundCallers `
            -TreasuryPemPath $TreasuryPemPath `
            -TargetEgld $CallerTargetEgld `
            -GatewayUrl $GatewayUrl `
            -ChainId $ChainId `
            -GasPrice $GasPrice `
            -PollIntervalSeconds $PollIntervalSeconds `
            -ConfirmTimeoutSeconds $ConfirmTimeoutSeconds `
            -HttpTimeoutSeconds $HttpTimeoutSeconds `
            -DryRun:$DryRun
    }
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

if (-not $SkipOperatorFunding) {
    Invoke-LoggedStep -Name ("operator EGLD funding target={0}" -f $OperatorTargetEgld) -Script {
        & $fundOperators `
            -TreasuryPemPath $TreasuryPemPath `
            -TargetEgld $OperatorTargetEgld `
            -GatewayUrl $GatewayUrl `
            -ChainId $ChainId `
            -GasPrice $GasPrice `
            -PollIntervalSeconds $PollIntervalSeconds `
            -ConfirmTimeoutSeconds $ConfirmTimeoutSeconds `
            -HttpTimeoutSeconds $HttpTimeoutSeconds `
            -DryRun:$DryRun
    }
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

if (-not $DryRun) {
    if ($SkipWegldBalanceCheck) {
        if ($AutoWrapWegld) {
            Invoke-WegldWrap -AmountBase $requiredWegld
        }
    } else {
        $currentWegld = Get-TokenBalance -Address $TreasuryAddress -TokenIdentifier $WegldTokenId
        if ($currentWegld -lt $requiredWegld) {
            $deficit = $requiredWegld - $currentWegld
            if (-not $AutoWrapWegld) {
                throw "Treasury WEGLD deficit: need $requiredWegld, have $currentWegld, missing $deficit. Re-run with -AutoWrapWegld or wrap manually first."
            }

            Invoke-LoggedStep -Name ("treasury WEGLD auto-wrap deficit={0}" -f $deficit) -Script {
                Invoke-WegldWrap -AmountBase $deficit
                Wait-TokenBalanceAtLeast -Address $TreasuryAddress -TokenIdentifier $WegldTokenId -MinimumBalance $requiredWegld -TimeoutSeconds $ConfirmTimeoutSeconds
            }
        }
    }
}

Invoke-LoggedStep -Name "treasury WEGLD distribution to caller pools" -Script {
    & $wegldPrep `
        -TreasuryPemPath $TreasuryPemPath `
        -TreasuryAddress $TreasuryAddress `
        -TokenId $WegldTokenId `
        -SyncPerWallet $SyncWegldPerWallet `
        -TeSwap1PerWallet $TeSwap1WegldPerWallet `
        -TeSwap2PerWallet "0" `
        -Async1PerWallet $Async1WegldPerWallet `
        -Async2PerWallet $Async2WegldPerWallet `
        -GatewayUrl $GatewayUrl `
        -ChainId $ChainId `
        -GasLimit $GasLimit `
        -GasPrice $GasPrice `
        -BatchSize $BatchSize `
        -PollIntervalSeconds $PollIntervalSeconds `
        -ConfirmTimeoutSeconds $ConfirmTimeoutSeconds `
        -HttpTimeoutSeconds $HttpTimeoutSeconds `
        -DryRun:$DryRun
}

if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

if ([decimal]$TeSwap2UsdcPerWallet -le 0) {
    Write-Host "skip pool=shard1_te_swap2_bias amount=0"
    exit 0
}

if (-not $StarterUsdcSourcePemPath -or -not $StarterUsdcSourceAddress) {
    $operator = Get-Shard1Operator
    if (-not $StarterUsdcSourcePemPath) {
        $StarterUsdcSourcePemPath = [System.IO.Path]::GetFullPath($operator.pemPath)
    }
    if (-not $StarterUsdcSourceAddress) {
        $StarterUsdcSourceAddress = $operator.address
    }
}

if (-not $DryRun) {
    Invoke-LoggedStep -Name ("starter USDC mint on shard 1 operator={0}" -f $StarterUsdcSourceAddress) -Script {
        & $wrapScript -Shard 1 -Amount $TeSwap2StarterWegldAmount
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }

        & $swap1Script -Shard 1 -Method direct -WegldAmount $TeSwap2StarterWegldAmount -UsdcAmountMin $TeSwap2StarterUsdcMin
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }
} else {
    Write-Host ("Dry-run starter USDC mint: shard=1 wrapAmount={0} minUsdcOut={1}" -f $TeSwap2StarterWegldAmount, $TeSwap2StarterUsdcMin)
}

$args = @(
    "./cmd/distributeesdt",
    "-allocation", $allocationPath,
    "-pool", "shard1_te_swap2_bias",
    "-token-id", $UsdcTokenId,
    "-amount-base", $TeSwap2UsdcPerWallet,
    "-treasury-address", $StarterUsdcSourceAddress,
    "-treasury-pem", $StarterUsdcSourcePemPath,
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

Write-StepBanner -Message "START starter USDC distribution to shard1_te_swap2_bias"
$usdcDistStartedAt = Get-Date
go run @args
Write-StepBanner -Message ("DONE  starter USDC distribution to shard1_te_swap2_bias in {0}" -f (Format-Elapsed -Elapsed ((Get-Date) - $usdcDistStartedAt)))
exit $LASTEXITCODE
