package bulksprint

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	GatewayURL   string
	ChainID      string
	TxVersion    uint32
	ManifestPath string

	SenderStatuses   []string
	RequiredTags     []string
	ShardFilter      int
	MaxActiveWallets int
	RoutingMode      string

	Duration           time.Duration
	TargetTx           int
	BatchSize          int
	MaxNonceLookahead  int
	MaxConcurrentReads int
	LoopIdleSleep      time.Duration
	ConfirmWorkers     int
	PollInterval       time.Duration
	ConfirmTimeout     time.Duration
	WaitConfirm        bool
	HTTPTimeout        time.Duration

	Value    string
	GasLimit uint64
	GasPrice uint64
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

func mustDurationSeconds(key string, defSeconds int) (time.Duration, error) {
	n, err := mustInt(key, defSeconds)
	if err != nil {
		return 0, err
	}
	return time.Duration(n) * time.Second, nil
}

func mustDurationMillis(key string, defMillis int) (time.Duration, error) {
	n, err := mustInt(key, defMillis)
	if err != nil {
		return 0, err
	}
	return time.Duration(n) * time.Millisecond, nil
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
	targetTx, err := mustInt("BULK_TARGET_TX", 1000000)
	if err != nil {
		return Config{}, err
	}
	if targetTx <= 0 {
		return Config{}, fmt.Errorf("BULK_TARGET_TX must be > 0")
	}
	duration, err := mustDurationSeconds("BULK_DURATION_SECONDS", 0)
	if err != nil {
		return Config{}, err
	}
	batchSize, err := mustInt("BULK_BATCH_SIZE", 50)
	if err != nil {
		return Config{}, err
	}
	if batchSize <= 0 {
		return Config{}, fmt.Errorf("BULK_BATCH_SIZE must be > 0")
	}
	lookahead, err := mustInt("BULK_MAX_NONCE_LOOKAHEAD", 75)
	if err != nil {
		return Config{}, err
	}
	if lookahead <= 0 {
		return Config{}, fmt.Errorf("BULK_MAX_NONCE_LOOKAHEAD must be > 0")
	}
	maxReads, err := mustInt("BULK_MAX_CONCURRENT_READS", 24)
	if err != nil {
		return Config{}, err
	}
	if maxReads <= 0 {
		return Config{}, fmt.Errorf("BULK_MAX_CONCURRENT_READS must be > 0")
	}
	maxWallets, err := mustInt("BULK_MAX_ACTIVE_WALLETS", 0)
	if err != nil {
		return Config{}, err
	}
	idleSleep, err := mustDurationMillis("BULK_IDLE_SLEEP_MS", 300)
	if err != nil {
		return Config{}, err
	}
	confirmWorkers, err := mustInt("BULK_CONFIRM_WORKERS", 32)
	if err != nil {
		return Config{}, err
	}
	if confirmWorkers <= 0 {
		return Config{}, fmt.Errorf("BULK_CONFIRM_WORKERS must be > 0")
	}
	pollInterval, err := mustDurationSeconds("BULK_POLL_INTERVAL_SECONDS", 3)
	if err != nil {
		return Config{}, err
	}
	confirmTimeout, err := mustDurationSeconds("BULK_CONFIRM_TIMEOUT_SECONDS", 0)
	if err != nil {
		return Config{}, err
	}
	waitConfirmRaw := getenv("BULK_WAIT_CONFIRM", "false")
	waitConfirm, err := strconv.ParseBool(waitConfirmRaw)
	if err != nil {
		return Config{}, fmt.Errorf("BULK_WAIT_CONFIRM must be bool: %w", err)
	}
	httpTO, err := mustDurationSeconds("HTTP_TIMEOUT_SECONDS", 12)
	if err != nil {
		return Config{}, err
	}
	shardFilter, err := mustInt("BULK_SHARD_FILTER", -1)
	if err != nil {
		return Config{}, err
	}
	gasLimit, err := mustUint64("GAS_LIMIT", 50000)
	if err != nil {
		return Config{}, err
	}
	gasPrice, err := mustUint64("GAS_PRICE", 1000000000)
	if err != nil {
		return Config{}, err
	}
	routingMode := strings.ToLower(strings.TrimSpace(getenv("BULK_ROUTING_MODE", "same-shard")))
	switch routingMode {
	case "same-shard", "cross-shard":
	default:
		return Config{}, fmt.Errorf("BULK_ROUTING_MODE must be same-shard or cross-shard")
	}

	return Config{
		GatewayURL:         getenv("GATEWAY_URL", "https://api.battleofnodes.com"),
		ChainID:            getenv("CHAIN_ID", "B"),
		TxVersion:          uint32(txVersion),
		ManifestPath:       getenv("WALLETS_MANIFEST", "./configs/wallets-manifest.json"),
		SenderStatuses:     splitCSV(getenv("BULK_SENDER_STATUSES", "active_candidate,warm_reserve,window_b_reserve")),
		RequiredTags:       splitCSV(getenv("BULK_REQUIRED_TAGS", "sender")),
		ShardFilter:        shardFilter,
		MaxActiveWallets:   maxWallets,
		RoutingMode:        routingMode,
		Duration:           duration,
		TargetTx:           targetTx,
		BatchSize:          batchSize,
		MaxNonceLookahead:  lookahead,
		MaxConcurrentReads: maxReads,
		LoopIdleSleep:      idleSleep,
		ConfirmWorkers:     confirmWorkers,
		PollInterval:       pollInterval,
		ConfirmTimeout:     confirmTimeout,
		WaitConfirm:        waitConfirm,
		HTTPTimeout:        httpTO,
		Value:              getenv("BULK_VALUE", "50000000000000"),
		GasLimit:           gasLimit,
		GasPrice:           gasPrice,
	}, nil
}
