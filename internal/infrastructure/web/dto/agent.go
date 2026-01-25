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
