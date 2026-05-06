# syntax=docker/dockerfile:1.6
#
# Monorepo Dockerfile with multiple build targets.
#
# Usage:
#   API:
#     docker build -t eop-api:latest --target api .
#
#   Dashboard (Vite SPA served by Nginx):
#     docker build -t eop-dashboard:latest --target dashboard \
#       --build-arg VITE_BACKEND_URL=https://eop-api.example.com \
#       --build-arg CSP_CONNECT_SRC=https://eop-api.example.com \
#       .
#

############################
# API
############################

FROM golang:1.25-alpine AS api-builder
WORKDIR /src

# Cache layer для зависимостей
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend ./
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags='-s -w' \
    -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot AS api
WORKDIR /
COPY --from=api-builder /out/api /api

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/api", "--healthcheck"] || exit 1

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]


############################
# Dashboard
############################

FROM node:20-alpine AS dashboard-builder
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

# Vite build-time args
ARG VITE_BACKEND_URL=https://eop-api.rysdavletov.org
ENV VITE_BACKEND_URL=${VITE_BACKEND_URL}
RUN pnpm -F @eop/dashboard build

FROM nginx:1.27-alpine AS dashboard

# CSP connect-src должен включать API origin, чтобы fetch() работал.
ARG CSP_CONNECT_SRC="https://eop-api.rysdavletov.org"
ENV CSP_CONNECT_SRC=${CSP_CONNECT_SRC}

COPY --from=dashboard-builder /repo/dashboard/dist /usr/share/nginx/html

# Nginx envsubst templates: docker-entrypoint.sh generates conf.d/*.conf
RUN rm -f /etc/nginx/conf.d/default.conf \
    && mkdir -p /etc/nginx/templates \
    && cat > /etc/nginx/templates/default.conf.template <<'NGINXCONF'
server {
  listen 8080;
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
    CMD wget -q --spider http://localhost:8080/ || exit 1

EXPOSE 8080

