param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("0", "1", "2")]
    [string]$Shard,
    [Parameter(Mandatory = $true)]
    [ValidateSet("direct", "sync", "async1", "async2", "te")]
    [string]$Method,
    [Parameter(Mandatory = $true)]
    [string]$WegldAmount,
    [string]$UsdcAmountMin = "1",
    [string]$ConfigPath,
    [string]$StatePath
)

& (Join-Path $PSScriptRoot "run-challenge4-interactor.ps1") `
    -Shard $Shard `
    -ConfigPath $ConfigPath `
    -StatePath $StatePath `
    -InteractorArgs @("swap1", $Method, "-a", $WegldAmount, "-m", $UsdcAmountMin)

exit $LASTEXITCODE
