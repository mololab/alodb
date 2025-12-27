package dto

import domainAgent "github.com/mololab/alodb/internal/domain/agent"

type ChatRequest struct {
	SessionID        string `json:"session_id,omitempty"`
	Message          string `json:"message" binding:"required"`
	ConnectionString string `json:"connection_string" binding:"required"`
	Model            string `json:"model,omitempty"`
}

func (r *ChatRequest) ToDomain() domainAgent.ChatRequest {
	return domainAgent.ChatRequest{
		SessionID:        r.SessionID,
		Message:          r.Message,
		ConnectionString: r.ConnectionString,
		Model:            r.Model,
	}
}

type Query struct {
	Title       string `json:"title"`
	Query       string `json:"query"`
	Description string `json:"description"`
}

type ChatResponse struct {
	Success   bool    `json:"success"`
	SessionID string  `json:"session_id,omitempty"`
	Message   string  `json:"message,omitempty"`
	Queries   []Query `json:"queries,omitempty"`
	Error     string  `json:"error,omitempty"`
}

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

func ErrorResponse(err string) ChatResponse {
	return ChatResponse{
		Success: false,
		Error:   err,
	}
}

type Model struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

type ModelsResponse struct {
	Models []Model `json:"models"`
}

func ModelsResponseFromDomain(models []domainAgent.Model) ModelsResponse {
	result := ModelsResponse{
		Models: make([]Model, len(models)),
	}

	for i, m := range models {
		result.Models[i] = Model{
			Slug:     m.Slug,
			Name:     m.Name,
			Provider: string(m.Provider),
		}
	}

	return result
}
