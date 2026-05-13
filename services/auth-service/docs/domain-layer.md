# Auth Service - Documentacao da Camada Domain

## Objetivo deste documento
Explicar de forma detalhada a camada de dominio do auth-service: o que existe, porque existe, como as pastas se relacionam e qual problema arquitetural cada decisao resolve.

## Contexto
A camada de dominio representa o nucleo de regra de negocio do servico de autenticacao. Ela deve permanecer independente de tecnologia especifica como MongoDB, HTTP, RabbitMQ, bibliotecas de JWT e frameworks.

Essa separacao segue Clean Architecture.

## Principios que guiam esta camada
- Regra de negocio primeiro.
- Baixo acoplamento com infraestrutura.
- Alta testabilidade.
- Linguagem de negocio explicita no codigo.
- Evolucao segura sem quebrar o nucleo da aplicacao.

## Estrutura atual
```txt
internal/domain/
  entities/
    credential.go
    refresh_session.go
  valueobjects/
    email.go
    credential_status.go
    password_policy.go
    token_ttl.go
    token_claims.go
    auth_tokens.go
  errors/
    domain_errors.go
  ports/
    repositories.go
    security.go
    system.go
```

## Visao geral das pastas

### entities
Contem as entidades de dominio. Entidade e um objeto com identidade propria e ciclo de vida.

No auth-service:
- Credential: credencial de login do usuario.
- RefreshSession: sessao de refresh token com revogacao e rotacao.

Por que entidade:
- Possui identidade (`ID`, `UserID`).
- Muda de estado ao longo do tempo.
- Carrega regras de transicao de estado.

### valueobjects
Contem objetos de valor. Value Object nao tem identidade propria; ele representa um conceito validado e imutavel por contrato semantico.

No auth-service:
- Email: normaliza e valida formato de email.
- CredentialStatus: enum de estado da credencial.
- PasswordPolicy: regras de senha forte.
- TokenTTL: regras de duracao de tokens.
- TokenClaims: claims de autenticacao.
- AuthTokens: pacote de retorno de tokens.

Por que value object:
- Evita strings soltas no sistema.
- Centraliza validacao e semantica.
- Reduz bug de regra duplicada em varios pontos.

### errors
Contem erros de dominio. Sao erros semanticos de regra de negocio, nao erros tecnicos de IO.

No auth-service:
- Exemplo: `ErrInvalidEmail`, `ErrCredentialDisabled`, `ErrRefreshTokenExpired`.

Por que erros de dominio:
- Padronizam o significado das falhas.
- Facilitam mapeamento para HTTP depois.
- Evitam acoplamento da camada de dominio com codigos HTTP.

### ports
Contem interfaces (contratos) que o dominio/use cases precisam para operar sem conhecer implementacoes concretas.

No auth-service:
- repositories.go: contratos de persistencia.
- security.go: contratos de hash e token.
- system.go: contratos de tempo e geracao de IDs.

Por que ports:
- Dominio nao depende de Mongo/JWT diretamente.
- Testes unitarios podem usar doubles simples.
- Infraestrutura pode trocar sem quebrar regra de negocio.

## Entidades detalhadas

### Credential
Responsavel por representar autenticacao principal do usuario.

Campos principais:
- `ID`, `UserID`: identidade.
- `Email`: VO validado.
- `PasswordHash`: hash da senha.
- `Status`: estado da credencial.
- `FailedLoginAttempts`: tentativas falhas.
- `LastLoginAt`, `CreatedAt`, `UpdatedAt`: rastreabilidade temporal.

Comportamentos:
- `CanLogin()`: decide se pode autenticar.
- `RegisterFailedAttempt(now)`: incrementa falha.
- `RegisterSuccessfulLogin(now)`: limpa falhas e registra login.
- `ChangePassword(newHash, now)`: atualiza hash.
- `Disable(now)`: desativa credencial.

Decisao arquitetural:
- A entidade guarda regras de transicao de estado para evitar regra espalhada em handlers/use cases.

### RefreshSession
Responsavel por controlar sessao de refresh token.

Campos principais:
- `TokenHash`: nunca guardar refresh token em claro.
- `ExpiresAt`: expiracao.
- `RevokedAt`: revogacao explicita.
- `RotatedFromSessionID`: encadeamento de rotacao.

Comportamentos:
- `IsExpired(now)`
- `IsRevoked()`
- `CanBeUsed(now)`
- `Revoke(now)`
- `Rotate(...)`: revoga sessao atual e cria nova sessao.

Decisao arquitetural:
- Colocar a logica de revogacao e rotacao no dominio garante consistencia e diminui risco de erro de seguranca.

## Value Objects detalhados

### Email
- Entrada normalizada (`trim + lowercase`).
- Formato validado por regex.

Beneficio:
- Unicidade de comportamento para email em todo o sistema.

### CredentialStatus
- Estado explicito da credencial.
- Valor valido apenas dentro de conjunto permitido.

Beneficio:
- Evita status invalido em runtime.

### PasswordPolicy
- Define parametros de senha forte (tamanho, upper, lower, numero, especial).

Beneficio:
- Politica de senha centralizada e facilmente testavel.

### TokenTTL
- Garante que `refresh` seja maior que `access`.

Beneficio:
- Evita configuracao insegura de expiracao.

### TokenClaims
- Estrutura padrao de claims e regra de expiracao.

Beneficio:
- Uniformidade na validacao de token.

### AuthTokens
- Estrutura de transporte para resposta de autenticacao.

Beneficio:
- Contrato claro entre use case e camada de entrega.

## Ports detalhados

### repositories.go
Contratos da persistencia necessaria ao dominio de auth.

- `CredentialRepository`
- `RefreshTokenRepository`

Motivo:
- Use case nao conhece MongoDB.
- Repositorios concretos ficam na camada infrastructure.

### security.go
Contratos de seguranca.

- `PasswordHasher`
- `TokenService`

Motivo:
- Dominio nao fica preso a bcrypt/jwt library especifica.

### system.go
Contratos transversais para testabilidade.

- `Clock`
- `IDGenerator`

Motivo:
- Testes deterministas sem depender de tempo real ou UUID real.

## Fluxo de dependencia correto
Regra geral:
- `domain` nao importa `infrastructure`.
- `application` depende de `domain`.
- `infrastructure` implementa os `ports` de `domain`.
- `delivery` chama `application`.

Isso reduz acoplamento e facilita manutencao.

## Por que essa modelagem e importante para microsservicos
- Cada servico precisa evoluir de forma independente.
- A camada de dominio tende a ser a parte mais estavel e valiosa.
- Se dominio estiver acoplado a tecnologia, qualquer troca tecnica vira refatoracao cara.

## Trade-offs
- Mais arquivos no inicio.
- Curva de aprendizado maior.
- Ganho de medio/longo prazo em qualidade, testes e evolucao.

## Como essa camada se conecta com as proximas fases
- Fase 2 (application): use cases vao consumir entidades, VOs, erros e ports.
- Fase 3 (security infra): implementa `PasswordHasher` e `TokenService`.
- Fase 4 (mongo infra): implementa repositorios.
- Fase 5 (http delivery): mapeia erros de dominio para respostas HTTP.

## Boas praticas de manutencao
- Toda nova regra de negocio entra primeiro no dominio.
- Evitar tipo primitivo quando houver conceito de negocio (preferir VO).
- Nao importar pacote de infraestrutura em `internal/domain`.
- Criar testes de dominio antes de integrar com banco/HTTP.

## Checklist de saude da camada domain
- Entidades sem dependencia de HTTP, Mongo, Rabbit.
- Value Objects com validacao explicita.
- Ports cobrindo tudo que dominio precisa externamente.
- Erros de dominio com semantica clara.
- Testes unitarios cobrindo regras e transicoes.

## Conclusao
A camada domain do auth-service foi estruturada para preservar regras de autenticacao no centro do sistema, desacopladas de tecnologia. Isso permite evoluir infraestrutura sem reescrever regra critica, acelerar testes e manter consistencia em cenarios de seguranca como login, revogacao e rotacao de refresh token.
