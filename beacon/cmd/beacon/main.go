package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"beacon/config"
	"beacon/internal/di"
)

func main() {
	cfg := config.Load()

	if cfg.DiscordWebhookURL == "" {
		log.Fatal("❌ DISCORD_WEBHOOK_URL is required")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	svc := di.Wire(cfg)

	if err := svc.Run(ctx); err != nil {
		log.Fatalf("❌ Beacon failed: %v", err)
	}
}
