Battle of Nodes — Guild Wars
Challenge 3: Crossover — Guild Brief
Challenge starts 16:00 UTC, March 24 2026 | Subject to change
Note: Challenge timing and parameters are tentative and subject to change based on network conditions. All updates communicated via t.me/BoN_Guild_Leaders.


What Is This Challenge
Challenges 1 and 2 tested raw transaction throughput and network capacity using MoveBalance — with no constraint on which shard the receiver was on. Both were designed to measure volume.

Challenge 3 — Crossover — changes the objective. Transactions must now cross shard boundaries. Only transfers where the sender and receiver are in different shards are counted. Intra-shard transactions are excluded entirely from scoring.

What is counted: Total successful cross-shard MoveBalance transactions submitted by all unique guild wallets, across both parts combined. Only cross-shard sends count — the receiver must be on a different shard from the sender.

Cross-Shard Transactions — What This Means
Cross-shard means the sender wallet and receiver wallet are on different shards of the network.
A transaction where both sender and receiver are on the same shard is intra-shard and will not count.
Wallet placement across shards is determined by address. Distribute your sending wallets across all 3 shards before the challenge starts.


Full Schedule — 24 March 2026 (All Times UTC)
Time (UTC)
Event
Details
15:45
Roster locks (new guilds)
Rosters for guilds newly registered during BoN are automatically locked. Existing guilds unaffected.
15:45
Funds distributed
2,500 EGLD sent directly to each guild leader wallet. Distribute to sending wallets immediately — 15 minutes to Part 1.
16:00
Part 1 starts
Challenge begins. All scripts must be running.
16:30
Part 1 ends
All Part 1 activity after 16:30 UTC excluded from scoring.
16:30–17:00
Break (30 min)
Adjust scripts, wallets, and strategy for Part 2. Part 2 requires a new set of 500 wallets.
17:00
Part 2 starts
Challenge resumes.
17:30
Challenge closes
All onchain activity after 17:30 UTC excluded from scoring.
17:30–05:30+1
Content submission window
12 hours to submit content tasks via the portal. Note: some tasks must have been posted during the challenge window — submission deadline is separate from posting deadline.


Part 1 — Capacity (16:00–16:30 UTC)
Maximize cross-shard transactions. Your full fee budget is available. No constraints beyond wallet count and budget.

Objective
Maximize successful cross-shard transactions within the time window and fee budget
Network
Post-Supernova shadow fork — 600ms block times
Transaction type
Cross-shard MoveBalance (sender and receiver in different shards)
Window
16:00–16:30 UTC (30 minutes)
Fee budget per guild
2,000 EGLD — fee and transaction value budget. Set transaction value to at least 1×10⁻¹⁸ EGLD. Fee spend and minimum transaction value will be verified. Subject to final confirmation.
Max sending wallets per guild
500 — newly created wallets, funded after 15:45 UTC. Must be unique across all guilds. See Wallet Rules.
Metric
Total successful cross-shard transactions submitted by all guild wallets during the window. Only cross-shard sends count.


Part 2 — Capability (17:00–17:30 UTC)
Same objective. Tighter budget. Higher minimum transaction value. Use the 30-minute break to create and fund a fresh set of wallets and adjust your scripts.

Objective
Maximize successful cross-shard transactions within the time window and fee budget
Network
Post-Supernova shadow fork — 600ms block times
Transaction type
Cross-shard MoveBalance (sender and receiver in different shards)
Window
17:00–17:30 UTC (30 minutes)
Fee budget per guild
500 EGLD — budget for fees and transaction value. Set transaction value to minimum of 0.01 EGLD. Fee spend and minimum transaction value will be verified. Subject to final confirmation.
Max sending wallets per guild
500 — must be a different set of newly created wallets from Part 1. See Wallet Rules.
Metric
Total successful cross-shard transactions submitted by all guild wallets during the window. Only cross-shard sends count.


Funding & Wallet Setup
At 15:45 UTC, 2,500 EGLD is sent directly to each guild leader’s registered wallet address. There is no faucet — funds arrive automatically. You have 15 minutes to distribute to your Part 1 wallets before the challenge starts. Part 2 wallet creation and funding happens during the break.

Part
Fee Budget
Max Wallets
Part 1
2,000 EGLD
500
Part 2
500 EGLD
500
Total distributed
2,500 EGLD *




Fee budget — important: The EGLD amounts above are fee and transaction value budgets, not transfer values. Set the value of each transaction to the minimum required (1×10⁻¹⁸ EGLD for Part 1; 0.01 EGLD for Part 2). Fee spend will be verified against the budget cap. Exceeding the cap will affect your score.

Minimum transaction value: Part 1 and Part 2 have different minimum transaction values. These are enforced separately. Review and update your scripts for each part.

Unused budget: Unused EGLD from Part 1 cannot be rolled into Part 2. The Part 2 budget is fixed at 500 EGLD regardless of Part 1 spend.

Tracing: All guild activity is traced back through the guild leader wallet. Each sending wallet must be funded directly by the guild leader wallet — no intermediary wallets. A chain of guild leader → wallet A → wallet B will not be counted. Only wallets funded directly by the guild leader wallet are attributed to your guild.

GL wallet role: The guild leader wallet is a funding and operator wallet only. It receives the 2,500 EGLD and distributes to your sending wallets. The guild leader wallet must not submit any MoveBalance transactions itself — only the sending wallets should appear in the transaction data. Keep the GL wallet clean.

Cross-shard requirement: Only transactions where the sender and receiver are in different shards are counted. Wallet placement is determined by address — distribute your wallets across all 3 shards.

Note: Unused EGLD from Part 1 cannot be carried into Part 2. The Part 2 budget is fixed at 500 EGLD regardless of Part 1 spending.


Wallet Rules
Each guild is allowed a maximum of 500 unique sending wallets per part.
Part 1 and Part 2 must each use a different set of 500 newly created wallets. Wallets used in Part 1 cannot be reused in Part 2. Wallets used in previous challenges cannot be reused in either part.
Wallet creation and fund distribution may happen before the challenge begins, but must occur after funds are distributed to the guild leader wallet at 15:45 UTC. Part 1 wallets should be created and funded in the 15-minute window before 16:00 UTC. Part 2 wallets should be created and funded during the 30-minute break.
The guild leader wallet is a funding and operator wallet only. It must not submit MoveBalance transactions. Only transactions from the sending wallets are counted.
Wallet addresses must be unique to your guild. If a wallet address is found to be shared with another guild — intentionally or not — transactions from that wallet will not be counted.
Guilds found to have deliberately used wallets belonging to other guilds may be disqualified for unsportsmanlike conduct.
All sending wallets must be funded directly by the guild leader wallet. No intermediary wallets are permitted. A chain where the guild leader funds wallet A and wallet A funds wallet B means wallet B will not be attributed to your guild.

Do not submit MoveBalance transactions from the guild leader wallet — it is a funding wallet only and will not be counted.
Do not reuse wallets between parts — Part 1 and Part 2 each require a fresh set of 500 previously-unused wallets.
Do not use intermediary wallets — all sending wallets must be funded directly by the guild leader wallet.
Do not use wallet addresses that may be in use by another guild — shared wallets will not count and may result in disqualification.
Only cross-shard transactions count — intra-shard transactions are excluded from scoring entirely.


Technical Setup
Step 1 — Before 15:45 UTC
Ensure your transaction scripts target cross-shard sends — sender and receiver must be on different shards
Plan your wallet distribution across all 3 shards to maximize valid cross-shard sends
Confirm Part 1 and Part 2 have different minimum transaction values — configure your scripts for each part separately
Prepare two wallet generation scripts: one for Part 1 (run at 15:45 UTC) and one for Part 2 (run during the break)
Test your scripts on the post-Supernova network to verify cross-shard transaction detection
Have your distribution script ready to run the moment funds arrive
Post your pre-challenge update on Twitter/X and tag @MultiversX

Step 2 — At 15:45 UTC
Receive 2,500 EGLD in your guild leader wallet
Generate your 500 Part 1 wallets — must be newly created
Distribute approximately 2,000 EGLD directly from the guild leader wallet to all 500 Part 1 wallets — no intermediary wallets
Reserve 500 EGLD for Part 2

Step 3 — At 16:00 UTC
Part 1 starts — fire your transaction scripts
Monitor live scores at bon.multiversx.com/guild-wars — updated every ~30 seconds
Step 4 — At 16:30 UTC (Break)
Part 1 closes — stop submission scripts
Review live scores and assess your position
Generate your 500 Part 2 wallets — must be a new set, different from Part 1
Distribute 500 EGLD directly from the guild leader wallet to all 500 Part 2 wallets
Update scripts for Part 2 — minimum transaction value increases to 0.01 EGLD
Prepare for Part 2 start at 17:00 UTC

Step 5 — At 17:00 UTC
Part 2 starts — fire your optimized scripts
Continue live updates on Twitter/X
Monitor milestone progress — first 3 guilds to 1,000,000 cross-shard transactions earn bonus points

Step 6 — After 17:30 UTC
Challenge closes — all onchain activity stops
Technical scores are final — computed from onchain data
You have 12 hours (until 05:30 UTC March 25) to submit content tasks via the portal
Post your recap thread, video, and push your scripts to GitHub
Submit all content links at bon.multiversx.com/guild-wars — guild leader only

Scoring
Scoring structure is identical to previous challenges. Technical scores are calculated from onchain data. Content scores from portal submissions. Final scores published after the content submission window closes.

Component
Max Score
Weight
How It Works
Technical Score
100 pts
75%
Top guild = 100 pts. Others scored proportionally. Tracked live every ~30 seconds. Max contribution: 75 pts.
Content Score
100 pts
25%
Sum of completed social tasks. All guilds can achieve full marks independently. Max contribution: 25 pts.
Milestone Bonus
+10/+7/+5
Flat
First 3 guilds to cross 1,000,000 cross-shard transactions. Independent of leaderboard position.
Maximum per challenge
110 pts


Tech 75 + Content 25 + Bonus 10


Total = (Tech Score × 0.75) + (Content Score × 0.25) + BonusTech Score = (guild cross-shard tx / top guild cross-shard tx) × 100 [max 100 pts → contributes max 75]Content Score = sum of completed tasks [max 100 pts → contributes max 25]Bonus = +10 / +7 / +5 (1st / 2nd / 3rd to cross 1,000,000 cross-shard tx)


Technical Score
Formula: (your guild’s total cross-shard tx / top guild’s total cross-shard tx) × 100

The guild with the most cross-shard transactions receives 100 points. All other guilds are scored proportionally. Live scores are visible during the challenge at bon.multiversx.com/guild-wars and updated approximately every 30 seconds.

Content Score
Sum of completed content tasks, max 100 points. All guilds can achieve full marks independently. Submitted by the guild leader only via the portal within 12 hours of challenge close.

Important — posting deadlines apply: Tasks 1 and 2 must be posted during the challenge window. Tasks 3, 4, and 5 can be posted any time within the 12-hour submission window, but must be submitted via the portal before 05:30 UTC March 25.

Content Tasks
Prerequisites (unscored — required to be eligible for content scoring)
Prerequisite
How to complete
Guild Twitter/X account exists, name identifiable as the guild
Must be set up and submitted via portal before challenge starts.
Guild GitHub repo exists and is public
Must be set up and submitted via portal before challenge starts.


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
16:00–17:30 UTC
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
Public repo updated after challenge closes. Content must be relevant to Crossover.
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
A flat bonus is awarded to the first three guilds to cross 1,000,000 total cross-shard transactions during the challenge window (both parts combined). Tracked in real time. Independent of final leaderboard position.

Milestone
Bonus
1st guild to cross 1,000,000 cross-shard transactions
+10 pts
2nd guild to cross 1,000,000 cross-shard transactions
+7 pts
3rd guild to cross 1,000,000 cross-shard transactions
+5 pts


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
mx-chain-go: github.com/multiversx/mx-chain-go 
mx-chain-scripts: github.com/multiversx/mx-chain-scripts 
mx-sdk-py-cli: github.com/multiversx/mx-sdk-py-cli 