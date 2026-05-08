# Deploy на Dokploy

Eye of Providence деплоится через Dokploy как 4 сервиса:

```
┌──────────────┐  https   ┌──────────────┐
│  dashboard   │ ───────► │  api         │
│  (Caddy SPA) │          │  (Go/Fiber)  │
└──────────────┘          └──┬───────────┘
                             │
                ┌────────────┼────────────┐
                ▼            ▼            ▼
         ┌──────────┐ ┌──────────┐ ┌─────────┐
         │ postgres │ │ clickhouse│ │ redis   │
         └──────────┘ └──────────┘ └─────────┘
```

## 0. Prerequisites

- Сервер с установленным Dokploy (https://docs.dokploy.com/docs/core/installation).
- Домены, направленные на сервер (например `eop-api.rysdavletov.org`, `eop-dash.rysdavletov.org`).
- Этот репозиторий, форкнутый или клонированный к тебе на GitHub.

## 1. Postgres (Database)

В Dokploy → **Databases** → **+ Create Database** → **PostgreSQL**.

- Name: `eop-postgres`
- Database: `eop`
- User: `eop`
- Password: сгенерируй (`openssl rand -hex 16`) и сохрани в password manager.
- Version: `16`

После создания запусти. Запиши **Connection URL** — пригодится в шаге 4.

## 2. ClickHouse (Database)

Dokploy не имеет встроенного ClickHouse template, но поддерживает **Compose** деплой:

→ **Compose** → **+ Create Compose** → имя `eop-clickhouse` → вставь `docker-compose.yml`:

```yaml
services:
  clickhouse:
    image: clickhouse/clickhouse-server:24.10-alpine
    environment:
      CLICKHOUSE_DB: eop
      CLICKHOUSE_USER: eop
      CLICKHOUSE_PASSWORD: ${CH_PASSWORD}
      CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT: 1
    volumes:
      - clickhouse_data:/var/lib/clickhouse
    ulimits:
      nofile: 262144
    expose:
      - "9000"
      - "8123"

volumes:
  clickhouse_data:
```

В Environment добавь `CH_PASSWORD=` (свой пароль). Deploy.

После запуска внутренний адрес: `eop-clickhouse-clickhouse-1:9000`. Точное имя — в Dokploy → Container.

## 3. Redis (Database)

→ **Databases** → **+ Create Database** → **Redis**. Name `eop-redis`. Запиши URL.

## 4. API (Application)

→ **Applications** → **+ Create Application** → **Dockerfile**.

- **Source**: Git → твой репо, branch `main`.
- **Build Path**: `.`  (корень репо — там лежит `api.Dockerfile`)
- **Dockerfile**: `api.Dockerfile`
- **Port**: `8080`
- **Domain**: `eop-api.rysdavletov.org` (Dokploy auto-привяжет Let's Encrypt).

**Environment Variables:**

```
EOP_ENV=production
EOP_HTTP_ADDR=:8080
EOP_POSTGRES_DSN=postgres://eop:PASSWORD@eop-postgres-db-1:5432/eop?sslmode=disable
EOP_CLICKHOUSE_DSN=clickhouse://eop:CH_PASSWORD@eop-clickhouse-clickhouse-1:9000/eop
EOP_REDIS_ADDR=eop-redis-db-1:6379
EOP_JWT_SECRET=<openssl rand -hex 32>      # ≥32 chars, иначе API не стартует в production
EOP_ALLOWED_ORIGINS=https://eop-dash.rysdavletov.org
EOP_AUTO_MIGRATE=true                       # API сам прогонит SQL-миграции на старте
EOP_INVITE_ONLY=true                        # регистрация только по invite (первый user — bootstrap)
EOP_ENABLE_DEV_TOKEN=false                  # никогда не включать в production
EOP_GEMINI_API_KEY=<https://aistudio.google.com/apikey>
EOP_GITHUB_CLIENT_ID=<опц., GitHub OAuth>
EOP_GITHUB_CLIENT_SECRET=<опц.>
EOP_GITHUB_CALLBACK_URL=https://eop-api.rysdavletov.org/v1/auth/github/callback
EOP_REPORTS_CRON_SEC=21600
```

> 🔒 **Production guard rails:** API падает на старте, если `EOP_JWT_SECRET` пуст / равен default / короче 32 символов, или `EOP_ALLOWED_ORIGINS=*`, или DSN'ы пустые. Это защита от случайного дефолтного secret в проде.

> ⚠️ Точные имена hostname'ов postgres/clickhouse/redis смотри в Dokploy → Container → внутренней сетке. Dokploy создаёт DNS вида `<app-name>-<service>-<index>`.

Deploy.

## 5. Миграции

API **сам прогонит** все SQL-миграции при старте, если `EOP_AUTO_MIGRATE=true` (default). Они embed'нуты в бинарь, идемпотентны (`CREATE … IF NOT EXISTS`).

Если хочешь применять вручную — выставь `EOP_AUTO_MIGRATE=false`:

```bash
# Postgres
psql -h <host> -U eop -d eop -f backend/migrations/001_init.up.sql
psql -h <host> -U eop -d eop -f backend/migrations/002_teams_auth.up.sql
psql -h <host> -U eop -d eop -f backend/migrations/003_multi_team_projects.up.sql

# ClickHouse
clickhouse-client --host <host> --user eop --password <CH_PASSWORD> --database eop \
  --multiquery < backend/migrations/clickhouse_001_init.sql
```

## 6. Dashboard (Application)

→ **Applications** → **+ Create Application** → **Dockerfile**.

- **Source**: тот же git repo.
- **Build Path**: `.` (КОРЕНЬ репо — нужно для ui/ workspace).
- **Dockerfile**: `dashboard.Dockerfile`
- **Port**: `3000` (Nginx внутри слушает 3000).
- **Domain**: `eop-dash.rysdavletov.org`

**Build Arguments** (важно!):
```
VITE_BACKEND_URL=https://eop-api.rysdavletov.org
CSP_CONNECT_SRC=https://eop-api.rysdavletov.org
```

`VITE_BACKEND_URL` — это compile-time (иначе dashboard будет ходить в localhost).
`CSP_CONNECT_SRC` — добавляется в Content-Security-Policy header'ы Caddy, чтобы fetch() из дашборда был разрешён к API origin.

**Environment Variables**: пустой, статике ничего не нужно.

Deploy.

## 7. Проверка

```bash
curl https://eop-api.rysdavletov.org/healthz
# {"service":"api","status":"ok"}

curl -X POST https://eop-api.rysdavletov.org/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"hunter2hunter2","display_name":"Test"}'
# {"token":"...", "user_id":"..."}
```

Открой `https://eop-dash.rysdavletov.org` → Регистрация → Создание команды.

## 8. Создать super_admin

После первой регистрации:

```bash
psql -h <host> -U eop -d eop -c "UPDATE users SET global_role='super_admin' WHERE email='you@example.com';"
```

## 9. Auto-deploy

В Dokploy → Application → **Deployments** → **Auto Deploy** → подключи GitHub webhook. После каждого `git push` в `main` сервис передеплоится автоматически.

## 10. Backups

- **Postgres**: Dokploy → eop-postgres → **Backups** → Schedule + Provider (S3/local).
- **ClickHouse**: руками через `clickhouse-backup` или volume snapshot.

## 11. Update flow

После каждого `git push`:
- Если auto-deploy включен — сам обновится через 1-2 минуты.
- Иначе: Dokploy → Application → **Deploy** (пересоберёт и перезапустит).

Миграции — добавляй НОВЫЕ файлы (`004_xxx.up.sql`), они идемпотентны через `IF NOT EXISTS`. Применяются вручную или через init-job (можно завести отдельный one-shot service в Compose).

## 12. Locally test the same Docker images

```bash
make docker-build VITE_BACKEND_URL=http://localhost:8080
docker run -p 8080:8080 \
  -e EOP_POSTGRES_DSN=... \
  -e EOP_CLICKHOUSE_DSN=... \
  eop-api:latest
docker run -p 8081:3000 eop-dashboard:latest
```

Открой `http://localhost:8081`.

## Troubleshooting

- **502 от Dashboard** → проверь что `VITE_BACKEND_URL` build arg правильный (`docker history eop-dashboard:latest | grep VITE`).
- **CORS ошибки** → `EOP_ALLOWED_ORIGINS` должен включать твой dashboard-домен.
- **API не подключается к БД** → имена hostnames внутренней сетки Dokploy специфичны. Зайди в backend контейнер `Console` → `nslookup eop-postgres-db-1`.
- **Миграции не применились** → API стартанёт, но `register` упадёт на FK. Применяй миграции вручную.
