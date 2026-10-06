# Runbook

A compact runbook manager: people sign in to write, review, and publish operational notes; machines fetch published notes using revocable API keys. The backend is Go + SQLite, and the React SPA is embedded into the single Go binary.

## Local setup

1. Copy `.env.example` to `.env`, change the seed password, and set `COOKIE_SECURE=false` for local HTTP.
2. `make build`
3. `./bin/runbook`
4. Sign in using `SEED_ADMIN_EMAIL` and `SEED_ADMIN_PASSWORD` from `.env`.

For frontend hot reload, run `make dev` in one terminal and `cd web && npm run dev` in another. The Vite dev server proxies `/api` to the backend.

## API

Session endpoints use an HttpOnly, SameSite=Lax cookie. Create API keys under **API keys**; the secret is shown only once. Send it as `Authorization: Bearer rb_…` to `/api/v1/runbooks`. Public endpoints return published content only.

Run `make test` and `make vet` for checks. Configure `ADDR` (host only), `APP_PORT`, `DB_PATH`, `COOKIE_SECURE`, `SESSION_TTL_HOURS`, `SEED_ADMIN_EMAIL`, and `SEED_ADMIN_PASSWORD` through environment variables or `.env`.
