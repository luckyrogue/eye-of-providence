# Production playbook — Eye of Providence

Минимальный runbook для продакшена. Кратко: что развернуто, как обновлять, как восстанавливать.

## Topology

```
┌──────────────────────────────────────────────┐
│  Hetzner VM (Dokploy)                        │
│                                              │
│  ┌────────────────────────────────────────┐  │
│  │  combined image (ghcr.io/.../eop)      │  │
│  │   ├─ nginx :3000  (SPA + /api proxy)   │  │
│  │   └─ Go API :8080 (cmd/api)            │  │
│  └────────────────────────────────────────┘  │
│                                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ Postgres │  │  Redis   │  │  Traefik │   │
│  │ (self)   │  │ (self)   │  │  (DPL)   │   │
│  └──────────┘  └──────────┘  └──────────┘   │
└──────────────────────────────────────────────┘
                  │
                  ▼ TLS
        ┌────────────────────┐
        │ ClickHouse Cloud   │
        └────────────────────┘
```

## Required env-vars (Dokploy → Service → Environment)

| Var | Описание | Пример |
|---|---|---|
| `EOP_ENV` | Должно быть `production` | `production` |
| `EOP_JWT_SECRET` | Случайный 64+ char hex. **НИКОГДА** default. | `openssl rand -hex 64` |
| `EOP_POSTGRES_DSN` | DSN до self-hosted PG | `postgres://eop:STRONG_PASS@eop-postgres:5432/eop?sslmode=disable` |
| `EOP_CLICKHOUSE_DSN` | ClickHouse Cloud с `?secure=true` | `clickhouse://default:KEY@xxx.clickhouse.cloud:9440/default?secure=true` |
| `EOP_REDIS_ADDR` | host:port | `eop-redis:6379` |
| `EOP_ALLOWED_ORIGINS` | Точные origins, comma-separated | `https://app.example.com` |
| `EOP_GITHUB_CLIENT_ID` | OAuth | `Iv1...` |
| `EOP_GITHUB_CLIENT_SECRET` | OAuth | `...` |
| `EOP_GITHUB_CALLBACK_URL` | Должно совпадать с тем что в GitHub OAuth app | `https://app.example.com/api/v1/auth/github/callback` |
| `EOP_GEMINI_API_KEY` | Для cron-репортов | `...` |
| `EOP_AUTO_MIGRATE` | На prod = `true` (advisory_lock защищает от race) | `true` |
| `EOP_BETA_TEAM_LIMIT` | Sanity-cap | `3` |
| `EOP_INVITE_ONLY` | `true` пока бета | `true` |
| `EOP_ENABLE_DEV_TOKEN` | **Должно быть `false`** на prod | `false` |
| `EOP_BODY_LIMIT_BYTES` | Дефолт 1 MiB; ingest батчам нужно больше | `5242880` |

## Required GitHub Actions secrets / vars

| Name | Тип | Что | Пример |
|---|---|---|---|
| `DOKPLOY_WEBHOOK` | secret | URL для триггера redeploy | `https://dokploy.../api/deploy/...` |
| `VITE_BACKEND_URL` | var | путь до API из браузера | `/api` |
| `CSP_CONNECT_SRC` | var | CSP для статики | `'self'` |
| `HEALTH_URL` | var | URL `/healthz` для post-deploy verification | `https://app.example.com/healthz` |

Если `HEALTH_URL` не задан, deploy job не ждёт healthcheck (curl webhook → exit). Установка → CI получает реальную обратную связь о деплое.

## Deploy flow

1. `git push` → main
2. CI: `backend-lint` → `backend` → `frontend` → `agent-rust` → `docker` → `deploy`
3. `docker` job: build, push в GHCR, Trivy scan (CRITICAL = fail)
4. `deploy` job: `curl POST $DOKPLOY_WEBHOOK` → polls `$HEALTH_URL` до 200 (5min timeout)
5. Если healthcheck timed out → CI красный, runbook'а нет, **смотрим логи в Dokploy**

## Rollback

Через Dokploy UI:
1. Open service → Deployments → выбрать предыдущий tag (`ghcr.io/.../eop:<prev-sha>`)
2. Redeploy

Через CLI на VM:
```bash
docker tag ghcr.io/<owner>/eop:<prev-sha> ghcr.io/<owner>/eop:latest
docker service update --force eop_app
```

## Postgres backups

Self-hosted PG = единственная копия данных users/teams/subscriptions. **Обязателен** offsite backup.

Setup:
```bash
# На VM:
sudo cp /path/to/repo/infra/backup-postgres.sh /usr/local/bin/eop-backup-postgres
sudo chmod +x /usr/local/bin/eop-backup-postgres

# Конфиг:
sudo tee /etc/eop-backup.env <<EOF
EOP_PG_CONTAINER=eop-postgres
EOP_PG_USER=eop
EOP_PG_DB=eop
EOP_BACKUP_DIR=/var/backups/eop
EOP_BACKUP_RETENTION=14
# Опционально: rclone до B2/Hetzner Storage Box
EOP_BACKUP_REMOTE=b2:eop-backups
EOF
sudo chmod 600 /etc/eop-backup.env

# Cron (root):
sudo crontab -e
17 3 * * *  /usr/local/bin/eop-backup-postgres >> /var/log/eop-backup.log 2>&1

# Тест прямо сейчас:
sudo /usr/local/bin/eop-backup-postgres
ls -lh /var/backups/eop/
```

Restore:
```bash
gunzip < /var/backups/eop/pg-eop-YYYYMMDDTHHMMSSZ.sql.gz \
  | docker exec -i eop-postgres psql -U eop eop
```

ClickHouse Cloud делает backup за тебя (default 24h retention; за деньгами больше).

## Migrations (golang-migrate)

Schema versioning через `golang-migrate/migrate/v4` поверх `schema_migrations` таблицы.

**Up** делает API при старте (если `EOP_AUTO_MIGRATE=true`). Безопасен на rolling deploy: golang-migrate берёт PG advisory_lock, остальные replica'и ждут.

**Down/force/version** — только через CLI:

```bash
# В контейнере (Dokploy → exec):
docker exec -it eop_app sh
# /usr/local/bin/migrate уже встроен в combined image
migrate -db postgres version
migrate -db postgres up
migrate -db postgres down 1            # шаг назад на 1 миграцию
migrate -db postgres goto 3            # привести к версии 3
EOP_MIGRATE_CONFIRM=yes-i-mean-it migrate -db postgres down all
```

**Первый deploy этой версии на старую prod-БД:** прежний idempotent-runner не вёл `schema_migrations`. Нужно один раз вручную пометить состояние:

```bash
migrate -db postgres force 5         # помечает 1..5 как applied без запуска SQL
migrate -db clickhouse force 2       # 001_events + 002_attribution_events
```

После этого auto-migrate при старте API увидит, что всё applied, и будет no-op.

**Откат прод-миграции:** только при идентифицированной проблеме, всегда после backup'а:
1. `migrate -db postgres version` — узнать current
2. `infra/backup-postgres.sh` — снять дамп
3. `migrate -db postgres down 1` — шаг назад
4. Передеплоить старый image (`Rollback` в Dokploy)

## Healthcheck semantics

`GET /healthz` через nginx → Go API:
- `200 {"status":"ok","postgres":"ok","clickhouse":"ok"}` — всё хорошо
- `503 {"status":"degraded",...}` — PG или CH вернули ошибку при ping (с 2s timeout)

Dockerfile `HEALTHCHECK` пробивает `127.0.0.1:3000/healthz` каждые 15s. start-period=60s даёт ClickHouse Cloud TLS-handshake'у и миграциям закончиться.

## Troubleshooting

| Симптом | Где смотреть | Что обычно |
|---|---|---|
| `503` после deploy | `docker logs eop_app` | Проверить env-vars, особенно DSN'ы |
| Restart-loop сразу после старта | `docker logs eop_app` | Чаще всего — сменился JWT_SECRET и старые сессии **не валидны** (это OK, юзеры перелогинятся). Или migration упала |
| Slow `/v1/teams/<id>/summary` | `docker logs eop_app` + ClickHouse Cloud query log | Часто N+1 — теперь должен быть один query, проверь |
| Auth ошибки `invalid token subject` | — | После смены схемы JWT все старые токены инвалидны. Это by design (revocation hardening) |

## Что НЕ делать

- `docker compose -f infra/docker-compose.yml up` на prod-VM — это dev-конфиг с публичными паролями (порты bind'ятся на 127.0.0.1, но всё равно)
- Коммитить `EOP_JWT_SECRET` в репо
- Сетить `EOP_ENABLE_DEV_TOKEN=true` на prod — открывает `/v1/auth/dev-token` без auth
- `git push --force` на main — CI dispatcher отправит redeploy на тот же sha-тэг
