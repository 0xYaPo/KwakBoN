package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kwakbon/internal/funding"
)

func main() {
	var dryRun bool
	flag.BoolVar(&dryRun, "dry-run", false, "Preview matched wallets and deficits without sending")
	flag.Parse()

	cfg, err := funding.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	app, err := funding.New(cfg, funding.Options{DryRun: dryRun})
	if err != nil {
		log.Fatalf("init error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\n[fundwallets] stopping...")
		cancel()
	}()

	start := time.Now()
	if err := app.Run(ctx); err != nil {
		log.Fatalf("run error: %v", err)
	}
	fmt.Printf("[fundwallets] done in %s\n", time.Since(start))
}
