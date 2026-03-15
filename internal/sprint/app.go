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

	stateMu         sync.Mutex
	inflight        int
	nextEligibleAt  time.Time
	transientErrors int
}

type receiverTarget struct {
	Address string
	Shard   int
}

type App struct {
	cfg       Config
	opt       Options
	gw        *gateway.Client
	senders   []senderWallet
	receivers []receiverTarget
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
	receivers := make([]receiverTarget, 0)

	if len(cfg.ReceiverAddresses) > 0 {
		for _, address := range cfg.ReceiverAddresses {
			shard, ok := findWalletShard(mf, address)
			if !ok {
				return nil, fmt.Errorf("receiver %s not found in manifest", address)
			}
			for i := 0; i < max(1, cfg.ReceiverWeight); i++ {
				receivers = append(receivers, receiverTarget{Address: address, Shard: shard})
			}
		}
		if cfg.IncludeTreasuryReceiver {
			for _, record := range mf.Wallets {
				if !record.Enabled {
					continue
				}
				if strings.EqualFold(record.Status, "treasury") {
					for i := 0; i < max(1, cfg.TreasuryReceiverWeight); i++ {
						receivers = append(receivers, receiverTarget{Address: record.Address, Shard: record.Shard})
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
				receivers = append(receivers, receiverTarget{Address: record.Address, Shard: record.Shard})
			}
			continue
		}
		if len(cfg.ReceiverAddresses) == 0 && cfg.IncludeTreasuryReceiver && strings.EqualFold(record.Status, "treasury") {
			for i := 0; i < max(1, cfg.TreasuryReceiverWeight); i++ {
				receivers = append(receivers, receiverTarget{Address: record.Address, Shard: record.Shard})
			}
		}
	}

	if len(senders) == 0 {
		return nil, fmt.Errorf("no senders matched manifest filters")
	}
	if len(receivers) == 0 {
		return nil, fmt.Errorf("no receivers matched manifest filters")
	}
	if err := validateShardCoverage(senders, receivers); err != nil {
		return nil, err
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
	nonceSem := make(chan struct{}, 32)
	nonceErrCh := make(chan error, len(a.senders))
	var nonceWg sync.WaitGroup
	for i := range a.senders {
		nonceWg.Add(1)
		go func(i int) {
			defer nonceWg.Done()
			nonceSem <- struct{}{}
			defer func() { <-nonceSem }()
			nonce, err := a.gw.GetAccountNonce(ctx, a.senders[i].Address)
			if err != nil {
				nonceErrCh <- fmt.Errorf("get nonce for %s: %w", a.senders[i].Address, err)
				return
			}
			a.senders[i].SetNonce(nonce)
		}(i)
	}
	nonceWg.Wait()
	close(nonceErrCh)
	if err := <-nonceErrCh; err != nil {
		return err
	}

	log.Printf(
		"[sprint] network=%s chainID=%s gateway=%s senders=%d receivers=%d targetTx=%d duration=%s tps=%d workers=%d confirmWorkers=%d value=%s gasLimit=%d gasPrice=%d waitConfirm=%t continueOnTransient=%t confirmTarget=%d confirmTimeout=%s shardFilter=%d maxInflightPerWallet=%d successCooldown=%s transientCooldown=%s quarantineCooldown=%s quarantineThreshold=%d",
		a.cfg.Network,
		a.cfg.ChainID,
		a.cfg.GatewayURL,
		len(a.senders),
		len(a.receivers),
		a.cfg.TargetTx,
		a.cfg.Duration,
		a.cfg.SustainedTPS,
		a.cfg.Workers,
		a.cfg.ConfirmWorkers,
		a.cfg.Value,
		a.cfg.GasLimit,
		a.cfg.GasPrice,
		a.cfg.WaitConfirm,
		a.cfg.ContinueOnTransientSendError,
		a.cfg.ConfirmSuccessTarget,
		a.cfg.ConfirmTimeout,
		a.cfg.ShardFilter,
		a.cfg.MaxInflightPerWallet,
		a.cfg.SuccessCooldown,
		a.cfg.TransientCooldown,
		a.cfg.QuarantineCooldown,
		a.cfg.QuarantineTransientThreshold,
	)

	if a.opt.DryRun {
		for i := 0; i < min(5, len(a.senders)); i++ {
			receiver, ok := a.pickReceiver(a.senders[i].Shard, i)
			if !ok {
				return fmt.Errorf("no receiver available for sender shard %d", a.senders[i].Shard)
			}
			log.Printf("[dry-run %d] sender=%s shard=%d receiver=%s receiverShard=%d value=%s", i+1, a.senders[i].Address, a.senders[i].Shard, receiver.Address, receiver.Shard, a.cfg.Value)
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
	var transientSendErrors uint64
	var rrSender uint64
	var rrRecv uint64
	hashes := make([]string, 0, a.cfg.TargetTx)
	hashMu := sync.Mutex{}
	start := time.Now()
	sendDeadline := time.Time{}
	if a.cfg.Duration > 0 {
		sendDeadline = start.Add(a.cfg.Duration)
	}

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
				if !sendDeadline.IsZero() && time.Now().After(sendDeadline) {
					return
				}
				select {
				case <-runCtx.Done():
					return
				case <-permits:
				}
				if !sendDeadline.IsZero() && time.Now().After(sendDeadline) {
					return
				}

				n := int(atomic.AddUint64(&issued, 1))
				if n > a.cfg.TargetTx {
					return
				}

				sender, ok := a.acquireSender(int(atomic.AddUint64(&rrSender, 1) - 1))
				if !ok {
					if sleepErr := sleepWithContext(runCtx, 10*time.Millisecond); sleepErr != nil {
						return
					}
					continue
				}
				receiver, ok := a.pickReceiver(sender.Shard, int(atomic.AddUint64(&rrRecv, 1)-1))
				if !ok {
					sender.release()
					pushErr(fmt.Errorf("no receiver available for sender shard %d", sender.Shard))
					return
				}

				var hash string
				for {
					if !sendDeadline.IsZero() && time.Now().After(sendDeadline) {
						return
					}
					var err error
					hash, err = a.sendOne(runCtx, sender, receiver.Address)
					if err == nil {
						sender.markSuccess(a.cfg.SuccessCooldown)
						break
					}
					if a.cfg.ContinueOnTransientSendError && gateway.IsTransientSendError(err) {
						quarantined := sender.markTransient(
							a.cfg.TransientCooldown,
							a.cfg.QuarantineCooldown,
							a.cfg.QuarantineTransientThreshold,
						)
						cur := atomic.AddUint64(&transientSendErrors, 1)
						if cur <= 20 || cur%100 == 0 {
							log.Printf("[sprint transient-send-error] count=%d err=%v", cur, err)
						}
						if quarantined {
							log.Printf("[sprint wallet-quarantine] sender=%s cooldown=%s", sender.Address, a.cfg.QuarantineCooldown)
						}
						if sleepErr := sleepWithContext(runCtx, 100*time.Millisecond); sleepErr != nil {
							return
						}
						continue
					}
					sender.release()
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

	if a.cfg.Duration == 0 && len(hashes) != a.cfg.TargetTx {
		return fmt.Errorf("sent %d/%d txs", len(hashes), a.cfg.TargetTx)
	}
	log.Printf("[sprint] send phase done issued=%d sent=%d transientSendErrors=%d elapsed=%s", issued, len(hashes), transientSendErrors, time.Since(start))
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

func (a *App) sendOne(ctx context.Context, sender *senderWallet, receiver string) (string, error) {
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
	sender.SetNonceAtLeast(networkNonce)
	return "", fmt.Errorf("send tx nonce=%d hit lowerNonceInTx; refreshed network nonce=%d: %w", nonce, networkNonce, err)
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

func findWalletShard(mf *manifest.Manifest, address string) (int, bool) {
	for _, record := range mf.Wallets {
		if strings.EqualFold(record.Address, address) {
			return record.Shard, true
		}
	}
	return 0, false
}

func validateShardCoverage(senders []senderWallet, receivers []receiverTarget) error {
	receiverShards := make(map[int]struct{}, len(receivers))
	for _, receiver := range receivers {
		receiverShards[receiver.Shard] = struct{}{}
	}
	senderShards := make(map[int]struct{}, len(senders))
	for _, sender := range senders {
		senderShards[sender.Shard] = struct{}{}
	}
	for shard := range senderShards {
		if _, ok := receiverShards[shard]; !ok {
			return fmt.Errorf("no receiver available for sender shard %d", shard)
		}
	}
	return nil
}

func (a *App) pickReceiver(senderShard int, start int) (receiverTarget, bool) {
	for i := 0; i < len(a.receivers); i++ {
		idx := (start + i) % len(a.receivers)
		receiver := a.receivers[idx]
		if receiver.Shard == senderShard {
			return receiver, true
		}
	}
	return receiverTarget{}, false
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

func (a *App) acquireSender(start int) (*senderWallet, bool) {
	now := time.Now()
	for i := 0; i < len(a.senders); i++ {
		idx := (start + i) % len(a.senders)
		if a.senders[idx].tryAcquire(now, a.cfg.MaxInflightPerWallet) {
			return &a.senders[idx], true
		}
	}
	return nil, false
}

func (w *senderWallet) tryAcquire(now time.Time, maxInflight int) bool {
	w.stateMu.Lock()
	defer w.stateMu.Unlock()
	if w.inflight >= maxInflight {
		return false
	}
	if now.Before(w.nextEligibleAt) {
		return false
	}
	w.inflight++
	return true
}

func (w *senderWallet) markSuccess(cooldown time.Duration) {
	w.stateMu.Lock()
	defer w.stateMu.Unlock()
	if w.inflight > 0 {
		w.inflight--
	}
	w.transientErrors = 0
	w.nextEligibleAt = time.Now().Add(cooldown)
}

func (w *senderWallet) markTransient(cooldown time.Duration, quarantine time.Duration, threshold int) bool {
	w.stateMu.Lock()
	defer w.stateMu.Unlock()
	if w.inflight > 0 {
		w.inflight--
	}
	w.transientErrors++
	nextCooldown := cooldown
	quarantined := false
	if w.transientErrors >= threshold {
		nextCooldown = quarantine
		w.transientErrors = 0
		quarantined = true
	}
	w.nextEligibleAt = time.Now().Add(nextCooldown)
	return quarantined
}

func (w *senderWallet) release() {
	w.stateMu.Lock()
	defer w.stateMu.Unlock()
	if w.inflight > 0 {
		w.inflight--
	}
}
