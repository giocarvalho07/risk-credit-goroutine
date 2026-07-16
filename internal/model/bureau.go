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
