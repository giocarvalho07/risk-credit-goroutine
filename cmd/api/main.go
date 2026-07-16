package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seu-usuario/risk-credit/internal/config"
	"github.com/seu-usuario/risk-credit/internal/database"
	"github.com/seu-usuario/risk-credit/internal/handler"
	"github.com/seu-usuario/risk-credit/internal/repository"
	"github.com/seu-usuario/risk-credit/internal/service"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Erro ao carregar configuração: %v", err)
	}

	dbConfig := database.DBConfig{
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		DBName:       cfg.Database.DBName,
		SSLMode:      cfg.Database.SSLMode,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
		ConnMaxLife:  cfg.Database.ConnMaxLife * time.Minute,
	}

	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatalf("Erro ao conectar com PostgreSQL: %v", err)
	}
	defer db.Close()

	if err := database.InitDatabase(db); err != nil {
		log.Fatalf("Erro ao inicializar banco de dados: %v", err)
	}

	customerRepo := repository.NewCustomerRepository(db)
	creditRuler := service.NewCreditRuler()
	creditService := service.NewCreditService(customerRepo, creditRuler, cfg.Bureau.DataPath)
	creditHandler := handler.NewCreditHandler(creditService)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	router.POST("/api/credit/analyze", creditHandler.Analyze)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Infof("Servidor iniciando na porta %s", cfg.Server.Port)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
