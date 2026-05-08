# syntax=docker/dockerfile:1.6
# Объединённый production image: Go API + dashboard SPA в одном контейнере.
# Nginx раздаёт статику dashboard и проксирует /api/* на локальный Go-бинарь.
#
# Dokploy:
#   Provider: Docker
#   Docker Image: ghcr.io/<owner>/eop:latest
#   Port: 3000
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
    -o /out/api ./cmd/api

############################
# Dashboard builder
############################
FROM node:22-alpine AS dashboard-builder
WORKDIR /repo
RUN corepack enable && corepack prepare pnpm@9.12.3 --activate

COPY pnpm-workspace.yaml package.json pnpm-lock.yaml* ./
COPY ui/package.json ui/
COPY dashboard/package.json dashboard/
RUN pnpm install --frozen-lockfile=false

COPY ui ui
COPY dashboard dashboard

# В combined-режиме фронт ходит относительно — на тот же хост
ARG VITE_BACKEND_URL=/api
ENV VITE_BACKEND_URL=${VITE_BACKEND_URL}
RUN pnpm -F @eop/dashboard build

############################
# Final image: nginx + Go API
############################
FROM nginx:1.27-alpine

# Go-бинарь
COPY --from=api-builder /out/api /usr/local/bin/api

# Статика dashboard
COPY --from=dashboard-builder /repo/dashboard/dist /usr/share/nginx/html

ARG CSP_CONNECT_SRC="'self'"
ENV CSP_CONNECT_SRC=${CSP_CONNECT_SRC}

# Nginx config — генерируется через envsubst при старте
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
  add_header Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src ${CSP_CONNECT_SRC}; object-src 'none'; frame-ancestors 'none'; base-uri 'self';" always;

  # API: /api/v1/foo  →  Go-бинарь :8080  →  /v1/foo
  location /api/ {
    proxy_pass http://127.0.0.1:8080/;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 60s;
  }

  # Прямой healthcheck Go-бинаря
  location = /healthz {
    proxy_pass http://127.0.0.1:8080/healthz;
  }

  # SPA fallback
  location / {
    try_files $uri $uri/ /index.html;
  }
}
NGINXCONF

# Запускаем оба процесса. wait -n валит контейнер, если любой упал.
RUN cat > /entrypoint.sh <<'SH' && chmod +x /entrypoint.sh
#!/bin/sh
set -e

# Стандартный nginx entrypoint выполняет envsubst по templates → conf.d
/docker-entrypoint.sh nginx -t

/usr/local/bin/api &
API_PID=$!

nginx -g "daemon off;" &
NGINX_PID=$!

wait -n "$API_PID" "$NGINX_PID"
EXIT=$?

# Если умер один — добиваем второй и выходим
kill -TERM "$API_PID" "$NGINX_PID" 2>/dev/null || true
exit $EXIT
SH

EXPOSE 3000

# Healthcheck выключен: entrypoint.sh завершается, если умрёт nginx или api,
# поэтому Swarm всё равно увидит crash. HEALTHCHECK + Swarm любил рубить
# контейнер по timeout, чем создавал restart-loop.

ENTRYPOINT ["/entrypoint.sh"]
