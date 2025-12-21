package dto

import domainAgent "github.com/mololab/alodb/internal/domain/agent"

// ChatRequest represents the incoming chat request from the API
type ChatRequest struct {
	SessionID        string `json:"session_id,omitempty"`
	Message          string `json:"message" binding:"required"`
	ConnectionString string `json:"connection_string" binding:"required"`
}

// ToDomain converts the DTO to a domain object
func (r *ChatRequest) ToDomain() domainAgent.ChatRequest {
	return domainAgent.ChatRequest{
		SessionID:        r.SessionID,
		Message:          r.Message,
		ConnectionString: r.ConnectionString,
	}
}

// Query represents a SQL query in the API response
type Query struct {
	Title       string `json:"title"`
	Query       string `json:"query"`
	Description string `json:"description"`
}

// ChatResponse represents the API response
type ChatResponse struct {
	Success   bool    `json:"success"`
	SessionID string  `json:"session_id,omitempty"`
	Message   string  `json:"message,omitempty"`
	Queries   []Query `json:"queries,omitempty"`
	Error     string  `json:"error,omitempty"`
}

// FromDomain creates a ChatResponse from a domain response
func ChatResponseFromDomain(resp *domainAgent.ChatResponse) ChatResponse {
	var queries []Query
	for _, q := range resp.Queries {
		queries = append(queries, Query{
			Title:       q.Title,
			Query:       q.Query,
			Description: q.Description,
		})
	}

	return ChatResponse{
		Success:   true,
		SessionID: resp.SessionID,
		Message:   resp.Message,
		Queries:   queries,
	}
}

// ErrorResponse creates an error response
func ErrorResponse(err string) ChatResponse {
	return ChatResponse{
		Success: false,
		Error:   err,
	}
}
