package agent

import "time"

type Query struct {
	Title       string   `json:"title"`
	Query       string   `json:"query"`
	Description string   `json:"description"`
	Diagram     string   `json:"diagram,omitempty"`
	UsedTables  []string `json:"-"`
}

// parsed response from agent
type ChatResponse struct {
	SessionID string  `json:"session_id"`
	Message   string  `json:"message"`
	Queries   []Query `json:"queries,omitempty"`
}

type AgentConfig struct {
	SchemaCacheTTL time.Duration
}
