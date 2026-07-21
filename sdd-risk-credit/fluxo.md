# Fluxo da API - Risk Credit

## Rotas Disponíveis

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/health` | Health check - verifica status da API |
| POST | `/api/credit/analyze` | Análise de crédito - retorna decisão baseada no CPF |

---

## Configuração do Servidor

| Parâmetro | Valor |
|-----------|-------|
| Host | definido em `config.yaml` |
| Porta | definida em `config.yaml` |
| Framework | Gin (Go) |

---

## Banco de Dados

| Propriedade | Valor |
|-------------|-------|
| Tipo | PostgreSQL |
| Driver | `github.com/lib/pq` |
| Tabela | `customers` |
| Configuração | via `config.yaml` |

### Estrutura da Tabela `customers`

| Coluna | Tipo | Constraints |
|--------|------|-------------|
| id | SERIAL | PRIMARY KEY |
| nome | VARCHAR(100) | NOT NULL |
| cpf | VARCHAR(14) | NOT NULL, UNIQUE |
| salario | DECIMAL(10,2) | NOT NULL |
| profissao | VARCHAR(100) | NOT NULL |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP |

### Dados Semente (5 registros)

| Nome | CPF | Salário | Profissão |
|------|-----|---------|-----------|
| João Silva | 123.456.789-00 | 8.500,00 | Engenheiro |
| Maria Santos | 987.654.321-00 | 12.000,00 | Doutora |
| Pedro Oliveira | 456.789.123-00 | 3.200,00 | Estudante |
| Ana Costa | 321.654.987-00 | 15.000,00 | Advogada |
| Carlos Ferreira | 789.123.456-00 | 2.800,00 | Vendedor |

### Pool de Conexões

| Parâmetro | Descrição |
|-----------|-----------|
| MaxOpenConns | Máximo de conexões abertas |
| MaxIdleConns | Máximo de conexões ociosas |
| ConnMaxLife | Tempo máximo de vida da conexão |

---

## Fonte de Dados - Bureau de Crédito

| Propriedade | Valor |
|-------------|-------|
| Tipo | Arquivo JSON estático |
| Caminho | `internal/data/bureau_data.json` |
| Carregamento | `os.ReadFile` + unmarshal |

### Estrutura do JSON Bureau

| Campo | Tipo | Descrição |
|-------|------|-----------|
| id | int | Identificador único |
| cpf | string | CPF do cliente |
| score_credito | int | Pontuação de crédito (350-820) |
| emprestimo_existente | bool | Possui empréstimo ativo |
| tipo_emprestimo | string | Tipo do empréstimo (consignado, pessoal, rotativo) |
| instituicao | string | Instituição financeira |

### Registros do Bureau

| CPF | Score | Empréstimo | Tipo | Instituição |
|-----|-------|------------|------|-------------|
| 123.456.789-00 | 750 | Não | - | - |
| 987.654.321-00 | 820 | Sim | Consignado | Banco do Brasil |
| 456.789.123-00 | 450 | Sim | Pessoal | Bradesco |
| 321.654.987-00 | 680 | Não | - | - |
| 789.123.456-00 | 350 | Sim | Rotativo | Cielo |

---

## Regras de Risco

| Condição | Nível | Ação |
|----------|-------|------|
| Instituição == "Santander" | ALTO | NEGAR |
| Score >= 700 | BAIXO | APROVAR |
| Score >= 500 | MEDIO | ANALISAR |
| Score < 500 | ALTO | NEGAR |

---

## Mensagens de Retorno

| Ação | Mensagem |
|------|----------|
| APROVAR | "Parabéns! Seu crédito foi aprovado..." |
| NEGAR | "Desculpe, seu crédito foi negado..." |
| ANALISAR | "Seu crédito está em análise..." |

---

## Resposta da API

| Campo | Tipo | Descrição |
|-------|------|-----------|
| nome | string | Nome do cliente |
| cpf | string | CPF do cliente |
| score_credito | int | Pontuação de crédito |
| nivel_risco | string | BAIXO, MEDIO, ALTO |
| acao | string | APROVAR, NEGAR, ANALISAR |
| descricao | string | Descrição do resultado |
| mensagem | string | Mensagem para o cliente |

---

## Códigos de Erro

| HTTP | Descrição |
|------|-----------|
| 200 | Sucesso |
| 400 | CPF obrigatório / inválido |
| 404 | Cliente não encontrado |
| 500 | Erro interno do servidor |

---

## Estrutura do Projeto

```
risk-credit/
├── cmd/api/main.go              # Entry point
├── config.yaml                  # Configuração
├── internal/
│   ├── config/config.go         # Carrega config
│   ├── database/connection.go   # Conexão PostgreSQL
│   ├── handler/
│   │   └── credit_handler.go    # HTTP handlers
│   ├── service/
│   │   ├── credit_service.go    # Lógica de negócio
│   │   └── credit_ruler.go      # Regras de risco
│   ├── repository/
│   │   └── customer_repository.go # Acesso ao banco
│   ├── model/
│   │   ├── customer.go          # Modelo Customer
│   │   ├── bureau.go            # Modelo Bureau
│   │   └── credit_response.go   # Modelo Response
│   └── data/
│       ├── bureau_data.json     # Dados do bureau
│       └── init.sql             # SQL de referência
├── Dockerfile                   # Build multi-stage
├── docker-compose.yml           # Container setup
├── go.mod                       # Dependências Go
└── go.sum                       # Checksums
```

---

## Dependências Go

| Pacote | Descrição |
|--------|-----------|
| github.com/gin-gonic/gin | Framework HTTP |
| github.com/lib/pq | Driver PostgreSQL |
| github.com/spf13/viper | Configuração YAML |
| github.com/sirupsen/logrus | Logging JSON |

---

## Docker

| Comando | Descrição |
|---------|-----------|
| `docker-compose up --build` | Build e inicia containers |
| `docker-compose down` | Para e remove containers |

### Portas

| Serviço | Porta |
|---------|-------|
| API | 8080 |
| PostgreSQL | 5432 |

---

## Execução Local

```bash
# Instalar dependências
go mod download

# Executar
go run cmd/api/main.go

# Ou buildar
go build -o api.exe ./cmd/api
./api.exe
```

---

## Visão Geral

Request HTTP → Gin Router → Handler → Service (2 goroutines em paralelo) → Repository + Bureau → Database/JSON → Service → Handler → Response HTTP

---

## Fluxo de Entrada (Request → Database)

```
1. POST /api/credit/analyze
   │
   ▼
2. Gin Router (main.go)
   │  Rota registrada: POST /api/credit/analyze
   │
   ▼
3. Handler (credit_handler.go)
   │  • Bind JSON → CreditRequest{CPF}
   │  • Validação: CPF obrigatório (400 se vazio)
   │  • Chama service.Analyze(cpf)
   │
   ▼
4. Service (credit_service.go)
   │  • Lança 2 goroutines em paralelo (WaitGroup + Mutex)
   │
   ├──► Goroutine 1 ──► Repository (customer_repository.go)
   │      • SQL: SELECT * FROM customers WHERE cpf = $1
   │      • Retorna Customer ou nil
   │
   └──► Goroutine 2 ──► loadBureauData() (service)
          • Lê bureau_data.json (os.ReadFile)
          • Busca registro por CPF
          • Retorna BureauData ou nil
```

---

## Fluxo de Processamento (Service)

```
5. Sincronização (wg.Wait)
   │  Ambas goroutines terminam
   │
   ▼
6. Validação
   │  • Erro DB ou Bureau → retorna erro (500)
   │  • Customer ou Bureau nil → retorna nil (404)
   │
   ▼
7. Ruler (credit_ruler.go)
   │  • Instituição == "Santander"? → ALTO / NEGAR
   │  • Score >= 700? → BAIXO / APROVAR
   │  • Score >= 500? → MEDIO / ANALISAR
   │  • Score < 500?  → ALTO / NEGAR
   │
   ▼
8. Monta CreditResponse
   │  • Nome, CPF (do PostgreSQL)
   │  • Score (do JSON)
   │  • NivelRisco, Acao, Descricao (do Ruler)
   │  • Mensagem (texto fixo por ação)
   │
   ▼
9. Retorna para Handler
```

---

## Fluxo de Saída (Response)

```
10. Handler (credit_handler.go)
    │  • Sucesso → HTTP 200 + JSON CreditResponse
    │  • Não encontrado → HTTP 404
    │  • Erro interno → HTTP 500
    │
    ▼
11. HTTP Client recebe:
    {
      "nome": "Joao Silva",
      "cpf": "123.456.789-00",
      "score_credito": 750,
      "nivel_risco": "BAIXO",
      "acao": "APROVAR",
      "descricao": "Score de credito alto...",
      "mensagem": "Parabens! Seu credito foi aprovado..."
    }
```

---

## Camadas (de fora para dentro)

```
cmd/api/main.go          → Entry point, DI, rotas
internal/handler/         → HTTP (request/response)
internal/service/         → Lógica de negócio
internal/repository/      → Acesso ao banco
internal/model/           → Estruturas de dados
internal/database/        → Conexão PostgreSQL
internal/config/          → Configuração YAML
internal/data/            → Dados estáticos (JSON, SQL)
```

---

## Dependências

```
main.go cria:
  config.LoadConfig()        → Config
  database.NewConnection()   → sql.DB
  database.InitDatabase()    → cria tabela + semente
  repository.NewRepository() → CustomerRepository
  service.NewRuler()         → CreditRuler
  service.NewService()       → CreditService
  handler.NewHandler()       → CreditHandler
  router.POST(...)           → registra rota
  router.Run()               → inicia servidor
```
