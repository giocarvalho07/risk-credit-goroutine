package service

import (
	"github.com/seu-usuario/risk-credit/internal/model"
)

type CreditRuler struct{}

type RiskResult struct {
	NivelRisco string
	Acao       string
	Descricao  string
}

func NewCreditRuler() *CreditRuler {
	return &CreditRuler{}
}

func (r *CreditRuler) Evaluate(bureau model.BureauData) RiskResult {
	if bureau.Instituicao == "Santander" {
		return RiskResult{
			NivelRisco: "ALTO",
			Acao:       "NEGAR",
			Descricao:  "Cliente possui empréstimo com o Santander",
		}
	}

	switch {
	case bureau.ScoreCredito >= 700:
		return RiskResult{
			NivelRisco: "BAIXO",
			Acao:       "APROVAR",
			Descricao:  "Score de crédito alto, cliente confiável",
		}
	case bureau.ScoreCredito >= 500:
		return RiskResult{
			NivelRisco: "MEDIO",
			Acao:       "ANALISAR",
			Descricao:  "Score de crédito médio, requer análise adicional",
		}
	default:
		return RiskResult{
			NivelRisco: "ALTO",
			Acao:       "NEGAR",
			Descricao:  "Score de crédito baixo, cliente de alto risco",
		}
	}
}
