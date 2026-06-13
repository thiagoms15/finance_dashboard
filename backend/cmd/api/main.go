package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thiago/finance/backend/internal/auth"
	"github.com/thiago/finance/backend/internal/config"
	"github.com/thiago/finance/backend/internal/observability"
	"github.com/thiago/finance/backend/internal/repository"
	"github.com/thiago/finance/backend/internal/service"
	httpHandlers "github.com/thiago/finance/backend/internal/transport/http/handlers"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := observability.NewLogger("api")
	ctx := context.Background()

	store, err := repository.NewPostgresStore(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer store.Close()
	instrumentedStore := observability.NewInstrumentedStore(store)

	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience)
	hasher := auth.PasswordHasher{
		Time:    cfg.Argon2Time,
		Memory:  cfg.Argon2Memory,
		Threads: cfg.Argon2Threads,
		KeyLen:  cfg.Argon2KeyLen,
	}
	svc := service.New(instrumentedStore, tokenManager, hasher, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	router := httpHandlers.NewRouter(cfg, svc, tokenManager, logger)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("api server started", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("api server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
	}
}
