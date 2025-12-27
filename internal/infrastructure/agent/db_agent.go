package agent

import (
	"context"
	"fmt"
	"os"
	"time"

	domainAgent "github.com/mololab/alodb/internal/domain/agent"
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
	agentName           = "alodb_agent"
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
	modelInfo, ok := domainAgent.GetModelBySlug(params.ModelSlug)
	if !ok {
		return nil, fmt.Errorf("unknown model: %s", params.ModelSlug)
	}

	if params.APIKey == "" {
		return nil, fmt.Errorf("API key is required for provider: %s", modelInfo.Provider)
	}

	instruction, err := loadInstruction()
	if err != nil {
		return nil, fmt.Errorf("failed to load agent instruction: %w", err)
	}

	logger.Debug().
		Str("model", params.ModelSlug).
		Str("provider", string(modelInfo.Provider)).
		Int("instruction_bytes", len(instruction)).
		Msg("creating agent")

	llmModel, err := createModel(ctx, modelInfo, params.APIKey)
	if err != nil {
		return nil, err
	}

	tools, err := createTools()
	if err != nil {
		return nil, err
	}

	dbAgent, err := llmagent.New(llmagent.Config{
		Name:        agentName,
		Model:       llmModel,
		Description: agentDescription,
		Instruction: instruction,
		Tools:       tools,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	agentRunner, err := runner.New(runner.Config{
		AppName:        agentName,
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

func createModel(ctx context.Context, m domainAgent.Model, apiKey string) (model.LLM, error) {
	switch m.Provider {
	case domainAgent.ProviderGoogle:
		return gemini.NewModel(ctx, m.Slug, &genai.ClientConfig{
			APIKey: apiKey,
		})
	case domainAgent.ProviderOpenAI:
		return nil, fmt.Errorf("OpenAI provider not yet supported")
	default:
		return nil, fmt.Errorf("unsupported provider: %s", m.Provider)
	}
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
