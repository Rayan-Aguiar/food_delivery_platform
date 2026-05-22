# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src/services/auth-service

# Dependencias de build para mod download e compilacao.
RUN apk add --no-cache ca-certificates git

# Copia os mod files primeiro para aproveitar cache.
COPY services/auth-service/go.mod services/auth-service/go.sum ./
COPY shared/go.mod /src/shared/go.mod

RUN go mod download

# Copia codigo do servico e modulo shared (replace ../../shared).
COPY shared /src/shared
COPY services/auth-service /src/services/auth-service

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/auth-service ./cmd/api

FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /out/auth-service /app/auth-service

ENV SERVICE_NAME=auth-service \
    HTTP_PORT=8081 \
    LOG_LEVEL=info

EXPOSE 8081

USER app

HEALTHCHECK --interval=15s --timeout=3s --start-period=20s --retries=3 \
  CMD wget -qO- http://127.0.0.1:${HTTP_PORT}/health/live >/dev/null || exit 1

ENTRYPOINT ["/app/auth-service"]
