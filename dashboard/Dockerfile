# Dashboard как статика, обслуживается через Caddy (HTTP/2, авто-gzip).
# VITE_BACKEND_URL задаётся в build args (см. fly.toml).

FROM node:20-alpine AS builder
WORKDIR /repo

# pnpm через corepack (приходит с Node)
RUN corepack enable && corepack prepare pnpm@9.12.3 --activate

# Лучший layer-cache: сначала только манифесты
COPY pnpm-workspace.yaml package.json pnpm-lock.yaml ./
COPY ui/package.json ui/
COPY dashboard/package.json dashboard/
RUN pnpm install --frozen-lockfile=false

# Затем код
COPY ui ui
COPY dashboard dashboard

ARG VITE_BACKEND_URL=https://eop-api.fly.dev
ENV VITE_BACKEND_URL=${VITE_BACKEND_URL}
RUN pnpm -F @eop/dashboard build

FROM caddy:2-alpine
COPY --from=builder /repo/dashboard/dist /srv
COPY dashboard/Caddyfile /etc/caddy/Caddyfile
EXPOSE 8080
