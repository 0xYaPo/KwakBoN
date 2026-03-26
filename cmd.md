set -a && source .env && set +a
go run ./cmd/fundwallets --dry-run
bash scripts/check_treasury.sh
go run ./cmd/fundwallets
go run ./cmd/windowsprint --dry-run
go run ./cmd/windowsprint
go run ./cmd/sweepwallets
