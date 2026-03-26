# Supernova Day-Of Checklist

Date:
- `2026-03-16`

Primary references:
- [Supernova-Surge-2026-03-16.md](/C:/Users/y.pochon/Dev/Mvx/KwakBoN/docs/Supernova-Surge-2026-03-16.md)
- [Supernova-Runbook.md](/C:/Users/y.pochon/Dev/Mvx/KwakBoN/docs/Supernova-Runbook.md)
- [Supernova-Balanced-250-Set.md](/C:/Users/y.pochon/Dev/Mvx/KwakBoN/docs/Supernova-Balanced-250-Set.md)
- [BulkSprint-Benchmarks-2026-03-16.md](/C:/Users/y.pochon/Dev/Mvx/KwakBoN/docs/BulkSprint-Benchmarks-2026-03-16.md)

## Challenge Timing

All times UTC.

- `14:00` brief published
- `15:45` rosters lock
- `15:45` `2,500 EGLD` sent to guild leader wallet
- `15:45` content portal opens
- `16:00` Window A starts
- `16:30` Window A ends
- `16:30-17:00` break
- `17:00` Window B starts
- `17:30` challenge closes
- `17:30-05:30+1` content submission window

## Rules To Keep In Mind

- max `500` unique sending wallets
- guild leader wallet must not send counted `MoveBalance`
- all sender wallets must be funded directly by the guild leader wallet
- sender matters for scoring, receiver shard does not
- Window A fee budget: `2,000 EGLD`
- Window B fee budget: `500 EGLD`

## Validated Live Sender Results

Balanced `250` wallet set:
- manifest: `configs/wallets-manifest.supernova-balanced-250.json`

`windowsprint`:
- `10 min`
- `1200 TPS` target
- result:
  - `324,160 sent`

`bulksprint`:
- tuned profile:
  - `BULK_BATCH_SIZE=25`
  - `BULK_MAX_NONCE_LOOKAHEAD=50`
  - `BULK_MAX_CONCURRENT_READS=24`
- `10 min` result:
  - `1,747,130 sent`
  - `1,746,002 accepted`
  - `15 errors`

Current operating decision:
- primary sender: `bulksprint`
- fallback sender: `windowsprint`

## Before 15:45 UTC

- verify system clock sync
- verify guild leader wallet address and PEM path
- verify balanced manifest is present
- verify bulk timed launcher script works in `--dry-run`
- keep fallback `windowsprint` command ready

## At 15:45 UTC

- confirm `2,500 EGLD` arrived in the guild leader wallet
- fund sender wallets directly from the guild leader wallet
- preserve Window B fee discipline

Funding targets:
- Window A balanced set:
  - `250` wallets
  - `7.0 EGLD` target per wallet
  - total sender allocation: `1,750 EGLD`
- Window B balanced set:
  - same `250` wallets
  - `1.9 EGLD` target per wallet
  - total sender allocation: `475 EGLD`
- rationale:
  - Window A keeps meaningful room under the `2,000 EGLD` fee budget
  - Window B supports either `bulksprint` or fallback `windowsprint` on the same balanced set

Funding command:

```powershell
cd C:\Users\y.pochon\Dev\Mvx\KwakBoN

$env:WALLETS_MANIFEST='.\configs\wallets-manifest.supernova-balanced-250.json'
$env:TREASURY_ADDRESS='erd1n28dse0sej7m2rz02ceftr30596arzx2a0trcegl9p3ztdagjahqa9szcx'
$env:TREASURY_PEM_PATH='C:\secure\mvx\treasury-shard2.pem'
$env:FUND_INCLUDE_STATUSES='supernova_balanced_active'
$env:FUND_REQUIRED_TAGS='sender'
$env:FUND_SHARD_FILTER='-1'
$env:FUND_TARGET_ACTIVE_EGLD='7.0'
$env:FUND_TARGET_SUPERNOVA_BALANCED_ACTIVE_EGLD='7.0'
$env:FUND_TARGET_RESERVE_EGLD='1.0'
$env:FUND_TARGET_WINDOW_B_RESERVE_EGLD='1.0'

go run ./cmd/fundwallets --dry-run
go run ./cmd/fundwallets
```

Window B top-up command:

```powershell
cd C:\Users\y.pochon\Dev\Mvx\KwakBoN

$env:WALLETS_MANIFEST='.\configs\wallets-manifest.supernova-balanced-250.json'
$env:TREASURY_ADDRESS='erd1n28dse0sej7m2rz02ceftr30596arzx2a0trcegl9p3ztdagjahqa9szcx'
$env:TREASURY_PEM_PATH='C:\secure\mvx\treasury-shard2.pem'
$env:FUND_INCLUDE_STATUSES='supernova_balanced_active'
$env:FUND_REQUIRED_TAGS='sender'
$env:FUND_SHARD_FILTER='-1'
$env:FUND_TARGET_ACTIVE_EGLD='1.9'
$env:FUND_TARGET_SUPERNOVA_BALANCED_ACTIVE_EGLD='1.9'
$env:FUND_TARGET_RESERVE_EGLD='1.0'
$env:FUND_TARGET_WINDOW_B_RESERVE_EGLD='1.0'

go run ./cmd/fundwallets --dry-run
go run ./cmd/fundwallets
```

## Window A

Preferred start:
- start `bulksprint` on the balanced set
- monitor accepted pace and error count
- keep `windowsprint` ready as fallback

Timed bulk launch:

```powershell
cd C:\Users\y.pochon\Dev\Mvx\KwakBoN
.\run-bulk-at-time.ps1 -LaunchAtUtc "2026-03-16T16:00:00Z" -DurationSeconds 1800
```

Immediate bulk launch:

```powershell
cd C:\Users\y.pochon\Dev\Mvx\KwakBoN

$env:WALLETS_MANIFEST='.\configs\wallets-manifest.supernova-balanced-250.json'
$env:BULK_SENDER_STATUSES='supernova_balanced_active'
$env:BULK_REQUIRED_TAGS='sender'
$env:BULK_SHARD_FILTER='-1'
$env:BULK_TARGET_TX='9999999'
$env:BULK_DURATION_SECONDS='1800'
$env:BULK_BATCH_SIZE='25'
$env:BULK_MAX_NONCE_LOOKAHEAD='50'
$env:BULK_MAX_CONCURRENT_READS='24'
$env:BULK_MAX_ACTIVE_WALLETS='0'
$env:BULK_IDLE_SLEEP_MS='300'
$env:BULK_WAIT_CONFIRM='false'
$env:BULK_CONFIRM_WORKERS='32'
$env:BULK_POLL_INTERVAL_SECONDS='3'
$env:BULK_CONFIRM_TIMEOUT_SECONDS='240'

go run ./cmd/bulksprint
```

Fallback sprint:

```powershell
cd C:\Users\y.pochon\Dev\Mvx\KwakBoN

$env:WALLETS_MANIFEST='.\configs\wallets-manifest.supernova-balanced-250.json'
$env:SPRINT_SENDER_STATUSES='supernova_balanced_active'
$env:SPRINT_REQUIRED_SENDER_TAGS='sender'
$env:SPRINT_RECEIVER_ADDRESSES='erd1m8sctsag8qvy6s3a53pz73wlwnh94lduz8luzjr0th65g4rl2xgqjga7lx,erd1d940d7urvw8yh42ssnk82vwr5jlwrjrcp28p9e7spj96yvt64m8s667eq4,erd16sjyu26hhsgfhqemv335l3u3s6mvzuae78ldgjh4wcrkwtmtwprqj6d66t'
$env:INCLUDE_TREASURY_RECEIVER='false'
$env:SPRINT_SHARD_FILTER='-1'
$env:SPRINT_TARGET_TX='9999999'
$env:SPRINT_DURATION_SECONDS='1800'
$env:SUSTAINED_TPS='1200'
$env:SPRINT_WORKERS='256'
$env:SPRINT_CONFIRM_WORKERS='64'
$env:SPRINT_MAX_INFLIGHT_PER_WALLET='2'
$env:SPRINT_SUCCESS_COOLDOWN_MS='40'
$env:SPRINT_TRANSIENT_COOLDOWN_MS='1200'
$env:SPRINT_QUARANTINE_COOLDOWN_MS='8000'
$env:SPRINT_QUARANTINE_TRANSIENT_THRESHOLD='3'
$env:WAIT_CONFIRM='false'
$env:CONFIRM_SUCCESS_TARGET='1'
$env:CONFIRM_TIMEOUT_SECONDS='240'
$env:CONTINUE_ON_TRANSIENT_SEND_ERROR='true'

go run ./cmd/windowsprint
```

## Break

- stop Window A activity cleanly
- assess fee spend
- decide whether to keep bulk primary for Window B
- top up or rebalance if needed

## Window B

- same preferred sender: `bulksprint`, if fee discipline still looks acceptable
- otherwise use `windowsprint`
- do not run both senders at the same time

## Final Reminder

- do not send counted `MoveBalance` from the guild leader wallet
- do not use intermediary funding wallets
- do not exceed the fee budget
- if bulk behavior becomes unhealthy, switch to sprint instead of trying to improvise
