package dto

import domainAgent "github.com/mololab/alodb/internal/domain/agent"

type ChatRequest struct {
	SessionID        string `json:"session_id,omitempty"`
	Message          string `json:"message" binding:"required"`
	ConnectionString string `json:"connection_string" binding:"required"`
	Model            string `json:"model,omitempty"`
}

func (r *ChatRequest) ToDomain(apiKey string) domainAgent.ChatRequest {
	return domainAgent.ChatRequest{
		SessionID:        r.SessionID,
		Message:          r.Message,
		ConnectionString: r.ConnectionString,
		Model:            r.Model,
		APIKey:           apiKey,
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

type ProviderMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ProviderModelsResponse struct {
	HeaderKey string           `json:"header_key"`
	Metadata  ProviderMetadata `json:"metadata"`
	Models    []Model          `json:"models"`
}

type ModelsResponse struct {
	Providers []ProviderModelsResponse `json:"providers"`
}

func ModelsResponseFromDomain(providerModels []domainAgent.ProviderModels) ModelsResponse {
	result := ModelsResponse{
		Providers: make([]ProviderModelsResponse, 0, len(providerModels)),
	}

	for _, pm := range providerModels {
		models := make([]Model, len(pm.Models))
		for i, m := range pm.Models {
			models[i] = Model{
				Slug:     m.Slug,
				Name:     m.Name,
				Provider: string(m.Provider),
			}
		}

		result.Providers = append(result.Providers, ProviderModelsResponse{
			HeaderKey: pm.HeaderKey,
			Metadata: ProviderMetadata{
				Name:        pm.Metadata.Name,
				Description: pm.Metadata.Description,
			},
			Models: models,
		})
	}

	return result
}
