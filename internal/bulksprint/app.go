package bulksprint

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"kwakbon/internal/gateway"
	"kwakbon/internal/manifest"
	"kwakbon/internal/txsign"
	"kwakbon/internal/wallets"
)

type Options struct {
	DryRun bool
}

type senderWallet struct {
	*wallets.Wallet
	Shard    int
	Receiver string
	Name     string
}

type App struct {
	cfg            Config
	opt            Options
	gw             *gateway.Client
	senders        []senderWallet
	readSemaphore  chan struct{}
	reservedTarget atomic.Uint64
	totalSent      atomic.Uint64
	totalAccepted  atomic.Uint64
	totalBatches   atomic.Uint64
	totalErrors    atomic.Uint64
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

	grouped := make(map[int][]senderWallet)
	for _, record := range mf.Wallets {
		if !record.Enabled {
			continue
		}
		if cfg.ShardFilter >= 0 && record.Shard != cfg.ShardFilter {
			continue
		}
		if !containsFold(cfg.SenderStatuses, record.Status) || !containsAllTags(record.Tags, cfg.RequiredTags) {
			continue
		}
		if record.PemPath == "" && !opt.DryRun {
			return nil, fmt.Errorf("sender %s missing pemPath", record.Address)
		}

		w := &wallets.Wallet{Address: record.Address}
		if !opt.DryRun {
			signer, err := wallets.NewPemSigner(record.PemPath, record.Address)
			if err != nil {
				return nil, fmt.Errorf("load signer for %s: %w", record.Address, err)
			}
			w.Signer = signer
		}
		grouped[record.Shard] = append(grouped[record.Shard], senderWallet{
			Wallet: w,
			Shard:  record.Shard,
			Name:   firstNonEmpty(record.WalletID, record.Address),
		})
	}

	if len(grouped) == 0 {
		return nil, fmt.Errorf("no bulk senders matched manifest filters")
	}

	senders := make([]senderWallet, 0)
	shards := make([]int, 0, len(grouped))
	for shard := range grouped {
		shards = append(shards, shard)
	}
	sort.Ints(shards)

	for _, shard := range shards {
		group := grouped[shard]
		if cfg.MaxActiveWallets > 0 && len(senders) >= cfg.MaxActiveWallets {
			break
		}
		sort.Slice(group, func(i, j int) bool {
			return strings.Compare(group[i].Address, group[j].Address) < 0
		})
		if len(group) < 2 {
			return nil, fmt.Errorf("shard %d has only %d sender(s); need at least 2 for ring routing", shard, len(group))
		}
		if cfg.MaxActiveWallets > 0 && len(senders)+len(group) > cfg.MaxActiveWallets {
			remaining := cfg.MaxActiveWallets - len(senders)
			if remaining < 2 {
				break
			}
			group = group[:remaining]
		}
		for i := range group {
			group[i].Receiver = group[(i+1)%len(group)].Address
		}
		senders = append(senders, group...)
	}

	if len(senders) == 0 {
		return nil, fmt.Errorf("no bulk senders available after max wallet filter")
	}

	return &App{
		cfg:           cfg,
		opt:           opt,
		gw:            gw,
		senders:       senders,
		readSemaphore: make(chan struct{}, cfg.MaxConcurrentReads),
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	log.Printf(
		"[bulksprint] gateway=%s senders=%d targetTx=%d duration=%s batchSize=%d lookahead=%d shardFilter=%d value=%s gasLimit=%d gasPrice=%d waitConfirm=%t confirmWorkers=%d confirmTimeout=%s",
		a.cfg.GatewayURL,
		len(a.senders),
		a.cfg.TargetTx,
		a.cfg.Duration,
		a.cfg.BatchSize,
		a.cfg.MaxNonceLookahead,
		a.cfg.ShardFilter,
		a.cfg.Value,
		a.cfg.GasLimit,
		a.cfg.GasPrice,
		a.cfg.WaitConfirm,
		a.cfg.ConfirmWorkers,
		a.cfg.ConfirmTimeout,
	)

	if a.opt.DryRun {
		for i := 0; i < min(5, len(a.senders)); i++ {
			log.Printf("[bulk dry-run %d] sender=%s shard=%d receiver=%s value=%s batchSize=%d", i+1, a.senders[i].Address, a.senders[i].Shard, a.senders[i].Receiver, a.cfg.Value, a.cfg.BatchSize)
		}
		return nil
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	var hashesCh chan []string
	var collectedHashes []string
	var collectorWG sync.WaitGroup
	var hashesMu sync.Mutex
	if a.cfg.WaitConfirm {
		hashesCh = make(chan []string, len(a.senders)*8)
		collectorWG.Add(1)
		go func() {
			defer collectorWG.Done()
			for hashes := range hashesCh {
				if len(hashes) == 0 {
					continue
				}
				hashesMu.Lock()
				collectedHashes = append(collectedHashes, hashes...)
				hashesMu.Unlock()
			}
		}()
	}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	start := time.Now()
	deadline := time.Time{}
	if a.cfg.Duration > 0 {
		deadline = start.Add(a.cfg.Duration)
		runCtx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
	}

	for i := range a.senders {
		wg.Add(1)
		go func(sw *senderWallet) {
			defer wg.Done()
			if err := a.runWallet(runCtx, deadline, sw, hashesCh); err != nil && runCtx.Err() == nil {
				select {
				case errCh <- err:
				default:
				}
				cancel()
			}
		}(&a.senders[i])
	}

	wg.Wait()
	if hashesCh != nil {
		close(hashesCh)
		collectorWG.Wait()
	}
	select {
	case err := <-errCh:
		return err
	default:
	}

	log.Printf(
		"[bulksprint] done sent=%d accepted=%d batches=%d errors=%d elapsed=%s",
		a.totalSent.Load(),
		a.totalAccepted.Load(),
		a.totalBatches.Load(),
		a.totalErrors.Load(),
		time.Since(start),
	)
	if !a.cfg.WaitConfirm {
		return nil
	}

	hashes := collectedHashes
	confirmCtx, confirmCancel := context.WithCancel(ctx)
	defer confirmCancel()
	successes, failed, err := a.waitForFinalStatuses(confirmCtx, hashes)
	if err != nil {
		return err
	}
	log.Printf("[bulksprint] confirmations done success=%d failed=%d total=%d", successes, failed, len(hashes))
	return nil
}

func (a *App) runWallet(ctx context.Context, deadline time.Time, sw *senderWallet, hashesCh chan<- []string) error {
	var localNonce uint64
	nonceInitialized := false
	perTxCost, ok := txTotalCost(a.cfg.Value, a.cfg.GasLimit, a.cfg.GasPrice)
	if !ok || perTxCost.Sign() <= 0 {
		return fmt.Errorf("invalid tx cost configuration")
	}

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !deadline.IsZero() && time.Now().After(deadline) {
			return nil
		}
		if a.cfg.TargetTx > 0 && int(a.reservedTarget.Load()) >= a.cfg.TargetTx {
			return nil
		}

		confirmedNonce, balanceRaw, err := a.getAccountSafe(ctx, sw.Address)
		if err != nil {
			if err := sleepWithContext(ctx, 500*time.Millisecond); err != nil {
				return err
			}
			continue
		}
		balance, ok := new(big.Int).SetString(balanceRaw, 10)
		if !ok {
			return fmt.Errorf("parse balance for %s: %q", sw.Address, balanceRaw)
		}
		if balance.Cmp(perTxCost) < 0 {
			return nil
		}
		if !nonceInitialized || confirmedNonce > localNonce {
			localNonce = confirmedNonce
			nonceInitialized = true
		}

		nonceGap := int(localNonce - confirmedNonce)
		nonceSlots := a.cfg.MaxNonceLookahead - nonceGap
		if nonceSlots <= 0 {
			if err := sleepWithContext(ctx, a.cfg.LoopIdleSleep); err != nil {
				return err
			}
			continue
		}

		maxByBalance := int(new(big.Int).Div(balance, perTxCost).Int64())
		if maxByBalance <= 0 {
			return nil
		}

		batchLimit := a.cfg.BatchSize
		batchLimit = min(batchLimit, nonceSlots)
		batchLimit = min(batchLimit, maxByBalance)
		batchCount, ok := a.reserveQuota(batchLimit)
		if !ok {
			return nil
		}
		if batchCount <= 0 {
			if err := sleepWithContext(ctx, a.cfg.LoopIdleSleep); err != nil {
				return err
			}
			continue
		}

		txs := make([]gateway.TxSendRequest, 0, batchCount)
		for j := 0; j < batchCount; j++ {
			tx := gateway.TxSendRequest{
				Nonce:    localNonce + uint64(j),
				Value:    a.cfg.Value,
				Receiver: sw.Receiver,
				Sender:   sw.Address,
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
				return fmt.Errorf("serialize tx nonce=%d sender=%s: %w", tx.Nonce, sw.Address, err)
			}
			sig, err := sw.Signer.SignTxBytes(toSign)
			if err != nil {
				return fmt.Errorf("sign tx nonce=%d sender=%s: %w", tx.Nonce, sw.Address, err)
			}
			tx.Signature = hex.EncodeToString(sig)
			txs = append(txs, tx)
		}

		hashes, err := a.gw.SendTxs(ctx, txs)
		if err != nil {
			a.releaseQuota(batchCount)
			a.totalErrors.Add(1)
			if strings.Contains(strings.ToLower(err.Error()), "lowernonceintx") || strings.Contains(strings.ToLower(err.Error()), "veryhighnonceintx") {
				networkNonce, _, nErr := a.getAccountSafe(ctx, sw.Address)
				if nErr == nil && networkNonce > localNonce {
					localNonce = networkNonce
				}
			}
			if err := sleepWithContext(ctx, a.cfg.LoopIdleSleep); err != nil {
				return err
			}
			continue
		}

		localNonce += uint64(batchCount)
		a.totalBatches.Add(1)
		a.totalSent.Add(uint64(batchCount))
		a.totalAccepted.Add(uint64(len(hashes)))
		if hashesCh != nil && len(hashes) > 0 {
			hashesCh <- hashes
		}
		cur := a.totalSent.Load()
		if cur <= uint64(batchCount) || cur%10000 < uint64(batchCount) {
			log.Printf("[bulksprint] sent=%d accepted=%d batches=%d", cur, a.totalAccepted.Load(), a.totalBatches.Load())
		}
	}
}

func (a *App) getAccountSafe(ctx context.Context, address string) (uint64, string, error) {
	select {
	case <-ctx.Done():
		return 0, "", ctx.Err()
	case a.readSemaphore <- struct{}{}:
	}
	defer func() { <-a.readSemaphore }()
	return a.gw.GetAccount(ctx, address)
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

func firstNonEmpty(items ...string) string {
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			return item
		}
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func txTotalCost(value string, gasLimit uint64, gasPrice uint64) (*big.Int, bool) {
	valueInt, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return nil, false
	}
	gasCost := new(big.Int).Mul(new(big.Int).SetUint64(gasLimit), new(big.Int).SetUint64(gasPrice))
	return new(big.Int).Add(valueInt, gasCost), true
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (a *App) reserveQuota(limit int) (int, bool) {
	if limit <= 0 {
		return 0, false
	}
	if a.cfg.TargetTx <= 0 {
		return limit, true
	}
	for {
		current := a.reservedTarget.Load()
		if current >= uint64(a.cfg.TargetTx) {
			return 0, false
		}
		remaining := a.cfg.TargetTx - int(current)
		n := min(limit, remaining)
		if n <= 0 {
			return 0, false
		}
		if a.reservedTarget.CompareAndSwap(current, current+uint64(n)) {
			return n, true
		}
	}
}

func (a *App) releaseQuota(n int) {
	if n <= 0 || a.cfg.TargetTx <= 0 {
		return
	}
	a.reservedTarget.Add(^uint64(n - 1))
}

func flattenHashes(ch <-chan []string) []string {
	out := make([]string, 0)
	for hashes := range ch {
		out = append(out, hashes...)
	}
	return out
}

func (a *App) waitForFinalStatuses(ctx context.Context, hashes []string) (int, int, error) {
	type void struct{}
	pending := make(map[string]void, len(hashes))
	for _, h := range hashes {
		pending[h] = void{}
	}

	successes := 0
	failed := 0
	start := time.Now()

	for len(pending) > 0 {
		select {
		case <-ctx.Done():
			return successes, failed, ctx.Err()
		default:
		}
		if a.cfg.ConfirmTimeout > 0 && time.Since(start) >= a.cfg.ConfirmTimeout {
			return successes, failed, fmt.Errorf("bulk confirmation timeout after %s: success=%d failed=%d pending=%d", a.cfg.ConfirmTimeout, successes, failed, len(pending))
		}

		resolvedMu := sync.Mutex{}
		resolved := make([]string, 0, len(pending))
		deltaSuccess := atomic.Int64{}
		deltaFailed := atomic.Int64{}
		checked := atomic.Int64{}
		roundTotal := len(pending)

		jobs := make(chan string)
		var wg sync.WaitGroup
		progressDone := make(chan struct{})
		go func(total int) {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-progressDone:
					return
				case <-ctx.Done():
					return
				case <-ticker.C:
					currentChecked := checked.Load()
					currentResolved := deltaSuccess.Load() + deltaFailed.Load()
					log.Printf(
						"[bulksprint confirm progress] checked=%d/%d resolved=%d successDelta=%d failedDelta=%d",
						currentChecked,
						total,
						currentResolved,
						deltaSuccess.Load(),
						deltaFailed.Load(),
					)
				}
			}
		}(roundTotal)
		for i := 0; i < a.cfg.ConfirmWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for h := range jobs {
					status, err := a.gw.GetTxStatus(ctx, h)
					checked.Add(1)
					if err != nil {
						continue
					}
					if gateway.IsFinalSuccessStatus(status) {
						deltaSuccess.Add(1)
						resolvedMu.Lock()
						resolved = append(resolved, h)
						resolvedMu.Unlock()
						continue
					}
					if gateway.IsFinalFailureStatus(status) {
						deltaFailed.Add(1)
						resolvedMu.Lock()
						resolved = append(resolved, h)
						resolvedMu.Unlock()
					}
				}
			}()
		}

		for h := range pending {
			jobs <- h
		}
		close(jobs)
		wg.Wait()
		close(progressDone)

		for _, h := range resolved {
			delete(pending, h)
		}
		successes += int(deltaSuccess.Load())
		failed += int(deltaFailed.Load())

		log.Printf("[bulksprint confirm] pending=%d success=%d failed=%d", len(pending), successes, failed)
		if len(pending) > 0 {
			time.Sleep(a.cfg.PollInterval)
		}
	}

	return successes, failed, nil
}
