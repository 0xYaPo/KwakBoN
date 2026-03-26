# Challenge 4 Operator Checklist

Reference brief:
- [Challenge4.md](../docs/Challenge4.md)
- [Challenge4-Prep-Plan.md](../docs/Challenge4-Prep-Plan.md)

## Challenge-Day Timeline

Use this as the default operating sequence for the live window.

### `T-60m` to `T-45m` - clean reset

Sweep callers, operators, and forwarders back to a clean state:

```powershell
.\run-sweep-challenge4-reset.ps1
```

Expected result:
- no meaningful `EGLD`, `WEGLD`, or `USDC` left on caller wallets
- no meaningful `WEGLD` or `USDC` left trapped in forwarders

### `T-45m` to `T-25m` - integrated live prep

Run the single prep helper with the current lean live defaults:

```powershell
.\run-challenge4-prep-inventory-treasury.ps1 `
  -TreasuryPemPath <path-to-treasury-pem> `
  -AutoWrapWegld
```

This now covers, in one pass:
- operator funding to `5 EGLD`
- caller funding to `0.5 EGLD`
- treasury `WEGLD` wrap and distribution
- starter `USDC` mint on shard `1`
- `USDC` distribution to the `swap2 te` pool

Live default opening inventory:
- shard `1` sync: `20` wallets with `0.15 WEGLD` each
- shard `1` TE `swap1`: `7` wallets with `0.20 WEGLD` each
- shard `1` TE `swap2`: `7` wallets with `3000000` base `USDC` each
- shard `0` async1: `33` wallets with `0.10 WEGLD` each
- shard `2` async2: `33` wallets with `0.10 WEGLD` each

### `T-25m` to `T-15m` - fast spot checks

Do not re-run full rehearsals. Only confirm that the pools are seeded and readable.

Examples:

```powershell
.\run-show-challenge4-pool.ps1
```

```powershell
.\run-challenge4-official-pool.ps1 `
  -Pool shard1_te_swap2_bias `
  -ForwarderShards 1 `
  -Action balances `
  -StartIndex 0 `
  -Limit 2
```

```powershell
.\run-challenge4-official-pool.ps1 `
  -Pool shard0_async1 `
  -ForwarderShards 0 `
  -Action balances `
  -StartIndex 0 `
  -Limit 2
```

```powershell
.\run-challenge4-official-pool.ps1 `
  -Pool shard2_async2 `
  -ForwarderShards 2 `
  -Action balances `
  -StartIndex 0 `
  -Limit 2
```

Expected outcome:
- caller wallets have gas
- `WEGLD` pools show seeded balances
- `swap2 te` pool shows `USDC`
- no need to “top up everything just in case”

### `T-15m` to `T-5m` - launch staging

Open the sustained command in a dedicated terminal and leave it ready to fire:

```powershell
.\run-challenge4-sustained.ps1 `
  -ParallelPools `
  -PoolMaxParallel 4 `
  -DrainEveryRounds 10 `
  -DurationSeconds 3600
```

Recommended terminal layout:
- terminal 1: sustained run only
- terminal 2: read-only spot checks / tx inspection
- terminal 3: emergency drains only

At this stage:
- do not run extra swap tests
- do not re-run prep unless there is clear evidence a pool is missing inventory

### `T-0` - launch

Start the sustained run:

```powershell
.\run-challenge4-sustained.ps1 `
  -ParallelPools `
  -PoolMaxParallel 4 `
  -DrainEveryRounds 10 `
  -DurationSeconds 3600
```

Target live lane mix:
- shard `1`: `blindSync`
- shard `1`: `blindTransfExec` via `swap1` and `swap2`
- shard `0`: `blindAsyncV1`
- shard `2`: `blindAsyncV2`

### `T+10m`, `T+20m`, `T+30m`, `T+45m` - monitor, do not thrash

Preferred checks:
- verify the sustained runner is still advancing rounds
- inspect only if there is a visible failure burst
- keep drains on the configured schedule unless a real issue demands intervention

Avoid:
- ad hoc pool reshuffles
- re-running prep mid-flight
- changing `PoolMaxParallel` without a clear reason

### `T+60m` - closeout

If the sustained runner exits cleanly, run a final explicit drain pass:

```powershell
.\run-challenge4-drain.ps1 -Shard 0
.\run-challenge4-drain.ps1 -Shard 1
.\run-challenge4-drain.ps1 -Shard 2
```

If the runner stops early or the network gets noisy, do the same final drain before any post-run assessment.

## Fallback Timing

Use this only if the integrated prep partially lands and funding already succeeded.

Retry token seeding without replaying the whole funding flow:

```powershell
.\run-challenge4-prep-inventory-treasury.ps1 `
  -TreasuryPemPath <path-to-treasury-pem> `
  -AutoWrapWegld `
  -SkipWegldBalanceCheck `
  -SkipCallerFunding `
  -SkipOperatorFunding `
  -ConfirmTimeoutSeconds 600
```

Use this when:
- callers and operators already have the intended `EGLD`
- only `WEGLD` or `USDC` distribution needs another pass
- the BoN indexing path is lagging and we want a patient retry instead of a full reset

## Files

Shard config templates:
- [shard0.toml](../configs/challenge4/shard0.toml)
- [shard1.toml](../configs/challenge4/shard1.toml)
- [shard2.toml](../configs/challenge4/shard2.toml)

State files written automatically after deploy:
- `state/challenge4/forwarder-blind-shard0.toml`
- `state/challenge4/forwarder-blind-shard1.toml`
- `state/challenge4/forwarder-blind-shard2.toml`

Caller wallet helpers:
- [run-gen-challenge4-caller-wallets.ps1](../run-gen-challenge4-caller-wallets.ps1)
- [run-fund-challenge4-callers.ps1](../run-fund-challenge4-callers.ps1)
- [run-fund-challenge4-operators.ps1](../run-fund-challenge4-operators.ps1)
- [run-challenge4-caller.ps1](../run-challenge4-caller.ps1)
- [run-challenge4-pool.ps1](../run-challenge4-pool.ps1)
- [run-challenge4-wrap-caller.ps1](../run-challenge4-wrap-caller.ps1)
- [run-challenge4-wrap-pool.ps1](../run-challenge4-wrap-pool.ps1)
- [run-challenge4-prep-wraps.ps1](../run-challenge4-prep-wraps.ps1)
- [run-challenge4-prep-wegld-treasury.ps1](../run-challenge4-prep-wegld-treasury.ps1)
- [run-challenge4-prep-inventory-treasury.ps1](../run-challenge4-prep-inventory-treasury.ps1)
- [run-challenge4-phase1.ps1](../run-challenge4-phase1.ps1)
- [run-challenge4-phase2.ps1](../run-challenge4-phase2.ps1)
- [run-challenge4-phase1-launch.ps1](../run-challenge4-phase1-launch.ps1)
- [run-challenge4-phase1-postdrain.ps1](../run-challenge4-phase1-postdrain.ps1)
- [run-challenge4-sustained.ps1](../run-challenge4-sustained.ps1)
- [run-sweep-challenge4-reset.ps1](../run-sweep-challenge4-reset.ps1)
- [run-show-challenge4-pool.ps1](../run-show-challenge4-pool.ps1)
- [run-gen-challenge4-pool-manifests.ps1](../run-gen-challenge4-pool-manifests.ps1)
- [wallet-allocation.challenge4-callers.json](../configs/challenge4/wallet-allocation.challenge4-callers.json)

## Before Running Anything

1. Update `wallet_pem` in each shard config file.
2. Confirm the DEX pair address is `erd1qqqqqqqqqqqqqpgqeel2kumf0r8ffyhth7pqdujjat9nx0862jpsg2pqa`.
3. Confirm `contract_code_path = 'forwarder-blind-bon.wasm'` in each shard config.
4. Confirm the shard wallet really belongs to the intended shard.
5. Ensure each wallet has EGLD for deploy gas and wrap operations.

## Generic Interactor

Generic runner:
- [run-challenge4-interactor.ps1](../run-challenge4-interactor.ps1)

Example:

```powershell
.\run-challenge4-interactor.ps1 -Shard 1 -InteractorArgs @("deploy")
```

## Deploy

Deploy the forwarder on each shard:

```powershell
.\run-challenge4-deploy.ps1 -Shard 0
.\run-challenge4-deploy.ps1 -Shard 1
.\run-challenge4-deploy.ps1 -Shard 2
```

After each deploy, record the contract address from the command output and from the matching state file.

Current sanctioned deploy addresses:

- shard `0`: `erd1qqqqqqqqqqqqqpgqqpvpk2zxn424w487g6g0fwgmp0mvhsmargvqr0p8v3`
- shard `1`: `erd1qqqqqqqqqqqqqpgqxcdu2hcry22q2ep9qdxav6527lzw4y40rphsrsh2up`
- shard `2`: `erd1qqqqqqqqqqqqqpgqctzj4st2dxwhlsjaqms4ath0fragw8lx3r9q8ykwxs`

## Wrap EGLD

Wrap `1 EGLD` on each shard wallet:

```powershell
.\run-challenge4-wrap.ps1 -Shard 0 -Amount 1000000000000000000
.\run-challenge4-wrap.ps1 -Shard 1 -Amount 1000000000000000000
.\run-challenge4-wrap.ps1 -Shard 2 -Amount 1000000000000000000
```

## Drain

Drain trapped WEGLD and USDC from a deployed forwarder back to the owner wallet:

```powershell
.\run-challenge4-drain.ps1 -Shard 0
.\run-challenge4-drain.ps1 -Shard 1
.\run-challenge4-drain.ps1 -Shard 2
```

Use this after:
- cross-shard `blindAsyncV1`
- cross-shard `blindAsyncV2`
- any `blindTransfExec`

## Swap Helpers

WEGLD -> USDC:

```powershell
.\run-challenge4-swap1.ps1 -Shard 1 -Method direct -WegldAmount 1000000000000000000
.\run-challenge4-swap1.ps1 -Shard 1 -Method sync -WegldAmount 1000000000000000000
.\run-challenge4-swap1.ps1 -Shard 0 -Method async1 -WegldAmount 1000000000000000000
.\run-challenge4-swap1.ps1 -Shard 2 -Method async2 -WegldAmount 1000000000000000000
.\run-challenge4-swap1.ps1 -Shard 0 -Method te -WegldAmount 1000000000000000000
```

USDC -> WEGLD:

```powershell
.\run-challenge4-swap2.ps1 -Shard 1 -Method direct -UsdcAmount 10000
.\run-challenge4-swap2.ps1 -Shard 1 -Method sync -UsdcAmount 10000
.\run-challenge4-swap2.ps1 -Shard 0 -Method async1 -UsdcAmount 10000
.\run-challenge4-swap2.ps1 -Shard 2 -Method async2 -UsdcAmount 10000
.\run-challenge4-swap2.ps1 -Shard 0 -Method te -UsdcAmount 10000
```

## Read-Only Checks

Liquidity:

```powershell
.\run-challenge4-liquidity.ps1 -Shard 1
```

Rate for `1 WEGLD`:

```powershell
.\run-challenge4-rate.ps1 -Shard 1 -WegldAmount 1000000000000000000
```

## Practical Notes

- Each shard uses the same interactor binary, but with a different config and state file.
- The deploy owner and drain caller must be the same shard wallet.
- `blind_sync` is only relevant on the same shard as the pair.
- Confirmed pair address for Challenge 4:
  - `erd1qqqqqqqqqqqqqpgqeel2kumf0r8ffyhth7pqdujjat9nx0862jpsg2pqa`
- The other address from the brief was confirmed to be an organizer copy-paste error.

## Default Swap Amounts

Current documented defaults for challenge-day operation:

- `swap1` (`WEGLD -> USDC`)
  - `0.01 WEGLD`
  - base amount: `10000000000000000`

- `swap2` (`USDC -> WEGLD`)
  - `10000`
  - practical amount: `0.01 USDC` if using `6` decimals

Why these defaults:

- both values are already validated on our current setup
- they are small enough to preserve working capital and make drain cycles easier
- they are still large enough to avoid unnecessary tiny-amount edge cases

We can test lower amounts later, but these are the safe baseline values to document and operate with for now.

For the threshold-first caller plan, the preferred prep path is now:
- wrap larger WEGLD chunks at treasury
- distribute WEGLD directly to the caller pools
- keep wallet-by-wallet wrapping as fallback only

The treasury WEGLD preparation helper uses these per-wallet working amounts:

- shard `1` sync callers:
  - `0.15 WEGLD` per wallet
- shard `1` TE `swap1` callers:
  - `0.20 WEGLD` per wallet
- shard `0` async1 callers:
  - `0.10 WEGLD` per wallet
- shard `2` async2 callers:
  - `0.10 WEGLD` per wallet
- shard `1` TE `swap2` callers:
  - currently `0` in the wrap prep helper

This is intentional: the helper is sized for challenge-day threshold work, not for a single smoke-test call.

For lower-interruption sustained runs, use treasury-side inventory seeding instead of caller-side wraps:

- [run-challenge4-prep-inventory-treasury.ps1](../run-challenge4-prep-inventory-treasury.ps1)

Revised live-start inventory targets:

- `WEGLD`
  - shard `1` sync: `0.15 WEGLD` per wallet
  - shard `1` TE `swap1`: `0.20 WEGLD` per wallet
  - shard `0` async1: `0.10 WEGLD` per wallet
  - shard `2` async2: `0.10 WEGLD` per wallet
- `USDC`
  - shard `1` TE `swap2`: `3000000` base per wallet initial seed

This keeps the opening capital low while keeping all four call types live from the start.

Important operational caveat:

- the treasury `WEGLD` balance/indexing path on the BoN API can lag or return incomplete data
- the prep helper now supports:
  - `-SkipWegldBalanceCheck`
  - `-SkipCallerFunding`
  - `-SkipOperatorFunding`
- in practice this means:
  - use the full prep flow after a clean reset
  - if funding already succeeded and only token seeding needs a retry, skip the funding steps and extend confirmation timeout instead of replaying everything blindly

## Current Tested Baseline

Validated on March 25, 2026:

- sanctioned deploy artifact in use:
  - wasm SHA256 `808f5beb4edd726b887cd387c474201e5d288ba65363bed68fd58d0f524eb19b`

- direct `swap1` on shard `1`: works
- `swap1 sync` on shard `1`: works
- `swap1 async1` on shard `0`: works, then `drain` works
- `swap1 async2` on shard `2`: works, then `drain` works
- `swap1 te` on shard `1`: works, then `drain` works
- `swap2 te` on shard `1`: works, then `drain` works
- `swap1 te` on shard `0`: not a clean lane right now
  - sanctioned-binary retest still failed on-chain as `user error`
- `swap1 te` on shard `2`: not a clean lane right now
  - sanctioned-binary retest still failed on-chain as `user error`
- `swap2 te` on shards `0` and `2`: not a clean lane right now
  - sanctioned-binary retest still failed on-chain as `user error`

Current operator recommendation:

1. use shard `1` direct swap and `blindSync` as the clean same-shard lane
2. use shards `0` and `2` with `blindAsyncV1` / `blindAsyncV2` as the cross-shard lanes
3. use shard `1` `swap1 te` and `swap2 te` as the validated `blindTransfExec` lanes
4. if needed, let shard `0` and `2` wallets call the shard `1` forwarder for additional TE volume while keeping the forwarder-to-pair path same-shard
5. treat true forwarder-cross-shard `blindTransfExec` as non-baseline on this pair until proven otherwise

Investigation note:

- a successful organizer `swap2 te` example was decoded and its payload shape matches our local `swap2_te()` builder
- a second successful `swap1 te` example was decoded from:
  - `6c8a4dd28eeaa6ae775a1a460ac261fe508b9f5cfbf6b8461e141db3fa41315a`
- its payload also matches our local `swap1_te()` builder
- it is useful because it confirms a successful `WEGLD -> USDC` `blindTransfExec` structure
- however, it is still a shard `1` forwarder calling the shard `1` pair
- so the unresolved issue remains the true cross-shard forwarder-to-pair runtime behavior, not the basic payload template

## Live Scoring Priorities

1. get all four call types above `300` successful calls
2. get total successful calls above `1,500`
3. race to `2,500` for the milestone bonus
4. then maximize total successful calls

This order matters more than chasing one high-throughput lane too early.

## Suggested Live Matrix

- shard `1`
  - primary: `swap1 sync`
  - optional: `swap1 te`, `swap2 te`
  - drain when using any `te` lane

- shard `0`
  - primary: `swap1 async1`
  - optional: `swap2 async1`
  - drain after each controlled batch or phase

- shard `2`
  - primary: `swap1 async2`
  - optional: `swap2 async2`
  - drain after each controlled batch or phase

## Threshold Plan

- `blindSync`
  - shard `1`
  - target first threshold: `300`

- `blindAsyncV1`
  - shard `0`
  - target first threshold: `300`

- `blindAsyncV2`
  - shard `2`
  - target first threshold: `300`

- `blindTransfExec`
  - shard `1` only
  - target first threshold: `300`
  - can be split across `swap1 te` and `swap2 te`
  - shard `0` and `2` wallets can be used as callers into the shard `1` forwarder if needed, but this should be described as wallet-cross-shard rather than true forwarder-cross-shard TE

## Execution Matrix

| Lane | Shard `1` | Shard `0` | Shard `2` | Live stance |
|------|-----------|-----------|-----------|-------------|
| `swap1 sync` | validated | n/a | n/a | primary |
| `swap1 async1` | optional | validated | optional | cross-shard primary on shard `0` |
| `swap1 async2` | optional | optional | validated | cross-shard primary on shard `2` |
| `swap1 te` | validated with drain | failed as true forwarder-cross-shard | failed as true forwarder-cross-shard | use shard `1` forwarder; wallet-cross-shard acceptable for scoring |
| `swap2 te` | validated with drain | failed as true forwarder-cross-shard | failed as true forwarder-cross-shard | use shard `1` forwarder; wallet-cross-shard acceptable for scoring |

## Recommended Sequence

1. Verify shard configs still point to `erd1qqqqqqqqqqqqqpgqeel2kumf0r8ffyhth7pqdujjat9nx0862jpsg2pqa`.
2. Keep shard `1` as the clean same-shard baseline.
3. Use shards `0` and `2` for cross-shard async lanes.
4. Keep `swap1 te` and `swap2 te` on the shard `1` forwarder.
5. If needed, route shard `0` and `2` wallets into the shard `1` forwarder for more TE volume.
6. Do not use true forwarder-cross-shard `te` in the primary scoring plan.
7. Secure `300` successes per call type before focusing on total count.
8. Drain after each planned batch, not reactively after every single call.

## Challenge-Day Command Groups

These command groups assume:

- sanctioned forwarders are already deployed
- shard configs are correct
- state files already contain the deployed forwarder addresses
- the operator wallets are funded

### Phase 0: Final Preflight

Rate / liquidity sanity:

```powershell
.\run-challenge4-liquidity.ps1 -Shard 1
.\run-challenge4-rate.ps1 -Shard 1 -WegldAmount 100000000000000000
```

Top up WEGLD working balances:

```powershell
.\run-challenge4-wrap.ps1 -Shard 0 -Amount 1000000000000000000
.\run-challenge4-wrap.ps1 -Shard 1 -Amount 1000000000000000000
.\run-challenge4-wrap.ps1 -Shard 2 -Amount 1000000000000000000
```

### Phase 1: Secure the Four Thresholds

`blindSync` baseline on shard `1`:

```powershell
.\run-challenge4-swap1.ps1 -Shard 1 -Method sync -WegldAmount 100000000000000000
```

`blindAsyncV1` baseline on shard `0`:

```powershell
.\run-challenge4-swap1.ps1 -Shard 0 -Method async1 -WegldAmount 100000000000000000
```

`blindAsyncV2` baseline on shard `2`:

```powershell
.\run-challenge4-swap1.ps1 -Shard 2 -Method async2 -WegldAmount 100000000000000000
```

`blindTransfExec` baseline on shard `1`:

```powershell
.\run-challenge4-swap1.ps1 -Shard 1 -Method te -WegldAmount 100000000000000000
.\run-challenge4-swap2.ps1 -Shard 1 -Method te -UsdcAmount 10000
```

Drain after TE / cross-shard async batches:

```powershell
.\run-challenge4-drain.ps1 -Shard 0
.\run-challenge4-drain.ps1 -Shard 1
.\run-challenge4-drain.ps1 -Shard 2
```

### Phase 2: Add TE Volume Without Unproven Forwarder-Cross-Shard TE

For TE, prefer keeping the forwarder on shard `1`. If wallet routing tooling later sends shard `0` and `2` wallets into the shard `1` forwarder, that is acceptable for scoring and still stays on the validated forwarder-to-pair path.

Primary TE commands remain:

```powershell
.\run-challenge4-swap1.ps1 -Shard 1 -Method te -WegldAmount 100000000000000000
.\run-challenge4-swap2.ps1 -Shard 1 -Method te -UsdcAmount 10000
```

Drain shard `1` on schedule:

```powershell
.\run-challenge4-drain.ps1 -Shard 1
```

### Phase 3: Push Stable Lanes

Same-shard sync:

```powershell
.\run-challenge4-swap1.ps1 -Shard 1 -Method sync -WegldAmount 100000000000000000
```

Cross-shard asyncs:

```powershell
.\run-challenge4-swap1.ps1 -Shard 0 -Method async1 -WegldAmount 100000000000000000
.\run-challenge4-swap1.ps1 -Shard 2 -Method async2 -WegldAmount 100000000000000000
```

Periodic drain:

```powershell
.\run-challenge4-drain.ps1 -Shard 0
.\run-challenge4-drain.ps1 -Shard 2
```

### Avoid During The Main Run

Do not use these as baseline scoring lanes unless we later prove them:

```powershell
.\run-challenge4-swap1.ps1 -Shard 0 -Method te -WegldAmount 100000000000000000
.\run-challenge4-swap1.ps1 -Shard 2 -Method te -WegldAmount 100000000000000000
.\run-challenge4-swap2.ps1 -Shard 0 -Method te -UsdcAmount 10000
.\run-challenge4-swap2.ps1 -Shard 2 -Method te -UsdcAmount 10000
```

## Wallet Plan

Within the `100` wallet cap:

- distribute the full `100` wallets approximately:
  - shard `1`: `34`
  - shard `0`: `33`
  - shard `2`: `33`

This is the default balanced split until more throughput testing says otherwise.

Recommended funding split:

- operator wallets:
  - `3` wallets
  - `5 EGLD` each
  - total `15 EGLD`
- caller wallets:
  - `100` wallets
  - target `0.5 EGLD` each
  - total `50 EGLD`
- opening `WEGLD` inventory:
  - shard `1` sync: `3.0 EGLD`
  - shard `1` TE `swap1`: `1.4 EGLD`
  - shard `0` async1: `3.3 EGLD`
  - shard `2` async2: `3.3 EGLD`
  - total `11.0 EGLD`
- startup `USDC` seed for `swap2 te`:
  - `7 x 3000000` base USDC
  - total `21000000` base USDC
- opening capital committed:
  - about `81 EGLD`
- treasury reserve:
  - about `419 EGLD`

Recommended generation:

```powershell
.\run-gen-challenge4-caller-wallets.ps1
```

Recommended funding:

```powershell
.\run-fund-challenge4-callers.ps1 `
  -TreasuryPemPath <path-to-treasury-pem> `
  -TargetEgld 0.5
```

Operator funding:

```powershell
.\run-fund-challenge4-operators.ps1 `
  -TreasuryPemPath <path-to-treasury-pem> `
  -TargetEgld 5.0
```

Funding dry-run:

```powershell
.\run-fund-challenge4-callers.ps1 `
  -TreasuryPemPath <path-to-treasury-pem> `
  -TargetEgld 0.5 `
  -DryRun
```

## Concrete Caller Allocation

Use the `100` caller wallets as four logical pools, while keeping the actual wallet inventory balanced by shard.

### Pool Split

- Pool A: shard `1` sync callers
  - `20` wallets
  - primary lane: `blindSync`
  - fallback lane: shard `1` `swap1 te` if sync is already safely above threshold

- Pool B: shard `1` TE callers
  - `14` wallets
  - primary lane: `blindTransfExec`
  - split between:
    - `swap1 te` (`WEGLD -> USDC`)
    - `swap2 te` (`USDC -> WEGLD`)

- Pool C: shard `0` async callers
  - `33` wallets
  - primary lane: `blindAsyncV1`

- Pool D: shard `2` async callers
  - `33` wallets
  - primary lane: `blindAsyncV2`

This maps exactly to the funded shard split:

- shard `1`: `34` wallets total = `20` sync + `14` TE
- shard `0`: `33` wallets total = async1 pool
- shard `2`: `33` wallets total = async2 pool

The current concrete pool membership is written to:

- [wallet-allocation.challenge4-callers.json](../configs/challenge4/wallet-allocation.challenge4-callers.json)

Quick inspection:

```powershell
.\run-show-challenge4-pool.ps1
.\run-show-challenge4-pool.ps1 -Pool shard1_sync
```

Generate per-pool manifests:

```powershell
.\run-gen-challenge4-pool-manifests.ps1
```

Single caller execution:

```powershell
.\run-challenge4-caller.ps1 -WalletId c4call-001 -ForwarderShard 1 -Action swap1 -Method sync -Amount 100000000000000000
```

Pool execution:

```powershell
.\run-challenge4-pool.ps1 -Pool shard1_sync -ForwarderShard 1 -Action swap1 -Method sync -Amount 100000000000000000 -WhatIf
.\run-challenge4-pool.ps1 -Pool shard0_async1 -ForwarderShard 0 -Action swap1 -Method async1 -Amount 100000000000000000
```

Caller wrap:

```powershell
.\run-challenge4-wrap-caller.ps1 -WalletId c4call-001 -Amount 100000000000000000
.\run-challenge4-wrap-pool.ps1 -Pool shard1_sync -Amount 100000000000000000 -WhatIf
.\run-challenge4-prep-wraps.ps1 -WhatIf
.\run-challenge4-prep-wraps.ps1
.\\run-challenge4-prep-wegld-treasury.ps1 -TreasuryPemPath <path-to-treasury-pem> -DryRun
.\\run-challenge4-prep-wegld-treasury.ps1 -TreasuryPemPath <path-to-treasury-pem>
.\\run-sweep-challenge4-reset.ps1 -DryRun
.\\run-sweep-challenge4-reset.ps1
```

Phase helpers:

```powershell
.\run-challenge4-phase1.ps1 -WhatIf -ParallelPools -PoolMaxParallel 4
.\run-challenge4-phase1.ps1 -ParallelPools -PoolMaxParallel 4
.\run-challenge4-phase2.ps1 -WhatIf -ParallelPools -PoolMaxParallel 4
.\run-challenge4-phase2.ps1 -ParallelPools -PoolMaxParallel 4
.\run-challenge4-phase1-launch.ps1 -WhatIf -ParallelPools -PoolMaxParallel 4
.\run-challenge4-phase1-launch.ps1 -ParallelPools -PoolMaxParallel 4
.\run-challenge4-phase1-postdrain.ps1 -WhatIf -ParallelPools -PoolMaxParallel 4
.\run-challenge4-phase1-postdrain.ps1 -ParallelPools -PoolMaxParallel 4
```

Performance guidance:
- Prefer parallel pool execution over sequential lane-by-lane execution.
- Recommended starting point: `-ParallelPools -PoolMaxParallel 4`.
- If the network stays healthy, try `-PoolMaxParallel 6` next.
- The interactor now builds once and reuses the compiled binary, which removes most of the old `cargo run` overhead.
- Pool logs now print `callerShard`, `forwarderShard`, `pairShard`, and `lane`, so the true cross-shard forwarder lanes are visible during rehearsal.
- For longer sustained runs, prefer one inventory prep plus one sustained runner instead of chaining many manual phase commands.
- After a full reset, caller wallets must be re-funded with `EGLD` before any sustained run. The current inventory prep helper now does that for both callers and operators.

### Threshold Objective By Pool

- Pool A:
  - secure `300` successful `blindSync`
- Pool B:
  - secure `300` successful `blindTransfExec`
- Pool C:
  - secure `300` successful `blindAsyncV1`
- Pool D:
  - secure `300` successful `blindAsyncV2`

### TE Direction Split

Inside Pool B, keep a simple directional split:

- `7` wallets biased to `swap1 te`
- `7` wallets biased to `swap2 te`

This is not for separate scoring categories. It is for operational balance:

- `swap1 te` consumes `WEGLD`
- `swap2 te` consumes `USDC`
- launch with both directions live by seeding a tiny starter `USDC` pool
- alternating both directions helps keep shard `1` caller balances and drain cycles healthier

### Challenge-Day Phase Use

Phase 1:

- Pool A runs `blindSync`
- Pool B runs `blindTransfExec`
- Pool C runs `blindAsyncV1`
- Pool D runs `blindAsyncV2`

Goal:

- get every call type above `300`

Operational note for TE:

- start Phase 1 with both `swap1 te` and `swap2 te` live
- keep the `swap2 te` seed intentionally small
- replenish only if the live run justifies it

Phase 2:

- keep Pool C and Pool D running steadily
- keep Pool A running if sync remains stable
- use Pool B in controlled bursts with planned drains

Goal:

- move from `1,200` threshold coverage to `1,500+`

Phase 3:

- scale the most reliable lanes:
  - Pool A `blindSync`
  - Pool C `blindAsyncV1`
  - Pool D `blindAsyncV2`
- use Pool B opportunistically when shard `1` drain cadence is under control

### Reserve / Flex Rule

We intentionally keep about `419 EGLD` in treasury reserve at launch.

Use that reserve only for:

- additional wraps
- replenishing TE lanes if they drift
- caller/operator top-ups if a specific pool becomes gas-constrained
- gas and execution comfort

Do not use it to redesign the allocation mid-run. The baseline should stay:

- `20` sync wallets on shard `1`
- `14` TE wallets on shard `1`
- `33` async1 wallets on shard `0`
- `33` async2 wallets on shard `2`

## Early-Game Timeline

Recommended launch posture:

- shard `1` `blindSync`
- shard `1` `swap1 te`
- shard `1` `swap2 te`
- shard `0` `blindAsyncV1`
- shard `2` `blindAsyncV2`

Goal:

- all four call types live from the start
- keep capital committed low
- use the first drain checkpoints for recovery, not for unlocking `swap2 te`

## Sustained Run Mode

For the live challenge, the preferred low-interruption mode is now:

1. treasury seeds `WEGLD` and `USDC` inventories to the caller pools
2. one sustained runner keeps all four call types active
3. drains happen only on periodic checkpoints

Dry-run the inventory plan:

```powershell
.\run-challenge4-prep-inventory-treasury.ps1 `
  -TreasuryPemPath <path-to-treasury-pem> `
  -CallerTargetEgld 0.5 `
  -OperatorTargetEgld 5.0 `
  -SyncWegldPerWallet 150000000000000000 `
  -TeSwap1WegldPerWallet 200000000000000000 `
  -Async1WegldPerWallet 100000000000000000 `
  -Async2WegldPerWallet 100000000000000000 `
  -TeSwap2UsdcPerWallet 0 `
  -DryRun
```

Real opening prep with automatic treasury WEGLD top-up if needed:

```powershell
.\run-challenge4-prep-inventory-treasury.ps1 `
  -TreasuryPemPath <path-to-treasury-pem> `
  -AutoWrapWegld
```

This single helper now also:

- funds callers to `0.5 EGLD`
- funds operators to `5 EGLD`
- wraps only the opening `WEGLD` inventory
- mints the shard `1` starter `USDC` pool from `5.0 WEGLD`
- distributes `3000000` base USDC to each `swap2 te` wallet

If the treasury `WEGLD` balance check is lagging but you know inventory is available or you are retrying token seeding after funding already succeeded:

```powershell
.\run-challenge4-prep-inventory-treasury.ps1 `
  -TreasuryPemPath <path-to-treasury-pem> `
  -SkipWegldBalanceCheck `
  -SkipCallerFunding `
  -SkipOperatorFunding `
  -ConfirmTimeoutSeconds 600
```

Then start the sustained run:

```powershell
.\run-challenge4-sustained.ps1 `
  -ParallelPools `
  -PoolMaxParallel 4 `
  -DrainEveryRounds 10 `
  -DurationSeconds 3600
```

Recommended live posture:

- start with `-PoolMaxParallel 4`
- keep all four lanes active from the start
- drain every `10` rounds unless on-chain behavior suggests tightening that cadence
- use the threshold-first manual phase helpers only as fallback or for debugging, not as the main live workflow

Validated sustained rehearsal notes:

- a full Challenge 4 reset now exists and clears:
  - forwarder balances via `drain`
  - caller and operator `WEGLD`
  - caller and operator `USDC`
  - caller and operator `EGLD`
- after reset, the prep flow must restore:
  - caller `EGLD`
  - operator `EGLD`
  - caller `WEGLD`
  - TE `swap2` caller `USDC`
- one-round sustained validation succeeded with:
  - shard `1` `sync`
  - shard `1` `swap1 te`
  - shard `1` `swap2 te`
  - shard `0` `async1`
  - shard `2` `async2`
  - final drain cycle successful on all three forwarders

Current practical baseline:

```powershell
.\run-challenge4-sustained.ps1 `
  -ParallelPools `
  -PoolMaxParallel 4 `
  -DrainEveryRounds 10 `
  -DurationSeconds 3600
```

## Real Run Result

Verified from BoN API transaction history for the real run window:

- start:
  - `2026-03-26T10:53:00Z`
- assessment window:
  - through `2026-03-26T11:53:30Z`

Successful calls by type:

- `blindSync`: `869`
- `blindAsyncV1`: `1426`
- `blindAsyncV2`: `1429`
- `blindTransfExec`: `315`

Total successful smart contract calls:

- `4039`

Practical conclusion:

- all four call types cleared the required `300` successful calls
- total successful calls cleared both:
  - the `1500` minimum
  - the `2500` milestone threshold

`blindTransfExec` direction split:

- `swap1 te` (`WEGLD -> USDC`) successes: `308`
- `swap2 te` (`USDC -> WEGLD`) successes: `7`
- `swap2 te` failures: `301`

Interpretation:

- `swap2 te` was not a viable production lane during this run
- however, Challenge 4 scoring is per call type, not per swap direction
- `blindTransfExec` still cleared the `300` threshold because `swap1 te` alone carried the lane

Operational takeaway:

- keep `swap1 te` as the primary `blindTransfExec` lane
- treat `swap2 te` as unstable until the failure mode is understood and fixed
- the sustained runner itself still needs hardening against config/state isolation failures under long concurrent execution

Official interactor sidecar:

- the organizer update did not change the sanctioned wasm or ABI
- we now keep an isolated official-style runner at:
  - [dex-interactor-official](../contracts/forwarder-blind/dex-interactor-official)
- this is for assessment only, not a replacement for the current validated Challenge 4 orchestration

Useful assessment commands:

```powershell
.\run-challenge4-official-pool.ps1 `
  -Pool shard0_async1 `
  -ForwarderShards 1 `
  -Action balances
```

```powershell
.\run-challenge4-official-pool.ps1 `
  -Pool shard0_async1 `
  -ForwarderShards 1 `
  -Action swap1 `
  -Method te `
  -Amount 10000000000000000 `
  -StartIndex 0 `
  -Limit 1 `
  -WhatIf
```

Interpretation:

- `ForwarderShards 1` with a shard `0` or `2` pool tests wallet-cross-shard, forwarder-same-shard TE
- `ForwarderShards 0` or `2` tests true forwarder-cross-shard lanes
- use small slices first with `-StartIndex` and `-Limit`
