package agent

import (
	"context"
	"fmt"

	domainAgent "github.com/mololab/alodb/internal/domain/agent"
	infraAgent "github.com/mololab/alodb/internal/infrastructure/agent"
	"github.com/mololab/alodb/pkg/logger"
)

type Service struct {
	config  domainAgent.AgentConfig
	manager *infraAgent.Manager
}

func NewService(config domainAgent.AgentConfig) *Service {
	return &Service{
		config:  config,
		manager: infraAgent.NewManager(config.Providers, config.SchemaCacheTTL),
	}
}

func (s *Service) Chat(ctx context.Context, req domainAgent.ChatRequest) (*domainAgent.ChatResponse, error) {
	modelSlug := req.Model
	if modelSlug == "" {
		modelSlug = domainAgent.GetDefaultModelSlug()
	}

	agent, err := s.manager.GetAgent(ctx, modelSlug)
	if err != nil {
		logger.Error().Err(err).Str("model", modelSlug).Msg("failed to get agent")
		return nil, fmt.Errorf("failed to get agent for model %s: %w", modelSlug, err)
	}

	logger.Debug().
		Str("session_id", req.SessionID).
		Str("model", modelSlug).
		Msg("processing chat request")

	return agent.Chat(ctx, req)
}

func (s *Service) GetAvailableModels() []domainAgent.Model {
	return s.manager.GetAvailableModels()
}

func (s *Service) Close() error {
	if s.manager != nil {
		return s.manager.Close()
	}
	return nil
}
