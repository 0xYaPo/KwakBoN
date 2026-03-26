# Challenge 4 Prep Plan

Challenge 4 tests contract-to-contract forwarding against the xExchange `WEGLD/USDC` pair.

Reference brief:
- [Challenge4.md](../docs/Challenge4.md)

## Current Reading

- Challenge date: `March 26, 2026`
- Challenge window: `16:00-17:00 UTC`
- Each guild needs one funded wallet in each shard: `0`, `1`, `2`
- Each guild deploys one forwarder contract per shard
- The forwarder then calls the xExchange pair using one of four call types
- Tokens can remain trapped in the forwarder depending on call type and shard relation, so drain tooling is mandatory
- Full brief now confirms:
  - max `100` wallets total
  - budget `500 EGLD`
  - all four call types required
  - minimum `300` successful calls per call type
  - minimum `1,500` successful calls total
  - first three guilds to `2,500` successful calls get milestone bonus

## Confirmed Pair Address

Confirmed final target pair:

- `erd1qqqqqqqqqqqqqpgqeel2kumf0r8ffyhth7pqdujjat9nx0862jpsg2pqa`

The other address present in the organizer brief:

- `erd1qqqqqqqqqqqqqpgqr8n2kjqhrupcrsceevkv6yydtjsgacvuqqqs23m8n6`

was confirmed to be a copy-paste error from the organizers.

All current local configs, deployed tests, and successful rehearsals already use the correct pair address above.

## Sanctioned Artifact Status

The organizer-sanctioned deploy artifact is now pinned locally in the interactor workspace:

- [forwarder-blind-bon.wasm](../contracts/forwarder-blind/dex-interactor/forwarder-blind-bon.wasm)
- [forwarder-blind-bon.abi.json](../contracts/forwarder-blind/dex-interactor/forwarder-blind-bon.abi.json)

Verified hashes:

- sanctioned wasm SHA256:
  - `808f5beb4edd726b887cd387c474201e5d288ba65363bed68fd58d0f524eb19b`
- sanctioned abi SHA256:
  - `2b4fc479fe377b37fcf61c2eb8b43c2a14c9b1be4de3e0139253b3c3b235a493`

The local interactor deploy path is now configured to read the sanctioned wasm bytes directly instead of deploying our local build output.

## Four Call Types

The upstream `forwarder-blind` contract exposes:

- `blind_sync`
- `blind_async_v1`
- `blind_async_v2`
- `blind_transf_exec`

Observed behavior from the pre-brief and upstream contract README:

- `blind_sync`
  - same-shard only
  - back-transfers go directly to caller
  - no drain required

- `blind_async_v1`
  - same-shard and cross-shard
  - same-shard back-transfers go to caller
  - cross-shard funds can remain in forwarder
  - drain required cross-shard

- `blind_async_v2`
  - same-shard and cross-shard
  - same-shard back-transfers go to caller
  - cross-shard funds can remain in forwarder
  - drain required cross-shard

- `blind_transf_exec`
  - same-shard and cross-shard
  - fire-and-forget, no callback
  - funds remain in forwarder
  - drain required always

## Upstream Reference

Vendored local contract:
- [contracts/forwarder-blind](../contracts/forwarder-blind)

Original upstream source:
- `mx-sdk-rs/contracts/feature-tests/composability/forwarder-blind`

Useful upstream notes:
- contract README confirms `blind_sync`, `blind_async_v1`, `blind_async_v2`, `blind_transf_exec`, and `drain`
- `drain` is `only_owner`
- the included `dex-interactor` already supports:
  - `deploy`
  - `wrap`
  - `swap1` (`WEGLD -> USDC`) via `direct|sync|async1|async2|te`
  - `swap2` (`USDC -> WEGLD`) via `direct|sync|async1|async2|te`
  - `get-rate`
  - `get-liquidity`
  - `drain`

## Local Interactor Patch

The vendored `dex-interactor` is now patched for real operator wallets instead of the upstream test wallet.

Current local behavior:

- wallet PEM path is loaded from `wallet_pem` in `config.toml`
- config file path can be overridden with `FORWARDER_BLIND_CONFIG`
- state file path can be overridden with `FORWARDER_BLIND_STATE`
- workspace path can be overridden with `FORWARDER_BLIND_WORKSPACE`
- default workspace path now points to our vendored copy:
  - `contracts/forwarder-blind/dex-interactor`

This gives us a clean per-shard operating model without duplicating the project:

- one config file per shard wallet
- one state file per deployed forwarder
- same binary, different env overrides
- sanctioned artifact path pinned per shard config

## Recommended Strategy

Primary recommendation:

- use the upstream `forwarder-blind` contract unchanged first
- use the upstream `dex-interactor` as the baseline operator path
- only consider custom contract changes after we have a stable deploy/drain workflow

Reason:

- the challenge is new and technically deeper than prior ones
- correctness and speed of setup matter more than clever contract changes on day one
- the official interactor already covers the exact call families we need

## Challenge-Day Strategy

The challenge-day plan should separate:

- clean same-shard lanes
- validated cross-shard lanes
- experimental lanes that should not carry the score plan
- threshold/eligibility work
- scaling work after eligibility is secured

### Scoring Implication

The scoring model changes the operating priority:

1. reach `300` successful calls in each of the four call types
2. cross `1,500` total successful calls
3. race for `2,500` total successful calls
4. only then optimize for maximum total count

This means the correct early-game objective is not raw throughput on one lane. It is eligibility.

### Timing Posture

The live runbook is now intentionally simple:

1. `T-60m` to `T-45m`
   - reset wallets and forwarders
2. `T-45m` to `T-25m`
   - run the integrated inventory prep once
3. `T-25m` to `T-15m`
   - do only light balance spot checks
4. `T-15m` to `T-0`
   - stage the sustained command and stop changing inputs
5. `T-0` to `T+60m`
   - run the sustained profile with scheduled drains

The detailed command timeline lives in:
- [Challenge4-Operator-Checklist.md](../docs/Challenge4-Operator-Checklist.md)

Important scoring interpretation:

- the brief defines same-shard vs cross-shard by the shard of the deployed forwarder relative to the pair on shard `1`
- so a wallet on shard `0` or `2` calling a forwarder on shard `1` is still a same-shard forwarder-to-pair TE path
- however, if those transactions succeed on-chain, they should still count toward the required `300` successful `blindTransfExec` calls
- we should describe those lanes honestly as wallet-cross-shard / forwarder-same-shard TE

### Strategy Matrix

| Shard | Relative to pair | Primary lane | Secondary lane | Avoid as baseline |
|------|-------------------|--------------|----------------|-------------------|
| `1` | same-shard | `blindSync` | `blindTransfExec` for `USDC -> WEGLD` or `WEGLD -> USDC`, including wallet-cross-shard feeds into the shard `1` forwarder | none of the tested same-shard lanes |
| `0` | cross-shard | `blindAsyncV1` | `blindAsyncV2` | `blindTransfExec` in both tested directions |
| `2` | cross-shard | `blindAsyncV2` | `blindAsyncV1` | `blindTransfExec` in both tested directions |

### Execution Matrix

| Direction | Shard `1` same-shard | Shard `0` cross-shard | Shard `2` cross-shard | Drain expectation |
|----------|-----------------------|-----------------------|-----------------------|------------------|
| `swap1 sync` (`WEGLD -> USDC`) | validated | not applicable | not applicable | none |
| `swap1 async1` (`WEGLD -> USDC`) | not primary | validated | not primary | yes cross-shard |
| `swap1 async2` (`WEGLD -> USDC`) | not primary | not primary | validated | yes cross-shard |
| `swap1 te` (`WEGLD -> USDC`) | validated | failed | failed | yes if same-shard only |
| `swap2 te` (`USDC -> WEGLD`) | validated | failed | failed | yes if same-shard only |

For TE planning, keep these three cases distinct:

- true same-shard TE:
  - shard `1` wallet -> shard `1` forwarder -> shard `1` pair
- wallet-cross-shard TE:
  - shard `0` or `2` wallet -> shard `1` forwarder -> shard `1` pair
- true forwarder-cross-shard TE:
  - shard `0` or `2` forwarder -> shard `1` pair

Only the first two are currently usable in our scoring plan.

### `blindTransfExec` Investigation Notes

We now have two organizer-aligned successful `blindTransfExec` references whose payloads match our local builders.

Known-good `swap2 te` reference:

- tx: `b6fc5737bee25cd848d1530682ad77ded88210b3e335a7f9a13e353ab709bcbe`
- input token: `USDC-c76f1f`
- forwarded call: `swapTokensFixedInput(WEGLD-bd4d79, 1)`
- result pattern:
  - swap fee goes to fee collector
  - resulting `WEGLD` lands in the forwarder
  - later drain is required

Known-good `swap1 te` reference:

- tx: `6c8a4dd28eeaa6ae775a1a460ac261fe508b9f5cfbf6b8461e141db3fa41315a`
- input token: `WEGLD-bd4d79`
- input amount: `0.2 WEGLD`
- forwarded call: `swapTokensFixedInput(USDC-c76f1f, 1)`
- API-confirmed properties:
  - top-level `function = blindTransfExec`
  - tx `status = success`
  - wallet sender shard = `2`
  - forwarder receiver shard = `1`
  - forwarded destination = confirmed pair `erd1qqqqqqqqqqqqqpgqeel2kumf0r8ffyhth7pqdujjat9nx0862jpsg2pqa`
  - resulting `USDC` lands in the forwarder on shard `1`

Important interpretation:

- both known-good examples support that our `swap1 te` and `swap2 te` payload construction is correct
- both examples still route from a shard `1` forwarder into the shard `1` pair
- so they validate same-shard forwarder-to-pair `blindTransfExec`
- they do not yet resolve the remaining discrepancy on true cross-shard forwarder-to-pair `blindTransfExec`

### Practical Use

- shard `1`
  - use for the cleanest throughput lane
  - `blindSync` is the best same-shard contract path
  - `swap1 te` and `swap2 te` are usable as additional same-shard lanes when we intentionally manage drain
  - shard `0` and `2` wallets can also feed the shard `1` forwarder if we need more TE volume while staying on a validated forwarder-to-pair path

- shard `0`
  - use as a cross-shard async lane
  - prefer `blindAsyncV1`
  - drain after controlled batches or after a time slice, not after every single tx

- shard `2`
  - use as the second cross-shard async lane
  - prefer `blindAsyncV2`
  - same drain discipline as shard `0`

### Drain Model

The contract strategy only stays usable if drains are predictable.

- `blindSync`
  - no routine drain expected

- same-shard `blindTransfExec`
  - drain required
  - validated in both tested directions
  - use in planned bursts, not mixed randomly with other lanes

- cross-shard `blindAsyncV1` / `blindAsyncV2`
  - expect funds to accumulate in the forwarder
  - drain between batches or phases
  - do not let trapped balances pile up across too many runs

- cross-shard `blindTransfExec`
  - do not use in the scoring baseline

### Recommended Operating Shape

If scoring rewards completed valid forwarded calls without special weighting, the safest production shape is:

1. shard `1` same-shard `blindSync`
2. shard `0` cross-shard `blindAsyncV1`
3. shard `2` cross-shard `blindAsyncV2`
4. optional shard `1` `swap1 te` and `swap2 te` lanes only if we intentionally budget drain operations
5. if needed, route shard `0` and `2` wallets into the shard `1` forwarder for additional TE count without relying on unproven forwarder-cross-shard TE

### Live Scoring Plan

Recommended phase plan for the `60` minute window:

#### Phase 0: Final sanity before `16:00 UTC`

- confirm the sanctioned forwarder binaries are the deployed ones
- confirm shard configs still point to `erd1qqqqqqqqqqqqqpgqeel2kumf0r8ffyhth7pqdujjat9nx0862jpsg2pqa`
- confirm the three forwarders and owner wallets are ready
- pre-stage drain commands for shard `0`, `1`, and `2`

#### Phase 1: Eligibility first

Goal:

- `300` successful `blindSync`
- `300` successful `blindAsyncV1`
- `300` successful `blindAsyncV2`
- `300` successful `blindTransfExec`

Recommended lane mapping:

- `blindSync`
  - shard `1`
- `blindAsyncV1`
  - shard `0`
- `blindAsyncV2`
  - shard `2`
- `blindTransfExec`
  - shard `1` only
  - split across `swap1 te` and `swap2 te` if useful
  - shard `0` and `2` wallets may be used as callers into the shard `1` forwarder if we need more wallet-level concurrency while preserving a validated TE path

This gets us to `1,200` successful calls while preserving eligibility.

#### Phase 2: Push to `1,500`

After all four types are above `300`, add volume on the most stable lanes:

- shard `1` `blindSync`
- shard `0` `blindAsyncV1`
- shard `2` `blindAsyncV2`

This phase should take us safely above the `1,500` minimum.

#### Phase 3: Milestone race to `2,500`

Once eligibility is secured, the milestone race matters more than perfect lane symmetry.

Preferred lanes to scale:

1. shard `1` `blindSync`
2. shard `0` `blindAsyncV1`
3. shard `2` `blindAsyncV2`
4. shard `1` `blindTransfExec` only if drain cadence is under control

#### Phase 4: Remainder of the hour

Optimize for total successful calls, but do not starve the drain flow:

- keep cross-shard async lanes moving
- keep shard `1` same-shard lanes efficient
- avoid introducing cross-shard `te` just to chase count

### Wallet Budgeting Within The `100` Wallet Cap

The full brief allows up to `100` wallets total, not just three operators.

That suggests a two-layer model:

- `3` owner/operator wallets
  - one per shard
  - deploy and drain authority
- additional caller wallets
  - up to the remaining wallet budget
  - funded and assigned by lane

Suggested starting allocation model across the full `100`-wallet cap:

- shard `1` same-shard pool: `34` wallets
- shard `0` cross-shard pool: `33` wallets
- shard `2` cross-shard pool: `33` wallets

Reason:

- simple and balanced
- sums exactly to `100`
- aligns with the three-lane primary structure
- shard `1` gets one extra wallet because it carries both `blindSync` and `blindTransfExec`

This should be treated as the default allocation unless further throughput testing suggests a better split.

### Capital Model

The `500 EGLD` budget covers:

- gas
- wrapped WEGLD / swapped USDC working capital

Because swaps recycle value, the main burn is gas. The operational risk is not running out of principal first; it is:

- failing one of the four type thresholds
- trapping balances without draining
- wasting time on unstable lanes

Revised live-start funding model:

- operators:
  - `3 x 5 EGLD = 15 EGLD`
- callers:
  - `100 x 0.5 EGLD = 50 EGLD`
- opening caller `WEGLD` inventory:
  - shard `1` sync: `3.0 EGLD`
  - shard `1` TE `swap1`: `1.4 EGLD`
  - shard `0` async1: `3.3 EGLD`
  - shard `2` async2: `3.3 EGLD`
  - total `11.0 EGLD`
- startup `USDC` seed for `swap2 te`:
  - `7 x 3000000` base USDC
  - minted from a shard `1` operator `5.0 WEGLD -> USDC` starter swap

Opening capital committed:

- about `81 EGLD`

Treasury reserve retained at launch:

- about `419 EGLD`

This is the preferred model over the older “fund every caller heavily” approach. It keeps the launch simple and preserves most of the budget for adaptation during the hour.

Implementation note:

- [run-challenge4-prep-inventory-treasury.ps1](../run-challenge4-prep-inventory-treasury.ps1) now covers the full opening prep in one step:
  - caller funding
  - operator funding
  - treasury `WEGLD` wrap if needed
  - opening `WEGLD` distribution
  - shard `1` starter `USDC` mint
  - `swap2 te` pool `USDC` distribution

### What Not To Build Around

- cross-shard `blindTransfExec`
- any lane that depends on ambiguous pair-address assumptions
- frequent ad hoc drains without a phase plan

The main lesson from testing is simple:

- async cross-shard works and is recoverable with drain
- same-shard sync is clean
- transfer-execute is shard-sensitive in our testing
- same-shard works in both tested directions
- cross-shard failed in both tested directions

## Validated Findings - March 25, 2026

Operator setup completed:

- one funded operator wallet per shard
- one deployed forwarder per shard
- contract artifact build working locally
- wrap flow working on all three wallets

Current sanctioned-binary forwarders:

- shard `0`: `erd1qqqqqqqqqqqqqpgqqpvpk2zxn424w487g6g0fwgmp0mvhsmargvqr0p8v3`
- shard `1`: `erd1qqqqqqqqqqqqqpgqxcdu2hcry22q2ep9qdxav6527lzw4y40rphsrsh2up`
- shard `2`: `erd1qqqqqqqqqqqqqpgqctzj4st2dxwhlsjaqms4ath0fragw8lx3r9q8ykwxs`

Validated call paths:

- shard `1` direct `WEGLD -> USDC`
  - success
  - `1 WEGLD -> 4490794 USDC`

- shard `1` `blindSync`
  - success
  - same-shard path works as expected

- shard `0` cross-shard `blindAsyncV1`
  - success
  - USDC stayed in the forwarder
  - `drain` recovered the trapped USDC successfully

- shard `2` cross-shard `blindAsyncV2`
  - success
  - USDC stayed in the forwarder
  - `drain` recovered the trapped USDC successfully

- shard `1` same-shard `blindTransfExec` for `USDC -> WEGLD`
  - success
  - `40000 USDC -> 8854253696473274 WEGLD`
  - resulting WEGLD stayed in the forwarder
  - `drain` recovered the trapped WEGLD successfully

- shard `1` same-shard `blindTransfExec` for `WEGLD -> USDC`
  - success after controlled retest with confirmed fresh WEGLD balance
  - `0.1 WEGLD -> 449039 USDC`
  - resulting USDC stayed in the forwarder
  - `drain` recovered the trapped USDC successfully

- sanctioned-binary retest of cross-shard `blindTransfExec`
  - shard `0` `swap1 te`: `1bb01773d5b68f126d6d67f6a37c774edf5a22e436811743c08f9f5e5adc1280`
  - shard `2` `swap1 te`: `a20ee339b9ea4fb2d7bf6147abf385bb5bc3750742ad8dc89a8d6ce81cdd898c`
  - shard `0` `swap2 te`: `e36688dd6cd3db92538ccb72fdd9d4f0d1464d6b8678b96bf02cd3a3dff76bd1`
  - shard `2` `swap2 te`: `393684d2d04c4d0f3676c4cd56d4946cc0f9a5e50ea2a43bab7830381a649a47`
  - all four still failed on-chain as `user error`
  - no residual WEGLD or USDC remained in the sanctioned forwarders afterwards

Problematic path:

- cross-shard `blindTransfExec`
  - still not usable in our testing even after switching to the sanctioned binary
  - failed in both tested directions on shards `0` and `2`
  - practical conclusion: keep it out of the scoring baseline until we obtain a meaningful successful organizer example and reproduce it

Sustained-run preparation findings:

- a full Challenge 4 reset flow now exists:
  - [run-sweep-challenge4-reset.ps1](../run-sweep-challenge4-reset.ps1)
  - it drains forwarders, then sweeps caller/operator `WEGLD`, `USDC`, and `EGLD`
- after such a reset, caller wallets are genuinely clean and therefore have no gas balance left
- the sustained inventory prep flow must therefore restore:
  - caller `EGLD`
  - operator `EGLD`
  - caller `WEGLD`
  - TE `swap2` caller `USDC`
- [run-challenge4-prep-inventory-treasury.ps1](../run-challenge4-prep-inventory-treasury.ps1) now includes caller/operator funding before token seeding
- the treasury `WEGLD` balance/indexing path on the BoN API is flaky enough that:
  - `-SkipWegldBalanceCheck` is sometimes necessary for rehearsal work
  - reruns should target only the incomplete prep slice where possible, not blindly replay the full prep

Sustained-run execution finding:

- a one-round sustained rehearsal succeeded with the live intended lane mix:
  - shard `1` `blindSync`
  - shard `1` `swap1 te`
  - shard `1` `swap2 te`
  - shard `0` `blindAsyncV1`
  - shard `2` `blindAsyncV2`
- final drain cycle succeeded on all three forwarders
- this validates the sustained execution model itself even though treasury `WEGLD` visibility remained noisy during prep

Current live recommendation:

1. if starting from a clean reset, run full inventory prep
2. prefer the sustained runner over many manual phase transitions
3. use:
   - `-ParallelPools`
   - `-PoolMaxParallel 4`
   - `-DrainEveryRounds 10`
4. only fall back to manual phase helpers for debugging or recovery

## `blindTransfExec` Investigation Note

We now have a meaningful successful organizer example for `swap2 te` and can compare its payload shape with our local builder.

Organizer example:

- tx hash:
  - `b6fc5737bee25cd848d1530682ad77ded88210b3e335a7f9a13e353ab709bcbe`
- raw input:
  - `ESDTTransfer@555344432d633736663166@9c40@626c696e645472616e736645786563@00000000000000000500ce7eab736978ce9492ebbf8206f252eacb333cfa5483@73776170546f6b656e734669786564496e707574@5745474c442d626434643739@01`

Decoded structure:

- outer token:
  - `USDC-c76f1f`
- outer amount:
  - `10000` for sustained production default
- forwarder endpoint:
  - `blindTransfExec`
- forwarded destination:
  - `erd1qqqqqqqqqqqqqpgqeel2kumf0r8ffyhth7pqdujjat9nx0862jpsg2pqaq`
- forwarded function:
  - `swapTokensFixedInput`
- output token requested:
  - `WEGLD-bd4d79`
- min out:
  - `1`

This matches the structure produced by our local `swap2_te()` builder in [interact.rs](../contracts/forwarder-blind/dex-interactor/src/interact.rs):

- `blind_transf_exec(&pair_address, swap_function_call)`
- outer payment in `USDC-c76f1f`
- forwarded `swap_tokens_fixed_input(WEGLD-bd4d79, minOut)`

Practical conclusion from the comparison:

- our `swap2 te` payload construction appears correct
- the same-shard success path is consistent with both:
  - the organizer example
  - our own sanctioned-binary same-shard tests
- the remaining discrepancy is therefore not best explained by payload shape
- the open problem is specifically cross-shard `blindTransfExec` runtime behavior

So the current investigation focus should be:

1. compare successful same-shard SCR chains against failing cross-shard SCR chains
2. obtain a meaningful successful cross-shard `blindTransfExec` example if possible
3. keep cross-shard `te` out of the baseline until that discrepancy is resolved

Current practical ranking:

1. shard `1` direct swap for base validation
2. shard `1` `blindSync` for same-shard contract validation
3. shard `0` / `2` `blindAsyncV1` and `blindAsyncV2` for cross-shard challenge lanes
4. shard `1` same-shard `blindTransfExec` in both tested directions as a validated special lane
5. cross-shard `blindTransfExec` only as a separate investigation track, not a primary operator lane in either tested direction

## Deployment Model

Prepare:

- shard `0` wallet -> deploy shard `0` forwarder
- shard `1` wallet -> deploy shard `1` forwarder
- shard `2` wallet -> deploy shard `2` forwarder

This gives:

- shard `1` forwarder for same-shard interaction with the pair
- shard `0` and `2` forwarders for cross-shard interaction with the pair

## Pre-Challenge Must-Haves

Before March 26:

1. Create and fund three shard-specific wallets.
2. Wrap EGLD into WEGLD on all three.
3. Deploy one forwarder contract per shard.
4. Record all wallet and forwarder addresses in a checked local operator note.
5. Test all four call types at least once.
6. Test both directions:
   - `WEGLD -> USDC`
   - `USDC -> WEGLD`
7. Test drain on every forwarder that accumulates funds.
8. Benchmark at least:
   - shard `1` same-shard `blind_sync`
   - shard `0` or `2` cross-shard `blind_async_v1`
   - shard `0` or `2` cross-shard `blind_async_v2`
   - one `blind_transf_exec` path plus drain

## Main Risks

1. Wrong pair address from the brief inconsistency.
2. Forgetting that `blind_sync` is same-shard only.
3. Losing usable balances inside forwarders because drain tooling is not ready.
4. Underestimating gas needs for forwarded contract calls.
5. Overbuilding a custom contract before we have baseline working deployments.

## Recommended Repo Work

Next tasks for this repo:

1. Add a `docs/Challenge4-Operator-Checklist.md`
2. Add helper scripts for:
   - shard wallet generation notes
   - wrap on shard wallets
   - deploy per shard
   - drain per shard
3. Add a small local config format for:
   - wallet PEM path by shard
   - forwarder address by shard
   - pair address
   - wrapper address
   - token IDs
4. Decide whether to vendor the upstream forwarder/interactor into this repo or keep them in a sibling workspace

## Recommendation On Vendoring

Current state:

- the upstream `forwarder-blind` contract and `dex-interactor` are now vendored into this repo under `contracts/forwarder-blind`
- local changes should remain minimal and well isolated until we have baseline deploy/drain success

Reason:

- challenge-day operations are easier if deployment, wrap, swap, and drain commands live under our control
- relying on an external clone at the last minute adds avoidable operational risk

## Immediate Next Step

Wire a small operator layer around the patched interactor:

1. shard-specific config files
2. deploy command per shard
3. wrap command per shard
4. drain command per shard
5. address/state note for the three deployed forwarders

## Real Run Outcome

Verified from BoN API transaction history for the real run window:

- start:
  - `2026-03-26T10:53:00Z`
- assessment window:
  - through `2026-03-26T11:53:30Z`

Successful calls by Challenge 4 scoring type:

- `blindSync`: `869`
- `blindAsyncV1`: `1426`
- `blindAsyncV2`: `1429`
- `blindTransfExec`: `315`

Total successful smart contract calls:

- `4039`

This means the run met the technical objectives:

- each required call type exceeded `300`
- total successful calls exceeded `1500`
- total successful calls also exceeded `2500`

Important nuance on `blindTransfExec`:

- `swap1 te` (`WEGLD -> USDC`) successes: `308`
- `swap2 te` (`USDC -> WEGLD`) successes: `7`
- `swap2 te` failures: `301`

So:

- `swap2 te` was effectively non-viable in production during this run
- but this did not block eligibility because scoring is per call type, not per direction
- `blindTransfExec` still cleared `300` through `swap1 te`

Post-run conclusion:

- the lane mix was sufficient for Challenge 4 scoring
- the main unresolved production issue is not eligibility anymore
- it is long-run stability:
  - `swap2 te` reliability
  - sustained runner state/config isolation under concurrent execution
