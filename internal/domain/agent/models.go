package agent

import "sort"

type Provider string

const (
	ProviderGoogle Provider = "google"
	ProviderOpenAI Provider = "openai"
)

type Model struct {
	Slug     string   `json:"slug"`
	Name     string   `json:"name"`
	Provider Provider `json:"provider"`
}

var GoogleModels = []Model{
	{Slug: "gemini-3-pro-preview", Name: "Gemini 3 Pro Preview", Provider: ProviderGoogle},
	{Slug: "gemini-3-flash-preview", Name: "Gemini 3 Flash Preview", Provider: ProviderGoogle},
	{Slug: "gemini-2.5-flash", Name: "Gemini 2.5 Flash", Provider: ProviderGoogle},
	{Slug: "gemini-2.5-flash-preview-09-2025", Name: "Gemini 2.5 Flash Preview", Provider: ProviderGoogle},
	{Slug: "gemini-2.5-flash-lite", Name: "Gemini 2.5 Flash Lite", Provider: ProviderGoogle},
	{Slug: "gemini-2.5-pro", Name: "Gemini 2.5 Pro", Provider: ProviderGoogle},
}

var OpenAIModels = []Model{}

type ProviderMetadata struct {
	Name        string
	Description string
}

type ProviderConfig struct {
	HeaderKey string
	Metadata  ProviderMetadata
	Models    []Model
}

var ProviderRegistry = map[Provider]ProviderConfig{
	ProviderGoogle: {
		HeaderKey: "X-Gemini-Api-Key",
		Metadata: ProviderMetadata{
			Name:        "Google Gemini",
			Description: "Google's Gemini family of multimodal AI models",
		},
		Models: GoogleModels,
	},
	// TODO: add when adk supports openai
	ProviderOpenAI: {
		HeaderKey: "X-Openai-Api-Key",
		Metadata: ProviderMetadata{
			Name:        "OpenAI",
			Description: "OpenAI's GPT family of language models",
		},
		Models: OpenAIModels,
	},
}

func GetDefaultModelSlug() string {
	return GoogleModels[0].Slug
}

func GetModelBySlug(slug string) (Model, bool) {
	for _, cfg := range ProviderRegistry {
		for _, m := range cfg.Models {
			if m.Slug == slug {
				return m, true
			}
		}
	}
	return Model{}, false
}

func GetProviderByModel(model Model) (ProviderConfig, bool) {
	cfg, ok := ProviderRegistry[model.Provider]
	return cfg, ok
}

type ProviderModels struct {
	Provider  Provider         `json:"provider"`
	HeaderKey string           `json:"header_key"`
	Metadata  ProviderMetadata `json:"metadata"`
	Models    []Model          `json:"models"`
}

func GetAllProviderModels() []ProviderModels {
	var result []ProviderModels

	for provider, cfg := range ProviderRegistry {
		if len(cfg.Models) == 0 {
			continue
		}

		result = append(result, ProviderModels{
			Provider:  provider,
			HeaderKey: cfg.HeaderKey,
			Metadata:  cfg.Metadata,
			Models:    cfg.Models,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Metadata.Name < result[j].Metadata.Name
	})

	return result
}
