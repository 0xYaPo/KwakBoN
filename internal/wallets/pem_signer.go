package wallets

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
)

type Signer interface {
	SignTxBytes(txBytes []byte) ([]byte, error)
}

type PemSigner struct {
	priv ed25519.PrivateKey
}

func NewPemSigner(pemPath string, expectedBech32Sender string) (*PemSigner, error) {
	b, err := os.ReadFile(pemPath)
	if err != nil {
		return nil, fmt.Errorf("read pem: %w", err)
	}

	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("pem decode failed: no PEM block found in %s", pemPath)
	}

	body := strings.TrimSpace(string(block.Bytes))
	body = strings.ReplaceAll(body, "\n", "")
	body = strings.ReplaceAll(body, "\r", "")
	body = strings.ReplaceAll(body, " ", "")

	raw, err := hex.DecodeString(body)
	if err != nil {
		return nil, fmt.Errorf("pem body is not hex (expected MultiversX hex-in-base64 format): %w", err)
	}

	var priv ed25519.PrivateKey
	switch len(raw) {
	case ed25519.SeedSize:
		priv = ed25519.NewKeyFromSeed(raw)
	case ed25519.PrivateKeySize:
		priv = ed25519.PrivateKey(raw)
	default:
		return nil, fmt.Errorf("unexpected decoded key length: %d bytes (expected 32 seed)", len(raw))
	}

	if expected := strings.TrimSpace(expectedBech32Sender); expected != "" {
		pub := priv.Public().(ed25519.PublicKey)
		derived, derr := bech32EncodeErdFromPubKey(pub)
		if derr != nil {
			return nil, fmt.Errorf("derive bech32: %w", derr)
		}
		if derived != expected {
			return nil, fmt.Errorf("pem does not match sender address. expected=%s derived=%s", expected, derived)
		}
	}

	return &PemSigner{priv: priv}, nil
}

func (p *PemSigner) SignTxBytes(txBytes []byte) ([]byte, error) {
	if len(p.priv) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size")
	}
	return ed25519.Sign(p.priv, txBytes), nil
}
