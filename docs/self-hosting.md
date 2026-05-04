# Self-hosting

## Quickstart (single host, Docker)

```bash
git clone https://github.com/luckyrogue/eye-of-providence.git
cd eye-of-providence

# Поднять весь стек (postgres + clickhouse + redis + api)
docker compose -f infra/docker-compose.full.yml up -d --build

# Применить миграции (один раз после первого запуска)
docker exec -i eop-postgres-1 psql -U eop -d eop < backend/migrations/001_init.up.sql
docker exec -i eop-clickhouse-1 clickhouse-client --user eop --password eop_dev \
  --database eop --multiquery < backend/migrations/clickhouse_001_init.sql
```

API будет доступен на `http://localhost:8080`.

## Конфигурация (env)

| Переменная | Описание | Default |
|---|---|---|
| `EOP_ENV` | `development` или `production` | `development` |
| `EOP_HTTP_ADDR` | listen address | `:8080` |
| `EOP_POSTGRES_DSN` | Postgres DSN; пусто = in-memory fallback | `postgres://eop:eop_dev@localhost:5432/eop?sslmode=disable` |
| `EOP_CLICKHOUSE_DSN` | ClickHouse DSN; пусто = in-memory fallback | `clickhouse://eop:eop_dev@localhost:9000/eop` |
| `EOP_REDIS_ADDR` | (зарезервировано — пока не используется) | `localhost:6379` |
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

- [ ] `EOP_JWT_SECRET` — длинный (≥ 32 байта), случайный.
- [ ] HTTPS (reverse proxy: nginx, Caddy, Traefik).
- [ ] Postgres backups (`pg_dump` cron).
- [ ] ClickHouse storage tier (TTL уже выставлен на 18 мес в `events` таблице).
- [ ] Метрики: scrape `/metrics` — пока не выставлены, добавить в Phase 7.5.
- [ ] Rate limit на ingest endpoint — Phase 7.5.

## Удаление пользователя

`DELETE /v1/me/data` (с Bearer token):
- Стирает все события из ClickHouse (`ALTER ... DELETE WHERE user_id = ?`).
- Стирает reports + user row + связанные таблицы из Postgres.
- Возвращает `{"status": "ok", "deleted_user": "<uuid>"}`.

Дашборд: Settings → Danger zone → Delete all my data.
