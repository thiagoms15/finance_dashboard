CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(32) NOT NULL,
    name VARCHAR(255) NOT NULL,
    exchange VARCHAR(32) NOT NULL,
    currency CHAR(3) NOT NULL,
    sector VARCHAR(128) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(symbol, exchange)
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE RESTRICT,
    type VARCHAR(16) NOT NULL CHECK (type IN ('BUY', 'SELL', 'DIVIDEND')),
    quantity NUMERIC(20,8) NOT NULL DEFAULT 0,
    price NUMERIC(20,8) NOT NULL DEFAULT 0,
    fees NUMERIC(20,8) NOT NULL DEFAULT 0,
    currency CHAR(3) NOT NULL,
    transaction_date TIMESTAMPTZ NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_date
    ON transactions(user_id, transaction_date DESC);

CREATE TABLE IF NOT EXISTS dividends (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE RESTRICT,
    amount NUMERIC(20,8) NOT NULL,
    currency CHAR(3) NOT NULL,
    payment_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_dividends_user_date
    ON dividends(user_id, payment_date DESC);

CREATE TABLE IF NOT EXISTS asset_prices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    price NUMERIC(20,8) NOT NULL,
    previous_close NUMERIC(20,8) NOT NULL DEFAULT 0,
    currency CHAR(3) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_asset_prices_asset_timestamp
    ON asset_prices(asset_id, timestamp DESC);

CREATE TABLE IF NOT EXISTS exchange_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base CHAR(3) NOT NULL,
    quote CHAR(3) NOT NULL,
    rate NUMERIC(20,8) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_exchange_rates_base_quote_timestamp
    ON exchange_rates(base, quote, timestamp DESC);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user
    ON password_reset_tokens(user_id, expires_at DESC);

CREATE TRIGGER users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER assets_set_updated_at
    BEFORE UPDATE ON assets
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER transactions_set_updated_at
    BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER dividends_set_updated_at
    BEFORE UPDATE ON dividends
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

INSERT INTO assets (symbol, name, exchange, currency, sector)
VALUES
    ('AAPL', 'Apple Inc.', 'NASDAQ', 'USD', 'Technology'),
    ('MSFT', 'Microsoft Corporation', 'NASDAQ', 'USD', 'Technology'),
    ('PETR4', 'Petrobras PN', 'B3', 'BRL', 'Energy'),
    ('VALE3', 'Vale ON', 'B3', 'BRL', 'Materials'),
    ('BTC', 'Bitcoin', 'CRYPTO', 'USD', 'Crypto'),
    ('ETH', 'Ethereum', 'CRYPTO', 'USD', 'Crypto')
ON CONFLICT (symbol, exchange) DO NOTHING;
