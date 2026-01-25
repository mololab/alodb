package agent

import (
	"time"

	"github.com/mololab/alodb/internal/domain/database"
	"github.com/mololab/alodb/internal/infrastructure/agent/cache"
	"github.com/mololab/alodb/internal/infrastructure/agent/schema"
	"github.com/mololab/alodb/pkg/logger"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

const defaultSchemaCacheTTL = 1 * time.Hour

type schemaReaderInput struct{}

type schemaReaderOutput struct {
	Status  string                   `json:"status"`
	Schema  *database.DatabaseSchema `json:"schema,omitempty"`
	Message string                   `json:"message,omitempty"`
}

func createSchemaReaderTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "read_schema",
			Description: "Reads and returns the complete database schema including all tables, columns, primary keys, foreign keys, and indexes. The database connection is already configured. Just call this tool to get the schema.",
		},
		schemaReaderHandler,
	)
}

func schemaReaderHandler(toolCtx tool.Context, input schemaReaderInput) (schemaReaderOutput, error) {
	logger.Debug().Msg("read_schema tool called")

	cacheTTL := getCacheTTL(toolCtx)
	schemaCache := cache.NewSchemaCache(cacheTTL)
	if cachedSchema := schemaCache.Get(toolCtx); cachedSchema != nil {
		logger.Debug().Msg("returning cached schema")
		return schemaReaderOutput{
			Status:  "success",
			Schema:  cachedSchema,
			Message: "Schema loaded from cache.",
		}, nil
	}

	executor, ok := toolCtx.Value(queryExecutorKey).(schema.QueryExecutor)
	if !ok {
		logger.Error().Msg("no query executor in context")
		return schemaReaderOutput{
			Status:  "error",
			Message: "No database connection available. Please connect via WebSocket.",
		}, nil
	}

	reader := schema.NewReader(executor)
	dbSchema, err := reader.ReadSchema()
	if err != nil {
		logger.Error().Err(err).Msg("schema extraction failed")
		return schemaReaderOutput{
			Status:  "error",
			Message: "Failed to read schema: " + err.Error(),
		}, nil
	}

	if err := schemaCache.Set(toolCtx, dbSchema); err != nil {
		logger.Warn().Err(err).Msg("failed to cache schema")
	}

	logger.Info().Int("tables", len(dbSchema.Tables)).Msg("schema extracted via client")
	return schemaReaderOutput{
		Status: "success",
		Schema: dbSchema,
	}, nil
}

func getCacheTTL(toolCtx tool.Context) time.Duration {
	if ttl, ok := toolCtx.Value(schemaCacheTTLKey).(time.Duration); ok && ttl > 0 {
		return ttl
	}
	return defaultSchemaCacheTTL
}
