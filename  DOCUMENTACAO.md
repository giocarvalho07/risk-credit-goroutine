# Risk Credit - Documentação do Projeto

## Visão Geral

O **Risk Credit** é uma API REST em Go que realiza análise de crédito automatizada. O sistema consulta dados do cliente em **PostgreSQL** e informações de bureau em **JSON em paralelo** usando goroutines, aplica uma régua de risco e retorna a decisão de crédito (APROVAR/NEGAR/ANALISAR).

---

## Arquitetura Clean Architecture

O projeto segue o padrão **Clean Architecture**, onde cada camada tem uma responsabilidade específica e depende apenas das camadas internas:

```
┌─────────────────────────────────────────────────────────┐
│                    ROTAS (main.go)                      │
│              Configura endpoints HTTP                   │
│                   GET /health                           │
│              POST /api/credit/analyze                   │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│              HANDLER (credit_handler.go)                │
│           Camada de Apresentação / API                  │
│                                                         │
│  • Recebe requisições HTTP (JSON)                       │
│  • Valida entrada (CPF obrigatório)                     │
│  • Chama o Service                                      │
│  • Retorna resposta HTTP (200, 400, 404, 500)           │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│              SERVICE (credit_service.go)                │
│            Camada de Caso de Uso / Negócio              │
│                                                         │
│  • Orquestra o fluxo de análise                         │
│  • Lança 2 goroutines paralelas:                        │
│    → Goroutine 1: Busca cliente no PostgreSQL           │
│    → Goroutine 2: Busca dados bureau no JSON            │
│  • Sincroniza com WaitGroup + Mutex                     │
│  • Aplica régua de risco                                │
│  • Monta e retorna resposta                             │
└───────────┬──────────────────────┬──────────────────────┘
            │                      │
            ▼                      ▼
┌──────────────────────┐  ┌───────────────────────────────┐
│  REPOSITORY          │  │  RULER (credit_ruler.go)      │
│  (customer_          │  │  Regras de negócio            │
│   repository.go)     │  │                               │
│                      │  │  Score >= 700 → APROVAR       │
│  • Acesso ao banco   │  │  Score >= 500 → ANALISAR      │
│  • Queries SQL       │  │  Score < 500  → NEGAR         │
│  • Busca por CPF     │  │  Santander    → NEGAR         │
└──────────┬───────────┘  └───────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────┐
│              DATABASE (connection.go)                   │
│           Camada de Infraestrutura                      │
│                                                         │
│  • Conexão com PostgreSQL (Render)                      │
│  • Pool de conexões (25 max open, 25 idle)              │
│  • Inicialização da tabela + dados                      │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│              MODEL (model/)                             │
│           Camada de Modelo de Dados                     │
│                                                         │
│  • Customer: dados do cliente (id, nome, cpf, etc)     │
│  • BureauData: dados do bureau (score, empréstimo)      │
│  • CreditResponse: resposta da análise                  │
└─────────────────────────────────────────────────────────┘
```

---

## Fluxo Completo: Da Request até a Resposta

### Passo 1: O Usuário faz a requisição HTTP

```bash
curl -X POST http://localhost:8080/api/credit/analyze \
  -H "Content-Type: application/json" \
  -d '{"cpf": "123.456.789-00"}'
```

O Gin Framework recebe a request e roteia para o handler correto.

---

### Passo 2: O Handler recebe e valida (`credit_handler.go:23`)

```go
func (h *CreditHandler) Analyze(c *gin.Context) {
    var request CreditRequest

    // Deserializa o JSON do body para a struct
    if err := c.ShouldBindJSON(&request); err != nil {
        // Erro: JSON inválido ou CPF ausente → 400
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "CPF é obrigatório",
        })
        return
    }

    // Chama o Service com o CPF
    response, err := h.service.Analyze(request.CPF)
    // ... tratamento de erro e retorno
}
```

**O que acontece aqui:**
- O JSON `{"cpf": "123.456.789-00"}` é deserializado para a struct `CreditRequest`
- Se o campo `cpf` estiver vazio ou faltando, retorna erro 400
- Se tudo OK, chama `service.Analyze(cpf)`

---

### Passo 3: O Service lança as Goroutines (`credit_service.go:27`)

Este é o ponto principal. O Service cria **duas goroutines** que executam **simultaneamente**:

```go
func (s *CreditService) Analyze(cpf string) (*model.CreditResponse, error) {
    var (
        wg        sync.WaitGroup   // Conta quantas goroutines precisam terminar
        mu        sync.Mutex       // Protege acesso às variáveis compartilhadas
        customer  *model.Customer  // Recebe resultado da Goroutine 1
        bureau    *model.BureauData // Recebe resultado da Goroutine 2
        errDB     error            // Erro da Goroutine 1
        errBureau error            // Erro da Goroutine 2
    )

    wg.Add(2)  // Diz ao WaitGroup que haverá 2 goroutines

    // ═══════════════════════════════════════════════════
    // GOROUTINE 1: Busca cliente no PostgreSQL
    // ═══════════════════════════════════════════════════
    go func() {
        defer wg.Done()  // Sinaliza que terminou quando a função retornar
        c, err := s.repo.GetByCPF(cpf)  // Query no banco
        mu.Lock()        // Trava o acesso às variáveis compartilhadas
        defer mu.Unlock() // Destrava ao final
        customer = c     // Salva o resultado
        errDB = err      // Salva o erro (ou nil)
    }()

    // ═══════════════════════════════════════════════════
    // GOROUTINE 2: Busca dados do bureau no JSON
    // ═══════════════════════════════════════════════════
    go func() {
        defer wg.Done()
        b, err := s.loadBureauData(cpf)  // Lê arquivo JSON
        mu.Lock()
        defer mu.Unlock()
        bureau = b
        errBureau = err
    }()

    // ═══════════════════════════════════════════════════
    // BLOQUEIA até ambas goroutines terminarem
    // ═══════════════════════════════════════════════════
    wg.Wait()

    // ... continuação do fluxo
}
```

**O que acontece aqui:**
1. `wg.Add(2)` → informa que 2 goroutines vão rodar
2. Goroutine 1 inicia → faz query no PostgreSQL
3. Goroutine 2 inicia → lê arquivo JSON
4. Ambas rodam **ao mesmo tempo** (paralelismo)
5. `wg.Wait()` → bloqueia até ambas terminarem
6. Quando terminam, o fluxo continua

---

### Passo 4: Repository busca no PostgreSQL (`customer_repository.go:18`)

```go
func (r *CustomerRepository) GetByCPF(cpf string) (*model.Customer, error) {
    query := "SELECT id, nome, cpf, salario, profissao FROM customers WHERE cpf = $1"

    var customer model.Customer
    err := r.db.QueryRow(query, cpf).Scan(
        &customer.ID,
        &customer.Nome,
        &customer.CPF,
        &customer.Salario,
        &customer.Profissao,
    )

    if err == sql.ErrNoRows {
        return nil, nil  // Cliente não encontrado
    }

    log.Infof("Cliente encontrado: %s", customer.Nome)
    return &customer, nil
}
```

**O que acontece aqui:**
- Executa query SQL no PostgreSQL
- Usa `$1` (sintaxe PostgreSQL) em vez de `?` (MySQL)
- Scaneia os resultados para a struct `Customer`
- Retorna o cliente ou nil se não encontrar

---

### Passo 5: Load Bureau lê o JSON (`credit_service.go:95`)

```go
func (s *CreditService) loadBureauData(cpf string) (*model.BureauData, error) {
    // Lê o arquivo JSON inteiro
    file, err := os.ReadFile(s.bureauPath)

    // Deserializa para a struct BureauResponse
    var bureauResponse model.BureauResponse
    json.Unmarshal(file, &bureauResponse)

    // Procura o CPF no array de dados
    for _, data := range bureauResponse.Bureau {
        if data.CPF == cpf {
            return &data, nil  // Encontrou
        }
    }

    return nil, nil  // Não encontrou
}
```

**Exemplo de `bureau_data.json`:**
```json
{
  "bureau": [
    {
      "id": 1,
      "cpf": "123.456.789-00",
      "score_credito": 750,
      "emprestimo_existente": false,
      "tipo_emprestimo": "",
      "instituicao": ""
    }
  ]
}
```

---

### Passo 6: Ruler aplica as regras de risco (`credit_ruler.go:19`)

```go
func (r *CreditRuler) Evaluate(bureau model.BureauData) RiskResult {
    // Regra especial: Santander sempre nega
    if bureau.Instituicao == "Santander" {
        return RiskResult{
            NivelRisco: "ALTO",
            Acao:       "NEGAR",
            Descricao:  "Cliente possui empréstimo com o Santander",
        }
    }

    // Regras baseadas no score
    switch {
    case bureau.ScoreCredito >= 700:
        return RiskResult{NivelRisco: "BAIXO", Acao: "APROVAR", ...}
    case bureau.ScoreCredito >= 500:
        return RiskResult{NivelRisco: "MEDIO", Acao: "ANALISAR", ...}
    default:
        return RiskResult{NivelRisco: "ALTO", Acao: "NEGAR", ...}
    }
}
```

**Tabela de Decisões:**

| Score | Instituição | Nível Risco | Ação |
|-------|-------------|-------------|------|
| >= 700 | Qualquer (exceto Santander) | BAIXO | APROVAR |
| 500-699 | Qualquer (exceto Santander) | MEDIO | ANALISAR |
| < 500 | Qualquer (exceto Santander) | ALTO | NEGAR |
| Qualquer | Santander | ALTO | NEGAR |

---

### Passo 7: Response é montada e retornada (`credit_service.go:82`)

```go
response := &model.CreditResponse{
    Nome:       customer.Nome,       // "João Silva"
    CPF:        customer.CPF,        // "123.456.789-00"
    Score:      bureau.ScoreCredito, // 750
    NivelRisco: result.NivelRisco,   // "BAIXO"
    Acao:       result.Acao,         // "APROVAR"
    Descricao:  result.Descricao,    // "Score de crédito alto..."
    Mensagem:   mensagem,            // "Parabéns! Seu crédito foi aprovado..."
}

return response, nil
```

**Resposta JSON retornada:**
```json
{
  "nome": "João Silva",
  "cpf": "123.456.789-00",
  "score_credito": 750,
  "nivel_risco": "BAIXO",
  "acao": "APROVAR",
  "descricao": "Score de crédito alto, cliente confiável",
  "mensagem": "Parabéns! Seu crédito foi aprovado. Entre em contato conosco para mais detalhes."
}
```

---

## Goroutines: Como Funcionam

### O que é uma Goroutine?

Uma **goroutine** é uma thread leve gerenciada pelo runtime do Go. Elas são extremamente baratas (cerca de 2KB de memória cada) e permitem executar milhares de tarefas simultaneamente.

### No Nosso Projeto

```
                        CreditService.Analyze()
                                  │
                    ┌─────────────┴─────────────┐
                    │                           │
                    ▼                           ▼
            ┌──────────────┐            ┌──────────────┐
            │  GOROUTINE 1 │            │  GOROUTINE 2 │
            │              │            │              │
            │  PostgreSQL  │            │  JSON File   │
            │  (banco)     │            │  (disco)     │
            │              │            │              │
            │  Query SQL   │            │  ReadFile    │
            │  ~50ms       │            │  ~10ms       │
            └──────┬───────┘            └──────┬───────┘
                   │                           │
                   └─────────────┬─────────────┘
                                 │
                                 ▼
                          wg.Wait() ← bloqueia até ambas terminarem
                                 │
                                 ▼
                          Continua o fluxo
```

**Sem goroutines (sequencial):** ~60ms (50ms + 10ms)
**Com goroutines (paralelo):** ~50ms (tempo da mais lenta)

### Sincronização

| Mecanismo | Função |
|-----------|--------|
| `sync.WaitGroup` | Conta goroutines pendentes. `wg.Add(2)` adiciona 2, `wg.Done()` remove 1, `wg.Wait()` bloqueia até chegar a 0 |
| `sync.Mutex` | Trava o acesso a variáveis compartilhadas. `mu.Lock()` trava, `mu.Unlock()` destrava |

### Por que usar Mutex?

Sem o Mutex, ambas as goroutines tentariam escrever nas variáveis `customer`, `bureau`, `errDB`, `errBureau` ao mesmo tempo, causando um **data race** (condição de corrida). O Mutex garante que apenas uma goroutine por vez escreva nessas variáveis.

---

## Onde Ver as Goroutines nos Logs

O projeto usa **Logrus** com formato JSON. Cada mensagem de log indica de qual camada veio.

### Logs de Inicialização

```json
{"level":"info","msg":"Configuração carregada com sucesso","time":"2026-07-16T02:30:00Z"}
{"level":"info","msg":"Tabela customers criada/verificada com sucesso","time":"2026-07-16T02:30:01Z"}
{"level":"info","msg":"Dados iniciais inseridos com sucesso","time":"2026-07-16T02:30:01Z"}
{"level":"info","msg":"Conexão com PostgreSQL estabelecida com sucesso","time":"2026-07-16T02:30:01Z"}
{"level":"info","msg":"Servidor iniciando na porta 8080","time":"2026-07-16T02:30:01Z"}
```

### Logs Durante a Análise (mostra as goroutines em ação)

Quando você faz uma requisição, os logs aparecem nestas ordem:

```json
// 1. Repository - Goroutine 1 encontrou o cliente
{"level":"info","msg":"Cliente encontrado: João Silva","time":"2026-07-16T02:31:00Z"}

// 2. Service - Análise completou com sucesso
{"level":"info","msg":"Crédito analisado para CPF 123.456.789-00: APROVAR","time":"2026-07-16T02:31:00Z"}

// 3. GIN - Request HTTP completou
[GIN] 2026/07/16 - 02:31:00 | 200 | 50ms | 172.19.0.1 | POST "/api/credit/analyze"
```

### Logs de Erro (quando algo falha)

```json
// Erro de conexão com banco
{"level":"error","msg":"Erro ao conectar com PostgreSQL: ...","time":"..."}
{"level":"fatal","msg":"Erro ao conectar com PostgreSQL: ...","time":"..."}

// Erro ao buscar cliente
{"level":"error","msg":"Erro ao buscar cliente por CPF 123.456.789-00: ...","time":"..."}

// Cliente não encontrado
{"level":"warning","msg":"Cliente não encontrado para CPF: 000.000.000-00","time":"..."}

// Erro interno no handler
{"level":"error","msg":"Erro ao analisar crédito: ...","time":"..."}
```

### Como Acompanhar as Goroutines

Para ver as goroutines em tempo real:

```bash
# Ver logs do container em tempo real
docker-compose logs -f app

# Ou rodar localmente
go run cmd/api/main.go
```

**Dica:** As duas goroutines (PostgreSQL e JSON) rodam quase ao mesmo tempo. Nos logs, você verá as mensagens de "Cliente encontrado" e possíveis erros de bureau aparecendo com timestamps muito próximos, indicando execução paralela.

---

## Estrutura de Diretórios

```
risk-credit/
├── cmd/
│   └── api/
│       └── main.go                    ← Ponto de entrada
├── internal/
│   ├── config/
│   │   └── config.go                  ← Leitura do config.yaml (Viper)
│   ├── database/
│   │   └── connection.go              ← Conexão PostgreSQL + InitDatabase
│   ├── model/
│   │   ├── customer.go                ← Modelo Customer
│   │   ├── bureau.go                  ← Modelo BureauData
│   │   └── credit_response.go         ← Modelo CreditResponse
│   ├── repository/
│   │   └── customer_repository.go     ← Queries SQL (GetByCPF)
│   ├── service/
│   │   ├── credit_service.go          ← Orquestração + Goroutines
│   │   └── credit_ruler.go            ← Regras de risco
│   ├── handler/
│   │   └── credit_handler.go          ← Endpoint HTTP
│   └── data/
│       ├── init.sql                   ← Script SQL de inicialização
│       └── bureau_data.json           ← Dados do bureau
├── config.yaml                        ← Configurações (host, porta, etc)
├── docker-compose.yml                 ← Orquestração Docker
├── Dockerfile                         ← Build da imagem
├── go.mod                             ← Dependências Go
└── risk-credit.postman_collection.json ← Collection Postman
```

---

## Modelos de Dados

### Customer (`internal/model/customer.go`)
```go
type Customer struct {
    ID        int     `json:"id"`         // ID auto-increment (SERIAL)
    Nome      string  `json:"nome"`       // Nome completo
    CPF       string  `json:"cpf"`        // CPF formatado (único)
    Salario   float64 `json:"salario"`    // Salário mensal R$
    Profissao string  `json:"profissao"`  // Profissão
}
```

### BureauData (`internal/model/bureau.go`)
```go
type BureauData struct {
    ID                  int    `json:"id"`                   // ID do registro
    CPF                 string `json:"cpf"`                  // CPF do cliente
    ScoreCredito        int    `json:"score_credito"`        // Score 0-1000
    EmprestimoExistente bool   `json:"emprestimo_existente"` // Tem empréstimo?
    TipoEmprestimo      string `json:"tipo_emprestimo"`      // Tipo (pessoal, consignado, etc)
    Instituicao         string `json:"instituicao"`          // Banco/Instituição
}
```

### CreditResponse (`internal/model/credit_response.go`)
```go
type CreditResponse struct {
    Nome       string `json:"nome"`        // Nome do cliente
    CPF        string `json:"cpf"`         // CPF
    Score      int    `json:"score_credito"` // Score obtido
    NivelRisco string `json:"nivel_risco"` // BAIXO / MEDIO / ALTO
    Acao       string `json:"acao"`        // APROVAR / ANALISAR / NEGAR
    Descricao  string `json:"descricao"`   // Descrição da decisão
    Mensagem   string `json:"mensagem"`    // Mensagem amigável
}
```

---

## Inicialização do Banco (InitDatabase)

Quando a aplicação inicia, `database.InitDatabase(db)` é chamado automaticamente:

1. **Cria a tabela** `customers` se não existir (`CREATE TABLE IF NOT EXISTS`)
2. **Insere 5 clientes** de exemplo usando `ON CONFLICT (cpf) DO NOTHING` (não duplica se já existir)

Isso garante que o banco esteja sempre pronto para uso, sem precisar rodar scripts manualmente.

---

## Configuração (`config.yaml`)

```yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  host: "dpg-d9c3bvhkh4rs73bvdjog-a.oregon-postgres.render.com"
  port: 5432
  user: "root"
  password: "6oz39OKVaYhgy6n7i5pni0rz5BRL0SuX"
  name: "db_credit"
  sslmode: "require"
  max_open_conns: 25
  max_idle_conns: 25
  conn_max_lifetime: 5

bureau:
  data_path: "internal/data/bureau_data.json"
```

---

## Como Rodar

### Com Docker (recomendado para deploy)
```bash
docker-compose up --build
```

### Localmente
```bash
go mod tidy
go run cmd/api/main.go
```

### Testar
```bash
# Health check
curl http://localhost:8080/health

# Análise de crédito
curl -X POST http://localhost:8080/api/credit/analyze \
  -H "Content-Type: application/json" \
  -d '{"cpf": "123.456.789-00"}'
```

Ou importe a collection Postman `risk-credit.postman_collection.json` no Postman.
