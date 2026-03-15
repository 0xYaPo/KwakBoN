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

	"kwakbon/internal/sweep"
)

func main() {
	var dryRun bool
	flag.BoolVar(&dryRun, "dry-run", false, "Preview matched wallets and sweepable amounts without sending")
	flag.Parse()

	cfg, err := sweep.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	app, err := sweep.New(cfg, sweep.Options{DryRun: dryRun})
	if err != nil {
		log.Fatalf("init error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\n[sweepwallets] stopping...")
		cancel()
	}()

	start := time.Now()
	if err := app.Run(ctx); err != nil {
		log.Fatalf("run error: %v", err)
	}
	fmt.Printf("[sweepwallets] done in %s\n", time.Since(start))
}
