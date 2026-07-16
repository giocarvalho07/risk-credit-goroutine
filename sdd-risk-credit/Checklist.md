# Checklist - Sistema de Avaliação de Crédito

## Pré-Desenvolvimento

### Ambiente
- [ ] Go 1.21+ instalado
- [ ] Docker instalado
- [ ] Docker Compose instalado
- [ ] Editor de código configurado
- [ ] Git configurado

### Estrutura
- [x] Diretório do projeto criado
- [x] Estrutura de diretórios criada
- [x] `go mod init` executado
- [x] `go.mod` criado com dependências

## Infraestrutura

### Docker
- [x] `Dockerfile` criado
- [x] `docker-compose.yml` criado
- [x] Serviço App configurado
- [x] Variáveis de ambiente configuradas

### Banco de Dados
- [x] Script `init.sql` criado
- [x] Tabela `customers` criada
- [x] Dados de exemplo inseridos
- [x] Conexão PostgreSQL testada

## Código

### Configurações
- [x] `config.yaml` criado
- [x] `config/config.go` implementado
- [x] Leitura de configurações via Viper

### Modelos
- [x] `customer.go` implementado
- [x] `bureau.go` implementado
- [x] `credit_response.go` implementado

### Infraestrutura
- [x] `database/connection.go` implementado
- [x] Pool de conexões configurado
- [x] Tratamento de erros implementado

### Repositório
- [x] `customer_repository.go` implementado
- [x] Método `GetByCPF` implementado
- [x] Queries SQL funcionais

### Lógica de Negócio
- [x] `credit_ruler.go` implementado
- [x] Regra Score >= 700 (BAIXO/APROVAR)
- [x] Regra 500 <= Score < 700 (MEDIO/ANALISAR)
- [x] Regra Score < 500 (ALTO/NEGAR)
- [x] Regra Santander (ALTO/NEGAR)

### Service com Goroutines
- [x] `credit_service.go` implementado
- [x] `sync` importado
- [x] Variáveis compartilhadas declaradas
- [x] `sync.WaitGroup` declarado
- [x] `sync.Mutex` declarado
- [x] `wg.Add(2)` implementado
- [x] Goroutine 1 (PostgreSQL) implementada
- [x] Goroutine 2 (Bureau JSON) implementada
- [x] `defer wg.Done()` em cada goroutine
- [x] `mu.Lock()` / `mu.Unlock()` em cada goroutine
- [x] `wg.Wait()` implementado
- [x] Verificação de erros pós-sincronização
- [x] Método `loadBureauData` implementado
- [x] Método `getMensagem` implementado

### Handler
- [x] `credit_handler.go` implementado
- [x] Validação de CPF
- [x] Tratamento de erros
- [x] Respostas JSON padronizadas

### Entry Point
- [x] `cmd/api/main.go` implementado
- [x] Inicialização de dependências
- [x] Configuração de rotas
- [x] Endpoint POST /api/credit/analyze
- [x] Endpoint GET /health

## Dados

### Dados de Teste
- [x] `internal/data/bureau_data.json` criado
- [x] Dados de bureau para testes
- [x] `test.json` criado
- [x] Cenários de teste documentados

## Testes

### Funcionais
- [ ] Health check funciona
- [ ] Análise com score alto → APROVAR
- [ ] Análise com score médio → ANALISAR
- [ ] Análise com score baixo → NEGAR
- [ ] Análise com empréstimo Santander → NEGAR
- [ ] Cliente não encontrado → Erro 404
- [ ] CPF inválido → Erro 400
- [ ] Resposta no formato correto

### Goroutines
- [ ] Busca PostgreSQL e JSON ocorrem em paralelo
- [ ] WaitGroup sincroniza corretamente
- [ ] Mutex protege variáveis compartilhadas
- [ ] Erros são capturados corretamente
- [ ] **Race detector não reporta problemas** (`go run -race`)
- [ ] **Não há data race entre goroutines**

### Docker
- [ ] `docker-compose up --build` funciona
- [ ] App inicia corretamente
- [ ] Logs são exibidos corretamente

### Performance
- [ ] Tempo de resposta < 100ms
- [ ] Pool de conexões funciona
- [ ] Não há vazamento de conexões
- [ ] **Goroutines reduzem tempo de I/O**

## Validação Final

### Endpoints
- [ ] POST /api/credit/analyze aceita CPF
- [ ] POST /api/credit/analyze retorna dados corretos
- [ ] POST /api/credit/analyze trata erros
- [ ] GET /health retorna {"status": "ok"}

### Dados
- [ ] PostgreSQL contém 5 clientes
- [ ] Bureau JSON contém dados de teste
- [ ] Todos os CPFs são válidos

### Concorrência
- [ ] Goroutines funcionam corretamente
- [ ] Não há race conditions
- [ ] Sistema é thread-safe
- [ ] Performa melhor que versão sequencial

### Documentação
- [ ] README.md criado
- [ ] Instruções de execução documentadas
- [ ] Exemplos de uso documentados

## Entregáveis

- [ ] Código fonte em Go
- [ ] Dockerfile + Docker Compose
- [ ] Banco de dados PostgreSQL com dados de exemplo
- [ ] Arquivo JSON com dados do bureau
- [ ] Régua de risco implementada
- [ ] API REST documentada
- [ ] Arquivo de teste (.json)
- [ ] README com instruções
- [ ] Arquitetura Clean implementada
- [ ] **Goroutines para buscas paralelas implementadas**

## Status

| Categoria | Progresso |
|-----------|-----------|
| Pré-Desenvolvimento | 0% |
| Infraestrutura | 0% |
| Código | 0% |
| Goroutines | 0% |
| Dados | 0% |
| Testes | 0% |
| Validação Final | 0% |
| **Geral** | **0%** |
