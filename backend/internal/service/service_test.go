package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/thiago/finance/backend/internal/auth"
	"github.com/thiago/finance/backend/internal/domain"
)

type mockStore struct {
	createUserFn           func(ctx context.Context, name, email, passwordHash string) (domain.User, error)
	getUserByEmailFn       func(ctx context.Context, email string) (domain.User, error)
	listTransactionsFn     func(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Transaction, error)
	listDividendsFn        func(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Dividend, error)
	listLatestPricesFn     func(ctx context.Context) ([]domain.AssetPrice, error)
	listAssetsFn           func(ctx context.Context, search string, limit int) ([]domain.Asset, error)
	getLatestRatesFn       func(ctx context.Context) ([]domain.ExchangeRate, error)
}

func (m *mockStore) CreateUser(ctx context.Context, name, email, passwordHash string) (domain.User, error) {
	if m.createUserFn != nil {
		return m.createUserFn(ctx, name, email, passwordHash)
	}
	return domain.User{}, errors.New("create user not implemented")
}

func (m *mockStore) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	if m.getUserByEmailFn != nil {
		return m.getUserByEmailFn(ctx, email)
	}
	return domain.User{}, errors.New("get user by email not implemented")
}

func (m *mockStore) UpdateUserPassword(context.Context, uuid.UUID, string) error { return nil }
func (m *mockStore) UpdateUserName(context.Context, uuid.UUID, string) error     { return nil }
func (m *mockStore) CreatePasswordResetToken(context.Context, uuid.UUID, string, time.Time) error {
	return nil
}
func (m *mockStore) ConsumePasswordResetToken(context.Context, string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *mockStore) ListAssets(ctx context.Context, search string, limit int) ([]domain.Asset, error) {
	if m.listAssetsFn != nil {
		return m.listAssetsFn(ctx, search, limit)
	}
	return nil, nil
}
func (m *mockStore) GetAssetByID(context.Context, uuid.UUID) (domain.Asset, error)         { return domain.Asset{}, nil }
func (m *mockStore) CreateAsset(context.Context, domain.Asset) (domain.Asset, error)        { return domain.Asset{}, nil }
func (m *mockStore) ListTransactions(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Transaction, error) {
	if m.listTransactionsFn != nil {
		return m.listTransactionsFn(ctx, userID, limit)
	}
	return nil, nil
}
func (m *mockStore) CreateTransaction(context.Context, domain.Transaction) (domain.Transaction, error) {
	return domain.Transaction{}, nil
}
func (m *mockStore) UpdateTransaction(context.Context, domain.Transaction) (domain.Transaction, error) {
	return domain.Transaction{}, nil
}
func (m *mockStore) DeleteTransaction(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (m *mockStore) ListDividends(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Dividend, error) {
	if m.listDividendsFn != nil {
		return m.listDividendsFn(ctx, userID, limit)
	}
	return nil, nil
}
func (m *mockStore) CreateDividend(context.Context, domain.Dividend) (domain.Dividend, error) {
	return domain.Dividend{}, nil
}
func (m *mockStore) UpdateDividend(context.Context, domain.Dividend) (domain.Dividend, error) {
	return domain.Dividend{}, nil
}
func (m *mockStore) DeleteDividend(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (m *mockStore) ListLatestPrices(ctx context.Context) ([]domain.AssetPrice, error) {
	if m.listLatestPricesFn != nil {
		return m.listLatestPricesFn(ctx)
	}
	return nil, nil
}
func (m *mockStore) UpsertAssetPrice(context.Context, domain.AssetPrice) error { return nil }
func (m *mockStore) GetLatestExchangeRates(ctx context.Context) ([]domain.ExchangeRate, error) {
	if m.getLatestRatesFn != nil {
		return m.getLatestRatesFn(ctx)
	}
	return nil, nil
}
func (m *mockStore) UpsertExchangeRate(context.Context, domain.ExchangeRate) error { return nil }

func newTestService(store Store) *AppService {
	return New(
		store,
		auth.NewTokenManager("test-secret", "test-issuer", "test-aud"),
		auth.PasswordHasher{Time: 1, Memory: 64 * 1024, Threads: 2, KeyLen: 32},
		15*time.Minute,
		7*24*time.Hour,
	)
}

func TestRegisterNormalizesInputAndReturnsToken(t *testing.T) {
	t.Parallel()

	var gotName, gotEmail, gotHash string
	store := &mockStore{
		createUserFn: func(_ context.Context, name, email, passwordHash string) (domain.User, error) {
			gotName, gotEmail, gotHash = name, email, passwordHash
			return domain.User{
				ID:           uuid.New(),
				Name:         name,
				Email:        email,
				PasswordHash: passwordHash,
				CreatedAt:    time.Now().UTC(),
				UpdatedAt:    time.Now().UTC(),
			}, nil
		},
	}
	svc := newTestService(store)

	out, err := svc.Register(context.Background(), RegisterInput{
		Name:     "  John Doe  ",
		Email:    "  JOHN.DOE@EXAMPLE.COM ",
		Password: "12345678",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if gotName != "John Doe" {
		t.Fatalf("stored name = %q, want %q", gotName, "John Doe")
	}
	if gotEmail != "john.doe@example.com" {
		t.Fatalf("stored email = %q, want %q", gotEmail, "john.doe@example.com")
	}
	if gotHash == "" || gotHash == "12345678" {
		t.Fatalf("password hash was not generated correctly: %q", gotHash)
	}
	if out.AccessToken == "" {
		t.Fatal("access token should not be empty")
	}
	if out.User.Name != "John Doe" {
		t.Fatalf("response user name = %q, want %q", out.User.Name, "John Doe")
	}
}

func TestLoginRejectsInvalidPassword(t *testing.T) {
	t.Parallel()

	hasher := auth.PasswordHasher{Time: 1, Memory: 64 * 1024, Threads: 2, KeyLen: 32}
	hash, err := hasher.Hash("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	store := &mockStore{
		getUserByEmailFn: func(_ context.Context, email string) (domain.User, error) {
			return domain.User{
				ID:           uuid.New(),
				Name:         "User",
				Email:        email,
				PasswordHash: hash,
			}, nil
		},
	}
	svc := New(store, auth.NewTokenManager("test-secret", "test-issuer", "test-aud"), hasher, 15*time.Minute, 7*24*time.Hour)

	_, err = svc.Login(context.Background(), "user@example.com", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestPortfolioSnapshotHandlesDescendingTransactionsFromStore(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	assetID := uuid.New()
	now := time.Now().UTC()
	store := &mockStore{
		listTransactionsFn: func(_ context.Context, uid uuid.UUID, _ int) ([]domain.Transaction, error) {
			if uid != userID {
				t.Fatalf("unexpected user id: %v", uid)
			}
			// Store returns DESC order (latest first), which service must handle.
			return []domain.Transaction{
				{
					UserID:          userID,
					AssetID:         assetID,
					Type:            domain.TransactionTypeSell,
					Quantity:        decimal.RequireFromString("2"),
					Price:           decimal.RequireFromString("120"),
					Currency:        "USD",
					TransactionDate: now.Add(2 * time.Hour),
					CreatedAt:       now.Add(2 * time.Hour),
				},
				{
					UserID:          userID,
					AssetID:         assetID,
					Type:            domain.TransactionTypeBuy,
					Quantity:        decimal.RequireFromString("5"),
					Price:           decimal.RequireFromString("100"),
					Currency:        "USD",
					TransactionDate: now,
					CreatedAt:       now,
				},
			}, nil
		},
		listDividendsFn: func(_ context.Context, _ uuid.UUID, _ int) ([]domain.Dividend, error) {
			return nil, nil
		},
		listLatestPricesFn: func(_ context.Context) ([]domain.AssetPrice, error) {
			return []domain.AssetPrice{
				{
					AssetID:       assetID,
					Price:         decimal.RequireFromString("130"),
					PreviousClose: decimal.RequireFromString("125"),
					Currency:      "USD",
					Timestamp:     now.Add(3 * time.Hour),
				},
			}, nil
		},
		listAssetsFn: func(_ context.Context, _ string, _ int) ([]domain.Asset, error) {
			return []domain.Asset{
				{
					ID:       assetID,
					Symbol:   "AAPL",
					Exchange: "NASDAQ",
					Currency: "USD",
				},
			}, nil
		},
		getLatestRatesFn: func(_ context.Context) ([]domain.ExchangeRate, error) {
			return nil, nil
		},
	}
	svc := newTestService(store)

	snapshot, err := svc.PortfolioSnapshot(context.Background(), userID, "USD")
	if err != nil {
		t.Fatalf("PortfolioSnapshot() error = %v", err)
	}
	if got, want := len(snapshot.Positions), 1; got != want {
		t.Fatalf("positions length = %d, want %d", got, want)
	}
	if got, want := snapshot.Positions[0].Quantity.StringFixed(0), "3"; got != want {
		t.Fatalf("remaining quantity = %s, want %s", got, want)
	}
}
