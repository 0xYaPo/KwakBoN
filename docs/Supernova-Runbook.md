# Supernova Runbook

Date context:
- `2026-03-16`
- Challenge 2: `Supernova Surge`

Principle:
- `bulksprint` is now a serious primary engine candidate on live Supernova
- `windowsprint` remains the fallback engine and the safer operational baseline
- never run both at the same time

Live evidence on `2026-03-16`:
- balanced `windowsprint`, `250` wallets, `10 min`, `1200 TPS` target:
  - `324,160 sent`
- balanced `bulksprint`, same `250` wallets, `10 min`, tuned `25/50` profile:
  - `1,746,002 accepted`

## Window A

### Preferred Plan

1. start with `bulksprint` on the balanced `250` wallet set
2. use the tuned profile only:
   - `BULK_BATCH_SIZE=25`
   - `BULK_MAX_NONCE_LOOKAHEAD=50`
   - `BULK_MAX_CONCURRENT_READS=24`
3. keep `windowsprint` ready as the fallback engine if bulk behavior becomes unhealthy

Why:
- live Supernova testing moved bulk from “interesting finisher” to “real primary candidate”
- the balanced sender topology materially improved behavior for both senders
- the fixed bulk deadlock at `50k` is resolved

### 16:00-16:20 UTC

- run `bulksprint`
- do not switch away early unless there is a clear operational reason
- first priority is maximizing accepted throughput while the path is healthy

Monitor:
- send pace
- accepted pace
- gateway/API errors
- nonce anomalies
- approximate fee burn
- leaderboard position

Decision point at about `16:18-16:20 UTC`:
- is bulk still materially outperforming fallback expectations?
- is the API path still healthy enough for bulk?
- are accepted tx still tracking closely with sent tx?
- is fee headroom still comfortable?
- is the operator ready to stop bulk and move to `windowsprint` cleanly if needed?

If the answers are mostly yes:
- keep `bulksprint`

If not:
- switch to `windowsprint`

### 16:20-16:30 UTC

Two options only:

#### Option A: stay on bulk

- keep `bulksprint`
- preferred if bulk remains healthy and fee usage is still acceptable

#### Option B: fall back to sprint

1. stop `bulksprint`
2. start `windowsprint`
3. use the balanced sender set if possible

Do not:
- run both together
- improvise a more aggressive bulk profile live than the validated `25/50`

## Window B

### 17:00-17:20 UTC

- default: start from the same tuned `bulksprint` profile only if fee discipline is clear
- otherwise use `windowsprint`
- be more conservative because Window B has the tighter fee budget

Monitor:
- send pace
- accepted pace
- gateway/API errors
- nonce anomalies
- approximate fee burn
- leaderboard position

### 17:20-17:30 UTC

- only stay on or switch to bulk if:
  - the API path looks clean
  - fee headroom is still comfortable
  - accepted tx remain close to sent tx
  - the operator can perform a clean switch if needed

Otherwise:
- stay on `windowsprint` or fall back to it

## Switch Checklist

Stay on `bulksprint` only if all or almost all are true:
- bulk accepted pace is materially ahead of sprint expectations
- no rising API timeout or rejection pattern is visible
- accepted tx are not drifting far behind sent tx
- fee headroom remains comfortable
- the operator can perform a clean stop/start switch quickly if needed

Fall back to `windowsprint` if:
- gateway behavior is shaky
- accepted tx start lagging significantly behind sent tx
- fee tracking is unclear
- operator focus is already stretched
- too little time remains for a clean changeover

## Practical Execution

Window A preferred start:
- launch `bulksprint` with the tuned `25/50` profile on the balanced set

Fallback mode:
1. stop `bulksprint`
2. launch `windowsprint`
3. monitor stable sent pace and transient errors

Window B:
- same logic, but with a higher bar for staying on bulk because the budget is tighter

## Team Roles

Best case:
- one operator runs commands
- one person watches metrics, leaderboard, and fee position
- one person records decisions and timestamps

Minimum:
- one operator
- one watcher

## Core Rule

- `bulksprint` is now the preferred high-throughput option on live Supernova
- `windowsprint` is the fallback safety net
- switching remains optional and should stay operator-controlled
