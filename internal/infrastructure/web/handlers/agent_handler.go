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
	return &AgentHandler{agentService: agentService}
}

func (h *AgentHandler) GetModels(c *gin.Context) {
	providerModels := h.agentService.GetAllProviderModels()
	c.JSON(http.StatusOK, dto.ModelsResponseFromDomain(providerModels))
}

func (h *AgentHandler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	err := h.agentService.DeleteSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true, "session_id": sessionID})
}
