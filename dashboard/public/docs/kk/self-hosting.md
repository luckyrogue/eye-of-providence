# Self-hosting

## Quickstart (бір host, Docker)

```bash
git clone https://github.com/luckyrogue/eye-of-providence.git
cd eye-of-providence

# 1. Секреттерді баптау — .env ішінде CHANGE_ME қалдырмаңыз
cp .env.example .env
# POSTGRES_PASSWORD, CLICKHOUSE_PASSWORD, EOP_JWT_SECRET, EOP_ALLOWED_ORIGINS өңдеңіз

# 2. (Міндетті емес, бірақ ұсынылады) — deploy алдында image қолтаңбасын тексеру.
# Қараңыз .github/SECURITY.md → "Verifying release artifacts". Сондай-ақ
# .env-те нақты SHA бекітіңіз: EOP_IMAGE=ghcr.io/luckyrogue/eop:<sha>

# 3. Толық стекті көтеру (postgres + clickhouse + redis + unified eop image)
docker compose -f infra/docker-compose.full.yml up -d
```

Dashboard + API `http://localhost:3000` (порт `.env` ішіндегі
`EOP_PUBLIC_PORT`). Production үшін алдына TLS бар reverse proxy қойыңыз
(Caddy / Traefik / nginx).

Миграциялар API стартында автоматты (`EOP_AUTO_MIGRATE=true`).
Қолмен бақылау үшін — `EOP_AUTO_MIGRATE=false` қойып
`docker exec eop-app /usr/local/bin/migrate` бөлек іске қосыңыз.

## Конфигурация (env)

| Айнымалы | Сипаттама | Default |
|---|---|---|
| `EOP_ENV` | `development` немесе `production` | `development` |
| `EOP_HTTP_ADDR` | listen address | `:8080` |
| `EOP_POSTGRES_DSN` | Postgres DSN; бос = in-memory fallback | `postgres://eop:eop_dev@localhost:5432/eop?sslmode=disable` |
| `EOP_CLICKHOUSE_DSN` | ClickHouse DSN; бос = in-memory fallback | `clickhouse://eop:eop_dev@localhost:9000/eop` |
| `EOP_REDIS_ADDR` | Redis: analytics cache, WebAuthn challenge (full compose-та қажет) | `localhost:6379` |
| `EOP_JWT_SECRET` | JWT қолтаңбасы секреті | dev-only, **production-да міндетті өзгерту** |
| `EOP_GEMINI_API_KEY` | Google AI Studio кілті; бос = mock (dev) | бос |
| `EOP_GITHUB_CLIENT_ID` | GitHub OAuth app | бос |
| `EOP_GITHUB_CLIENT_SECRET` | GitHub OAuth app | бос |
| `EOP_REPORTS_CRON_SEC` | weekly cron жиілігі, 0 = өшіру | 0 (`docker-compose.full.yml` — 21600 = 6h) |

## Кілттер

- **Gemini**: https://aistudio.google.com/apikey → API key жасау. Онсыз reports mock режимінде.
- **GitHub OAuth**: https://github.com/settings/developers → New OAuth App. Callback URL = `https://YOUR-DOMAIN/v1/auth/github/callback`.

## Frontend

`dashboard/dist/` — `pnpm -F @eop/dashboard build` кейін статика, кез келген CDN/nginx. Әдепкі `http://localhost:8080`; production-да build кезінде `VITE_BACKEND_URL`.

## Production checklist

- [ ] `EOP_JWT_SECRET` — ұзын (≥ 32 байт), кездейсоқ (`openssl rand -hex 32`).
- [ ] `POSTGRES_PASSWORD` / `CLICKHOUSE_PASSWORD` — CHANGE_ME емес, default емес.
- [ ] HTTPS (reverse proxy: nginx, Caddy, Traefik).
- [ ] Image pinning — `EOP_IMAGE=ghcr.io/luckyrogue/eop:<sha>` `:latest` орнына.
- [ ] Image verification — cosign + SLSA provenance (`.github/SECURITY.md`).
- [ ] Postgres backup (`pg_dump` cron, `postgres_data` volume).
- [ ] ClickHouse storage tier (`events` кестесінде TTL 18 ай).
- [ ] Reverse-proxy rate-limit `/v1/ingest` алдында — backend
      өзі 120 req/min, бірақ edge API exhaustion-нан қорғайды.
- [ ] `EOP_INVITE_ONLY=true` жария тіркелуді қаламасаңыз.
- [ ] `EOP_ENABLE_DEV_TOKEN=false` — debug token өшірілгенін растаңыз.

## Пайдаланушыны өшіру

`DELETE /v1/me/data` (Bearer token):
- ClickHouse events өшіреді (`ALTER ... DELETE WHERE user_id = ?`).
- Postgres reports + user + байланысты кестелер.
- `{"status": "ok", "deleted_user": "<uuid>"}` қайтарады.

Dashboard: Settings → Danger zone → Delete all my data.
