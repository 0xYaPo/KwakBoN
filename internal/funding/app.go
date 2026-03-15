package funding

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"sort"
	"strings"
	"sync"
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
	cfg      Config
	opt      Options
	gw       *gateway.Client
	treasury *wallets.PemSigner
	records  []manifest.WalletRecord
}

type fundingTarget struct {
	WalletID      string
	Address       string
	Shard         int
	Status        string
	CurrentBase   *big.Int
	TargetBase    *big.Int
	DeficitBase   *big.Int
}

func New(cfg Config, opt Options) (*App, error) {
	gw := gateway.New(cfg.GatewayURL, cfg.HTTPTimeout)
	var signer *wallets.PemSigner
	if !opt.DryRun {
		loadedSigner, err := wallets.NewPemSigner(cfg.TreasuryPemPath, cfg.TreasuryAddress)
		if err != nil {
			return nil, fmt.Errorf("load treasury signer: %w", err)
		}
		signer = loadedSigner
	}
	mf, err := manifest.Load(cfg.ManifestPath)
	if err != nil {
		return nil, err
	}
	if err := mf.Validate(); err != nil {
		return nil, err
	}

	records := make([]manifest.WalletRecord, 0, len(mf.Wallets))
	for _, record := range mf.Wallets {
		if !record.Enabled {
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
		if _, ok := cfg.TargetAmountByStatus[record.Status]; !ok {
			continue
		}
		records = append(records, record)
	}

	return &App{
		cfg:      cfg,
		opt:      opt,
		gw:       gw,
		treasury: signer,
		records:  records,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	if len(a.records) == 0 {
		return fmt.Errorf("no manifest wallets matched the funding filters")
	}

	treasuryNonce, err := a.gw.GetAccountNonce(ctx, a.cfg.TreasuryAddress)
	if err != nil {
		return fmt.Errorf("get treasury nonce: %w", err)
	}

	targets, totalDeficit, err := a.computeTargets(ctx)
	if err != nil {
		return err
	}

	log.Printf("[fundwallets] treasury=%s matched=%d needingTopUp=%d shardFilter=%d include=%s tags=%s totalDeficitBase=%s", a.cfg.TreasuryAddress, len(a.records), len(targets), a.cfg.ShardFilter, strings.Join(a.cfg.IncludeStatuses, ","), strings.Join(a.cfg.RequiredTags, ","), totalDeficit.String())

	if a.opt.DryRun {
		for i := 0; i < min(10, len(targets)); i++ {
			t := targets[i]
			log.Printf("[dry-run] wallet=%s status=%s shard=%d current=%s target=%s deficit=%s", t.Address, t.Status, t.Shard, t.CurrentBase.String(), t.TargetBase.String(), t.DeficitBase.String())
		}
		return nil
	}

	nonce := treasuryNonce
	hashes := make([]string, 0, len(targets))
	for i, target := range targets {
		hash, err := a.sendTopUp(ctx, nonce, target)
		if err != nil {
			return fmt.Errorf("fund %s: %w", target.Address, err)
		}
		nonce++
		hashes = append(hashes, hash)
		if (i+1)%25 == 0 || i+1 == len(targets) {
			log.Printf("[fundwallets] sent=%d/%d", i+1, len(targets))
		}
	}

	if !a.cfg.WaitConfirm {
		return nil
	}
	successes, failed, err := a.waitAll(ctx, hashes)
	if err != nil {
		return err
	}
	log.Printf("[fundwallets] confirmations done success=%d failed=%d total=%d", successes, failed, len(hashes))
	if successes != len(hashes) {
		return fmt.Errorf("only %d/%d funding txs succeeded", successes, len(hashes))
	}
	return nil
}

func (a *App) computeTargets(ctx context.Context) ([]fundingTarget, *big.Int, error) {
	type result struct {
		target fundingTarget
		skip   bool
		err    error
	}

	results := make([]result, len(a.records))
	sem := make(chan struct{}, 32)
	var wg sync.WaitGroup

	for i, record := range a.records {
		wg.Add(1)
		go func(i int, record manifest.WalletRecord) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			targetHuman := a.cfg.TargetAmountByStatus[record.Status]
			targetBase, err := esdt.AmountToBaseUnits(targetHuman, 18)
			if err != nil {
				results[i] = result{err: fmt.Errorf("invalid target amount for status %s: %w", record.Status, err)}
				return
			}
			balanceStr, err := a.gw.GetAccountBalance(ctx, record.Address)
			if err != nil {
				results[i] = result{err: fmt.Errorf("get balance for %s: %w", record.Address, err)}
				return
			}
			currentBase, ok := new(big.Int).SetString(balanceStr, 10)
			if !ok {
				results[i] = result{err: fmt.Errorf("invalid balance for %s: %s", record.Address, balanceStr)}
				return
			}
			if currentBase.Cmp(targetBase) >= 0 {
				results[i] = result{skip: true}
				return
			}
			deficit := new(big.Int).Sub(targetBase, currentBase)
			results[i] = result{target: fundingTarget{
				WalletID:    record.WalletID,
				Address:     record.Address,
				Shard:       record.Shard,
				Status:      record.Status,
				CurrentBase: currentBase,
				TargetBase:  targetBase,
				DeficitBase: deficit,
			}}
		}(i, record)
	}
	wg.Wait()

	out := make([]fundingTarget, 0, len(a.records))
	total := big.NewInt(0)
	for _, r := range results {
		if r.err != nil {
			return nil, nil, r.err
		}
		if r.skip {
			continue
		}
		total.Add(total, r.target.DeficitBase)
		out = append(out, r.target)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Status == out[j].Status {
			return out[i].WalletID < out[j].WalletID
		}
		return out[i].Status < out[j].Status
	})

	return out, total, nil
}

func (a *App) sendTopUp(ctx context.Context, nonce uint64, target fundingTarget) (string, error) {
	tx := gateway.TxSendRequest{
		Nonce:    nonce,
		Value:    target.DeficitBase.String(),
		Receiver: target.Address,
		Sender:   a.cfg.TreasuryAddress,
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
	sig, err := a.treasury.SignTxBytes(toSign)
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
		log.Printf("[fundwallets confirm] pending=%d success=%d failed=%d", len(pending), successes, failed)
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
