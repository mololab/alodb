package agent

type Provider string

const (
	ProviderGoogle Provider = "google"
)

const DefaultModelSlug = "gemini-2.5-flash"

type Model struct {
	Slug     string   `json:"slug"`
	Name     string   `json:"name"`
	Provider Provider `json:"provider"`
}

type ProviderMetadata struct {
	Name        string
	Description string
}

type ProviderConfig struct {
	HeaderKey string
	Metadata  ProviderMetadata
}

var ProviderRegistry = map[Provider]ProviderConfig{
	ProviderGoogle: {
		HeaderKey: "X-Gemini-Api-Key",
		Metadata: ProviderMetadata{
			Name:        "Google Gemini",
			Description: "Google's Gemini family of multimodal AI models",
		},
	},
}

type ProviderInfo struct {
	Provider  Provider         `json:"provider"`
	HeaderKey string           `json:"header_key"`
	Metadata  ProviderMetadata `json:"metadata"`
}

func GetProviders() []ProviderInfo {
	var result []ProviderInfo
	for provider, cfg := range ProviderRegistry {
		result = append(result, ProviderInfo{
			Provider:  provider,
			HeaderKey: cfg.HeaderKey,
			Metadata:  cfg.Metadata,
		})
	}
	return result
}
