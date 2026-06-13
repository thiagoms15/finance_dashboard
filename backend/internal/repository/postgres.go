package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/thiago/finance/backend/internal/domain"
)

var ErrNotFound = errors.New("not found")

type PostgresStore struct {
	db *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &PostgresStore{db: pool}, nil
}

func (s *PostgresStore) Close() {
	s.db.Close()
}

func (s *PostgresStore) CreateUser(ctx context.Context, name, email, passwordHash string) (domain.User, error) {
	row := s.db.QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, password_hash, failed_login_attempts, first_failed_login_at, locked_until, created_at, updated_at
	`, strings.TrimSpace(name), strings.ToLower(strings.TrimSpace(email)), passwordHash)

	return scanUser(row)
}

func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, name, email, password_hash, failed_login_attempts, first_failed_login_at, locked_until, created_at, updated_at
		FROM users
		WHERE email = $1
	`, strings.ToLower(strings.TrimSpace(email)))

	return scanUser(row)
}

func (s *PostgresStore) UpdateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE users
		SET password_hash = $2
		WHERE id = $1
	`, userID, passwordHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) UpdateUserName(ctx context.Context, userID uuid.UUID, name string) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE users
		SET name = $2
		WHERE id = $1
	`, userID, strings.TrimSpace(name))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (s *PostgresStore) ConsumePasswordResetToken(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	var userID uuid.UUID
	row := tx.QueryRow(ctx, `
		SELECT user_id
		FROM password_reset_tokens
		WHERE token_hash = $1
		  AND used_at IS NULL
		  AND expires_at > now()
		FOR UPDATE
	`, tokenHash)
	if err := row.Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrNotFound
		}
		return uuid.Nil, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE password_reset_tokens
		SET used_at = now()
		WHERE token_hash = $1
	`, tokenHash)
	if err != nil {
		return uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func (s *PostgresStore) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (s *PostgresStore) RotateRefreshToken(ctx context.Context, tokenHash, newTokenHash string, expiresAt time.Time) (domain.User, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.User{}, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		SELECT u.id, u.name, u.email, u.password_hash, u.failed_login_attempts, u.first_failed_login_at, u.locked_until, u.created_at, u.updated_at
		FROM refresh_tokens rt
		JOIN users u ON u.id = rt.user_id
		WHERE rt.token_hash = $1
		  AND rt.revoked_at IS NULL
		  AND rt.expires_at > now()
		FOR UPDATE
	`, tokenHash)
	user, err := scanUser(row)
	if err != nil {
		return domain.User{}, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = now(),
		    replaced_by_token_hash = $2
		WHERE token_hash = $1
	`, tokenHash, newTokenHash); err != nil {
		return domain.User{}, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, user.ID, newTokenHash, expiresAt); err != nil {
		return domain.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (s *PostgresStore) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > now()
	`, tokenHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) ResetLoginFailures(ctx context.Context, userID uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE users
		SET failed_login_attempts = 0,
		    first_failed_login_at = NULL,
		    locked_until = NULL
		WHERE id = $1
	`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) RecordFailedLogin(ctx context.Context, userID uuid.UUID, now time.Time, window, lockDuration time.Duration) (*time.Time, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var (
		attempts    int
		firstFailed *time.Time
		lockedUntil *time.Time
	)
	if err := tx.QueryRow(ctx, `
		SELECT failed_login_attempts, first_failed_login_at, locked_until
		FROM users
		WHERE id = $1
		FOR UPDATE
	`, userID).Scan(&attempts, &firstFailed, &lockedUntil); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if lockedUntil != nil && lockedUntil.After(now) {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return lockedUntil, nil
	}

	if firstFailed == nil || firstFailed.Add(window).Before(now) {
		attempts = 1
		firstFailed = &now
	} else {
		attempts++
	}

	var newLockedUntil *time.Time
	if attempts >= 5 {
		lockUntil := now.Add(lockDuration)
		newLockedUntil = &lockUntil
	}

	if _, err := tx.Exec(ctx, `
		UPDATE users
		SET failed_login_attempts = $2,
		    first_failed_login_at = $3,
		    locked_until = $4
		WHERE id = $1
	`, userID, attempts, firstFailed, newLockedUntil); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return newLockedUntil, nil
}

func (s *PostgresStore) ListAssets(ctx context.Context, search string, limit int) ([]domain.Asset, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	query := `
		SELECT id, symbol, name, exchange, currency, sector, created_at, updated_at
		FROM assets
	`
	args := []any{}
	if search = strings.TrimSpace(search); search != "" {
		query += ` WHERE symbol ILIKE $1 OR name ILIKE $1 OR exchange ILIKE $1 `
		args = append(args, "%"+search+"%")
	}
	query += fmt.Sprintf(" ORDER BY symbol ASC LIMIT %d", limit)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assets := make([]domain.Asset, 0, limit)
	for rows.Next() {
		asset, err := scanAsset(rows)
		if err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}
	return assets, rows.Err()
}

func (s *PostgresStore) GetAssetByID(ctx context.Context, id uuid.UUID) (domain.Asset, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, symbol, name, exchange, currency, sector, created_at, updated_at
		FROM assets
		WHERE id = $1
	`, id)
	return scanAsset(row)
}

func (s *PostgresStore) CreateAsset(ctx context.Context, asset domain.Asset) (domain.Asset, error) {
	row := s.db.QueryRow(ctx, `
		INSERT INTO assets (symbol, name, exchange, currency, sector)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (symbol, exchange)
		DO UPDATE SET
			name = EXCLUDED.name,
			currency = EXCLUDED.currency,
			sector = EXCLUDED.sector
		RETURNING id, symbol, name, exchange, currency, sector, created_at, updated_at
	`, strings.ToUpper(asset.Symbol), asset.Name, strings.ToUpper(asset.Exchange), strings.ToUpper(asset.Currency), asset.Sector)
	return scanAsset(row)
}

func (s *PostgresStore) ListTransactions(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Transaction, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, asset_id, type, quantity, price, fees, currency, transaction_date, notes, created_at, updated_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY transaction_date DESC, created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []domain.Transaction
	for rows.Next() {
		txn, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		txns = append(txns, txn)
	}
	return txns, rows.Err()
}

func (s *PostgresStore) ListTransactionsByAssetID(ctx context.Context, userID, assetID uuid.UUID, limit int) ([]domain.Transaction, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, asset_id, type, quantity, price, fees, currency, transaction_date, notes, created_at, updated_at
		FROM transactions
		WHERE user_id = $1 AND asset_id = $2
		ORDER BY transaction_date DESC, created_at DESC
		LIMIT $3
	`, userID, assetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []domain.Transaction
	for rows.Next() {
		txn, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		txns = append(txns, txn)
	}
	return txns, rows.Err()
}

func (s *PostgresStore) CreateTransaction(ctx context.Context, txn domain.Transaction) (domain.Transaction, error) {
	row := s.db.QueryRow(ctx, `
		INSERT INTO transactions (user_id, asset_id, type, quantity, price, fees, currency, transaction_date, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, asset_id, type, quantity, price, fees, currency, transaction_date, notes, created_at, updated_at
	`, txn.UserID, txn.AssetID, txn.Type, txn.Quantity.String(), txn.Price.String(), txn.Fees.String(), strings.ToUpper(txn.Currency), txn.TransactionDate, txn.Notes)
	return scanTransaction(row)
}

func (s *PostgresStore) UpdateTransaction(ctx context.Context, txn domain.Transaction) (domain.Transaction, error) {
	row := s.db.QueryRow(ctx, `
		UPDATE transactions
		SET asset_id = $3,
		    type = $4,
		    quantity = $5,
		    price = $6,
		    fees = $7,
		    currency = $8,
		    transaction_date = $9,
		    notes = $10
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, asset_id, type, quantity, price, fees, currency, transaction_date, notes, created_at, updated_at
	`, txn.ID, txn.UserID, txn.AssetID, txn.Type, txn.Quantity.String(), txn.Price.String(), txn.Fees.String(), strings.ToUpper(txn.Currency), txn.TransactionDate, txn.Notes)
	return scanTransaction(row)
}

func (s *PostgresStore) DeleteTransaction(ctx context.Context, userID, transactionID uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM transactions WHERE id = $1 AND user_id = $2`, transactionID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) ListDividends(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Dividend, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, asset_id, amount, currency, payment_date, created_at, updated_at
		FROM dividends
		WHERE user_id = $1
		ORDER BY payment_date DESC, created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dividends []domain.Dividend
	for rows.Next() {
		dividend, err := scanDividend(rows)
		if err != nil {
			return nil, err
		}
		dividends = append(dividends, dividend)
	}
	return dividends, rows.Err()
}

func (s *PostgresStore) CreateDividend(ctx context.Context, dividend domain.Dividend) (domain.Dividend, error) {
	row := s.db.QueryRow(ctx, `
		INSERT INTO dividends (user_id, asset_id, amount, currency, payment_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, asset_id, amount, currency, payment_date, created_at, updated_at
	`, dividend.UserID, dividend.AssetID, dividend.Amount.String(), strings.ToUpper(dividend.Currency), dividend.PaymentDate)
	return scanDividend(row)
}

func (s *PostgresStore) UpdateDividend(ctx context.Context, dividend domain.Dividend) (domain.Dividend, error) {
	row := s.db.QueryRow(ctx, `
		UPDATE dividends
		SET asset_id = $3,
		    amount = $4,
		    currency = $5,
		    payment_date = $6
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, asset_id, amount, currency, payment_date, created_at, updated_at
	`, dividend.ID, dividend.UserID, dividend.AssetID, dividend.Amount.String(), strings.ToUpper(dividend.Currency), dividend.PaymentDate)
	return scanDividend(row)
}

func (s *PostgresStore) DeleteDividend(ctx context.Context, userID, dividendID uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM dividends WHERE id = $1 AND user_id = $2`, dividendID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) ListLatestPrices(ctx context.Context) ([]domain.AssetPrice, error) {
	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT ON (asset_id)
			id, asset_id, price, previous_close, currency, timestamp
		FROM asset_prices
		ORDER BY asset_id, timestamp DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []domain.AssetPrice
	for rows.Next() {
		price, err := scanAssetPrice(rows)
		if err != nil {
			return nil, err
		}
		prices = append(prices, price)
	}
	return prices, rows.Err()
}

func (s *PostgresStore) ListLatestPricesByAssetIDs(ctx context.Context, assetIDs []uuid.UUID) ([]domain.AssetPrice, error) {
	if len(assetIDs) == 0 {
		return []domain.AssetPrice{}, nil
	}

	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT ON (asset_id)
			id, asset_id, price, previous_close, currency, timestamp
		FROM asset_prices
		WHERE asset_id = ANY($1::uuid[])
		ORDER BY asset_id, timestamp DESC
	`, assetIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []domain.AssetPrice
	for rows.Next() {
		price, err := scanAssetPrice(rows)
		if err != nil {
			return nil, err
		}
		prices = append(prices, price)
	}
	return prices, rows.Err()
}

func (s *PostgresStore) ListPriceHistoryByAssetIDs(ctx context.Context, assetIDs []uuid.UUID, since time.Time) ([]domain.AssetPrice, error) {
	if len(assetIDs) == 0 {
		return []domain.AssetPrice{}, nil
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, asset_id, price, previous_close, currency, timestamp
		FROM asset_prices
		WHERE asset_id = ANY($1::uuid[])
		  AND timestamp >= $2
		ORDER BY asset_id ASC, timestamp ASC
	`, assetIDs, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prices := make([]domain.AssetPrice, 0, len(assetIDs)*32)
	for rows.Next() {
		price, err := scanAssetPrice(rows)
		if err != nil {
			return nil, err
		}
		prices = append(prices, price)
	}
	return prices, rows.Err()
}

func (s *PostgresStore) UpsertAssetPrice(ctx context.Context, price domain.AssetPrice) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO asset_prices (asset_id, price, previous_close, currency, timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`, price.AssetID, price.Price.String(), price.PreviousClose.String(), strings.ToUpper(price.Currency), price.Timestamp)
	return err
}

func (s *PostgresStore) GetLatestExchangeRates(ctx context.Context) ([]domain.ExchangeRate, error) {
	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT ON (base, quote)
			id, base, quote, rate, timestamp
		FROM exchange_rates
		ORDER BY base, quote, timestamp DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []domain.ExchangeRate
	for rows.Next() {
		rate, err := scanExchangeRate(rows)
		if err != nil {
			return nil, err
		}
		rates = append(rates, rate)
	}
	return rates, rows.Err()
}

func (s *PostgresStore) UpsertExchangeRate(ctx context.Context, rate domain.ExchangeRate) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO exchange_rates (base, quote, rate, timestamp)
		VALUES ($1, $2, $3, $4)
	`, strings.ToUpper(rate.Base), strings.ToUpper(rate.Quote), rate.Rate.String(), rate.Timestamp)
	return err
}

func (s *PostgresStore) ListAssetsByIDs(ctx context.Context, assetIDs []uuid.UUID) ([]domain.Asset, error) {
	if len(assetIDs) == 0 {
		return []domain.Asset{}, nil
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, symbol, name, exchange, currency, sector, created_at, updated_at
		FROM assets
		WHERE id = ANY($1::uuid[])
		ORDER BY symbol ASC
	`, assetIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assets := make([]domain.Asset, 0, len(assetIDs))
	for rows.Next() {
		asset, err := scanAsset(rows)
		if err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}
	return assets, rows.Err()
}

func scanUser(row interface{ Scan(dest ...any) error }) (domain.User, error) {
	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.FailedLoginAttempts,
		&user.FirstFailedLoginAt,
		&user.LockedUntil,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, err
	}
	return user, nil
}

func scanAsset(row interface{ Scan(dest ...any) error }) (domain.Asset, error) {
	var asset domain.Asset
	err := row.Scan(&asset.ID, &asset.Symbol, &asset.Name, &asset.Exchange, &asset.Currency, &asset.Sector, &asset.CreatedAt, &asset.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Asset{}, ErrNotFound
		}
		return domain.Asset{}, err
	}
	return asset, nil
}

func scanTransaction(row interface{ Scan(dest ...any) error }) (domain.Transaction, error) {
	var txn domain.Transaction
	var qty, price, fees string
	err := row.Scan(&txn.ID, &txn.UserID, &txn.AssetID, &txn.Type, &qty, &price, &fees, &txn.Currency, &txn.TransactionDate, &txn.Notes, &txn.CreatedAt, &txn.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Transaction{}, ErrNotFound
		}
		return domain.Transaction{}, err
	}
	txn.Quantity, _ = decimal.NewFromString(qty)
	txn.Price, _ = decimal.NewFromString(price)
	txn.Fees, _ = decimal.NewFromString(fees)
	return txn, nil
}

func scanDividend(row interface{ Scan(dest ...any) error }) (domain.Dividend, error) {
	var dividend domain.Dividend
	var amount string
	err := row.Scan(&dividend.ID, &dividend.UserID, &dividend.AssetID, &amount, &dividend.Currency, &dividend.PaymentDate, &dividend.CreatedAt, &dividend.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Dividend{}, ErrNotFound
		}
		return domain.Dividend{}, err
	}
	dividend.Amount, _ = decimal.NewFromString(amount)
	return dividend, nil
}

func scanAssetPrice(row interface{ Scan(dest ...any) error }) (domain.AssetPrice, error) {
	var price domain.AssetPrice
	var value, previous string
	err := row.Scan(&price.ID, &price.AssetID, &value, &previous, &price.Currency, &price.Timestamp)
	if err != nil {
		return domain.AssetPrice{}, err
	}
	price.Price, _ = decimal.NewFromString(value)
	price.PreviousClose, _ = decimal.NewFromString(previous)
	return price, nil
}

func scanExchangeRate(row interface{ Scan(dest ...any) error }) (domain.ExchangeRate, error) {
	var rate domain.ExchangeRate
	var value string
	err := row.Scan(&rate.ID, &rate.Base, &rate.Quote, &value, &rate.Timestamp)
	if err != nil {
		return domain.ExchangeRate{}, err
	}
	rate.Rate, _ = decimal.NewFromString(value)
	return rate, nil
}
