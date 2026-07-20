package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gabrielnakaema/project-chat/internal/platform/config"
	"github.com/gabrielnakaema/project-chat/internal/platform/kafka"
	"github.com/gabrielnakaema/project-chat/internal/platform/logger"
	"github.com/gabrielnakaema/project-chat/internal/platform/outbox"
	"github.com/gabrielnakaema/project-chat/internal/platform/postgres"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("outbox-relay: load config: %v", err)
	}

	log := logger.Init(cfg)

	pool, err := postgres.NewPool(cfg)
	if err != nil {
		log.Error("outbox-relay: create db pool", "error", err.Error())
		os.Exit(1)
	}
	defer pool.Close()

	producer, err := kafka.NewSyncPublisher(cfg)
	if err != nil {
		log.Error("outbox-relay: create sync producer", "error", err.Error())
		os.Exit(1)
	}

	relay := outbox.NewRelay(pool, producer, log, cfg.OutboxPollInterval, cfg.OutboxBatchSize)
	defer relay.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go relay.Run(ctx)
	log.Info("outbox relay started", "poll_interval", cfg.OutboxPollInterval.String(), "batch_size", cfg.OutboxBatchSize)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit

	log.Info("outbox relay shutting down", "signal", s.String())
	cancel()
}
