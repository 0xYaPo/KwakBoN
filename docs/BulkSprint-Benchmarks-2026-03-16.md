# BulkSprint Benchmarks 2026-03-16

Date: `2026-03-16`

Purpose:
- evaluate the new offensive bulk sender against the shared Battle API path
- compare `cmd/bulksprint` against the hardened `cmd/windowsprint` sender
- identify a first viable bulk profile for real competition use

Tooling:
- sender engine: `cmd/bulksprint`
- gateway path: `transaction/send-multiple`
- sender pool: `497` wallets
- routing: same-shard ring generated directly from the selected sender set
- read path: one `nonce + balance` fetch per batch

Reference `windowsprint` benchmark:
- profile: `497` senders, all shards, same-shard receivers, `700 TPS`
- result: `20,000/20,000 success`
- send phase: about `38s`
- full confirmations: about `3m27s`

## Bulk V1 initial benchmark

Profile:
- `BULK_BATCH_SIZE=50`
- `BULK_MAX_NONCE_LOOKAHEAD=75`
- `BULK_MAX_CONCURRENT_READS=24`
- `20,000 tx`

Result without confirmation:
- `20,000/20,000 accepted`
- send phase: about `3.47s`
- no overshoot after quota fix

Result with lightweight confirmation:
- `20,000/20,000 accepted`
- after `3m` confirmation window:
  - `19,507 success`
  - `0 failed`
  - `493 pending`
- conclusion:
  - broadcast path is extremely fast
  - this profile is too aggressive for fast finality

## Bulk tuned benchmark

Profile:
- `BULK_BATCH_SIZE=25`
- `BULK_MAX_NONCE_LOOKAHEAD=50`
- `BULK_MAX_CONCURRENT_READS=24`
- `20,000 tx`

Result with lightweight confirmation:
- `20,000/20,000 accepted`
- send phase: about `4.05s`
- full confirmations:
  - `20,000/20,000 success`
  - `0 failed`
- intermediate confirmation state:
  - after about `3m`: `19,790 success`, `210 pending`
  - a few seconds later: all confirmed

Conclusion:
- this is the first bulk profile that keeps the very large broadcast advantage
- it also reaches full final success in a timeframe competitive with `windowsprint`

## Higher-volume tuned benchmark

Profile:
- `BULK_BATCH_SIZE=25`
- `BULK_MAX_NONCE_LOOKAHEAD=50`
- `BULK_MAX_CONCURRENT_READS=24`
- `50,000 tx`
- `BULK_CONFIRM_TIMEOUT_SECONDS=240`

Result:
- `50,000/50,000 accepted`
- send phase: about `13.0s`
- after `4m` confirmation window:
  - `49,958 success`
  - `0 failed`
  - `42 pending`
- process exited on confirmation timeout, not on observed tx failure

Conclusion:
- the tuned bulk profile scales well beyond the `20,000 tx` benchmark
- remaining risk at this stage is confirmation lag, not broadcast instability

## Current Recommended Bulk Profile

For further testing:
- `BULK_BATCH_SIZE=25`
- `BULK_MAX_NONCE_LOOKAHEAD=50`
- `BULK_MAX_CONCURRENT_READS=24`

Current interpretation:
- `windowsprint` remains the robust general-purpose competition sender
- `bulksprint` is now a credible offensive sender for shared-API throughput tests
- the tuned `25/50` profile is the first one that looks suitable for real competitive experimentation

## Live Supernova Results

Environment:
- date: `2026-03-16`
- network: live post-Supernova Battle API
- sender set: balanced `250` wallets across shards `0/1/2`
- manifest: `configs/wallets-manifest.supernova-balanced-250.json`

Balanced `windowsprint` reference:
- `10 min`
- `1200 TPS` target
- `250` senders
- result:
  - `324,160 sent`
  - `15 transient send errors`
  - elapsed: about `10m`

Balanced `bulksprint` fixed-target verification:
- `100,000 tx`
- `BULK_WAIT_CONFIRM=false`
- result:
  - `100,000/100,000 accepted`
  - `0 errors`
  - elapsed: about `32.7s`

Balanced `bulksprint` live `10 min` run:
- `BULK_BATCH_SIZE=25`
- `BULK_MAX_NONCE_LOOKAHEAD=50`
- `BULK_MAX_CONCURRENT_READS=24`
- `BULK_WAIT_CONFIRM=false`
- result:
  - `1,747,130 sent`
  - `1,746,002 accepted`
  - `69,924 batches`
  - `15 errors`
  - elapsed: `10m0s`

Important implementation note:
- an internal send-phase deadlock was found and fixed on `2026-03-16`
- root cause:
  - accepted tx hashes were buffered during send but not drained until after the send phase
  - this created a hard plateau around `50,000 tx`
- after the fix:
  - fixed-target and duration bulk runs completed normally
  - the live `10 min` benchmark above became possible

Updated interpretation:
- on live Supernova, `bulksprint` is no longer just a tactical finisher
- with the balanced `250` wallet set and the tuned `25/50` profile, it massively outperformed `windowsprint` on raw accepted throughput
- `windowsprint` remains the safer fallback and simpler operator path
- `bulksprint` is now a serious primary-candidate sender for Window A style throughput pushes
