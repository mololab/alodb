# Agent Tools

Tools are functions that the LLM agent can call to interact with external systems.

## Available Tools

### read_schema

Reads the complete PostgreSQL database schema.

**Purpose**: Provides the agent with database structure information so it can generate accurate SQL queries.

**Input**: None (connection string comes from secure context)

**Output**:

```json
{
  "status": "success",
  "schema": {
    "database_name": "mydb",
    "tables": [...]
  },
  "message": "Schema loaded from cache."
}
```

**Caching**: Schema is cached in session state for performance:

- First request in session: reads from database
- Subsequent requests: returns cached schema
- Cache expires after configured TTL (default: 1 hour)

## Schema Caching

To avoid hitting the database on every request, the schema is cached in session state.

### How It Works

```
First Request (new session):
  └── read_schema called
  └── No cache found → query database
  └── Store schema + timestamp in session state
  └── Return schema

Subsequent Requests (same session):
  └── read_schema called
  └── Cache found, not expired → return cached schema
  └── No database query!
```

### Cache Configuration

Set `SCHEMA_CACHE_TTL` in your environment or `app.env`:

```env
SCHEMA_CACHE_TTL=1h    # 1 hour (default)
SCHEMA_CACHE_TTL=30m   # 30 minutes
SCHEMA_CACHE_TTL=24h   # 24 hours
```

### Cache Storage

The cache uses ADK session state:

- `cached_schema`: JSON-encoded database schema
- `schema_cached_at`: RFC3339 timestamp

Cache is automatically invalidated when:

- TTL expires
- New session is created
- Server restarts (in-memory sessions)

## Implementation

### File Structure

```
internal/infrastructure/agent/
├── tools.go                    # Tool creation and handler
├── cache/
│   └── schema_cache.go         # Schema caching logic
└── tools/
    └── schema_reader.go        # Database schema extraction
```

### Tool Handler Flow

```go
func schemaReaderHandler(toolCtx tool.Context, input Input) (Output, error) {
    // 1. Get connection string from context
    connStr := toolCtx.Value(connectionStringKey)

    // 2. Check cache
    schemaCache := cache.NewSchemaCache(ttl)
    if cached := schemaCache.Get(toolCtx); cached != nil {
        return cached  // Cache hit!
    }

    // 3. Cache miss - read from database
    result := tools.ReadSchemaFromDatabase(connStr)

    // 4. Store in cache for next time
    schemaCache.Set(toolCtx, result.Schema)

    return result
}
```

## Security

The connection string is **never exposed to the LLM**:

1. Client sends connection string in request
2. Server stores it in Go's `context.Context`
3. Tool reads it at execution time
4. LLM only sees the schema result

## Adding New Tools

1. Create implementation in `tools/`
2. Create wrapper in `tools.go`
3. Register in `createTools()` function
4. Update agent prompt in `prompts/agent_instruction.md`

## Planned Tools

| Tool              | Purpose                   | Status         |
| ----------------- | ------------------------- | -------------- |
| `read_schema`     | Read database schema      | ✅ Implemented |
| `query_executor`  | Execute read-only queries | 🔜 Planned     |
| `query_optimizer` | Analyze and optimize SQL  | 🔜 Planned     |
