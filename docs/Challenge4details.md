Battle of Nodes — Guild Wars
Challenge 4: Contract Storm — Guild Brief
Challenge starts 16:00 UTC, March 26 2026 | Subject to change
Note: Challenge timing and parameters are subject to change based on network conditions.
All updates communicated via t.me/BoN_Guild_Leaders.


What Is This Challenge
Challenge 4 tests smart contract composability — specifically how contracts interact with each other, both same-shard and cross-shard.
Each guild deploys a forwarder contract in each shard. Your scripts call the forwarder, which forwards the call to the xExchange DEX pair contract (WEGLD/USDC). You are stress-testing four distinct call mechanisms that MultiversX supports for contract-to-contract interaction.

Target DEX pair:
erd1qqqqqqqqqqqqqpgqr8n2kjqhrupcrsceevkv6yydtjsgacvuqqqs23m8n6

Reference implementation (forwarder-blind contract):
github.com/multiversx/mx-sdk-rs/tree/master/contracts/feature-tests/composability/forwarder-blind

Example interactor (use directly or as inspiration):
https://github.com/multiversx/mx-sdk-rs/tree/master/contracts/feature-tests/composability/forwarder-blind/dex-interactor

Setup
Complete this before the challenge starts. The longer preparation window is the point — do not leave this for the day of the challenge.

Step 1 — Wallets
You need one wallet in each shard (Shard 0, 1, 2) — at minimum. Generate them using:
cargo install multiversx-sc-meta
sc-meta wallet new --shard X

Each guild may use up to 100 wallets in total. Fund each wallet with EGLD before the challenge window.
Step 2 — Deploy Forwarder Contracts
Deploy one forwarder contract per shard, using the shard-specific wallet. This determines which shard the forwarder lives in, which controls whether your calls are same-shard or cross-shard relative to the DEX pair (which is on Shard 1).

To make sure that everyone is performing all the call types as intended, everyone will be required to only deploy the standard sanctioned compiled binaries of forwarder-blind. 
You can find them in the dex-interactor crate, the wasm binary and the ABI, for reference.

Step 3 — Wrap EGLD
The DEX pair requires WEGLD. Wrap EGLD on all three wallets before the challenge:
Example wrap transaction

As a warm-up, you can also do direct swaps:
WEGLD → USDC: bon-explorer.multiversx.com/transactions/895ece7843ef2748df8c1d6edafe59858d179ff99812874cfeb564b8c90a6744 
USDC → WEGLD: bon-explorer.multiversx.com/transactions/a9874e395fe39ddd0baa799b5380c028c2c6966c8fa442236b77a98c79d1c315 

Step 4 — Test and Explore
Use the preparation window to experiment before the challenge starts. 
Things worth exploring:
Which call types perform best cross-shard vs. same-shard under load
Drain flows — make sure you can recover tokens from forwarders reliably
Script throughput — how many calls per minute can you sustain
Custom contract optimizations if you are going that route

The Four Call Types
The forwarder-blind contract supports four methods for forwarding calls to the DEX pair. Each behaves differently for same-shard vs. cross-shard interactions. Your guild must execute all four call types during the challenge window — each type has a minimum threshold that must be met for full scoring eligibility.

Arguments are the same for all four call types:
Destination address: erd1qqqqqqqqqqqqqpgqeel2kumf0r8ffyhth7pqdujjat9nx0862jpsg2pqa
Endpoint to call: swapTokensFixedInput
Arguments: expected token to be returned (USDC-c76f1f or WEGLD-bd4d79), minimum amount (slippage protection)

The sender also needs to send the ESDT tokens to swap in the transaction to the forwarder (USDC-c76f1f or WEGLD-bd4d79).

1. Sync Call (blindSync)
Same-shard only — forwarder must be on Shard 1 (same shard as the DEX pair)
Tokens are returned directly to the caller
Example: transaction example

2. Async V1 (blindAsyncV1)
Works both same-shard and cross-shard
Same-shard: tokens returned to caller
Cross-shard: tokens stay in forwarder — call drain@USDC-c76f1f@ and drain@WEGLD-bd4d79@ to recover them
Example: transaction example

3. Async V2 / Promises (blindAsyncV2)
Works both same-shard and cross-shard
Same behavior as Async V1 regarding token returns — drain required cross-shard
Example: transaction example

4. Transfer-Execute (blindTransfExec)
Async call without a callback
Works both same-shard and cross-shard
Tokens always stay in forwarder — drain required for both same-shard and cross-shard
Example: transaction example

Call Type
Cross-Shard
Same-Shard
Drain Required
blindSync
No
Yes
No
blindAsyncV1
Yes
Yes
Cross-shard only
blindAsyncV2
Yes
Yes
Cross-shard only
blindTransfExec
Yes
Yes
Always



Challenge Window
Parameter
Details
Date
Thursday, March 26 2026
Window
16:00–17:00 UTC (60 minutes)
Network
Post-Supernova shadow fork — 600ms block times
Budget per guild
500 EGLD (covers gas fees and swap amounts); funds will be distributed to Guild Leader wallets at ~15:45 UTC
Max wallets per guild
100
Required call types
All four — blindSync, blindAsyncV1, blindAsyncV2, blindTransfExec
Minimum total txs
1,500 successful SC calls
Minimum per call type
300 successful calls per type


Gas Budget
Each guild receives 500 EGLD distributed to the guild leader wallet before the challenge. This covers both gas fees and swap amounts — unlike previous challenges, swapped tokens are not consumed. WEGLD swaps to USDC, which can be drained from the forwarder and reused. Your capital recycles; your ongoing cost is gas per call.
Gas spend will be verified against the budget cap. Exceeding the cap will affect your score.
Scoring
Component
Max Score
Weight
How It Works
Technical Score
100 pts
75%
Top guild = 100 pts. Others scored proportionally. All call types weighted equally — 1 successful call = 1 point. All four types must meet the minimum threshold to be eligible.
Content Score
100 pts
25%
Sum of completed social tasks. All guilds can achieve full marks independently.
Milestone Bonus
+10/+7/+5
Flat
First 3 guilds to cross 2,500 total successful SC calls. Independent of leaderboard position.
Maximum per challenge
110 pts


Tech 75 + Content 25 + Bonus 10


Technical Score
Formula: (your guild’s total successful SC calls / top guild’s total successful SC calls) × 100
All four call types are weighted equally — each successful call counts as one point regardless of type. However, all four types must individually reach the 300-call minimum. A guild that does not meet the minimum in any one call type is not eligible for full technical scoring.
Live scores visible at bon.multiversx.com/guild-wars, updated approximately every 30 seconds.
Content Score
Sum of completed content tasks, max 100 points. All guilds can achieve full marks independently. Submitted by the guild leader only via the portal within 12 hours of challenge close.

Important — posting deadlines apply: Tasks 1 and 2 must be posted during the challenge window. Tasks 3, 4, and 5 can be posted any time within the 12-hour submission window, but must be submitted via the portal before 05:00 UTC March 27.

Content Tasks
Prerequisites (unscored — required to be eligible for content scoring)
Prerequisite
How to complete
Guild Twitter/X account exists, name identifiable as the guild
Must be set up and submitted via portal before challenge starts
Guild GitHub repo exists and is public
Must be set up and submitted via portal before challenge starts


Note: If prerequisites are not completed and submitted via the portal, content tasks cannot be verified and will not be scored.


Per-challenge tasks
#
Task
Pts
Posting deadline
Requirements
Submit via portal
1
Pre-challenge post
10
Before 16:00 UTC
Tag @MultiversX. No hashtags.
X/Twitter link
2
3+ live updates during challenge
20
16:00–17:00 UTC
Timestamps within challenge window. @MultiversX in first update only. No hashtags.
X/Twitter link (1), (2), (3)
3
Post-challenge recap
20 +5
Within 12h of close
Twitter/X: min 3 connected tweets. No hashtags. +5 if cross-posted to Reddit (r/MultiversXOfficial or r/CryptoCurrency, 200+ words).
X/Twitter link (required); Reddit link (optional, +5)
4
Short video covering the challenge
20 +5
Within 12h of close
Twitter/X: 60s minimum. No hashtags. +5 if on YouTube (3–5 hashtags in description: #MultiversX #BattleOfNodes #blockchain #GuildWars) or Reddit.
X/Twitter link (required); YouTube or Reddit link (optional, +5)
5
Scripts or README on guild GitHub
20
Within 12h of close
Public repo updated after challenge closes. Content must be relevant to Contract Storm.
GitHub repo or file link


Tagging and hashtag rules
Platform
Tagging
Hashtags
Twitter/X
@MultiversX — in first tweet or first live update only
None
Reddit
No tagging required
None
YouTube
No tagging required
3–5 in description: #MultiversX #BattleOfNodes #blockchain #GuildWars



Milestone Bonus
A flat bonus is awarded to the first three guilds to cross 2,500 total successful SC calls during the challenge window. Tracked in real time. Independent of final leaderboard position.

Milestone
Bonus
1st guild to cross 2,500 successful SC calls
+10 pts
2nd guild to cross 2,500 successful SC calls
+7 pts
3rd guild to cross 2,500 successful SC calls
+5 pts


Do not start contract calls before 16:00 UTC — pre-challenge activity is excluded from scoring.


Do not neglect any of the four call types — failing to reach 300 calls in any single type affects your scoring eligibility.


Do not forget to drain forwarder contracts — tokens left in forwarders cross-shard are stuck until drained.


Do not use the guild leader wallet to submit transactions — it is a funding wallet only.


Prize Pool
Prize
Distribution
$7,500 EGLD — Overall Leaderboard
Top 10 guilds by total score across all 5 challenges
$2,500 EGLD — Per-Challenge Prizes
5 challenges × $500 EGLD per challenge (1st: $250 / 2nd: $150 / 3rd: $100)
$10,000 EGLD — Total Guild Wars Prize Pool




Note: Guild size is not normalized. No score adjustment based on number of members. Optimal guild size is 5.


Key Links
Guild Wars portal: bon.multiversx.com/guild-wars 
Network explorer: bon-explorer.multiversx.com 
Guild leader Telegram: t.me/BoN_Guild_Leaders 

Developer Resources
BoN API docs: api.battleofnodes.com/docs 
BoN gateway: gateway.battleofnodes.com/network/config 
MultiversX GitHub: github.com/multiversx 
mx-sdk-rs (forwarder-blind): github.com/multiversx/mx-sdk-rs 
mx-sdk-py-cli: github.com/multiversx/mx-sdk-py-cli 
