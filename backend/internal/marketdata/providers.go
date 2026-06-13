package marketdata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/thiago/finance/backend/internal/config"
	"github.com/thiago/finance/backend/internal/domain"
)

type Provider interface {
	LatestPrices(ctx context.Context, assets []domain.Asset) ([]domain.AssetPrice, error)
	LatestExchangeRates(ctx context.Context) ([]domain.ExchangeRate, error)
}

type LiveProvider struct {
	client       *http.Client
	brapiToken   string
	coinIDCache  map[string]string
	coinIDMu     sync.RWMutex
}

func NewLiveProvider(cfg config.Config) *LiveProvider {
	return &LiveProvider{
		client: &http.Client{Timeout: 5 * time.Second},
		brapiToken: strings.TrimSpace(cfg.BRAPIToken),
		coinIDCache: map[string]string{
			"BTC": "bitcoin",
			"ETH": "ethereum",
			"SOL": "solana",
			"USDT": "tether",
			"USDC": "usd-coin",
			"BNB": "binancecoin",
			"XRP": "ripple",
			"ADA": "cardano",
			"DOGE": "dogecoin",
			"AVAX": "avalanche-2",
		},
	}
}

func (p *LiveProvider) LatestPrices(ctx context.Context, assets []domain.Asset) ([]domain.AssetPrice, error) {
	prices := make([]domain.AssetPrice, 0, len(assets))
	var errs []string

	grouped := groupAssetsByExchange(assets)

	if b3Assets := grouped["B3"]; len(b3Assets) > 0 {
		b3Prices, err := p.fetchB3Prices(ctx, b3Assets)
		if err != nil {
			errs = append(errs, fmt.Sprintf("B3: %v", err))
		}
		prices = append(prices, b3Prices...)
	}

	if nasdaqAssets := append(grouped["NASDAQ"], grouped["NYSE"]...); len(nasdaqAssets) > 0 {
		stockPrices, err := p.fetchYahooPrices(ctx, nasdaqAssets, false)
		if err != nil {
			errs = append(errs, fmt.Sprintf("NASDAQ/NYSE: %v", err))
		}
		prices = append(prices, stockPrices...)
	}

	if cryptoAssets := grouped["CRYPTO"]; len(cryptoAssets) > 0 {
		cryptoPrices, err := p.fetchCoinGeckoPrices(ctx, cryptoAssets)
		if err != nil {
			errs = append(errs, fmt.Sprintf("CRYPTO: %v", err))
		}
		prices = append(prices, cryptoPrices...)
	}

	if len(prices) == 0 && len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, "; "))
	}

	return prices, nil
}

func (p *LiveProvider) LatestExchangeRates(ctx context.Context) ([]domain.ExchangeRate, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.frankfurter.dev/v1/latest?base=USD&symbols=BRL", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "finance-dashboard/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("frankfurter returned status %d", resp.StatusCode)
	}

	var payload struct {
		Base  string             `json:"base"`
		Date  string             `json:"date"`
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return nil, err
	}

	brlRate, ok := payload.Rates["BRL"]
	if !ok || brlRate <= 0 {
		return nil, fmt.Errorf("frankfurter did not return USD/BRL")
	}

	now := time.Now().UTC()
	brl := decimal.NewFromFloat(brlRate)
	return []domain.ExchangeRate{
		{ID: uuid.New(), Base: "USD", Quote: "BRL", Rate: brl, Timestamp: now},
		{ID: uuid.New(), Base: "BRL", Quote: "USD", Rate: decimal.NewFromInt(1).Div(brl), Timestamp: now},
	}, nil
}

func groupAssetsByExchange(assets []domain.Asset) map[string][]domain.Asset {
	grouped := make(map[string][]domain.Asset)
	for _, asset := range assets {
		exchange := strings.ToUpper(strings.TrimSpace(asset.Exchange))
		grouped[exchange] = append(grouped[exchange], asset)
	}
	return grouped
}

func (p *LiveProvider) fetchB3Prices(ctx context.Context, assets []domain.Asset) ([]domain.AssetPrice, error) {
	if len(assets) == 0 {
		return nil, nil
	}

	symbols := make([]string, 0, len(assets))
	assetBySymbol := make(map[string]domain.Asset, len(assets))
	for _, asset := range assets {
		symbol := strings.ToUpper(strings.TrimSpace(asset.Symbol))
		symbols = append(symbols, symbol)
		assetBySymbol[symbol] = asset
	}

	endpoint := "https://brapi.dev/api/quote/" + strings.Join(symbols, ",")
	values := url.Values{}
	if p.brapiToken != "" {
		values.Set("token", p.brapiToken)
	}
	if encoded := values.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "finance-dashboard/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return p.fetchYahooPrices(ctx, assets, true)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return p.fetchYahooPrices(ctx, assets, true)
	}

	var payload struct {
		Results []struct {
			Symbol                     string  `json:"symbol"`
			Currency                   string  `json:"currency"`
			RegularMarketPrice         float64 `json:"regularMarketPrice"`
			RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
		} `json:"results"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&payload); err != nil {
		return p.fetchYahooPrices(ctx, assets, true)
	}

	var prices []domain.AssetPrice
	for _, result := range payload.Results {
		asset, ok := assetBySymbol[strings.ToUpper(result.Symbol)]
		if !ok || result.RegularMarketPrice <= 0 {
			continue
		}
		prices = append(prices, domain.AssetPrice{
			ID:            uuid.New(),
			AssetID:       asset.ID,
			Price:         decimal.NewFromFloat(result.RegularMarketPrice),
			PreviousClose: decimal.NewFromFloat(defaultPreviousClose(result.RegularMarketPrice, result.RegularMarketPreviousClose)),
			Currency:      firstNonEmpty(result.Currency, asset.Currency),
			Timestamp:     time.Now().UTC(),
		})
	}

	if len(prices) == 0 {
		return p.fetchYahooPrices(ctx, assets, true)
	}
	return prices, nil
}

func (p *LiveProvider) fetchYahooPrices(ctx context.Context, assets []domain.Asset, b3Suffix bool) ([]domain.AssetPrice, error) {
	if len(assets) == 0 {
		return nil, nil
	}

	prices := make([]domain.AssetPrice, 0, len(assets))
	var errs []string
	for _, asset := range assets {
		yahooSymbol := strings.ToUpper(strings.TrimSpace(asset.Symbol))
		if b3Suffix {
			yahooSymbol += ".SA"
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://query1.finance.yahoo.com/v8/finance/chart/"+url.PathEscape(yahooSymbol)+"?range=5d&interval=1d", nil)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", yahooSymbol, err))
			continue
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 finance-dashboard/1.0")

		resp, err := p.client.Do(req)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", yahooSymbol, err))
			continue
		}

		func() {
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errs = append(errs, fmt.Sprintf("%s: yahoo chart status %d", yahooSymbol, resp.StatusCode))
				return
			}

			var payload struct {
				Chart struct {
					Result []struct {
						Meta struct {
							Currency           string  `json:"currency"`
							RegularMarketPrice float64 `json:"regularMarketPrice"`
							ChartPreviousClose float64 `json:"chartPreviousClose"`
						} `json:"meta"`
					} `json:"result"`
				} `json:"chart"`
			}
			if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&payload); err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", yahooSymbol, err))
				return
			}
			if len(payload.Chart.Result) == 0 || payload.Chart.Result[0].Meta.RegularMarketPrice <= 0 {
				errs = append(errs, fmt.Sprintf("%s: empty chart result", yahooSymbol))
				return
			}

			meta := payload.Chart.Result[0].Meta
			prices = append(prices, domain.AssetPrice{
				ID:            uuid.New(),
				AssetID:       asset.ID,
				Price:         decimal.NewFromFloat(meta.RegularMarketPrice),
				PreviousClose: decimal.NewFromFloat(defaultPreviousClose(meta.RegularMarketPrice, meta.ChartPreviousClose)),
				Currency:      firstNonEmpty(meta.Currency, asset.Currency),
				Timestamp:     time.Now().UTC(),
			})
		}()
	}

	if len(prices) == 0 && len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, "; "))
	}

	return prices, nil
}

func (p *LiveProvider) fetchCoinGeckoPrices(ctx context.Context, assets []domain.Asset) ([]domain.AssetPrice, error) {
	if len(assets) == 0 {
		return nil, nil
	}

	type lookup struct {
		asset domain.Asset
		id    string
	}
	lookups := make([]lookup, 0, len(assets))
	ids := make([]string, 0, len(assets))
	for _, asset := range assets {
		id, err := p.coinGeckoID(ctx, asset.Symbol)
		if err != nil || id == "" {
			continue
		}
		lookups = append(lookups, lookup{asset: asset, id: id})
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no crypto ids resolved")
	}

	vsCurrencies := "usd,brl"
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.coingecko.com/api/v3/simple/price?ids="+url.QueryEscape(strings.Join(ids, ","))+"&vs_currencies="+vsCurrencies+"&include_24hr_change=true",
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "finance-dashboard/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko returned status %d", resp.StatusCode)
	}

	var payload map[string]map[string]float64
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&payload); err != nil {
		return nil, err
	}

	prices := make([]domain.AssetPrice, 0, len(lookups))
	for _, lookup := range lookups {
		values, ok := payload[lookup.id]
		if !ok {
			continue
		}
		currency := strings.ToLower(strings.TrimSpace(lookup.asset.Currency))
		if !slices.Contains([]string{"usd", "brl"}, currency) {
			currency = "usd"
		}

		current := values[currency]
		if current <= 0 {
			continue
		}
		changePct := values[currency+"_24h_change"]
		previous := current
		factor := 1 + (changePct / 100)
		if factor > 0 {
			previous = current / factor
		}

		prices = append(prices, domain.AssetPrice{
			ID:            uuid.New(),
			AssetID:       lookup.asset.ID,
			Price:         decimal.NewFromFloat(current),
			PreviousClose: decimal.NewFromFloat(previous),
			Currency:      strings.ToUpper(currency),
			Timestamp:     time.Now().UTC(),
		})
	}

	return prices, nil
}

func (p *LiveProvider) coinGeckoID(ctx context.Context, symbol string) (string, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	p.coinIDMu.RLock()
	if id, ok := p.coinIDCache[symbol]; ok {
		p.coinIDMu.RUnlock()
		return id, nil
	}
	p.coinIDMu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.coingecko.com/api/v3/search?query="+url.QueryEscape(symbol), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "finance-dashboard/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("coingecko search returned status %d", resp.StatusCode)
	}

	var payload struct {
		Coins []struct {
			ID            string `json:"id"`
			Symbol        string `json:"symbol"`
			MarketCapRank int    `json:"market_cap_rank"`
		} `json:"coins"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return "", err
	}

	bestID := ""
	bestRank := int(^uint(0) >> 1)
	for _, coin := range payload.Coins {
		if strings.EqualFold(coin.Symbol, symbol) {
			rank := coin.MarketCapRank
			if rank <= 0 {
				rank = bestRank + 1
			}
			if bestID == "" || rank < bestRank {
				bestID = coin.ID
				bestRank = rank
			}
		}
	}
	if bestID == "" {
		return "", fmt.Errorf("coingecko id not found for %s", symbol)
	}

	p.coinIDMu.Lock()
	p.coinIDCache[symbol] = bestID
	p.coinIDMu.Unlock()
	return bestID, nil
}

func defaultPreviousClose(current, previous float64) float64 {
	if previous > 0 {
		return previous
	}
	return current
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.ToUpper(strings.TrimSpace(value))
		}
	}
	return ""
}
