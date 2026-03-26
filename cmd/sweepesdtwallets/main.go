package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math/big"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"kwakbon/internal/esdt"
	"kwakbon/internal/gateway"
	"kwakbon/internal/manifest"
	"kwakbon/internal/txsign"
	"kwakbon/internal/wallets"
)

type config struct {
	GatewayURL      string
	ChainID         string
	TxVersion       uint32
	GasPrice        uint64
	GasLimit        uint64
	ManifestPath    string
	TreasuryAddress string
	TokenID         string
	WaitConfirm     bool
	ConfirmWorkers  int
	PollInterval    time.Duration
	ConfirmTimeout  time.Duration
	HTTPTimeout     time.Duration
	IncludeStatuses []string
	ExcludeStatuses []string
	RequiredTags    []string
	ShardFilter     int
}

type target struct {
	WalletID    string
	Address     string
	Shard       int
	Status      string
	TokenBase   *big.Int
	PemPath     string
}

func main() {
	var (
		cfg    config
		dryRun bool
	)

	flag.StringVar(&cfg.GatewayURL, "gateway", "https://api.battleofnodes.com", "gateway/api base URL")
	flag.StringVar(&cfg.ChainID, "chain-id", "B", "chain id")
	txVersion := flag.Uint("tx-version", 2, "tx version")
	flag.Uint64Var(&cfg.GasPrice, "gas-price", 1000000000, "gas price")
	flag.Uint64Var(&cfg.GasLimit, "gas-limit", 500000, "gas limit for ESDT sweep")
	flag.StringVar(&cfg.ManifestPath, "manifest", "./configs/challenge4/wallets-manifest.challenge4-callers.json", "wallet manifest path")
	flag.StringVar(&cfg.TreasuryAddress, "treasury-address", "erd1n28dse0sej7m2rz02ceftr30596arzx2a0trcegl9p3ztdagjahqa9szcx", "treasury bech32")
	flag.StringVar(&cfg.TokenID, "token-id", "WEGLD-bd4d79", "token identifier to sweep")
	flag.BoolVar(&cfg.WaitConfirm, "wait-confirm", true, "wait for final status")
	flag.IntVar(&cfg.ConfirmWorkers, "confirm-workers", 32, "number of confirmation workers")
	pollSeconds := flag.Int("poll-interval-seconds", 3, "confirmation poll interval seconds")
	confirmTimeoutSeconds := flag.Int("confirm-timeout-seconds", 180, "confirmation timeout seconds")
	httpTimeoutSeconds := flag.Int("http-timeout-seconds", 25, "http timeout seconds")
	includeStatuses := flag.String("include-statuses", "", "comma-separated allowed statuses")
	excludeStatuses := flag.String("exclude-statuses", "treasury,receiver", "comma-separated excluded statuses")
	requiredTags := flag.String("required-tags", "", "comma-separated required tags")
	flag.IntVar(&cfg.ShardFilter, "shard-filter", -1, "restrict to shard, -1 means all")
	flag.BoolVar(&dryRun, "dry-run", false, "preview matched wallets and token balances without sending")
	flag.Parse()

	cfg.TxVersion = uint32(*txVersion)
	cfg.PollInterval = time.Duration(*pollSeconds) * time.Second
	cfg.ConfirmTimeout = time.Duration(*confirmTimeoutSeconds) * time.Second
	cfg.HTTPTimeout = time.Duration(*httpTimeoutSeconds) * time.Second
	cfg.IncludeStatuses = splitCSV(*includeStatuses)
	cfg.ExcludeStatuses = splitCSV(*excludeStatuses)
	cfg.RequiredTags = splitCSV(*requiredTags)

	if cfg.TreasuryAddress == "" {
		log.Fatal("missing -treasury-address")
	}
	if cfg.TokenID == "" {
		log.Fatal("missing -token-id")
	}
	if cfg.ConfirmWorkers <= 0 {
		log.Fatal("confirm-workers must be > 0")
	}

	app, err := newApp(cfg, dryRun)
	if err != nil {
		log.Fatalf("init error: %v", err)
	}

	start := time.Now()
	if err := app.run(context.Background()); err != nil {
		log.Fatalf("run error: %v", err)
	}
	fmt.Printf("[sweepesdtwallets] done in %s\n", time.Since(start))
}

type app struct {
	cfg     config
	dryRun  bool
	gw      *gateway.Client
	records []manifest.WalletRecord
}

func newApp(cfg config, dryRun bool) (*app, error) {
	gw := gateway.New(cfg.GatewayURL, cfg.HTTPTimeout)
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

	return &app{cfg: cfg, dryRun: dryRun, gw: gw, records: records}, nil
}

func (a *app) run(ctx context.Context) error {
	if len(a.records) == 0 {
		return fmt.Errorf("no manifest wallets matched the sweep filters")
	}

	targets, total, err := a.computeTargets(ctx)
	if err != nil {
		return err
	}

	log.Printf(
		"[sweepesdtwallets] token=%s treasury=%s matched=%d sweepable=%d shardFilter=%d include=%s tags=%s totalBase=%s",
		a.cfg.TokenID,
		a.cfg.TreasuryAddress,
		len(a.records),
		len(targets),
		a.cfg.ShardFilter,
		strings.Join(a.cfg.IncludeStatuses, ","),
		strings.Join(a.cfg.RequiredTags, ","),
		total.String(),
	)

	if a.dryRun {
		for i := 0; i < min(10, len(targets)); i++ {
			t := targets[i]
			log.Printf("[dry-run] wallet=%s status=%s shard=%d token=%s amount=%s", t.Address, t.Status, t.Shard, a.cfg.TokenID, t.TokenBase.String())
		}
		return nil
	}

	hashes := make([]string, 0, len(targets))
	for i, target := range targets {
		hash, err := a.sendSweep(ctx, target)
		if err != nil {
			return fmt.Errorf("sweep token %s from %s: %w", a.cfg.TokenID, target.Address, err)
		}
		hashes = append(hashes, hash)
		if (i+1)%25 == 0 || i+1 == len(targets) {
			log.Printf("[sweepesdtwallets] sent=%d/%d", i+1, len(targets))
		}
	}

	if !a.cfg.WaitConfirm {
		return nil
	}

	successes, failed, err := a.waitAll(ctx, hashes)
	if err != nil {
		return err
	}
	log.Printf("[sweepesdtwallets] confirmations done success=%d failed=%d total=%d", successes, failed, len(hashes))
	if successes != len(hashes) {
		return fmt.Errorf("only %d/%d ESDT sweep txs succeeded", successes, len(hashes))
	}
	return nil
}

func (a *app) computeTargets(ctx context.Context) ([]target, *big.Int, error) {
	out := make([]target, 0, len(a.records))
	total := big.NewInt(0)
	start := time.Now()
	scanned := 0
	sweepable := 0

	for _, record := range a.records {
		scanned++
		balanceStr, err := a.getTokenBalanceWithRetry(ctx, record.Address)
		if err != nil {
			return nil, nil, fmt.Errorf("get token balance for %s: %w", record.Address, err)
		}
		balance, ok := new(big.Int).SetString(balanceStr, 10)
		if !ok {
			return nil, nil, fmt.Errorf("invalid token balance for %s: %s", record.Address, balanceStr)
		}
		if balance.Sign() <= 0 {
			if scanned%25 == 0 || scanned == len(a.records) {
				log.Printf("[sweepesdtwallets scan] token=%s scanned=%d/%d sweepable=%d totalBase=%s elapsed=%s", a.cfg.TokenID, scanned, len(a.records), sweepable, total.String(), time.Since(start).Round(time.Second))
			}
			continue
		}

		total.Add(total, balance)
		sweepable++
		out = append(out, target{
			WalletID:  record.WalletID,
			Address:   record.Address,
			Shard:     record.Shard,
			Status:    record.Status,
			TokenBase: balance,
			PemPath:   record.PemPath,
		})

		if scanned%25 == 0 || scanned == len(a.records) {
			log.Printf("[sweepesdtwallets scan] token=%s scanned=%d/%d sweepable=%d totalBase=%s elapsed=%s", a.cfg.TokenID, scanned, len(a.records), sweepable, total.String(), time.Since(start).Round(time.Second))
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Status == out[j].Status {
			return out[i].WalletID < out[j].WalletID
		}
		return out[i].Status < out[j].Status
	})

	return out, total, nil
}

func (a *app) getTokenBalanceWithRetry(ctx context.Context, address string) (string, error) {
	var lastErr error
	for attempt := 1; attempt <= 4; attempt++ {
		balanceStr, err := a.gw.GetAccountTokenBalance(ctx, address, a.cfg.TokenID)
		if err == nil {
			return balanceStr, nil
		}
		lastErr = err
		if attempt == 4 || !isTransientBalanceError(err) {
			break
		}
		log.Printf("[sweepesdtwallets balance-retry] wallet=%s token=%s attempt=%d err=%v", address, a.cfg.TokenID, attempt, err)
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
		}
	}
	return "", lastErr
}

func isTransientBalanceError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "temporarily unavailable") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "eof") ||
		strings.Contains(msg, "bad request") ||
		strings.Contains(msg, "too many requests") ||
		strings.Contains(msg, "bad gateway") ||
		strings.Contains(msg, "service unavailable") ||
		strings.Contains(msg, "gateway timeout") ||
		strings.Contains(msg, "invalid character '<'")
}

func (a *app) sendSweep(ctx context.Context, target target) (string, error) {
	signer, err := wallets.NewPemSigner(target.PemPath, target.Address)
	if err != nil {
		return "", fmt.Errorf("load signer: %w", err)
	}
	nonce, err := a.gw.GetAccountNonce(ctx, target.Address)
	if err != nil {
		return "", fmt.Errorf("get nonce: %w", err)
	}

	rawData := esdt.ESDTTransferData(a.cfg.TokenID, target.TokenBase)
	gatewayData := base64.StdEncoding.EncodeToString([]byte(rawData))

	tx := gateway.TxSendRequest{
		Nonce:    nonce,
		Value:    "0",
		Receiver: a.cfg.TreasuryAddress,
		Sender:   target.Address,
		GasPrice: a.cfg.GasPrice,
		GasLimit: a.cfg.GasLimit,
		Data:     gatewayData,
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
		Data:     rawData,
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

func (a *app) waitAll(ctx context.Context, hashes []string) (int, int, error) {
	pending := make(map[string]struct{}, len(hashes))
	for _, h := range hashes {
		pending[h] = struct{}{}
	}
	successes := 0
	failed := 0
	start := time.Now()

	for len(pending) > 0 {
		if a.cfg.ConfirmTimeout > 0 && time.Since(start) >= a.cfg.ConfirmTimeout {
			return successes, failed, fmt.Errorf(
				"esdt sweep confirmation timeout after %s: pending=%d success=%d failed=%d samplePending=%s",
				a.cfg.ConfirmTimeout,
				len(pending),
				successes,
				failed,
				strings.Join(samplePendingHashes(pending, 5), ","),
			)
		}

		resolvedMu := sync.Mutex{}
		resolvedSuccess := make([]string, 0, len(pending))
		resolvedFailed := make([]string, 0, len(pending))
		checked := atomic.Int64{}
		deltaSuccess := atomic.Int64{}
		deltaFailed := atomic.Int64{}
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
					log.Printf("[sweepesdtwallets confirm progress] token=%s checked=%d/%d resolved=%d successDelta=%d failedDelta=%d", a.cfg.TokenID, checked.Load(), total, deltaSuccess.Load()+deltaFailed.Load(), deltaSuccess.Load(), deltaFailed.Load())
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
						resolvedSuccess = append(resolvedSuccess, h)
						resolvedMu.Unlock()
						continue
					}
					if gateway.IsFinalFailureStatus(status) {
						deltaFailed.Add(1)
						resolvedMu.Lock()
						resolvedFailed = append(resolvedFailed, h)
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

		for _, h := range resolvedSuccess {
			delete(pending, h)
		}
		for _, h := range resolvedFailed {
			delete(pending, h)
		}
		successes += len(resolvedSuccess)
		failed += len(resolvedFailed)

		log.Printf("[sweepesdtwallets confirm] token=%s pending=%d success=%d failed=%d", a.cfg.TokenID, len(pending), successes, failed)
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

func samplePendingHashes(pending map[string]struct{}, limit int) []string {
	if limit <= 0 || len(pending) == 0 {
		return nil
	}
	hashes := make([]string, 0, len(pending))
	for h := range pending {
		hashes = append(hashes, h)
	}
	sort.Strings(hashes)
	if len(hashes) > limit {
		hashes = hashes[:limit]
	}
	return hashes
}

func splitCSV(in string) []string {
	if strings.TrimSpace(in) == "" {
		return nil
	}
	parts := strings.Split(in, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		v := strings.TrimSpace(part)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
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
