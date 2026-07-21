# Fluxo da API - Risk Credit

## Visão Geral

Request HTTP → Gin Router → Handler → Service (2 goroutines em paralelo) → Repository + Bureau → Database/JSON → Service → Handler → Response HTTP

---

## Códigos de API

| HTTP | Descrição |
|------|-----------|
| 200 | Sucesso |
| 400 | CPF obrigatório / inválido |
| 404 | Cliente não encontrado |
| 500 | Erro interno do servidor |

---

## Rotas Disponíveis

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/health` | Health check - verifica status da API |
| POST | `/api/credit/analyze` | Análise de crédito - retorna decisão baseada no CPF |

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
