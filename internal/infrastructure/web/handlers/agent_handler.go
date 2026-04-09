package handlers

import (
	"net/http"

	agentApp "github.com/mololab/alodb/internal/application/agent"
	domainAgent "github.com/mololab/alodb/internal/domain/agent"
	"github.com/mololab/alodb/internal/infrastructure/web/dto"

	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	agentService *agentApp.Service
}

func NewAgentHandler(agentService *agentApp.Service) *AgentHandler {
	return &AgentHandler{agentService: agentService}
}

func (h *AgentHandler) GetProviders(c *gin.Context) {
	providers := h.agentService.GetProviders()
	c.JSON(http.StatusOK, dto.ProvidersResponseFromDomain(providers))
}

func (h *AgentHandler) GetProviderModels(c *gin.Context) {
	provider := domainAgent.Provider(c.Param("provider"))

	cfg, ok := domainAgent.ProviderRegistry[provider]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown provider: " + string(provider)})
		return
	}

	apiKey := c.Query("api_key")
	if apiKey == "" {
		apiKey = c.GetHeader(cfg.HeaderKey)
	}

	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required via api_key query param or " + cfg.HeaderKey + " header"})
		return
	}

	models, err := h.agentService.GetModels(c.Request.Context(), provider, apiKey)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch models: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ModelsResponseFromDomain(models))
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
