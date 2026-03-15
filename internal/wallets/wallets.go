package wallets

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"kwakbon/internal/manifest"
)

type Wallet struct {
	Address   string
	NextNonce uint64
	Signer    Signer

	mu sync.Mutex
}

func (w *Wallet) SetNonce(n uint64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.NextNonce = n
}

func (w *Wallet) ReserveNonce() uint64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	n := w.NextNonce
	w.NextNonce++
	return n
}

type walletFile struct {
	Wallets []struct {
		Address string `json:"address"`
		Nonce   uint64 `json:"nonce"`
		PemPath string `json:"pemPath"`
	} `json:"wallets"`
	Senders []struct {
		Address string `json:"address"`
		Nonce   uint64 `json:"nonce"`
		PemPath string `json:"pemPath"`
	} `json:"senders"`
}

func Load(path string) ([]*Wallet, error) {
	if mf, err := manifest.Load(path); err == nil {
		if err := mf.Validate(); err != nil {
			return nil, err
		}
		return fromManifest(mf), nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read wallets file: %w", err)
	}

	var wf walletFile
	if err := json.Unmarshal(b, &wf); err != nil {
		return nil, fmt.Errorf("parse wallets file: %w", err)
	}

	items := wf.Wallets
	if len(items) == 0 {
		items = wf.Senders
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no wallets in %s", path)
	}

	out := make([]*Wallet, 0, len(items))
	for _, item := range items {
		var signer Signer
		if item.PemPath != "" {
			ps, err := NewPemSigner(item.PemPath, item.Address)
			if err != nil {
				return nil, fmt.Errorf("load pem signer for %s: %w", item.Address, err)
			}
			signer = ps
		}

		out = append(out, &Wallet{
			Address:   item.Address,
			NextNonce: item.Nonce,
			Signer:    signer,
		})
	}

	return out, nil
}

func fromManifest(mf *manifest.Manifest) []*Wallet {
	out := make([]*Wallet, 0, len(mf.Wallets))
	for _, item := range mf.Wallets {
		var signer Signer
		if item.PemPath != "" {
			ps, err := NewPemSigner(item.PemPath, item.Address)
			if err == nil {
				signer = ps
			}
		}

		out = append(out, &Wallet{
			Address:   item.Address,
			NextNonce: 0,
			Signer:    signer,
		})
	}
	return out
}
