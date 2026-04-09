package agent

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	domainAgent "github.com/mololab/alodb/internal/domain/agent"
	"github.com/mololab/alodb/pkg/logger"
	"google.golang.org/genai"
)

const modelCacheTTL = 1 * time.Hour

// ModelFetcher fetches available models from Google's API and caches the results.
type ModelFetcher struct {
	mu           sync.RWMutex
	cachedModels []domainAgent.Model
	lastFetched  time.Time
}

func NewModelFetcher() *ModelFetcher {
	return &ModelFetcher{}
}

// FetchModels returns available Google models, using a cached result if fresh enough.
func (f *ModelFetcher) FetchModels(ctx context.Context, apiKey string) ([]domainAgent.Model, error) {
	f.mu.RLock()
	if f.cachedModels != nil && time.Since(f.lastFetched) < modelCacheTTL {
		models := f.cachedModels
		f.mu.RUnlock()
		return models, nil
	}
	f.mu.RUnlock()

	return f.refreshModels(ctx, apiKey)
}

func (f *ModelFetcher) refreshModels(ctx context.Context, apiKey string) ([]domainAgent.Model, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// double-check after acquiring write lock
	if f.cachedModels != nil && time.Since(f.lastFetched) < modelCacheTTL {
		return f.cachedModels, nil
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	var models []domainAgent.Model
	for model, err := range client.Models.All(ctx) {
		if err != nil {
			return nil, fmt.Errorf("failed to list models: %w", err)
		}

		slug := strings.TrimPrefix(model.Name, "models/")

		if !isAgentCompatible(slug, model.SupportedActions) {
			continue
		}

		models = append(models, domainAgent.Model{
			Slug:     slug,
			Name:     model.DisplayName,
			Provider: domainAgent.ProviderGoogle,
		})
	}

	slices.Reverse(models)

	f.cachedModels = models
	f.lastFetched = time.Now()

	logger.Info().Int("count", len(models)).Msg("fetched available models from Google")

	return models, nil
}

// isAgentCompatible checks if a model supports generateContent and function calling.
// many models (TTS, image gen, Gemma, Lyria, robotics, etc.) support generateContent
// but not function calling, which the agent requires.
func isAgentCompatible(slug string, actions []string) bool {
	if !slices.Contains(actions, "generateContent") {
		return false
	}

	// only core gemini models
	if !strings.HasPrefix(slug, "gemini-") {
		return false
	}

	// exclude specialized model variants
	excludePatterns := []string{"-tts", "-image", "computer-use", "deep-research", "robotics", "customtools", "nano-banana"}
	for _, pattern := range excludePatterns {
		if strings.Contains(slug, pattern) {
			return false
		}
	}

	return true
}
