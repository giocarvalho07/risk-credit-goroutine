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
