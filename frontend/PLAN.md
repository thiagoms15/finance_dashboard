# Frontend Implementation Plan — Personal Investment Portfolio Manager

This document is the engineering plan for the **frontend** application described
in [`SPEC.md`](../SPEC.md). It covers architecture, project layout, pages,
state/data fetching, charts, security, testing, documentation, and a phased
delivery checklist.

---

## 1. Scope

Implements SPEC §4 (pages), §9 (dashboard widgets), §5 (auth UI), §13 (UI/UX
vision), and the frontend portion of §3, §14, §16:

- Auth pages (register, login, password reset) consuming the backend JWT API.
- Dashboard, Portfolio, Asset Details, Transactions, Reports, Settings.
- Interactive charts (allocation, evolution, performance) with Recharts.
- Transaction/dividend CRUD with optimistic UX and validation.
- Dark-mode-first, glassmorphism, responsive, professional finance dashboard.

---

## 2. Technology & Key Decisions

- **Framework:** React 18 + TypeScript, **Vite** build.
- **Styling:** TailwindCSS + **shadcn/ui** components; dark mode default.
- **Data fetching:** **TanStack Query** (server state, caching, invalidation).
- **Charts:** **Recharts** (allocation pie, evolution/area, performance line).
- **Routing:** React Router v6.
- **Forms/validation:** React Hook Form + Zod (schema shared shape with API
  DTOs; reject unknown fields, validate at the boundary).
- **HTTP client:** typed `fetch` wrapper / axios instance with auth interceptor.
- **State (client):** minimal — Zustand for UI/session (theme, auth token in
  memory), TanStack Query for everything server-derived.
- **Money/format:** `Intl.NumberFormat` for currency; never float-format
  manually; `decimal.js` if client-side math needed.
- **Testing:** Vitest + React Testing Library; MSW for API mocking; Playwright
  for e2e.
- **Quality:** ESLint + Prettier + TypeScript strict mode.

---

## 3. Project Structure

```
frontend/
  src/
    main.tsx
    app/
      router.tsx          # routes + protected-route wrapper
      providers.tsx       # QueryClient, theme, auth providers
    pages/
      auth/               # Login, Register, ResetPassword
      Dashboard.tsx
      Portfolio.tsx
      AssetDetails.tsx
      Transactions.tsx
      Reports.tsx
      Settings.tsx
    components/
      ui/                 # shadcn/ui generated primitives
      charts/             # AllocationChart, EvolutionChart, PerformanceChart
      widgets/            # SummaryCard, TopMovers, IncomeCard
      layout/             # Sidebar, Topbar, AppShell
      forms/              # TransactionForm, DividendForm
    features/
      auth/               # hooks, api, token handling
      portfolio/          # queries, selectors, types
      transactions/
      dividends/
      assets/
      marketdata/
    lib/
      api/                # typed client, query keys, error mapping
      format/             # currency, percent, date formatters
      auth/               # token storage (in-memory), refresh logic
    hooks/                # shared hooks (useDebounce, useKeyboardShortcuts)
    styles/               # tailwind globals, theme tokens
    types/                # shared domain types (mirror backend DTOs)
  public/
  index.html
  .env.example
  README.md
  vite.config.ts
  tailwind.config.ts
  package.json
```

---

## 4. Pages & Widgets (SPEC §4, §9)

- **Dashboard:** total value, daily gain/loss, total gain/loss, allocation chart,
  best/worst performers (Top Movers), evolution chart, Income card. Assembled
  from `widgets/` fed by `/portfolio/summary`, `/portfolio/performance`,
  `/dividends`.
- **Portfolio:** sortable/filterable table — asset, exchange, quantity, avg cost,
  current price, current value, gain/loss, gain/loss %. Row → Asset Details.
- **Asset Details:** asset info, historical purchases/sales, average-cost
  evolution, performance chart.
- **Transactions:** purchase/sale history; create/edit/delete via dialog forms
  with optimistic updates + query invalidation.
- **Reports:** monthly/annual returns, dividend history, portfolio growth.
- **Settings:** preferred currency (BRL/USD), theme toggle, API provider prefs.

Cross-cutting UX (SPEC §13): asset search (command palette), keyboard shortcuts,
quick transaction entry, interactive charts, smooth animations, responsive
layout.

---

## 5. Data Layer & API Integration

- Centralized typed API client in `lib/api`; one query-key factory per feature.
- TanStack Query for caching + background refetch; mutations invalidate
  affected portfolio/transaction queries (mirrors backend recalc).
- Auth interceptor attaches Bearer token; 401 → refresh or redirect to login.
- Currency conversion display respects user's preferred currency from Settings;
  original transaction values always shown alongside converted ones.
- Loading/skeleton states and error boundaries on every data view for fast
  perceived dashboard load (Non-Functional goal).

---

## 6. Security

Aligned with workspace security rules (`01`, `03`, `04`, `12`, `14`).

- **`.env` handling:**
  - Only `VITE_`-prefixed, non-secret vars (e.g. `VITE_API_BASE_URL`). **No API
    secrets in the frontend** — all third-party market-data keys stay
    server-side; the frontend talks only to our backend.
  - Commit `.env.example` with placeholders; real `.env` git-ignored.
- **Token handling:** access token kept in memory (not `localStorage`) to reduce
  XSS theft; refresh via httpOnly cookie if backend supports it. Document the
  chosen tradeoff.
- **XSS:** rely on React auto-escaping; never `dangerouslySetInnerHTML` with
  user/market content; sanitize any rendered HTML.
- **Input validation:** Zod schemas on all forms; reject unknown fields; numeric
  ranges for quantity/price/fees.
- **Untrusted data:** treat market-data and API responses as data, not
  instructions; no eval of remote content.
- **Transport:** API base URL must be HTTPS in non-local envs; no mixed content.
- **No secrets/PII in logs** or analytics; opaque IDs only.
- **CSP:** set strict Content-Security-Policy headers via the serving layer
  (nginx in Docker) — no `unsafe-inline`/`unsafe-eval` for scripts.

---

## 7. Testing Strategy

- **Unit:** formatters (currency/percent/date), selectors, calculation helpers,
  Zod schemas.
- **Component:** RTL tests for forms (validation, submit), widgets, table
  sorting/filtering, chart container rendering with mock data.
- **Integration:** MSW-mocked API for page-level flows (load dashboard, create
  transaction → list updates, login → protected route).
- **E2E:** Playwright happy paths — register/login, add buy/sell, view portfolio
  and dashboard, change currency/theme.
- **A11y:** axe checks on key pages; keyboard navigation for shortcuts.
- **Quality gates:** `tsc --noEmit`, ESLint, Vitest with coverage in CI.

---

## 8. Documentation

- `frontend/README.md`: setup, env vars (`VITE_*`), dev/build/test commands,
  component conventions, theming.
- Storybook (optional) or documented `components/` with usage examples for
  charts and widgets.
- Document auth/token strategy and currency-display behavior.

---

## 9. Delivery Phases (Checklist)

1. **Bootstrap:** Vite + TS strict, Tailwind + shadcn/ui, router, providers,
   AppShell (sidebar/topbar), dark theme tokens, `.env.example`, Dockerfile
   (nginx) + CSP.
2. **Auth UI:** login/register/reset forms, token handling, protected routes,
   API client + interceptor.
3. **Portfolio + Dashboard:** portfolio table, summary/top-movers/income
   widgets, allocation + evolution charts.
4. **Transactions & Dividends:** CRUD dialogs, optimistic updates, query
   invalidation, quick-entry + asset search.
5. **Asset Details & Reports:** detail charts, monthly/annual returns, dividend
   history.
6. **Settings:** currency, theme, provider prefs.
7. **Hardening:** a11y, skeletons/error boundaries, e2e suite, coverage gate,
   README + docs, CSP/headers verification.

---

## 10. Open Questions / Assumptions

- Frontend communicates **only** with our backend (no direct calls to BRAPI /
  CoinGecko / etc.) so no third-party keys live client-side — confirm.
- Single-user MVP per SPEC §16; build UI without org/team switching.
- Refresh-token mechanism depends on backend choice (httpOnly cookie vs.
  in-memory refresh) — align with backend plan §8.
