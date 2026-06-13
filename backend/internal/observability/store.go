package observability

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/thiago/finance/backend/internal/domain"
)

type DBStore interface {
	CreateUser(ctx context.Context, name, email, passwordHash string) (domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	UpdateUserName(ctx context.Context, userID uuid.UUID, name string) error
	CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	ConsumePasswordResetToken(ctx context.Context, tokenHash string) (uuid.UUID, error)
	CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	RotateRefreshToken(ctx context.Context, tokenHash, newTokenHash string, expiresAt time.Time) (domain.User, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	ResetLoginFailures(ctx context.Context, userID uuid.UUID) error
	RecordFailedLogin(ctx context.Context, userID uuid.UUID, now time.Time, window, lockDuration time.Duration) (*time.Time, error)
	ListAssets(ctx context.Context, search string, limit int) ([]domain.Asset, error)
	GetAssetByID(ctx context.Context, id uuid.UUID) (domain.Asset, error)
	CreateAsset(ctx context.Context, asset domain.Asset) (domain.Asset, error)
	ListTransactions(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Transaction, error)
	ListTransactionsByAssetID(ctx context.Context, userID, assetID uuid.UUID, limit int) ([]domain.Transaction, error)
	CreateTransaction(ctx context.Context, txn domain.Transaction) (domain.Transaction, error)
	UpdateTransaction(ctx context.Context, txn domain.Transaction) (domain.Transaction, error)
	DeleteTransaction(ctx context.Context, userID, transactionID uuid.UUID) error
	ListDividends(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Dividend, error)
	CreateDividend(ctx context.Context, dividend domain.Dividend) (domain.Dividend, error)
	UpdateDividend(ctx context.Context, dividend domain.Dividend) (domain.Dividend, error)
	DeleteDividend(ctx context.Context, userID, dividendID uuid.UUID) error
	ListLatestPrices(ctx context.Context) ([]domain.AssetPrice, error)
	ListLatestPricesByAssetIDs(ctx context.Context, assetIDs []uuid.UUID) ([]domain.AssetPrice, error)
	ListPriceHistoryByAssetIDs(ctx context.Context, assetIDs []uuid.UUID, since time.Time) ([]domain.AssetPrice, error)
	UpsertAssetPrice(ctx context.Context, price domain.AssetPrice) error
	GetLatestExchangeRates(ctx context.Context) ([]domain.ExchangeRate, error)
	UpsertExchangeRate(ctx context.Context, rate domain.ExchangeRate) error
	ListAssetsByIDs(ctx context.Context, assetIDs []uuid.UUID) ([]domain.Asset, error)
}

type InstrumentedStore struct {
	next DBStore
}

func NewInstrumentedStore(next DBStore) *InstrumentedStore {
	return &InstrumentedStore{next: next}
}

func (s *InstrumentedStore) CreateUser(ctx context.Context, name, email, passwordHash string) (domain.User, error) {
	startedAt := time.Now()
	user, err := s.next.CreateUser(ctx, name, email, passwordHash)
	ObserveDBQuery("create_user", startedAt, err)
	return user, err
}

func (s *InstrumentedStore) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	startedAt := time.Now()
	user, err := s.next.GetUserByEmail(ctx, email)
	ObserveDBQuery("get_user_by_email", startedAt, err)
	return user, err
}

func (s *InstrumentedStore) UpdateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	startedAt := time.Now()
	err := s.next.UpdateUserPassword(ctx, userID, passwordHash)
	ObserveDBQuery("update_user_password", startedAt, err)
	return err
}

func (s *InstrumentedStore) UpdateUserName(ctx context.Context, userID uuid.UUID, name string) error {
	startedAt := time.Now()
	err := s.next.UpdateUserName(ctx, userID, name)
	ObserveDBQuery("update_user_name", startedAt, err)
	return err
}

func (s *InstrumentedStore) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	startedAt := time.Now()
	err := s.next.CreatePasswordResetToken(ctx, userID, tokenHash, expiresAt)
	ObserveDBQuery("create_password_reset_token", startedAt, err)
	return err
}

func (s *InstrumentedStore) ConsumePasswordResetToken(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	startedAt := time.Now()
	userID, err := s.next.ConsumePasswordResetToken(ctx, tokenHash)
	ObserveDBQuery("consume_password_reset_token", startedAt, err)
	return userID, err
}

func (s *InstrumentedStore) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	startedAt := time.Now()
	err := s.next.CreateRefreshToken(ctx, userID, tokenHash, expiresAt)
	ObserveDBQuery("create_refresh_token", startedAt, err)
	return err
}

func (s *InstrumentedStore) RotateRefreshToken(ctx context.Context, tokenHash, newTokenHash string, expiresAt time.Time) (domain.User, error) {
	startedAt := time.Now()
	user, err := s.next.RotateRefreshToken(ctx, tokenHash, newTokenHash, expiresAt)
	ObserveDBQuery("rotate_refresh_token", startedAt, err)
	return user, err
}

func (s *InstrumentedStore) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	startedAt := time.Now()
	err := s.next.RevokeRefreshToken(ctx, tokenHash)
	ObserveDBQuery("revoke_refresh_token", startedAt, err)
	return err
}

func (s *InstrumentedStore) ResetLoginFailures(ctx context.Context, userID uuid.UUID) error {
	startedAt := time.Now()
	err := s.next.ResetLoginFailures(ctx, userID)
	ObserveDBQuery("reset_login_failures", startedAt, err)
	return err
}

func (s *InstrumentedStore) RecordFailedLogin(ctx context.Context, userID uuid.UUID, now time.Time, window, lockDuration time.Duration) (*time.Time, error) {
	startedAt := time.Now()
	lockedUntil, err := s.next.RecordFailedLogin(ctx, userID, now, window, lockDuration)
	ObserveDBQuery("record_failed_login", startedAt, err)
	return lockedUntil, err
}

func (s *InstrumentedStore) ListAssets(ctx context.Context, search string, limit int) ([]domain.Asset, error) {
	startedAt := time.Now()
	assets, err := s.next.ListAssets(ctx, search, limit)
	ObserveDBQuery("list_assets", startedAt, err)
	return assets, err
}

func (s *InstrumentedStore) GetAssetByID(ctx context.Context, id uuid.UUID) (domain.Asset, error) {
	startedAt := time.Now()
	asset, err := s.next.GetAssetByID(ctx, id)
	ObserveDBQuery("get_asset_by_id", startedAt, err)
	return asset, err
}

func (s *InstrumentedStore) CreateAsset(ctx context.Context, asset domain.Asset) (domain.Asset, error) {
	startedAt := time.Now()
	created, err := s.next.CreateAsset(ctx, asset)
	ObserveDBQuery("create_asset", startedAt, err)
	return created, err
}

func (s *InstrumentedStore) ListTransactions(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Transaction, error) {
	startedAt := time.Now()
	items, err := s.next.ListTransactions(ctx, userID, limit)
	ObserveDBQuery("list_transactions", startedAt, err)
	return items, err
}

func (s *InstrumentedStore) ListTransactionsByAssetID(ctx context.Context, userID, assetID uuid.UUID, limit int) ([]domain.Transaction, error) {
	startedAt := time.Now()
	items, err := s.next.ListTransactionsByAssetID(ctx, userID, assetID, limit)
	ObserveDBQuery("list_transactions_by_asset_id", startedAt, err)
	return items, err
}

func (s *InstrumentedStore) CreateTransaction(ctx context.Context, txn domain.Transaction) (domain.Transaction, error) {
	startedAt := time.Now()
	item, err := s.next.CreateTransaction(ctx, txn)
	ObserveDBQuery("create_transaction", startedAt, err)
	return item, err
}

func (s *InstrumentedStore) UpdateTransaction(ctx context.Context, txn domain.Transaction) (domain.Transaction, error) {
	startedAt := time.Now()
	item, err := s.next.UpdateTransaction(ctx, txn)
	ObserveDBQuery("update_transaction", startedAt, err)
	return item, err
}

func (s *InstrumentedStore) DeleteTransaction(ctx context.Context, userID, transactionID uuid.UUID) error {
	startedAt := time.Now()
	err := s.next.DeleteTransaction(ctx, userID, transactionID)
	ObserveDBQuery("delete_transaction", startedAt, err)
	return err
}

func (s *InstrumentedStore) ListDividends(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Dividend, error) {
	startedAt := time.Now()
	items, err := s.next.ListDividends(ctx, userID, limit)
	ObserveDBQuery("list_dividends", startedAt, err)
	return items, err
}

func (s *InstrumentedStore) CreateDividend(ctx context.Context, dividend domain.Dividend) (domain.Dividend, error) {
	startedAt := time.Now()
	item, err := s.next.CreateDividend(ctx, dividend)
	ObserveDBQuery("create_dividend", startedAt, err)
	return item, err
}

func (s *InstrumentedStore) UpdateDividend(ctx context.Context, dividend domain.Dividend) (domain.Dividend, error) {
	startedAt := time.Now()
	item, err := s.next.UpdateDividend(ctx, dividend)
	ObserveDBQuery("update_dividend", startedAt, err)
	return item, err
}

func (s *InstrumentedStore) DeleteDividend(ctx context.Context, userID, dividendID uuid.UUID) error {
	startedAt := time.Now()
	err := s.next.DeleteDividend(ctx, userID, dividendID)
	ObserveDBQuery("delete_dividend", startedAt, err)
	return err
}

func (s *InstrumentedStore) ListLatestPrices(ctx context.Context) ([]domain.AssetPrice, error) {
	startedAt := time.Now()
	items, err := s.next.ListLatestPrices(ctx)
	ObserveDBQuery("list_latest_prices", startedAt, err)
	return items, err
}

func (s *InstrumentedStore) ListLatestPricesByAssetIDs(ctx context.Context, assetIDs []uuid.UUID) ([]domain.AssetPrice, error) {
	startedAt := time.Now()
	items, err := s.next.ListLatestPricesByAssetIDs(ctx, assetIDs)
	ObserveDBQuery("list_latest_prices_by_asset_ids", startedAt, err)
	return items, err
}

func (s *InstrumentedStore) ListPriceHistoryByAssetIDs(ctx context.Context, assetIDs []uuid.UUID, since time.Time) ([]domain.AssetPrice, error) {
	startedAt := time.Now()
	items, err := s.next.ListPriceHistoryByAssetIDs(ctx, assetIDs, since)
	ObserveDBQuery("list_price_history_by_asset_ids", startedAt, err)
	return items, err
}

func (s *InstrumentedStore) UpsertAssetPrice(ctx context.Context, price domain.AssetPrice) error {
	startedAt := time.Now()
	err := s.next.UpsertAssetPrice(ctx, price)
	ObserveDBQuery("upsert_asset_price", startedAt, err)
	return err
}

func (s *InstrumentedStore) GetLatestExchangeRates(ctx context.Context) ([]domain.ExchangeRate, error) {
	startedAt := time.Now()
	items, err := s.next.GetLatestExchangeRates(ctx)
	ObserveDBQuery("get_latest_exchange_rates", startedAt, err)
	return items, err
}

func (s *InstrumentedStore) UpsertExchangeRate(ctx context.Context, rate domain.ExchangeRate) error {
	startedAt := time.Now()
	err := s.next.UpsertExchangeRate(ctx, rate)
	ObserveDBQuery("upsert_exchange_rate", startedAt, err)
	return err
}

func (s *InstrumentedStore) ListAssetsByIDs(ctx context.Context, assetIDs []uuid.UUID) ([]domain.Asset, error) {
	startedAt := time.Now()
	assets, err := s.next.ListAssetsByIDs(ctx, assetIDs)
	ObserveDBQuery("list_assets_by_ids", startedAt, err)
	return assets, err
}
