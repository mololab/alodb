package agent

import "time"

type QueryType string

const (
	QueryTypeRead   QueryType = "read"
	QueryTypeCreate QueryType = "create"
	QueryTypeUpdate QueryType = "update"
	QueryTypeDelete QueryType = "delete"
)

var validQueryTypes = map[QueryType]bool{
	QueryTypeRead:   true,
	QueryTypeCreate: true,
	QueryTypeUpdate: true,
	QueryTypeDelete: true,
}

func IsValidQueryType(t QueryType) bool {
	return validQueryTypes[t]
}

type Query struct {
	Title       string    `json:"title"`
	Query       string    `json:"query"`
	Description string    `json:"description"`
	Type        QueryType `json:"type"`
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
