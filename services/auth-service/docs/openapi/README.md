# OpenAPI - Auth Service

## Versoes da spec
- `openapi.v1.yaml`: versao atual do contrato HTTP do auth-service.

## Escopo documentado
- Health checks (`/health/live`, `/health/ready`, `/auth/health`)
- Endpoints de autenticacao (`/auth/register`, `/auth/login`, `/auth/refresh`, `/auth/logout`)
- Modelos de request/response
- Erros padronizados
- Exemplos de payload

## Regras de versionamento
- Nao alterar retroativamente `openapi.v1.yaml` com breaking changes.
- Para breaking changes, criar nova spec com sufixo de versao (ex.: `openapi.v2.yaml`).
- Mudancas aditivas e retrocompativeis podem incrementar o campo `info.version` da mesma major.
