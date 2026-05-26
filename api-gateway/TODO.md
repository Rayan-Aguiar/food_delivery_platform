# API Gateway - TODO Mestre por Fases

## Objetivo geral
Construir o api-gateway de forma incremental, com foco em aprendizado pratico de arquitetura de borda, iniciando pela integracao com o auth-service e evoluindo para seguranca, resiliencia e operacao em ambiente real.

## Por que usar API Gateway
- Centraliza entrada de requests em um ponto unico.
- Evita duplicar autenticacao, rate limit, CORS e observabilidade em cada servico.
- Facilita evolucao de contratos de API para clientes.
- Melhora seguranca de borda (validacoes, headers, politicas).
- Ajuda na governanca de trafego (timeout, retry, circuit breaker).

## Escopo funcional planejado
- Publicar rotas de auth, user, restaurant e order.
- Proteger rotas privadas com JWT.
- Permitir rotas publicas com limite de taxa.
- Propagar request_id, correlation_id e traceparent.

## Definition of Done do gateway (visao final)
- Rotas principais funcionando via proxy para upstreams.
- JWT validado nas rotas privadas.
- Rate limit ativo para rotas publicas e privadas.
- Timeouts e resiliencia por upstream.
- Logs, metricas e tracing ativos na borda.
- Health checks live e ready funcionais.
- Dockerfile e testes basicos de regressao.

---

## Fase 0 - Fundacao minima do gateway
Objetivo da fase:
Preparar base executavel do servico com configuracao, bootstrap HTTP e health checks.

Voce vai aprender:
Como iniciar um servico de borda em Go com configuracao por ambiente e ciclo de vida limpo.

### Checklist
- [ ] Criar modulo Go do api-gateway.
- [ ] Definir estrutura inicial em cmd e internal.
- [ ] Criar carregamento de envs obrigatorias.
- [ ] Subir servidor HTTP com timeout global.
- [ ] Expor /health/live e /health/ready.
- [ ] Integrar logger compartilhado e middleware base (request_id e correlation_id).

### Criterio de pronto
- [ ] Servico sobe localmente sem dependencias externas.
- [ ] Health endpoints respondem 200.
- [ ] Logs de request contem request_id e correlation_id.

---

## Fase 1 - Primeira integracao: gateway <-> auth-service
Objetivo da fase:
Entregar o primeiro valor real: gateway encaminhando rotas de auth para o auth-service.

Voce vai aprender:
Como um reverse proxy funciona na pratica e como o gateway desacopla cliente de servicos internos.

### Checklist
- [ ] Definir upstream AUTH_SERVICE_URL.
- [ ] Mapear rotas publicas de auth:
  - [ ] POST /auth/register
  - [ ] POST /auth/login
  - [ ] POST /auth/refresh
- [ ] Implementar proxy HTTP para auth com preservacao de metodo, path e body.
- [ ] Propagar headers de contexto (request_id e correlation_id).
- [ ] Tratar erros de upstream com resposta padronizada.

### Criterio de pronto
- [ ] Chamadas para /auth/* funcionam via gateway.
- [ ] Sem quebra de contrato esperado pelo cliente.
- [ ] Falhas de upstream retornam erro consistente.

---

## Fase 2 - Router e contratos de roteamento
Objetivo da fase:
Separar roteamento por dominios e preparar escalabilidade de manutencao.

Voce vai aprender:
Como organizar um gateway por contexto de negocio sem virar um monolito de regras.

### Checklist
- [ ] Criar modulo internal/router com tabela de rotas.
- [ ] Definir rotas para user, restaurant e order (mesmo sem implementar tudo).
- [ ] Padronizar mapeamento rota publica vs privada.
- [ ] Adicionar validacao de rotas e metodos permitidos.

### Criterio de pronto
- [ ] Mapa de rotas centralizado e legivel.
- [ ] Inclusao de novas rotas sem alterar codigo espalhado.

---

## Fase 3 - Seguranca de borda (JWT, CORS e headers)
Objetivo da fase:
Aplicar controles de seguranca no gateway antes de abrir novas rotas privadas.

Voce vai aprender:
Como proteger APIs na borda e reduzir superficie de ataque em arquitetura distribuida.

### Checklist
- [ ] Implementar JWTAuthMiddleware com chave publica configuravel.
- [ ] Proteger rotas privadas por regra de roteamento.
- [ ] Configurar CORS restritivo por ambiente.
- [ ] Adicionar headers de seguranca basicos.
- [ ] Definir estrategia para erros 401 e 403 padronizados.

### Criterio de pronto
- [ ] Rotas privadas exigem token valido.
- [ ] Rotas publicas continuam acessiveis sem token.
- [ ] Politica CORS aplicada corretamente.

---

## Fase 4 - Rate limit e politicas de trafego
Objetivo da fase:
Controlar abuso e proteger upstreams de picos e bursts.

Voce vai aprender:
Como aplicar governanca de trafego por IP/token sem comprometer experiencia do cliente.

### Checklist
- [ ] Implementar RateLimitMiddleware configuravel (RATE_LIMIT_RPS).
- [ ] Definir limite diferenciado para rotas publicas e privadas.
- [ ] Implementar resposta 429 padronizada.
- [ ] Registrar metricas de rejeicao.

### Criterio de pronto
- [ ] Excesso de requests recebe 429 com contrato estavel.
- [ ] Upstreams ficam protegidos contra burst simples.

---

## Fase 5 - Observabilidade de borda
Objetivo da fase:
Instrumentar o gateway para diagnosticar latencia, erros e gargalos.

Voce vai aprender:
Como medir comportamento da borda e criar base para operacao orientada a dados.

### Checklist
- [ ] Implementar logs estruturados por rota e upstream.
- [ ] Expor metricas Prometheus:
  - [ ] gateway_http_requests_total
  - [ ] gateway_http_request_duration_seconds
  - [ ] gateway_rate_limit_rejections_total
  - [ ] gateway_upstream_errors_total
- [ ] Implementar TracingMiddleware com span pai por request.
- [ ] Propagar traceparent para upstreams.

### Criterio de pronto
- [ ] Latencia e erros visiveis em metricas e logs.
- [ ] Traces distribuidos permitem seguir chamadas entre servicos.

---

## Fase 6 - Resiliencia de chamadas upstream
Objetivo da fase:
Evitar falhas em cascata quando um servico interno estiver degradado.

Voce vai aprender:
Como timeout, retry e circuit breaker mudam o comportamento do sistema em cenarios reais de falha.

### Checklist
- [ ] Definir timeout por rota/upstream.
- [ ] Implementar retry somente para GET idempotente.
- [ ] Implementar circuit breaker por upstream.
- [ ] Mapear erros de resiliencia para respostas claras ao cliente.

### Criterio de pronto
- [ ] Gateway nao bloqueia indefinidamente em upstream lento.
- [ ] Falhas repetidas ativam protecao de circuito.
- [ ] Retry nao causa duplicidade de operacoes nao idempotentes.

---

## Fase 7 - Expansao de dominios de negocio
Objetivo da fase:
Adicionar gradualmente rotas de user, restaurant e order com regras de acesso corretas.

Voce vai aprender:
Como escalar um gateway multi-dominio mantendo organizacao e clareza de responsabilidades.

### Checklist
- [ ] User:
  - [ ] GET /users/me
  - [ ] PUT /users/me
  - [ ] GET /users/me/orders
- [ ] Restaurant:
  - [ ] GET /restaurants
  - [ ] GET /restaurants/{id}
  - [ ] GET /restaurants/{id}/menu
- [ ] Order:
  - [ ] POST /orders
  - [ ] GET /orders/{id}
- [ ] Revisar regras publica/privada por rota.

### Criterio de pronto
- [ ] Rotas essenciais dos dominios publicadas no gateway.
- [ ] Controle de acesso respeita regras de negocio.

---

## Fase 8 - Testes do gateway
Objetivo da fase:
Garantir confiabilidade de evolucao com cobertura de cenarios criticos.

Voce vai aprender:
Como testar middleware, roteamento e proxy sem depender de ambiente manual.

### Checklist
- [ ] Testes unitarios de middlewares (auth, rate limit, recovery).
- [ ] Testes de integracao de roteamento com upstream mock.
- [ ] Cenarios obrigatorios:
  - [ ] token invalido
  - [ ] rate limit excedido
  - [ ] timeout de upstream
  - [ ] propagacao de headers de contexto

### Criterio de pronto
- [ ] Suite minima de testes rodando em CI local.
- [ ] Falhas comuns detectadas antes de deploy.

---

## Fase 9 - Docker e operacao local (divida tecnica por enquanto)
Objetivo da fase:
Containerizar gateway e facilitar execucao padrao no ambiente de desenvolvimento.

Voce vai aprender:
Como empacotar servico de borda para reproducibilidade e integracao via compose.

Status atual:
- [ ] Adiada (divida tecnica)

### Checklist
- [ ] Criar Dockerfile multi-stage.
- [ ] Definir envs e portas no compose.
- [ ] Validar healthchecks no container.

### Criterio de pronto
- [ ] Gateway sobe via Docker com configuracao previsivel.

---

## Fase 10 - OpenAPI agregada e governanca de contrato (divida tecnica por enquanto)
Objetivo da fase:
Documentar a API de borda e preparar governanca de contratos para clientes.

Voce vai aprender:
Como manter documentacao viva no ponto unico de entrada da plataforma.

Status atual:
- [ ] Adiada (divida tecnica)

### Checklist
- [ ] Criar docs/openapi do gateway.
- [ ] Descrever contratos expostos na borda.
- [ ] Validar consistencia entre rota publicada e comportamento real.

### Criterio de pronto
- [ ] Contrato de borda versionado e confiavel para consumidores.

---

## Ordem recomendada de execucao (aprendizado incremental)
1. Fase 0
2. Fase 1
3. Fase 2
4. Fase 3
5. Fase 4
6. Fase 5
7. Fase 6
8. Fase 7
9. Fase 8
10. Fase 9 (quando retirar da divida tecnica)
11. Fase 10 (quando retirar da divida tecnica)
