# OpenAPI - Restaurant Service

## Versoes da spec
- `openapi.v1.yaml`: versao atual do contrato HTTP do restaurant-service.

## Escopo documentado
- Health e operacao (`/health/live`, `/health/ready`, `/metrics`)
- Catalogo publico (`GET /restaurants`, `GET /restaurants/{id}`, `GET /restaurants/{id}/menu`)
- Administracao com JWT (`POST /restaurants`, `POST /restaurants/{id}/menu/items`, `PATCH /restaurants/{id}/menu/items/{itemId}`)
- Contrato interno para o order-service (`POST /restaurants/{id}/menu/validate`)
- Erros padronizados (`ErrorResponse`) e exemplos de payload

## Regras de versionamento
- Nao alterar retroativamente `openapi.v1.yaml` com breaking changes.
- Para breaking changes, criar nova spec com sufixo de versao (ex.: `openapi.v2.yaml`).
- Mudancas aditivas e retrocompativeis podem incrementar `info.version` da mesma major.
- O contrato do endpoint de validacao tambem e' fixado por teste
  (`internal/delivery/http/validate_contract_test.go`): mude os dois juntos.
