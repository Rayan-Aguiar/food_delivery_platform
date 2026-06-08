# Restaurant Service - TODO Mestre por Fases

## Objetivo geral
Construir o restaurant-service de forma incremental, com foco em dominio de catalogo (restaurantes e cardapio), validacao para o fluxo de pedidos e boas praticas de arquitetura em microsservicos.

## Por que este servico existe
- Centraliza regras de catalogo e disponibilidade de itens.
- Evita duplicacao de regra de menu no order-service.
- Entrega consultas publicas para app/web via gateway.
- Suporta validacao interna de itens para criacao de pedido.

## Escopo funcional planejado
- Consulta publica de restaurantes e menu.
- CRUD administrativo de restaurante e itens.
- Endpoint interno de validacao de itens para order-service.
- Publicacao de eventos de mudanca de menu/disponibilidade.

## Definition of Done do restaurant-service (visao final)
- Endpoints publicos e admin funcionais.
- Endpoint interno de validacao pronto para order-service.
- Persistencia Mongo com indices adequados.
- Logs, metricas e tracing ativos.
- Testes unitarios e de integracao cobrindo fluxos criticos.
- OpenAPI e Dockerfile prontos.

---

## Fase 0 - Fundacao tecnica
Objetivo da fase:
Criar base executavel do servico com bootstrap HTTP, configuracao por ambiente e health checks.

Voce vai aprender:
Como iniciar um novo microsservico padronizado no monorepo com ciclo de vida limpo.

Status atual:
- [x] Concluida

### Checklist
- [x] Criar modulo Go do restaurant-service.
- [x] Definir estrutura inicial em cmd e internal.
- [x] Criar configuracao por env (`HTTP_PORT`, `MONGO_URI`, `MONGO_DB_NAME`).
- [x] Subir servidor HTTP com timeout global.
- [x] Expor `/health/live` e `/health/ready`.
- [x] Integrar logger e middlewares base (request_id, correlation_id, recovery, access log).

### Criterio de pronto
- [x] Servico sobe localmente.
- [x] Health endpoints respondem 200.
- [x] Logs de request com request_id e correlation_id.

---

## Fase 1 - Modelagem de dominio de catalogo
Objetivo da fase:
Modelar entidades e regras centrais de restaurante e menu sem acoplamento com infraestrutura.

Voce vai aprender:
Como traduzir regras de negocio em entidades, value objects e contratos de repositorio.

Status atual:
- [x] Concluida

### Checklist
- [x] Modelar `Restaurant` (id, nome, endereco, telefone, status, taxa de entrega `float64`).
- [x] Modelar `MenuItem` (id, restaurant_id, nome, preco, categoria, available).
- [x] Restringir categorias do menu para: `pizzaria`, `hamburgueria`, `japones`, `comida brasileira`, `sorveteria`.
- [x] Definir estados de restaurante (ativo/inativo) e invariantes.
- [x] Definir interfaces `RestaurantRepository` e `MenuRepository`.
- [x] Definir erros de dominio para item indisponivel e restaurante inativo.

### Criterio de pronto
- [x] Dominio compila sem dependencia de Mongo/HTTP.
- [x] Regras de dominio cobertas por testes unitarios.

---

## Fase 2 - Casos de uso de consulta publica
Objetivo da fase:
Implementar fluxos de leitura para listagem e detalhes de restaurantes/menu.

Voce vai aprender:
Como organizar casos de uso de leitura com separacao clara entre aplicacao e infraestrutura.

### Checklist
- [ ] Implementar `ListRestaurants`.
- [ ] Implementar `GetRestaurantDetails`.
- [ ] Implementar `ListRestaurantMenu`.
- [ ] Definir filtros/paginacao inicial de leitura (quando aplicavel).

### Criterio de pronto
- [ ] Casos de uso retornam dados esperados para consultas publicas.
- [ ] Testes unitarios cobrindo cenarios basicos e limites.

---

## Fase 3 - Persistencia MongoDB e indices
Objetivo da fase:
Conectar dominio ao MongoDB com repositorios e desempenho minimo esperado.

Voce vai aprender:
Como modelar collections e indices para consultas de catalogo.

### Checklist
- [ ] Implementar repositorio Mongo de restaurantes.
- [ ] Implementar repositorio Mongo de menu.
- [ ] Criar collections `restaurants` e `menu_items`.
- [ ] Garantir indices:
  - [ ] `restaurant_id` em `menu_items`
  - [ ] `available` em `menu_items`
  - [ ] indice de texto por nome (restaurante/item)

### Criterio de pronto
- [ ] Leitura publica funcionando com dados persistidos no Mongo.
- [ ] Indices aplicados automaticamente no startup.

---

## Fase 4 - API publica HTTP
Objetivo da fase:
Expor os endpoints publicos de catalogo para consumo via API Gateway.

Voce vai aprender:
Como transformar casos de uso em handlers HTTP com contratos estaveis.

### Checklist
- [ ] Implementar `GET /restaurants`.
- [ ] Implementar `GET /restaurants/{id}`.
- [ ] Implementar `GET /restaurants/{id}/menu`.
- [ ] Padronizar respostas e erros HTTP.

### Criterio de pronto
- [ ] Endpoints publicos funcionam ponta a ponta.
- [ ] Contratos HTTP consistentes e previsiveis.

---

## Fase 5 - API administrativa (CRUD de catalogo)
Objetivo da fase:
Permitir operacoes de manutencao do catalogo por rotas protegidas de administracao.

Voce vai aprender:
Como separar leitura publica de escrita administrativa com seguranca de acesso.

### Checklist
- [ ] Implementar `POST /restaurants` (admin).
- [ ] Implementar `POST /restaurants/{id}/menu/items` (admin).
- [ ] Implementar `PATCH /restaurants/{id}/menu/items/{itemId}` (admin).
- [ ] Preparar idempotencia opcional por `idempotency-key` para upsert admin.

### Criterio de pronto
- [ ] CRUD admin funcional e validado.
- [ ] Alteracoes de catalogo refletidas nas consultas publicas.

---

## Fase 6 - Endpoint interno de validacao para pedidos
Objetivo da fase:
Entregar contrato interno para order-service validar itens e disponibilidade antes de criar pedido.

Voce vai aprender:
Como construir endpoint interno orientado a regra critica de negocio (hot path).

### Checklist
- [ ] Implementar `POST /restaurants/{id}/menu/validate`.
- [ ] Validar restaurante ativo/inativo.
- [ ] Validar existencia e disponibilidade de itens.
- [ ] Validar preco retornado para composicao de pedido.

### Criterio de pronto
- [ ] Endpoint interno retorna validacao confiavel para o order-service.
- [ ] RN003 coberta: item indisponivel nao pode ser pedido.

---

## Fase 7 - Seguranca e politicas de acesso
Objetivo da fase:
Aplicar controles de acesso em rotas admin e limites em consultas publicas.

Voce vai aprender:
Como proteger endpoints de escrita e evitar abuso em endpoints de leitura.

### Checklist
- [ ] Middleware JWT em rotas admin.
- [ ] Rate limit para consultas publicas.
- [ ] Revisao de erros 401/403/429 padronizados.

### Criterio de pronto
- [ ] Rotas admin protegidas corretamente.
- [ ] Rotas publicas com protecao basica contra burst.

---

## Fase 8 - Eventos de catalogo
Objetivo da fase:
Publicar eventos de mudanca para integracao com outros servicos e analytics.

Voce vai aprender:
Como emitir eventos versionados de alteracao de estado de dominio.

### Checklist
- [ ] Publicar `restaurant.menu.updated.v1`.
- [ ] Publicar `restaurant.availability.changed.v1`.
- [ ] Incluir correlation_id e envelope padrao.
- [ ] Definir estrategia minima de retry de publish.

### Criterio de pronto
- [ ] Eventos publicados com contrato consistente.
- [ ] Consumidores conseguem reagir a mudancas de catalogo.

---

## Fase 9 - Observabilidade e qualidade
Objetivo da fase:
Instrumentar o servico para operacao confiavel e diagnostico rapido.

Voce vai aprender:
Como medir comportamento de endpoints publicos e internos com sinais de operacao.

### Checklist
- [ ] Logs estruturados por endpoint.
- [ ] Metricas de leitura de menu e validacao.
- [ ] Tracing na rota de validacao de itens.
- [ ] Testes unitarios de regras de disponibilidade e preco.
- [ ] Testes de integracao Mongo.
- [ ] Teste de contrato do endpoint de validacao.

### Criterio de pronto
- [ ] Observabilidade minima ativa.
- [ ] Suite de testes cobrindo fluxos criticos.

---

## Fase 10 - Documentacao e empacotamento
Objetivo da fase:
Finalizar entrega com contrato documentado e empacotamento para execucao padronizada.

Voce vai aprender:
Como preparar servico para consumo por times e execucao em ambiente containerizado.

### Checklist
- [ ] Gerar/atualizar OpenAPI com rotas publicas, admin e interna.
- [ ] Criar Dockerfile multi-stage.
- [ ] Validar variaveis de ambiente no README do servico.

### Criterio de pronto
- [ ] OpenAPI publica e valida.
- [ ] Build de imagem funcionando.
- [ ] Servico pronto para compose e integracao no roadmap.

---

## Ordem recomendada de execucao
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
