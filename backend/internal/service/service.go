package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/thiago/finance/backend/internal/auth"
	"github.com/thiago/finance/backend/internal/domain"
	"github.com/thiago/finance/backend/internal/domain/portfolio"
	"github.com/thiago/finance/backend/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidSession     = errors.New("invalid session")
)

const (
	loginFailureWindow = 15 * time.Minute
	lockoutDuration    = 15 * time.Minute
)

type AccountLockedError struct {
	LockedUntil time.Time
}

func (e AccountLockedError) Error() string {
	return fmt.Sprintf("account locked until %s", e.LockedUntil.UTC().Format(time.RFC3339))
}

type Store interface {
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
	CreateTransaction(ctx context.Context, txn domain.Transaction) (domain.Transaction, error)
	UpdateTransaction(ctx context.Context, txn domain.Transaction) (domain.Transaction, error)
	DeleteTransaction(ctx context.Context, userID, transactionID uuid.UUID) error
	ListDividends(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Dividend, error)
	CreateDividend(ctx context.Context, dividend domain.Dividend) (domain.Dividend, error)
	UpdateDividend(ctx context.Context, dividend domain.Dividend) (domain.Dividend, error)
	DeleteDividend(ctx context.Context, userID, dividendID uuid.UUID) error
	ListLatestPrices(ctx context.Context) ([]domain.AssetPrice, error)
	UpsertAssetPrice(ctx context.Context, price domain.AssetPrice) error
	GetLatestExchangeRates(ctx context.Context) ([]domain.ExchangeRate, error)
	UpsertExchangeRate(ctx context.Context, rate domain.ExchangeRate) error
}

type AppService struct {
	store        Store
	tokenManager *auth.TokenManager
	hasher       auth.PasswordHasher
	accessTTL    time.Duration
	refreshTTL   time.Duration
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginOutput struct {
	AccessToken       string      `json:"accessToken"`
	TokenType         string      `json:"tokenType"`
	ExpiresIn         int64       `json:"expiresIn"`
	User              domain.User `json:"user"`
	RefreshToken      string      `json:"-"`
	RefreshExpiresAt  time.Time   `json:"-"`
}

type TransactionInput struct {
	AssetID         uuid.UUID
	Type            domain.TransactionType
	Quantity        decimal.Decimal
	Price           decimal.Decimal
	Fees            decimal.Decimal
	Currency        string
	TransactionDate time.Time
	Notes           string
}

type DividendInput struct {
	AssetID     uuid.UUID
	Amount      decimal.Decimal
	Currency    string
	PaymentDate time.Time
}

type AssetInput struct {
	Symbol   string
	Name     string
	Exchange string
	Currency string
	Sector   string
}

func New(store Store, tokenManager *auth.TokenManager, hasher auth.PasswordHasher, accessTTL, refreshTTL time.Duration) *AppService {
	return &AppService{
		store:        store,
		tokenManager: tokenManager,
		hasher:       hasher,
		accessTTL:    accessTTL,
		refreshTTL:   refreshTTL,
	}
}

func (s *AppService) Register(ctx context.Context, input RegisterInput) (LoginOutput, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if name == "" || !strings.Contains(email, "@") || len(input.Password) < 8 {
		return LoginOutput{}, ErrInvalidInput
	}

	hash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return LoginOutput{}, err
	}

	user, err := s.store.CreateUser(ctx, name, email, hash)
	if err != nil {
		return LoginOutput{}, err
	}

	return s.createLoginOutput(ctx, user)
}

func (s *AppService) Login(ctx context.Context, email, password string) (LoginOutput, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return LoginOutput{}, ErrInvalidCredentials
		}
		return LoginOutput{}, err
	}

	if user.LockedUntil != nil && user.LockedUntil.After(time.Now().UTC()) {
		return LoginOutput{}, AccountLockedError{LockedUntil: *user.LockedUntil}
	}

	ok, err := s.hasher.Verify(password, user.PasswordHash)
	if err != nil {
		return LoginOutput{}, err
	}
	if !ok {
		lockedUntil, err := s.store.RecordFailedLogin(ctx, user.ID, time.Now().UTC(), loginFailureWindow, lockoutDuration)
		if err != nil {
			return LoginOutput{}, err
		}
		if lockedUntil != nil {
			return LoginOutput{}, AccountLockedError{LockedUntil: *lockedUntil}
		}
		return LoginOutput{}, ErrInvalidCredentials
	}

	if err := s.store.ResetLoginFailures(ctx, user.ID); err != nil {
		return LoginOutput{}, err
	}

	return s.createLoginOutput(ctx, user)
}

func (s *AppService) RefreshSession(ctx context.Context, refreshToken string) (LoginOutput, error) {
	_, tokenHash, err := auth.GenerateRefreshTokenFromRaw(refreshToken)
	if err != nil {
		return LoginOutput{}, ErrInvalidSession
	}

	rawToken, newTokenHash, err := auth.GenerateRefreshToken()
	if err != nil {
		return LoginOutput{}, err
	}
	refreshExpiresAt := time.Now().UTC().Add(s.refreshTTL)
	user, err := s.store.RotateRefreshToken(ctx, tokenHash, newTokenHash, refreshExpiresAt)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return LoginOutput{}, ErrInvalidSession
		}
		return LoginOutput{}, err
	}

	return s.createAccessOutput(user, rawToken, refreshExpiresAt)
}

func (s *AppService) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return nil
	}

	_, tokenHash, err := auth.GenerateRefreshTokenFromRaw(refreshToken)
	if err != nil {
		return nil
	}
	if err := s.store.RevokeRefreshToken(ctx, tokenHash); err != nil && !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	return nil
}

func (s *AppService) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", nil
		}
		return "", err
	}

	raw, hash, err := auth.GeneratePasswordResetToken()
	if err != nil {
		return "", err
	}

	if err := s.store.CreatePasswordResetToken(ctx, user.ID, hash, time.Now().UTC().Add(time.Hour)); err != nil {
		return "", err
	}

	return raw, nil
}

func (s *AppService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrInvalidInput
	}

	hash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}

	_, tokenHash, err := auth.GeneratePasswordResetTokenFromRaw(token)
	if err != nil {
		return err
	}

	userID, err := s.store.ConsumePasswordResetToken(ctx, tokenHash)
	if err != nil {
		return err
	}

	return s.store.UpdateUserPassword(ctx, userID, hash)
}

func (s *AppService) ListAssets(ctx context.Context, search string, limit int) ([]domain.Asset, error) {
	return s.store.ListAssets(ctx, search, limit)
}

func (s *AppService) GetAsset(ctx context.Context, assetID uuid.UUID) (domain.Asset, error) {
	return s.store.GetAssetByID(ctx, assetID)
}

func (s *AppService) CreateAsset(ctx context.Context, input AssetInput) (domain.Asset, error) {
	symbol := strings.ToUpper(strings.TrimSpace(input.Symbol))
	exchange := strings.ToUpper(strings.TrimSpace(input.Exchange))
	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	name := strings.TrimSpace(input.Name)
	sector := strings.TrimSpace(input.Sector)

	if symbol == "" || exchange == "" || currency == "" {
		return domain.Asset{}, ErrInvalidInput
	}
	if name == "" {
		name = symbol
	}

	return s.store.CreateAsset(ctx, domain.Asset{
		Symbol:   symbol,
		Name:     name,
		Exchange: exchange,
		Currency: currency,
		Sector:   sector,
	})
}

func (s *AppService) CreateTransaction(ctx context.Context, userID uuid.UUID, input TransactionInput) (domain.Transaction, error) {
	if err := validateTransactionInput(input); err != nil {
		return domain.Transaction{}, err
	}

	if _, err := s.store.GetAssetByID(ctx, input.AssetID); err != nil {
		return domain.Transaction{}, err
	}

	return s.store.CreateTransaction(ctx, domain.Transaction{
		UserID:          userID,
		AssetID:         input.AssetID,
		Type:            input.Type,
		Quantity:        input.Quantity,
		Price:           input.Price,
		Fees:            input.Fees,
		Currency:        strings.ToUpper(input.Currency),
		TransactionDate: input.TransactionDate.UTC(),
		Notes:           strings.TrimSpace(input.Notes),
	})
}

func (s *AppService) UpdateTransaction(ctx context.Context, userID, transactionID uuid.UUID, input TransactionInput) (domain.Transaction, error) {
	if err := validateTransactionInput(input); err != nil {
		return domain.Transaction{}, err
	}

	return s.store.UpdateTransaction(ctx, domain.Transaction{
		ID:              transactionID,
		UserID:          userID,
		AssetID:         input.AssetID,
		Type:            input.Type,
		Quantity:        input.Quantity,
		Price:           input.Price,
		Fees:            input.Fees,
		Currency:        strings.ToUpper(input.Currency),
		TransactionDate: input.TransactionDate.UTC(),
		Notes:           strings.TrimSpace(input.Notes),
	})
}

func (s *AppService) DeleteTransaction(ctx context.Context, userID, transactionID uuid.UUID) error {
	return s.store.DeleteTransaction(ctx, userID, transactionID)
}

func (s *AppService) ListTransactions(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Transaction, error) {
	return s.store.ListTransactions(ctx, userID, limit)
}

func (s *AppService) CreateDividend(ctx context.Context, userID uuid.UUID, input DividendInput) (domain.Dividend, error) {
	if input.Amount.LessThanOrEqual(decimal.Zero) || strings.TrimSpace(input.Currency) == "" {
		return domain.Dividend{}, ErrInvalidInput
	}
	if _, err := s.store.GetAssetByID(ctx, input.AssetID); err != nil {
		return domain.Dividend{}, err
	}

	return s.store.CreateDividend(ctx, domain.Dividend{
		UserID:      userID,
		AssetID:     input.AssetID,
		Amount:      input.Amount,
		Currency:    strings.ToUpper(input.Currency),
		PaymentDate: input.PaymentDate.UTC(),
	})
}

func (s *AppService) UpdateDividend(ctx context.Context, userID, dividendID uuid.UUID, input DividendInput) (domain.Dividend, error) {
	if input.Amount.LessThanOrEqual(decimal.Zero) || strings.TrimSpace(input.Currency) == "" {
		return domain.Dividend{}, ErrInvalidInput
	}

	return s.store.UpdateDividend(ctx, domain.Dividend{
		ID:          dividendID,
		UserID:      userID,
		AssetID:     input.AssetID,
		Amount:      input.Amount,
		Currency:    strings.ToUpper(input.Currency),
		PaymentDate: input.PaymentDate.UTC(),
	})
}

func (s *AppService) DeleteDividend(ctx context.Context, userID, dividendID uuid.UUID) error {
	return s.store.DeleteDividend(ctx, userID, dividendID)
}

func (s *AppService) ListDividends(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Dividend, error) {
	return s.store.ListDividends(ctx, userID, limit)
}

func (s *AppService) PortfolioSnapshot(ctx context.Context, userID uuid.UUID, preferredCurrency string) (portfolio.Snapshot, error) {
	txns, err := s.store.ListTransactions(ctx, userID, 1000)
	if err != nil {
		return portfolio.Snapshot{}, err
	}
	dividends, err := s.store.ListDividends(ctx, userID, 1000)
	if err != nil {
		return portfolio.Snapshot{}, err
	}
	prices, err := s.store.ListLatestPrices(ctx)
	if err != nil {
		return portfolio.Snapshot{}, err
	}

	assetsMap, err := s.assetsMap(ctx)
	if err != nil {
		return portfolio.Snapshot{}, err
	}

	rates, err := s.exchangeRateMap(ctx)
	if err != nil {
		return portfolio.Snapshot{}, err
	}

	return portfolio.Calculate(assetsMap, txns, dividends, prices, strings.ToUpper(preferredCurrency), converter(rates, preferredCurrency))
}

func (s *AppService) PortfolioPerformance(ctx context.Context, userID uuid.UUID, preferredCurrency string) ([]portfolio.PerformancePoint, error) {
	txns, err := s.store.ListTransactions(ctx, userID, 1000)
	if err != nil {
		return nil, err
	}
	rates, err := s.exchangeRateMap(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(txns, func(i, j int) bool {
		return txns[i].TransactionDate.Before(txns[j].TransactionDate)
	})
	return portfolio.BuildPerformanceSeries(txns, strings.ToUpper(preferredCurrency), converter(rates, preferredCurrency)), nil
}

func (s *AppService) createLoginOutput(ctx context.Context, user domain.User) (LoginOutput, error) {
	refreshToken, refreshTokenHash, err := auth.GenerateRefreshToken()
	if err != nil {
		return LoginOutput{}, err
	}
	refreshExpiresAt := time.Now().UTC().Add(s.refreshTTL)
	if err := s.store.CreateRefreshToken(ctx, user.ID, refreshTokenHash, refreshExpiresAt); err != nil {
		return LoginOutput{}, err
	}

	return s.createAccessOutput(user, refreshToken, refreshExpiresAt)
}

func (s *AppService) createAccessOutput(user domain.User, refreshToken string, refreshExpiresAt time.Time) (LoginOutput, error) {
	token, err := s.tokenManager.CreateAccessToken(user.ID, user.Email, s.accessTTL)
	if err != nil {
		return LoginOutput{}, err
	}

	user.PasswordHash = ""
	return LoginOutput{
		AccessToken:      token,
		TokenType:        "Bearer",
		ExpiresIn:        int64(s.accessTTL.Seconds()),
		User:             user,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func (s *AppService) assetsMap(ctx context.Context) (map[uuid.UUID]domain.Asset, error) {
	assets, err := s.store.ListAssets(ctx, "", 1000)
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]domain.Asset, len(assets))
	for _, asset := range assets {
		out[asset.ID] = asset
	}
	return out, nil
}

func (s *AppService) exchangeRateMap(ctx context.Context) (map[string]decimal.Decimal, error) {
	rates, err := s.store.GetLatestExchangeRates(ctx)
	if err != nil {
		return nil, err
	}
	out := map[string]decimal.Decimal{}
	for _, rate := range rates {
		out[rate.Base+"_"+rate.Quote] = rate.Rate
	}
	return out, nil
}

func converter(rates map[string]decimal.Decimal, preferredCurrency string) func(amount decimal.Decimal, from, to string) decimal.Decimal {
	return func(amount decimal.Decimal, from, to string) decimal.Decimal {
		from = strings.ToUpper(from)
		to = strings.ToUpper(to)
		if from == to {
			return amount
		}

		key := from + "_" + to
		rate, ok := rates[key]
		if !ok {
			return amount
		}
		return amount.Mul(rate)
	}
}

func validateTransactionInput(input TransactionInput) error {
	if input.Quantity.LessThanOrEqual(decimal.Zero) || input.Price.LessThan(decimal.Zero) || input.Fees.LessThan(decimal.Zero) {
		return ErrInvalidInput
	}
	switch input.Type {
	case domain.TransactionTypeBuy, domain.TransactionTypeSell:
	default:
		return fmt.Errorf("%w: unsupported transaction type", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Currency) == "" {
		return ErrInvalidInput
	}
	return nil
}
