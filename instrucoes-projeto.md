# Food Delivery Platform
## Sistema Distribuído de Delivery baseado em Microsserviços

---

# 1. Objetivo do Projeto

Desenvolver uma plataforma de delivery inspirada em sistemas como iFood/Uber Eats, utilizando arquitetura de microsserviços, comunicação assíncrona com RabbitMQ e banco de dados MongoDB.

O objetivo principal deste projeto é construir um sistema distribuído escalável, resiliente e desacoplado, capaz de processar pedidos, pagamentos, entregas e notificações em tempo real.

Além do produto final, este projeto tem como objetivo educacional proporcionar aprendizado prático em:

- Arquitetura de microsserviços
- Event-Driven Architecture
- RabbitMQ
- Saga Pattern
- Concorrência em Golang
- Comunicação síncrona e assíncrona
- Escalabilidade
- Resiliência
- Observabilidade
- Dockerização
- Design Patterns
- Boas práticas de backend moderno

---

# 2. Stack Tecnológica

## Backend
- Golang

## Banco de Dados
- MongoDB

## Mensageria
- RabbitMQ

## Containerização
- Docker
- Docker Compose

## API
- REST API
- gRPC (opcional)

## Observabilidade
- Prometheus
- Grafana
- Jaeger

## Documentação
- Swagger / OpenAPI

## Autenticação
- JWT
- Refresh Token

## Testes
- Testes Unitários
- Testes de Integração

## CI/CD
- GitHub Actions

---

# 3. Arquitetura do Sistema

O sistema será baseado em microsserviços independentes.

Cada serviço:
- possui responsabilidade única
- possui banco próprio
- pode ser escalado individualmente
- se comunica via HTTP/gRPC e RabbitMQ

---

# 4. Microsserviços

## 4.1 Auth Service

Responsável por:
- Registro de usuários
- Login
- Geração de JWT
- Refresh Token
- Controle de autenticação

### Banco
- MongoDB

### Endpoints
- POST /auth/register
- POST /auth/login
- POST /auth/refresh

---

## 4.2 User Service

Responsável por:
- Perfil do usuário
- Endereço
- Histórico de pedidos

### Endpoints
- GET /users/me
- PUT /users/me
- GET /users/me/orders

---

## 4.3 Restaurant Service

Responsável por:
- Cadastro de restaurantes
- Cardápios
- Categorias
- Disponibilidade de itens

### Endpoints
- GET /restaurants
- GET /restaurants/:id
- GET /restaurants/:id/menu

---

## 4.4 Order Service

Responsável por:
- Criação de pedidos
- Atualização de status
- Controle de pedidos

### Endpoints
- POST /orders
- GET /orders/:id
- PATCH /orders/:id/status

### Eventos Produzidos
- order.created
- order.cancelled
- order.confirmed

---

## 4.5 Payment Service

Responsável por:
- Processamento de pagamento
- Aprovação/Rejeição
- Simulação de gateway

### Eventos Consumidos
- order.created

### Eventos Produzidos
- payment.approved
- payment.failed

---

## 4.6 Delivery Service

Responsável por:
- Busca de entregador
- Rastreamento de entrega
- Atualização de status

### Eventos Consumidos
- payment.approved

### Eventos Produzidos
- delivery.started
- delivery.completed

---

## 4.7 Notification Service

Responsável por:
- Email
- SMS (mock)
- Push notifications

### Eventos Consumidos
- payment.approved
- payment.failed
- delivery.started
- delivery.completed

---

# 5. Comunicação Entre Serviços

## Comunicação Síncrona
Utilizada para:
- consultas rápidas
- validações
- autenticação

### Tecnologias
- REST
- gRPC

---

## Comunicação Assíncrona
Utilizada para:
- processamento de eventos
- desacoplamento
- workflows distribuídos

### Tecnologia
- RabbitMQ

---

# 6. RabbitMQ — Estrutura

## Exchanges

### order.exchange
Responsável por eventos de pedidos.

### payment.exchange
Responsável por eventos financeiros.

### delivery.exchange
Responsável por eventos de entrega.

### notification.exchange
Responsável por notificações.

---

# 7. Filas

## Principais Filas

- order.created.queue
- payment.processing.queue
- delivery.processing.queue
- notification.queue

---

# 8. Dead Letter Queue (DLQ)

Todas as filas críticas devem possuir DLQ.

## Exemplo
- payment.processing.dlq

Objetivo:
- armazenar mensagens inválidas
- facilitar debug
- evitar perda de eventos

---

# 9. Retry Strategy

Mensagens falhadas devem:
- ser reenviadas automaticamente
- utilizar retry exponencial
- possuir limite máximo de tentativas

---

# 10. Saga Pattern

## Objetivo

Garantir consistência distribuída entre microsserviços sem utilizar transações distribuídas.

---

## Exemplo do Fluxo

### 1. Pedido criado
Order Service publica:
- order.created

### 2. Pagamento processado
Payment Service:
- aprova pagamento
- publica payment.approved

OU

- rejeita pagamento
- publica payment.failed

### 3. Entrega iniciada
Delivery Service consome:
- payment.approved

### 4. Caso ocorra falha
Order Service executa ação compensatória:
- cancela pedido

---

## Compensações

### Exemplo
Se pagamento falhar:
- pedido deve ser cancelado automaticamente

Se entrega falhar:
- pagamento pode ser estornado

---

# 11. Design Patterns Utilizados

## Clean Architecture
Separação entre:
- domain
- application
- infrastructure
- delivery

---

## Repository Pattern
Abstração de acesso ao banco.

---

## Dependency Injection
Desacoplamento entre componentes.

---

## Event-Driven Architecture
Comunicação baseada em eventos.

---

## Worker Pool
Processamento concorrente de jobs.

---

## Circuit Breaker (opcional)
Proteção contra falhas em cascata.

---

## Outbox Pattern (opcional)
Garantir consistência entre banco e eventos.

---

# 12. Requisitos Funcionais

## RF001
O sistema deve permitir cadastro de usuários.

## RF002
O sistema deve permitir autenticação via JWT.

## RF003
O usuário deve conseguir visualizar restaurantes.

## RF004
O usuário deve conseguir visualizar cardápios.

## RF005
O usuário deve conseguir criar pedidos.

## RF006
O sistema deve processar pagamentos.

## RF007
O sistema deve atualizar status do pedido.

## RF008
O sistema deve iniciar entrega após pagamento aprovado.

## RF009
O sistema deve notificar usuários sobre alterações no pedido.

## RF010
O sistema deve manter histórico de pedidos.

---

# 13. Requisitos Não Funcionais

## RNF001
O sistema deve ser escalável horizontalmente.

## RNF002
O sistema deve tolerar falhas temporárias.

## RNF003
O sistema deve possuir observabilidade.

## RNF004
O sistema deve possuir logs estruturados.

## RNF005
O sistema deve suportar comunicação assíncrona.

## RNF006
O sistema deve ser containerizado.

## RNF007
O sistema deve possuir documentação de APIs.

## RNF008
O sistema deve possuir testes automatizados.

## RNF009
O sistema deve garantir rastreabilidade de requests.

---

# 14. Regras de Negócio

## RN001
Pedidos só podem ser enviados para entrega após pagamento aprovado.

## RN002
Pedidos cancelados não podem ser reativados.

## RN003
Itens indisponíveis não podem ser adicionados ao pedido.

## RN004
Pedidos pagos devem possuir registro financeiro.

## RN005
O usuário deve receber notificação em mudanças de status.

## RN006
Pedidos entregues não podem ser alterados.

## RN007
O sistema deve impedir duplicação de pagamentos.

---

# 15. Observabilidade

## Logs
Todos os serviços devem possuir:
- logs estruturados
- correlation id
- request id

---

## Métricas
Monitoramento de:
- tempo de resposta
- filas
- erros
- throughput

---

## Tracing Distribuído
Rastrear:
- request
- eventos
- comunicação entre serviços

---

# 16. Segurança

## Requisitos
- JWT
- Refresh Token
- Senhas criptografadas
- Rate Limiting
- Middleware de autenticação

---

# 17. Estrutura de Pastas

```txt
/services
  /auth-service
  /user-service
  /restaurant-service
  /order-service
  /payment-service
  /delivery-service
  /notification-service

/shared
  /broker
  /logger
  /events
  /middlewares
  /utils

/docker

/scripts

README.md
docker-compose.yml
Makefile