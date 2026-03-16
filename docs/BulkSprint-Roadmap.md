# BulkSprint Roadmap

Status:
- idea backlog only
- not planned for immediate implementation
- keep current focus on validating the existing live strategy first

Current live preference:
- use `windowsprint` as the baseline sender
- optionally switch to `bulksprint` for a late finish mode
- do not run both in parallel

Target live strategy to validate when Supernova is fully live:
- `windowsprint` for the first `20` minutes
- `bulksprint` for the last `10` minutes

## Quick Wins

Low risk, high value improvements for `bulksprint`.

### 1. Per-shard metrics

Add:
- sent per shard
- accepted per shard
- pending per shard
- errors per shard

Why:
- much better live visibility
- easier to spot shard-specific degradation

### 2. Better `accepted` vs `final success` tracking

Add:
- clearer logs on acceptance vs finality
- confirmation catch-up visibility
- better summary after each run

Why:
- this is the key distinction between a fast sender and a useful competition sender

### 3. Smarter partial-accept retry

Improve:
- isolate the unaccepted subset after `send-multiple`
- retry only the remainder
- log partial acceptance clearly

Why:
- better control under imperfect API behavior

### 4. Bulk budget guardrails

Add:
- soft fee stop
- hard fee stop
- explicit fee budget awareness in bulk mode

Why:
- bulk can burn budget very quickly when healthy

## Medium Complexity

Useful improvements with moderate implementation risk.

### 5. Partial pre-generation mode

Idea:
- pre-build and pre-sign a small rolling buffer of batches
- regenerate in the background as batches are sent

Why:
- keeps the blast advantage
- avoids the rigidity of full pre-generation

### 6. Explicit shard scheduler

Add:
- clearer shard-level orchestration
- shard-level knobs if needed
- shard-level operational visibility

Why:
- more control
- better live diagnosis

### 7. Read-path tuning

Explore:
- nonce/balance read frequency
- concurrent read limits
- better account-state source if available

Why:
- bulk performance depends strongly on account-state freshness

### 8. Easier finish-mode operation

Improve:
- operator-friendly start/stop flow
- tighter logs
- simpler late-window invocation

Why:
- likely the most realistic live use case

## High Risk / High Upside

Interesting, but not for now.

### 9. Full pre-generation mode

Idea:
- generate and sign the whole run upfront

Upside:
- maximum local send speed

Risks:
- rigid
- harder to recover from partial acceptance
- harder to adapt live

### 10. Multi-endpoint bulk strategy

Idea:
- split or rotate bulk traffic across multiple public endpoints

Upside:
- could reduce frontend API pressure

Risks:
- more difficult diagnosis
- may not help if the backend path is effectively shared

### 11. Dynamic bulk adaptation

Idea:
- adjust batch size or lookahead based on live behavior

Upside:
- more intelligent sender under changing conditions

Risks:
- more complexity
- harder to trust under competition pressure

## Priority Order

Recommended future order:
1. per-shard metrics
2. smarter partial-accept retry
3. budget guardrails
4. partial pre-generation
5. clearer shard scheduler

## Current Decision

Do not implement this roadmap yet.

Near-term priority:
- keep the current `bulksprint` implementation as-is
- validate the real Supernova strategy live first
- especially the sequential plan:
  - `windowsprint` first
  - `bulksprint` as optional late acceleration
