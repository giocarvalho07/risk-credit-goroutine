package repository

import (
	"database/sql"

	"github.com/seu-usuario/risk-credit/internal/model"
	log "github.com/sirupsen/logrus"
)

type CustomerRepository struct {
	db *sql.DB
}

func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

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
		log.Warnf("Cliente não encontrado para CPF: %s", cpf)
		return nil, nil
	}

	if err != nil {
		log.Errorf("Erro ao buscar cliente por CPF %s: %v", cpf, err)
		return nil, err
	}

	log.Infof("Cliente encontrado: %s", customer.Nome)
	return &customer, nil
}
