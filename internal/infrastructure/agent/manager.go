package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	domainAgent "github.com/mololab/alodb/internal/domain/agent"
	"github.com/mololab/alodb/pkg/logger"

	"google.golang.org/adk/session"
)

type Manager struct {
	agents         map[string]*DBAgent
	mu             sync.RWMutex
	sessionService session.Service
	providers      map[domainAgent.Provider]string
	schemaCacheTTL time.Duration
}

func NewManager(providers map[domainAgent.Provider]string, schemaCacheTTL time.Duration) *Manager {
	return &Manager{
		agents:         make(map[string]*DBAgent),
		sessionService: session.InMemoryService(),
		providers:      providers,
		schemaCacheTTL: schemaCacheTTL,
	}
}

func (m *Manager) GetAgent(ctx context.Context, modelSlug string) (*DBAgent, error) {
	m.mu.RLock()
	if agent, exists := m.agents[modelSlug]; exists {
		m.mu.RUnlock()
		return agent, nil
	}
	m.mu.RUnlock()

	return m.createAgent(ctx, modelSlug)
}

func (m *Manager) createAgent(ctx context.Context, modelSlug string) (*DBAgent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if agent, exists := m.agents[modelSlug]; exists {
		return agent, nil
	}

	model, ok := domainAgent.GetModelBySlug(modelSlug)
	if !ok {
		return nil, fmt.Errorf("unknown model: %s", modelSlug)
	}

	apiKey, ok := m.providers[model.Provider]
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("provider %s is not configured", model.Provider)
	}

	logger.Info().
		Str("model", modelSlug).
		Str("provider", string(model.Provider)).
		Msg("initializing agent")

	agent, err := NewDBAgent(ctx, AgentParams{
		ModelSlug:      modelSlug,
		APIKey:         apiKey,
		SchemaCacheTTL: m.schemaCacheTTL,
		SessionService: m.sessionService,
	})
	if err != nil {
		return nil, err
	}

	m.agents[modelSlug] = agent
	logger.Info().Str("model", modelSlug).Msg("agent initialized")

	return agent, nil
}

func (m *Manager) GetAvailableModels() []domainAgent.Model {
	var models []domainAgent.Model

	for provider, apiKey := range m.providers {
		if apiKey == "" {
			continue
		}

		cfg, ok := domainAgent.ProviderRegistry[provider]
		if !ok {
			continue
		}

		models = append(models, cfg.Models...)
	}

	return models
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for slug, agent := range m.agents {
		if err := agent.Close(); err != nil {
			logger.Error().Err(err).Str("model", slug).Msg("error closing agent")
		}
	}

	m.agents = make(map[string]*DBAgent)
	return nil
}
