Battle of Nodes — Guild Wars
Challenge 4: Contract Storm — Pre-Brief
Challenge starts 16:00 UTC, March 26 2026 | Subject to change
This is a pre-brief. Challenge 4 runs on Thursday, March 26 at 16:00 UTC.


We are publishing setup details a day early because this challenge requires more preparation than previous ones.
Use the time to deploy your contracts, test all four call types, and explore optimizations.


Specific scoring metrics will be announced on March 25.
All updates via t.me/BoN_Guild_Leaders.


What Is This Challenge
Challenge 4 tests smart contract composability — specifically how contracts interact with each other, both same-shard and cross-shard.

The setup: each guild deploys a forwarder contract in each shard. Your scripts call the forwarder, which forwards the call to the xExchange DEX pair contract (WEGLD/USDC). You are stress-testing four distinct call mechanisms that MultiversX supports for contract-to-contract interaction.

This is the most technically involved Guild Wars challenge. The extra day is intentional — use it.

Target DEX pair:
erd1qqqqqqqqqqqqqpgqr8n2kjqhrupcrsceevkv6yydtjsgacvuqqqs23m8n6

Reference implementation (forwarder-blind contract):
github.com/multiversx/mx-sdk-rs/tree/master/contracts/feature-tests/composability/forwarder-blind

Example interactor (use directly or as inspiration):
github.com/multiversx/mx-sdk-rs/pull/2315

Setup
Complete this before the challenge starts. The preparation window is the point — do not leave this for the day of the challenge.

Step 1 — Wallets
You need one wallet in each shard (Shard 0, 1, 2). Generate them using:

cargo install multiversx-sc-meta
sc-meta wallet new --shard X
Fund each wallet with EGLD before the challenge starts.

Step 2 — Deploy Forwarder Contracts
Deploy one forwarder contract per shard, using the shard-specific wallet. This determines which shard the forwarder lives in, which controls whether your calls are same-shard or cross-shard relative to the DEX pair (which is on Shard 1).

You have two options:
Deploy the provided forwarder-blind contract as-is. No modifications required — this is the simplest path to getting started.
Build your own contract. You may implement your own forwarder as long as it fulfills the same functional requirements (forwarding calls to the DEX pair via the supported call types). Custom contracts are permitted.

Step 3 — Wrap EGLD
The DEX pair requires WEGLD. Wrap EGLD on all three wallets before the challenge:
Example wrap transaction

As a warm-up, you can also do direct swaps:
WEGLD → USDC: bon-explorer.multiversx.com/transactions/895ece7843ef2748df8c1d6edafe59858d179ff99812874cfeb564b8c90a6744
USDC → WEGLD: bon-explorer.multiversx.com/transactions/a9874e395fe39ddd0baa799b5380c028c2c6966c8fa442236b77a98c79d1c315

Step 4 — Test and Explore
The preparation window is your chance to experiment before scoring is locked. Things worth exploring:
Which call types perform best cross-shard vs. same-shard under load
Drain flows — make sure you can recover tokens from forwarders reliably
Script throughput — how many calls per minute can you sustain
Custom contract optimizations if you are going that route

Scoring metrics will be announced on March 25.
Use this time to get comfortable with the contracts and explore optimizations.


The Four Call Types
The forwarder-blind contract supports four methods for forwarding calls to the DEX pair. Each behaves differently for same-shard vs. cross-shard interactions.

Arguments are the same for all four call types:
Destination address: erd1qqqqqqqqqqqqqpgqeel2kumf0r8ffyhth7pqdujjat9nx0862jpsg2pqa
Endpoint to call: swapTokensFixedInput
Arguments: Token (USDC-c76f1f), Amount

1. Sync Call (blindSync)
Same-shard only — forwarder must be on Shard 1 (same shard as the DEX pair)
Tokens are returned directly to the caller
Example: transaction example

2. Async V1 (blindAsyncV1)
Works both same-shard and cross-shard
Same-shard: tokens returned to caller
Cross-shard: tokens stay in forwarder — call drain@USDC-c76f1f@ and drain@WEGLD-bd4d79@ to recover them
Example: transaction example

Drain transaction example

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
Scoring metric
Announced March 25


Scoring Structure
Scoring follows the same structure as previous Guild Wars challenges. Technical scores are calculated from onchain data. Content scores from portal submissions. Full scoring details, content task deadlines, and portal submission instructions will be published in the full brief on March 25.

Component
Max Score
Weight
How It Works
Technical Score
100 pts
75%
Top guild = 100 pts. Others scored proportionally. Tracked live every ~30 seconds.
Content Score
100 pts
25%
Sum of completed social tasks. All guilds can achieve full marks independently.
Milestone Bonus
TBA
Flat
TBA


Do not wait until March 26 to deploy your contracts — use the preparation window.


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
