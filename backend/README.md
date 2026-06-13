# Backend

Go backend for the Personal Investment Portfolio Manager.

## Stack

- Go + Gin
- PostgreSQL + `pgx`
- JWT auth
- Argon2id password hashing
- Decimal-safe portfolio calculations

## Environment

Copy `.env.example` to `.env` and fill in local values.

Important variables:

- `DATABASE_URL`
- `MIGRATIONS_DATABASE_URL`
- `JWT_SECRET`
- `JWT_ISSUER`
- `JWT_AUDIENCE`
- `JWT_ACCESS_TTL`
- `JWT_REFRESH_TTL`
- `REFRESH_COOKIE_NAME`
- `REFRESH_COOKIE_PATH`
- `REFRESH_COOKIE_DOMAIN`
- `REFRESH_COOKIE_SAME_SITE`
- `REFRESH_COOKIE_SECURE`
- `CORS_ORIGINS`

Do not commit `.env` or any real secrets.

## Commands

```bash
make migrate-up
make run-api
make run-worker
make test
make fmt
```

## API

Main routes:

- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `POST /api/auth/logout`
- `POST /api/auth/password-reset/request`
- `POST /api/auth/password-reset/confirm`
- `GET /api/assets`
- `GET /api/assets/:id`
- `GET /api/assets/:id/icon`
- `GET /api/transactions`
- `POST /api/transactions`
- `PUT /api/transactions/:id`
- `DELETE /api/transactions/:id`
- `GET /api/dividends`
- `POST /api/dividends`
- `PUT /api/dividends/:id`
- `DELETE /api/dividends/:id`
- `GET /api/portfolio`
- `GET /api/portfolio/summary`
- `GET /api/portfolio/performance`

## Notes

- Password reset returns a raw token only in `development` mode to keep local
  testing possible without email delivery.
- Login and registration issue a short-lived access token in JSON plus a
  hashed, DB-backed refresh token in a secure HttpOnly cookie. Logout revokes
  the refresh token server-side.
- Failed login attempts are tracked per account; 5 failures within 15 minutes
  lock the account for 15 minutes.
- Users now have a `name` field, returned in auth responses and used by the
  frontend welcome UI.
- Live market data currently uses:
  - BRAPI for B3 quotes when available, with public Yahoo Finance quote fallback
  - Yahoo Finance public quote endpoint for NASDAQ/NYSE-style symbols
  - CoinGecko public API for crypto prices
  - Frankfurter public API for USD/BRL FX rates
- Asset icons are proxied through the backend:
  - B3 icons use BRAPI metadata with direct `icons.brapi.dev` fallback
  - U.S. stock icons use FMP image endpoints
  - Crypto icons use CoinGecko images with direct fallbacks for common coins
- The Yahoo quote path is an unofficial public endpoint, so it may be more
  fragile than keyed commercial providers.
- Auth and crypto code paths should get human security review before merge.
