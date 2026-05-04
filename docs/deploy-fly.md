# Deploy на Fly.io

Целевая топология:

```
┌──────────────────────┐      ┌────────────────────┐
│  eop-dashboard       │      │  eop-api           │
│  (Caddy + React SPA) │ ───► │  (Go + Fiber)      │
└──────────────────────┘      └─────────┬──────────┘
                                        │ .flycast (private)
                          ┌─────────────┼─────────────┐
                          ▼             ▼             ▼
                  ┌──────────────┐ ┌──────────────┐ (Redis — V1)
                  │  eop-postgres│ │ eop-clickhouse│
                  │  (managed)   │ │  + volume     │
                  └──────────────┘ └──────────────┘
```

Секреты — через `fly secrets set` (зашифрованы at-rest, available только в runtime).

---

## 0. Prerequisites

```bash
# Fly CLI
curl -L https://fly.io/install.sh | sh

# Логин (откроется браузер)
fly auth login

# Опционально: создать org для команды (по умолчанию personal)
# fly orgs create acme
```

## 1. Один скрипт (если не хочешь читать)

```bash
cd infra/fly
FLY_ORG=personal FLY_REGION=ams ./setup.sh
```

Скрипт спросит пароль ClickHouse, JWT secret, опц. Gemini ключ и GitHub OAuth, всё остальное сделает сам. Дальше — пропусти до §6.

## 2. Postgres (managed)

```bash
fly postgres create \
  --name eop-postgres \
  --region ams \
  --initial-cluster-size 1 \
  --vm-size shared-cpu-1x \
  --volume-size 1
```

Запиши пароль admin user'а из вывода — пригодится. Применить миграции:

```bash
fly postgres connect -a eop-postgres -d eop \
  < backend/migrations/001_init.up.sql
```

## 3. ClickHouse (как отдельный Fly app)

```bash
fly apps create eop-clickhouse
fly volumes create clickhouse_data --app eop-clickhouse --region ams --size 5
fly secrets set --app eop-clickhouse CLICKHOUSE_PASSWORD="..."

cd infra/fly/clickhouse
fly deploy --app eop-clickhouse

# Применить миграции
fly ssh console -a eop-clickhouse -C "clickhouse-client --user eop --password $CH_PASS --database eop --multiquery" \
  < backend/migrations/clickhouse_001_init.sql
```

Note: ClickHouse слушает на `eop-clickhouse.flycast:9000` (TCP) и `:8123` (HTTP) — internal-only.

## 4. API

```bash
fly apps create eop-api
fly postgres attach eop-postgres --app eop-api    # выставит DATABASE_URL secret

fly secrets set --app eop-api \
  EOP_POSTGRES_DSN="postgres://eop:PASS@eop-postgres.flycast:5432/eop?sslmode=disable" \
  EOP_CLICKHOUSE_DSN="clickhouse://eop:CH_PASS@eop-clickhouse.flycast:9000/eop" \
  EOP_JWT_SECRET="$(openssl rand -hex 32)" \
  EOP_GEMINI_API_KEY="..." \
  EOP_GITHUB_CLIENT_ID="..." \
  EOP_GITHUB_CLIENT_SECRET="..."

cd backend
fly deploy --app eop-api
```

API будет доступен на `https://eop-api.fly.dev`. Healthcheck: `/healthz`.

## 5. Dashboard

Dockerfile собирается из корня монорепо (нужен доступ к `ui/` пакету):

```bash
fly apps create eop-dashboard
fly deploy -c dashboard/fly.toml -a eop-dashboard \
  --build-arg VITE_BACKEND_URL=https://eop-api.fly.dev .
```

Dashboard будет на `https://eop-dashboard.fly.dev`.

Не забудь обновить CORS в backend если хочешь чтобы dashboard ходил с production-домена:

```go
// backend/cmd/api/main.go
AllowOrigins: "https://eop-dashboard.fly.dev"
```

(или сделать env-driven, V1).

## 6. GitHub OAuth callback URL

В https://github.com/settings/developers → твоя OAuth App → Authorization callback URL:

```
https://eop-api.fly.dev/v1/auth/github/callback
```

## 7. Verify

```bash
curl https://eop-api.fly.dev/healthz
# {"service":"api","status":"ok"}

curl -X POST https://eop-api.fly.dev/v1/auth/dev-token
# {"token": "...", "user_id": "..."}

# Open dashboard:
open https://eop-dashboard.fly.dev
```

## 8. Логи и метрики

```bash
fly logs -a eop-api
fly logs -a eop-clickhouse
fly status -a eop-api

# Metrics (Prometheus)
curl https://eop-api.fly.dev/metrics
```

Fly автоматически собирает скрейпит `/metrics` если в `fly.toml` есть `[metrics]` секция (уже есть в backend/fly.toml). Графики — в Fly dashboard → Monitoring.

## 9. Production checklist

- [ ] **Custom domain**: `fly certs add eop.dev --app eop-dashboard`.
- [ ] **CORS**: env `EOP_ALLOWED_ORIGINS=https://eop.dev`, а не hardcoded localhost.
- [ ] **dev-token endpoint**: убедиться что `EOP_ENV=production` его отключает (V1 TODO).
- [ ] **Backups**: Fly Postgres делает automated, но проверь schedule в Fly dashboard.
- [ ] **Volume snapshots для ClickHouse**: `fly volumes snapshots list -a eop-clickhouse`.
- [ ] **Scale**: `fly scale count 2 --app eop-api` если RPS вырастет (Fly размажет по machines).
- [ ] **Memory**: ClickHouse 1GB достаточно для ~1M events; для больше — `fly scale memory 4096 -a eop-clickhouse`.

## 10. Стоимость (примерная, на январь 2026)

| Компонент | Размер | $/мес |
|---|---|---|
| `eop-api` shared-cpu-1x 256MB | min 1 машина | ~$2 |
| `eop-postgres` shared-cpu-1x 256MB | 1GB volume | ~$2 |
| `eop-clickhouse` shared-cpu-1x 1GB | 5GB volume | ~$5 |
| `eop-dashboard` shared-cpu-1x 256MB | auto-stop когда нет трафика | ~$0.50 |
| **Итого** | | ~$10 |

Для прод-нагрузки придётся апгрейдить (Postgres → dedicated, ClickHouse → 4GB+, redundancy).

## 11. Troubleshooting

**"target machine not found"** — приложение не задеплоено или machine остановлена. `fly machine list -a eop-api` → `fly machine start <id>`.

**"connect: connection refused" с .flycast** — приложение существует, но машина не запущена. `auto_start_machines = true` решает большинство случаев.

**ClickHouse "Authentication failed"** — пароль не совпадает между secret и DSN в backend. Перепроверь:
```bash
fly secrets list -a eop-clickhouse
fly secrets list -a eop-api
```

**Migrations не применились** — нет idempotency, повторный запуск сломается на CREATE TABLE. Используй `IF NOT EXISTS` или применяй через goose/atlas (V1).
