# Order Service - Plano Detalhado

## 1. Visao geral
Servico central do dominio de pedidos e orquestracao da Saga de compra/entrega.

## 2. Responsabilidades
- Criar pedido.
- Evoluir status do pedido.
- Orquestrar compensacoes de falha (pagamento/entrega).
- Publicar eventos de ciclo de vida do pedido.

## 3. Regras de negocio
- RN001: entrega somente apos pagamento aprovado.
- RN002: pedido cancelado nao pode ser reativado.
- RN006: pedido entregue nao pode ser alterado.
- RN007: impedir pagamento duplicado (idempotencia de criacao/processamento).

## 4. Arquitetura interna
- `domain`: `Order`, `OrderItem`, `OrderStatusTransition`, `PaymentSnapshot`.
- `application`: use cases de criacao, consulta, alteracao de status, compensacao.
- `infrastructure`: Mongo, publisher RabbitMQ, consumer RabbitMQ, client restaurant-service.
- `delivery/http`: endpoints REST.

## 5. Estrutura de pastas
```txt
services/order-service/
  cmd/api/main.go
  internal/domain/
  internal/application/
  internal/infrastructure/mongo/
  internal/infrastructure/messaging/
  internal/infrastructure/httpclient/
  internal/delivery/http/
  internal/delivery/messaging/
  configs/
  docs/openapi/
  tests/
  go.mod
  Dockerfile
```

## 6. Models e collections MongoDB
- `orders`
  - `_id`, `order_id`, `user_id`, `restaurant_id`, `items[]`, `total_amount`, `status`, `status_history[]`, `created_at`, `updated_at`
  - indices: `order_id` unico, `user_id`, `status`
- `processed_events`
  - `_id`, `event_id`, `consumer`, `processed_at`
  - indice: `event_id` unico (idempotencia)
- `outbox_events` (opcional/recomendado)
  - `_id`, `aggregate_id`, `event_type`, `payload`, `status`, `created_at`, `published_at`

## 7. Endpoints
- `POST /orders`
- `GET /orders/{id}`
- `PATCH /orders/{id}/status`
- `GET /orders/health`

## 8. Eventos
Produzidos:
- `order.created.v1`
  - payload: `order_id`, `user_id`, `restaurant_id`, `total_amount`, `items`
- `order.confirmed.v1`
  - payload: `order_id`, `confirmed_at`
- `order.cancelled.v1`
  - payload: `order_id`, `reason`, `cancelled_at`
- `order.status.changed.v1`
  - payload: `order_id`, `old_status`, `new_status`
- `delivery.requested.v1`
  - payload: `order_id`, `pickup`, `dropoff`

Consumidos:
- `payment.approved.v1`
- `payment.failed.v1`
- `delivery.started.v1`
- `delivery.completed.v1`
- `delivery.failed.v1`

## 9. Filas RabbitMQ
- Publicacao: `order.exchange`, `delivery.exchange`
- Consumo:
  - `order.payment.result.queue` (bind `payment.*`)
  - `order.delivery.result.queue` (bind `delivery.*`)
- DLQ:
  - `order.payment.result.dlq`
  - `order.delivery.result.dlq`

## 10. Fluxo de execucao (Saga)
- Criacao:
  1. Validar itens via restaurant-service.
  2. Persistir pedido em `PENDING_PAYMENT`.
  3. Publicar `order.created.v1`.
- Compensacao de pagamento:
  - `payment.failed.v1` -> `CANCELLED` + `order.cancelled.v1`.
- Aprovacao pagamento:
  - `payment.approved.v1` -> `PAID` + `order.confirmed.v1` + `delivery.requested.v1`.
- Entrega:
  - `delivery.started.v1` -> `OUT_FOR_DELIVERY`.
  - `delivery.completed.v1` -> `DELIVERED`.
  - `delivery.failed.v1` -> `CANCELLED` + `payment.refund.requested.v1` (fase 2).

## 11. Use cases
- CreateOrder
- GetOrderByID
- UpdateOrderStatus
- HandlePaymentApproved
- HandlePaymentFailed
- HandleDeliveryStarted
- HandleDeliveryCompleted
- HandleDeliveryFailed

## 12. Handlers/repositories/services/consumers/publishers
Handlers:
- CreateOrderHandler
- GetOrderHandler
- PatchStatusHandler

Repositories:
- OrderRepository
- ProcessedEventRepository
- OutboxRepository

Consumers:
- PaymentApprovedConsumer
- PaymentFailedConsumer
- DeliveryStartedConsumer
- DeliveryCompletedConsumer
- DeliveryFailedConsumer

Publishers:
- OrderEventPublisher
- DeliveryRequestPublisher

## 13. Validacoes
- Item deve existir e estar disponivel.
- Total calculado deve bater com itens.
- Transicao de status valida por maquina de estados.

## 14. Middlewares
- Auth JWT (usuario) em criacao/consulta de pedido.
- Correlation ID.
- Recovery, timeout e rate limit.

## 15. Observabilidade
- Logs por transicao de status.
- Metricas:
  - `orders_created_total`
  - `orders_cancelled_total`
  - `order_status_transition_total`
  - `saga_compensation_total`
- Tracing: spans de chamadas sync e consumo/publish de eventos.

## 16. Retry, DLQ e idempotencia
- Retry exponencial 3x no consumidor.
- Deduplicacao por `event_id`.
- Idempotencia em `POST /orders` por `Idempotency-Key`.

## 17. Testes
- Unitarios de maquina de estados.
- Integracao Mongo e RabbitMQ.
- Testes de saga (happy path e compensacao).
- Testes de contrato de evento.

## 18. Docker/env/config/Swagger/health
- Env:
  - `HTTP_PORT`, `MONGO_URI`, `MONGO_DB_NAME`
  - `RABBITMQ_URL`
  - `RESTAURANT_SERVICE_URL`
  - `WORKER_POOL_SIZE`
- Health:
  - `/health/live`
  - `/health/ready` (mongo + rabbit)
- Swagger: endpoints de pedidos.

## 19. Dependencias externas
- MongoDB.
- RabbitMQ.
- restaurant-service.
- api-gateway.

## 20. TODO detalhado
1. Definir entidades e maquina de estados.
2. Implementar endpoint `POST /orders` com validacao de menu.
3. Implementar persistencia e indices Mongo.
4. Implementar publisher de `order.created.v1`.
5. Implementar consumidores de `payment.*` e `delivery.*`.
6. Implementar compensacoes da saga.
7. Adicionar idempotencia HTTP e mensageria.
8. Adicionar metricas/tracing/logs.
9. Implementar testes de saga.
10. Swagger, Dockerfile e hardening operacional.
