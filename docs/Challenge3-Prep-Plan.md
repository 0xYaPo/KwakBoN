# Challenge 3 Prep Plan

Challenge 3 only counts cross-shard `MoveBalance` transactions. Sender and receiver must be on different shards.

This repo now supports cross-shard routing in both sender tools:

- `windowsprint` via `SPRINT_ROUTING_MODE=cross-shard`
- `bulksprint` via `BULK_ROUTING_MODE=cross-shard`

## Current State

- `bulksprint` remains the primary candidate for throughput.
- `windowsprint` is the fallback path.
- Current checked-in manifest is not suitable for final Challenge 3 ops.
- Current shard distribution is too skewed toward shard `2`:
  - shard `0`: `69`
  - shard `1`: `122`
  - shard `2`: `308`

Challenge 3 requires:

- Part 1: `500` fresh wallets
- Part 2: another `500` fresh wallets
- both sets funded directly by the guild leader wallet
- cross-shard routing only
- Part 1 minimum tx value: `1e-18 EGLD`
- Part 2 minimum tx value: `0.01 EGLD`

## Operating Model

Use two distinct manifests:

- `configs/wallets-manifest.challenge3-part1.live.json`
- `configs/wallets-manifest.challenge3-part2.live.json`

Reason:

- lower operator risk
- easier funding filters
- easier to guarantee no wallet reuse across parts
- easier post-run auditing
- avoids accidental overwrite or PEM mismatch with rehearsal files

## Wallet Strategy

Target a balanced sender spread across all three shards for each part.

Practical target:

- shard `0`: about `166`
- shard `1`: about `167`
- shard `2`: about `167`

Exact counts do not need to be perfect, but each shard must have enough senders to sustain cross-shard traffic.

Recommended tags/statuses per manifest:

- status: `challenge3_part1_sender` or `challenge3_part2_sender`
- tags:
  - `sender`
  - `challenge3`
  - `part1` or `part2`
  - `shard0` / `shard1` / `shard2`

Treasury / GL wallet should remain separate and must never be used as a scoring sender.

## Tool Choice

Primary:

- `bulksprint`

Fallback:

- `windowsprint`

Reason:

- `bulksprint` has the best throughput history
- `windowsprint` is simpler to reason about if the bulk path becomes unstable

## Part 1 Preset

Use:

- cross-shard routing
- `BULK_VALUE=1`
- `BULK_WAIT_CONFIRM=false` for main race mode
- optional bounded confirm probe before challenge

Budget reminder:

- Part 1 budget includes both fees and tx value
- keep tx value at the minimum required

## Part 2 Preset

Use:

- cross-shard routing
- `BULK_VALUE=10000000000000000` (`0.01 EGLD`)
- fresh Part 2 manifest only
- no wallet reuse from Part 1

Part 2 is a one-shot working-capital problem. The operating rule is:

- fund once from GL wallet
- do not refill without sweeping/resetting first
- optimize for total counted tx within `30` minutes, not perfectly flat wallet balances

Recommended opening model:

- sender set is also the receiver set
- no sink wallets
- cross-shard circulation:
  - shard `0 -> 1`
  - shard `1 -> 2`
  - shard `2 -> 0`
- `bulksprint` receiver-pool rotation within the target shard, not fixed one-to-one pairs

Recommended starting funding:

- tested rehearsal baseline: `0.75 EGLD` per wallet
- live full-deployment option: `1.0 EGLD` per wallet
- keep the actual live choice consistent with the Part 2 budget rule before launch

Recommended opening `bulksprint` profile:

- `BULK_VALUE=10000000000000000`
- `BULK_ROUTING_MODE=cross-shard`
- `BULK_BATCH_SIZE=12`
- `BULK_MAX_NONCE_LOOKAHEAD=15`
- `BULK_MAX_CONCURRENT_READS=24`
- `BULK_WAIT_CONFIRM=false`

Reasoning:

- `0.01 EGLD` principal recycles
- fees are the true burn
- `12/15` outperformed `15/20` in full-set rehearsal
- receiver-pool routing removed the catastrophic dust tail seen with fixed pair routing

Current best rehearsal result:

- full Part 2 set: `500` wallets
- funding: `0.75 EGLD` per wallet
- runtime: `10` minutes
- result: `1,286,553` accepted tx
- errors: `546`
- estimated fee burn at `0.00005 EGLD / tx`: about `64.33 EGLD`

Observed post-run balance distribution after receiver-pool routing:

- shard `0`: min `0.128500`, avg `0.713178`, median `0.701225`, max `1.635950`, `<0.1` = `0`
- shard `1`: min `0.119850`, avg `0.480599`, median `0.431300`, max `1.576850`, `<0.1` = `0`
- shard `2`: min `0.159200`, avg `0.670726`, median `0.651450`, max `1.525950`, `<0.1` = `0`
- overall: min `0.119850`, avg `0.621318`, median `0.551525`, max `1.635950`, `<0.1` = `0`

Interpretation:

- principal is circulating instead of getting stranded
- balance loss closely matches fee burn
- the major pre-patch drift problem is materially reduced
- shard `1` still runs weaker than shards `0` and `2`, but not catastrophically

Current Part 2 recommendation:

- primary profile: `500 wallets`, `0.75 EGLD`, `12/15`, patched receiver-pool routing
- do not re-fund mid-run
- sweep/reset before any new attempt
- keep `windowsprint` only as fallback

Part 2 fallback `windowsprint` profile:

- `SPRINT_ROUTING_MODE=cross-shard`
- `SPRINT_REUSE_SENDERS_AS_RECEIVERS=true`
- `SPRINT_TX_VALUE=10000000000000000`
- start with `SPRINT_MAX_INFLIGHT_PER_WALLET=1`
- use only if bulk mode becomes unstable

## Timeline

Before `15:45 UTC` on March 24, 2026:

- finalize cross-shard sender configs
- prepare Part 1 and Part 2 manifests
- rehearse funding and launch scripts
- prepare content checklist

At `15:45 UTC`:

- receive `2,500 EGLD`
- fund Part 1 wallets directly from GL wallet
- keep Part 2 reserve untouched

At `16:00 UTC`:

- launch `bulksprint` Part 1 profile

At `16:30 UTC`:

- stop Part 1
- generate or load Part 2 manifest
- fund Part 2 wallets directly from GL wallet
- switch minimum tx value to `0.01 EGLD`

At `17:00 UTC`:

- launch Part 2 profile

## Remaining Work

1. Build or prepare two Challenge 3 manifests with fresh wallets.
2. Rehearse direct GL funding against the intended Part 1 and Part 2 statuses.
3. Run small live cross-shard probes with both `bulksprint` and `windowsprint`.
4. Decide final challenge-day profile for batch size and lookahead.
5. Prepare content submission checklist.

## Operator Commands

Generate fresh Part 1 wallets:

```powershell
.\run-gen-challenge3-part1-live.ps1
```

Generate fresh Part 2 wallets:

```powershell
.\run-gen-challenge3-part2-live.ps1
```

Sweep Part 2 set:

```powershell
.\run-sweep-challenge3-part2.ps1
```

Refill Part 2 set to `0.75 EGLD`:

```powershell
.\run-fund-challenge3-part2.ps1 `
  -TreasuryPemPath C:\secure\mvx\treasury-shard2.pem `
  -TargetEgld 1.0
```

Run Part 2 baseline rehearsal / race profile:

```powershell
.\run-bulk-challenge3-part2.ps1 `
  -ManifestPath .\configs\wallets-manifest.challenge3-part2.json `
  -TargetTx 9999999 `
  -DurationSeconds 600 `
  -BatchSize 12 `
  -MaxNonceLookahead 15
```

Assess post-run refill pressure:

```powershell
.\run-fund-challenge3-part2.ps1 `
  -TreasuryPemPath C:\secure\mvx\treasury-shard2.pem `
  -TargetEgld 0.75 `
  -DryRun
```

Assess post-run balance drift:

```powershell
.\tmp-test\balance_stats.ps1
```

Optional late-run shard rebalance dry-run:

```powershell
.\run-rebalance-challenge3-part2-shard.ps1 `
  -Shard 1 `
  -TreasuryPemPath C:\secure\mvx\treasury-shard2.pem `
  -TargetEgld 0.75 `
  -DryRun
```

Optional late-run shard rebalance live:

```powershell
.\run-rebalance-challenge3-part2-shard.ps1 `
  -Shard 1 `
  -TreasuryPemPath C:\secure\mvx\treasury-shard2.pem `
  -TargetEgld 0.75
```
