# ADR 0002 — Observability: structured logging and metrics

- **Status:** Accepted (partially implemented — see §Implementation status)
- **Date:** 2026-06-13
- **Deciders:** Engineering / Platform
- **Related:** [`docs/architecture.md`](../architecture.md) §12, workspace rule
  `14-logging-observability`

---

## Context

The system runs as several long-lived services (API, market-data worker,
Postgres, Redis) behind an nginx-served SPA. When something misbehaves —
a failed login spike, slow dashboards, stale prices, a slow query — we need to
diagnose it from signals, not guesswork. We also need those signals to be safe:
logs are a persistent, widely-shared data sink and must never carry secrets or
PII.

We need to decide how the application emits **logs** and **metrics**, and what we
commit to tracking.

---

## Decision

### Structured logging

- **Emit structured JSON logs** with stable field names, so they can be queried,
  filtered and scrubbed mechanically. Every log line for a domain action should
  carry the actor and the action, e.g.:

  ```json
  {
    "level": "INFO",
    "user_id": "123",
    "action": "CREATE_TRANSACTION"
  }
  ```

- **Standard fields**: `level`, `time`, `msg`, `request_id`, and where applicable
  `user_id`, `action`, and a result/duration. Use opaque IDs for correlation
  (`request_id`), never user emails.
- **Redaction at the boundary.** Never log passwords, tokens, JWTs, API keys,
  full request/response bodies on auth or PII paths, or raw provider responses.
  Use an allowlist of fields for sensitive endpoints (rule
  `14-logging-observability`).
- **Correlation.** Propagate `request_id` from the inbound request through to
  every log line and back to the client via the `X-Request-ID` header.

### Metrics

- **Use Grafana** for dashboards and visualization, backed by a metrics store
  (Prometheus-style scrape of a `/metrics` endpoint is the intended source).
- **Track, at minimum:**
  - **API latency** — request duration histograms per route/method/status.
  - **Failed logins** — counter of authentication failures (for abuse / lockout
    visibility), labeled by reason, never by raw credential.
  - **Market data refreshes** — worker sync runs: success/failure counts,
    duration, and number of assets/rates updated per run.
  - **Database query times** — query/transaction latency to spot slow paths.

---

## Implementation status

| Item | Status | Notes |
|---|---|---|
| Structured logging via `slog` | ✅ Implemented | Backend + worker use `log/slog`; worker logs synced asset/rate counts |
| `request_id` correlation | ✅ Implemented | `middleware.RequestID()` sets/propagates and echoes `X-Request-ID` |
| Health endpoint | ✅ Implemented | `GET /healthz`; Postgres/Redis container healthchecks |
| JSON log handler everywhere | ⛏️ Partial | Ensure all entrypoints use the JSON `slog` handler (not text) and consistent field names |
| `user_id` + `action` on domain events | ⛏️ Planned | Add structured fields to mutations (CREATE/UPDATE/DELETE transaction, dividend, login) |
| `/metrics` endpoint + Prometheus | ⛏️ Planned | No metrics exporter wired yet |
| Grafana dashboards | ⛏️ Planned | Add to compose; provision dashboards for the four tracked signals |
| API latency / DB query / failed-login / refresh metrics | ⛏️ Planned | Instrument middleware, repository, auth and worker |

---

## Consequences

### Positive

- JSON logs are greppable and shippable to any aggregator; structured
  `user_id`/`action` makes audit trails and incident forensics tractable.
- Boundary redaction means a missed call site cannot leak a secret into logs.
- The four core metrics cover the most common failure modes: slowness (latency,
  DB time), abuse (failed logins), and data freshness (market refreshes).

### Negative / costs

- Adds a metrics exporter dependency and a Grafana/Prometheus pair to the
  deployment (more containers, more config).
- Instrumentation is cross-cutting work across middleware, repository, auth and
  worker.
- Log/metric retention must match data class; metrics labels must avoid
  high-cardinality identifiers (no raw `user_id` as a Prometheus label).

### Follow-ups

1. Standardize a JSON `slog` handler and a small logging helper that always sets
   `request_id` and accepts `user_id`/`action` fields.
2. Add structured audit log lines on all user-owned mutations.
3. Add a Prometheus-compatible `/metrics` endpoint and instrument: per-route
   latency histogram, `auth_login_failures_total`, market-sync run metrics,
   DB query latency.
4. Add `prometheus` + `grafana` services to compose with provisioned dashboards.
5. Confirm log retention/scrubbing against `14-logging-observability` and the
   data-classification rule.

---

## Alternatives considered

- **Plain text logs.** Easier to read by eye, but hard to query, scrub and
  correlate at scale; rejected in favor of structured JSON.
- **APM SaaS (Datadog, New Relic).** Powerful, but sends telemetry off-host and
  raises data-classification questions for a self-hosted finance app; Grafana +
  Prometheus keeps signals local by default.
- **Logs-only (no metrics).** Cheaper, but latency/throughput trends and
  alerting are far weaker without real metrics; rejected.
