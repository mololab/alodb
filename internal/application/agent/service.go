package agent

import (
	"context"
	"fmt"
	"sync"

	domainAgent "github.com/mololab/alodb/internal/domain/agent"
	infraAgent "github.com/mololab/alodb/internal/infrastructure/agent"
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

	agent, err := infraAgent.NewDBAgent(ctx, s.config)
	if err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	s.agent = agent
	return nil
}

// Chat sends a message to the agent and returns the response
func (s *Service) Chat(ctx context.Context, req domainAgent.ChatRequest) (*domainAgent.ChatResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.agent == nil {
		return nil, fmt.Errorf("agent not initialized")
	}

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
