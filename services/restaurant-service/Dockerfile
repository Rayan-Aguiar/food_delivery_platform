# syntax=docker/dockerfile:1

# Build a partir da raiz do repositorio (contexto ../../ no compose) por causa do
# replace ../../shared no go.mod do servico.

FROM golang:1.25-alpine AS builder

WORKDIR /src/services/restaurant-service

RUN apk add --no-cache ca-certificates git

# Mod files primeiro: a camada de download so e' refeita quando dependencias mudam.
COPY services/restaurant-service/go.mod services/restaurant-service/go.sum ./
COPY shared/go.mod shared/go.sum /src/shared/

RUN go mod download

COPY shared /src/shared
COPY services/restaurant-service /src/services/restaurant-service

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/restaurant-service ./cmd/api

FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /out/restaurant-service /app/restaurant-service

ENV SERVICE_NAME=restaurant-service \
    HTTP_PORT=8083 \
    LOG_LEVEL=info

EXPOSE 8083

# Usuario sem privilegios: o binario nao precisa de root para nada.
USER app

HEALTHCHECK --interval=15s --timeout=3s --start-period=20s --retries=3 \
  CMD wget -qO- http://127.0.0.1:${HTTP_PORT}/health/live >/dev/null || exit 1

ENTRYPOINT ["/app/restaurant-service"]
