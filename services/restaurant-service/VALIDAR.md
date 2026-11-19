# Validar antes de mergear na main

Este trabalho (observabilidade, testes, OpenAPI, Dockerfile e compose do
restaurant-service) foi commitado **sem validação completa**: o PC antigo não
aguentava rodar os containers Docker. Esta branch existe para não perder o
código; só entre na `main` depois de passar pelo checklist abaixo no PC novo.

Apague este arquivo no commit que concluir a validação.

## Já verificado (sem Docker, 2026-11-19)

- [x] `go build ./...` compila
- [x] `go vet ./...` sem avisos
- [x] `go test ./...` passa (testes unitários de application, http, entities,
      messaging, mongo, security e observability)

## Falta validar (precisa de Docker)

Rodar a partir de `deploy/compose/`, com `.env.restaurant` criado a partir de
`.env.restaurant.example` (o `JWT_SECRET` tem que ser o mesmo do auth-service).

### 1. Testes de integração com Mongo
- [ ] Subir só o Mongo do stack de restaurante
- [ ] `MONGO_TEST_URI=mongodb://localhost:27018 go test -tags=integration ./internal/infrastructure/mongo/...`
      (rodar em `services/restaurant-service/`)

### 2. Build da imagem
- [ ] `docker compose -f docker-compose.restaurant.yml --env-file .env.restaurant build restaurant-service`
      (o contexto é a raiz do repo por causa do `replace ../../shared`)
- [ ] Imagem roda como usuário `app`, sem root

### 3. Serviço no compose
- [ ] `docker compose -f docker-compose.restaurant.yml --env-file .env.restaurant up -d`
- [ ] `restaurant-service` fica `healthy` (healthcheck em `/health/live`)
- [ ] Conecta no Mongo (`restaurant-mongo`) e no RabbitMQ (`restaurant-rabbitmq`)
- [ ] Rotas admin aceitam um token emitido pelo auth-service (mesmo `JWT_SECRET` e `JWT_ISSUER=auth-service`)

### 4. Comportamentos novos (testar com o serviço no ar)
- [ ] `GET /restaurants?page=1&page_size=10`: paginação e `has_more` corretos; `page_size` acima de 100 é limitado
- [ ] Body acima de 64 KiB retorna 413 `PAYLOAD_TOO_LARGE`
- [ ] Validação de itens com mais de 100 `item_ids` é rejeitada
- [ ] `POST /restaurants/{id}/menu/items` repetido com a mesma `Idempotency-Key` devolve o item original (antes dava 500)
- [ ] A mesma `Idempotency-Key` usada em outro restaurante retorna 409
- [ ] Evento de catálogo chega na fila do RabbitMQ (painel em http://localhost:15673)

### 5. Observabilidade
- [ ] `GET http://localhost:8083/metrics` expõe `http_request_duration_seconds` e `restaurant_menu_item_validations_total`
- [ ] Prometheus (stack de observabilidade) mostra o alvo `restaurant-service` como UP em http://localhost:9090/targets
      (coleta via `host.docker.internal:8083`; confirmar que isso funciona no Docker do PC novo)
- [ ] Com `OTEL_ENABLED=true` e `OTEL_EXPORTER_OTLP_ENDPOINT` apontando para um coletor, aparecem spans da rota de validação e do Mongo
- [ ] Logs saem em JSON com `request_id`/`correlation_id`

### 6. OpenAPI / Swagger
- [ ] Lint do `docs/openapi/openapi.v1.yaml` (ex.: `npx @redocly/cli lint docs/openapi/openapi.v1.yaml`)
- [ ] Swagger UI (http://localhost:8085) mostra o seletor com `auth-service` e `restaurant-service`, e os dois specs carregam

## Pendências conhecidas (não bloqueiam o merge)

Já anotadas no `TODO.md`:
- Endpoint interno de validação sem autenticação serviço-a-serviço
- `/health/ready` não checa Mongo/RabbitMQ
- Retry de `POST /restaurants` com a mesma `Idempotency-Key` faz upsert em vez de devolver o original

## Quando tudo passar

1. Marcar no `TODO.md` os critérios de pronto da última fase
2. Apagar este `VALIDAR.md`
3. Mergear a branch na `main`
