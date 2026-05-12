# Auth Service - Plano Detalhado

## 1. Visao geral
Servico responsavel por identidade: registro, login, emissao de JWT, refresh token e revogacao.

## 2. Responsabilidades
- Cadastro de usuario (credenciais).
- Autenticacao por email/senha.
- Emissao e renovacao de access token.
- Revogacao de refresh token.
- Publicacao de evento de usuario criado para sincronizacao minima.

## 3. Regras de negocio
- Email unico no sistema.
- Senha com politica minima (tamanho, complexidade).
- Refresh token com expiracao e revogavel.
- Bloqueio progressivo apos tentativas invalidas (opcional fase 2).

## 4. Arquitetura interna
- `domain`: entidades `Credential`, `SessionToken`.
- `application`: casos de uso de auth.
- `infrastructure`: Mongo repository, hash provider, JWT provider.
- `delivery/http`: handlers REST.

## 5. Estrutura de pastas
```txt
services/auth-service/
  cmd/api/main.go
  internal/domain/
  internal/application/
  internal/infrastructure/mongo/
  internal/infrastructure/security/
  internal/delivery/http/
  internal/delivery/messaging/
  configs/
  docs/openapi/
  tests/
  go.mod
  Dockerfile
```

## 6. Models e collections MongoDB
- `credentials`
  - `_id`, `user_id`, `email`, `password_hash`, `status`, `created_at`, `updated_at`
  - indices: `email` unico
- `refresh_tokens`
  - `_id`, `user_id`, `token_hash`, `expires_at`, `revoked_at`, `ip`, `user_agent`
  - indices: `user_id`, `expires_at` TTL

## 7. Endpoints
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`
- `POST /auth/logout`
- `GET /auth/health`

## 8. Eventos
Produzidos:
- `user.auth.registered.v1`
  - payload: `user_id`, `email`, `registered_at`
- `auth.login.succeeded.v1`
  - payload: `user_id`, `logged_at`

Consumidos:
- Nenhum obrigatorio na primeira fase.

## 9. Use cases
- RegisterUser
- LoginUser
- RefreshAccessToken
- LogoutSession
- ValidateToken (uso interno)

## 10. Handlers, repositories e services
Handlers:
- RegisterHandler
- LoginHandler
- RefreshHandler
- LogoutHandler

Repositories:
- CredentialRepository
- RefreshTokenRepository

Services:
- PasswordHasher
- TokenService
- AuthAuditService

## 11. Middlewares e seguranca
- Request ID.
- Correlation ID.
- Recovery + timeout.
- Rate limit por IP/rota sensivel.
- Validacao de payload.

## 12. Observabilidade
Logs:
- `auth_action`, `user_id`, `request_id`, `correlation_id`.

Metricas:
- `auth_login_attempts_total`
- `auth_login_failures_total`
- `auth_token_refresh_total`

Tracing:
- spans por endpoint e por acesso a Mongo.

## 13. Retry, DLQ e idempotencia
- Nao aplica fila critica local inicialmente.
- Idempotencia de `register` por `email` + `idempotency-key` opcional no gateway.

## 14. Testes
- Unitarios: policy senha, token, use cases.
- Integracao: Mongo real em container.
- Contrato HTTP: respostas e codigos.
- Seguranca: senha invalida, token expirado, replay de refresh.

## 15. Dockerizacao e operacao
- Dockerfile multi-stage (build Go + runtime distroless/alpine).
- Variaveis:
  - `HTTP_PORT`
  - `MONGO_URI`
  - `MONGO_DB_NAME`
  - `JWT_SECRET`
  - `JWT_ACCESS_TTL`
  - `JWT_REFRESH_TTL`
  - `LOG_LEVEL`
- Healthcheck:
  - `/health/live`
  - `/health/ready`

## 16. Swagger
- Documentar endpoints e modelos de erro padrao.

## 17. Dependencias externas
- MongoDB.
- RabbitMQ (somente publisher de evento opcional).

## 18. TODO detalhado
1. Inicializar modulo e bootstrap do servico.
2. Definir entidades e contratos de repositorio.
3. Implementar repositorios Mongo e indices.
4. Implementar hash de senha e JWT.
5. Implementar endpoints `register/login/refresh/logout`.
6. Adicionar middlewares de observabilidade e seguranca.
7. Adicionar metricas e tracing.
8. Gerar OpenAPI.
9. Criar testes unitarios e integracao.
10. Criar Dockerfile e compose override para ambiente local.
