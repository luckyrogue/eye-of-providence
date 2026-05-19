# Self-hosting

## Quickstart (single host, Docker)

```bash
git clone https://github.com/luckyrogue/eye-of-providence.git
cd eye-of-providence

# 1. Configure secrets — do NOT leave CHANGE_ME in .env
cp .env.example .env
# edit POSTGRES_PASSWORD, CLICKHOUSE_PASSWORD, EOP_JWT_SECRET, EOP_ALLOWED_ORIGINS

# 2. (Optional but recommended) — verify image signature before deploy.
# See .github/SECURITY.md → "Verifying release artifacts". Also
# pin a specific SHA in .env: EOP_IMAGE=ghcr.io/luckyrogue/eop:<sha>

# 3. Start the full stack (postgres + clickhouse + redis + unified eop image)
docker compose -f infra/docker-compose.full.yml up -d
```

Dashboard + API are available at `http://localhost:3000` (port via
`EOP_PUBLIC_PORT` in `.env`). Put a reverse proxy with TLS in front for
production (Caddy / Traefik / nginx).

Migrations run automatically on API start (`EOP_AUTO_MIGRATE=true`).
For manual control — set `EOP_AUTO_MIGRATE=false` and run
`docker exec eop-app /usr/local/bin/migrate` separately.

## Configuration (env)

| Variable | Description | Default |
|---|---|---|
| `EOP_ENV` | `development` or `production` | `development` |
| `EOP_HTTP_ADDR` | listen address | `:8080` |
| `EOP_POSTGRES_DSN` | Postgres DSN; empty = in-memory fallback | `postgres://eop:eop_dev@localhost:5432/eop?sslmode=disable` |
| `EOP_CLICKHOUSE_DSN` | ClickHouse DSN; empty = in-memory fallback | `clickhouse://eop:eop_dev@localhost:9000/eop` |
| `EOP_REDIS_ADDR` | Redis: analytics cache, WebAuthn challenge store (required in full compose) | `localhost:6379` |
| `EOP_JWT_SECRET` | secret for JWT signing | dev-only secret, **must change in production** |
| `EOP_GEMINI_API_KEY` | Google AI Studio key; empty = mock mode (dev) | empty |
| `EOP_GITHUB_CLIENT_ID` | GitHub OAuth app | empty |
| `EOP_GITHUB_CLIENT_SECRET` | GitHub OAuth app | empty |
| `EOP_REPORTS_CRON_SEC` | how often to run weekly cron, 0 = off | 0 (in `docker-compose.full.yml` — 21600 = 6h) |

## Keys

- **Gemini**: https://aistudio.google.com/apikey → create API key. Without it reports run in mock mode.
- **GitHub OAuth**: https://github.com/settings/developers → New OAuth App. Authorization callback URL = `https://YOUR-DOMAIN/v1/auth/github/callback`.

## Frontend

`dashboard/dist/` after `pnpm -F @eop/dashboard build` — static assets served by any CDN or nginx. By default calls `http://localhost:8080`; in production set `VITE_BACKEND_URL` at build time.

## Production checklist

- [ ] `EOP_JWT_SECRET` — long (≥ 32 bytes), random (`openssl rand -hex 32`).
- [ ] `POSTGRES_PASSWORD` / `CLICKHOUSE_PASSWORD` — not CHANGE_ME, not defaults.
- [ ] HTTPS (reverse proxy: nginx, Caddy, Traefik).
- [ ] Image pinning — `EOP_IMAGE=ghcr.io/luckyrogue/eop:<sha>` instead of `:latest`.
- [ ] Image verification — cosign + SLSA provenance (see `.github/SECURITY.md`).
- [ ] Postgres backups (`pg_dump` cron on named volume `postgres_data`).
- [ ] ClickHouse storage tier (TTL already set to 18 months on `events` table).
- [ ] Reverse-proxy rate-limit (Caddy/Traefik) in front of `/v1/ingest` — backend
      limits 120 req/min itself, but edge layer protects the API from exhaustion.
- [ ] `EOP_INVITE_ONLY=true` if you do not want public registration.
- [ ] `EOP_ENABLE_DEV_TOKEN=false` — ensure debug tokens are disabled.

## User deletion

`DELETE /v1/me/data` (with Bearer token):
- Deletes all events from ClickHouse (`ALTER ... DELETE WHERE user_id = ?`).
- Deletes reports + user row + related tables from Postgres.
- Returns `{"status": "ok", "deleted_user": "<uuid>"}`.

Dashboard: Settings → Danger zone → Delete all my data.
