# Roadmap de Implementacao - Plataforma Completa

## 1. Dependencias entre microsservicos
- `api-gateway` depende de `auth-service`, `user-service`, `restaurant-service`, `order-service`.
- `order-service` depende sync de `restaurant-service` e async de `payment-service` e `delivery-service`.
- `payment-service` depende de evento `order.created.v1`.
- `delivery-service` depende de evento `delivery.requested.v1` (emitido pelo order-service apos aprovacao financeira).
- `notification-service` depende dos eventos de order/payment/delivery.
- `user-service` pode depender de evento `user.auth.registered.v1` e opcionalmente `order.status.changed.v1`.

## 2. Ordem recomendada de desenvolvimento
1. Fundacao compartilhada (`shared/contracts`, `shared/events`, `shared/logger`, `shared/middleware`, `shared/broker`).
2. auth-service (MongoDB).
3. api-gateway (com auth pronto).
4. restaurant-service (MongoDB).
5. order-service (PostgreSQL, happy path sem compensacoes completas).
6. payment-service (PostgreSQL) + integracao de evento.
7. delivery-service (MongoDB) + integracao de evento.
8. notification-service (MongoDB).
9. compensacoes avancadas, estorno e hardening de resiliencia.
10. observabilidade e testes E2E finais.

## 3. Marco por iteracao
- Iteracao A: autenticacao + catalogo + gateway.
- Iteracao B: criacao de pedido + pagamento aprovado.
- Iteracao C: entrega iniciada/concluida + notificacoes.
- Iteracao D: compensacoes da saga e estorno.
- Iteracao E: performance, seguranca, chaos/resilience tests.

## 4. Riscos criticos e mitigacao
- Risco: quebra de contrato de evento.
  - Mitigacao: versionamento de eventos + testes de contrato CI.
- Risco: duplicidade de consumo de mensagem.
  - Mitigacao: tabela `processed_events` + chaves idempotentes.
- Risco: crescimento de backlog de fila.
  - Mitigacao: autoscaling de consumers por lag + alertas Prometheus.
- Risco: indisponibilidade de servico central (order-service).
  - Mitigacao: replicas, timeout, circuit breaker e fallback de consulta.

## 5. Padrao minimo de pronto por servico (Definition of Done)
- Endpoints implementados + OpenAPI.
- Health endpoints liveness/readiness.
- Logs estruturados com correlation ID.
- Metricas Prometheus + tracing OpenTelemetry.
- Retry, DLQ e idempotencia quando consumir eventos.
- Testes unitarios e integracao passando.
- Dockerfile pronto e integrado ao compose.

## 6. Decisoes arquiteturais e trade-offs
- Orquestracao de saga no `order-service`:
  - Beneficio: visibilidade central do fluxo de pedido.
  - Trade-off: acoplamento de coordenacao em um servico.
- Topic exchanges no RabbitMQ:
  - Beneficio: flexibilidade de assinaturas futuras.
  - Trade-off: disciplina maior em naming de routing key.
- Banco por servico:
  - Beneficio: autonomia e escalabilidade.
  - Trade-off: consistencia eventual e duplicacao controlada de dados.

## 7. Estrategia de persistencia por dominio (decisao consolidada)
Depois da analise da arquitetura e dos requisitos de negocio, cheguei a conclusao de que o projeto deve adotar persistencia poliglota.

Distribuicao definida:
- `auth-service`: MongoDB
- `user-service`: MongoDB
- `restaurant-service`: MongoDB
- `delivery-service`: MongoDB
- `notification-service`: MongoDB
- `order-service`: PostgreSQL
- `payment-service`: PostgreSQL

Justificativa principal:
- `payment-service` exige consistencia transacional forte, auditoria e confiabilidade de escrita para operacoes financeiras.
- `order-service` possui transicoes criticas de estado do pedido e regras de integridade que se beneficiam de modelo relacional.
- Servicos orientados a dados mais documentais e flexiveis continuam em MongoDB para evolucao rapida.

Impacto no plano de execucao:
- Na implementacao de `order-service` e `payment-service`, incluir configuracao e camada de persistencia para PostgreSQL desde o inicio.
- Nos demais servicos, manter MongoDB conforme planejamento original.

