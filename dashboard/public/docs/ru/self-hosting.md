# Self-hosting

## Quickstart (single host, Docker)

```bash
git clone https://github.com/luckyrogue/eye-of-providence.git
cd eye-of-providence

# 1. Настроить секреты — НЕ оставляй CHANGE_ME в .env
cp .env.example .env
# отредактируй POSTGRES_PASSWORD, CLICKHOUSE_PASSWORD, EOP_JWT_SECRET, EOP_ALLOWED_ORIGINS

# 2. (Опционально, но рекомендовано) — verify image signature перед deploy'ем.
# См. .github/SECURITY.md → "Verifying release artifacts". Заодно
# зафиксируй конкретный SHA в .env: EOP_IMAGE=ghcr.io/luckyrogue/eop:<sha>

# 3. Поднять весь стек (postgres + clickhouse + redis + unified eop image)
docker compose -f infra/docker-compose.full.yml up -d
```

Дашборд + API доступны на `http://localhost:3000` (порт меняется через
`EOP_PUBLIC_PORT` в `.env`). Поставь reverse proxy с TLS перед ним для
production (Caddy / Traefik / nginx).

Миграции применяются автоматически на старте API (`EOP_AUTO_MIGRATE=true`).
Если хочешь ручной контроль — выстави `EOP_AUTO_MIGRATE=false` и запусти
`docker exec eop-app /usr/local/bin/migrate` отдельно.

## Конфигурация (env)

| Переменная | Описание | Default |
|---|---|---|
| `EOP_ENV` | `development` или `production` | `development` |
| `EOP_HTTP_ADDR` | listen address | `:8080` |
| `EOP_POSTGRES_DSN` | Postgres DSN; пусто = in-memory fallback | `postgres://eop:eop_dev@localhost:5432/eop?sslmode=disable` |
| `EOP_CLICKHOUSE_DSN` | ClickHouse DSN; пусто = in-memory fallback | `clickhouse://eop:eop_dev@localhost:9000/eop` |
| `EOP_REDIS_ADDR` | Redis: analytics cache, WebAuthn challenge store (required в full compose) | `localhost:6379` |
| `EOP_JWT_SECRET` | секрет для подписи JWT | dev-only-секрет, **обязательно поменять в production** |
| `EOP_GEMINI_API_KEY` | ключ Google AI Studio; пусто = mock-режим (dev) | пусто |
| `EOP_GITHUB_CLIENT_ID` | GitHub OAuth app | пусто |
| `EOP_GITHUB_CLIENT_SECRET` | GitHub OAuth app | пусто |
| `EOP_REPORTS_CRON_SEC` | как часто прогонять weekly cron, 0 = выкл | 0 (в `docker-compose.full.yml` — 21600 = 6h) |

## Ключи

- **Gemini**: https://aistudio.google.com/apikey → создать API key. Без него reports работают в mock-режиме.
- **GitHub OAuth**: https://github.com/settings/developers → New OAuth App. Authorization callback URL = `https://YOUR-DOMAIN/v1/auth/github/callback`.

## Frontend

`dashboard/dist/` после `pnpm -F @eop/dashboard build` — статика, отдаётся любым CDN или nginx. По умолчанию ходит в `http://localhost:8080`; в production задать `VITE_BACKEND_URL` при сборке.

## Production checklist

- [ ] `EOP_JWT_SECRET` — длинный (≥ 32 байта), случайный (`openssl rand -hex 32`).
- [ ] `POSTGRES_PASSWORD` / `CLICKHOUSE_PASSWORD` — не CHANGE_ME, не дефолтные.
- [ ] HTTPS (reverse proxy: nginx, Caddy, Traefik).
- [ ] Image pinning — `EOP_IMAGE=ghcr.io/luckyrogue/eop:<sha>` вместо `:latest`.
- [ ] Image verification — cosign + SLSA provenance (см. `.github/SECURITY.md`).
- [ ] Postgres backups (`pg_dump` cron на named volume `postgres_data`).
- [ ] ClickHouse storage tier (TTL уже выставлен на 18 мес в `events` таблице).
- [ ] Reverse-proxy rate-limit (Caddy/Traefik) перед `/v1/ingest` — backend
      сам режет 120 req/min, но edge-layer защищает API от exhaustion.
- [ ] `EOP_INVITE_ONLY=true` если не хочешь публичной регистрации.
- [ ] `EOP_ENABLE_DEV_TOKEN=false` — гарантированно отключить debug-токены.

## Удаление пользователя

`DELETE /v1/me/data` (с Bearer token):
- Стирает все события из ClickHouse (`ALTER ... DELETE WHERE user_id = ?`).
- Стирает reports + user row + связанные таблицы из Postgres.
- Возвращает `{"status": "ok", "deleted_user": "<uuid>"}`.

Дашборд: Settings → Danger zone → Delete all my data.
