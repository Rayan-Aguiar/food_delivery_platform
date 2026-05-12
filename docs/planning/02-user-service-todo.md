# User Service - Plano Detalhado

## 1. Visao geral
Servico de perfil e dados de usuario, incluindo endereco e historico de pedidos consultavel.

## 2. Responsabilidades
- Manter perfil do usuario.
- Gerenciar multiplos enderecos.
- Expor historico resumido de pedidos do usuario.

## 3. Regras de negocio
- Um endereco pode ser marcado como padrao.
- Usuario so altera o proprio perfil.
- Historico retorna apenas pedidos pertencentes ao usuario.

## 4. Arquitetura interna
- `domain`: `UserProfile`, `Address`, `OrderSummary`.
- `application`: casos de uso de perfil/endereco/historico.
- `infrastructure`: Mongo + client para order-service (consulta).
- `delivery/http`: handlers protegidos por JWT.

## 5. Estrutura de pastas
```txt
services/user-service/
  cmd/api/main.go
  internal/domain/
  internal/application/
  internal/infrastructure/mongo/
  internal/infrastructure/httpclient/
  internal/delivery/http/
  configs/
  docs/openapi/
  tests/
  go.mod
  Dockerfile
```

## 6. Models e collections MongoDB
- `users`
  - `_id`, `user_id`, `name`, `phone`, `created_at`, `updated_at`
  - indice: `user_id` unico
- `addresses`
  - `_id`, `user_id`, `label`, `street`, `number`, `city`, `state`, `zip`, `is_default`
  - indices: `user_id`, `is_default`

## 7. Endpoints
- `GET /users/me`
- `PUT /users/me`
- `GET /users/me/addresses`
- `POST /users/me/addresses`
- `PUT /users/me/addresses/{id}`
- `DELETE /users/me/addresses/{id}`
- `GET /users/me/orders`

## 8. Eventos
Produzidos:
- `user.profile.updated.v1`

Consumidos:
- `user.auth.registered.v1` (criar perfil base automaticamente)
- `order.status.changed.v1` (projecao opcional de historico local)

## 9. Use cases
- GetMyProfile
- UpdateMyProfile
- AddAddress
- UpdateAddress
- DeleteAddress
- ListMyOrders

## 10. Handlers, repositories e services
Repositories:
- UserRepository
- AddressRepository

Services:
- UserPolicyService
- OrderHistoryService (query)

## 11. Middlewares e autenticacao
- JWT mandatory para rotas `/users/me`.
- Correlation ID.
- Validacao de payload.

## 12. Observabilidade
- Logs estruturados com `user_id`.
- Metricas:
  - `user_profile_updates_total`
  - `user_address_operations_total`
- Tracing: spans em chamadas ao order-service.

## 13. Retry, DLQ e idempotencia
- Consumer `user.auth.registered.v1` com retry e DLQ.
- Idempotencia por `event_id` em consumo de evento.

## 14. Testes
- Unitarios de validacao de endereco e ownership.
- Integracao com Mongo.
- Contrato de endpoint autenticado.

## 15. Dockerizacao, env e healthcheck
- Variaveis:
  - `HTTP_PORT`, `MONGO_URI`, `MONGO_DB_NAME`, `ORDER_SERVICE_URL`, `JWT_PUBLIC_KEY`
- Health:
  - `/health/live`, `/health/ready`

## 16. Swagger
- Schemas de perfil, endereco e erros padronizados.

## 17. Dependencias externas
- MongoDB.
- order-service (consulta de historico se nao houver projecao local).
- RabbitMQ para evento de cadastro.

## 18. TODO detalhado
1. Bootstrap do servico + middleware JWT.
2. Entidades e repositorios Mongo.
3. Endpoints de perfil e endereco.
4. Integracao de historico de pedidos.
5. Consumer de `user.auth.registered.v1`.
6. Observabilidade completa.
7. Testes unitarios/integracao.
8. Swagger e Dockerfile.
