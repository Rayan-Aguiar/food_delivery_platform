# Auth Service - Fase 0

## Variaveis de ambiente
- `SERVICE_NAME` (default: `auth-service`)
- `HTTP_PORT` (default: `8081`)
- `LOG_LEVEL` (default: `info`)
- `REQUEST_TIMEOUT_SECONDS` (default: `10`)
- `MONGO_URI` (default: `mongodb://localhost:27017`)
- `MONGO_DB_NAME` (default: `auth_db`)
- `RABBITMQ_URL` (default: `amqp://guest:guest@localhost:5672/`)

## Endpoints disponiveis na fase 0
- `GET /health/live`
- `GET /health/ready`
- `GET /auth/health`

## Executar local
```bash
cd services/auth-service
go mod tidy
go run ./cmd/api
```
