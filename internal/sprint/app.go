package sprint

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"kwakbon/internal/gateway"
	"kwakbon/internal/manifest"
	"kwakbon/internal/ratelimit"
	"kwakbon/internal/txsign"
	"kwakbon/internal/wallets"
)

type Options struct {
	DryRun bool
}

type senderWallet struct {
	*wallets.Wallet
	Shard int
}

type App struct {
	cfg       Config
	opt       Options
	gw        *gateway.Client
	senders   []senderWallet
	receivers []string
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

	senders := make([]senderWallet, 0)
	receivers := make([]string, 0)

	if len(cfg.ReceiverAddresses) > 0 {
		for _, address := range cfg.ReceiverAddresses {
			for i := 0; i < max(1, cfg.ReceiverWeight); i++ {
				receivers = append(receivers, address)
			}
		}
		if cfg.IncludeTreasuryReceiver {
			for _, record := range mf.Wallets {
				if !record.Enabled {
					continue
				}
				if strings.EqualFold(record.Status, "treasury") {
					for i := 0; i < max(1, cfg.TreasuryReceiverWeight); i++ {
						receivers = append(receivers, record.Address)
					}
					break
				}
			}
		}
	}

	for _, record := range mf.Wallets {
		if !record.Enabled {
			continue
		}
		if cfg.ShardFilter >= 0 && record.Shard != cfg.ShardFilter {
			continue
		}

		if containsFold(cfg.SenderStatuses, record.Status) && containsAllTags(record.Tags, cfg.RequiredSenderTags) {
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
			senders = append(senders, senderWallet{Wallet: w, Shard: record.Shard})
			continue
		}

		if len(cfg.ReceiverAddresses) == 0 && containsFold(cfg.ReceiverStatuses, record.Status) && containsAllTags(record.Tags, cfg.RequiredReceiverTags) {
			for i := 0; i < max(1, cfg.ReceiverWeight); i++ {
				receivers = append(receivers, record.Address)
			}
			continue
		}
		if len(cfg.ReceiverAddresses) == 0 && cfg.IncludeTreasuryReceiver && strings.EqualFold(record.Status, "treasury") {
			for i := 0; i < max(1, cfg.TreasuryReceiverWeight); i++ {
				receivers = append(receivers, record.Address)
			}
		}
	}

	if len(senders) == 0 {
		return nil, fmt.Errorf("no senders matched manifest filters")
	}
	if len(receivers) == 0 {
		return nil, fmt.Errorf("no receivers matched manifest filters")
	}

	return &App{
		cfg:       cfg,
		opt:       opt,
		gw:        gw,
		senders:   senders,
		receivers: receivers,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	for i := range a.senders {
		nonce, err := a.gw.GetAccountNonce(ctx, a.senders[i].Address)
		if err != nil {
			return fmt.Errorf("get nonce for %s: %w", a.senders[i].Address, err)
		}
		a.senders[i].SetNonce(nonce)
	}

	log.Printf(
		"[sprint] network=%s chainID=%s gateway=%s senders=%d receivers=%d targetTx=%d tps=%d workers=%d confirmWorkers=%d value=%s gasLimit=%d gasPrice=%d waitConfirm=%t confirmTarget=%d confirmTimeout=%s shardFilter=%d",
		a.cfg.Network,
		a.cfg.ChainID,
		a.cfg.GatewayURL,
		len(a.senders),
		len(a.receivers),
		a.cfg.TargetTx,
		a.cfg.SustainedTPS,
		a.cfg.Workers,
		a.cfg.ConfirmWorkers,
		a.cfg.Value,
		a.cfg.GasLimit,
		a.cfg.GasPrice,
		a.cfg.WaitConfirm,
		a.cfg.ConfirmSuccessTarget,
		a.cfg.ConfirmTimeout,
		a.cfg.ShardFilter,
	)

	if a.opt.DryRun {
		for i := 0; i < min(5, len(a.senders)); i++ {
			log.Printf("[dry-run %d] sender=%s shard=%d receiver=%s value=%s", i+1, a.senders[i].Address, a.senders[i].Shard, a.receivers[i%len(a.receivers)], a.cfg.Value)
		}
		return nil
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	permits := make(chan struct{}, 4096)
	ctrl := ratelimit.NewTPSController(int64(a.cfg.SustainedTPS))
	go ctrl.Run(runCtx, permits)

	var issued uint64
	var sent uint64
	var rrSender uint64
	var rrRecv uint64
	hashes := make([]string, 0, a.cfg.TargetTx)
	hashMu := sync.Mutex{}

	errCh := make(chan error, 1)
	pushErr := func(err error) {
		select {
		case errCh <- err:
		default:
		}
		cancel()
	}

	var wg sync.WaitGroup
	for w := 0; w < a.cfg.Workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-runCtx.Done():
					return
				case <-permits:
				}

				n := int(atomic.AddUint64(&issued, 1))
				if n > a.cfg.TargetTx {
					return
				}

				si := int(atomic.AddUint64(&rrSender, 1)-1) % len(a.senders)
				sender := a.senders[si]
				ri := int(atomic.AddUint64(&rrRecv, 1)-1) % len(a.receivers)
				receiver := a.receivers[ri]

				hash, err := a.sendOne(runCtx, sender, receiver)
				if err != nil {
					pushErr(err)
					return
				}

				hashMu.Lock()
				hashes = append(hashes, hash)
				hashMu.Unlock()

				curSent := atomic.AddUint64(&sent, 1)
				if curSent%10000 == 0 || curSent == uint64(a.cfg.TargetTx) {
					log.Printf("[sprint] sent=%d/%d", curSent, a.cfg.TargetTx)
				}
			}
		}()
	}

	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
	}

	if len(hashes) != a.cfg.TargetTx {
		return fmt.Errorf("sent %d/%d txs", len(hashes), a.cfg.TargetTx)
	}
	if !a.cfg.WaitConfirm {
		return nil
	}

	successes, failed, err := a.waitForFinalStatuses(runCtx, hashes)
	if err != nil {
		return err
	}
	log.Printf("[sprint] confirmations done success=%d failed=%d total=%d", successes, failed, len(hashes))
	if successes < a.cfg.ConfirmSuccessTarget {
		return fmt.Errorf("only %d/%d required transactions succeeded on-chain", successes, a.cfg.ConfirmSuccessTarget)
	}
	return nil
}

func (a *App) sendOne(ctx context.Context, sender senderWallet, receiver string) (string, error) {
	nonce := sender.ReserveNonce()
	tx := gateway.TxSendRequest{
		Nonce:    nonce,
		Value:    a.cfg.Value,
		Receiver: receiver,
		Sender:   sender.Address,
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
		return "", fmt.Errorf("serialize tx nonce=%d: %w", nonce, err)
	}
	sig, err := sender.Signer.SignTxBytes(toSign)
	if err != nil {
		return "", fmt.Errorf("sign tx nonce=%d: %w", nonce, err)
	}
	tx.Signature = hex.EncodeToString(sig)

	hash, err := a.gw.SendTx(ctx, tx)
	if err == nil {
		return hash, nil
	}
	if !strings.Contains(err.Error(), "lowerNonceInTx") {
		return "", fmt.Errorf("send tx nonce=%d: %w", nonce, err)
	}

	networkNonce, nErr := a.gw.GetAccountNonce(ctx, sender.Address)
	if nErr != nil {
		return "", fmt.Errorf("send tx nonce=%d failed with lowerNonceInTx and nonce refresh failed: %w", nonce, nErr)
	}
	sender.SetNonce(networkNonce)
	retryNonce := sender.ReserveNonce()
	tx.Nonce = retryNonce
	unsigned.Nonce = int64(retryNonce)

	toSign, err = txsign.SerializeForSigning(unsigned)
	if err != nil {
		return "", fmt.Errorf("serialize retried tx nonce=%d: %w", retryNonce, err)
	}
	sig, err = sender.Signer.SignTxBytes(toSign)
	if err != nil {
		return "", fmt.Errorf("sign retried tx nonce=%d: %w", retryNonce, err)
	}
	tx.Signature = hex.EncodeToString(sig)

	hash, err = a.gw.SendTx(ctx, tx)
	if err != nil {
		return "", fmt.Errorf("send retried tx nonce=%d: %w", retryNonce, err)
	}
	return hash, nil
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
		if successes >= a.cfg.ConfirmSuccessTarget {
			return successes, failed, nil
		}
		select {
		case <-ctx.Done():
			return successes, failed, ctx.Err()
		default:
		}
		if a.cfg.ConfirmTimeout > 0 && time.Since(start) >= a.cfg.ConfirmTimeout {
			if successes >= a.cfg.ConfirmSuccessTarget {
				return successes, failed, nil
			}
			return successes, failed, fmt.Errorf("confirmation timeout after %s: success=%d failed=%d pending=%d required=%d", a.cfg.ConfirmTimeout, successes, failed, len(pending), a.cfg.ConfirmSuccessTarget)
		}

		resolvedMu := sync.Mutex{}
		resolved := make([]string, 0, len(pending))
		deltaSuccess := atomic.Int64{}
		deltaFailed := atomic.Int64{}

		jobs := make(chan string)
		var wg sync.WaitGroup
		for i := 0; i < a.cfg.ConfirmWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for h := range jobs {
					status, err := a.gw.GetTxStatus(ctx, h)
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

		for _, h := range resolved {
			delete(pending, h)
		}
		successes += int(deltaSuccess.Load())
		failed += int(deltaFailed.Load())

		log.Printf("[sprint confirm] pending=%d success=%d failed=%d", len(pending), successes, failed)
		if len(pending) > 0 {
			time.Sleep(a.cfg.PollInterval)
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
