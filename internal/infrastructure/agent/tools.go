package agent

import (
	"log"
	"time"

	"github.com/mololab/alodb/internal/infrastructure/agent/cache"
	"github.com/mololab/alodb/internal/infrastructure/agent/tools"

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
	log.Printf("[TOOL] read_schema called")

	// connection string from context
	connStr, ok := toolCtx.Value(connectionStringKey).(string)
	log.Printf("[TOOL] Connection string from context: ok=%v, empty=%v", ok, connStr == "")

	if !ok || connStr == "" {
		log.Printf("[TOOL] ERROR: No connection string in context")
		return tools.SchemaReaderOutput{
			Status:  "error",
			Message: "No database connection configured for this session.",
		}, nil
	}

	cacheTTL := getCacheTTL(toolCtx)
	schemaCache := cache.NewSchemaCache(cacheTTL)

	if cachedSchema := schemaCache.Get(toolCtx); cachedSchema != nil {
		log.Printf("[TOOL] Returning cached schema")
		return tools.SchemaReaderOutput{
			Status:  "success",
			Schema:  cachedSchema,
			Message: "Schema loaded from cache.",
		}, nil
	}

	// cache miss
	log.Printf("[TOOL] Cache miss, reading from database...")
	result, err := tools.ReadSchemaFromDatabase(connStr)
	if err != nil {
		return result, err
	}

	// cache the schema if successful
	if result.Status == "success" && result.Schema != nil {
		if err := schemaCache.Set(toolCtx, result.Schema); err != nil {
			log.Printf("[TOOL] Warning: Failed to cache schema: %v", err)
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
