package dto

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PasswordResetRequest struct {
	Email string `json:"email"`
}

type PasswordResetConfirmRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

type AssetRequest struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Exchange string `json:"exchange"`
	Currency string `json:"currency"`
	Sector   string `json:"sector"`
}

type TransactionRequest struct {
	AssetID         string `json:"assetId"`
	Type            string `json:"type"`
	Quantity        string `json:"quantity"`
	Price           string `json:"price"`
	Fees            string `json:"fees"`
	Currency        string `json:"currency"`
	TransactionDate string `json:"transactionDate"`
	Notes           string `json:"notes"`
}

type DividendRequest struct {
	AssetID     string `json:"assetId"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	PaymentDate string `json:"paymentDate"`
}
