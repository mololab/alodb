package websocket

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	domainAgent "github.com/mololab/alodb/internal/domain/agent"
	ws "github.com/mololab/alodb/internal/domain/websocket"
	infraAgent "github.com/mololab/alodb/internal/infrastructure/agent"
	"github.com/mololab/alodb/pkg/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub     *Hub
	manager *infraAgent.Manager
}

func NewHandler(hub *Hub, manager *infraAgent.Manager) *Handler {
	return &Handler{
		hub:     hub,
		manager: manager,
	}
}

// ServeWS handles websocket connection upgrade and setup
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	apiKey := r.URL.Query().Get("api_key")
	if apiKey == "" {
		apiKey = r.Header.Get("X-Gemini-Api-Key")
	}

	model := r.URL.Query().Get("model")
	if model == "" {
		model = domainAgent.DefaultModelSlug
	}

	if apiKey == "" {
		http.Error(w, "API key required", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error().Err(err).Msg("websocket upgrade failed")
		return
	}

	client := NewClient(conn, h.hub, h)
	client.APIKey = apiKey
	client.Model = model

	h.hub.Register(client)

	client.SendEvent(ws.EventSessionCreated, ws.SessionCreatedPayload{
		SessionID: client.SessionID,
	})

	logger.Info().
		Str("client_id", client.ID).
		Str("session_id", client.SessionID).
		Str("model", model).
		Msg("websocket connection established")

	go client.WritePump()
	go client.ReadPump()
}

// HandleChat processes incoming chat messages from client
func (h *Handler) HandleChat(client *Client, payload *ws.ChatPayload) {
	logger.Debug().
		Str("session_id", client.SessionID).
		Str("message", truncateForLog(payload.Message, 100)).
		Msg("handling chat message")

	model := payload.Model
	if model == "" {
		model = client.Model
	}

	ctx := context.Background()
	agent, err := h.manager.GetAgent(ctx, model, client.APIKey)
	if err != nil {
		logger.Error().Err(err).Str("model", model).Msg("failed to get agent")
		client.SendEvent(ws.EventError, ws.ErrorPayload{
			Code:    "agent_error",
			Message: "Failed to initialize agent: " + err.Error(),
		})
		return
	}

	// create query executor that sends queries to client for execution
	executor := infraAgent.NewClientQueryExecutor(client.ExecuteQuery)

	callbacks := infraAgent.StreamCallbacks{
		OnThinking: func(status string) {
			client.SendEvent(ws.EventThinking, ws.ThinkingPayload{
				Status: status,
			})
		},
		OnToolCall: func(tool string) {
			logger.Debug().Str("tool", tool).Msg("tool being called")
		},
		OnTextDelta: func(delta string) {
			client.SendEvent(ws.EventTextDelta, ws.TextDeltaPayload{
				Delta: delta,
			})
		},
		OnResponseComplete: func(resp *domainAgent.ChatResponse) {
			queries := make([]ws.GeneratedSQL, len(resp.Queries))
			for i, q := range resp.Queries {
				queries[i] = ws.GeneratedSQL{
					Title:       q.Title,
					Query:       q.Query,
					Description: q.Description,
					Type:        string(q.Type),
				}
			}
			client.SendEvent(ws.EventResponseComplete, ws.ResponseCompletePayload{
				Success: true,
				Message: resp.Message,
				Queries: queries,
			})
		},
		OnError: func(err error) {
			client.SendEvent(ws.EventError, ws.ErrorPayload{
				Code:    "processing_error",
				Message: err.Error(),
			})
		},
	}

	go agent.StreamChat(ctx, client.SessionID, payload.Message, executor, callbacks)
}

func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
