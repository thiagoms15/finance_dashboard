package portfolio

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/thiago/finance/backend/internal/domain"
)

func TestCalculateSnapshot(t *testing.T) {
	t.Parallel()

	assetID := uuid.New()
	assets := map[uuid.UUID]domain.Asset{
		assetID: {
			ID:       assetID,
			Symbol:   "AAPL",
			Exchange: "NASDAQ",
			Currency: "USD",
		},
	}

	transactions := []domain.Transaction{
		{
			AssetID:         assetID,
			Type:            domain.TransactionTypeBuy,
			Quantity:        decimal.RequireFromString("10"),
			Price:           decimal.RequireFromString("100"),
			Fees:            decimal.RequireFromString("5"),
			Currency:        "USD",
			TransactionDate: time.Now().Add(-48 * time.Hour),
		},
		{
			AssetID:         assetID,
			Type:            domain.TransactionTypeSell,
			Quantity:        decimal.RequireFromString("3"),
			Price:           decimal.RequireFromString("120"),
			Fees:            decimal.RequireFromString("2"),
			Currency:        "USD",
			TransactionDate: time.Now().Add(-24 * time.Hour),
		},
	}

	dividends := []domain.Dividend{
		{
			AssetID:  assetID,
			Amount:   decimal.RequireFromString("10"),
			Currency: "USD",
		},
	}

	prices := []domain.AssetPrice{
		{
			AssetID:       assetID,
			Price:         decimal.RequireFromString("130"),
			PreviousClose: decimal.RequireFromString("125"),
			Currency:      "USD",
			Timestamp:     time.Now(),
		},
	}

	snapshot, err := Calculate(assets, transactions, dividends, prices, "USD", func(amount decimal.Decimal, from, to string) decimal.Decimal {
		return amount
	})
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	if got, want := len(snapshot.Positions), 1; got != want {
		t.Fatalf("len(snapshot.Positions) = %d, want %d", got, want)
	}

	position := snapshot.Positions[0]
	if got, want := position.Quantity.StringFixed(0), "7"; got != want {
		t.Fatalf("position.Quantity = %s, want %s", got, want)
	}
	if got, want := position.RealizedPL.StringFixed(1), "56.5"; got != want {
		t.Fatalf("position.RealizedPL = %s, want %s", got, want)
	}
	if got, want := snapshot.Summary.DividendsReceived.StringFixed(0), "10"; got != want {
		t.Fatalf("summary.DividendsReceived = %s, want %s", got, want)
	}
}

func TestCalculateRejectsOversell(t *testing.T) {
	t.Parallel()

	assetID := uuid.New()
	assets := map[uuid.UUID]domain.Asset{
		assetID: {ID: assetID, Symbol: "AAPL", Currency: "USD"},
	}

	_, err := Calculate(
		assets,
		[]domain.Transaction{
			{
				AssetID:  assetID,
				Type:     domain.TransactionTypeSell,
				Quantity: decimal.RequireFromString("1"),
				Price:    decimal.RequireFromString("100"),
				Currency: "USD",
			},
		},
		nil,
		nil,
		"USD",
		func(amount decimal.Decimal, from, to string) decimal.Decimal { return amount },
	)
	if err == nil {
		t.Fatal("Calculate() error = nil, want oversell error")
	}
}

func TestCalculateSortsTransactionsChronologically(t *testing.T) {
	t.Parallel()

	assetID := uuid.New()
	assets := map[uuid.UUID]domain.Asset{
		assetID: {
			ID:       assetID,
			Symbol:   "AAPL",
			Exchange: "NASDAQ",
			Currency: "USD",
		},
	}

	now := time.Now()
	snapshot, err := Calculate(
		assets,
		[]domain.Transaction{
			{
				AssetID:         assetID,
				Type:            domain.TransactionTypeSell,
				Quantity:        decimal.RequireFromString("2"),
				Price:           decimal.RequireFromString("120"),
				Currency:        "USD",
				TransactionDate: now.Add(2 * time.Hour),
				CreatedAt:       now.Add(2 * time.Hour),
			},
			{
				AssetID:         assetID,
				Type:            domain.TransactionTypeBuy,
				Quantity:        decimal.RequireFromString("5"),
				Price:           decimal.RequireFromString("100"),
				Currency:        "USD",
				TransactionDate: now,
				CreatedAt:       now,
			},
		},
		nil,
		[]domain.AssetPrice{
			{
				AssetID:       assetID,
				Price:         decimal.RequireFromString("130"),
				PreviousClose: decimal.RequireFromString("125"),
				Currency:      "USD",
				Timestamp:     now.Add(3 * time.Hour),
			},
		},
		"USD",
		func(amount decimal.Decimal, from, to string) decimal.Decimal { return amount },
	)
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	if got, want := len(snapshot.Positions), 1; got != want {
		t.Fatalf("len(snapshot.Positions) = %d, want %d", got, want)
	}

	if got, want := snapshot.Positions[0].Quantity.StringFixed(0), "3"; got != want {
		t.Fatalf("position.Quantity = %s, want %s", got, want)
	}
}
