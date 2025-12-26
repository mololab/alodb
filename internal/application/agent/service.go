package agent

import (
	"context"
	"fmt"
	"sync"

	domainAgent "github.com/mololab/alodb/internal/domain/agent"
	infraAgent "github.com/mololab/alodb/internal/infrastructure/agent"
	"github.com/mololab/alodb/pkg/logger"
)

// Service handles agent operations
type Service struct {
	config domainAgent.AgentConfig
	agent  *infraAgent.DBAgent
	mu     sync.RWMutex
}

// NewService creates a new agent service
func NewService(config domainAgent.AgentConfig) *Service {
	return &Service{
		config: config,
	}
}

// Initialize initializes the agent
func (s *Service) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.agent != nil {
		return nil
	}

	logger.Info().Msg("initializing agent")
	agent, err := infraAgent.NewDBAgent(ctx, s.config)
	if err != nil {
		logger.Error().Err(err).Msg("failed to initialize agent")
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	s.agent = agent
	logger.Info().Msg("agent initialized successfully")
	return nil
}

// Chat sends a message to the agent and returns the response
func (s *Service) Chat(ctx context.Context, req domainAgent.ChatRequest) (*domainAgent.ChatResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.agent == nil {
		logger.Error().Msg("chat called before agent initialization")
		return nil, fmt.Errorf("agent not initialized")
	}

	logger.Debug().Str("session_id", req.SessionID).Msg("processing chat request")
	return s.agent.Chat(ctx, req)
}

// Close cleans up the agent resources
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.agent != nil {
		if err := s.agent.Close(); err != nil {
			return err
		}
		s.agent = nil
	}

	return nil
}
