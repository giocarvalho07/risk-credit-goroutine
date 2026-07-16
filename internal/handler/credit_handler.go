package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seu-usuario/risk-credit/internal/service"
	log "github.com/sirupsen/logrus"
)

type CreditHandler struct {
	service *service.CreditService
}

type CreditRequest struct {
	CPF string `json:"cpf" binding:"required"`
}

func NewCreditHandler(service *service.CreditService) *CreditHandler {
	return &CreditHandler{service: service}
}

func (h *CreditHandler) Analyze(c *gin.Context) {
	var request CreditRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Errorf("Erro ao validar requisição: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "CPF é obrigatório",
			"message": "Por favor, informe um CPF válido",
		})
		return
	}

	if request.CPF == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "CPF não pode ser vazio",
			"message": "Por favor, informe um CPF válido",
		})
		return
	}

	response, err := h.service.Analyze(request.CPF)
	if err != nil {
		log.Errorf("Erro ao analisar crédito: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erro interno do servidor",
			"message": "Não foi possível processar a solicitação",
		})
		return
	}

	if response == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Cliente não encontrado",
			"message": "Não foi possível encontrar cliente com o CPF informado",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
