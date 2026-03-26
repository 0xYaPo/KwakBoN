param(
    [int]$DurationSeconds = 3600,
    [int]$DrainEveryRounds = 10,
    [int]$MaxRounds = 0,
    [string]$Swap1Amount = "10000000000000000",
    [string]$Swap2Amount = "10000",
    [string]$MinAmount = "1",
    [int]$PoolMaxParallel = 4,
    [switch]$ParallelPools
)

if ($DurationSeconds -le 0) {
    throw "DurationSeconds must be > 0"
}
if ($DrainEveryRounds -le 0) {
    throw "DrainEveryRounds must be > 0"
}
if ($PoolMaxParallel -lt 1) {
    throw "PoolMaxParallel must be >= 1"
}

$repoRoot = $PSScriptRoot
$phaseRunner = Join-Path $repoRoot "run-challenge4-phase1.ps1"
$drainRunner = Join-Path $repoRoot "run-challenge4-drain.ps1"

$start = Get-Date
$deadline = $start.AddSeconds($DurationSeconds)
$round = 0

Write-Host ("Challenge 4 sustained run start={0:o} deadline={1:o} poolMaxParallel={2} drainEveryRounds={3}" -f $start.ToUniversalTime(), $deadline.ToUniversalTime(), $PoolMaxParallel, $DrainEveryRounds)

while ((Get-Date) -lt $deadline) {
    if ($MaxRounds -gt 0 -and $round -ge $MaxRounds) {
        break
    }

    $round++
    Write-Host ("Sustained round={0} elapsed={1}" -f $round, ((Get-Date) - $start))

    & $phaseRunner `
        -Swap1Amount $Swap1Amount `
        -Swap2Amount $Swap2Amount `
        -MinAmount $MinAmount `
        -PoolMaxParallel $PoolMaxParallel `
        -ParallelPools:$ParallelPools `
        -SkipDrain
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }

    if ($round % $DrainEveryRounds -eq 0) {
        Write-Host ("Sustained drain checkpoint round={0}" -f $round)
        & $drainRunner -Shard 0
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
        & $drainRunner -Shard 1
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
        & $drainRunner -Shard 2
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    }
}

Write-Host "Final drain cycle"
& $drainRunner -Shard 0
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
& $drainRunner -Shard 1
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
& $drainRunner -Shard 2
exit $LASTEXITCODE
