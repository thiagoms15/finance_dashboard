package marketdata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/thiago/finance/backend/internal/config"
	"github.com/thiago/finance/backend/internal/domain"
)

var ErrIconNotFound = errors.New("asset icon not found")

type IconPayload struct {
	ContentType string
	Body        []byte
}

type IconResolver struct {
	client     *http.Client
	brapiToken string
}

var knownCryptoIconURLs = map[string]string{
	"BTC": "https://coin-images.coingecko.com/coins/images/1/large/bitcoin.png",
	"ETH": "https://coin-images.coingecko.com/coins/images/279/large/ethereum.png",
}

func NewIconResolver(cfg config.Config) *IconResolver {
	return &IconResolver{
		client: &http.Client{Timeout: 10 * time.Second},
		brapiToken: strings.TrimSpace(cfg.BRAPIToken),
	}
}

func (r *IconResolver) FetchAssetIcon(ctx context.Context, asset domain.Asset) (IconPayload, error) {
	switch strings.ToUpper(strings.TrimSpace(asset.Exchange)) {
	case "B3":
		return r.fetchB3Icon(ctx, asset.Symbol)
	case "CRYPTO":
		return r.fetchCryptoIcon(ctx, asset.Symbol)
	case "NASDAQ", "NYSE":
		return r.fetchUSStockIcon(ctx, asset.Symbol)
	default:
		return IconPayload{}, ErrIconNotFound
	}
}

func (r *IconResolver) fetchB3Icon(ctx context.Context, symbol string) (IconPayload, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return IconPayload{}, ErrIconNotFound
	}

	endpoint := "https://brapi.dev/api/quote/" + url.PathEscape(strings.ToUpper(strings.TrimSpace(symbol)))
	if r.brapiToken != "" {
		endpoint += "?token=" + url.QueryEscape(r.brapiToken)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return IconPayload{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "finance-dashboard/1.0")

	resp, err := r.client.Do(req)
	if err != nil {
		return IconPayload{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return r.fetchRemoteImage(ctx, "https://icons.brapi.dev/icons/"+url.PathEscape(symbol)+".svg")
	}

	var payload struct {
		Results []struct {
			LogoURL string `json:"logourl"`
		} `json:"results"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return IconPayload{}, err
	}
	if len(payload.Results) == 0 || strings.TrimSpace(payload.Results[0].LogoURL) == "" {
		return r.fetchRemoteImage(ctx, "https://icons.brapi.dev/icons/"+url.PathEscape(symbol)+".svg")
	}

	return r.fetchRemoteImage(ctx, payload.Results[0].LogoURL)
}

func (r *IconResolver) fetchCryptoIcon(ctx context.Context, symbol string) (IconPayload, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if iconURL, ok := knownCryptoIconURLs[symbol]; ok {
		return r.fetchRemoteImage(ctx, iconURL)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.coingecko.com/api/v3/search?query="+url.QueryEscape(strings.ToUpper(strings.TrimSpace(symbol))), nil)
	if err != nil {
		return IconPayload{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "finance-dashboard/1.0")

	resp, err := r.client.Do(req)
	if err != nil {
		return IconPayload{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return IconPayload{}, fmt.Errorf("coingecko returned status %d", resp.StatusCode)
	}

	var payload struct {
		Coins []struct {
			Symbol string `json:"symbol"`
			Large  string `json:"large"`
			Thumb  string `json:"thumb"`
		} `json:"coins"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return IconPayload{}, err
	}

	target := strings.ToUpper(strings.TrimSpace(symbol))
	for _, coin := range payload.Coins {
		if strings.EqualFold(coin.Symbol, target) {
			iconURL := firstNonEmpty(coin.Large, coin.Thumb)
			if strings.TrimSpace(iconURL) == "" {
				return IconPayload{}, ErrIconNotFound
			}
			return r.fetchRemoteImage(ctx, iconURL)
		}
	}

	return IconPayload{}, ErrIconNotFound
}

func (r *IconResolver) fetchUSStockIcon(ctx context.Context, symbol string) (IconPayload, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return IconPayload{}, ErrIconNotFound
	}

	return r.fetchRemoteImage(ctx, "https://financialmodelingprep.com/image-stock/"+url.PathEscape(symbol)+".png")
}

func (r *IconResolver) fetchRemoteImage(ctx context.Context, remoteURL string) (IconPayload, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, remoteURL, nil)
	if err != nil {
		return IconPayload{}, err
	}
	req.Header.Set("Accept", "image/*")
	req.Header.Set("User-Agent", "finance-dashboard/1.0")

	resp, err := r.client.Do(req)
	if err != nil {
		return IconPayload{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return IconPayload{}, fmt.Errorf("icon source returned status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		return IconPayload{}, fmt.Errorf("unexpected icon content type %q", contentType)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return IconPayload{}, err
	}

	return IconPayload{
		ContentType: contentType,
		Body:        body,
	}, nil
}
