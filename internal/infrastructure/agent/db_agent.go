package agent

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mololab/alodb/pkg/logger"

	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"
)

const (
	agentDescription    = "A database assistant that helps users understand their database schema and generate SQL queries."
	instructionFilePath = "prompts/agent_instruction.md"
)

type AgentParams struct {
	ModelSlug      string
	APIKey         string
	SchemaCacheTTL time.Duration
	SessionService session.Service
}

func NewDBAgent(ctx context.Context, params AgentParams) (*DBAgent, error) {
	if params.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	instruction, err := loadInstruction()
	if err != nil {
		return nil, fmt.Errorf("failed to load agent instruction: %w", err)
	}

	logger.Debug().
		Str("model", params.ModelSlug).
		Int("instruction_bytes", len(instruction)).
		Msg("creating agent")

	llmModel, err := createModel(ctx, params.ModelSlug, params.APIKey)
	if err != nil {
		return nil, err
	}

	tools, err := createTools()
	if err != nil {
		return nil, err
	}

	dbAgent, err := llmagent.New(llmagent.Config{
		Name:        AppName,
		Model:       llmModel,
		Description: agentDescription,
		Instruction: instruction,
		Tools:       tools,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	agentRunner, err := runner.New(runner.Config{
		AppName:        AppName,
		Agent:          dbAgent,
		SessionService: params.SessionService,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	return &DBAgent{
		agent:          dbAgent,
		runner:         agentRunner,
		sessionService: params.SessionService,
		modelSlug:      params.ModelSlug,
		schemaCacheTTL: params.SchemaCacheTTL,
	}, nil
}

func createModel(ctx context.Context, modelSlug string, apiKey string) (model.LLM, error) {
	return gemini.NewModel(ctx, modelSlug, &genai.ClientConfig{
		APIKey: apiKey,
	})
}

func loadInstruction() (string, error) {
	data, err := os.ReadFile(instructionFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read instruction file: %w", err)
	}
	return string(data), nil
}

func createTools() ([]tool.Tool, error) {
	schemaReaderTool, err := createSchemaReaderTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create schema reader tool: %w", err)
	}

	return []tool.Tool{
		schemaReaderTool,
	}, nil
}

func (a *DBAgent) Close() error {
	return nil
}

func (a *DBAgent) ModelSlug() string {
	return a.modelSlug
}

func (a *DBAgent) ensureSession(ctx context.Context, sessionID string) error {
	_, err := a.sessionService.Get(ctx, &session.GetRequest{
		AppName:   AppName,
		UserID:    sessionID,
		SessionID: sessionID,
	})
	if err == nil {
		return nil // session exists
	}

	_, err = a.sessionService.Create(ctx, &session.CreateRequest{
		AppName:   AppName,
		UserID:    sessionID,
		SessionID: sessionID,
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}
