# Frontend

React + TypeScript + Vite frontend for the Personal Investment Portfolio Manager.

## Environment

Copy `.env.example` to `.env` for local development.

Only non-secret `VITE_*` variables belong here:

- `VITE_API_BASE_URL`

All provider API keys stay on the backend.

## Commands

```bash
npm install
npm run dev
npm run build
npm run test
npm run lint
npm run typecheck
```

## Notes

- Auth sessions are persisted locally with an expiration timestamp so reloads do
  not immediately sign the user out.
- The production container serves the SPA through nginx and proxies `/api` to the backend.
- CSP headers are configured in `nginx.conf`.
- Asset detail pages render backend-proxied icons using CSP-compatible `data:`
  image URLs.
- The authenticated shell shows a friendly welcome using the user's display
  name when available.
