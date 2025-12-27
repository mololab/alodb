package agent

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

type ProviderConfig struct {
	EnvKey string
	Models []Model
}

var ProviderRegistry = map[Provider]ProviderConfig{
	ProviderGoogle: {
		EnvKey: "GOOGLE_API_KEY",
		Models: GoogleModels,
	},
	ProviderOpenAI: {
		EnvKey: "OPENAI_API_KEY",
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
