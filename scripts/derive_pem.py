#!/usr/bin/env python3
"""Derive a MultiversX wallet PEM from a 24-word BIP39 mnemonic.

Usage:
    python3 scripts/derive_pem.py <seed.txt> <output.pem> [account_index]

account_index defaults to 0 (first wallet on path m/44'/508'/0'/0'/0').

The output PEM uses the MultiversX format: base64(hex(seed32))
with header "PRIVATE KEY for erd1...".
"""

import sys
import hashlib
import hmac
import struct
import unicodedata
import base64


# ---------------------------------------------------------------------------
# BIP39: mnemonic → 64-byte seed
# ---------------------------------------------------------------------------

def bip39_to_seed(mnemonic: str, passphrase: str = "") -> bytes:
    m = unicodedata.normalize("NFKD", mnemonic).encode("utf-8")
    s = unicodedata.normalize("NFKD", "mnemonic" + passphrase).encode("utf-8")
    return hashlib.pbkdf2_hmac("sha512", m, s, 2048)


# ---------------------------------------------------------------------------
# SLIP-0010: seed → Ed25519 private key via hardened BIP44 path
# ---------------------------------------------------------------------------

def slip10_derive(seed: bytes, path: list) -> bytes:
    """Return 32-byte private key at the given hardened path."""
    I = hmac.new(b"ed25519 seed", seed, hashlib.sha512).digest()
    kL, kR = I[:32], I[32:]
    for index in path:
        if index < 0x80000000:
            raise ValueError(f"Ed25519 SLIP-0010 only supports hardened indices, got {index}")
        data = b"\x00" + kL + struct.pack(">I", index)
        I = hmac.new(kR, data, hashlib.sha512).digest()
        kL, kR = I[:32], I[32:]
    return kL


# ---------------------------------------------------------------------------
# Ed25519 public key derivation
# ---------------------------------------------------------------------------

def ed25519_pubkey(seed32: bytes) -> bytes:
    try:
        from cryptography.hazmat.primitives.asymmetric.ed25519 import Ed25519PrivateKey
    except ImportError:
        print("ERROR: install the cryptography package:  pip3 install cryptography", file=sys.stderr)
        sys.exit(1)
    priv = Ed25519PrivateKey.from_private_bytes(seed32)
    return priv.public_key().public_bytes_raw()


# ---------------------------------------------------------------------------
# Bech32 / MultiversX address encoding (erd1...)
# Mirrors the Go implementation in internal/wallets/bech32_erd.go
# ---------------------------------------------------------------------------

CHARSET = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"


def _convertbits(data, frombits, tobits, pad=True):
    acc, bits, ret = 0, 0, []
    maxv = (1 << tobits) - 1
    maxacc = (1 << (frombits + tobits - 1)) - 1
    for v in data:
        acc = ((acc << frombits) | v) & maxacc
        bits += frombits
        while bits >= tobits:
            bits -= tobits
            ret.append((acc >> bits) & maxv)
    if pad and bits:
        ret.append((acc << (tobits - bits)) & maxv)
    return ret


def _polymod(values):
    GEN = [0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3]
    chk = 1
    for v in values:
        b = chk >> 25
        chk = ((chk & 0x1ffffff) << 5) ^ v
        for i in range(5):
            if (b >> i) & 1:
                chk ^= GEN[i]
    return chk


def _hrp_expand(hrp):
    return [ord(x) >> 5 for x in hrp] + [0] + [ord(x) & 31 for x in hrp]


def _bech32_encode(hrp, data5):
    combined = data5 + [0] * 6
    mod = _polymod(_hrp_expand(hrp) + combined) ^ 1
    checksum = [(mod >> (5 * (5 - i))) & 31 for i in range(6)]
    return hrp + "1" + "".join(CHARSET[d] for d in data5 + checksum)


def pubkey_to_erd(pubkey32: bytes) -> str:
    data5 = _convertbits(list(pubkey32), 8, 5, True)
    return _bech32_encode("erd", data5)


# ---------------------------------------------------------------------------
# PEM writer — MultiversX format: base64(hex(seed32))
# ---------------------------------------------------------------------------

def write_mvx_pem(seed32: bytes, address: str, outpath: str):
    hex_seed = seed32.hex()                      # 64-char hex string
    b64 = base64.b64encode(hex_seed.encode()).decode()
    lines = [b64[i:i+64] for i in range(0, len(b64), 64)]
    tag = f"PRIVATE KEY for {address}"
    with open(outpath, "w") as f:
        f.write(f"-----BEGIN {tag}-----\n")
        for line in lines:
            f.write(line + "\n")
        f.write(f"-----END {tag}-----\n")


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    if len(sys.argv) < 3:
        print(__doc__)
        sys.exit(1)

    seed_file = sys.argv[1]
    out_pem   = sys.argv[2]
    account   = int(sys.argv[3]) if len(sys.argv) > 3 else 0

    with open(seed_file) as f:
        tokens = f.read().split()

    # Support numbered format: "1. word 2. word ..." or "1 word 2 word ..."
    words = [t for t in tokens if not t.rstrip(".").isdigit()]
    mnemonic = " ".join(words)

    if len(words) != 24:
        print(f"ERROR: expected 24 words, got {len(words)}", file=sys.stderr)
        sys.exit(1)

    bip39_seed = bip39_to_seed(mnemonic)

    H = 0x80000000
    path = [H | 44, H | 508, H | 0, H | 0, H | account]
    seed32 = slip10_derive(bip39_seed, path)

    pubkey  = ed25519_pubkey(seed32)
    address = pubkey_to_erd(pubkey)

    write_mvx_pem(seed32, address, out_pem)

    print(f"address : {address}")
    print(f"account : {account}  (path m/44'/508'/0'/0'/{account}')")
    print(f"pem     : {out_pem}")


if __name__ == "__main__":
    main()
