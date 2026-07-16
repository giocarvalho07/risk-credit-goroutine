PRD - Sistema de Avaliação de Crédito
1. VISÃO GERAL DO PROJETO
1.1 Objetivo
Sistema simplificado de avaliação de crédito que consolida dados do cliente (PostgreSQL), informações de bureau (JSON) e aplica uma régua de risco para tomar decisões de crédito, utilizando goroutines para buscas paralelas.

1.2 Escopo
- Buscar dados do cliente no PostgreSQL
- Carregar dados do bureau via arquivo JSON
- Buscar dados do cliente e do bureau **em paralelo** com goroutines
- Aplicar regras de risco
- Retornar decisão (APROVAR/NEGAR/ANALISAR)
- API REST com um único endpoint
- Containerização com Docker

1.3 Tecnologias
- Linguagem: Go 1.21+
- Framework: Gin (API REST)
- Banco de Dados: PostgreSQL (Render)
- Container: Docker + Docker Compose
- Arquitetura: Clean Architecture (simplificada)
- Concorrência: Goroutines + sync.WaitGroup + sync.Mutex

2. ARQUITETURA CLEAN (CAMADAS)
```
┌─────────────────────────────────────────────┐
│           CAMADA DE APRESENTAÇÃO            │
│           (handlers/credit_handler.go)      │
│         - Recebe requisições HTTP           │
│         - Valida entrada                    │
│         - Retorna respostas JSON            │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│           CAMADA DE CASO DE USO             │
│           (services/credit_service.go)      │
│         - Orquestra o fluxo de negócio      │
│         - Busca dados do cliente (goroutine)│
│         - Carrega dados do bureau (goroutine)│
│         - Aplica régua de risco             │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│           CAMADA DE REPOSITÓRIO             │
│        (repository/customer_repository.go)  │
│         - Interface com banco de dados      │
│         - Queries SQL                       │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│           CAMADA DE INFRAESTRUTURA          │
│        (database/connection.go)             │
│         - Conexão com PostgreSQL            │
│         - Pool de conexões                  │
└─────────────────────────────────────────────┘
```

2.1 Fluxo de Concorrência (Goroutines)
```
CreditService.Analyze(cpf)
│
├─── wg.Add(2)
│
├─── goroutine 1 ──► repo.GetByCPF(cpf)  ──► PostgreSQL
│        │                                       │
│        └──► mu.Lock() → customer = c           │
│                                       │
├─── goroutine 2 ──► loadBureauData(cpf) ──► JSON│
│        │                                       │
│        └──► mu.Lock() → bureau = b             │
│                                       │
└─── wg.Wait() ◄────────────────────────────────┘
         │
         ├── Verifica erros (PostgreSQL e Bureau)
         ├── ruler.Evaluate(bureau)
         └── Monta e retorna CreditResponse
```

3. REQUISITOS FUNCIONAIS
3.1 Banco de Dados - PostgreSQL
Tabela: customers

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| id | SERIAL | Sim | Chave primária auto-increment |
| nome | VARCHAR(100) | Sim | Nome completo do cliente |
| cpf | VARCHAR(14) | Sim | CPF formatado (único) |
| salario | DECIMAL(10,2) | Sim | Salário mensal em R$ |
| profissao | VARCHAR(50) | Não | Profissão do cliente |

SQL de Criação:

```sql
CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(100) NOT NULL,
    cpf VARCHAR(14) UNIQUE NOT NULL,
    salario DECIMAL(10,2) DEFAULT 0.00,
    profissao VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

3.2 Arquivo JSON - Bureau
Arquivo: internal/data/bureau_data.json

| Campo | Tipo | Descrição |
|-------|------|-----------|
| id | INT | Identificador único |
| cpf | STRING | CPF do cliente |
| score_credito | INT | Score de crédito (0-1000) |
| emprestimo_existente | BOOLEAN | Possui empréstimo ativo? |
| tipo_emprestimo | STRING | Tipo do empréstimo |
| instituicao | STRING | Instituição financeira |

Exemplo do JSON:

```json
{
  "bureau": [
    {
      "id": 1,
      "cpf": "123.456.789-00",
      "score_credito": 750,
      "emprestimo_existente": true,
      "tipo_emprestimo": "Imobiliário",
      "instituicao": "Banco do Brasil"
    }
  ]
}
```

3.3 Régua de Risco - Módulo Go

| Nível de Risco | Condição | Ação | Descrição |
|----------------|----------|------|-----------|
| BAIXO | Score >= 700 | APROVAR | Cliente com excelente histórico |
| MÉDIO | Score >= 500 e < 700 | ANALISAR | Análise manual necessária |
| ALTO | Score < 500 | NEGAR | Alto risco de inadimplência |
| ALTO | Empréstimo no Santander | NEGAR | Instituição de alto risco |

3.4 Endpoint da API
POST /api/credit/analyze

Request Body:

```json
{
  "cpf": "123.456.789-00"
}
```

Response Body:

```json
{
  "nome": "João Silva",
  "cpf": "123.456.789-00",
  "score_credito": 750,
  "nivel_risco": "BAIXO",
  "acao": "APROVAR",
  "descricao": "Score elevado, baixo risco",
  "mensagem": "Crédito aprovado!"
}
```

4. ESTRUTURA DE DIRETÓRIOS
```
credit-risk-analyzer/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point da aplicação
├── internal/
│   ├── config/
│   │   └── config.go               # Configurações (Viper)
│   ├── database/
│   │   └── connection.go           # Conexão PostgreSQL
│   ├── model/
│   │   ├── customer.go             # Modelo do cliente
│   │   ├── bureau.go               # Modelo do bureau
│   │   └── credit_response.go      # Modelo de resposta
│   ├── repository/
│   │   └── customer_repository.go  # Acesso ao banco
│   ├── service/
│   │   ├── credit_service.go       # Caso de uso (com goroutines)
│   │   └── credit_ruler.go         # Régua de risco
│   ├── handler/
│   │   └── credit_handler.go       # API Handler
│   └── data/
│       └── bureau_data.json        # Dados do bureau
├── pkg/
│   └── utils/
│       └── logger.go               # Utilitário de logs
├── docker-compose.yml              # Orquestração Docker
├── Dockerfile                      # Imagem da aplicação
├── init.sql                        # Script de inicialização
├── config.yaml                     # Configurações
├── go.mod                          # Dependências
├── test.json                       # JSON para testes
└── README.md                       # Documentação
```

5. IMPLEMENTAÇÃO PASSO A PASSO

Passo 1: Estrutura do Projeto
```bash
mkdir credit-risk-analyzer
cd credit-risk-analyzer
go mod init github.com/seu-usuario/credit-risk-analyzer
```

Passo 2: Configurações (config.yaml)
```yaml
server:
  port: 8080

database:
  host: dpg-d9c3bvhkh4rs73bvdjog-a.oregon-postgres.render.com
  port: 5432
  user: root
  password: 6oz39OKVaYhgy6n7i5pni0rz5BRL0SuX
  name: db_credit
  sslmode: require
```

Passo 3: Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY config.yaml .
COPY internal/data/ ./internal/data/
EXPOSE 8080
CMD ["./main"]
```

Passo 4: Docker Compose
```yaml
version: '3.8'

services:
  app:
    build: .
    container_name: credit_api
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=dpg-d9c3bvhkh4rs73bvdjog-a.oregon-postgres.render.com
      - DB_PORT=5432
      - DB_USER=root
      - DB_PASSWORD=6oz39OKVaYhgy6n7i5pni0rz5BRL0SuX
      - DB_NAME=db_credit
      - DB_SSLMODE=require
```

Passo 5: Script SQL (init.sql)
```sql
CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(100) NOT NULL,
    cpf VARCHAR(14) UNIQUE NOT NULL,
    salario DECIMAL(10,2) DEFAULT 0.00,
    profissao VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO customers (nome, cpf, salario, profissao) VALUES
('João Silva', '123.456.789-00', 5000.00, 'Engenheiro'),
('Maria Santos', '987.654.321-00', 8000.00, 'Médica'),
('Pedro Costa', '111.222.333-44', 3500.00, 'Professor'),
('Ana Lima', '555.666.777-88', 12000.00, 'Advogada'),
('Carlos Souza', '999.888.777-66', 4500.00, 'Programador')
ON CONFLICT (cpf) DO NOTHING;
```

Passo 6: Código Go - Camadas

6.1 Models (internal/model/)

```go
// customer.go
package model

type Customer struct {
    ID        int     `json:"id"`
    Nome      string  `json:"nome"`
    CPF       string  `json:"cpf"`
    Salario   float64 `json:"salario"`
    Profissao string  `json:"profissao"`
}
```

```go
// bureau.go
package model

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

```go
// credit_response.go
package model

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

6.2 Database Connection (internal/database/connection.go)

```go
package database

import (
    "database/sql"
    "fmt"
    "time"
    _ "github.com/lib/pq"
)

type DBConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    Database string
    SSLMode  string
}

func NewPostgresConnection(config DBConfig) (*sql.DB, error) {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(time.Hour)

    if err := db.Ping(); err != nil {
        return nil, err
    }

    return db, nil
}
```

6.3 Repository (internal/repository/customer_repository.go)

```go
package repository

import (
    "database/sql"
    "fmt"
    "github.com/seu-usuario/credit-risk-analyzer/internal/model"
)

type CustomerRepository struct {
    db *sql.DB
}

func NewCustomerRepository(db *sql.DB) *CustomerRepository {
    return &CustomerRepository{db: db}
}

func (r *CustomerRepository) GetByCPF(cpf string) (*model.Customer, error) {
    query := `SELECT id, nome, cpf, salario, profissao FROM customers WHERE cpf = $1`

    var customer model.Customer
    err := r.db.QueryRow(query, cpf).Scan(
        &customer.ID,
        &customer.Nome,
        &customer.CPF,
        &customer.Salario,
        &customer.Profissao,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("cliente não encontrado: %s", cpf)
    }

    if err != nil {
        return nil, err
    }

    return &customer, nil
}
```

6.4 Service - Régua de Risco (internal/service/credit_ruler.go)

```go
package service

import "github.com/seu-usuario/credit-risk-analyzer/internal/model"

type CreditRuler struct{}

func NewCreditRuler() *CreditRuler {
    return &CreditRuler{}
}

type RiskResult struct {
    NivelRisco string
    Acao       string
    Descricao  string
}

func (r *CreditRuler) Evaluate(bureau model.BureauData) RiskResult {
    if bureau.ScoreCredito >= 700 {
        return RiskResult{
            NivelRisco: "BAIXO",
            Acao:       "APROVAR",
            Descricao:  "Score elevado, baixo risco",
        }
    }

    if bureau.ScoreCredito >= 500 && bureau.ScoreCredito < 700 {
        return RiskResult{
            NivelRisco: "MEDIO",
            Acao:       "ANALISAR",
            Descricao:  "Score médio, análise manual necessária",
        }
    }

    if bureau.ScoreCredito < 500 {
        return RiskResult{
            NivelRisco: "ALTO",
            Acao:       "NEGAR",
            Descricao:  "Score baixo, alto risco",
        }
    }

    if bureau.EmprestimoExistente && bureau.Instituicao == "Santander" {
        return RiskResult{
            NivelRisco: "ALTO",
            Acao:       "NEGAR",
            Descricao:  "Cliente com empréstimo no Santander - Alto risco",
        }
    }

    return RiskResult{
        NivelRisco: "MEDIO",
        Acao:       "ANALISAR",
        Descricao:  "Análise manual recomendada",
    }
}
```

6.5 Service - Caso de Uso com Goroutines (internal/service/credit_service.go)

```go
package service

import (
    "encoding/json"
    "fmt"
    "os"
    "sync"
    "github.com/seu-usuario/credit-risk-analyzer/internal/model"
    "github.com/seu-usuario/credit-risk-analyzer/internal/repository"
)

type CreditService struct {
    repo  *repository.CustomerRepository
    ruler *CreditRuler
}

func NewCreditService(repo *repository.CustomerRepository, ruler *CreditRuler) *CreditService {
    return &CreditService{
        repo:  repo,
        ruler: ruler,
    }
}

func (s *CreditService) Analyze(cpf string) (*model.CreditResponse, error) {
    var (
        wg        sync.WaitGroup
        mu        sync.Mutex
        customer  *model.Customer
        bureau    *model.BureauData
        errDB     error
        errBureau error
    )

    wg.Add(2)

    // Goroutine 1: Busca cliente no PostgreSQL
    go func() {
        defer wg.Done()
        c, err := s.repo.GetByCPF(cpf)
        mu.Lock()
        defer mu.Unlock()
        customer = c
        errDB = err
    }()

    // Goroutine 2: Carrega dados do bureau (JSON)
    go func() {
        defer wg.Done()
        b, err := s.loadBureauData(cpf)
        mu.Lock()
        defer mu.Unlock()
        bureau = b
        errBureau = err
    }()

    // Aguarda ambas goroutines finalizarem
    wg.Wait()

    // Verifica erros
    if errDB != nil {
        return nil, errDB
    }
    if errBureau != nil {
        return nil, errBureau
    }

    // Aplica régua de risco
    riskResult := s.ruler.Evaluate(*bureau)

    // Monta resposta
    response := &model.CreditResponse{
        Nome:       customer.Nome,
        CPF:        customer.CPF,
        Score:      bureau.ScoreCredito,
        NivelRisco: riskResult.NivelRisco,
        Acao:       riskResult.Acao,
        Descricao:  riskResult.Descricao,
        Mensagem:   s.getMensagem(riskResult.Acao),
    }

    return response, nil
}

func (s *CreditService) loadBureauData(cpf string) (*model.BureauData, error) {
    data, err := os.ReadFile("internal/data/bureau_data.json")
    if err != nil {
        return nil, fmt.Errorf("erro ao ler arquivo bureau: %w", err)
    }

    var bureauResp model.BureauResponse
    if err := json.Unmarshal(data, &bureauResp); err != nil {
        return nil, fmt.Errorf("erro ao parsear bureau: %w", err)
    }

    for _, b := range bureauResp.Bureau {
        if b.CPF == cpf {
            return &b, nil
        }
    }

    return nil, fmt.Errorf("dados do bureau não encontrados para CPF: %s", cpf)
}

func (s *CreditService) getMensagem(acao string) string {
    mensagens := map[string]string{
        "APROVAR":  "Crédito aprovado!",
        "NEGAR":    "Crédito negado devido ao alto risco.",
        "ANALISAR": "Em análise manual. Retornaremos em breve.",
    }
    return mensagens[acao]
}
```

6.6 Handler (internal/handler/credit_handler.go)

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/seu-usuario/credit-risk-analyzer/internal/service"
)

type CreditHandler struct {
    service *service.CreditService
}

func NewCreditHandler(service *service.CreditService) *CreditHandler {
    return &CreditHandler{service: service}
}

type CreditRequest struct {
    CPF string `json:"cpf" binding:"required"`
}

func (h *CreditHandler) Analyze(c *gin.Context) {
    var req CreditRequest

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "erro": "CPF é obrigatório",
        })
        return
    }

    response, err := h.service.Analyze(req.CPF)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "erro": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, response)
}
```

6.7 Main (cmd/api/main.go)

```go
package main

import (
    "log"
    "github.com/gin-gonic/gin"
    "github.com/spf13/viper"
    "github.com/seu-usuario/credit-risk-analyzer/internal/database"
    "github.com/seu-usuario/credit-risk-analyzer/internal/handler"
    "github.com/seu-usuario/credit-risk-analyzer/internal/repository"
    "github.com/seu-usuario/credit-risk-analyzer/internal/service"
)

func main() {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.ReadInConfig()

    dbConfig := database.DBConfig{
        Host:     viper.GetString("database.host"),
        Port:     viper.GetInt("database.port"),
        User:     viper.GetString("database.user"),
        Password: viper.GetString("database.password"),
        Database: viper.GetString("database.name"),
        SSLMode:  viper.GetString("database.sslmode"),
    }

    db, err := database.NewPostgresConnection(dbConfig)
    if err != nil {
        log.Fatalf("Erro ao conectar ao PostgreSQL: %v", err)
    }
    defer db.Close()

    repo := repository.NewCustomerRepository(db)
    ruler := service.NewCreditRuler()
    creditService := service.NewCreditService(repo, ruler)
    creditHandler := handler.NewCreditHandler(creditService)

    router := gin.Default()
    router.POST("/api/credit/analyze", creditHandler.Analyze)
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    port := viper.GetString("server.port")
    router.Run(":" + port)
}
```

7. MÓDULOS GO UTILIZADOS
```
module github.com/seu-usuario/credit-risk-analyzer

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1      // Framework HTTP
    github.com/lib/pq v1.10.9             // Driver PostgreSQL
    github.com/spf13/viper v1.18.2       // Configurações
    github.com/sirupsen/logrus v1.9.3    // Logs
)
```

8. TESTES

Arquivo de Teste (test.json)
```json
{
  "tests": [
    {
      "descricao": "Cliente com score alto - Aprovar",
      "cpf": "123.456.789-00",
      "esperado": "APROVAR"
    },
    {
      "descricao": "Cliente com score médio - Analisar",
      "cpf": "999.888.777-66",
      "esperado": "ANALISAR"
    },
    {
      "descricao": "Cliente com score baixo - Negar",
      "cpf": "111.222.333-44",
      "esperado": "NEGAR"
    },
    {
      "descricao": "Cliente não encontrado",
      "cpf": "000.000.000-00",
      "esperado": "ERRO"
    }
  ]
}
```

Como Testar
```bash
# 1. Iniciar containers
docker-compose up -d

# 2. Testar endpoint
curl -X POST http://localhost:8080/api/credit/analyze \
  -H "Content-Type: application/json" \
  -d '{"cpf": "123.456.789-00"}'

# 3. Testar com arquivo JSON
curl -X POST http://localhost:8080/api/credit/analyze \
  -H "Content-Type: application/json" \
  -d @test.json
```

9. COMANDOS PARA EXECUÇÃO
```bash
docker-compose up --build -d
docker-compose logs -f app
docker-compose down
go mod tidy
go run cmd/api/main.go
go build -o credit-api cmd/api/main.go
./credit-api
```

10. DECISÕES DE NEGÓCIO

10.1 Política de Crédito

| Cenário | Decisão | Justificativa |
|---------|---------|---------------|
| Score ≥ 700 | APROVAR | Baixo risco, cliente com bom histórico |
| 500 ≤ Score < 700 | ANALISAR | Risco moderado, requer análise manual |
| Score < 500 | NEGAR | Alto risco de inadimplência |
| Empréstimo no Santander | NEGAR | Instituição com alta taxa de inadimplência |

10.2 Regras de Negócio
- Score de Crédito é o principal indicador de risco
- Empréstimos existentes aumentam o risco
- Instituições específicas podem ser consideradas de alto risco
- Decisões são tomadas automaticamente para scores extremos
- Scores médios requerem intervenção humana

10.3 Métricas de Sucesso
- Tempo de resposta < 100ms
- Taxa de aprovação automática > 70%
- Precisão das decisões > 95%
- Zero downtime durante atualizações

11. ENTREGÁVEIS
- Código fonte em Go
- Dockerfile + Docker Compose
- Banco de dados PostgreSQL com dados de exemplo
- Arquivo JSON com dados do bureau
- Régua de risco implementada
- API REST documentada
- Arquivo de teste (.json)
- README com instruções
- Arquitetura Clean implementada
- Goroutines para buscas paralelas (PostgreSQL + Bureau)
