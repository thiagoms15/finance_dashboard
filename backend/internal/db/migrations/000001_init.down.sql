DROP TRIGGER IF EXISTS dividends_set_updated_at ON dividends;
DROP TRIGGER IF EXISTS transactions_set_updated_at ON transactions;
DROP TRIGGER IF EXISTS assets_set_updated_at ON assets;
DROP TRIGGER IF EXISTS users_set_updated_at ON users;

DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS exchange_rates;
DROP TABLE IF EXISTS asset_prices;
DROP TABLE IF EXISTS dividends;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS users;

DROP FUNCTION IF EXISTS set_updated_at;
