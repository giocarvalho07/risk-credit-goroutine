package config

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Bureau   BureauConfig
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	User         string        `mapstructure:"user"`
	Password     string        `mapstructure:"password"`
	DBName       string        `mapstructure:"name"`
	SSLMode      string        `mapstructure:"sslmode"`
	MaxOpenConns int           `mapstructure:"max_open_conns"`
	MaxIdleConns int           `mapstructure:"max_idle_conns"`
	ConnMaxLife  time.Duration `mapstructure:"conn_max_lifetime"`
}

type BureauConfig struct {
	DataPath string `mapstructure:"data_path"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("Erro ao ler configuração: %v", err)
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Errorf("Erro ao deserializar configuração: %v", err)
		return nil, err
	}

	log.Info("Configuração carregada com sucesso")
	return &config, nil
}
