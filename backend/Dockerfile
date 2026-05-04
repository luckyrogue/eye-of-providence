# syntax=docker/dockerfile:1.6
FROM golang:1.25-alpine AS builder
WORKDIR /src

# Сначала зависимости — для лучшего layer-кэша
COPY go.mod go.sum ./
RUN go mod download

# Затем код
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/api ./cmd/api

# Финальный образ — distroless (нет shell, минимальная attack surface)
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=builder /out/api /api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]
