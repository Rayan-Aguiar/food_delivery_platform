# API Gateway - Plano Detalhado

## 1. Visao geral
Ponto unico de entrada para clientes, responsavel por roteamento, autenticacao, rate limit e observabilidade na borda.

## 2. Responsabilidades
- Encaminhar requests para microsservicos.
- Validar JWT (ou delegar introspeccao ao auth-service).
- Aplicar rate limit e politicas CORS.
- Propagar `request_id`, `correlation_id`, `traceparent`.

## 3. Regras de negocio
- Rotas privadas exigem token valido.
- Rotas publicas (catalogo) sem token, mas com limite de taxa.

## 4. Arquitetura interna
- `internal/router`: definicao de rotas e upstreams.
- `internal/middleware`: auth, rate limit, tracing, logging.
- `internal/proxy`: reverse proxy HTTP com timeout/circuit breaker.

## 5. Estrutura de pastas
```txt
api-gateway/
  cmd/api/main.go
  internal/router/
  internal/middleware/
  internal/proxy/
  internal/config/
  docs/openapi/
  tests/
  go.mod
  Dockerfile
```

## 6. Endpoints (publicados pelo gateway)
- Auth:
  - `POST /auth/register`
  - `POST /auth/login`
  - `POST /auth/refresh`
- User:
  - `GET /users/me`
  - `PUT /users/me`
  - `GET /users/me/orders`
- Restaurant:
  - `GET /restaurants`
  - `GET /restaurants/{id}`
  - `GET /restaurants/{id}/menu`
- Order:
  - `POST /orders`
  - `GET /orders/{id}`

## 7. Eventos
- Gateway nao deve carregar regra de negocio orientada a eventos.
- Opcional: publicar `api.request.logged.v1` para analytics (fase 3).

## 8. Use cases internos
- RouteRequest
- ValidateAccessToken
- ApplyRateLimit
- ForwardWithResilience

## 9. Middlewares
- `RequestIDMiddleware`
- `CorrelationIDMiddleware`
- `JWTAuthMiddleware`
- `RateLimitMiddleware`
- `RecoveryMiddleware`
- `MetricsMiddleware`
- `TracingMiddleware`

## 10. Observabilidade
- Logs estruturados de entrada e latencia por rota upstream.
- Metricas:
  - `gateway_http_requests_total`
  - `gateway_http_request_duration_seconds`
  - `gateway_rate_limit_rejections_total`
  - `gateway_upstream_errors_total`
- Tracing: span pai da requisicao distribuida.

## 11. Retry, circuit breaker e resiliencia
- Retry somente para GET idempotente.
- Circuit breaker por upstream.
- Timeout por rota.
- Bulkhead (pool de conexao por upstream) opcional.

## 12. Seguranca
- Validacao de assinatura JWT.
- CORS restritivo por ambiente.
- Headers de seguranca basicos.
- Protecao contra burst por token/IP.

## 13. Testes
- Unitarios de middlewares.
- Integracao de roteamento para upstream mock.
- Cenarios de rate limit, token invalido e timeout.

## 14. Docker/env/config/Swagger/health
- Env:
  - `HTTP_PORT`
  - `AUTH_SERVICE_URL`
  - `USER_SERVICE_URL`
  - `RESTAURANT_SERVICE_URL`
  - `ORDER_SERVICE_URL`
  - `JWT_PUBLIC_KEY`
  - `RATE_LIMIT_RPS`
- Health:
  - `/health/live`
  - `/health/ready` (checagem de upstreams criticos)

## 15. Dependencias externas
- auth-service.
- user-service.
- restaurant-service.
- order-service.

## 16. TODO detalhado
1. Definir mapa de rotas e contratos de upstream.
2. Implementar reverse proxy por dominio.
3. Implementar middlewares de seguranca e observabilidade.
4. Implementar validacao JWT com chave publica.
5. Implementar rate limiting configuravel.
6. Implementar timeout/retry/circuit breaker.
7. Criar testes de gateway.
8. Criar Dockerfile e OpenAPI agregada de borda.
