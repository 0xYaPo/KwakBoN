param(
    [string]$ManifestPath = "./configs/wallets-manifest.challenge3-part2.live.json",
    [string]$WalletsDir = "./wallets/challenge3-part2-live",
    [string]$Prefix = "c3p2live",
    [int]$TargetShard0 = 166,
    [int]$TargetShard1 = 167,
    [int]$TargetShard2 = 167
)

go run ./cmd/genwallets `
    -mode challenge3 `
    -manifest $ManifestPath `
    -wallets-dir $WalletsDir `
    -prefix $Prefix `
    -target-shard0 $TargetShard0 `
    -target-shard1 $TargetShard1 `
    -target-shard2-c3 $TargetShard2 `
    -status challenge3_part2_sender `
    -extra-tags sender,challenge3,part2 `
    -notes "Live Challenge 3 Part 2 wallets."

exit $LASTEXITCODE
