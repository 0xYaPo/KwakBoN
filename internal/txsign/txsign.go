package txsign

import (
	"encoding/json"
	"fmt"
)

func SerializeForSigning(tx UnsignedTx) ([]byte, error) {
	if tx.Nonce < 0 {
		return nil, fmt.Errorf("invalid nonce")
	}

	payload := frontendUnsignedTx{
		Nonce:    uint64(tx.Nonce),
		Value:    tx.Value,
		Receiver: tx.Receiver,
		Sender:   tx.Sender,
		GasPrice: tx.GasPrice,
		GasLimit: tx.GasLimit,
		Data:     []byte(tx.Data),
		ChainID:  tx.ChainID,
		Version:  tx.Version,
		Options:  tx.Options,
		Guardian: tx.Guardian,
		Relayer:  tx.Relayer,
	}

	return json.Marshal(payload)
}

type frontendUnsignedTx struct {
	Nonce    uint64 `json:"nonce"`
	Value    string `json:"value"`
	Receiver string `json:"receiver"`
	Sender   string `json:"sender"`
	GasPrice uint64 `json:"gasPrice"`
	GasLimit uint64 `json:"gasLimit"`
	Data     []byte `json:"data,omitempty"`
	ChainID  string `json:"chainID"`
	Version  uint32 `json:"version"`
	Options  uint32 `json:"options,omitempty"`
	Guardian string `json:"guardian,omitempty"`
	Relayer  string `json:"relayer,omitempty"`
}

type UnsignedTx struct {
	Nonce    int64
	Value    string
	Receiver string
	Sender   string
	GasPrice uint64
	GasLimit uint64
	Data     string
	ChainID  string
	Version  uint32
	Options  uint32
	Guardian string
	Relayer  string
}
