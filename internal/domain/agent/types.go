package agent

import "time"

type ChatRequest struct {
	SessionID        string
	Message          string
	ConnectionString string
	Model            string
	APIKey           string // from header
}

type Query struct {
	Title       string `json:"title"`
	Query       string `json:"query"`
	Description string `json:"description"`
}

type ChatResponse struct {
	SessionID string  `json:"session_id"`
	Message   string  `json:"message"`
	Queries   []Query `json:"queries,omitempty"`
}

type AgentConfig struct {
	SchemaCacheTTL time.Duration
}
