# Delivery Service - Plano Detalhado

## 1. Visão geral
Servico de alocacao de entregador e rastreamento da entrega.

## 2. Responsabilidades
- Consumir pedido apto para entrega.
- Alocar entregador.
- Atualizar status de entrega.
- Publicar eventos de inicio/fim/falha.

## 3. Regras de negocio
- Entrega inicia somente apos pagamento aprovado (RN001).
- Entrega finalizada nao volta de estado.

## 4. Arquitetura interna
- `domain`: `Delivery`, `Courier`, `TrackingEvent`.
- `application`: iniciar entrega, concluir, falhar.
- `infrastructure`: Mongo e RabbitMQ.

## 5. Estrutura de pastas
```txt
services/delivery-service/
  cmd/api/main.go
  internal/domain/
  internal/application/
  internal/infrastructure/mongo/
  internal/infrastructure/messaging/
  internal/delivery/http/
  internal/delivery/messaging/
  configs/
  docs/openapi/
  tests/
  go.mod
  Dockerfile
```

## 6. Models e collections MongoDB
- `deliveries`
  - `_id`, `delivery_id`, `order_id`, `courier_id`, `status`, `started_at`, `completed_at`
  - indice unico: `order_id`
- `couriers`
  - `_id`, `courier_id`, `name`, `status`, `last_location`
- `processed_events`
  - dedupe por `event_id`

## 7. Endpoints
- `GET /deliveries/{orderId}`
- `PATCH /deliveries/{orderId}/status`
- `GET /deliveries/health`

## 8. Eventos
Consumidos:
- `delivery.requested.v1`

Produzidos:
- `delivery.started.v1`
  - payload: `order_id`, `delivery_id`, `courier_id`, `started_at`
- `delivery.completed.v1`
  - payload: `order_id`, `delivery_id`, `completed_at`
- `delivery.failed.v1`
  - payload: `order_id`, `reason`, `failed_at`

## 9. Filas RabbitMQ
- `delivery.processing.queue` <- `delivery.requested.v1`
- DLQ: `delivery.processing.dlq`

## 10. Use cases
- StartDelivery
- CompleteDelivery
- FailDelivery
- GetDeliveryByOrder

## 11. Consumers/publishers
Consumer:
- DeliveryRequestedConsumer

Publishers:
- DeliveryStartedPublisher
- DeliveryCompletedPublisher
- DeliveryFailedPublisher

## 12. Validacoes
- Nao iniciar entrega sem `order_id` valido.
- Nao concluir entrega sem estado `IN_PROGRESS`.

## 13. Observabilidade
- Metricas:
  - `deliveries_started_total`
  - `deliveries_completed_total`
  - `deliveries_failed_total`
- Tracing para consumo e transicao de entrega.

## 14. Retry, DLQ e idempotencia
- Retry exponencial no consumer.
- Idempotencia por `order_id` no inicio de entrega.
- DLQ para mensagens invalidas/nao processaveis.

## 15. Testes
- Unitarios de transicao de estado.
- Integracao Mongo + Rabbit.
- Teste de duplicidade de evento `delivery.requested.v1`.

## 16. Docker/env/config/Swagger/health
- Env: `HTTP_PORT`, `MONGO_URI`, `MONGO_DB_NAME`, `RABBITMQ_URL`, `WORKER_POOL_SIZE`.
- Health: `/health/live`, `/health/ready`.

## 17. Dependencias externas
- MongoDB.
- RabbitMQ.

## 18. TODO detalhado
1. Modelar entidade de entrega.
2. Implementar consumer de `delivery.requested.v1`.
3. Implementar alocacao de entregador (mock).
4. Publicar eventos de ciclo de entrega.
5. Implementar endpoints de consulta/status.
6. Adicionar observabilidade completa.
7. Implementar retry, DLQ e idempotencia.
8. Testes e Dockerfile.
