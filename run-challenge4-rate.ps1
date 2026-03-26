param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("0", "1", "2")]
    [string]$Shard,
    [string]$WegldAmount = "1000000000000000000",
    [string]$ConfigPath,
    [string]$StatePath
)

& (Join-Path $PSScriptRoot "run-challenge4-interactor.ps1") `
    -Shard $Shard `
    -ConfigPath $ConfigPath `
    -StatePath $StatePath `
    -InteractorArgs @("get-rate", "-a", $WegldAmount)

exit $LASTEXITCODE
