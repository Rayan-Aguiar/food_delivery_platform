# Payment Service - Plano Detalhado

## 1. Visão geral
Servico financeiro para autorizacao de pagamento, controle de transacoes e resultado da etapa financeira da saga.

## 2. Responsabilidades
- Consumir pedido criado.
- Processar pagamento (simulador de gateway).
- Publicar aprovacao/reprovacao.
- Garantir nao duplicidade de cobranca.

## 3. Regras de negocio
- RN004: pedido pago deve ter registro financeiro.
- RN007: impedir duplicacao de pagamento por `order_id` + `idempotency_key`.

## 4. Arquitetura interna
- `domain`: `Payment`, `PaymentAttempt`, `Refund`.
- `application`: processar pagamento e estorno.
- `infrastructure`: Mongo, RabbitMQ consumer/publisher, mock gateway.

## 5. Estrutura de pastas
```txt
services/payment-service/
  cmd/api/main.go
  internal/domain/
  internal/application/
  internal/infrastructure/mongo/
  internal/infrastructure/messaging/
  internal/infrastructure/gateway/
  internal/delivery/http/
  internal/delivery/messaging/
  configs/
  docs/openapi/
  tests/
  go.mod
  Dockerfile
```

## 6. Models e collections MongoDB
- `payments`
  - `_id`, `payment_id`, `order_id`, `user_id`, `amount`, `status`, `method`, `provider_ref`, `created_at`, `updated_at`
  - indice unico: `order_id`
- `payment_attempts`
  - `_id`, `order_id`, `attempt`, `result`, `error`, `created_at`
- `processed_events`
  - dedupe por `event_id`

## 7. Endpoints
- `GET /payments/{orderId}`
- `POST /payments/{orderId}/refund` (fase 2)
- `GET /payments/health`

## 8. Eventos
Consumidos:
- `order.created.v1`
- `payment.refund.requested.v1` (fase 2)

Produzidos:
- `payment.approved.v1`
  - payload: `order_id`, `payment_id`, `amount`, `approved_at`
- `payment.failed.v1`
  - payload: `order_id`, `reason`, `failed_at`
- `payment.refunded.v1` (fase 2)

## 9. Filas RabbitMQ
- `payment.processing.queue` <- `order.created.v1`
- `payment.refund.queue` <- `payment.refund.requested.v1`
- DLQ:
  - `payment.processing.dlq`
  - `payment.refund.dlq`

## 10. Use cases
- ProcessPayment
- GetPaymentStatus
- RefundPayment

## 11. Consumers/publishers
Consumers:
- OrderCreatedConsumer
- RefundRequestedConsumer

Publishers:
- PaymentApprovedPublisher
- PaymentFailedPublisher
- PaymentRefundedPublisher

## 12. Validacoes
- Valor do pagamento deve coincidir com valor do pedido.
- Mesmo pedido nao processa pagamento mais de uma vez.

## 13. Observabilidade
- Metricas:
  - `payments_processed_total`
  - `payments_approved_total`
  - `payments_failed_total`
  - `payment_processing_seconds`
- Tracing no fluxo de consumo e gateway.

## 14. Retry, DLQ e idempotencia
- Retry 3x com backoff exponencial.
- Idempotencia por `order_id` + `event_id`.
- DLQ com payload + causa de falha.

## 15. Testes
- Unitarios de regra financeira.
- Integracao com Mongo + Rabbit.
- Testes de duplicidade de evento.
- Testes de contrato de evento publicado.

## 16. Docker/env/config/Swagger/health
- Env:
  - `HTTP_PORT`, `MONGO_URI`, `MONGO_DB_NAME`
  - `RABBITMQ_URL`, `PAYMENT_GATEWAY_MODE`, `WORKER_POOL_SIZE`
- Health:
  - `/health/live`, `/health/ready`

## 17. Dependencias externas
- MongoDB.
- RabbitMQ.

## 18. TODO detalhado
1. Definir entidade `Payment` e estados.
2. Implementar consumer de `order.created.v1`.
3. Implementar simulador de gateway.
4. Publicar `payment.approved.v1` e `payment.failed.v1`.
5. Implementar deduplicacao de eventos.
6. Implementar retry + DLQ.
7. Adicionar metricas/tracing/logs.
8. Implementar testes completos.
9. Dockerfile e OpenAPI.
