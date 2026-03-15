package funding

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
	PollInterval    time.Duration
	HTTPTimeout     time.Duration
	ManifestPath    string
	TreasuryAddress string
	TreasuryPemPath string

	IncludeStatuses []string
	ExcludeStatuses []string
	RequiredTags    []string
	ShardFilter     int

	TargetAmountByStatus map[string]string
	NonceRefreshEvery    int
	SendCooldown         time.Duration
	NonceRetryCooldown   time.Duration
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
	gasPrice, err := mustUint64("GAS_PRICE", 1000000000)
	if err != nil {
		return Config{}, err
	}
	gasLimit, err := mustUint64("FUND_GAS_LIMIT", 50000)
	if err != nil {
		return Config{}, err
	}
	waitConfirm, err := mustBool("WAIT_CONFIRM", true)
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
	shardFilter, err := mustInt("FUND_SHARD_FILTER", -1)
	if err != nil {
		return Config{}, err
	}
	nonceRefreshEvery, err := mustInt("FUND_NONCE_REFRESH_EVERY", 75)
	if err != nil {
		return Config{}, err
	}
	if nonceRefreshEvery <= 0 {
		return Config{}, fmt.Errorf("FUND_NONCE_REFRESH_EVERY must be > 0")
	}
	sendCooldown, err := mustDurationMillis("FUND_SEND_COOLDOWN_MS", 25)
	if err != nil {
		return Config{}, err
	}
	nonceRetryCooldown, err := mustDurationMillis("FUND_NONCE_RETRY_COOLDOWN_MS", 750)
	if err != nil {
		return Config{}, err
	}

	targets := map[string]string{
		"active_candidate": getenv("FUND_TARGET_ACTIVE_EGLD", "0.6"),
		"warm_reserve":     getenv("FUND_TARGET_RESERVE_EGLD", "0.2"),
		"window_b_reserve": getenv("FUND_TARGET_WINDOW_B_RESERVE_EGLD", "0.2"),
	}

	cfg := Config{
		GatewayURL:      getenv("GATEWAY_URL", "https://api.battleofnodes.com"),
		ChainID:         getenv("CHAIN_ID", "B"),
		TxVersion:       uint32(txVersion),
		GasPrice:        gasPrice,
		GasLimit:        gasLimit,
		WaitConfirm:     waitConfirm,
		PollInterval:    poll,
		HTTPTimeout:     httpTO,
		ManifestPath:    getenv("WALLETS_MANIFEST", "./configs/wallets-manifest.json"),
		TreasuryAddress: getenv("TREASURY_ADDRESS", ""),
		TreasuryPemPath: getenv("TREASURY_PEM_PATH", ""),
		IncludeStatuses: splitCSV(getenv("FUND_INCLUDE_STATUSES", "active_candidate,warm_reserve")),
		ExcludeStatuses: splitCSV(getenv("FUND_EXCLUDE_STATUSES", "treasury,receiver")),
		RequiredTags:    splitCSV(getenv("FUND_REQUIRED_TAGS", "window_a")),
		ShardFilter:     shardFilter,
		TargetAmountByStatus: targets,
		NonceRefreshEvery:    nonceRefreshEvery,
		SendCooldown:         sendCooldown,
		NonceRetryCooldown:   nonceRetryCooldown,
	}

	if cfg.TreasuryAddress == "" || cfg.TreasuryPemPath == "" {
		return Config{}, fmt.Errorf("TREASURY_ADDRESS and TREASURY_PEM_PATH are required")
	}
	return cfg, nil
}
