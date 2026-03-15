# KwakBoN

Clean workspace for the Guild Wars competition on MultiversX.

## Principles

- Keep this repo shareable with collaborators.
- Copy reusable code in deliberately instead of importing the whole previous workspace.
- Never commit wallets, PEM files, `.env` files, challenge notes with private data, or local run outputs.

## Suggested layout

- `cmd/` entrypoints for runnable tools
- `internal/` private application packages
- `configs/` checked-in non-secret config examples
- `docs/` runbooks and competition notes safe to share
- `scripts/` helper scripts for local execution

## Initial workflow

1. Add only the reusable pieces from `mvx-esdt-loadgen`.
2. Replace competition-specific naming, defaults, and docs.
3. Keep secrets in local `.env` files or external wallet paths outside the repo.
4. Add a reviewed `README` section for each tool before sharing it.

## Reusable core included

- [internal/gateway/client.go](/C:/Users/portyp/Mvx/KwakBoN/internal/gateway/client.go) for nonce, balance, shard, status, and tx broadcast calls
- [internal/wallets/wallets.go](/C:/Users/portyp/Mvx/KwakBoN/internal/wallets/wallets.go) and [internal/wallets/pem_signer.go](/C:/Users/portyp/Mvx/KwakBoN/internal/wallets/pem_signer.go) for local wallet loading and PEM signing
- [internal/txsign/txsign.go](/C:/Users/portyp/Mvx/KwakBoN/internal/txsign/txsign.go) for canonical MultiversX tx signing payloads
- [internal/ratelimit/tps.go](/C:/Users/portyp/Mvx/KwakBoN/internal/ratelimit/tps.go) for simple TPS control
- [internal/esdt/esdt.go](/C:/Users/portyp/Mvx/KwakBoN/internal/esdt/esdt.go) for token amount and payload helpers

These are generic enough to reuse across Guild Wars tools without pulling in old challenge runners.

## Wallet Manifest

- Single source of truth: `configs/wallets-manifest.json`
- Example file: [configs/wallets-manifest.example.json](/C:/Users/portyp/Mvx/KwakBoN/configs/wallets-manifest.example.json)
- Validation/summary command:

```bash
go run ./cmd/walletmanifest -manifest ./configs/wallets-manifest.json
```

- Expected fields per wallet:
  - `walletId`
  - `address`
  - `shard`
  - `pemFile`
  - `pemPath`
  - `status`
  - `tags`
  - `enabled`
  - `notes`

## Wallet Generation

Generate wallets until the manifest reaches the desired totals and shard-2 quota:

```bash
go run ./cmd/genwallets \
  -manifest ./configs/wallets-manifest.json \
  -wallets-dir ./wallets \
  -target-total 497 \
  -target-shard2 250 \
  -active-shard2 150 \
  -reserve-shard2 100
```

This command:
- creates PEM files in `./wallets`
- derives wallet address and shard locally
- appends new entries to the manifest
- marks shard-2 wallets as `active_candidate` then `warm_reserve`
- marks other wallets as `window_b_reserve`

## Funding

Fund Window A senders from the treasury wallet up to target EGLD balances:

```bash
go run ./cmd/fundwallets --dry-run
```

Key environment variables:
- `WALLETS_MANIFEST`
- `TREASURY_ADDRESS`
- `TREASURY_PEM_PATH`
- `FUND_INCLUDE_STATUSES`
- `FUND_REQUIRED_TAGS`
- `FUND_SHARD_FILTER`
- `FUND_TARGET_ACTIVE_EGLD`
- `FUND_TARGET_RESERVE_EGLD`

Helper script:
- [run-fund-window-a.ps1](/C:/Users/portyp/Mvx/KwakBoN/run-fund-window-a.ps1)

## Sweep Back

Drain funded wallets back to the treasury wallet after Window A:

```bash
go run ./cmd/sweepwallets --dry-run
```

Key environment variables:
- `WALLETS_MANIFEST`
- `TREASURY_ADDRESS`
- `SWEEP_INCLUDE_STATUSES`
- `SWEEP_EXCLUDE_STATUSES`
- `SWEEP_REQUIRED_TAGS`
- `SWEEP_SHARD_FILTER`
- `SWEEP_MIN_REMAIN_EGLD`

Helper script:
- [run-sweep-window-a.ps1](/C:/Users/portyp/Mvx/KwakBoN/run-sweep-window-a.ps1)

## Window A Sender

Run the Window A manifest-driven `MoveBalance` sender:

```bash
go run ./cmd/windowsprint --dry-run
```

Key environment variables:
- `WALLETS_MANIFEST`
- `SPRINT_SENDER_STATUSES`
- `SPRINT_RECEIVER_STATUSES`
- `SPRINT_REQUIRED_SENDER_TAGS`
- `SPRINT_REQUIRED_RECEIVER_TAGS`
- `SPRINT_SHARD_FILTER`
- `SPRINT_TARGET_TX`
- `SUSTAINED_TPS`
- `SPRINT_WORKERS`
- `CONFIRM_SUCCESS_TARGET`

Helper script:
- [run-window-a.ps1](/C:/Users/portyp/Mvx/KwakBoN/run-window-a.ps1)

## Local setup

Create a local `.env` from `.env.example` once the first tool lands in the repo.
