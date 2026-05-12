# Notification Service - Plano Detalhado

## 1. Visao geral
Servico de notificacao multi-canal (email, SMS mock, push mock) acionado por eventos de negocio.

## 2. Responsabilidades
- Consumir eventos relevantes da jornada do pedido.
- Compor mensagem por template e canal.
- Enviar notificacao e registrar status.

## 3. Regras de negocio
- RN005: notificar usuario em mudancas de status do pedido.
- Priorizar email como canal padrao quando push/SMS indisponiveis.

## 4. Arquitetura interna
- `domain`: `Notification`, `Template`, `ChannelResult`.
- `application`: dispatch por canal, politica de fallback.
- `infrastructure`: Mongo, RabbitMQ, providers de notificacao.

## 5. Estrutura de pastas
```txt
services/notification-service/
  cmd/api/main.go
  internal/domain/
  internal/application/
  internal/infrastructure/mongo/
  internal/infrastructure/messaging/
  internal/infrastructure/providers/
  internal/delivery/http/
  internal/delivery/messaging/
  configs/
  docs/openapi/
  tests/
  go.mod
  Dockerfile
```

## 6. Models e collections MongoDB
- `notifications`
  - `_id`, `notification_id`, `user_id`, `order_id`, `channel`, `template`, `payload`, `status`, `sent_at`, `error`
  - indices: `user_id`, `order_id`, `status`
- `processed_events`
  - dedupe por `event_id`

## 7. Endpoints
- `GET /notifications/{orderId}`
- `GET /notifications/health`

## 8. Eventos
Consumidos:
- `payment.approved.v1`
- `payment.failed.v1`
- `delivery.started.v1`
- `delivery.completed.v1`
- `order.cancelled.v1`

Produzidos (opcional):
- `notification.sent.v1`
- `notification.failed.v1`

## 9. Filas RabbitMQ
- `notification.queue` <- binds multiplos de `payment.*`, `delivery.*`, `order.cancelled`
- DLQ: `notification.dlq`

## 10. Use cases
- NotifyPaymentApproved
- NotifyPaymentFailed
- NotifyDeliveryStarted
- NotifyDeliveryCompleted
- NotifyOrderCancelled

## 11. Consumers/publishers
Consumers:
- PaymentApprovedConsumer
- PaymentFailedConsumer
- DeliveryStartedConsumer
- DeliveryCompletedConsumer
- OrderCancelledConsumer

Publishers:
- NotificationSentPublisher (opcional)
- NotificationFailedPublisher (opcional)

## 12. Validacoes
- Template obrigatorio por tipo de evento.
- Canal deve estar habilitado no perfil do usuario (fase 2 via user-service).

## 13. Observabilidade
- Metricas:
  - `notifications_sent_total`
  - `notifications_failed_total`
  - `notification_dispatch_seconds`
- Tracing no consumo e envio ao provider.
- Logs com `channel`, `template`, `event_type`.

## 14. Retry, DLQ e idempotencia
- Retry com backoff para falha transiente de provider.
- Idempotencia por `event_id + channel`.
- DLQ apos limite de tentativas.

## 15. Testes
- Unitarios para template e fallback de canal.
- Integracao Rabbit/Mongo.
- Testes de resiliencia em falha de provider.

## 16. Docker/env/config/Swagger/health
- Env:
  - `HTTP_PORT`, `MONGO_URI`, `MONGO_DB_NAME`, `RABBITMQ_URL`
  - `EMAIL_PROVIDER_MODE`, `SMS_PROVIDER_MODE`, `PUSH_PROVIDER_MODE`
  - `WORKER_POOL_SIZE`
- Health:
  - `/health/live`, `/health/ready`

## 17. Dependencias externas
- MongoDB.
- RabbitMQ.
- Providers de notificacao (mock na fase inicial).

## 18. TODO detalhado
1. Modelar notificacao e templates.
2. Implementar consumers dos eventos de negocio.
3. Implementar dispatcher por canal.
4. Persistir resultado de envio.
5. Adicionar retry, DLQ e idempotencia.
6. Adicionar metricas, logs e tracing.
7. Implementar testes de falha/fallback.
8. Dockerfile e OpenAPI.
