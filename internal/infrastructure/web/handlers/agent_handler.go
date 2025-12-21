package handlers

import (
	"net/http"

	agentApp "github.com/mololab/alodb/internal/application/agent"
	"github.com/mololab/alodb/internal/infrastructure/web/dto"

	"github.com/gin-gonic/gin"
)

// AgentHandler handles agent-related HTTP requests
type AgentHandler struct {
	agentService *agentApp.Service
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(agentService *agentApp.Service) *AgentHandler {
	return &AgentHandler{
		agentService: agentService,
	}
}

// Chat handles the chat endpoint
// @Summary Chat with the database agent
// @Description Send a message to the database agent to get help with queries
// @Tags agent
// @Accept json
// @Produce json
// @Param request body dto.ChatRequest true "Chat request"
// @Success 200 {object} dto.ChatResponse
// @Failure 400 {object} dto.ChatResponse
// @Failure 500 {object} dto.ChatResponse
// @Router /v1/agent/chat [post]
func (h *AgentHandler) Chat(c *gin.Context) {
	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("invalid request: "+err.Error()))
		return
	}

	// Call the agent service with domain object
	resp, err := h.agentService.Chat(c.Request.Context(), req.ToDomain())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("failed to process message: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.ChatResponseFromDomain(resp))
}
