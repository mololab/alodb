package dto

import domainAgent "github.com/mololab/alodb/internal/domain/agent"

type Model struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

type ProviderMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ProviderResponse struct {
	Provider  string           `json:"provider"`
	HeaderKey string           `json:"header_key"`
	Metadata  ProviderMetadata `json:"metadata"`
}

type ProvidersResponse struct {
	Providers []ProviderResponse `json:"providers"`
}

func ProvidersResponseFromDomain(providers []domainAgent.ProviderInfo) ProvidersResponse {
	result := ProvidersResponse{
		Providers: make([]ProviderResponse, 0, len(providers)),
	}

	for _, p := range providers {
		result.Providers = append(result.Providers, ProviderResponse{
			Provider:  string(p.Provider),
			HeaderKey: p.HeaderKey,
			Metadata: ProviderMetadata{
				Name:        p.Metadata.Name,
				Description: p.Metadata.Description,
			},
		})
	}

	return result
}

type ModelsResponse struct {
	Models []Model `json:"models"`
}

func ModelsResponseFromDomain(models []domainAgent.Model) ModelsResponse {
	result := ModelsResponse{
		Models: make([]Model, 0, len(models)),
	}

	for _, m := range models {
		result.Models = append(result.Models, Model{
			Slug:     m.Slug,
			Name:     m.Name,
			Provider: string(m.Provider),
		})
	}

	return result
}
