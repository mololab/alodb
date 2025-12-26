package cache

import (
	"encoding/json"
	"time"

	"github.com/mololab/alodb/internal/domain/database"
	"github.com/mololab/alodb/pkg/logger"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/tool"
)

// Session state keys for schema caching
const (
	SchemaStateKey    = "cached_schema"
	SchemaCachedAtKey = "schema_cached_at"
)

// SchemaCache handles caching of database schemas in session state
type SchemaCache struct {
	ttl time.Duration
}

// NewSchemaCache creates a new schema cache with the given TTL
func NewSchemaCache(ttl time.Duration) *SchemaCache {
	return &SchemaCache{ttl: ttl}
}

// Get retrieves the cached schema if it exists and is not expired
func (c *SchemaCache) Get(toolCtx tool.Context) *database.DatabaseSchema {
	callbackCtx, ok := toolCtx.(agent.CallbackContext)
	if !ok {
		return nil
	}

	state := callbackCtx.State()

	schemaJSON, err := state.Get(SchemaStateKey)
	if err != nil || schemaJSON == nil {
		return nil
	}

	schemaStr, ok := schemaJSON.(string)
	if !ok || schemaStr == "" {
		return nil
	}

	cachedAtVal, err := state.Get(SchemaCachedAtKey)
	if err != nil || cachedAtVal == nil {
		return nil
	}

	cachedAtStr, ok := cachedAtVal.(string)
	if !ok || cachedAtStr == "" {
		return nil
	}

	cachedAt, err := time.Parse(time.RFC3339, cachedAtStr)
	if err != nil {
		return nil
	}

	if time.Since(cachedAt) > c.ttl {
		logger.Debug().Dur("age", time.Since(cachedAt)).Dur("ttl", c.ttl).Msg("cache expired")
		return nil
	}

	var schema database.DatabaseSchema
	if err := json.Unmarshal([]byte(schemaStr), &schema); err != nil {
		logger.Warn().Err(err).Msg("failed to unmarshal cached schema")
		return nil
	}

	logger.Debug().Dur("age", time.Since(cachedAt).Round(time.Second)).Msg("cache hit")
	return &schema
}

// Set stores the schema in session state with current timestamp
func (c *SchemaCache) Set(toolCtx tool.Context, schema *database.DatabaseSchema) error {
	callbackCtx, ok := toolCtx.(agent.CallbackContext)
	if !ok {
		return nil
	}

	state := callbackCtx.State()

	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return err
	}

	if err := state.Set(SchemaStateKey, string(schemaJSON)); err != nil {
		return err
	}

	if err := state.Set(SchemaCachedAtKey, time.Now().Format(time.RFC3339)); err != nil {
		return err
	}

	logger.Debug().Int("tables", len(schema.Tables)).Dur("ttl", c.ttl).Msg("schema cached")
	return nil
}

// TTL returns the cache TTL duration
func (c *SchemaCache) TTL() time.Duration {
	return c.ttl
}
