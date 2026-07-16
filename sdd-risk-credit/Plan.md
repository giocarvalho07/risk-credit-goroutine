# Plan - Sistema de Avaliação de Crédito

## 1. Estratégia de Implementação

### Abordagem
Desenvolvimento iterativo e incremental, começando pela infraestrutura e evoluindo até a funcionalidade completa com goroutines.

### Metodologia
- Desenvolvimento por camadas (Clean Architecture)
- Testes contínuos
- Versionamento com Git

## 2. Fases do Projeto

### Fase 1: Configuração do Ambiente
**Objetivo:** Estruturar o projeto e configurar o ambiente de desenvolvimento

| Atividade | Prioridade | Status |
|-----------|------------|--------|
| Criar estrutura de diretórios | Alta | ⬜ |
| Inicializar módulo Go | Alta | ⬜ |
| Criar go.mod com dependências | Alta | ⬜ |
| Configurar Dockerfile | Alta | ⬜ |
| Configurar docker-compose.yml | Alta | ⬜ |
| Criar config.yaml | Alta | ⬜ |
| Criar init.sql | Alta | ⬜ |

**Entregáveis:**
- Estrutura de diretórios completa
- Arquivo go.mod
- Dockerfile funcional
- docker-compose.yml funcional
- Script SQL de inicialização

### Fase 2: Camada de Infraestrutura
**Objetivo:** Implementar conexão com banco de dados e configurações

| Atividade | Prioridade | Status |
|-----------|------------|--------|
| Implementar database/connection.go | Alta | ⬜ |
| Implementar config/config.go | Alta | ⬜ |
| Testar conexão PostgreSQL | Alta | ⬜ |

**Entregáveis:**
- Conexão PostgreSQL configurada
- Pool de conexões otimizado
- Leitura de configurações via Viper

### Fase 3: Modelos de Dados
**Objetivo:** Definir estruturas de dados do sistema

| Atividade | Prioridade | Status |
|-----------|------------|--------|
| Criar model/customer.go | Alta | ⬜ |
| Criar model/bureau.go | Alta | ⬜ |
| Criar model/credit_response.go | Alta | ⬜ |

**Entregáveis:**
- Modelo Customer
- Modelo BureauData
- Modelo CreditResponse

### Fase 4: Camada de Repositório
**Objetivo:** Implementar acesso ao banco de dados

| Atividade | Prioridade | Status |
|-----------|------------|--------|
| Criar interface CustomerRepository | Alta | ⬜ |
| Implementar GetByCPF | Alta | ⬜ |
| Testar queries SQL | Alta | ⬜ |

**Entregáveis:**
- Repository com consulta por CPF
- Tratamento de erros adequado

### Fase 5: Lógica de Negócio
**Objetivo:** Implementar regras de avaliação de crédito

| Atividade | Prioridade | Status |
|-----------|------------|--------|
| Criar credit_ruler.go | Alta | ⬜ |
| Implementar regra Score >= 700 | Alta | ⬜ |
| Implementar regra 500 <= Score < 700 | Alta | ⬜ |
| Implementar regra Score < 500 | Alta | ⬜ |
| Implementar regra Santander | Alta | ⬜ |
| Criar loadBureauData | Alta | ⬜ |

**Entregáveis:**
- Régua de risco completa
- Leitura de dados do bureau

### Fase 6: Camada de Serviço com Goroutines
**Objetivo:** Orquestrar fluxo de negócio com buscas paralelas

| Atividade | Prioridade | Status |
|-----------|------------|--------|
| Criar credit_service.go | Alta | ⬜ |
| Implementar sync.WaitGroup | Alta | ⬜ |
| Implementar sync.Mutex | Alta | ⬜ |
| Criar goroutine para PostgreSQL | Alta | ⬜ |
| Criar goroutine para Bureau JSON | Alta | ⬜ |
| Implementar wg.Wait() | Alta | ⬜ |
| Integrar repository + ruler | Alta | ⬜ |

**Entregáveis:**
- Serviço com goroutines funcionais
- Buscas paralelas PostgreSQL + JSON
- Sincronização correta com WaitGroup
- Proteção de dados com Mutex

### Fase 7: API REST
**Objetivo:** Criar endpoint HTTP

| Atividade | Prioridade | Status |
|-----------|------------|--------|
| Criar credit_handler.go | Alta | ⬜ |
| Implementar POST /api/credit/analyze | Alta | ⬜ |
| Validar entrada (CPF) | Alta | ⬜ |
| Retornar respostas JSON | Alta | ⬜ |
| Criar endpoint de health check | Média | ⬜ |

**Entregáveis:**
- Endpoint funcional
- Validação de entrada
- Respostas padronizadas

### Fase 8: Entry Point
**Objetivo:** Ponto de entrada da aplicação

| Atividade | Prioridade | Status |
|-----------|------------|--------|
| Criar cmd/api/main.go | Alta | ⬜ |
| Configurar dependências | Alta | ⬜ |
| Inicializar router Gin | Alta | ⬜ |
| Executar aplicação | Alta | ⬜ |

**Entregáveis:**
- main.go funcional
- Inicialização completa

### Fase 9: Dados de Teste
**Objetivo:** Preparar dados para validação

| Atividade | Prioridade | Status |
|-----------|------------|--------|
| Criar dados de exemplo no init.sql | Alta | ⬜ |
| Criar bureau_data.json | Alta | ⬜ |
| Criar test.json | Média | ⬜ |

**Entregáveis:**
- Dados de teste no PostgreSQL
- Dados de bureau em JSON

### Fase 10: Testes e Validação
**Objetivo:** Garantir qualidade do sistema

| Atividade | Prioridade | Status |
|-----------|------------|--------|
| Testar containerização | Alta | ⬜ |
| Testar endpoint com curl | Alta | ⬜ |
| Validar todas as regras de risco | Alta | ⬜ |
| Testar cenários de erro | Média | ⬜ |
| **Validar goroutines (buscas paralelas)** | **Alta** | ⬜ |
| **Verificar ausência de race conditions** | **Alta** | ⬜ |

**Entregáveis:**
- Todos os cenários testados
- Sistema funcionando em Docker
- Goroutines funcionando corretamente

## 3. Cronograma

```
Fase 1-2:  ████████████░░░░░░░░░░░░░░░░░░  (Infraestrutura)
Fase 3-4:  ░░░░░░░░████████████░░░░░░░░░░  (Modelos + Repository)
Fase 5:    ░░░░░░░░░░░░░░░░████░░░░░░░░░░  (Lógica de Negócio)
Fase 6:    ░░░░░░░░░░░░░░░░░░░████████░░░  (Goroutines + Service)
Fase 7-8:  ░░░░░░░░░░░░░░░░░░░░░░░░████░░  (API + Main)
Fase 9-10: ░░░░░░░░░░░░░░░░░░░░░░░░░░░░██  (Testes)
```

## 4. Dependências Externas

| Dependência | Versão | Finalidade |
|-------------|--------|------------|
| Go | 1.21+ | Linguagem |
| PostgreSQL | Render | Banco de dados |
| Docker | latest | Containerização |
| Docker Compose | latest | Orquestração |

## 5. Riscos e Mitigações

| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| Conexão PostgreSQL falhar | Média | Alto | Health check + retry |
| Arquivo JSON não encontrado | Baixa | Médio | Validação de path |
| Score fora do range | Baixa | Baixo | Tratamento de fallback |
| Porta em uso | Baixa | Médio | Configuração via config.yaml |
| **Race condition em goroutines** | **Baixa** | **Alto** | **Uso de sync.Mutex** |
| **Goroutine com erro não tratado** | **Baixa** | **Médio** | **Captura de erro via Mutex** |

## 6. Critérios de Aceite

- [ ] API responde POST /api/credit/analyze
- [ ] Dados do cliente são buscados do PostgreSQL
- [ ] Dados do bureau são lidos do JSON
- [ ] **Buscas PostgreSQL e JSON ocorrem em paralelo (goroutines)**
- [ ] **WaitGroup sincroniza corretamente as goroutines**
- [ ] **Mutex protege variáveis compartilhadas**
- [ ] Régua de risco é aplicada corretamente
- [ ] Decisão é retornada no formato especificado
- [ ] Sistema roda via Docker
- [ ] Health check retorna status ok
- [ ] **Race detector não reporta problemas** (`go run -race`)
