# Window A Rehearsal 2026-03-15

Date: `2026-03-15`

Purpose:
- validate the Window A sender architecture before Guild Wars on March 16, 2026

Setup:
- sender engine: `cmd/windowsprint`
- manifest-driven sender pool
- active senders: `150` shard-2 wallets
- warm reserves funded but not used as senders in these runs
- rehearsal receiver override:
  - `erd1tn5ar9t25p4355kr0fgazjyar39yumykw99e7dzdjze38dasqfwqrydsaa`
- treasury receiver excluded during rehearsal routing

Funding rehearsal:
- tool: `cmd/fundwallets`
- result: `250/250 success`
- distributed:
  - `150` active senders to `0.6 EGLD`
  - `100` warm reserves to `0.2 EGLD`
- total distributed: `110 EGLD`

Sweep rehearsal:
- tool: `cmd/sweepwallets`
- result: `250/250 success`
- swept back to treasury: `107.112499999999942500 EGLD`
- the difference versus the initial `110 EGLD` distribution is explained by:
  - funding tx fees
  - sender rehearsal fees
  - sweep tx fees

Sender rehearsal results:

1. `500 tx @ 50 TPS`
- success: `500/500`
- failures: `0`
- runtime: about `44.7s`

2. `2,000 tx @ 100 TPS`
- success: `2,000/2,000`
- failures: `0`
- runtime: about `76s`

3. `5,000 tx @ 200 TPS`
- success: `5,000/5,000`
- failures: `0`
- runtime: about `2m05s`

4. `10,000 tx @ 400 TPS`
- success: `10,000/10,000`
- failures: `0`
- runtime: about `3m`

5. `20,000 tx @ 600 TPS`
- success: `20,000/20,000`
- failures: `0`
- runtime: about `4m49s`

6. `20,000 tx @ 800 TPS`
- success: `20,000/20,000`
- failures: `0`
- runtime: about `4m21s`

Observed conclusions:
- no failures observed in any rehearsal tier
- nonce handling stayed stable
- confirmation lag remained manageable
- the current sender architecture is validated at least through `800 TPS`
- the planned Window A operating band of `650-750 TPS` has headroom
- funding and post-window treasury recovery both worked cleanly across the full shard-2 sender pool

Recommended live baseline for Window A:
- senders: `150` shard-2 wallets
- initial live TPS: `700`
- live adjustment band: `700-800 TPS`
- increase only if early live acceptance remains clean
