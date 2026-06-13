package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/thiago/finance/backend/internal/config"
	"github.com/thiago/finance/backend/internal/jobs"
	"github.com/thiago/finance/backend/internal/marketdata"
	"github.com/thiago/finance/backend/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	store, err := repository.NewPostgresStore(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	job := jobs.NewPriceSyncJob(store, marketdata.NewLiveProvider(cfg), logger)
	if err := job.Schedule(ctx); err != nil && err != context.Canceled {
		logger.Error("worker stopped with error", "error", err)
		os.Exit(1)
	}
}
