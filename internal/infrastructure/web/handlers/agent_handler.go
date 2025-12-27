package handlers

import (
	"net/http"

	agentApp "github.com/mololab/alodb/internal/application/agent"
	"github.com/mololab/alodb/internal/infrastructure/web/dto"

	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	agentService *agentApp.Service
}

func NewAgentHandler(agentService *agentApp.Service) *AgentHandler {
	return &AgentHandler{
		agentService: agentService,
	}
}

func (h *AgentHandler) Chat(c *gin.Context) {
	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("invalid request: "+err.Error()))
		return
	}

	resp, err := h.agentService.Chat(c.Request.Context(), req.ToDomain())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("failed to process message: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.ChatResponseFromDomain(resp))
}

func (h *AgentHandler) GetModels(c *gin.Context) {
	models := h.agentService.GetAvailableModels()
	c.JSON(http.StatusOK, dto.ModelsResponseFromDomain(models))
}
