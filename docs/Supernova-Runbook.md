# Supernova Runbook

Date context:
- `2026-03-16`
- Challenge 2: `Supernova Surge`

Principle:
- `windowsprint` is the default engine
- `bulksprint` is an optional late-window accelerator
- never run both at the same time

## Window A

### 16:00-16:20 UTC

- run `windowsprint`
- do not switch early just because bulk is faster on paper
- first priority is stable in-window success

Monitor:
- send pace
- confirmation pace
- gateway/API errors
- nonce anomalies
- approximate fee burn
- leaderboard position

Decision point at about `16:18-16:20 UTC`:
- are we behind target enough that a late acceleration is worth the risk?
- is the API path healthy enough to support bulk?
- are confirmations catching up cleanly?
- is fee headroom still comfortable?
- is the operator ready to stop one engine and start the other cleanly?

If the answers are mostly yes:
- prepare finish mode

If not:
- stay on `windowsprint` to the end

### 16:20-16:30 UTC

Two options only:

#### Option A: no switch

- keep `windowsprint`
- preferred if network conditions are uncertain

#### Option B: finish mode

1. stop `windowsprint`
2. start `bulksprint`
3. use only the tuned bulk profile:
   - `BULK_BATCH_SIZE=25`
   - `BULK_MAX_NONCE_LOOKAHEAD=50`
   - `BULK_MAX_CONCURRENT_READS=24`

Do not:
- run both together
- improvise a more aggressive bulk profile live

## Window B

### 17:00-17:20 UTC

- default: `windowsprint`
- be more conservative because Window B has the tighter fee budget

Monitor:
- send pace
- confirmation pace
- gateway/API errors
- nonce anomalies
- approximate fee burn
- leaderboard position

### 17:20-17:30 UTC

- only consider bulk finish mode if:
  - the API path looks clean
  - fee headroom is still comfortable
  - a late push is actually needed
  - the operator can perform a clean switch

Otherwise:
- stay on `windowsprint`

## Switch Checklist

Switch to `bulksprint` only if all or almost all are true:
- `windowsprint` has been stable
- no serious nonce issues have appeared
- no rising API timeout or rejection pattern is visible
- confirmations are not badly lagging
- fee headroom remains comfortable
- more raw pace is needed than `windowsprint` is currently delivering
- the operator can perform a clean stop/start switch quickly

Do not switch if:
- gateway behavior is shaky
- confirmations are already lagging significantly
- fee tracking is unclear
- operator focus is already stretched
- too little time remains for a clean changeover

## Practical Execution

Window A start:
- launch `windowsprint`

Possible Window A finish mode:
1. stop `windowsprint`
2. launch `bulksprint` with the tuned profile
3. monitor accepted pace and pending behavior

Window B:
- same logic, but with a higher bar before switching

## Team Roles

Best case:
- one operator runs commands
- one person watches metrics, leaderboard, and fee position
- one person records decisions and timestamps

Minimum:
- one operator
- one watcher

## Core Rule

- `windowsprint` is the baseline
- `bulksprint` is a tactical late-window option
- switching is optional, never automatic
