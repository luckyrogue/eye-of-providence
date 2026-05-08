# syntax=docker/dockerfile:1.6
# Production image для dashboard (Vite SPA → Nginx).
#
# Dokploy:
#   Dockerfile:   dashboard.Dockerfile
#   Build Path:   .
#   Port:         3000
#   Build Args:
#     VITE_BACKEND_URL=https://eop-api.rysdavletov.org
#     CSP_CONNECT_SRC=https://eop-api.rysdavletov.org
#
# Локально:
#   docker build -t eop-dashboard -f dashboard.Dockerfile \
#     --build-arg VITE_BACKEND_URL=https://eop-api.rysdavletov.org .

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

# Vite build-time args (важно — VITE_BACKEND_URL запекается в bundle)
ARG VITE_BACKEND_URL=https://eop-api.rysdavletov.org
ENV VITE_BACKEND_URL=${VITE_BACKEND_URL}
RUN pnpm -F @eop/dashboard build

# Финальный образ — Nginx раздаёт статику + SPA fallback
FROM nginx:1.27-alpine

# CSP connect-src должен включать API origin, чтобы fetch() работал.
ARG CSP_CONNECT_SRC="https://eop-api.rysdavletov.org"
ENV CSP_CONNECT_SRC=${CSP_CONNECT_SRC}

COPY --from=builder /repo/dashboard/dist /usr/share/nginx/html

# Nginx envsubst templates: docker-entrypoint.sh подставляет ${CSP_CONNECT_SRC}
# в /etc/nginx/conf.d/default.conf при старте контейнера. $uri и др. nginx-внутренние
# переменные envsubst не трогает (их нет в env).
RUN rm -f /etc/nginx/conf.d/default.conf \
    && mkdir -p /etc/nginx/templates \
    && cat > /etc/nginx/templates/default.conf.template <<'NGINXCONF'
server {
  listen 3000;
  server_name _;

  root /usr/share/nginx/html;
  index index.html;

  gzip on;
  gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript image/svg+xml;

  add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
  add_header X-Frame-Options "DENY" always;
  add_header X-Content-Type-Options "nosniff" always;
  add_header Referrer-Policy "strict-origin-when-cross-origin" always;
  add_header Permissions-Policy "camera=(), microphone=(), geolocation=(), interest-cohort=()" always;
  add_header Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self' ${CSP_CONNECT_SRC}; object-src 'none'; frame-ancestors 'none'; base-uri 'self';" always;

  location / {
    try_files $uri $uri/ /index.html;
  }
}
NGINXCONF

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:3000/index.html || exit 1

EXPOSE 3000
