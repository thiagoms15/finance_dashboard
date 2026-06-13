package portfolio

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/thiago/finance/backend/internal/domain"
)

type Position struct {
	Asset           domain.Asset    `json:"asset"`
	Quantity        decimal.Decimal `json:"quantity"`
	AverageCost     decimal.Decimal `json:"averageCost"`
	TotalCost       decimal.Decimal `json:"totalCost"`
	CurrentPrice    decimal.Decimal `json:"currentPrice"`
	CurrentValue    decimal.Decimal `json:"currentValue"`
	CurrentCurrency string          `json:"currentCurrency"`
	UnrealizedPL    decimal.Decimal `json:"unrealizedPL"`
	UnrealizedPLPct decimal.Decimal `json:"unrealizedPLPct"`
	DailyChange     decimal.Decimal `json:"dailyChange"`
	DailyChangePct  decimal.Decimal `json:"dailyChangePct"`
	RealizedPL      decimal.Decimal `json:"realizedPL"`
}

type Snapshot struct {
	Positions []Position       `json:"positions"`
	Summary   PortfolioSummary `json:"summary"`
}

type PortfolioSummary struct {
	PreferredCurrency string          `json:"preferredCurrency"`
	TotalInvested     decimal.Decimal `json:"totalInvested"`
	CurrentValue      decimal.Decimal `json:"currentValue"`
	TotalProfitLoss   decimal.Decimal `json:"totalProfitLoss"`
	DailyGainLoss     decimal.Decimal `json:"dailyGainLoss"`
	RealizedProfit    decimal.Decimal `json:"realizedProfit"`
	DividendsReceived decimal.Decimal `json:"dividendsReceived"`
}

type PerformancePoint struct {
	Date  time.Time       `json:"date"`
	Value decimal.Decimal `json:"value"`
}

type priceLookup map[uuid.UUID]domain.AssetPrice

func Calculate(
	assets map[uuid.UUID]domain.Asset,
	transactions []domain.Transaction,
	dividends []domain.Dividend,
	prices []domain.AssetPrice,
	preferredCurrency string,
	convert func(amount decimal.Decimal, from, to string) decimal.Decimal,
) (Snapshot, error) {
	orderedTransactions := append([]domain.Transaction(nil), transactions...)
	sort.SliceStable(orderedTransactions, func(i, j int) bool {
		if orderedTransactions[i].TransactionDate.Equal(orderedTransactions[j].TransactionDate) {
			return orderedTransactions[i].CreatedAt.Before(orderedTransactions[j].CreatedAt)
		}
		return orderedTransactions[i].TransactionDate.Before(orderedTransactions[j].TransactionDate)
	})

	positions := make(map[uuid.UUID]*Position)
	priceMap := make(priceLookup, len(prices))
	for _, price := range prices {
		priceMap[price.AssetID] = price
	}

	summary := PortfolioSummary{
		PreferredCurrency: preferredCurrency,
	}

	for _, txn := range orderedTransactions {
		asset, ok := assets[txn.AssetID]
		if !ok {
			return Snapshot{}, fmt.Errorf("asset %s not found", txn.AssetID)
		}

		position, ok := positions[txn.AssetID]
		if !ok {
			position = &Position{Asset: asset}
			positions[txn.AssetID] = position
		}

		switch txn.Type {
		case domain.TransactionTypeBuy:
			totalSpend := txn.Quantity.Mul(txn.Price).Add(txn.Fees)
			position.TotalCost = position.TotalCost.Add(convert(totalSpend, txn.Currency, preferredCurrency))
			position.Quantity = position.Quantity.Add(txn.Quantity)
			if !position.Quantity.IsZero() {
				position.AverageCost = position.TotalCost.Div(position.Quantity)
			}
		case domain.TransactionTypeSell:
			if position.Quantity.LessThan(txn.Quantity) {
				return Snapshot{}, fmt.Errorf("cannot sell %s units of %s when holding %s", txn.Quantity, asset.Symbol, position.Quantity)
			}

			costRemoved := position.AverageCost.Mul(txn.Quantity)
			sellTotal := txn.Quantity.Mul(txn.Price).Sub(txn.Fees)
			sellTotalConverted := convert(sellTotal, txn.Currency, preferredCurrency)
			position.RealizedPL = position.RealizedPL.Add(sellTotalConverted.Sub(costRemoved))
			position.Quantity = position.Quantity.Sub(txn.Quantity)
			position.TotalCost = position.TotalCost.Sub(costRemoved)
			if position.Quantity.IsZero() {
				position.AverageCost = decimal.Zero
				position.TotalCost = decimal.Zero
			}
		}
	}

	for _, dividend := range dividends {
		summary.DividendsReceived = summary.DividendsReceived.Add(
			convert(dividend.Amount, dividend.Currency, preferredCurrency),
		)
	}

	result := make([]Position, 0, len(positions))
	for assetID, position := range positions {
		price, ok := priceMap[assetID]
		if ok {
			priceConverted := convert(price.Price, price.Currency, preferredCurrency)
			prevCloseConverted := convert(price.PreviousClose, price.Currency, preferredCurrency)
			position.CurrentPrice = priceConverted
			position.CurrentCurrency = preferredCurrency
			position.CurrentValue = priceConverted.Mul(position.Quantity)
			position.UnrealizedPL = position.CurrentValue.Sub(position.TotalCost)
			if !position.TotalCost.IsZero() {
				position.UnrealizedPLPct = position.UnrealizedPL.Div(position.TotalCost).Mul(decimal.NewFromInt(100))
			}
			position.DailyChange = priceConverted.Sub(prevCloseConverted).Mul(position.Quantity)
			if !prevCloseConverted.IsZero() && !position.Quantity.IsZero() {
				position.DailyChangePct = priceConverted.Sub(prevCloseConverted).Div(prevCloseConverted).Mul(decimal.NewFromInt(100))
			}
		}

		summary.TotalInvested = summary.TotalInvested.Add(position.TotalCost)
		summary.CurrentValue = summary.CurrentValue.Add(position.CurrentValue)
		summary.RealizedProfit = summary.RealizedProfit.Add(position.RealizedPL)
		summary.DailyGainLoss = summary.DailyGainLoss.Add(position.DailyChange)

		if position.Quantity.GreaterThan(decimal.Zero) {
			result = append(result, *position)
		}
	}

	summary.TotalProfitLoss = summary.CurrentValue.Sub(summary.TotalInvested).Add(summary.RealizedProfit)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Asset.Symbol < result[j].Asset.Symbol
	})

	return Snapshot{Positions: result, Summary: summary}, nil
}

func BuildPerformanceSeries(transactions []domain.Transaction, preferredCurrency string, convert func(amount decimal.Decimal, from, to string) decimal.Decimal) []PerformancePoint {
	points := make([]PerformancePoint, 0, len(transactions))
	cumulative := decimal.Zero

	for _, txn := range transactions {
		switch txn.Type {
		case domain.TransactionTypeBuy:
			cumulative = cumulative.Add(convert(txn.Quantity.Mul(txn.Price).Add(txn.Fees), txn.Currency, preferredCurrency))
		case domain.TransactionTypeSell:
			cumulative = cumulative.Sub(convert(txn.Quantity.Mul(txn.Price).Sub(txn.Fees), txn.Currency, preferredCurrency))
		}

		points = append(points, PerformancePoint{
			Date:  txn.TransactionDate,
			Value: cumulative,
		})
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].Date.Before(points[j].Date)
	})

	return points
}
