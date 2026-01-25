package agent

import (
	"context"

	domainAgent "github.com/mololab/alodb/internal/domain/agent"
	infraAgent "github.com/mololab/alodb/internal/infrastructure/agent"
)

type Service struct {
	config  domainAgent.AgentConfig
	manager *infraAgent.Manager
}

func NewService(config domainAgent.AgentConfig) *Service {
	return &Service{
		config:  config,
		manager: infraAgent.NewManager(config.SchemaCacheTTL),
	}
}

func (s *Service) GetAllProviderModels() []domainAgent.ProviderModels {
	return domainAgent.GetAllProviderModels()
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
