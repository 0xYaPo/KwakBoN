package sprint

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Network    string
	GatewayURL string
	ChainID    string
	TxVersion  uint32

	ManifestPath string

	SenderStatuses []string
	ReceiverStatuses []string
	RequiredSenderTags []string
	RequiredReceiverTags []string
	ShardFilter int

	IncludeTreasuryReceiver bool
	ReceiverWeight          int
	TreasuryReceiverWeight  int

	TargetTx      int
	SustainedTPS  int
	Workers       int
	ConfirmWorkers int

	Value    string
	GasLimit uint64
	GasPrice uint64

	WaitConfirm          bool
	ConfirmSuccessTarget int
	ConfirmTimeout       time.Duration
	PollInterval         time.Duration
	HTTPTimeout          time.Duration
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func mustInt(key string, def int) (int, error) {
	v := getenv(key, "")
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s must be int: %w", key, err)
	}
	return n, nil
}

func mustUint64(key string, def uint64) (uint64, error) {
	v := getenv(key, "")
	if v == "" {
		return def, nil
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be uint64: %w", key, err)
	}
	return n, nil
}

func mustBool(key string, def bool) (bool, error) {
	v := getenv(key, "")
	if v == "" {
		return def, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("%s must be bool: %w", key, err)
	}
	return b, nil
}

func mustDurationSeconds(key string, defSeconds int) (time.Duration, error) {
	n, err := mustInt(key, defSeconds)
	if err != nil {
		return 0, err
	}
	return time.Duration(n) * time.Second, nil
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

func LoadConfigFromEnv() (Config, error) {
	txVersion, err := mustInt("TX_VERSION", 2)
	if err != nil {
		return Config{}, err
	}
	targetTx, err := mustInt("SPRINT_TARGET_TX", 1100000)
	if err != nil {
		return Config{}, err
	}
	if targetTx <= 0 {
		return Config{}, fmt.Errorf("SPRINT_TARGET_TX must be > 0")
	}
	tps, err := mustInt("SUSTAINED_TPS", 700)
	if err != nil {
		return Config{}, err
	}
	if tps <= 0 {
		return Config{}, fmt.Errorf("SUSTAINED_TPS must be > 0")
	}
	workers, err := mustInt("SPRINT_WORKERS", 64)
	if err != nil {
		return Config{}, err
	}
	if workers <= 0 {
		return Config{}, fmt.Errorf("SPRINT_WORKERS must be > 0")
	}
	confirmWorkers, err := mustInt("SPRINT_CONFIRM_WORKERS", 32)
	if err != nil {
		return Config{}, err
	}
	if confirmWorkers <= 0 {
		return Config{}, fmt.Errorf("SPRINT_CONFIRM_WORKERS must be > 0")
	}
	gasLimit, err := mustUint64("GAS_LIMIT", 50000)
	if err != nil {
		return Config{}, err
	}
	gasPrice, err := mustUint64("GAS_PRICE", 1000000000)
	if err != nil {
		return Config{}, err
	}
	waitConfirm, err := mustBool("WAIT_CONFIRM", true)
	if err != nil {
		return Config{}, err
	}
	confirmTarget, err := mustInt("CONFIRM_SUCCESS_TARGET", targetTx)
	if err != nil {
		return Config{}, err
	}
	if confirmTarget <= 0 || confirmTarget > targetTx {
		return Config{}, fmt.Errorf("CONFIRM_SUCCESS_TARGET must be between 1 and SPRINT_TARGET_TX")
	}
	confirmTimeout, err := mustDurationSeconds("CONFIRM_TIMEOUT_SECONDS", 0)
	if err != nil {
		return Config{}, err
	}
	poll, err := mustDurationSeconds("POLL_INTERVAL_SECONDS", 3)
	if err != nil {
		return Config{}, err
	}
	httpTO, err := mustDurationSeconds("HTTP_TIMEOUT_SECONDS", 12)
	if err != nil {
		return Config{}, err
	}
	shardFilter, err := mustInt("SPRINT_SHARD_FILTER", 2)
	if err != nil {
		return Config{}, err
	}
	includeTreasury, err := mustBool("INCLUDE_TREASURY_RECEIVER", true)
	if err != nil {
		return Config{}, err
	}
	receiverWeight, err := mustInt("RECEIVER_WEIGHT", 4)
	if err != nil {
		return Config{}, err
	}
	treasuryWeight, err := mustInt("TREASURY_RECEIVER_WEIGHT", 1)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Network:                getenv("DEFAULT_NETWORK", "battle"),
		GatewayURL:             getenv("GATEWAY_URL", "https://api.battleofnodes.com"),
		ChainID:                getenv("CHAIN_ID", "B"),
		TxVersion:              uint32(txVersion),
		ManifestPath:           getenv("WALLETS_MANIFEST", "./configs/wallets-manifest.json"),
		SenderStatuses:         splitCSV(getenv("SPRINT_SENDER_STATUSES", "active_candidate")),
		ReceiverStatuses:       splitCSV(getenv("SPRINT_RECEIVER_STATUSES", "receiver")),
		RequiredSenderTags:     splitCSV(getenv("SPRINT_REQUIRED_SENDER_TAGS", "window_a,sender")),
		RequiredReceiverTags:   splitCSV(getenv("SPRINT_REQUIRED_RECEIVER_TAGS", "window_a,sink")),
		ShardFilter:            shardFilter,
		IncludeTreasuryReceiver: includeTreasury,
		ReceiverWeight:         receiverWeight,
		TreasuryReceiverWeight: treasuryWeight,
		TargetTx:               targetTx,
		SustainedTPS:           tps,
		Workers:                workers,
		ConfirmWorkers:         confirmWorkers,
		Value:                  getenv("SPRINT_TX_VALUE", "1"),
		GasLimit:               gasLimit,
		GasPrice:               gasPrice,
		WaitConfirm:            waitConfirm,
		ConfirmSuccessTarget:   confirmTarget,
		ConfirmTimeout:         confirmTimeout,
		PollInterval:           poll,
		HTTPTimeout:            httpTO,
	}, nil
}
