# Auth Service - TODO Mestre de Implementacao

## Objetivo
Construir o auth-service completo em Go, com Clean Architecture, MongoDB, JWT + Refresh Token, observabilidade, testes e containerizacao, pronto para execucao local via Docker e integracao com o restante da plataforma.

## Escopo funcional (requisitos da doc)
- Registro de usuario
- Login
- Refresh token
- Logout (revogacao de refresh token)
- Health endpoints
- Publicacao de eventos de auth

## Escopo tecnico obrigatorio
- Clean Architecture
- Repository Pattern
- Dependency Injection
- Logs estruturados com request_id e correlation_id
- Tracing distribuido (OpenTelemetry)
- Metricas Prometheus
- Testes unitarios e integracao
- Dockerfile multi-stage
- Documentacao OpenAPI

## Regras de negocio a cumprir
- Email unico
- Senha com politica minima
- Refresh token com expiracao e revogacao
- Login falho nao gera token
- Refresh invalido/expirado/revogado deve falhar

## Dependencias externas
- MongoDB
- RabbitMQ (publisher de eventos)
- Shared module local (contracts, events, logger, middleware, errors, broker, utils)

## Estrutura alvo do servico
- cmd/api/main.go
- internal/domain
- internal/application
- internal/infrastructure
- internal/delivery/http
- internal/delivery/messaging
- configs
- docs/openapi
- tests
- go.mod
- Dockerfile

## Definition of Done do auth-service
- Endpoints principais funcionais e testados
- Erros padronizados
- Logs estruturados em todas as rotas
- Metricas e tracing ativos
- Mongo com indices criados automaticamente
- Publicacao de eventos de registro/login
- OpenAPI gerada e valida
- Docker image buildando e container subindo
- Testes unitarios e integracao passando

---

## Fase 0 - Kickoff tecnico
Objetivo: preparar base minima do servico para iniciar desenvolvimento.

### Checklist
- [x] Criar modulo Go proprio do auth-service
- [x] Configurar replace para shared local
- [x] Definir arquivo de configuracao por env
- [x] Criar bootstrap inicial do servidor HTTP
- [x] Criar endpoints de health: /health/live e /health/ready
- [x] Integrar logger e middlewares compartilhados
- [x] Criar padrao de resposta de erro HTTP
- [x] Configurar timeout global de requests

### Criterio de pronto
- [x] Servico sobe localmente
- [x] Health endpoints respondem 200
- [x] Requests geram logs com request_id e correlation_id

---

## Fase 1 - Dominio e contratos internos
Objetivo: modelar regras de autenticacao sem acoplamento com infraestrutura.

### Entidades
- [x] Credential
- [x] RefreshSession
- [x] AuthTokens

### Value Objects e regras
- [x] Email valido
- [x] PasswordPolicy (minimo de seguranca)
- [x] TokenTTL

### Interfaces
- [x] CredentialRepository
- [x] RefreshTokenRepository
- [x] PasswordHasher
- [x] TokenService
- [x] Clock (para testabilidade)
- [x] IDGenerator

### Criterio de pronto
- [x] Compila sem infraestrutura real
- [x] Regras de dominio cobertas por testes unitarios

---

## Fase 2 - Application use cases
Objetivo: implementar casos de uso com fluxo de negocio do auth.

### Casos de uso
- [x] RegisterUser
- [x] LoginUser
- [x] RefreshAccessToken
- [x] LogoutSession
- [x] ValidateAccessToken (uso interno)

### Regras por caso de uso
- RegisterUser
  - [x] Validar payload e senha
  - [x] Garantir email unico
  - [x] Hash de senha
  - [x] Persistir credencial
  - [x] Criar refresh session inicial (opcional no register)
- LoginUser
  - [x] Buscar credencial por email
  - [x] Comparar hash
  - [x] Gerar access token + refresh token
  - [x] Persistir hash do refresh token
- RefreshAccessToken
  - [x] Validar refresh token
  - [x] Verificar revogacao e expiracao
  - [x] Rotacionar refresh token
  - [x] Invalidar token anterior
- LogoutSession
  - [x] Revogar refresh token ativo

### Criterio de pronto
- [x] Todos os casos de uso com testes unitarios
- [x] Erros de dominio mapeados para erros padrao

---

## Fase 3 - Infra de seguranca
Objetivo: implementar provedores tecnicos de criptografia e token.

### Hash de senha
- [x] Implementar com bcrypt
- [x] Definir custo por ambiente
- [x] Cobrir erro de hash/compare

### JWT
- [x] Gerar access token com claims minimas: sub, iat, exp
- [x] Assinatura HMAC (fase atual)
- [x] Validacao de assinatura e expiracao
- [x] Preparar suporte futuro para RSA/JWKS

### Refresh token
- [x] Gerar token aleatorio seguro
- [x] Persistir somente hash do refresh token
- [x] Definir TTL configuravel

### Criterio de pronto
- [x] Testes de seguranca para tokens validos e invalidos
- [x] Sem persistir segredos em claro

---

## Fase 4 - Infra MongoDB
Objetivo: implementar persistencia e garantias de dados.

### Repositorios
- [x] MongoCredentialRepository
- [x] MongoRefreshTokenRepository

### Collections
- [x] credentials
- [x] refresh_tokens
- [ ] auth_audit (opcional fase 2)

### Indices
- [x] unique index em email
- [x] index por user_id
- [x] TTL index em expires_at para refresh_tokens

### Criterio de pronto
- [x] CRUD necessario dos casos de uso funcionando
- [x] Violacao de email unico tratada corretamente

---

## Fase 5 - Delivery HTTP
Objetivo: expor API REST final do auth-service.

### Endpoints
- [x] POST /auth/register
- [x] POST /auth/login
- [x] POST /auth/refresh
- [x] POST /auth/logout
- [x] GET /auth/health

### DTOs
- [x] Request/Response para register
- [x] Request/Response para login
- [x] Request/Response para refresh
- [x] Request/Response para logout

### Middlewares
- [x] Request ID
- [x] Correlation ID
- [x] Recovery
- [x] Access log
- [x] Timeout
- [x] Rate limit para login/refresh

### Criterio de pronto
- [x] Contratos HTTP estaveis e testados
- [x] Erros padronizados em todas as rotas

---

## Fase 6 - Eventos e mensageria
Objetivo: publicar eventos de autenticacao para integracao com outros servicos.

### Eventos produzidos
- [ ] user.auth.registered.v1
- [ ] auth.login.succeeded.v1

### Publicacao
- [ ] Integrar broker compartilhado
- [ ] Definir exchange e routing keys
- [ ] Propagar correlation_id nos headers

### Confiabilidade
- [ ] Estrategia minima de retry de publish
- [ ] Log de falha de publish com contexto
- [ ] Preparar suporte a outbox (fase posterior)

### Criterio de pronto
- [ ] Eventos publicados com envelope padrao
- [ ] Payload validado e versionado

---

## Fase 7 - Observabilidade completa
Objetivo: instrumentar servico para operacao real.

### Logs
- [ ] Log estruturado por endpoint
- [ ] Campos obrigatorios: service, request_id, correlation_id, status, duration

### Metricas
- [ ] auth_login_attempts_total
- [ ] auth_login_failures_total
- [ ] auth_token_refresh_total
- [ ] http_request_duration_seconds

### Tracing
- [ ] Instrumentacao HTTP server
- [ ] Instrumentacao Mongo operations
- [ ] Instrumentacao publish de eventos

### Criterio de pronto
- [ ] Dados observaveis no ambiente local
- [ ] Erros e latencia rastreaveis ponta a ponta

---

## Fase 8 - Testes completos
Objetivo: garantir qualidade e regressao controlada.

### Unitarios
- [ ] Dominio (password policy, email, expiracao)
- [ ] Use cases (register/login/refresh/logout)
- [ ] Security providers (hash/jwt/refresh)

### Integracao
- [ ] Repositorios Mongo com container
- [ ] Endpoints HTTP com banco real
- [ ] Publicacao de evento com broker de teste

### Contrato
- [ ] Schemas de request/response
- [ ] Schemas de eventos publicados

### Seguranca
- [ ] Senha invalida
- [ ] Credencial invalida
- [ ] Token expirado
- [ ] Token revogado
- [ ] Reuso de refresh token rotacionado

### Criterio de pronto
- [ ] go test ./... verde no auth-service
- [ ] Cobertura minima acordada (ex.: 80%+)

---

## Fase 9 - Documentacao OpenAPI
Objetivo: padronizar contrato para gateway e consumidores.

### Checklist
- [ ] Documentar endpoints e modelos
- [ ] Documentar erros padrao
- [ ] Documentar exemplos de payload
- [ ] Versionar spec

### Criterio de pronto
- [ ] Spec valida e revisada

---

## Fase 10 - Docker e runtime local
Objetivo: rodar auth-service conteinerizado com dependencias.

### Dockerfile
- [ ] Multi-stage build
- [ ] Imagem final enxuta
- [ ] Usuario nao-root
- [ ] Healthcheck HTTP

### Runtime
- [ ] Variaveis de ambiente documentadas
- [ ] Conexao Mongo por env
- [ ] Conexao Rabbit por env
- [ ] Exposicao de porta HTTP

### Compose
- [ ] Integracao com compose local do projeto
- [ ] Dependencia de Mongo e Rabbit

### Criterio de pronto
- [ ] docker build concluido
- [ ] container sobe e responde health
- [ ] endpoints auth funcionam em ambiente containerizado

---

## Fase 11 - Hardening e pendencias finais
Objetivo: fechar requisitos nao-funcionais e preparacao para proximas integracoes.

### Checklist
- [ ] Revisar tratamento de erros e mensagens seguras
- [ ] Revisar retries, timeouts e limites
- [ ] Revisar consistencia de logs e metricas
- [ ] Revisar mapeamento de requisitos funcionais e nao funcionais
- [ ] Revisar readiness para integracao com api-gateway

### Criterio de pronto
- [ ] Auth-service pronto para ser consumido pelo gateway
- [ ] Checklist inteiro concluido

---

## Matriz de requisitos atendidos por esta implementacao
- RF001 cadastro de usuarios: fases 2, 5
- RF002 autenticacao JWT: fases 2, 3, 5
- RNF003 observabilidade: fase 7
- RNF004 logs estruturados: fases 0, 7
- RNF006 containerizacao: fase 10
- RNF007 documentacao de APIs: fase 9
- RNF008 testes automatizados: fase 8
- RNF009 rastreabilidade request/evento: fases 0, 6, 7

## Ordem de execucao recomendada (sem pular)
1. Fase 0
2. Fase 1
3. Fase 2
4. Fase 3
5. Fase 4
6. Fase 5
7. Fase 6
8. Fase 7
9. Fase 8
10. Fase 9
11. Fase 10
12. Fase 11

## Regras de trabalho durante a execucao
- Nao iniciar fase nova com teste da fase anterior quebrado
- Nao acoplar regra de negocio na camada shared
- Nao quebrar contratos HTTP/eventos sem versionamento
- Todo endpoint novo deve nascer com teste
- Toda mudanca relevante deve atualizar este TODO
