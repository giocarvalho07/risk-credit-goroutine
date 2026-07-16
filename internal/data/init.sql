CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(100) NOT NULL,
    cpf VARCHAR(14) NOT NULL UNIQUE,
    salario DECIMAL(10,2) NOT NULL,
    profissao VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO customers (nome, cpf, salario, profissao) VALUES
('João Silva', '123.456.789-00', 8500.00, 'Engenheiro'),
('Maria Santos', '987.654.321-00', 12000.00, 'Doutora'),
('Pedro Oliveira', '456.789.123-00', 3200.00, 'Estudante'),
('Ana Costa', '321.654.987-00', 15000.00, 'Advogada'),
('Carlos Ferreira', '789.123.456-00', 2800.00, 'Vendedor')
ON CONFLICT (cpf) DO NOTHING;
