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

	"kwakbon/internal/sprint"
)

func main() {
	var dryRun bool
	flag.BoolVar(&dryRun, "dry-run", false, "Print sample sender/receiver routing but do not sign/send")
	flag.Parse()

	cfg, err := sprint.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	app, err := sprint.New(cfg, sprint.Options{DryRun: dryRun})
	if err != nil {
		log.Fatalf("init error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\n[windowsprint] stopping...")
		cancel()
	}()

	start := time.Now()
	if err := app.Run(ctx); err != nil {
		log.Fatalf("run error: %v", err)
	}
	fmt.Printf("[windowsprint] done in %s\n", time.Since(start))
}
