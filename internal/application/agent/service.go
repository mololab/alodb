package agent

import (
	"context"
	"fmt"

	domainAgent "github.com/mololab/alodb/internal/domain/agent"
	infraAgent "github.com/mololab/alodb/internal/infrastructure/agent"
)

type Service struct {
	config       domainAgent.AgentConfig
	manager      *infraAgent.Manager
	modelFetcher *infraAgent.ModelFetcher
}

func NewService(config domainAgent.AgentConfig) *Service {
	return &Service{
		config:       config,
		manager:      infraAgent.NewManager(config.SchemaCacheTTL),
		modelFetcher: infraAgent.NewModelFetcher(),
	}
}

func (s *Service) GetProviders() []domainAgent.ProviderInfo {
	return domainAgent.GetProviders()
}

func (s *Service) GetModels(ctx context.Context, provider domainAgent.Provider, apiKey string) ([]domainAgent.Model, error) {
	_, ok := domainAgent.ProviderRegistry[provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}

	switch provider {
	case domainAgent.ProviderGoogle:
		return s.modelFetcher.FetchModels(ctx, apiKey)
	default:
		return nil, fmt.Errorf("provider %s is not yet supported", provider)
	}
}

func (s *Service) GetManager() *infraAgent.Manager {
	return s.manager
}

func (s *Service) DeleteSession(ctx context.Context, sessionID string) error {
	return s.manager.DeleteSession(ctx, sessionID)
}

func (s *Service) Close() error {
	if s.manager != nil {
		return s.manager.Close()
	}
	return nil
}
