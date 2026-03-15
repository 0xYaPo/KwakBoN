#!/usr/bin/env bash
# Check treasury wallet balance on multiple networks.
# Usage:  ./scripts/check_treasury.sh [erd1address]
# If no address given, reads the treasury entry from configs/wallets-manifest.json.

set -euo pipefail

MANIFEST="${WALLETS_MANIFEST:-configs/wallets-manifest.json}"

# ── resolve address ────────────────────────────────────────────────────────────
if [[ $# -ge 1 ]]; then
  ADDRESS="$1"
else
  ADDRESS=$(python3 -c "
import json, sys
with open('$MANIFEST') as f:
    data = json.load(f)
wallets = data['wallets'] if isinstance(data, dict) else data
treasury = [w for w in wallets if w.get('status') == 'treasury']
if not treasury:
    print('ERROR: no treasury wallet found in manifest', file=sys.stderr)
    sys.exit(1)
print(treasury[0]['address'])
")
fi

echo "address : $ADDRESS"
echo ""

# ── query one network ──────────────────────────────────────────────────────────
check_network() {
  local label="$1"
  local api="$2"

  local raw
  raw=$(curl -sf --max-time 10 "${api}/address/${ADDRESS}" 2>/dev/null || true)

  if [[ -z "$raw" ]]; then
    printf "  %-12s  unreachable\n" "$label"
    return
  fi

  local balance nonce shard
  balance=$(echo "$raw" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('data',{}).get('account',{}).get('balance','?'))")
  nonce=$(echo   "$raw" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('data',{}).get('account',{}).get('nonce','?'))")

  # shard comes from the /accounts/ endpoint (API, not proxy)
  local shard_raw
  shard_raw=$(curl -sf --max-time 10 "${api}/accounts/${ADDRESS}" 2>/dev/null || true)
  if [[ -n "$shard_raw" ]]; then
    shard=$(echo "$shard_raw" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('shard','?'))" 2>/dev/null || echo "?")
  else
    shard="?"
  fi

  local egld
  egld=$(python3 -c "
b = '$balance'
try:
    v = int(b)
    print(f'{v / 10**18:.6f}')
except:
    print(b)
")

  printf "  %-12s  %s EGLD   (nonce=%s  shard=%s)\n" "$label" "$egld" "$nonce" "$shard"
}

# ── networks ───────────────────────────────────────────────────────────────────
check_network "mainnet"  "https://api.multiversx.com"
check_network "devnet"   "https://devnet-api.multiversx.com"
check_network "battle"   "https://api.battleofnodes.com"
