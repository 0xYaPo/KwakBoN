# KwakBoN

Clean workspace for the Guild Wars competition on MultiversX.

Current challenge brief:
- [docs/Supernova-Surge-2026-03-16.md](/C:/Users/portyp/Mvx/KwakBoN/docs/Supernova-Surge-2026-03-16.md)
- [docs/Supernova-Runbook.md](/C:/Users/portyp/Mvx/KwakBoN/docs/Supernova-Runbook.md)
- [docs/Supernova-Day-Of-Checklist.md](/C:/Users/portyp/Mvx/KwakBoN/docs/Supernova-Day-Of-Checklist.md)
- [docs/Supernova-Balanced-250-Set.md](/C:/Users/portyp/Mvx/KwakBoN/docs/Supernova-Balanced-250-Set.md)
- [docs/BulkSprint-Benchmarks-2026-03-16.md](/C:/Users/portyp/Mvx/KwakBoN/docs/BulkSprint-Benchmarks-2026-03-16.md)

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

Fund matched sender wallets from the treasury wallet up to target EGLD balances:

```bash
go run ./cmd/fundwallets --dry-run
```

Current funding flow is hardened for treasury nonce pressure:
- periodic treasury nonce refresh during large batches
- short cooldown between treasury sends
- nonce resync and retry on `veryHighNonceInTx` / `lowerNonceInTx`

Key environment variables:
- `WALLETS_MANIFEST`
- `TREASURY_ADDRESS`
- `TREASURY_PEM_PATH`
- `FUND_INCLUDE_STATUSES`
- `FUND_REQUIRED_TAGS`
- `FUND_SHARD_FILTER`
- `FUND_TARGET_ACTIVE_EGLD`
- `FUND_TARGET_RESERVE_EGLD`
- `FUND_NONCE_REFRESH_EVERY`
- `FUND_SEND_COOLDOWN_MS`
- `FUND_NONCE_RETRY_COOLDOWN_MS`
- `FUND_NONCE_RESYNC_ATTEMPTS`

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

## Sender

Run the manifest-driven `MoveBalance` sender:

```bash
go run ./cmd/windowsprint --dry-run
```

Current sender behavior:
- sender selection comes from manifest status/tag filters
- receiver routing is shard-aware: a sender only picks receivers from its own shard
- per-wallet inflight limit and cooldowns reduce nonce drift under congestion
- transient gateway errors do not immediately kill the run

Key environment variables:
- `WALLETS_MANIFEST`
- `SPRINT_RECEIVER_ADDRESSES`
- `SPRINT_SENDER_STATUSES`
- `SPRINT_RECEIVER_STATUSES`
- `SPRINT_REQUIRED_SENDER_TAGS`
- `SPRINT_REQUIRED_RECEIVER_TAGS`
- `SPRINT_SHARD_FILTER`
- `SPRINT_TARGET_TX`
- `SPRINT_DURATION_SECONDS`
- `SUSTAINED_TPS`
- `SPRINT_WORKERS`
- `SPRINT_MAX_INFLIGHT_PER_WALLET`
- `SPRINT_SUCCESS_COOLDOWN_MS`
- `SPRINT_TRANSIENT_COOLDOWN_MS`
- `SPRINT_QUARANTINE_COOLDOWN_MS`
- `SPRINT_QUARANTINE_TRANSIENT_THRESHOLD`
- `CONTINUE_ON_TRANSIENT_SEND_ERROR`
- `CONFIRM_SUCCESS_TARGET`

Helper script:
- [run-window-a.ps1](/C:/Users/portyp/Mvx/KwakBoN/run-window-a.ps1)

## Bulk Sender

Run the specialized bulk sender that uses `transaction/send-multiple`:

```bash
go run ./cmd/bulksprint --dry-run
```

Current bulk sender behavior:
- same-shard only ring routing built directly from the selected sender set
- one worker per active wallet
- one nonce+balance read per batch
- bounded nonce lookahead per wallet
- bulk broadcast through the gateway `send-multiple` path
- optional lightweight confirmation polling for aggregate finality checks

Key environment variables:
- `WALLETS_MANIFEST`
- `BULK_SENDER_STATUSES`
- `BULK_REQUIRED_TAGS`
- `BULK_SHARD_FILTER`
- `BULK_TARGET_TX`
- `BULK_DURATION_SECONDS`
- `BULK_BATCH_SIZE`
- `BULK_MAX_NONCE_LOOKAHEAD`
- `BULK_MAX_CONCURRENT_READS`
- `BULK_MAX_ACTIVE_WALLETS`
- `BULK_IDLE_SLEEP_MS`
- `BULK_VALUE`
- `BULK_WAIT_CONFIRM`
- `BULK_CONFIRM_WORKERS`
- `BULK_POLL_INTERVAL_SECONDS`
- `BULK_CONFIRM_TIMEOUT_SECONDS`

This mode is meant for throughput benchmarking on the shared API path. It is intentionally narrower and less defensive than `windowsprint`.

Current validated bulk profile:
- `BULK_BATCH_SIZE=25`
- `BULK_MAX_NONCE_LOOKAHEAD=50`
- `BULK_MAX_CONCURRENT_READS=24`

Earlier benchmark results:
- `20,000 tx`: send phase about `4s`, `20,000/20,000 success`
- `50,000 tx`: send phase about `13s`, `49,958/50,000 success` after `4m`, `42 pending`, `0 failed`

Live Supernova results with the balanced `250` wallet set:
- fixed target `100,000 tx`: `100,000/100,000 accepted` in about `32.7s`
- duration `10 min`: `1,747,130 sent`, `1,746,002 accepted`, `15 errors`

Current interpretation:
- `bulksprint` is now a serious primary candidate on live Supernova
- `windowsprint` remains the fallback sender and simpler operator path

Detailed notes:
- [docs/BulkSprint-Benchmarks-2026-03-16.md](/C:/Users/portyp/Mvx/KwakBoN/docs/BulkSprint-Benchmarks-2026-03-16.md)
- [docs/BulkSprint-Roadmap.md](/C:/Users/portyp/Mvx/KwakBoN/docs/BulkSprint-Roadmap.md)

## Validated Test Profiles

Observed on `prepSupernova`:
- Window A profile: `250` senders, shard `2`, same-shard receivers, `750 TPS`, `192` workers, `20,000/20,000 success`
- Window A upper-bound probe: `250` senders, `1500 TPS`, `20,000/20,000 success`
- Window B profile: `497` senders, all shards, shard-aware receivers, `700 TPS`, `192` workers, `20,000/20,000 success`
- Funding retry/resume for large treasury batches completed successfully after hardening

Observed on live Supernova:
- balanced `windowsprint`, `250` senders, `10 min`, `1200 TPS` target: `324,160 sent`
- balanced `bulksprint`, `250` senders, `10 min`, tuned `25/50` profile: `1,746,002 accepted`

## Local setup

Create a local `.env` from `.env.example` once the first tool lands in the repo.
