package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thiago/finance/backend/internal/config"
	"github.com/thiago/finance/backend/internal/jobs"
	"github.com/thiago/finance/backend/internal/marketdata"
	"github.com/thiago/finance/backend/internal/observability"
	"github.com/thiago/finance/backend/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := observability.NewLogger("worker")
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if cfg.MetricsAddr != "" {
		metricsServer := &http.Server{
			Addr:              cfg.MetricsAddr,
			Handler:           http.NewServeMux(),
			ReadHeaderTimeout: 5 * time.Second,
		}
		metricsServer.Handler.(*http.ServeMux).Handle("/metrics", observability.MetricsHandler())
		metricsServer.Handler.(*http.ServeMux).HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})

		go func() {
			logger.Info("worker metrics server started", "addr", cfg.MetricsAddr)
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("worker metrics server failed", "error", err)
			}
		}()

		go func() {
			<-ctx.Done()
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			_ = metricsServer.Shutdown(shutdownCtx)
		}()
	}

	store, err := repository.NewPostgresStore(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer store.Close()
	instrumentedStore := observability.NewInstrumentedStore(store)

	job := jobs.NewPriceSyncJob(instrumentedStore, marketdata.NewLiveProvider(cfg), logger)
	if err := job.Schedule(ctx); err != nil && err != context.Canceled {
		logger.Error("worker stopped with error", "error", err)
		os.Exit(1)
	}
}
