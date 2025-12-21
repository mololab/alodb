package cache

import (
	"encoding/json"
	"log"
	"time"

	"github.com/mololab/alodb/internal/domain/database"
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
// returns nil if no valid cache exists
func (c *SchemaCache) Get(toolCtx tool.Context) *database.DatabaseSchema {
	callbackCtx, ok := toolCtx.(agent.CallbackContext)
	if !ok {
		log.Printf("[CACHE] Cannot access state: tool context is not CallbackContext")
		return nil
	}

	state := callbackCtx.State()

	schemaJSON, err := state.Get(SchemaStateKey)
	if err != nil || schemaJSON == nil {
		log.Printf("[CACHE] No cached schema found")
		return nil
	}

	schemaStr, ok := schemaJSON.(string)
	if !ok || schemaStr == "" {
		log.Printf("[CACHE] Cached schema is not a string")
		return nil
	}

	cachedAtVal, err := state.Get(SchemaCachedAtKey)
	if err != nil || cachedAtVal == nil {
		log.Printf("[CACHE] No cache timestamp found")
		return nil
	}

	cachedAtStr, ok := cachedAtVal.(string)
	if !ok || cachedAtStr == "" {
		log.Printf("[CACHE] Cache timestamp is not a string")
		return nil
	}

	cachedAt, err := time.Parse(time.RFC3339, cachedAtStr)
	if err != nil {
		log.Printf("[CACHE] Invalid cache timestamp: %v", err)
		return nil
	}

	if time.Since(cachedAt) > c.ttl {
		log.Printf("[CACHE] Cache expired (age: %v, ttl: %v)", time.Since(cachedAt), c.ttl)
		return nil
	}

	var schema database.DatabaseSchema
	if err := json.Unmarshal([]byte(schemaStr), &schema); err != nil {
		log.Printf("[CACHE] Failed to unmarshal cached schema: %v", err)
		return nil
	}

	log.Printf("[CACHE] Cache hit! Schema cached %v ago (ttl: %v)", time.Since(cachedAt).Round(time.Second), c.ttl)
	return &schema
}

// Set stores the schema in session state with current timestamp
func (c *SchemaCache) Set(toolCtx tool.Context, schema *database.DatabaseSchema) error {
	callbackCtx, ok := toolCtx.(agent.CallbackContext)
	if !ok {
		log.Printf("[CACHE] Cannot access state: tool context is not CallbackContext")
		return nil // no error, just skip caching
	}

	state := callbackCtx.State()

	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return err
	}

	if err := state.Set(SchemaStateKey, string(schemaJSON)); err != nil {
		log.Printf("[CACHE] Failed to set schema in state: %v", err)
		return err
	}

	if err := state.Set(SchemaCachedAtKey, time.Now().Format(time.RFC3339)); err != nil {
		log.Printf("[CACHE] Failed to set timestamp in state: %v", err)
		return err
	}

	log.Printf("[CACHE] Schema cached (tables: %d, ttl: %v)", len(schema.Tables), c.ttl)
	return nil
}

// TTL returns the cache TTL duration
func (c *SchemaCache) TTL() time.Duration {
	return c.ttl
}
