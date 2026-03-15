# Window A Strategy Checklist

Date context:
- Guild Wars Challenge #1, Window A runs on March 16, 2026 from 16:00 UTC to 16:30 UTC.
- Guild funding arrives to the treasury wallet at 15:45 UTC on March 16, 2026.
- Window B starts at 17:00 UTC on March 16, 2026.

## Agreed Strategy

- tx primitive: plain minimum-positive-value `MoveBalance`
- topology: same-shard-first
- treasury/orchestrator: `erd1n28dse0sej7m2rz02ceftr30596arzx2a0trcegl9p3ztdagjahqa9szcx`
- second known shard-2 controlled receiver: `erd16sjyu26hhsgfhqemv335l3u3s6mvzuae78ldgjh4wcrkwtmtwprqj6d66t`
- target: `1.1M tx`
- minimum sustained TPS required: about `611`
- operational target TPS: `650-750`
- wallet universe: `497 total`
- shard-2 prepared set: `250`
- initial active sender pool: `150 shard-2 wallets`
- shard-2 warm reserve: `100 wallets`
- receivers: `2-3 controlled shard-2 receivers`
- active wallet preload target: `0.6 EGLD per wallet`
- warm reserve preload target: `0.15-0.2 EGLD per wallet`
- treasury retains the rest

## 1. Wallet Generation And Shard Sorting

1. Generate the full wallet universe: `497 wallets`.
2. For each wallet, record:
   - wallet id
   - address
   - shard
   - PEM filename/path
   - intended status: `shard2_candidate`, `other_shard`, `reserve`, `receiver_candidate`
3. Ensure at least `250 wallets` are on shard `2`.
4. Select from those 250:
   - `150 shard-2 active sender candidates`
   - `100 shard-2 warm reserve candidates`
5. Keep the remaining `247 non-shard-2 wallets` grouped by shard for possible Window B use.
6. Mark the known controlled shard-2 wallets separately:
   - `erd1n28dse0sej7m2rz02ceftr30596arzx2a0trcegl9p3ztdagjahqa9szcx` as treasury
   - `erd16sjyu26hhsgfhqemv335l3u3s6mvzuae78ldgjh4wcrkwtmtwprqj6d66t` as primary receiver
7. Identify or create the third shard-2 controlled receiver if possible.
8. Store the manifest in a form the team can use quickly:
   - address
   - shard
   - role
   - PEM reference
9. Verify shard assignment before competition day, not during the event.
10. Freeze the initial Window A wallet set once validated, so late changes are deliberate.

## 2. Receiver Set Completion

1. Confirm the shard of the two known receivers:
   - `erd1n28dse0sej7m2rz02ceftr30596arzx2a0trcegl9p3ztdagjahqa9szcx`
   - `erd16sjyu26hhsgfhqemv335l3u3s6mvzuae78ldgjh4wcrkwtmtwprqj6d66t`
2. Add a third controlled shard-2 receiver if possible.
3. Assign receiver roles:
   - treasury receiver: `erd1n28...szcx`
   - primary sink: `erd16sj...66t`
   - secondary sink: third shard-2 receiver
4. Decide the routing split for Window A:
   - primary sink: `45%`
   - secondary sink: `45%`
   - treasury receiver: `10%`
5. If only two receivers are ready:
   - primary sink: `70-80%`
   - treasury receiver: `20-30%`
6. Mark receivers explicitly in the manifest so they are excluded from the normal active sender pool unless deliberately used.
7. Validate operational properties before the event:
   - correct shard
   - correct address formatting
   - team knows which one is treasury
   - team knows which ones are sinks only
8. Keep at least one fallback receiver candidate ready in case a chosen sink becomes unusable.

## 3. Funding Workflow From Treasury At 15:45 UTC

1. Treasury wallet:
   - `erd1n28dse0sej7m2rz02ceftr30596arzx2a0trcegl9p3ztdagjahqa9szcx`
   - treat it as the central funding source first, not as a main sender
2. Predefine three wallet groups:
   - `active_shard2_150`
   - `warm_reserve_shard2_100`
   - `other_shards_window_b_pool`
3. Initial preload targets:
   - active shard-2 senders: `0.6 EGLD each`
   - warm shard-2 reserves: `0.15-0.2 EGLD each`
   - non-shard-2 wallets: no preload unless specifically needed
4. Estimated initial distribution:
   - `150 × 0.6 = 90 EGLD`
   - `100 × 0.15 to 0.2 = 15 to 20 EGLD`
   - initial total sent out: about `105 to 110 EGLD`
5. Keep the overwhelming remainder on treasury for:
   - emergency top-ups
   - scaling up active senders during Window A
   - flexibility for Window B after 16:30 UTC
6. Funding sequence:
   - fund active shard-2 senders first
   - fund warm shard-2 reserves second
   - do not fund extra wallets without a specific reason
7. Treasury should retain enough to:
   - top up underfunded active wallets during Window A
   - activate reserve shard-2 wallets quickly
   - pivot for Window B if the objective changes
8. Funding script should take an explicit wallet group as input and log funded wallet, amount, nonce, and tx hash.
9. Have the funding groups, target amounts, and script tested before March 16, 2026.
10. If treasury funds arrive late or funding speed is slower than expected, prioritize active sender wallets only.

## 4. Window A Sender Runbook

1. Final pre-run inputs must already be fixed:
   - active sender set: `150 shard-2 wallets`
   - warm reserve set: `100 shard-2 wallets`
   - receiver set: `2-3 shard-2 controlled receivers`
   - treasury wallet kept operational
   - minimum positive transfer value decided
   - funding completed
2. Sender behavior:
   - plain `MoveBalance`
   - no data payload
   - same-shard routing
   - local nonce tracking per wallet
   - async confirmation, not synchronous send-wait-send
3. Runtime target:
   - design target: `1.1M tx`
   - minimum required sustained TPS: about `611`
   - operational target: `650-750`
4. Receiver routing:
   - send mostly to non-treasury sinks
   - treasury wallet should receive only a minority share
   - use deterministic rotation
5. Start-of-window behavior at 16:00 UTC:
   - begin with the preselected `150 shard-2 active wallets`
   - ramp quickly but not chaotically
   - verify live success/failure trend immediately
6. During the run, monitor:
   - sent tx count
   - confirmed success count
   - failure/rejection count
   - per-wallet nonce drift indicators
   - treasury balance
   - active wallet balances
   - achieved TPS over time
7. Steering rules during the run:
   - if acceptance is healthy, maintain target throughput
   - if some wallets underperform, rotate in warm reserves
   - if a receiver becomes problematic, reduce or stop traffic to it
   - if failures spike, reduce waste before burning time and fee budget
8. Treasury role during the run:
   - do not make it a heavy sender by default
   - keep it available for emergency top-ups
   - keep it available for post-window reconfiguration
9. End-of-window behavior at 16:30 UTC:
   - stop new sends cleanly
   - preserve logs and summary
   - keep all wallet state available for a possible Window B pivot
10. Documentation during the run:
   - save exact command(s)
   - save active wallet count
   - save receiver list
   - save TPS snapshots and tx count progression
   - save final success count and notable issues
11. Tactical objective:
   - aim to cross `1,000,000 tx` early enough to contend for the bonus
   - do not destabilize the run just to chase burst optics

## 5. Sweep-Back Contingency Runbook

1. Objective:
   - drain funded wallets back to treasury after 16:30 UTC if Window B requirements make the current allocation suboptimal
2. Treasury destination:
   - `erd1n28dse0sej7m2rz02ceftr30596arzx2a0trcegl9p3ztdagjahqa9szcx`
3. Wallet groups the sweep must support:
   - active Window A senders
   - warm shard-2 reserves
   - unused funded wallets
   - optional shard filter
   - optional explicit wallet list
4. Sweep behavior:
   - compute each wallet’s available EGLD
   - leave enough for the sweep tx fee itself
   - send the remainder back to treasury
   - skip empty or dust-only wallets
   - log each sweep tx
5. Required filtering modes:
   - by role tag
   - by shard
   - by active/reserve status
   - by explicit inclusion/exclusion list
6. Required outputs:
   - wallet address
   - shard
   - starting balance
   - swept amount
   - fee estimate
   - tx hash
   - success/failure status
   - total recovered amount
7. Operational sequence after Window A:
   - evaluate Window B objective immediately
   - decide whether to keep current allocation or sweep
   - if sweep is needed, execute the prebuilt script, not manual transfers
   - prioritize the wallets whose balances matter most first
8. Safety constraints:
   - do not sweep the treasury wallet
   - do not sweep wallets still needed for immediate Window B use
   - allow dry-run / preview mode
   - allow partial sweep batches for time-sensitive adaptation
9. Tactical use cases:
   - Window B favors different shards
   - Window B requires different active wallet distribution
   - Window B uses fewer wallets but higher balance concentration
   - Window A overfunded many wallets that are no longer needed
10. Time-management value:
   - the sweep tool is important because the Window A to Window B adjustment window is only 30 minutes, from 16:30 UTC to 17:00 UTC on March 16, 2026
   - without this tool, re-pooling funds manually across hundreds of wallets is not realistic
11. Decision rule:
   - if Window B can reuse the Window A shard-2 funding pattern, do not sweep unnecessarily
   - if Window B needs a different topology, sweep early and decisively
