# Auth Service

## Variaveis de ambiente
- `SERVICE_NAME` (default: `auth-service`)
- `HTTP_PORT` (default: `8081`)
- `LOG_LEVEL` (default: `info`)
- `REQUEST_TIMEOUT_SECONDS` (default: `10`)
- `MONGO_URI` (default: `mongodb://localhost:27017`)
- `MONGO_DB_NAME` (default: `auth_db`)
- `RABBITMQ_URL` (default: `amqp://guest:guest@localhost:5672/`)
- `JWT_SECRET` (obrigatorio em ambiente real)
- `JWT_ISSUER` (default: `auth-service`)
- `JWT_ACCESS_TTL` (em minutos, default: `15`)
- `JWT_REFRESH_TTL` (em minutos, default: `10080`)
- `BCRYPT_COST` (default: `12`)

## Endpoints principais
- `GET /health/live`
- `GET /health/ready`
- `GET /auth/health`
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`
- `POST /auth/logout`

## OpenAPI
- Spec versionada: `docs/openapi/openapi.v1.yaml`
- Guia de versionamento: `docs/openapi/README.md`

### Validar spec
```bash
cd services/auth-service
npx -y @redocly/cli@latest lint docs/openapi/openapi.v1.yaml
```

### Abrir com Swagger UI (Docker)
```bash
cd deploy/compose
docker compose --env-file .env.auth -f docker-compose.swagger.yml up -d
```

URL local:
- `http://localhost:8085`

## Executar local
```bash
cd services/auth-service
go mod tidy
go run ./cmd/api
```

## Executar com Docker Compose (Mongo + Rabbit + Auth)
```bash
cd deploy/compose
docker compose -f docker-compose.auth.yml up --build
```

Servicos expostos:
- Auth API: `http://localhost:8081`
- RabbitMQ AMQP: `localhost:5672`
- RabbitMQ UI: `http://localhost:15672` (guest/guest)
- MongoDB: `localhost:27017`

## Build da imagem do auth-service
Use o contexto da raiz do repositorio por causa do `replace ../../shared` no `go.mod` do servico.

```bash
cd /caminho/para/food_delivery_platform
docker build -f services/auth-service/Dockerfile -t auth-service:local .
```
