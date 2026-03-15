package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Manifest struct {
	Version     int            `json:"version"`
	GeneratedAt string         `json:"generatedAt,omitempty"`
	Wallets     []WalletRecord `json:"wallets"`
}

type WalletRecord struct {
	WalletID string   `json:"walletId"`
	Address  string   `json:"address"`
	Shard    int      `json:"shard"`
	PemFile  string   `json:"pemFile,omitempty"`
	PemPath  string   `json:"pemPath,omitempty"`
	Status   string   `json:"status"`
	Tags     []string `json:"tags,omitempty"`
	Enabled  bool     `json:"enabled"`
	Notes    string   `json:"notes,omitempty"`
}

type Summary struct {
	TotalEnabled   int
	TotalDisabled  int
	ByShard        map[int]int
	ByStatus       map[string]int
	ByTag          map[string]int
	Shard2Senders  int
	Shard2Receivers int
}

func Load(path string) (*Manifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if m.Version == 0 {
		m.Version = 1
	}
	if len(m.Wallets) == 0 {
		return nil, fmt.Errorf("manifest has no wallets")
	}

	return &m, nil
}

func LoadIfExists(path string) (*Manifest, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Manifest{Version: 1, Wallets: []WalletRecord{}}, nil
		}
		return nil, fmt.Errorf("stat manifest: %w", err)
	}
	return Load(path)
}

func (m *Manifest) Save(path string) error {
	if m.Version == 0 {
		m.Version = 1
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create manifest dir: %w", err)
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	return nil
}

func (m *Manifest) Validate() error {
	if m.Version <= 0 {
		return fmt.Errorf("manifest version must be > 0")
	}

	ids := make(map[string]struct{}, len(m.Wallets))
	addrs := make(map[string]struct{}, len(m.Wallets))
	for i, w := range m.Wallets {
		if strings.TrimSpace(w.WalletID) == "" {
			return fmt.Errorf("wallet %d missing walletId", i)
		}
		if strings.TrimSpace(w.Address) == "" {
			return fmt.Errorf("wallet %s missing address", w.WalletID)
		}
		if strings.TrimSpace(w.Status) == "" {
			return fmt.Errorf("wallet %s missing status", w.WalletID)
		}
		if w.Shard < 0 {
			return fmt.Errorf("wallet %s has invalid shard %d", w.WalletID, w.Shard)
		}
		if _, ok := ids[w.WalletID]; ok {
			return fmt.Errorf("duplicate walletId %s", w.WalletID)
		}
		if _, ok := addrs[w.Address]; ok {
			return fmt.Errorf("duplicate address %s", w.Address)
		}
		ids[w.WalletID] = struct{}{}
		addrs[w.Address] = struct{}{}
	}

	return nil
}

func (m *Manifest) Summary() Summary {
	s := Summary{
		ByShard:  make(map[int]int),
		ByStatus: make(map[string]int),
		ByTag:    make(map[string]int),
	}

	for _, w := range m.Wallets {
		if w.Enabled {
			s.TotalEnabled++
		} else {
			s.TotalDisabled++
		}
		s.ByShard[w.Shard]++
		s.ByStatus[w.Status]++
		for _, tag := range w.Tags {
			s.ByTag[tag]++
		}
		if w.Enabled && w.Shard == 2 && hasTag(w.Tags, "sender") {
			s.Shard2Senders++
		}
		if w.Enabled && w.Shard == 2 && hasTag(w.Tags, "sink") {
			s.Shard2Receivers++
		}
	}

	return s
}

func (s Summary) SortedShards() []int {
	out := make([]int, 0, len(s.ByShard))
	for shard := range s.ByShard {
		out = append(out, shard)
	}
	sort.Ints(out)
	return out
}

func (s Summary) SortedStatuses() []string {
	out := make([]string, 0, len(s.ByStatus))
	for status := range s.ByStatus {
		out = append(out, status)
	}
	sort.Strings(out)
	return out
}

func (s Summary) SortedTags() []string {
	out := make([]string, 0, len(s.ByTag))
	for tag := range s.ByTag {
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}

func hasTag(tags []string, want string) bool {
	for _, tag := range tags {
		if strings.EqualFold(strings.TrimSpace(tag), want) {
			return true
		}
	}
	return false
}
