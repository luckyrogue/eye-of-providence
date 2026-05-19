# Self-hosting

## Inicio rápido (un host, Docker)

```bash
git clone https://github.com/luckyrogue/eye-of-providence.git
cd eye-of-providence

# 1. Configurar secretos — NO deje CHANGE_ME en .env
cp .env.example .env
# edite POSTGRES_PASSWORD, CLICKHOUSE_PASSWORD, EOP_JWT_SECRET, EOP_ALLOWED_ORIGINS

# 2. (Opcional pero recomendado) — verificar firma de imagen antes del despliegue.
# Véase .github/SECURITY.md → "Verifying release artifacts". También
# fije un SHA concreto en .env: EOP_IMAGE=ghcr.io/luckyrogue/eop:<sha>

# 3. Levantar el stack completo (postgres + clickhouse + redis + imagen eop unificada)
docker compose -f infra/docker-compose.full.yml up -d
```

Panel + API en `http://localhost:3000` (puerto vía
`EOP_PUBLIC_PORT` en `.env`). Ponga un proxy inverso con TLS delante en
producción (Caddy / Traefik / nginx).

Las migraciones se aplican al arrancar la API (`EOP_AUTO_MIGRATE=true`).
Para control manual — `EOP_AUTO_MIGRATE=false` y ejecute
`docker exec eop-app /usr/local/bin/migrate` por separado.

## Configuración (env)

| Variable | Descripción | Por defecto |
|---|---|---|
| `EOP_ENV` | `development` o `production` | `development` |
| `EOP_HTTP_ADDR` | dirección de escucha | `:8080` |
| `EOP_POSTGRES_DSN` | DSN Postgres; vacío = fallback en memoria | `postgres://eop:eop_dev@localhost:5432/eop?sslmode=disable` |
| `EOP_CLICKHOUSE_DSN` | DSN ClickHouse; vacío = fallback en memoria | `clickhouse://eop:eop_dev@localhost:9000/eop` |
| `EOP_REDIS_ADDR` | Redis: caché analytics, challenges WebAuthn (requerido en compose full) | `localhost:6379` |
| `EOP_JWT_SECRET` | secreto para firmar JWT | solo dev, **cambiar en producción** |
| `EOP_GEMINI_API_KEY` | clave Google AI Studio; vacío = modo mock (dev) | vacío |
| `EOP_GITHUB_CLIENT_ID` | app OAuth GitHub | vacío |
| `EOP_GITHUB_CLIENT_SECRET` | app OAuth GitHub | vacío |
| `EOP_REPORTS_CRON_SEC` | frecuencia del cron semanal, 0 = apagado | 0 (en `docker-compose.full.yml` — 21600 = 6 h) |

## Claves

- **Gemini**: https://aistudio.google.com/apikey → crear API key. Sin ella los informes van en modo mock.
- **GitHub OAuth**: https://github.com/settings/developers → New OAuth App. URL de callback = `https://YOUR-DOMAIN/v1/auth/github/callback`.

## Frontend

`dashboard/dist/` tras `pnpm -F @eop/dashboard build` — estáticos en cualquier CDN o nginx. Por defecto llama a `http://localhost:8080`; en producción defina `VITE_BACKEND_URL` al compilar.

## Lista de producción

- [ ] `EOP_JWT_SECRET` — largo (≥ 32 bytes), aleatorio (`openssl rand -hex 32`).
- [ ] `POSTGRES_PASSWORD` / `CLICKHOUSE_PASSWORD` — no CHANGE_ME ni valores por defecto.
- [ ] HTTPS (proxy inverso: nginx, Caddy, Traefik).
- [ ] Fijar imagen — `EOP_IMAGE=ghcr.io/luckyrogue/eop:<sha>` en lugar de `:latest`.
- [ ] Verificación de imagen — cosign + procedencia SLSA (`.github/SECURITY.md`).
- [ ] Copias Postgres (`pg_dump` cron en volumen `postgres_data`).
- [ ] Nivel de almacenamiento ClickHouse (TTL 18 meses ya en tabla `events`).
- [ ] Rate-limit en proxy delante de `/v1/ingest` — el backend
      limita 120 req/min, pero el borde protege de agotamiento.
- [ ] `EOP_INVITE_ONLY=true` si no quiere registro público.
- [ ] `EOP_ENABLE_DEV_TOKEN=false` — tokens de depuración desactivados.

## Eliminación de usuario

`DELETE /v1/me/data` (con Bearer token):
- Borra todos los eventos en ClickHouse (`ALTER ... DELETE WHERE user_id = ?`).
- Borra informes + fila de usuario + tablas relacionadas en Postgres.
- Devuelve `{"status": "ok", "deleted_user": "<uuid>"}`.

Panel: Settings → Danger zone → Delete all my data.
