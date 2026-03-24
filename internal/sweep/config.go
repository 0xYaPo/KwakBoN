package sweep

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	GatewayURL      string
	ChainID         string
	TxVersion       uint32
	GasPrice        uint64
	GasLimit        uint64
	WaitConfirm     bool
	ConfirmWorkers  int
	PollInterval    time.Duration
	ConfirmTimeout  time.Duration
	HTTPTimeout     time.Duration
	ManifestPath    string
	TreasuryAddress string

	IncludeStatuses []string
	ExcludeStatuses []string
	RequiredTags    []string
	ShardFilter     int
	MinRemainEGLD   string
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
	gasPrice, err := mustUint64("GAS_PRICE", 1000000000)
	if err != nil {
		return Config{}, err
	}
	gasLimit, err := mustUint64("SWEEP_GAS_LIMIT", 50000)
	if err != nil {
		return Config{}, err
	}
	waitConfirm, err := mustBool("WAIT_CONFIRM", true)
	if err != nil {
		return Config{}, err
	}
	confirmWorkers, err := mustInt("SWEEP_CONFIRM_WORKERS", 32)
	if err != nil {
		return Config{}, err
	}
	if confirmWorkers <= 0 {
		return Config{}, fmt.Errorf("SWEEP_CONFIRM_WORKERS must be > 0")
	}
	poll, err := mustDurationSeconds("POLL_INTERVAL_SECONDS", 3)
	if err != nil {
		return Config{}, err
	}
	confirmTimeout, err := mustDurationSeconds("SWEEP_CONFIRM_TIMEOUT_SECONDS", 0)
	if err != nil {
		return Config{}, err
	}
	httpTO, err := mustDurationSeconds("HTTP_TIMEOUT_SECONDS", 12)
	if err != nil {
		return Config{}, err
	}
	shardFilter, err := mustInt("SWEEP_SHARD_FILTER", -1)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		GatewayURL:      getenv("GATEWAY_URL", "https://api.battleofnodes.com"),
		ChainID:         getenv("CHAIN_ID", "B"),
		TxVersion:       uint32(txVersion),
		GasPrice:        gasPrice,
		GasLimit:        gasLimit,
		WaitConfirm:     waitConfirm,
		ConfirmWorkers:  confirmWorkers,
		PollInterval:    poll,
		ConfirmTimeout:  confirmTimeout,
		HTTPTimeout:     httpTO,
		ManifestPath:    getenv("WALLETS_MANIFEST", "./configs/wallets-manifest.json"),
		TreasuryAddress: getenv("TREASURY_ADDRESS", ""),
		IncludeStatuses: splitCSV(getenv("SWEEP_INCLUDE_STATUSES", "active_candidate,warm_reserve,window_b_reserve")),
		ExcludeStatuses: splitCSV(getenv("SWEEP_EXCLUDE_STATUSES", "treasury,receiver")),
		RequiredTags:    splitCSV(getenv("SWEEP_REQUIRED_TAGS", "")),
		ShardFilter:     shardFilter,
		MinRemainEGLD:   getenv("SWEEP_MIN_REMAIN_EGLD", "0"),
	}

	if cfg.TreasuryAddress == "" {
		return Config{}, fmt.Errorf("TREASURY_ADDRESS is required")
	}
	return cfg, nil
}
