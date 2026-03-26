param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("shard1_sync", "shard1_te_swap1_bias", "shard1_te_swap2_bias", "shard0_async1", "shard2_async2")]
    [string]$Pool,
    [Parameter(Mandatory = $true)]
    [ValidateSet("0", "1", "2")]
    [string[]]$ForwarderShards,
    [string]$AllocationPath = "./configs/challenge4/wallet-allocation.challenge4-callers.json",
    [int]$StartIndex = 0,
    [int]$Limit = 0,
    [string]$OutputPath,
    [switch]$PassThru
)

$repoRoot = $PSScriptRoot
$resolvedAllocation = [System.IO.Path]::GetFullPath((Join-Path $repoRoot $AllocationPath))
if (-not (Test-Path $resolvedAllocation)) {
    throw "Allocation file not found: $resolvedAllocation"
}

if (-not $OutputPath) {
    $sliceSuffix = "s{0}-l{1}" -f $StartIndex, $(if ($Limit -gt 0) { $Limit } else { "all" })
    $OutputPath = Join-Path $repoRoot ("tmp/challenge4-official/{0}.{1}.{2}.toml" -f $Pool, (($ForwarderShards | Sort-Object) -join ""), $sliceSuffix)
}

$resolvedOutputPath = [System.IO.Path]::GetFullPath($OutputPath)
$outputDir = Split-Path -Parent $resolvedOutputPath
if ($outputDir -and -not (Test-Path $outputDir)) {
    New-Item -ItemType Directory -Path $outputDir -Force | Out-Null
}

$allocation = Get-Content $resolvedAllocation -Raw | ConvertFrom-Json
$items = @($allocation.pools.$Pool)
if ($items.Count -eq 0) {
    throw "Pool $Pool is empty in allocation file."
}
if ($StartIndex -lt 0) {
    throw "StartIndex must be >= 0"
}
if ($StartIndex -ge $items.Count) {
    throw "StartIndex $StartIndex is outside pool $Pool size $($items.Count)"
}
$items = @($items | Select-Object -Skip $StartIndex)
if ($Limit -gt 0) {
    $items = @($items | Select-Object -First $Limit)
}

$pairAddress = "erd1qqqqqqqqqqqqqpgqeel2kumf0r8ffyhth7pqdujjat9nx0862jpsg2pqaq"
$wegldAddress = "erd1qqqqqqqqqqqqqpgqmuk0q2saj0mgutxm4teywre6dl8wqf58xamqdrukln"

$walletPemLines = foreach ($item in $items) {
    $pemPath = ([System.IO.Path]::GetFullPath($item.pemPath)) -replace "\\", "/"
    "    '{0}'," -f $pemPath
}

$contractAddressLines = foreach ($shard in ($ForwarderShards | Select-Object -Unique)) {
    $statePath = Join-Path $repoRoot ("state/challenge4/forwarder-blind-shard{0}.toml" -f $shard)
    if (-not (Test-Path $statePath)) {
        throw "Forwarder state file not found for shard ${shard}: $statePath"
    }
    $content = Get-Content $statePath -Raw
    $match = [regex]::Match($content, 'contract_address\s*=\s*"([^"]+)"')
    if (-not $match.Success) {
        throw "contract_address not found in state file: $statePath"
    }
    '    "{0}",' -f $match.Groups[1].Value
}

$toml = @(
    "chain_type = 'real'"
    "gateway_uri = 'https://gateway.battleofnodes.com'"
    "wegld_address = '$wegldAddress'"
    "pair_address = '$pairAddress'"
    "wegld_token_id = 'WEGLD-bd4d79'"
    "usdc_token_id = 'USDC-c76f1f'"
    ""
    "wallet_pem_paths = ["
    $walletPemLines
    "]"
    ""
    "contract_addresses = ["
    $contractAddressLines
    "]"
) -join "`r`n"

Set-Content -LiteralPath $resolvedOutputPath -Value $toml

if ($PassThru) {
    [pscustomobject]@{
        ConfigPath = $resolvedOutputPath
        Pool = $Pool
        ForwarderShards = @($ForwarderShards | Select-Object -Unique)
        WalletCount = $items.Count
        StartIndex = $StartIndex
        Limit = $Limit
    }
}
