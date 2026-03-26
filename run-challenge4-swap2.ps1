param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("0", "1", "2")]
    [string]$Shard,
    [Parameter(Mandatory = $true)]
    [ValidateSet("direct", "sync", "async1", "async2", "te")]
    [string]$Method,
    [Parameter(Mandatory = $true)]
    [string]$UsdcAmount,
    [string]$WegldAmountMin = "1",
    [string]$ConfigPath,
    [string]$StatePath
)

& (Join-Path $PSScriptRoot "run-challenge4-interactor.ps1") `
    -Shard $Shard `
    -ConfigPath $ConfigPath `
    -StatePath $StatePath `
    -InteractorArgs @("swap2", $Method, "-a", $UsdcAmount, "-m", $WegldAmountMin)

exit $LASTEXITCODE
