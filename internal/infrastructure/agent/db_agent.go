package agent

import (
	"context"
	"fmt"
	"log"
	"os"

	domainAgent "github.com/mololab/alodb/internal/domain/agent"

	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"
)

// Agent configuration constants
const (
	agentName           = "alodb_agent"
	agentDescription    = "A database assistant that helps users understand their database schema and generate SQL queries."
	defaultModelName    = "gemini-2.0-flash"
	instructionFilePath = "prompts/agent_instruction.md"
)

// NewDBAgent creates a new database agent
func NewDBAgent(ctx context.Context, config domainAgent.AgentConfig) (*DBAgent, error) {
	if config.GoogleAPIKey == "" {
		return nil, fmt.Errorf("google API key is required")
	}

	if config.ModelName == "" {
		config.ModelName = defaultModelName
	}

	instruction, err := loadInstruction()
	if err != nil {
		return nil, fmt.Errorf("failed to load agent instruction: %w", err)
	}
	log.Printf("[AGENT] Loaded instruction (%d bytes)", len(instruction))
	log.Printf("[AGENT] Schema cache TTL: %v", config.SchemaCacheTTL)

	model, err := gemini.NewModel(ctx, config.ModelName, &genai.ClientConfig{
		APIKey: config.GoogleAPIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini model: %w", err)
	}

	tools, err := createTools()
	if err != nil {
		return nil, err
	}

	dbAgent, err := llmagent.New(llmagent.Config{
		Name:        agentName,
		Model:       model,
		Description: agentDescription,
		Instruction: instruction,
		Tools:       tools,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	sessionService := session.InMemoryService()

	agentRunner, err := runner.New(runner.Config{
		AppName:        agentName,
		Agent:          dbAgent,
		SessionService: sessionService,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	return &DBAgent{
		agent:          dbAgent,
		runner:         agentRunner,
		sessionService: sessionService,
		config:         config,
		schemaCacheTTL: config.SchemaCacheTTL,
	}, nil
}

// loadInstruction loads the agent instruction from the prompts file
func loadInstruction() (string, error) {
	data, err := os.ReadFile(instructionFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read instruction file: %w", err)
	}
	return string(data), nil
}

// createTools creates all tools for the agent
func createTools() ([]tool.Tool, error) {
	schemaReaderTool, err := createSchemaReaderTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create schema reader tool: %w", err)
	}

	return []tool.Tool{
		schemaReaderTool,
	}, nil
}

// Close cleans up the agent resources
func (a *DBAgent) Close() error {
	return nil
}
