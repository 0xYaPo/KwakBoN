# KwakBoN

A workspace for the Guild Wars competition on MultiversX.

Current challenge brief:
- [docs/challenge3.md](docs/challenge3.md)
- [docs/Challenge3-Prep-Plan.md](docs/Challenge3-Prep-Plan.md)
- [docs/Challenge4.md](docs/Challenge4.md)
- [docs/Challenge4details.md](docs/Challenge4details.md)
- [docs/Challenge4-Prep-Plan.md](docs/Challenge4-Prep-Plan.md)
- [docs/Challenge4-Operator-Checklist.md](docs/Challenge4-Operator-Checklist.md)

Challenge 4 live command flow is now documented in one place:
- reset
- integrated prep
- spot checks
- sustained launch
- final drains
- see [docs/Challenge4-Operator-Checklist.md](docs/Challenge4-Operator-Checklist.md)

Challenge 4 summary:
- confirmed pair:
  - `erd1qqqqqqqqqqqqqpgqeel2kumf0r8ffyhth7pqdujjat9nx0862jpsg2pqa`
- sanctioned deploy artifact used
- validated working lanes:
  - shard `1` `blindSync`
  - shard `0` `blindAsyncV1`
  - shard `2` `blindAsyncV2`
  - shard `1` `blindTransfExec` via `swap1` and `swap2`
- validated live inventory model:
  - callers funded to `0.5 EGLD`
  - operators funded to `5 EGLD`
  - `swap2` default reduced to `10000`
  - `swap2` pool seeded with `3000000` base `USDC` per wallet
- verified real run outcome from API history:
  - `blindSync`: `869`
  - `blindAsyncV1`: `1426`
  - `blindAsyncV2`: `1429`
  - `blindTransfExec`: `315`
  - total successful calls: `4039`
- practical caveat:
  - the tx retrieval / confirmation path on the BoN API is flaky enough to stall or kill the current sustained runner
  - simple batches still produce successful on-chain calls, but long-run orchestration is not fully reliable yet

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

- [internal/gateway/client.go](internal/gateway/client.go) for nonce, balance, shard, status, and tx broadcast calls
- [internal/wallets/wallets.go](internal/wallets/wallets.go) and [internal/wallets/pem_signer.go](internal/wallets/pem_signer.go) for local wallet loading and PEM signing
- [internal/txsign/txsign.go](internal/txsign/txsign.go) for canonical MultiversX tx signing payloads
- [internal/ratelimit/tps.go](internal/ratelimit/tps.go) for simple TPS control
- [internal/esdt/esdt.go](internal/esdt/esdt.go) for token amount and payload helpers

These are generic enough to reuse across Guild Wars tools without pulling in old challenge runners.

## Wallet Manifest

- Generic example file: [configs/wallets-manifest.example.json](configs/wallets-manifest.example.json)
- Challenge 3 live files:
  - `configs/wallets-manifest.challenge3-part1.live.json`
  - `configs/wallets-manifest.challenge3-part2.live.json`
- Validation/summary command:

```bash
go run ./cmd/walletmanifest -manifest ./configs/wallets-manifest.challenge3-part1.live.json
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

## Challenge 3 Live Workflow

Challenge 3 only counts cross-shard `MoveBalance` transactions.

Live wallet generation helpers:
- [run-gen-challenge3-part1-live.ps1](run-gen-challenge3-part1-live.ps1)
- [run-gen-challenge3-part2-live.ps1](run-gen-challenge3-part2-live.ps1)

These generate fresh `500`-wallet sets with balanced shard targets:
- shard `0`: `166`
- shard `1`: `167`
- shard `2`: `167`

Part 1 live assets:
- manifest: `configs/wallets-manifest.challenge3-part1.live.json`
- wallet dir: `wallets/challenge3-part1-live`

Part 2 live assets:
- manifest: `configs/wallets-manifest.challenge3-part2.live.json`
- wallet dir: `wallets/challenge3-part2-live`

Do not reuse rehearsal manifests or wallet directories during the live challenge.

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

Challenge 3 funding helpers:
- [run-fund-challenge3.ps1](run-fund-challenge3.ps1)
- [run-fund-challenge3-part2.ps1](run-fund-challenge3-part2.ps1)

Current live defaults:
- Part 1: `4.0 EGLD` per wallet
- Part 2: `1.0 EGLD` per wallet
- funding uses direct GL -> sender wallet transfers only
- funding confirmation polling is bounded by timeout

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

Challenge 3 sweep helper:
- [run-sweep-challenge3-part2.ps1](run-sweep-challenge3-part2.ps1)

Sweep confirmation polling now uses concurrent status workers and bounded timeouts.

## Sender

Run the manifest-driven `MoveBalance` sender:

```bash
go run ./cmd/windowsprint --dry-run
```

Current sender behavior:
- sender selection comes from manifest status/tag filters
- routing supports both same-shard and cross-shard modes
- `SPRINT_REUSE_SENDERS_AS_RECEIVERS=true` allows sender-set reuse as receiver pool
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

Challenge 3 sprint helpers:
- [run-sprint-challenge3.ps1](run-sprint-challenge3.ps1)
- [run-sprint-challenge3-part2.ps1](run-sprint-challenge3-part2.ps1)

## Bulk Sender

Run the specialized bulk sender that uses `transaction/send-multiple`:

```bash
go run ./cmd/bulksprint --dry-run
```

Current bulk sender behavior:
- supports `BULK_ROUTING_MODE=same-shard|cross-shard`
- one worker per active wallet
- one nonce+balance read per batch
- bounded nonce lookahead per wallet
- bulk broadcast through the gateway `send-multiple` path
- optional lightweight confirmation polling for aggregate finality checks
- cross-shard mode now uses receiver-pool rotation within the target shard instead of fixed one-to-one pairings

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

Challenge 3 validated profiles:

Part 1:
- live full-set probe, `500` wallets, cross-shard, `25/50`, min value:
  - `1,042,100` accepted in `10` minutes

Part 2:
- full-set rehearsal, `500` wallets, `0.75 EGLD` funding, cross-shard receiver-pool routing, `12/15`:
  - `1,286,553` accepted in `10` minutes
- aggressive burst probe, `25/50`:
  - `149,913` accepted in `3m56s`
  - too unstable as a sustained 30-minute profile

Current interpretation:
- `bulksprint` is the primary Challenge 3 sender
- `windowsprint` remains the fallback path
- Part 1 favors aggressive throughput
- Part 2 requires more care because each tx carries `0.01 EGLD`
- receiver-pool rotation materially improved Part 2 balance drift

Challenge 3 bulk helpers:
- [run-bulk-challenge3.ps1](run-bulk-challenge3.ps1)
- [run-bulk-challenge3-part2.ps1](run-bulk-challenge3-part2.ps1)

## Local setup

Create a local `.env` from `.env.example` once the first tool lands in the repo.

## Challenge 4 Workflow

Challenge 4 is contract-call oriented and currently uses the sanctioned `forwarder-blind-bon.wasm` flow through the vendored `dex-interactor`.

Key docs:
- [Challenge4-Prep-Plan.md](docs/Challenge4-Prep-Plan.md)
- [Challenge4-Operator-Checklist.md](docs/Challenge4-Operator-Checklist.md)

Current local baseline:
- confirmed pair:
  - `erd1qqqqqqqqqqqqqpgqeel2kumf0r8ffyhth7pqdujjat9nx0862jpsg2pqa`
- sanctioned forwarders deployed on shards `0`, `1`, `2`
- caller wallet set:
  - `100` wallets
  - shard split `33 / 34 / 33`
- caller allocation:
  - shard `1` sync pool
  - shard `1` TE pools
  - shard `0` async1 pool
  - shard `2` async2 pool

Useful Challenge 4 helpers:
- wallet generation and funding:
  - [run-gen-challenge4-operator-wallets.ps1](run-gen-challenge4-operator-wallets.ps1)
  - [run-gen-challenge4-caller-wallets.ps1](run-gen-challenge4-caller-wallets.ps1)
  - [run-fund-challenge4-callers.ps1](run-fund-challenge4-callers.ps1)
  - [run-fund-challenge4-operators.ps1](run-fund-challenge4-operators.ps1)
- deploy and control:
  - [run-challenge4-deploy.ps1](run-challenge4-deploy.ps1)
  - [run-challenge4-wrap.ps1](run-challenge4-wrap.ps1)
  - [run-challenge4-drain.ps1](run-challenge4-drain.ps1)
  - [run-sweep-challenge4-reset.ps1](run-sweep-challenge4-reset.ps1)
- caller and pool execution:
  - [run-challenge4-caller.ps1](run-challenge4-caller.ps1)
  - [run-challenge4-pool.ps1](run-challenge4-pool.ps1)
  - [run-challenge4-phase1-launch.ps1](run-challenge4-phase1-launch.ps1)
  - [run-challenge4-phase1-postdrain.ps1](run-challenge4-phase1-postdrain.ps1)
  - [run-challenge4-prep-wraps.ps1](run-challenge4-prep-wraps.ps1)
  - [run-challenge4-prep-inventory-treasury.ps1](run-challenge4-prep-inventory-treasury.ps1)
  - [run-challenge4-sustained.ps1](run-challenge4-sustained.ps1)

Current documented default amounts:
- `swap1`: `0.01 WEGLD`
- `swap2`: `10000`

Current operational recommendation:
- full clean reset is available before rehearsal:
  - `run-sweep-challenge4-reset.ps1`
- preferred prep path for sustained runs:
  - restore caller/operator `EGLD` to the lean live targets:
    - callers `0.5 EGLD`
    - operators `5 EGLD`
  - treasury-seed only the opening `WEGLD` inventory:
    - about `11.0 EGLD`
  - seed only a tiny `USDC` starter pool for `swap2 te`
  - all of the above is now integrated into:
    - [run-challenge4-prep-inventory-treasury.ps1](run-challenge4-prep-inventory-treasury.ps1)
- preferred execution path:
  - `run-challenge4-sustained.ps1`
  - `-ParallelPools -PoolMaxParallel 4 -DrainEveryRounds 10`
- validated one-round sustained rehearsal succeeded with:
  - shard `1` sync
  - shard `1` `swap1 te`
  - shard `1` `swap2 te`
  - shard `0` async1
  - shard `2` async2
