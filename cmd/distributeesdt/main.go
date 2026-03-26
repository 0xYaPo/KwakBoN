package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"sort"
	"time"

	"kwakbon/internal/esdt"
	"kwakbon/internal/gateway"
	"kwakbon/internal/txsign"
	"kwakbon/internal/wallets"
)

type poolWallet struct {
	WalletID string `json:"walletId"`
	Address  string `json:"address"`
	Shard    int    `json:"shard"`
	PemPath  string `json:"pemPath"`
}

type allocationFile struct {
	Pools map[string][]poolWallet `json:"pools"`
}

func main() {
	var (
		allocationPath         = flag.String("allocation", "./configs/challenge4/wallet-allocation.challenge4-callers.json", "path to challenge4 allocation file")
		poolName               = flag.String("pool", "", "allocation pool name")
		tokenID                = flag.String("token-id", "WEGLD-bd4d79", "token identifier to distribute")
		amountBase             = flag.String("amount-base", "", "base-unit amount to send to each wallet")
		treasuryAddress        = flag.String("treasury-address", "erd1n28dse0sej7m2rz02ceftr30596arzx2a0trcegl9p3ztdagjahqa9szcx", "treasury bech32")
		treasuryPemPath        = flag.String("treasury-pem", "", "treasury pem path")
		gatewayURL             = flag.String("gateway", "https://api.battleofnodes.com", "gateway/api base URL")
		chainID                = flag.String("chain-id", "B", "chain id")
		gasLimit               = flag.Uint64("gas-limit", 500000, "gas limit for each ESDT transfer")
		gasPrice               = flag.Uint64("gas-price", 1000000000, "gas price")
		txVersion              = flag.Uint("tx-version", 2, "tx version")
		batchSize              = flag.Int("batch-size", 25, "send-multiple batch size")
		waitConfirm            = flag.Bool("wait-confirm", true, "wait for final status")
		confirmTimeoutSeconds  = flag.Int("confirm-timeout-seconds", 180, "confirmation timeout seconds")
		pollIntervalSeconds    = flag.Int("poll-interval-seconds", 3, "confirmation poll interval seconds")
		httpTimeoutSeconds     = flag.Int("http-timeout-seconds", 25, "http timeout seconds")
		nonceRefreshEvery      = flag.Int("nonce-refresh-every", 75, "refresh treasury nonce every N sends")
		sendCooldownMillis     = flag.Int("send-cooldown-ms", 0, "sleep between send batches")
		dryRun                 = flag.Bool("dry-run", false, "show distribution plan only")
	)
	flag.Parse()

	if *poolName == "" {
		log.Fatal("missing required -pool")
	}
	if *amountBase == "" {
		log.Fatal("missing required -amount-base")
	}
	if !*dryRun && *treasuryPemPath == "" {
		log.Fatal("missing required -treasury-pem")
	}
	if *batchSize <= 0 {
		log.Fatal("batch-size must be > 0")
	}
	if *pollIntervalSeconds <= 0 {
		log.Fatal("poll-interval-seconds must be > 0")
	}
	if *confirmTimeoutSeconds <= 0 {
		log.Fatal("confirm-timeout-seconds must be > 0")
	}
	if *nonceRefreshEvery <= 0 {
		log.Fatal("nonce-refresh-every must be > 0")
	}

	perWalletAmount, ok := new(big.Int).SetString(*amountBase, 10)
	if !ok || perWalletAmount.Sign() <= 0 {
		log.Fatalf("invalid amount-base: %s", *amountBase)
	}

	fileBytes, err := os.ReadFile(*allocationPath)
	if err != nil {
		log.Fatalf("read allocation file: %v", err)
	}
	var allocation allocationFile
	if err := json.Unmarshal(fileBytes, &allocation); err != nil {
		log.Fatalf("parse allocation file: %v", err)
	}
	targets := allocation.Pools[*poolName]
	if len(targets) == 0 {
		log.Fatalf("allocation pool not found or empty: %s", *poolName)
	}

	sort.Slice(targets, func(i, j int) bool {
		return targets[i].WalletID < targets[j].WalletID
	})

	total := new(big.Int).Mul(new(big.Int).Set(perWalletAmount), big.NewInt(int64(len(targets))))
	log.Printf("[distributeesdt] pool=%s wallets=%d token=%s amountPerWalletBase=%s totalBase=%s", *poolName, len(targets), *tokenID, perWalletAmount.String(), total.String())
	for i := 0; i < min(10, len(targets)); i++ {
		w := targets[i]
		log.Printf("[dry-run] walletId=%s shard=%d address=%s amountBase=%s", w.WalletID, w.Shard, w.Address, perWalletAmount.String())
	}
	if *dryRun {
		return
	}

	ctx := context.Background()
	gw := gateway.New(*gatewayURL, time.Duration(*httpTimeoutSeconds)*time.Second)
	signer, err := wallets.NewPemSigner(*treasuryPemPath, *treasuryAddress)
	if err != nil {
		log.Fatalf("load treasury signer: %v", err)
	}

	nonce, err := gw.GetAccountNonce(ctx, *treasuryAddress)
	if err != nil {
		log.Fatalf("get treasury nonce: %v", err)
	}

	hashes := make([]string, 0, len(targets))
	for start := 0; start < len(targets); start += *batchSize {
		if start > 0 && start%*nonceRefreshEvery == 0 {
			networkNonce, nErr := gw.GetAccountNonce(ctx, *treasuryAddress)
			if nErr != nil {
				log.Fatalf("refresh treasury nonce after %d sends: %v", start, nErr)
			}
			if networkNonce > nonce {
				log.Printf("[distributeesdt] treasury nonce refresh old=%d new=%d after=%d", nonce, networkNonce, start)
				nonce = networkNonce
			}
		}

		end := min(start+*batchSize, len(targets))
		batch := make([]gateway.TxSendRequest, 0, end-start)
		for _, target := range targets[start:end] {
			data := esdt.ESDTTransferData(*tokenID, perWalletAmount)
			gatewayData := base64.StdEncoding.EncodeToString([]byte(data))
			tx := gateway.TxSendRequest{
				Nonce:    nonce,
				Value:    "0",
				Receiver: target.Address,
				Sender:   *treasuryAddress,
				GasPrice: *gasPrice,
				GasLimit: *gasLimit,
				Data:     gatewayData,
				ChainID:  *chainID,
				Version:  uint32(*txVersion),
			}
			unsigned := txsign.UnsignedTx{
				Nonce:    int64(tx.Nonce),
				Value:    tx.Value,
				Receiver: tx.Receiver,
				Sender:   tx.Sender,
				GasPrice: tx.GasPrice,
				GasLimit: tx.GasLimit,
				Data:     data,
				ChainID:  tx.ChainID,
				Version:  tx.Version,
			}
			toSign, sErr := txsign.SerializeForSigning(unsigned)
			if sErr != nil {
				log.Fatalf("serialize tx for walletId=%s nonce=%d: %v", target.WalletID, nonce, sErr)
			}
			sig, sErr := signer.SignTxBytes(toSign)
			if sErr != nil {
				log.Fatalf("sign tx for walletId=%s nonce=%d: %v", target.WalletID, nonce, sErr)
			}
			tx.Signature = hex.EncodeToString(sig)
			batch = append(batch, tx)
			nonce++
		}

		batchHashes, sendErr := gw.SendTxs(ctx, batch)
		if sendErr != nil {
			log.Fatalf("send batch %d-%d: %v", start, end-1, sendErr)
		}
		hashes = append(hashes, batchHashes...)
		log.Printf("[distributeesdt] sent=%d/%d", end, len(targets))

		if *sendCooldownMillis > 0 {
			time.Sleep(time.Duration(*sendCooldownMillis) * time.Millisecond)
		}
	}

	if !*waitConfirm {
		return
	}

	successes, failed, err := waitAll(ctx, gw, hashes, time.Duration(*confirmTimeoutSeconds)*time.Second, time.Duration(*pollIntervalSeconds)*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[distributeesdt] confirmations done success=%d failed=%d total=%d", successes, failed, len(hashes))
	if successes != len(hashes) {
		log.Fatalf("only %d/%d ESDT transfers succeeded", successes, len(hashes))
	}
}

func waitAll(ctx context.Context, gw *gateway.Client, hashes []string, timeout time.Duration, poll time.Duration) (int, int, error) {
	pending := make(map[string]struct{}, len(hashes))
	for _, h := range hashes {
		pending[h] = struct{}{}
	}
	successes := 0
	failed := 0
	start := time.Now()

	for len(pending) > 0 {
		if time.Since(start) >= timeout {
			return successes, failed, fmt.Errorf("esdt distribution confirmation timeout after %s: pending=%d success=%d failed=%d", timeout, len(pending), successes, failed)
		}
		for h := range pending {
			status, err := gw.GetTxStatus(ctx, h)
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
		log.Printf("[distributeesdt confirm] pending=%d success=%d failed=%d", len(pending), successes, failed)
		if len(pending) == 0 {
			break
		}
		select {
		case <-ctx.Done():
			return successes, failed, ctx.Err()
		case <-time.After(poll):
		}
	}

	return successes, failed, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
