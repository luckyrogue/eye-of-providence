# syntax=docker/dockerfile:1.6
# Production image для eop-api (используется в Fly.io, Dokploy, любой Docker host).
# Build context: backend/ (Dokploy: "Build Path = backend").

FROM golang:1.25-alpine AS builder
WORKDIR /src

# Cache layer для зависимостей
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags='-s -w' \
    -o /out/api ./cmd/api

# Distroless final image — нет shell, минимальная attack surface, ~6 MB.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /

COPY --from=builder /out/api /api

# Health check (Dokploy поддерживает HEALTHCHECK для определения "health")
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/api", "--healthcheck"] || exit 1

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]
