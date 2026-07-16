# Task - Sistema de Avaliação de Crédito

## Tarefas

### 1. Estrutura do Projeto
- [x] Criar diretório raiz `credit-risk-analyzer`
- [x] Criar estrutura de diretórios:
  ```
  credit-risk-analyzer/
  ├── cmd/api/
  ├── internal/config/
  ├── internal/database/
  ├── internal/model/
  ├── internal/repository/
  ├── internal/service/
  ├── internal/handler/
  ├── internal/data/
  └── pkg/utils/
  ```
- [x] Executar `go mod init github.com/seu-usuario/credit-risk-analyzer`

### 2. Configuração
- [x] Criar arquivo `config.yaml` com configurações de servidor e banco
- [x] Criar arquivo `go.mod` com dependências:
  - gin v1.9.1
  - lib/pq v1.10.9
  - viper v1.18.2
  - logrus v1.9.3

### 3. Docker
- [x] Criar `Dockerfile` com build multi-stage
- [x] Criar `docker-compose.yml` com serviço app
- [x] Configurar variáveis de ambiente

### 4. Banco de Dados
- [x] Criar script `init.sql` com:
  - Criação da tabela `customers`
  - Inserção de dados de exemplo (5 clientes)

### 5. Modelos (internal/model/)
- [x] Criar `customer.go`:
  ```go
  type Customer struct {
      ID        int     `json:"id"`
      Nome      string  `json:"nome"`
      CPF       string  `json:"cpf"`
      Salario   float64 `json:"salario"`
      Profissao string  `json:"profissao"`
  }
  ```
- [x] Criar `bureau.go`:
  ```go
  type BureauData struct {
      ID                  int    `json:"id"`
      CPF                 string `json:"cpf"`
      ScoreCredito        int    `json:"score_credito"`
      EmprestimoExistente bool   `json:"emprestimo_existente"`
      TipoEmprestimo      string `json:"tipo_emprestimo"`
      Instituicao         string `json:"instituicao"`
  }

  type BureauResponse struct {
      Bureau []BureauData `json:"bureau"`
  }
  ```
- [x] Criar `credit_response.go`:
  ```go
  type CreditResponse struct {
      Nome       string `json:"nome"`
      CPF        string `json:"cpf"`
      Score      int    `json:"score_credito"`
      NivelRisco string `json:"nivel_risco"`
      Acao       string `json:"acao"`
      Descricao  string `json:"descricao"`
      Mensagem   string `json:"mensagem"`
  }
  ```

### 6. Infraestrutura (internal/database/)
- [x] Criar `connection.go` com:
  - Struct `DBConfig`
  - Função `NewPostgresConnection`
  - Função `InitDatabase`
  - Configuração de pool de conexões
  - Tratamento de erros

### 7. Configurações (internal/config/)
- [x] Criar `config.go` com leitura via Viper

### 8. Repositório (internal/repository/)
- [x] Criar `customer_repository.go`:
  - Struct `CustomerRepository`
  - Função `NewCustomerRepository`
  - Método `GetByCPF(cpf string) (*model.Customer, error)`

### 9. Lógica de Negócio (internal/service/)
- [x] Criar `credit_ruler.go`:
  - Struct `CreditRuler`
  - Struct `RiskResult`
  - Método `Evaluate(bureau model.BureauData) RiskResult`
  - Implementar regras:
    - Score >= 700 → BAIXO/APROVAR
    - 500 <= Score < 700 → MEDIO/ANALISAR
    - Score < 500 → ALTO/NEGAR
    - Santander → ALTO/NEGAR

### 10. Service com Goroutines (internal/service/)
- [x] Criar `credit_service.go`:
  - Struct `CreditService`
  - Função `NewCreditService`
  - Método `Analyze(cpf string) (*model.CreditResponse, error)`
  - Método `loadBureauData(cpf string) (*model.BureauData, error)`
  - Método `getMensagem(acao string) string`

### 11. Implementação de Goroutines
- [x] Importar `sync` no credit_service.go
- [x] Declarar variáveis compartilhadas com Mutex:
  ```go
  var (
      wg        sync.WaitGroup
      mu        sync.Mutex
      customer  *model.Customer
      bureau    *model.BureauData
      errDB     error
      errBureau error
  )
  ```
- [x] Implementar `wg.Add(2)` antes das goroutines
- [x] Criar goroutine 1 - Busca PostgreSQL:
  ```go
  go func() {
      defer wg.Done()
      c, err := s.repo.GetByCPF(cpf)
      mu.Lock()
      defer mu.Unlock()
      customer = c
      errDB = err
  }()
  ```
- [x] Criar goroutine 2 - Busca Bureau JSON:
  ```go
  go func() {
      defer wg.Done()
      b, err := s.loadBureauData(cpf)
      mu.Lock()
      defer mu.Unlock()
      bureau = b
      errBureau = err
  }()
  ```
- [x] Implementar `wg.Wait()` após goroutines
- [x] Verificar erros após sincronização

### 12. Handler (internal/handler/)
- [x] Criar `credit_handler.go`:
  - Struct `CreditHandler`
  - Função `NewCreditHandler`
  - Struct `CreditRequest`
  - Método `Analyze(c *gin.Context)`
  - Validação de entrada
  - Tratamento de erros

### 13. Entry Point (cmd/api/)
- [x] Criar `main.go`:
  - Configuração Viper
  - Inicialização PostgreSQL
  - Injeção de dependências
  - Configuração de rotas Gin
  - Endpoint POST /api/credit/analyze
  - Endpoint GET /health

### 14. Dados
- [x] Criar `internal/data/bureau_data.json` com dados de teste
- [x] Criar `test.json` com cenários de teste

### 15. Testes
- [ ] Testar containerização: `docker-compose up --build -d`
- [ ] Testar health check: `curl http://localhost:8080/health`
- [ ] Testar análise com score alto:
  ```bash
  curl -X POST http://localhost:8080/api/credit/analyze \
    -H "Content-Type: application/json" \
    -d '{"cpf": "123.456.789-00"}'
  ```
- [ ] Testar análise com score médio
- [ ] Testar análise com score baixo
- [ ] Testar cliente não encontrado
- [ ] Testar CPF inválido
- [ ] **Testar com race detector**: `go run -race cmd/api/main.go`
- [ ] **Verificar que não há race conditions**

### 16. Documentação
- [ ] Criar README.md com instruções

## Resumo de Tarefas

| Categoria | Total | Concluídas |
|-----------|-------|------------|
| Estrutura | 3 | 3 |
| Configuração | 2 | 2 |
| Docker | 3 | 3 |
| Banco de Dados | 1 | 1 |
| Modelos | 3 | 3 |
| Infraestrutura | 1 | 1 |
| Configurações | 1 | 1 |
| Repositório | 1 | 1 |
| Lógica | 1 | 1 |
| **Goroutines** | **8** | **8** |
| Handler | 1 | 1 |
| Main | 1 | 1 |
| Dados | 2 | 2 |
| Testes | 9 | 0 |
| Documentação | 1 | 0 |
| **Total** | **39** | **29** |
