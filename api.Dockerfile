# syntax=docker/dockerfile:1.6
# Production image для eop-api.
# Лежит в корне репо для удобства Dokploy / любого Docker host:
# default context = "." (корень), Dockerfile = "api.Dockerfile".
#
# Dokploy:
#   Docker File:         api.Dockerfile
#   Docker Context Path: .  (или пусто — default)
#
# Локально:
#   docker build -t eop-api -f api.Dockerfile .

FROM golang:1.25-alpine AS builder
WORKDIR /src

# Cache layer для зависимостей
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend ./
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags='-s -w' \
    -o /out/api ./cmd/api

# Distroless final image — нет shell, ~6 MB
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /

COPY --from=builder /out/api /api

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/api", "--healthcheck"] || exit 1

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]
