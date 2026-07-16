package service

import (
	"encoding/json"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/seu-usuario/risk-credit/internal/model"
	"github.com/seu-usuario/risk-credit/internal/repository"
	log "github.com/sirupsen/logrus"
)

type CreditService struct {
	repo       *repository.CustomerRepository
	ruler      *CreditRuler
	bureauPath string
}

func NewCreditService(repo *repository.CustomerRepository, ruler *CreditRuler, bureauPath string) *CreditService {
	return &CreditService{
		repo:       repo,
		ruler:      ruler,
		bureauPath: bureauPath,
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

	startTotal := time.Now()
	log.Infof("[MAIN][%s] Iniciando análise de crédito para CPF: %s", startTotal.Format("15:04:05.000"), cpf)

	wg.Add(2)

	// ═══════════════════════════════════════════════════
	// GOROUTINE 1: Busca cliente no PostgreSQL
	// ═══════════════════════════════════════════════════
	go func() {
		defer wg.Done()
		start := time.Now()
		goroutineID := getGoroutineID()
		log.Infof("[GOROUTINE-1][%s][%s] INICIADA - Buscando cliente no PostgreSQL para CPF: %s", goroutineID, start.Format("15:04:05.000"), cpf)

		c, err := s.repo.GetByCPF(cpf)

		mu.Lock()
		defer mu.Unlock()
		customer = c
		errDB = err

		end := time.Now()
		duration := end.Sub(start)

		if err != nil {
			log.Errorf("[GOROUTINE-1][%s][%s] FALHA (%.2fms) - Erro ao buscar cliente: %v", goroutineID, end.Format("15:04:05.000"), float64(duration.Microseconds())/1000, err)
		} else if c != nil {
			log.Infof("[GOROUTINE-1][%s][%s] CONCLUÍDA (%.2fms) - Cliente encontrado: %s (ID: %d)", goroutineID, end.Format("15:04:05.000"), float64(duration.Microseconds())/1000, c.Nome, c.ID)
		} else {
			log.Warnf("[GOROUTINE-1][%s][%s] CONCLUÍDA (%.2fms) - Cliente não encontrado para CPF: %s", goroutineID, end.Format("15:04:05.000"), float64(duration.Microseconds())/1000, cpf)
		}
	}()

	// ═══════════════════════════════════════════════════
	// GOROUTINE 2: Busca dados do bureau no JSON
	// ═══════════════════════════════════════════════════
	go func() {
		defer wg.Done()
		start := time.Now()
		goroutineID := getGoroutineID()
		log.Infof("[GOROUTINE-2][%s][%s] INICIADA - Buscando dados do bureau no JSON para CPF: %s", goroutineID, start.Format("15:04:05.000"), cpf)

		b, err := s.loadBureauData(cpf)

		mu.Lock()
		defer mu.Unlock()
		bureau = b
		errBureau = err

		end := time.Now()
		duration := end.Sub(start)

		if err != nil {
			log.Errorf("[GOROUTINE-2][%s][%s] FALHA (%.2fms) - Erro ao buscar bureau: %v", goroutineID, end.Format("15:04:05.000"), float64(duration.Microseconds())/1000, err)
		} else if b != nil {
			log.Infof("[GOROUTINE-2][%s][%s] CONCLUÍDA (%.2fms) - Bureau encontrado: Score=%d, Instituição=%s", goroutineID, end.Format("15:04:05.000"), float64(duration.Microseconds())/1000, b.ScoreCredito, b.Instituicao)
		} else {
			log.Warnf("[GOROUTINE-2][%s][%s] CONCLUÍDA (%.2fms) - Bureau não encontrado para CPF: %s", goroutineID, end.Format("15:04:05.000"), float64(duration.Microseconds())/1000, cpf)
		}
	}()

	log.Infof("[MAIN][%s] Aguardando goroutines finalizarem...", time.Now().Format("15:04:05.000"))
	wg.Wait()
	endTotal := time.Now()
	durationTotal := endTotal.Sub(startTotal)
	log.Infof("[MAIN][%s] Ambas goroutines finalizadas (%.2fms). Processando resultado...", endTotal.Format("15:04:05.000"), float64(durationTotal.Microseconds())/1000)

	if errDB != nil {
		log.Errorf("[MAIN] Erro retornado da Goroutine-1: %v", errDB)
		return nil, errDB
	}

	if errBureau != nil {
		log.Errorf("[MAIN] Erro retornado da Goroutine-2: %v", errBureau)
		return nil, errBureau
	}

	if customer == nil {
		log.Warnf("[MAIN] Cliente não encontrado para CPF: %s", cpf)
		return nil, nil
	}

	if bureau == nil {
		log.Warnf("[MAIN] Bureau não encontrado para CPF: %s", cpf)
		return nil, nil
	}

	log.Infof("[MAIN] Aplicando régua de risco - Score: %d, Instituição: %s", bureau.ScoreCredito, bureau.Instituicao)
	result := s.ruler.Evaluate(*bureau)
	log.Infof("[MAIN] Resultado da régua: %s → %s", result.NivelRisco, result.Acao)

	mensagem := s.getMensagem(result.Acao)

	response := &model.CreditResponse{
		Nome:       customer.Nome,
		CPF:        customer.CPF,
		Score:      bureau.ScoreCredito,
		NivelRisco: result.NivelRisco,
		Acao:       result.Acao,
		Descricao:  result.Descricao,
		Mensagem:   mensagem,
	}

	log.Infof("[MAIN] Análise concluída em %.2fms - Cliente: %s, Decisão: %s", float64(durationTotal.Microseconds())/1000, customer.Nome, result.Acao)

	return response, nil
}

func (s *CreditService) loadBureauData(cpf string) (*model.BureauData, error) {
	file, err := os.ReadFile(s.bureauPath)
	if err != nil {
		log.Errorf("Erro ao ler arquivo de bureau: %v", err)
		return nil, err
	}

	var bureauResponse model.BureauResponse
	if err := json.Unmarshal(file, &bureauResponse); err != nil {
		log.Errorf("Erro ao deserializar dados do bureau: %v", err)
		return nil, err
	}

	for _, data := range bureauResponse.Bureau {
		if data.CPF == cpf {
			return &data, nil
		}
	}

	return nil, nil
}

func (s *CreditService) getMensagem(acao string) string {
	switch acao {
	case "APROVAR":
		return "Parabéns! Seu crédito foi aprovado. Entre em contato conosco para mais detalhes."
	case "ANALISAR":
		return "Seu crédito está em análise. Entraremos em contato em breve."
	case "NEGAR":
		return "Infelizmente, seu crédito não foi aprovado neste momento."
	default:
		return "Entre em contato com nosso atendimento para mais informações."
	}
}

func getGoroutineID() string {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	id := ""
	for i := 10; i < n; i++ {
		if buf[i] >= '0' && buf[i] <= '9' {
			id += string(buf[i])
		} else {
			break
		}
	}
	return "goroutine-" + id
}
