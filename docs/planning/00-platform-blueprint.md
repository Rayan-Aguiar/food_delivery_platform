# Platform Blueprint - Food Delivery

## 1. Objetivo arquitetural
Construir uma plataforma de delivery distribuida, desacoplada e orientada a eventos com microsservicos em Go, banco isolado por servico (MongoDB), mensageria RabbitMQ, observabilidade ponta a ponta e padrao Saga para consistencia eventual.

## 2. Premissas obrigatorias
- Cada servico e independente: `go.mod`, Dockerfile, configs e banco proprios.
- Comunicacao sincronica somente para leitura/validacao/autenticacao de baixa latencia.
- Comunicacao assincrona para mudancas de estado e workflow de negocio.
- Eventos versionados (`event_type`, `event_version`) e idempotentes (`event_id`, `idempotency_key`).
- Correlation ID e Trace Context propagados em HTTP e RabbitMQ.

## 3. Responsabilidades por dominio
- auth-service: identidade, credenciais, JWT e refresh token.
- user-service: perfil, enderecos e historico de pedidos por usuario.
- restaurant-service: restaurantes, menu, disponibilidade de item.
- order-service: criacao e ciclo de vida de pedido (orquestracao de saga de pedido).
- payment-service: autorizacao/captura simulada e resultado financeiro.
- delivery-service: alocacao de entregador e ciclo de entrega.
- notification-service: envio de notificacoes por canal.
- api-gateway: borda de entrada, auth pass-through, rate limit, roteamento.

## 4. Contrato padrao de evento
Todos os eventos devem seguir envelope comum:

```json
{
  "event_id": "uuid",
  "event_type": "order.created",
  "event_version": 1,
  "occurred_at": "2026-05-12T12:00:00Z",
  "producer": "order-service",
  "correlation_id": "uuid",
  "causation_id": "uuid",
  "idempotency_key": "string",
  "traceparent": "00-...",
  "payload": {}
}
```

Beneficio: padroniza roteamento, tracing e resiliencia.
Trade-off: maior overhead de mensagem.

## 5. RabbitMQ - topologia recomendada
Tipo de exchange: `topic` para flexibilidade de roteamento.

- `order.exchange`
  - Routing keys: `order.created`, `order.confirmed`, `order.cancelled`, `order.status.changed`
- `payment.exchange`
  - Routing keys: `payment.requested`, `payment.approved`, `payment.failed`, `payment.refund.requested`, `payment.refunded`
- `delivery.exchange`
  - Routing keys: `delivery.requested`, `delivery.started`, `delivery.completed`, `delivery.failed`
- `notification.exchange`
  - Routing keys: `notification.requested`, `notification.sent`, `notification.failed`

Filas principais:
- `payment.processing.queue` <- `order.created`
- `delivery.processing.queue` <- `payment.approved`
- `order.compensation.queue` <- `payment.failed|delivery.failed`
- `notification.queue` <- eventos de negocio relevantes

DLQs:
- `payment.processing.dlq`
- `delivery.processing.dlq`
- `order.compensation.dlq`
- `notification.dlq`

Retry:
- 3 tentativas com backoff exponencial: 5s, 30s, 2m.
- Retry via filas de atraso (`x-message-ttl` + dead-letter-exchange) por tentativa.
- Apos limite, mensagem vai para DLQ com `failure_reason`.

## 6. Fluxo Saga principal (pedido)
1. API Gateway envia `POST /orders` para order-service.
2. order-service valida itens no restaurant-service (sync), cria pedido `PENDING_PAYMENT` e publica `order.created`.
3. payment-service consome `order.created` e processa pagamento:
- sucesso: publica `payment.approved`.
- falha: publica `payment.failed`.
4. order-service consome resposta de pagamento:
- `payment.approved` -> atualiza pedido para `PAID` e publica `order.confirmed` + `delivery.requested`.
- `payment.failed` -> acao compensatoria: `CANCELLED` + publica `order.cancelled`.
5. delivery-service consome `delivery.requested`:
- sucesso -> `delivery.started` e depois `delivery.completed`.
- falha -> `delivery.failed`.
6. order-service consome eventos de entrega e atualiza status final (`OUT_FOR_DELIVERY`, `DELIVERED` ou compensacao).
7. notification-service consome eventos e notifica usuario.

## 7. Estados de pedido e regras de negocio
Estado recomendado:
- `CREATED` -> `PENDING_PAYMENT` -> `PAID` -> `OUT_FOR_DELIVERY` -> `DELIVERED`
- Cancelamento: `CREATED|PENDING_PAYMENT|PAID` -> `CANCELLED` (com regra de compensacao financeira quando necessario)

Regras:
- RN001: entrega so inicia apos `payment.approved`.
- RN002: pedido cancelado nao reativa.
- RN006: pedido entregue nao altera.
- RN007: pagamento nao pode duplicar (idempotency).

## 8. Estrutura de monorepo recomendada

```txt
services/<service-name>/
  cmd/api/main.go
  internal/
    domain/
    application/
    infrastructure/
    delivery/http/
    delivery/messaging/
  docs/openapi/
  configs/
  migrations/ (se aplicavel)
  tests/
  go.mod
  Dockerfile

shared/
  broker/
  contracts/
  errors/
  events/
  logger/
  middleware/
  utils/
```

## 9. Padroes de codigo e contratos internos
- Clean Architecture por servico.
- Repository Pattern no dominio/aplicacao.
- Dependency Injection por construtores.
- Worker pool para consumidores RabbitMQ.
- Outbox Pattern recomendado em `order-service` e `payment-service` para confiabilidade.

## 10. Observabilidade padrao
- Logs JSON com campos: `timestamp`, `level`, `service`, `message`, `request_id`, `correlation_id`, `trace_id`, `span_id`, `error_code`.
- Metricas Prometheus:
  - `http_request_duration_seconds`
  - `http_requests_total`
  - `rabbitmq_consumer_lag`
  - `events_processed_total`
  - `events_failed_total`
- Tracing OpenTelemetry + Jaeger.

## 11. Seguranca
- JWT assinado com chave rotacionavel.
- Refresh token com revogacao e expiracao.
- Hash de senha: Argon2id (preferivel) ou bcrypt.
- Rate limit no gateway.
- Comunicacao interna protegida por mTLS (fase avancada).

## 12. Dependencias criticas
- RabbitMQ indisponivel: afetacao alta em workflows assincronos.
- MongoDB por servico: risco de latencia e lock em indices ruins.
- order-service: ponto central do fluxo de negocio.

Mitigacoes:
- Retry + DLQ + idempotencia.
- Circuit breaker em chamadas sincronas.
- Healthcheck de dependencia (`/health/ready`).

## 13. Riscos arquiteturais e gargalos
- Duplicacao de eventos sem deduplicacao por consumidor.
- Acoplamento acidental via shared package com regra de negocio.
- Crescimento de filas sem autoscaling de consumers.
- Falta de contrato versionado quebrando compatibilidade entre servicos.

## 14. Ordem de implementacao recomendada
1. Fundacao compartilhada: contratos de evento, logger, middleware de correlacao, broker client, error model.
2. auth-service + api-gateway (entrada e seguranca).
3. restaurant-service (catalogo e disponibilidade).
4. order-service (sem saga completa no primeiro incremento).
5. payment-service e integracao `order.created` -> `payment.*`.
6. delivery-service e integracao `payment.approved` -> `delivery.*`.
7. notification-service para eventos de negocio.
8. observabilidade completa, testes E2E e hardening de resiliencia.

## 15. Estrategia incremental
- Iteracao 1: happy path (pedido pago e entregue).
- Iteracao 2: compensacoes da saga (falha de pagamento e falha de entrega).
- Iteracao 3: idempotencia, retry avancado, DLQ operacional.
- Iteracao 4: performance, seguranca avancada e readiness para producao.
