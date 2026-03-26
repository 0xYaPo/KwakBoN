# Battle of Nodes - Guild Wars

## Challenge 2: Supernova Surge - Guild Brief

Released: `2026-03-16 14:00 UTC`  
Challenge starts: `2026-03-16 16:00 UTC`  
Status: tentative, subject to change

Primary update channel:
- `t.me/BoN_Guild_Leaders`

Note:
- challenge timing and parameters are explicitly marked as tentative
- any live change should be treated as authoritative only once communicated through the guild leaders channel

## What This Challenge Is

Challenge 1, `Transaction Sprint`, established the baseline on the pre-Supernova network.

Challenge 2, `Supernova Surge`, runs the same objective again on the post-Supernova network.

What stays the same:
- objective
- structure
- rules
- transaction type: `MoveBalance`

What changes:
- block times drop from `6 seconds` to `600 milliseconds`
- the network is effectively `10x` faster by block cadence
- move balance transfer capacity is estimated at about `4x` higher

What is counted:
- total successful `MoveBalance` transactions
- submitted by all unique guild wallets
- across both windows combined
- only the sender matters
- receiver shard is irrelevant and is not checked

## Supernova Upgrade - What Changed

Before Supernova:
- `6-second` block times
- about `10 blocks/minute`
- Challenge 1 baseline conditions

After Supernova:
- `600ms` block times
- about `100 blocks/minute`
- Challenge 2 real conditions

Operational implication:
- transactions can settle much faster
- throughput ceilings should be higher
- nonce handling, submission timing, and concurrency behavior may change materially
- scripts must be tested on the post-Supernova network before the live window

## Differences From Challenge 1

| Parameter | Challenge 1 - Transaction Sprint | Challenge 2 - Supernova Surge |
|---|---|---|
| Network | Pre-Supernova shadow fork | Post-Supernova shadow fork |
| Block time | 6 seconds | 600 milliseconds |
| Transaction type | MoveBalance | MoveBalance |
| Window A budget | 2,000 EGLD | 2,000 EGLD |
| Window B budget | 500 EGLD | 500 EGLD |
| Max wallets | 500 | 500 |
| Milestone threshold | 1,000,000 tx | 2,500,000 tx |

## Full Schedule - 2026-03-16 UTC

| Time (UTC) | Event | Details |
|---|---|---|
| 14:00 | Brief published | Window B specs revealed. 2 hours to prepare. |
| 15:45 | Roster locks | No roster changes after this point. |
| 15:45 | Funds distributed | `2,500 EGLD` sent to each guild leader wallet. |
| 15:45 | Portal opens | Content submissions accepted on `bon.multiversx.com/guild-wars`. |
| 16:00 | Window A starts | Challenge begins. |
| 16:30 | Window A ends | Activity after `16:30 UTC` excluded from Window A scoring. |
| 16:30 - 17:00 | Break | Adjust scripts and strategy for Window B. |
| 17:00 | Window B starts | Challenge resumes. |
| 17:30 | Challenge closes | Activity after `17:30 UTC` excluded from scoring. |
| 17:30 - 05:30+1 | Content submission window | `12h` to submit content tasks. Posting deadlines remain separate from submission deadlines. |

## Window A - Capacity

Window:
- `2026-03-16 16:00 UTC` to `16:30 UTC`

Parameters:
- objective: maximize successful transactions within the time window and fee budget
- network: post-Supernova shadow fork
- transaction type: simple `MoveBalance`
- fee budget: `2,000 EGLD`
- max unique wallets: `500`
- sender is the only wallet checked for attribution and scoring

Important notes:
- transaction value must be set to the minimum possible amount
- fee spend is what is budgeted and later verified
- receiver shard does not matter for scoring

## Window B - Capability

Window:
- `2026-03-16 17:00 UTC` to `17:30 UTC`

Parameters:
- objective: maximize successful transactions within the time window and fee budget
- network: post-Supernova shadow fork
- transaction type: simple `MoveBalance`
- fee budget: `500 EGLD`
- max unique wallets: `500`
- same wallet set as Window A
- sender is the only wallet checked for attribution and scoring

Important notes:
- same minimum-value rule applies
- Window B rewards better fee efficiency and better script tuning

## Funding And Wallet Setup

At `15:45 UTC`, `2,500 EGLD` is sent directly to the registered guild leader wallet.

Funding structure:

| Window | Fee Budget | Max Wallets |
|---|---|---|
| Window A | 2,000 EGLD | 500 |
| Window B | 500 EGLD | 500 |
| Total distributed | 2,500 EGLD | - |

Critical funding rules:
- the EGLD above is a fee budget, not intended transfer value
- set `MoveBalance` value to the minimum possible amount
- unused Window A budget cannot be rolled into Window B
- Window B budget is fixed at `500 EGLD`
- all sending wallets must be funded directly by the guild leader wallet
- intermediary wallets are not allowed for attribution

Tracing rule:
- only direct guild leader -> sender wallet funding counts for attribution
- guild leader -> wallet A -> wallet B means wallet B does not count

Guild leader wallet role:
- funding wallet only
- operator wallet only
- must not send counted `MoveBalance` traffic itself
- should stay clean during the challenge

## Wallet Rules

Rules carried over from Challenge 1:
- maximum `500` unique sending wallets per guild
- guild leader wallet must not submit counted `MoveBalance` transactions
- wallet addresses must be unique to a single guild
- shared wallets do not count and may lead to penalties or disqualification
- all sending wallets must be funded directly by the guild leader wallet
- the same `500` wallets from Challenge 1 may be reused

Practical warnings:
- do not send counted traffic from the guild leader wallet
- do not use intermediary wallets
- do not use addresses that may overlap with another guild

## Technical Setup Guidance

### Step 1 - Before 15:45 UTC

- test all transaction scripts on the post-Supernova network
- review nonce handling carefully for `600ms` blocks
- prepare any direct funding scripts in advance
- verify the full sender wallet set is ready and unique
- use explorer and developer resources to validate behavior

### Step 2 - At 15:45 UTC

- receive `2,500 EGLD` in the guild leader wallet
- fund senders directly from the guild leader wallet
- reserve about `2,000 EGLD` for Window A
- reserve about `500 EGLD` for Window B

### Step 3 - At 16:00 UTC

- start Window A scripts
- monitor live scores at `bon.multiversx.com/guild-wars`
- publish the first live update on Twitter/X and tag `@MultiversX`

### Step 4 - At 16:30 UTC

- stop Window A submission scripts
- review live scores
- redistribute or consolidate funds if needed
- tune scripts for Window B efficiency
- prepare for the `17:00 UTC` restart

### Step 5 - At 17:00 UTC

- start Window B scripts
- continue live updates
- monitor progress toward the `2,500,000 tx` milestone bonus

### Step 6 - After 17:30 UTC

- stop on-chain activity
- technical scores are computed from on-chain data
- content submission remains open for `12h`, until `2026-03-17 05:30 UTC`
- publish recap content and push scripts to GitHub
- submit all content links through the portal

## Operational Takeaways For This Repo

- receiver shard is irrelevant for scoring in Challenge 2
- only the sender is checked
- direct guild leader funding is mandatory
- the guild leader wallet must remain a funding/operator wallet only
- post-Supernova timing strongly favors retesting nonce and concurrency behavior
- Window B keeps the same sender wallet set and reduces the fee budget sharply
- the milestone target moves from `1,000,000` to `2,500,000` transactions
