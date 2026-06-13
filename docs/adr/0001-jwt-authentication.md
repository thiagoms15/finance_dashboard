# ADR 0001 — Authentication, authorization and secrets handling

- **Status:** Accepted and implemented (except Docker secrets future work)
- **Date:** 2026-06-13
- **Deciders:** Engineering / Security
- **Related:** [`docs/architecture.md`](../architecture.md) §9, workspace rules
  `00-global-security`, `01-secrets-and-data`, `11-cryptography`, `20-human-review`

---

## Context

The Personal Investment Portfolio Manager stores financial data tied to a user
account. Even though the MVP is effectively single-user, the system handles
credentials and personal investment records, so authentication, authorization
and secret handling must be designed correctly from day one — retrofitting
tenant isolation or credential hygiene later is expensive and error-prone.

We need to decide:

1. How users authenticate and how API requests are authorized.
2. How session lifetime and revocation are handled.
3. How data is isolated per user.
4. How secrets (JWT signing key, DB password, third-party API keys) are managed.

This is a security-sensitive path and therefore subject to human review before
changes.

---

## Decision

### Authentication

- **JWT access tokens, short-lived.** API requests are authorized with a stateless
  `Bearer` JWT validated on every protected route. Access TTL is short
  (default 15 minutes, `JWT_ACCESS_TTL`).
- **Refresh tokens stored in the database.** Longer-lived refresh tokens are
  persisted server-side (hashed) so they can be revoked and rotated, unlike the
  stateless access token. A refresh TTL is configured separately
  (`JWT_REFRESH_TTL`, default 7 days).
- **Password hashing with Argon2id** (preferred over bcrypt). Parameters are
  configurable (`ARGON2_TIME`, `ARGON2_MEMORY`, `ARGON2_THREADS`,
  `ARGON2_KEY_LEN`). Passwords are never stored or logged in plaintext.
- **Account lockout after repeated failed logins.** Repeated authentication
  failures must throttle and eventually lock the account for a cooldown window
  to resist credential stuffing / brute force, complemented by per-IP and
  per-account rate limiting.

### Authorization

- Model the domain hierarchy explicitly, even as a single-user app today:

  ```
  User
   └── Portfolio
        └── Transactions (and Dividends)
  ```

- **Every data query MUST filter by `user_id`.** No portfolio, transaction or
  dividend read/write may return or mutate rows that do not belong to the
  authenticated user. The `user_id` is taken from the verified JWT, never from
  request input.
- Foreign keys cascade user-owned data on user delete; shared reference data
  (assets, prices) is read-only to users.

### Secrets

- **Never store secrets in source code.** This includes:
  - third-party API keys (BRAPI, Finnhub, CoinGecko, FX provider),
  - the JWT signing secret,
  - the database password / connection string credentials.
- **Use `.env` files** (git-ignored) for local and compose configuration, loaded
  into the backend/worker via environment variables.
- **Docker secrets (future).** Migrate from plain `.env` injection to Docker /
  orchestrator secrets for production-grade deployments.
- **Provide a committed `.env.example`** with placeholder (non-secret) values so
  operators know which variables are required without leaking real values.

---

## Implementation status

Tracked honestly so reviewers know what is real vs. planned:

| Item | Status | Notes |
|---|---|---|
| Short-lived JWT access tokens | ✅ Implemented | `internal/auth/jwt.go`, `JWTAuth` middleware, `JWT_ACCESS_TTL` default 15m |
| Argon2id password hashing | ✅ Implemented | `internal/auth/password.go`, `ARGON2_*` config |
| `JWT_SECRET` validation (≥ 32 chars) | ✅ Implemented | `internal/config/config.go` |
| Per-user query isolation | ✅ Implemented | service/repository pass `user_id` from the JWT on all user-owned reads/writes |
| `.env` / `.env.example`, no secrets in code | ✅ Implemented | root + `backend/.env.example`, git-ignored real `.env` |
| Auth rate limiting | ✅ Implemented | `RateLimit(5, 10s)` on `/api/auth/*` |
| **Refresh tokens stored in DB** | ✅ Implemented | `refresh_tokens` table, hashed token storage, rotation on `/api/auth/refresh`, revocation on logout |
| **Account lockout** | ✅ Implemented | 5 failed logins in 15 minutes locks the account for 15 minutes |
| **Docker secrets** | ⛏️ Future | Currently `.env`-based injection |

---

## Consequences

### Positive

- Stateless access tokens keep the hot path cheap (no DB hit to authorize a
  request).
- DB-backed refresh tokens give us revocation/rotation that pure stateless JWT
  cannot.
- Argon2id is a current best-practice, memory-hard KDF.
- Enforcing `user_id` filtering everywhere makes multi-user support a small step
  rather than a rewrite.
- Keeping secrets out of source and out of the client image limits blast radius
  of a repo or image leak.

### Negative / costs

- Short access TTL requires a working refresh flow for good UX; this is now
  handled by DB-backed refresh tokens in a secure HttpOnly cookie.
- DB-backed refresh tokens add storage, rotation and cleanup logic.
- Account lockout must be tuned to avoid denial-of-service-by-lockout.

### Follow-ups

1. Add operational cleanup for expired/revoked refresh tokens.
2. Add structured audit logging for lockouts, refresh rotation and logout
   revocation (without logging tokens or secrets).
3. Plan migration to Docker/orchestrator secrets for production.

---

## Alternatives considered

- **Server-side sessions (cookie + session store).** Simpler revocation, but adds
  a stateful lookup on every request and CSRF surface; rejected in favor of
  stateless access + revocable refresh.
- **bcrypt instead of Argon2id.** Acceptable (cost ≥ 12) per the crypto policy,
  but Argon2id is preferred as the memory-hard default.
- **Long-lived access tokens, no refresh.** Simplest, but no practical revocation
  and a much larger window if a token leaks; rejected.
