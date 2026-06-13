package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionType string

const (
	TransactionTypeBuy      TransactionType = "BUY"
	TransactionTypeSell     TransactionType = "SELL"
	TransactionTypeDividend TransactionType = "DIVIDEND"
)

type User struct {
	ID                  uuid.UUID  `json:"id"`
	Name                string     `json:"name"`
	Email               string     `json:"email"`
	PasswordHash        string     `json:"-"`
	FailedLoginAttempts int        `json:"-"`
	FirstFailedLoginAt  *time.Time `json:"-"`
	LockedUntil         *time.Time `json:"-"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type Asset struct {
	ID        uuid.UUID `json:"id"`
	Symbol    string    `json:"symbol"`
	Name      string    `json:"name"`
	Exchange  string    `json:"exchange"`
	Currency  string    `json:"currency"`
	Sector    string    `json:"sector"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Transaction struct {
	ID              uuid.UUID       `json:"id"`
	UserID          uuid.UUID       `json:"userId"`
	AssetID         uuid.UUID       `json:"assetId"`
	Type            TransactionType `json:"type"`
	Quantity        decimal.Decimal `json:"quantity"`
	Price           decimal.Decimal `json:"price"`
	Fees            decimal.Decimal `json:"fees"`
	Currency        string          `json:"currency"`
	TransactionDate time.Time       `json:"transactionDate"`
	Notes           string          `json:"notes"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

type Dividend struct {
	ID          uuid.UUID       `json:"id"`
	UserID      uuid.UUID       `json:"userId"`
	AssetID     uuid.UUID       `json:"assetId"`
	Amount      decimal.Decimal `json:"amount"`
	Currency    string          `json:"currency"`
	PaymentDate time.Time       `json:"paymentDate"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

type AssetPrice struct {
	ID            uuid.UUID       `json:"id"`
	AssetID       uuid.UUID       `json:"assetId"`
	Price         decimal.Decimal `json:"price"`
	PreviousClose decimal.Decimal `json:"previousClose"`
	Currency      string          `json:"currency"`
	Timestamp     time.Time       `json:"timestamp"`
}

type ExchangeRate struct {
	ID        uuid.UUID       `json:"id"`
	Base      string          `json:"base"`
	Quote     string          `json:"quote"`
	Rate      decimal.Decimal `json:"rate"`
	Timestamp time.Time       `json:"timestamp"`
}

type PasswordResetToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"userId"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

type RefreshToken struct {
	ID                  uuid.UUID  `json:"id"`
	UserID              uuid.UUID  `json:"userId"`
	TokenHash           string     `json:"-"`
	ExpiresAt           time.Time  `json:"expiresAt"`
	RevokedAt           *time.Time `json:"revokedAt,omitempty"`
	ReplacedByTokenHash string     `json:"-"`
	CreatedAt           time.Time  `json:"createdAt"`
}
