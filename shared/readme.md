# Shared Foundation

Pacote compartilhado para infraestrutura transversal entre microsservicos:
- contracts: contratos técnicos
- events: envelope e convencoes de evento
- logger: logging estruturado
- middleware: middlewares HTTP comuns
- broker: camada comum RabbitMQ

Regras:
1. Nao colocar regra de negocio aqui.
2. Nao colocar entidade de dominio de servico aqui.
3. Apenas infraestrutura e contratos técnicos reutilizaveis.