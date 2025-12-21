package agent

import "time"

// ChatRequest represents a request to the database agent
type ChatRequest struct {
	SessionID        string
	Message          string
	ConnectionString string
}

// Query represents a single SQL query with its metadata
type Query struct {
	Title       string `json:"title"`
	Query       string `json:"query"`
	Description string `json:"description"`
}

// ChatResponse represents a response from the database agent
type ChatResponse struct {
	SessionID string  `json:"session_id"`
	Message   string  `json:"message"`
	Queries   []Query `json:"queries,omitempty"`
}

// AgentConfig represents configuration for the database agent
type AgentConfig struct {
	GoogleAPIKey   string
	ModelName      string
	SchemaCacheTTL time.Duration
}
