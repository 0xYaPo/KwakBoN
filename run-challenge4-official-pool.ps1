param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("shard1_sync", "shard1_te_swap1_bias", "shard1_te_swap2_bias", "shard0_async1", "shard2_async2")]
    [string]$Pool,
    [Parameter(Mandatory = $true)]
    [ValidateSet("0", "1", "2")]
    [string[]]$ForwarderShards,
    [Parameter(Mandatory = $true)]
    [ValidateSet("swap1", "swap2", "wrap", "deploy", "drain", "balances")]
    [string]$Action,
    [ValidateSet("direct", "sync", "async1", "async2", "te")]
    [string]$Method,
    [string]$Amount,
    [string]$MinAmount = "1",
    [int]$StartIndex = 0,
    [int]$Limit = 0,
    [string]$StatePath,
    [switch]$WhatIf
)

$repoRoot = $PSScriptRoot
$generator = Join-Path $repoRoot "run-gen-challenge4-official-config.ps1"
$runner = Join-Path $repoRoot "run-challenge4-official-interactor.ps1"

$configInfo = & $generator `
    -Pool $Pool `
    -ForwarderShards $ForwarderShards `
    -StartIndex $StartIndex `
    -Limit $Limit `
    -PassThru
$configPath = $configInfo.ConfigPath

if (-not $StatePath) {
    $StatePath = Join-Path $repoRoot ("state/challenge4/official/{0}.{1}.deploy.toml" -f $Pool, (($ForwarderShards | Sort-Object) -join ""))
}

switch ($Action) {
    "deploy" { $interactorArgs = @("deploy") }
    "wrap" {
        if (-not $Amount) {
            throw "Amount is required for Action=wrap"
        }
        $interactorArgs = @("wrap", "-a", $Amount)
    }
    "drain" { $interactorArgs = @("drain") }
    "balances" { $interactorArgs = @("balances") }
    "swap1" {
        if (-not $Method) {
            throw "Method is required for Action=swap1"
        }
        if (-not $Amount) {
            $Amount = "100000000000000000"
        }
        $interactorArgs = @("swap1", $Method, "-a", $Amount, "-m", $MinAmount)
    }
    "swap2" {
        if (-not $Method) {
            throw "Method is required for Action=swap2"
        }
        if (-not $Amount) {
            $Amount = "40000"
        }
        $interactorArgs = @("swap2", $Method, "-a", $Amount, "-m", $MinAmount)
    }
}

Write-Host ("official pool={0} wallets={1} startIndex={2} limit={3} forwarderShards={4} action={5}{6}" -f `
    $Pool,
    $configInfo.WalletCount,
    $StartIndex,
    $(if ($Limit -gt 0) { $Limit } else { "all" }),
    (($configInfo.ForwarderShards | Sort-Object) -join ","),
    $Action,
    ($(if ($Method) { " method=$Method" } else { "" })))
Write-Host ("config={0}" -f $configPath)

if ($WhatIf) {
    Write-Host (".\\run-challenge4-official-interactor.ps1 -ConfigPath `"{0}`" -StatePath `"{1}`" -InteractorArgs @({2})" -f `
        $configPath,
        $StatePath,
        (($interactorArgs | ForEach-Object { "'$_'" }) -join ", "))
    exit 0
}

& $runner -ConfigPath $configPath -StatePath $StatePath -InteractorArgs $interactorArgs
exit $LASTEXITCODE
