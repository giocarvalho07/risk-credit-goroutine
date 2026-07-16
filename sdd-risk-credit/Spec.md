# Spec - Sistema de Avaliação de Crédito

## 1. Visão Geral
Sistema simplificado de avaliação de crédito que consolida dados do cliente (PostgreSQL), informações de bureau (JSON) e aplica uma régua de risco para tomar decisões de crédito, utilizando goroutines para buscas paralelas.

## 2. Objetivo
Desenvolver uma API REST em Go que realiza análise de crédito automatizada, consultando dados do cliente em PostgreSQL e informações de bureau em JSON **em paralelo**, retornando decisão de crédito (APROVAR/NEGAR/ANALISAR).

## 3. Escopo

### 3.1 Incluído
- Buscar dados do cliente no PostgreSQL
- Carregar dados do bureau via arquivo JSON
- Buscar dados do cliente e do bureau **em paralelo com goroutines**
- Aplicar regras de risco
- Retornar decisão (APROVAR/NEGAR/ANALISAR)
- API REST com um único endpoint
- Containerização com Docker

### 3.2 Não Incluído
- Autenticação/autorização
- Múltiplos endpoints
- Interface web
- Integração com sistemas externos de bureau

## 4. Requisitos Funcionais

### RF01: Consulta de Cliente
- **Descrição:** Buscar dados do cliente por CPF no PostgreSQL
- **Entrada:** CPF (string formatada)
- **Saída:** Dados do cliente (id, nome, cpf, salario, profissao)
- **Critério:** Cliente deve existir no banco de dados
- **Execução:** Via goroutine concorrente

### RF02: Consulta de Bureau
- **Descrição:** Carregar dados de bureau do arquivo JSON
- **Entrada:** CPF do cliente
- **Saída:** Dados do bureau (score, empréstimo, instituição)
- **Critério:** Dados devem existir no arquivo JSON
- **Execução:** Via goroutine concorrente

### RF03: Busca Paralela (Goroutines)
- **Descrição:** Executar RF01 e RF02 simultaneamente
- **Mecanismo:** sync.WaitGroup + sync.Mutex
- **Fluxo:**
  1. Lança goroutine para buscar PostgreSQL
  2. Lança goroutine para buscar JSON
  3. WaitGroup.Wait() aguarda ambas
  4. Mutex protege escrita das variáveis compartilhadas
- **Resultado:** Redução de ~50% no tempo de I/O

### RF04: Análise de Risco
- **Descrição:** Aplicar régua de risco sobre os dados do bureau
- **Entrada:** Dados do bureau
- **Saída:** Classificação de risco e ação recomendada
- **Regras:**
  - Score >= 700: BAIXO risco → APROVAR
  - Score >= 500 e < 700: MÉDIO risco → ANALISAR
  - Score < 500: ALTO risco → NEGAR
  - Empréstimo no Santander: ALTO risco → NEGAR

### RF05: Retorno de Decisão
- **Descrição:** Retornar resposta completa da análise
- **Saída:** JSON com nome, cpf, score, nivel_risco, acao, descricao, mensagem

## 5. Requisitos Não Funcionais

### RNF01: Performance
- Tempo de resposta < 100ms
- Buscas paralelas via goroutines

### RNF02: Disponibilidade
- Zero downtime durante atualizações

### RNF03: Precisão
- Taxa de precisão das decisões > 95%

### RNF04: Portabilidade
- Execução via Docker em qualquer ambiente

### RNF05: Concorrência
- Uso de sync.WaitGroup para sincronização
- Uso de sync.Mutex para proteção de variáveis compartilhadas
- Goroutines para I/O paralelo (PostgreSQL + JSON)

## 6. Tecnologias

| Componente | Tecnologia |
|------------|------------|
| Linguagem | Go 1.21+ |
| Framework | Gin |
| Banco de Dados | PostgreSQL (Render) |
| Container | Docker + Docker Compose |
| Configurações | Viper |
| Logs | Logrus |
| Concorrência | Goroutines + sync.WaitGroup + sync.Mutex |

## 7. Modelo de Dados

### 7.1 Tabela Customers (PostgreSQL)
| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| id | SERIAL | Sim | Chave primária auto-increment |
| nome | VARCHAR(100) | Sim | Nome completo do cliente |
| cpf | VARCHAR(14) | Sim | CPF formatado (único) |
| salario | DECIMAL(10,2) | Sim | Salário mensal em R$ |
| profissao | VARCHAR(50) | Não | Profissão do cliente |

### 7.2 Arquivo Bureau (JSON)
| Campo | Tipo | Descrição |
|-------|------|-----------|
| id | INT | Identificador único |
| cpf | STRING | CPF do cliente |
| score_credito | INT | Score de crédito (0-1000) |
| emprestimo_existente | BOOLEAN | Possui empréstimo ativo? |
| tipo_emprestimo | STRING | Tipo do empréstimo |
| instituicao | STRING | Instituição financeira |

## 8. API

### Endpoint: POST /api/credit/analyze

**Request:**
```json
{
  "cpf": "123.456.789-00"
}
```

**Response (200):**
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

**Response (400):**
```json
{
  "erro": "CPF é obrigatório"
}
```

**Response (404):**
```json
{
  "erro": "cliente não encontrado: 123.456.789-00"
}
```

## 9. Arquitetura

### Clean Architecture (simplificada)
```
┌─────────────────────────────────────────────┐
│           CAMADA DE APRESENTAÇÃO            │
│           (handlers/credit_handler.go)      │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│           CAMADA DE CASO DE USO             │
│           (services/credit_service.go)      │
│         ┌───────┴───────┐                   │
│    goroutine 1    goroutine 2               │
│    (PostgreSQL)   (JSON)                    │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│           CAMADA DE REPOSITÓRIO             │
│        (repository/customer_repository.go)  │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│           CAMADA DE INFRAESTRUTURA          │
│        (database/connection.go)             │
└─────────────────────────────────────────────┘
```

### Fluxo de Concorrência
```
CreditService.Analyze(cpf)
│
├─── wg.Add(2)
│
├─── goroutine 1 ──► repo.GetByCPF(cpf)  ──► PostgreSQL
│        └──► mu.Lock() → customer = c
│
├─── goroutine 2 ──► loadBureauData(cpf) ──► JSON
│        └──► mu.Lock() → bureau = b
│
└─── wg.Wait() ◄─── ambas finalizam
         │
         ├── Verifica erros
         ├── ruler.Evaluate(bureau)
         └── Retorna CreditResponse
```

## 10. Dependências

```
github.com/gin-gonic/gin v1.9.1
github.com/lib/pq v1.10.9
github.com/spf13/viper v1.18.2
github.com/sirupsen/logrus v1.9.3
```
