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
    -o /out/migrate ./cmd/migrate && \
    CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags='-s -w' \
    -o /out/worker ./cmd/worker

############################
# Dashboard builder
############################
FROM node:22-alpine AS dashboard-builder
WORKDIR /repo
RUN corepack enable && corepack prepare pnpm@9.12.3 --activate

COPY pnpm-workspace.yaml package.json pnpm-lock.yaml* tsconfig.base.json ./
COPY ui/package.json ui/
COPY i18n/package.json i18n/
COPY dashboard/package.json dashboard/
RUN pnpm install --frozen-lockfile

COPY ui ui
COPY i18n i18n
COPY dashboard dashboard

# В combined-режиме фронт ходит относительно — на тот же хост
ARG VITE_BACKEND_URL=/api
ENV VITE_BACKEND_URL=${VITE_BACKEND_URL}
RUN pnpm -F @eop/dashboard build

############################
# Caddy from-source rebuild
############################
# caddy:2.11-alpine binary upstream был built с Go 1.26.0 + outdated deps
# (otel 1.40, go-jose v4.1.3) — содержит 9+ HIGH CVE по trivy. Rebuild from
# source через xcaddy дают clean stdlib + auto-bumped transitive deps:
#
#   stdlib:    1.26.0 → 1.26.3+ (CVE-2026-25679/27137/32280-83/33810 fixed)
#   otel:      1.40   → 1.41+   (CVE-2026-29181 fixed)
#   otel/sdk:  1.40   → 1.43+   (CVE-2026-39883 fixed)
#   go-jose:   v3+v4   → patched (CVE-2026-34986 fixed)
#
# caddy:2.11-builder-alpine содержит Go 1.26.3 + xcaddy + ca-certificates.
# `xcaddy build v2.11.3` builds tag explicitly — reproducible.
# v2.11.3 фиксит CVE-2026-45135 (FastCGI splitPos Unicode handling allows
# execution of non-PHP files). У нас FastCGI handler не настроен, но всё
# равно держим патч-уровень up-to-date.
FROM caddy:2.11-builder-alpine AS caddy-builder
# --with overrides force-bump sub-dependencies в caddy go.sum:
#   - go-jose v3+v4 → patched (CVE-2026-34986 JWE DoS)
#   - otel 1.43 → fixed CVE-2026-29181 (baggage header DoS)
#   - otel/sdk 1.43 → fixed CVE-2026-39883 (BSD kenv PATH hijack)
#   - otel/exporters otlp{trace,metric}http 1.43 — semver-aligned с sdk выше
# Все semver-compatible upgrades. xcaddy validates через go mod tidy.
#
# NB: НЕ добавлять --with github.com/smallstep/certificates — у root-модуля
# нет импортируемого пакета (multi-package layout), xcaddy не соберётся
# с "build constraints exclude all Go files". CVE smallstep уже waived
# в .trivyignore — not-exploitable в нашей config (нет SCEP/ACME endpoint).
#
# NB: НЕ добавлять --with otlploghttp@v0.x — log-exporter живёт на v0.x
# track и тянет несовместимую otel v0.19.0 через свой go.mod, ломая весь
# build. Дождаться v1.x stable релиза log-exporter'а.
RUN xcaddy build v2.11.3 \
    --with github.com/go-jose/go-jose/v3@v3.0.5 \
    --with github.com/go-jose/go-jose/v4@v4.1.4 \
    --with go.opentelemetry.io/otel@v1.43.0 \
    --with go.opentelemetry.io/otel/sdk@v1.43.0 \
    --with go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp@v1.43.0 \
    --with go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp@v1.43.0 \
    --with github.com/mholt/caddy-ratelimit \
    --output /out/caddy

############################
# Final image: Caddy + Go API
############################
# Pinned Alpine 3.23 (вместо caddy:2.11-alpine) — caddy upstream сидит на
# Alpine 3.22, где curl/libcurl застряли на 8.14.x. Alpine 3.23 main уже
# содержит curl 8.19.0-r0 с фиксами CVE-2025-14017/14524/14819, CVE-2026-1965/
# 3783/3784/3805. Свой caddy всё равно собираем из xcaddy выше, поэтому
# дополнительные слои base-image от caddy не нужны.
FROM alpine:3.23

# wget — для HEALTHCHECK; ca-certificates — для outbound TLS из Go API.
# apk upgrade подтягивает любые post-release patch'и (libssl3, zlib и пр.).
RUN apk upgrade --no-cache \
 && apk add --no-cache ca-certificates wget tzdata \
 && rm -rf /var/cache/apk/*

# Caddyfile location + state dirs (раньше предоставлялись caddy:2.11-alpine).
RUN mkdir -p /etc/caddy /data /config

# Caddy binary — наш custom build с current Go, replaces upstream vulnerable.
COPY --from=caddy-builder /out/caddy /usr/bin/caddy

# Go-бинарь + migrate CLI (для ручного rollback в проде) + attribution worker.
# Worker живёт в этом же образе, а не отдельным сервисом: прод на Dokploy
# разворачивает ОДИН контейнер из этого образа, отдельный compose-сервис
# туда не доедет. Запускается из entrypoint.sh (см. ниже).
COPY --from=api-builder /out/api /usr/local/bin/api
COPY --from=api-builder /out/migrate /usr/local/bin/migrate
COPY --from=api-builder /out/worker /usr/local/bin/worker

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
	#
	# rate_limit — global registration через "order" directive (caddy-ratelimit
	# нестандартный, не имеет implicit priority). После этого rate_limit можно
	# использовать в любом site block.
	order rate_limit before reverse_proxy
}

:3000 {
	root * /usr/share/caddy
	encode gzip zstd

	# Edge rate-limit: 1000 req/min per-IP across весь сайт. Защита от DDoS
	# до того как запрос дойдёт до Go backend. Backend имеет свои tight limits
	# на auth endpoints (10/min) + global /v1/ (120/min), Caddy layer —
	# coarse-grained "noisy IP" cutoff.
	#
	# RFC 6585: 429 + Retry-After header возвращается клиенту.
	# Trusted health-check (Dokploy/Docker) — UA "Wget" exempt (start-period).
	rate_limit {
		zone global_ip {
			key {client_ip}
			events 1000
			window 1m
		}
	}

	# Security headers — match старый nginx setup.
	header {
		Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
		X-Frame-Options "DENY"
		X-Content-Type-Options "nosniff"
		Referrer-Policy "strict-origin-when-cross-origin"
		Permissions-Policy "camera=(), microphone=(), geolocation=(), interest-cohort=()"
		# CSP — strict. script-src без 'unsafe-inline' благодаря /preboot.js
		# (см. dashboard/index.html). style-src/font-src whitelisting Google
		# Fonts (см. <link rel=stylesheet> в index.html). img-src https: —
		# changelog markdown и user avatars могут тянуть с любых HTTPS-хостов.
		# connect-src — задаётся через CSP_CONNECT_SRC env (e.g. backend host).
		#
		# script-src jsdelivr: @monaco-editor/react lazy-loads loader.js +
		# language workers с https://cdn.jsdelivr.net/npm/monaco-editor@... —
		# нужно для admin content + email-templates редакторов. CDN заменим
		# на bundled self-hosted в Sprint 2 (см. tech-debt.md C10).
		# worker-src blob:: Monaco создаёт web workers через blob URLs.
		Content-Security-Policy "default-src 'self'; script-src 'self' https://cdn.jsdelivr.net; worker-src 'self' blob:; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdn.jsdelivr.net; img-src 'self' data: blob: https:; font-src 'self' data: https://fonts.gstatic.com; connect-src {$CSP_CONNECT_SRC:'self'}; object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self';"
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

# Запускаем api + caddy + attribution worker. wait -n валит контейнер, если
# любой из них упал — Dokploy/Swarm перезапустит целиком.
#
# Worker опционален через EOP_WORKER_ENABLED=false: его позиция обработки
# хранится в PG глобально (`worker_state`), поэтому ДВЕ реплики контейнера
# будут дублировать записи в attribution_events. Если когда-нибудь появится
# вторая реплика — worker остаётся включённым ровно на одной.
RUN cat > /entrypoint.sh <<'SH' && chmod +x /entrypoint.sh
#!/bin/sh
set -e

# Validate Caddyfile перед запуском (быстрый fail).
caddy validate --config /etc/caddy/Caddyfile

/usr/local/bin/api &
API_PID=$!

caddy run --config /etc/caddy/Caddyfile --adapter caddyfile &
CADDY_PID=$!

PIDS="$API_PID $CADDY_PID"

if [ "${EOP_WORKER_ENABLED:-true}" != "false" ]; then
	# Worker в retry-цикле и СОЗНАТЕЛЬНО не в `wait -n` ниже. Он завершается
	# fatal'ом, если ClickHouse недоступен на старте, а CH в проде живёт
	# отдельным стеком. Будь он среди «критичных для liveness» процессов,
	# любая моргнувшая CH убивала бы весь контейнер вместе с API и
	# дашбордом — то есть аналитика роняла бы продукт. Атрибуция может
	# отставать, интерфейс отставать не может.
	(
		while true; do
			/usr/local/bin/worker || echo "entrypoint: worker exited, retry in 15s" >&2
			sleep 15
		done
	) &
	PIDS="$PIDS $!"
fi

# `|| EXIT=$?` обязателен: под `set -e` голый `wait -n` с ненулевым статусом
# оборвал бы скрипт до cleanup'а и оставил остальные процессы сиротами.
EXIT=0
wait -n "$API_PID" "$CADDY_PID" || EXIT=$?

# Умер критичный — гасим всё, включая worker-цикл, и выходим.
# PIDS намеренно расщепляется по словам.
# shellcheck disable=SC2086
kill -TERM $PIDS 2>/dev/null || true
exit $EXIT
SH

# Non-root user — defense-in-depth: если в API/Caddy найдут RCE,
# атакующий не сразу получает root внутри контейнера. uid 10001
# выбран как "high range" чтобы не пересекаться с автогенерёнными
# system-юзерами base-image'а.
#
# Owner rights:
#   /usr/share/caddy   — Caddy читает SPA-static (root:eop, mode 0755 default)
#   /etc/caddy         — Caddyfile readable
#   /data, /config     — Caddy persist storage (auto_https=off, но Caddy всё
#                        равно создаёт пустые dir'ы; chown заранее)
#   /usr/local/bin/api — Go binary executable
RUN addgroup -g 10001 eop \
 && adduser -D -u 10001 -G eop -s /sbin/nologin eop \
 && mkdir -p /data /config \
 && chown -R eop:eop /data /config /etc/caddy /usr/share/caddy

USER eop:eop

EXPOSE 3000

# start-period=60s даёт миграциям и ClickHouse Cloud TLS-handshake'у
# завершиться до того как Swarm начнёт считать неудачные probe'ы.
# /healthz пробивает Caddy → Go API → проверяет PG+CH; возвращает 503 если
# что-то деградировано — Swarm перенесёт трафик на здоровый instance.
HEALTHCHECK --start-period=60s --interval=15s --timeout=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:3000/healthz >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/entrypoint.sh"]
