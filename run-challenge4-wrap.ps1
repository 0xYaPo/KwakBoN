param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("0", "1", "2")]
    [string]$Shard,
    [Parameter(Mandatory = $true)]
    [string]$Amount,
    [string]$ConfigPath,
    [string]$StatePath
)

& (Join-Path $PSScriptRoot "run-challenge4-interactor.ps1") `
    -Shard $Shard `
    -ConfigPath $ConfigPath `
    -StatePath $StatePath `
    -InteractorArgs @("wrap", "-a", $Amount)

exit $LASTEXITCODE
