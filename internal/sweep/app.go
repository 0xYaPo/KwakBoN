package sweep

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"sort"
	"strings"
	"time"

	"kwakbon/internal/esdt"
	"kwakbon/internal/gateway"
	"kwakbon/internal/manifest"
	"kwakbon/internal/txsign"
	"kwakbon/internal/wallets"
)

type Options struct {
	DryRun bool
}

type App struct {
	cfg       Config
	opt       Options
	gw        *gateway.Client
	records   []manifest.WalletRecord
	minRemain *big.Int
	feeBase   *big.Int
}

type sweepTarget struct {
	WalletID    string
	Address     string
	Shard       int
	Status      string
	BalanceBase *big.Int
	SweepBase   *big.Int
	PemPath     string
}

func New(cfg Config, opt Options) (*App, error) {
	gw := gateway.New(cfg.GatewayURL, cfg.HTTPTimeout)
	mf, err := manifest.Load(cfg.ManifestPath)
	if err != nil {
		return nil, err
	}
	if err := mf.Validate(); err != nil {
		return nil, err
	}

	minRemain, err := esdt.AmountToBaseUnits(cfg.MinRemainEGLD, 18)
	if err != nil {
		return nil, fmt.Errorf("invalid SWEEP_MIN_REMAIN_EGLD: %w", err)
	}
	feeBase := new(big.Int).Mul(new(big.Int).SetUint64(cfg.GasPrice), new(big.Int).SetUint64(cfg.GasLimit))

	records := make([]manifest.WalletRecord, 0, len(mf.Wallets))
	for _, record := range mf.Wallets {
		if !record.Enabled {
			continue
		}
		if strings.EqualFold(record.Address, cfg.TreasuryAddress) {
			continue
		}
		if cfg.ShardFilter >= 0 && record.Shard != cfg.ShardFilter {
			continue
		}
		if len(cfg.IncludeStatuses) > 0 && !containsFold(cfg.IncludeStatuses, record.Status) {
			continue
		}
		if containsFold(cfg.ExcludeStatuses, record.Status) {
			continue
		}
		if len(cfg.RequiredTags) > 0 && !containsAllTags(record.Tags, cfg.RequiredTags) {
			continue
		}
		if record.PemPath == "" {
			continue
		}
		records = append(records, record)
	}

	return &App{
		cfg:       cfg,
		opt:       opt,
		gw:        gw,
		records:   records,
		minRemain: minRemain,
		feeBase:   feeBase,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	if len(a.records) == 0 {
		return fmt.Errorf("no manifest wallets matched the sweep filters")
	}

	targets, totalSweep, err := a.computeTargets(ctx)
	if err != nil {
		return err
	}
	log.Printf("[sweepwallets] treasury=%s matched=%d sweepable=%d shardFilter=%d include=%s tags=%s totalSweepBase=%s", a.cfg.TreasuryAddress, len(a.records), len(targets), a.cfg.ShardFilter, strings.Join(a.cfg.IncludeStatuses, ","), strings.Join(a.cfg.RequiredTags, ","), totalSweep.String())

	if a.opt.DryRun {
		for i := 0; i < min(10, len(targets)); i++ {
			t := targets[i]
			log.Printf("[dry-run] wallet=%s status=%s shard=%d balance=%s sweep=%s", t.Address, t.Status, t.Shard, t.BalanceBase.String(), t.SweepBase.String())
		}
		return nil
	}

	hashes := make([]string, 0, len(targets))
	for i, target := range targets {
		hash, err := a.sendSweep(ctx, target)
		if err != nil {
			return fmt.Errorf("sweep %s: %w", target.Address, err)
		}
		hashes = append(hashes, hash)
		if (i+1)%25 == 0 || i+1 == len(targets) {
			log.Printf("[sweepwallets] sent=%d/%d", i+1, len(targets))
		}
	}

	if !a.cfg.WaitConfirm {
		return nil
	}
	successes, failed, err := a.waitAll(ctx, hashes)
	if err != nil {
		return err
	}
	log.Printf("[sweepwallets] confirmations done success=%d failed=%d total=%d", successes, failed, len(hashes))
	if successes != len(hashes) {
		return fmt.Errorf("only %d/%d sweep txs succeeded", successes, len(hashes))
	}
	return nil
}

func (a *App) computeTargets(ctx context.Context) ([]sweepTarget, *big.Int, error) {
	out := make([]sweepTarget, 0, len(a.records))
	total := big.NewInt(0)
	threshold := new(big.Int).Add(new(big.Int).Set(a.feeBase), a.minRemain)

	for _, record := range a.records {
		balanceStr, err := a.gw.GetAccountBalance(ctx, record.Address)
		if err != nil {
			return nil, nil, fmt.Errorf("get balance for %s: %w", record.Address, err)
		}
		balance, ok := new(big.Int).SetString(balanceStr, 10)
		if !ok {
			return nil, nil, fmt.Errorf("invalid balance for %s: %s", record.Address, balanceStr)
		}
		if balance.Cmp(threshold) <= 0 {
			continue
		}

		sweepValue := new(big.Int).Sub(balance, a.feeBase)
		sweepValue.Sub(sweepValue, a.minRemain)
		if sweepValue.Sign() <= 0 {
			continue
		}

		total.Add(total, sweepValue)
		out = append(out, sweepTarget{
			WalletID:    record.WalletID,
			Address:     record.Address,
			Shard:       record.Shard,
			Status:      record.Status,
			BalanceBase: balance,
			SweepBase:   sweepValue,
			PemPath:     record.PemPath,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Status == out[j].Status {
			return out[i].WalletID < out[j].WalletID
		}
		return out[i].Status < out[j].Status
	})

	return out, total, nil
}

func (a *App) sendSweep(ctx context.Context, target sweepTarget) (string, error) {
	signer, err := wallets.NewPemSigner(target.PemPath, target.Address)
	if err != nil {
		return "", fmt.Errorf("load signer: %w", err)
	}
	nonce, err := a.gw.GetAccountNonce(ctx, target.Address)
	if err != nil {
		return "", fmt.Errorf("get nonce: %w", err)
	}

	tx := gateway.TxSendRequest{
		Nonce:    nonce,
		Value:    target.SweepBase.String(),
		Receiver: a.cfg.TreasuryAddress,
		Sender:   target.Address,
		GasPrice: a.cfg.GasPrice,
		GasLimit: a.cfg.GasLimit,
		Data:     "",
		ChainID:  a.cfg.ChainID,
		Version:  a.cfg.TxVersion,
	}

	unsigned := txsign.UnsignedTx{
		Nonce:    int64(tx.Nonce),
		Value:    tx.Value,
		Receiver: tx.Receiver,
		Sender:   tx.Sender,
		GasPrice: tx.GasPrice,
		GasLimit: tx.GasLimit,
		Data:     "",
		ChainID:  tx.ChainID,
		Version:  tx.Version,
	}

	toSign, err := txsign.SerializeForSigning(unsigned)
	if err != nil {
		return "", err
	}
	sig, err := signer.SignTxBytes(toSign)
	if err != nil {
		return "", err
	}
	tx.Signature = hex.EncodeToString(sig)
	return a.gw.SendTx(ctx, tx)
}

func (a *App) waitAll(ctx context.Context, hashes []string) (int, int, error) {
	pending := make(map[string]struct{}, len(hashes))
	for _, h := range hashes {
		pending[h] = struct{}{}
	}
	successes := 0
	failed := 0

	for len(pending) > 0 {
		for h := range pending {
			status, err := a.gw.GetTxStatus(ctx, h)
			if err != nil {
				continue
			}
			if gateway.IsFinalSuccessStatus(status) {
				successes++
				delete(pending, h)
				continue
			}
			if gateway.IsFinalFailureStatus(status) {
				failed++
				delete(pending, h)
			}
		}
		log.Printf("[sweepwallets confirm] pending=%d success=%d failed=%d", len(pending), successes, failed)
		if len(pending) == 0 {
			break
		}
		select {
		case <-ctx.Done():
			return successes, failed, ctx.Err()
		case <-time.After(a.cfg.PollInterval):
		}
	}

	return successes, failed, nil
}

func containsFold(items []string, want string) bool {
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item), want) {
			return true
		}
	}
	return false
}

func containsAllTags(tags []string, wants []string) bool {
	for _, want := range wants {
		if !containsFold(tags, want) {
			return false
		}
	}
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
