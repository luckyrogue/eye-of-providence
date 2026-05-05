# syntax=docker/dockerfile:1.6
# Production image для dashboard (Vite SPA → Caddy).
# Build context должен быть КОРНЕМ репозитория, чтобы Dockerfile видел ui/ workspace.
#
# Dokploy:
#   - Build Path: . (корень репо)
#   - Dockerfile Path: dashboard/Dockerfile
#   - Build Arg: VITE_BACKEND_URL=https://api.твой-домен.com
#
# Локально:
#   docker build -t eop-dashboard -f dashboard/Dockerfile \
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

ARG VITE_BACKEND_URL=http://localhost:8080
ENV VITE_BACKEND_URL=${VITE_BACKEND_URL}
RUN pnpm -F @eop/dashboard build

# Финальный образ — Caddy раздаёт статику + SPA fallback
FROM caddy:2-alpine
COPY --from=builder /repo/dashboard/dist /srv

# Inline Caddyfile (без отдельного файла — проще для Dokploy / любого деплоя)
RUN printf ':8080 {\n\
    root * /srv\n\
    encode gzip zstd\n\
    try_files {path} /index.html\n\
    file_server\n\
    header {\n\
        X-Frame-Options DENY\n\
        X-Content-Type-Options nosniff\n\
        Referrer-Policy strict-origin-when-cross-origin\n\
    }\n\
}\n' > /etc/caddy/Caddyfile

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/ || exit 1

EXPOSE 8080
