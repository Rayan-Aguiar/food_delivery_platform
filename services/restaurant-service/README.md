# Restaurant Service

Catalogo de restaurantes e cardapio: consulta publica, administracao protegida por
JWT, validacao interna de itens para o order-service e eventos de catalogo no RabbitMQ.

## Variaveis de ambiente

| Variavel | Obrigatoria | Default | Descricao |
|---|---|---|---|
| `SERVICE_NAME` | nao | `restaurant-service` | Nome nos logs e no tracing |
| `SERVICE_VERSION` | nao | `dev` | Versao no tracing |
| `ENVIRONMENT` | nao | `local` | `deployment.environment` no tracing |
| `HTTP_PORT` | nao | `8083` | Porta HTTP |
| `LOG_LEVEL` | nao | `info` | `debug`, `info`, `warn`, `error` |
| `REQUEST_TIMEOUT_SECONDS` | nao | `10` | Timeout global por request |
| `MONGO_URI` | **sim** | - | Ex.: `mongodb://localhost:27018` |
| `MONGO_DB_NAME` | nao | `restaurant_db` | Database do catalogo |
| `JWT_SECRET` | **sim** | - | Mesmo valor do `JWT_SECRET` do auth-service (minimo 32 caracteres) |
| `JWT_ISSUER` | nao | `auth-service` | Issuer esperado nos tokens |
| `RABBITMQ_URL` | nao | vazio | Sem valor, o servico sobe sem publicar eventos |
| `OTEL_ENABLED` | nao | `false` | Liga tracing OTLP/HTTP (endpoint via `OTEL_EXPORTER_OTLP_ENDPOINT`) |

O servico nao sobe sem `MONGO_URI` ou com `JWT_SECRET` ausente/curto (< 32 caracteres).

## Endpoints

| Metodo | Rota | Acesso |
|---|---|---|
| GET | `/health/live`, `/health/ready` | publico |
| GET | `/metrics` | publico (Prometheus) |
| GET | `/restaurants?page=&page_size=` | publico, rate limit por IP |
| GET | `/restaurants/{id}` | publico, rate limit por IP |
| GET | `/restaurants/{id}/menu` | publico, rate limit por IP |
| POST | `/restaurants` | JWT |
| POST | `/restaurants/{id}/menu/items` | JWT, aceita `Idempotency-Key` |
| PATCH | `/restaurants/{id}/menu/items/{itemId}` | JWT |
| POST | `/restaurants/{id}/menu/validate` | interno (order-service) |

Contrato completo: [`docs/openapi/openapi.v1.yaml`](docs/openapi/openapi.v1.yaml).

Limites: `page_size` ate 100 (default 20), `page` ate 1000, body ate 64 KiB e
ate 100 `item_ids` por validacao.

## Executar

### Local (infra no Docker, servico com `go run`)
```bash
cd deploy/compose
docker compose -p food-restaurant --env-file .env.restaurant -f docker-compose.restaurant.yml up -d restaurant-mongo restaurant-rabbitmq

cd ../../services/restaurant-service
cp .env.example .env   # ajuste JWT_SECRET para o mesmo do auth-service
go run ./cmd/api
```

### Tudo no Docker Compose
```bash
cd deploy/compose
cp .env.restaurant.example .env.restaurant   # ajuste JWT_SECRET
docker compose -p food-restaurant --env-file .env.restaurant -f docker-compose.restaurant.yml up -d --build
```

### Build da imagem
Use a raiz do repositorio como contexto por causa do `replace ../../shared` no `go.mod`:
```bash
docker build -f services/restaurant-service/Dockerfile -t restaurant-service:local .
```

## Testes

```bash
# unitarios (sem dependencias externas)
go test ./...

# integracao com Mongo real (database temporario, removido ao final)
MONGO_TEST_URI=mongodb://localhost:27018 go test -tags=integration ./internal/infrastructure/mongo/...
```

O contrato do endpoint de validacao e' fixado em
`internal/delivery/http/validate_contract_test.go`.

## Observabilidade

- **Logs**: JSON estruturado por request (`method`, `path`, `status`, `duration_ms`,
  `request_id`, `correlation_id`).
- **Metricas** em `/metrics`:
  - `http_request_duration_seconds{method,path,status}`: `path` e' o padrao da rota
    (ex.: `GET /restaurants/{id}/menu`), nunca o path com ids.
  - `restaurant_menu_item_validations_total{result}`: `valid`, `item_not_found`,
    `item_unavailable` ou `restaurant_inactive`.
- **Tracing** (`OTEL_ENABLED=true`): span por request nomeado pela rota, spans das
  queries Mongo e atributos `restaurant.id`, `menu.validation.item_count` e
  `menu.validation.all_valid` na validacao.

Exemplos de PromQL:
```promql
# p95 de leitura do cardapio
histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket{path="GET /restaurants/{id}/menu"}[5m])))

# taxa de itens rejeitados na validacao, por motivo
sum by (result) (rate(restaurant_menu_item_validations_total{result!="valid"}[5m]))
```
