package main

import (
	"flag"
	"fmt"
	"log"

	"kwakbon/internal/manifest"
)

func main() {
	path := flag.String("manifest", "./configs/wallets-manifest.json", "Path to wallets manifest JSON")
	flag.Parse()

	m, err := manifest.Load(*path)
	if err != nil {
		log.Fatalf("load manifest: %v", err)
	}
	if err := m.Validate(); err != nil {
		log.Fatalf("validate manifest: %v", err)
	}

	s := m.Summary()
	fmt.Printf("manifest=%s version=%d wallets=%d enabled=%d disabled=%d\n", *path, m.Version, len(m.Wallets), s.TotalEnabled, s.TotalDisabled)
	fmt.Println("shards:")
	for _, shard := range s.SortedShards() {
		fmt.Printf("  shard=%d count=%d\n", shard, s.ByShard[shard])
	}
	fmt.Println("statuses:")
	for _, status := range s.SortedStatuses() {
		fmt.Printf("  status=%s count=%d\n", status, s.ByStatus[status])
	}
	fmt.Printf("shard2_enabled_senders=%d\n", s.Shard2Senders)
	fmt.Printf("shard2_enabled_receivers=%d\n", s.Shard2Receivers)
}
