package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type DBConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLife  time.Duration
}

func NewPostgresConnection(config DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Errorf("Erro ao abrir conexão com PostgreSQL: %v", err)
		return nil, err
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLife)

	if err := db.Ping(); err != nil {
		log.Errorf("Erro ao conectar com PostgreSQL: %v", err)
		return nil, err
	}

	log.Info("Conexão com PostgreSQL estabelecida com sucesso")
	return db, nil
}

func InitDatabase(db *sql.DB) error {
	createTable := `
	CREATE TABLE IF NOT EXISTS customers (
		id SERIAL PRIMARY KEY,
		nome VARCHAR(100) NOT NULL,
		cpf VARCHAR(14) NOT NULL UNIQUE,
		salario DECIMAL(10,2) NOT NULL,
		profissao VARCHAR(100) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(createTable); err != nil {
		log.Errorf("Erro ao criar tabela customers: %v", err)
		return err
	}
	log.Info("Tabela customers criada/verificada com sucesso")

	insertData := `
	INSERT INTO customers (nome, cpf, salario, profissao) VALUES
	('João Silva', '123.456.789-00', 8500.00, 'Engenheiro'),
	('Maria Santos', '987.654.321-00', 12000.00, 'Doutora'),
	('Pedro Oliveira', '456.789.123-00', 3200.00, 'Estudante'),
	('Ana Costa', '321.654.987-00', 15000.00, 'Advogada'),
	('Carlos Ferreira', '789.123.456-00', 2800.00, 'Vendedor')
	ON CONFLICT (cpf) DO NOTHING;`

	if _, err := db.Exec(insertData); err != nil {
		log.Errorf("Erro ao inserir dados iniciais: %v", err)
		return err
	}
	log.Info("Dados iniciais inseridos com sucesso")

	return nil
}
