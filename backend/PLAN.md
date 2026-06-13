# Backend Implementation Plan — Personal Investment Portfolio Manager

This document is the engineering plan for the **backend** service described in
[`SPEC.md`](../SPEC.md). It covers architecture, project layout, data layer,
APIs, background jobs, security, testing, documentation, and a phased delivery
checklist. The backend also drives the `market-data-worker` since both share Go
domain code.

---

## 1. Scope

Implements sections 5–12 and the backend portions of 3, 14, 16 of the SPEC:

- JWT authentication (register, login, logout, password reset).
- Asset, transaction, and dividend management (CRUD).
- Portfolio calculations (average cost, realized/unrealized P/L, allocation,
  evolution).
- Market data aggregation (B3 / BRAPI, NASDAQ / Alpha Vantage|Finnhub|Twelve
  Data, Crypto / CoinGecko) and USD↔BRL currency conversion.
- Background price synchronization + portfolio recalculation.
- PostgreSQL persistence with migrations; optional Redis caching.

---

## 2. Technology & Key Decisions

- **Language:** Go 1.23+.
- **HTTP framework:** Gin (mature middleware ecosystem; SPEC allows Gin/Fiber —
  picking Gin).
- **DB access:** `pgx` v5 + `sqlc` for type-safe queries (no heavy ORM, keeps
  financial SQL explicit and auditable).
- **Migrations:** `golang-migrate` with versioned SQL files.
- **Auth:** `golang-jwt/jwt/v5`, `argon2id` password hashing (per security rule
  `11-cryptography` — Argon2id preferred).
- **Config:** `env`-based loader (`caarlos0/env` + `godotenv` for local dev).
- **Validation:** `go-playground/validator`.
- **Decimal math:** `shopspring/decimal` for all monetary values (never floats
  for money).
- **Logging:** `slog` (structured JSON), with secret/PII redaction at the writer.
- **Caching/queue:** Redis via `redis/go-redis/v9` (optional; feature-flagged).
- **Testing:** standard `testing` + `testify`, `testcontainers-go` for
  Postgres/Redis integration tests, `httptest` for handlers.
- **Docs:** OpenAPI 3.1 spec, served via Swagger UI; `golangci-lint`.

---

## 3. Project Structure

```
backend/
  cmd/
    api/main.go            # HTTP server entrypoint
    worker/main.go         # market-data-worker entrypoint (shares internal/)
    migrate/main.go        # migration runner
  internal/
    config/                # env loading + validation
    db/
      migrations/          # golang-migrate SQL files
      sqlc/                # generated type-safe queries
      queries/             # *.sql source for sqlc
    domain/                # entities + business rules (no I/O)
      portfolio/           # avg cost, P/L, allocation calculators
    repository/            # data access interfaces + pgx impls
    service/               # auth, asset, transaction, dividend, portfolio
    transport/http/
      handlers/            # gin handlers
      middleware/          # auth, request-id, rate-limit, recovery, CORS
      dto/                 # request/response structs + validation tags
    marketdata/            # provider clients (brapi, alphavantage, coingecko, fx)
    jobs/                  # price sync + recalculation scheduler
    auth/                  # jwt issue/verify, password hashing, reset tokens
    cache/                 # redis wrapper (no-op when disabled)
  api/openapi.yaml         # OpenAPI 3.1 contract
  test/                    # integration + e2e helpers, fixtures
  Dockerfile
  .env.example
  README.md
  Makefile
  go.mod
```

---

## 4. Data Layer & Migrations

Tables follow SPEC §10. Refinements for integrity and performance:

- All `id` are `UUID` (`gen_random_uuid()` via `pgcrypto`).
- `users.email` UNIQUE + `CITEXT`; `password_hash` not null.
- `transactions.type` constrained `CHECK (type IN ('BUY','SELL','DIVIDEND'))`.
- Monetary columns `NUMERIC(20,8)` (crypto precision), `currency CHAR(3)`.
- `assets` UNIQUE `(symbol, exchange)`.
- Foreign keys with `ON DELETE CASCADE` for user-owned rows; `RESTRICT` on
  `asset_id` from transactions.
- Indexes: `transactions(user_id, transaction_date)`, `asset_prices(asset_id,
  timestamp DESC)`, `dividends(user_id, payment_date)`.
- Add `password_reset_tokens` table (hashed token, expiry, used_at) — not in
  SPEC but required for password reset feature.
- Add `exchange_rates(base, quote, rate, timestamp)` for FX history.
- `created_at`/`updated_at TIMESTAMPTZ DEFAULT now()`; trigger to maintain
  `updated_at`.

Migrations are forward + down pairs, run via `cmd/migrate`. Seed migration adds
a small reference asset list for dev only (guarded, never in prod).

---

## 5. Domain & Portfolio Calculations (SPEC §8)

Pure functions in `internal/domain/portfolio`, fully unit-tested:

- **Average cost** = total invested (incl. fees) / total shares; recomputed on
  BUY; SELL reduces quantity at current average (no avg change).
- **Realized P/L** on SELL = (sell price − avg cost) × qty − fees.
- **Unrealized P/L** = current value − remaining cost basis.
- **Portfolio value** = Σ open positions × latest price, converted to preferred
  currency.
- **Daily performance** = current value − previous market close.
- **Currency** — always store original currency + amount; conversion computed at
  read time using `exchange_rates`. Original transaction values are immutable.

Edge cases to test: selling more than held (reject), zero-quantity positions,
multi-currency aggregation, dividends not affecting cost basis, fractional
crypto quantities.

---

## 6. REST API (SPEC §11)

Implement all listed endpoints under `/api`. Conventions:

- JSON; `Authorization: Bearer <jwt>`; consistent error envelope
  `{ "error": { "code", "message" } }`.
- All list endpoints paginated (`?limit`, `?cursor`) and scoped to the
  authenticated user.
- Input validation via DTO + validator; reject unknown fields.

Endpoints:

- **Auth:** `POST /auth/register`, `/auth/login`, `/auth/logout`,
  plus `POST /auth/password-reset/request` and `/auth/password-reset/confirm`.
- **Assets:** `GET /assets`, `GET /assets/:id` (search by symbol/exchange).
- **Portfolio:** `GET /portfolio`, `/portfolio/summary`,
  `/portfolio/performance?range=1M|3M|6M|1Y|5Y|ALL`.
- **Transactions:** `GET/POST /transactions`, `PUT/DELETE /transactions/:id`.
- **Dividends:** `GET/POST /dividends`, `PUT/DELETE /dividends/:id`.

Every mutating transaction/dividend call triggers portfolio recalculation
(SPEC §12).

---

## 7. Market Data Integration (SPEC §6) & Worker

- Provider interface `PriceProvider` with implementations: `brapi` (B3),
  `alphavantage`/`finnhub`/`twelvedata` (NASDAQ, one default + fallback),
  `coingecko` (crypto), and `fx` (Frankfurter default, ExchangeRate fallback).
- Each client: typed config, timeouts, retry with backoff, rate-limit respect,
  response size limits. Treat all external responses as untrusted input.
- `cmd/worker` scheduler (SPEC §12): every 1 min during market hours, every 15
  min otherwise; updates `asset_prices` + `exchange_rates`. Redis used to cache
  latest prices and dedupe in-flight fetches.
- API keys provided via env only; never logged.

---

## 8. Security (.env, DB, Auth)

Aligned with workspace security rules (`00`, `01`, `11`, `12`, `13`, `14`).

- **`.env` handling:**
  - Commit `.env.example` with placeholder keys only; real `.env` git-ignored
    (confirm root `.gitignore` covers `.env`, `.env.*`).
  - Required vars: `DATABASE_URL`, `JWT_SECRET` (≥32 bytes), `JWT_ACCESS_TTL`,
    `JWT_REFRESH_TTL`, `ARGON2_*` params, `REDIS_URL`, `BRAPI_TOKEN`,
    `ALPHAVANTAGE_KEY`, `COINGECKO_KEY`, `FX_PROVIDER_KEY`, `CORS_ORIGINS`,
    `ENV`.
  - Config loader fails fast if any required secret is missing or weak.
- **Database security:**
  - App connects as a least-privilege role (no superuser); migrations run with a
    separate role.
  - Parameterized queries only (sqlc/pgx) — no string-concatenated SQL.
  - TLS to Postgres in non-local envs (`sslmode=require`+).
  - Secrets/connection strings never printed to logs.
- **Auth:**
  - Argon2id password hashing; constant-time token comparison.
  - JWT `alg` allowlist (HS256 dev / consider RS256), verify `exp`/`iss`/`aud`;
    short access TTL + refresh.
  - Password-reset tokens hashed at rest, single-use, expiring.
- **Web defenses:** schema validation, strict CORS allowlist, security headers,
  rate limiting on `/auth/*` and expensive endpoints, recovery middleware.
- **Logging:** structured JSON, redact tokens/passwords/PII, opaque request IDs.
- **Note:** Auth and crypto paths are sensitive (`20-human-review`) — flag for
  security review before merge.

---

## 9. Testing Strategy

- **Unit tests:** all `domain/portfolio` calculators (table-driven, decimal
  precision, edge cases); auth (hashing, jwt, reset tokens); marketdata response
  parsing with recorded fixtures.
- **Repository/integration tests:** `testcontainers-go` Postgres; run real
  migrations; verify constraints, cascades, indexes.
- **Handler/API tests:** `httptest` against the Gin router with a containerized
  DB; cover auth flows, ownership isolation (user A cannot read user B), input
  validation, error envelopes.
- **Worker tests:** scheduler timing logic + provider clients with mocked HTTP.
- **Coverage gate:** ≥80% on `domain`/`service`; CI runs `go test ./... -race`.
- **Lint/security:** `golangci-lint`, `go vet`, `govulncheck`, `gosec`.

---

## 10. Documentation

- `backend/README.md`: setup, env vars, run/migrate/test commands, architecture
  overview.
- `api/openapi.yaml`: full OpenAPI 3.1 contract, served at `/docs` (Swagger UI)
  in non-prod.
- Inline package doc comments for `domain`, `service`, `marketdata`.
- `docs/` notes for calculation formulas and provider quirks.
- Threat-model note in PR for new auth + external integrations
  (`21-threat-modeling`).

---

## 11. Delivery Phases (Checklist)

1. **Bootstrap:** module, config loader, `.env.example`, Dockerfile, Makefile,
   CI skeleton, logging.
2. **DB foundation:** migrations for all tables + extras, sqlc setup, repository
   layer + integration tests.
3. **Auth:** register/login/logout/reset, JWT middleware, hashing — with tests +
   security review.
4. **Assets & Transactions:** CRUD, validation, ownership scoping.
5. **Portfolio engine:** domain calculators + `/portfolio*` endpoints + recalc
   trigger.
6. **Dividends:** CRUD + income aggregation.
7. **Market data + worker:** provider clients, FX, scheduler, Redis cache.
8. **Hardening:** rate limiting, headers, redaction, coverage gate, OpenAPI,
   README, govulncheck/gosec.

---

## 12. Open Questions / Assumptions

- NASDAQ provider default: assume **Finnhub** primary, Twelve Data fallback
  (free-tier friendliness) — confirm.
- Single-user MVP per SPEC §16, but schema is multi-user ready.
- `market-data-worker` lives in `backend/cmd/worker` (shared Go module) rather
  than a separate top-level `worker/` — adjust docker-compose `build` context
  accordingly.
