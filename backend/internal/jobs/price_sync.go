package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/thiago/finance/backend/internal/domain"
	"github.com/thiago/finance/backend/internal/marketdata"
	"github.com/thiago/finance/backend/internal/observability"
)

type PriceStore interface {
	ListAssets(ctx context.Context, search string, limit int) ([]domain.Asset, error)
	UpsertAssetPrice(ctx context.Context, price domain.AssetPrice) error
	UpsertExchangeRate(ctx context.Context, rate domain.ExchangeRate) error
}

type PriceSyncJob struct {
	store    PriceStore
	provider marketdata.Provider
	logger   *slog.Logger
}

func NewPriceSyncJob(store PriceStore, provider marketdata.Provider, logger *slog.Logger) *PriceSyncJob {
	return &PriceSyncJob{store: store, provider: provider, logger: logger}
}

func (j *PriceSyncJob) RunOnce(ctx context.Context) error {
	startedAt := time.Now()
	assets, err := j.store.ListAssets(ctx, "", 1000)
	if err != nil {
		observability.ObserveMarketRefresh("error", time.Since(startedAt), 0, 0)
		return err
	}
	prices, err := j.provider.LatestPrices(ctx, assets)
	if err != nil {
		observability.ObserveMarketRefresh("error", time.Since(startedAt), 0, 0)
		return err
	}
	for _, price := range prices {
		if err := j.store.UpsertAssetPrice(ctx, price); err != nil {
			observability.ObserveMarketRefresh("error", time.Since(startedAt), len(prices), 0)
			return err
		}
	}

	rates, err := j.provider.LatestExchangeRates(ctx)
	if err != nil {
		observability.ObserveMarketRefresh("error", time.Since(startedAt), len(prices), 0)
		return err
	}
	for _, rate := range rates {
		if err := j.store.UpsertExchangeRate(ctx, rate); err != nil {
			observability.ObserveMarketRefresh("error", time.Since(startedAt), len(prices), len(rates))
			return err
		}
	}

	duration := time.Since(startedAt)
	observability.ObserveMarketRefresh("success", duration, len(prices), len(rates))
	j.logger.Info(
		"price sync completed",
		"action", "MARKET_DATA_REFRESH",
		"assets", len(prices),
		"rates", len(rates),
		"duration_ms", duration.Milliseconds(),
	)
	return nil
}

func (j *PriceSyncJob) Schedule(ctx context.Context) error {
	if err := j.RunOnce(ctx); err != nil {
		j.logger.Error("initial price sync failed", "error", err)
	}

	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := j.RunOnce(ctx); err != nil {
				j.logger.Error("scheduled price sync failed", "error", err)
			}
		}
	}
}
