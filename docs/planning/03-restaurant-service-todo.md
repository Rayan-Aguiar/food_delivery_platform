# Restaurant Service - Plano Detalhado

## 1. Visao geral
Servico de catalogo: restaurantes, cardapio, categorias e disponibilidade de itens.

## 2. Responsabilidades
- CRUD de restaurante (admin).
- CRUD de item de menu (admin).
- Consulta publica de restaurantes e menu.
- Disponibilidade para validacao de pedido.

## 3. Regras de negocio
- Item indisponivel nao pode ser pedido (RN003).
- Restaurante inativo nao recebe pedidos.

## 4. Arquitetura interna
- `domain`: `Restaurant`, `MenuItem`, `Category`.
- `application`: casos de uso de catalogo e validacao de item.
- `delivery/http`: APIs publicas e admin.

## 5. Estrutura de pastas
```txt
services/restaurant-service/
  cmd/api/main.go
  internal/domain/
  internal/application/
  internal/infrastructure/mongo/
  internal/delivery/http/
  configs/
  docs/openapi/
  tests/
  go.mod
  Dockerfile
```

## 6. Models e collections MongoDB
- `restaurants`: `_id`, `name`, `status`, `delivery_fee`, `created_at`
- `menu_items`: `_id`, `restaurant_id`, `name`, `price`, `category`, `available`
- indices: `restaurant_id`, `available`, texto por nome.

## 7. Endpoints
- `GET /restaurants`
- `GET /restaurants/{id}`
- `GET /restaurants/{id}/menu`
- `POST /restaurants` (admin)
- `POST /restaurants/{id}/menu/items` (admin)
- `PATCH /restaurants/{id}/menu/items/{itemId}` (admin)
- `POST /restaurants/{id}/menu/validate` (interno order-service)

## 8. Eventos
Produzidos:
- `restaurant.menu.updated.v1`
- `restaurant.availability.changed.v1`

Consumidos:
- Nenhum obrigatorio inicial.

## 9. Use cases
- ListRestaurants
- GetRestaurantDetails
- ListRestaurantMenu
- ValidateMenuItems
- UpsertMenuItem

## 10. Repositories/services/handlers
Repositories:
- RestaurantRepository
- MenuRepository

Services:
- AvailabilityService
- PricingPolicyService

## 11. Middlewares
- JWT em rotas admin.
- Rate limit em consultas publicas.
- Correlation ID.

## 12. Observabilidade
- Metricas de leitura de menu e validacao.
- Tracing em rota de validacao de itens (hot path do pedido).

## 13. Retry/DLQ/idempotencia
- Sem consumidor critico fase 1.
- Idempotencia em upsert admin por `idempotency-key` opcional.

## 14. Testes
- Unitarios de disponibilidade e validacao de preco.
- Integracao Mongo.
- Teste de contrato do endpoint interno de validacao.

## 15. Docker/env/health
- Env: `HTTP_PORT`, `MONGO_URI`, `MONGO_DB_NAME`, `JWT_PUBLIC_KEY`.
- Health: `/health/live`, `/health/ready`.

## 16. Swagger
- APIs publicas + admin + endpoint interno.

## 17. Dependencias externas
- MongoDB.
- API Gateway.

## 18. TODO detalhado
1. Modelagem de entidades e indices.
2. Endpoints publicos.
3. Endpoints admin.
4. Endpoint de validacao para order-service.
5. Observabilidade e testes.
6. Swagger e Dockerfile.
