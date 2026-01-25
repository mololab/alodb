package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	domainAgent "github.com/mololab/alodb/internal/domain/agent"
	"github.com/mololab/alodb/pkg/logger"

	"google.golang.org/adk/session"
)

type cachedAgent struct {
	agent    *DBAgent
	lastUsed time.Time
}

type Manager struct {
	agents         map[string]*cachedAgent
	mu             sync.RWMutex
	sessionService session.Service
	schemaCacheTTL time.Duration
}

func NewManager(schemaCacheTTL time.Duration) *Manager {
	return &Manager{
		agents:         make(map[string]*cachedAgent),
		sessionService: session.InMemoryService(),
		schemaCacheTTL: schemaCacheTTL,
	}
}

func (m *Manager) GetAgent(ctx context.Context, modelSlug string, apiKey string) (*DBAgent, error) {
	cacheKey := m.buildCacheKey(modelSlug, apiKey)

	m.mu.RLock()
	if cached, exists := m.agents[cacheKey]; exists {
		cached.lastUsed = time.Now()
		m.mu.RUnlock()
		return cached.agent, nil
	}
	m.mu.RUnlock()

	return m.createAgent(ctx, modelSlug, apiKey, cacheKey)
}

func (m *Manager) buildCacheKey(modelSlug, apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return fmt.Sprintf("%s:%s", modelSlug, hex.EncodeToString(hash[:8]))
}

func (m *Manager) createAgent(ctx context.Context, modelSlug, apiKey, cacheKey string) (*DBAgent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cached, exists := m.agents[cacheKey]; exists {
		cached.lastUsed = time.Now()
		return cached.agent, nil
	}

	model, ok := domainAgent.GetModelBySlug(modelSlug)
	if !ok {
		return nil, fmt.Errorf("unknown model: %s", modelSlug)
	}

	if apiKey == "" {
		providerCfg, _ := domainAgent.GetProviderByModel(model)
		return nil, fmt.Errorf("API key is required, provide via header: %s", providerCfg.HeaderKey)
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

	m.agents[cacheKey] = &cachedAgent{
		agent:    agent,
		lastUsed: time.Now(),
	}
	logger.Info().Str("model", modelSlug).Msg("agent initialized")

	return agent, nil
}

func (m *Manager) DeleteSession(ctx context.Context, sessionID string) error {
	err := m.sessionService.Delete(ctx, &session.DeleteRequest{
		AppName:   AppName,
		UserID:    sessionID,
		SessionID: sessionID,
	})
	if err != nil {
		logger.Warn().Err(err).Str("session_id", sessionID).Msg("failed to delete session")
		return err
	}

	logger.Info().Str("session_id", sessionID).Msg("session deleted")
	return nil
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for key, cached := range m.agents {
		if err := cached.agent.Close(); err != nil {
			logger.Error().Err(err).Str("key", key).Msg("error closing agent")
		}
	}

	m.agents = make(map[string]*cachedAgent)
	return nil
}
