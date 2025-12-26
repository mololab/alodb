package agent

import (
	"time"

	"github.com/mololab/alodb/internal/infrastructure/agent/cache"
	"github.com/mololab/alodb/internal/infrastructure/agent/tools"
	"github.com/mololab/alodb/pkg/logger"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// default cache TTL
const defaultSchemaCacheTTL = 1 * time.Hour

// createSchemaReaderTool creates the schema reader tool for the agent
func createSchemaReaderTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "read_schema",
			Description: "Reads and returns the complete database schema including all tables, columns, primary keys, foreign keys, and indexes. The database connection is already configured. Just call this tool to get the schema.",
		},
		schemaReaderHandler,
	)
}

// schemaReaderHandler handles the schema reader tool invocation
func schemaReaderHandler(toolCtx tool.Context, input tools.SchemaReaderInput) (tools.SchemaReaderOutput, error) {
	logger.Debug().Msg("read_schema tool called")

	connStr, ok := toolCtx.Value(connectionStringKey).(string)
	if !ok || connStr == "" {
		logger.Warn().Msg("no connection string in context")
		return tools.SchemaReaderOutput{
			Status:  "error",
			Message: "No database connection configured for this session.",
		}, nil
	}

	cacheTTL := getCacheTTL(toolCtx)
	schemaCache := cache.NewSchemaCache(cacheTTL)

	if cachedSchema := schemaCache.Get(toolCtx); cachedSchema != nil {
		logger.Debug().Msg("returning cached schema")
		return tools.SchemaReaderOutput{
			Status:  "success",
			Schema:  cachedSchema,
			Message: "Schema loaded from cache.",
		}, nil
	}

	logger.Debug().Msg("cache miss, reading from database")
	result, err := tools.ReadSchemaFromDatabase(connStr)
	if err != nil {
		return result, err
	}

	if result.Status == "success" && result.Schema != nil {
		if err := schemaCache.Set(toolCtx, result.Schema); err != nil {
			logger.Warn().Err(err).Msg("failed to cache schema")
		}
	}

	return result, nil
}

// getCacheTTL extracts cache TTL from context or returns default
func getCacheTTL(toolCtx tool.Context) time.Duration {
	if ttl, ok := toolCtx.Value(schemaCacheTTLKey).(time.Duration); ok && ttl > 0 {
		return ttl
	}
	return defaultSchemaCacheTTL
}
