# syntax=docker/dockerfile:1.6
# Объединённый production image: Go API + dashboard SPA в одном контейнере.
# Caddy раздаёт статику dashboard и проксирует /api/* на локальный Go-бинарь.
#
# Почему Caddy вместо nginx:
#   - Caddyfile читает env-vars нативно ({$VAR}), no envsubst step
#   - Auto-https off + reverse_proxy + file_server + headers — out-of-the-box
#   - encode gzip zstd — лучше чем nginx-only gzip
#   - SPA fallback через `try_files` без regex
#   - Меньше boilerplate, легче поддерживать
#
# Dokploy:
#   Provider: Docker
#   Docker Image: ghcr.io/<owner>/eop:latest
#   Port: 3000  (Dokploy terminates TLS, Caddy listen plain HTTP внутри)
#
# Локально:
#   docker build -t eop .
#   docker run --rm -p 3000:3000 \
#     -e EOP_POSTGRES_DSN=... -e EOP_CLICKHOUSE_DSN=... -e EOP_REDIS_ADDR=... \
#     -e EOP_JWT_SECRET=... -e EOP_ALLOWED_ORIGINS=https://example.org \
#     eop

############################
# Go API builder
############################
FROM golang:1.25-alpine AS api-builder
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend ./
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags='-s -w' \
    -o /out/api ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags='-s -w' \
    -o /out/migrate ./cmd/migrate

############################
# Dashboard builder
############################
FROM node:22-alpine AS dashboard-builder
WORKDIR /repo
RUN corepack enable && corepack prepare pnpm@9.12.3 --activate

COPY pnpm-workspace.yaml package.json pnpm-lock.yaml* ./
COPY ui/package.json ui/
COPY dashboard/package.json dashboard/
RUN pnpm install --frozen-lockfile

COPY ui ui
COPY dashboard dashboard

# В combined-режиме фронт ходит относительно — на тот же хост
ARG VITE_BACKEND_URL=/api
ENV VITE_BACKEND_URL=${VITE_BACKEND_URL}
RUN pnpm -F @eop/dashboard build

############################
# Final image: Caddy + Go API
############################
FROM caddy:2-alpine

# Подтягиваем последние patch-версии apk-пакетов (libcrypto3, libssl3, libxml2,
# zlib и пр.), даже если base-image отстал на пару дней. Trivy gates на
# CRITICAL — фиксят vendors через apk-репозиторий быстрее, чем в base-tag.
RUN apk upgrade --no-cache && rm -rf /var/cache/apk/*

# Go-бинарь + migrate CLI (для ручного rollback в проде)
COPY --from=api-builder /out/api /usr/local/bin/api
COPY --from=api-builder /out/migrate /usr/local/bin/migrate

# Статика dashboard (Caddy default site root)
COPY --from=dashboard-builder /repo/dashboard/dist /usr/share/caddy

ARG CSP_CONNECT_SRC="'self'"
ENV CSP_CONNECT_SRC=${CSP_CONNECT_SRC}

# Caddyfile — читается caddy при старте, env-vars подставляются native
# через {$VAR:default} syntax.
RUN cat > /etc/caddy/Caddyfile <<'CADDY'
{
	auto_https off
	admin off
	# Dokploy в нашей setup terminates TLS, поэтому Caddy listen plain HTTP.
	# Если когда-то развернёмся на bare metal — раскомментить TLS via Let's Encrypt.
}

:3000 {
	root * /usr/share/caddy
	encode gzip zstd

	# Security headers — match старый nginx setup.
	header {
		Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
		X-Frame-Options "DENY"
		X-Content-Type-Options "nosniff"
		Referrer-Policy "strict-origin-when-cross-origin"
		Permissions-Policy "camera=(), microphone=(), geolocation=(), interest-cohort=()"
		Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src {$CSP_CONNECT_SRC:'self'}; object-src 'none'; frame-ancestors 'none'; base-uri 'self';"
		# Удаляем Server header чтобы не палить Caddy version.
		-Server
	}

	# /api/v1/foo  →  http://127.0.0.1:8080/v1/foo. handle_path strips /api prefix.
	handle_path /api/* {
		reverse_proxy 127.0.0.1:8080 {
			header_up Host {host}
			header_up X-Real-IP {remote_host}
			header_up X-Forwarded-For {remote_host}
			header_up X-Forwarded-Proto {scheme}
			transport http {
				read_timeout 60s
			}
		}
	}

	# Прямой healthcheck Go-бинаря.
	handle /healthz {
		reverse_proxy 127.0.0.1:8080
	}

	# SPA fallback: для всех неизвестных URL — отдаём index.html.
	# file_server попробует $uri; если не нашёл, try_files направит на /index.html.
	handle {
		try_files {path} /index.html
		file_server
	}
}
CADDY

# Запускаем оба процесса. wait -n валит контейнер, если любой упал.
RUN cat > /entrypoint.sh <<'SH' && chmod +x /entrypoint.sh
#!/bin/sh
set -e

# Validate Caddyfile перед запуском (быстрый fail).
caddy validate --config /etc/caddy/Caddyfile

/usr/local/bin/api &
API_PID=$!

caddy run --config /etc/caddy/Caddyfile --adapter caddyfile &
CADDY_PID=$!

wait -n "$API_PID" "$CADDY_PID"
EXIT=$?

# Если умер один — добиваем второй и выходим
kill -TERM "$API_PID" "$CADDY_PID" 2>/dev/null || true
exit $EXIT
SH

EXPOSE 3000

# start-period=60s даёт миграциям и ClickHouse Cloud TLS-handshake'у
# завершиться до того как Swarm начнёт считать неудачные probe'ы.
# /healthz пробивает Caddy → Go API → проверяет PG+CH; возвращает 503 если
# что-то деградировано — Swarm перенесёт трафик на здоровый instance.
HEALTHCHECK --start-period=60s --interval=15s --timeout=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:3000/healthz >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/entrypoint.sh"]
