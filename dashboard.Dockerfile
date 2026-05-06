# syntax=docker/dockerfile:1.6
# Production image для dashboard (Vite SPA → Caddy).
# Лежит в корне репо чтобы Dokploy / любой Docker host работал без path-tricks:
# default context = "." (корень), Dockerfile = "dashboard.Dockerfile".
#
# Dokploy:
#   Docker File:         dashboard.Dockerfile
#   Docker Context Path: .  (или оставь пустым — это default)
#   Build Arg:           VITE_BACKEND_URL=https://eop-api.rysdavletov.org
#
# Локально:
#   docker build -t eop-dashboard -f dashboard.Dockerfile \
#     --build-arg VITE_BACKEND_URL=http://localhost:8080 .

FROM node:20-alpine AS builder
WORKDIR /repo

RUN corepack enable && corepack prepare pnpm@9.12.3 --activate

# Cache layer для install
COPY pnpm-workspace.yaml package.json pnpm-lock.yaml* ./
COPY ui/package.json ui/
COPY dashboard/package.json dashboard/
RUN pnpm install --frozen-lockfile=false

# Source
COPY ui ui
COPY dashboard dashboard

ARG VITE_BACKEND_URL=https://eop-api.rysdavletov.org
ENV VITE_BACKEND_URL=${VITE_BACKEND_URL}
RUN pnpm -F @eop/dashboard build

# Финальный образ — Caddy раздаёт статику + SPA fallback.
FROM caddy:2-alpine

# CSP connect-src должен включать API origin, чтобы fetch() работал.
ARG CSP_CONNECT_SRC="https://eop-api.rysdavletov.org"
ENV CSP_CONNECT_SRC=${CSP_CONNECT_SRC}

# Статика
COPY --from=builder /repo/dashboard/dist /srv

# Inline Caddyfile (через shell heredoc — работает в любом docker, не требует BuildKit).
# Caddy сам подставляет {$CSP_CONNECT_SRC} из ENV при загрузке конфига.
RUN cat > /etc/caddy/Caddyfile <<'CADDYFILE' \
    && mkdir -p /data/caddy /config/caddy \
    && chown -R nobody:nobody /srv /data/caddy /config/caddy /etc/caddy
:8080 {
    root * /srv
    encode gzip zstd
    try_files {path} /index.html
    file_server
    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
        X-Frame-Options "DENY"
        X-Content-Type-Options "nosniff"
        Referrer-Policy "strict-origin-when-cross-origin"
        Permissions-Policy "camera=(), microphone=(), geolocation=(), interest-cohort=()"
        Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self' {$CSP_CONNECT_SRC}; object-src 'none'; frame-ancestors 'none'; base-uri 'self';"
        -Server
    }
}
CADDYFILE

USER nobody:nobody

# Healthcheck бьётся в локальный Caddy — НЕ во внешний API
# (внешний URL мог быть недоступен и контейнер считался бы unhealthy).
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/ || exit 1

EXPOSE 8080
