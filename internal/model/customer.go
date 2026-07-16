package model

type Customer struct {
	ID        int     `json:"id"`
	Nome      string  `json:"nome"`
	CPF       string  `json:"cpf"`
	Salario   float64 `json:"salario"`
	Profissao string  `json:"profissao"`
}
