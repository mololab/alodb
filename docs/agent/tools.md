# Agent Tools

Tools are functions that the LLM agent can call to interact with external systems.

## Available Tools

### read_schema

Reads the complete PostgreSQL database schema via **client-side query execution**.

**Purpose**: Provides the agent with database structure information so it can generate accurate SQL queries.

**Input**: None (query executor comes from secure context)

**Output**:

```json
{
  "status": "success",
  "schema": {
    "database_name": "mydb",
    "tables": [...],
    "enums": [
      { "name": "order_status", "values": ["pending", "shipped", "delivered"] }
    ]
  },
  "message": "Schema loaded from cache."
}
```

**How It Works**:

The `read_schema` tool uses the `QueryExecutor` interface to send SQL queries to the client for execution. The server **never** connects directly to the database - all queries are executed by the client.

```
read_schema called
  └── Check cache
  └── If not cached:
      └── Send "Getting database name" query → Client executes → Returns result
      └── Send "Discovering tables" query → Client executes → Returns result
      └── Send "Reading enum types" query → Client executes → Returns result
      └── For each table:
          └── Send "Reading columns for X" → Client executes → Returns result
          └── Send "Finding primary key for X" → Client executes → Returns result
          └── Send "Finding foreign keys for X" → Client executes → Returns result
          └── Send "Reading indexes for X" → Client executes → Returns result
      └── Cache the complete schema
  └── Return schema to agent
```

**Caching**: Schema is cached in session state for performance:

- First request in session: sends queries to client
- Subsequent requests: returns cached schema
- Cache expires after configured TTL (default: 1 hour)

## Schema Queries

The tool sends predefined queries to the client for execution. Each query includes:

| Field         | Description                        |
| ------------- | ---------------------------------- |
| `name`        | Human-readable name for UI display |
| `description` | What the query does                |
| `query`       | The SQL to execute                 |
| `step`        | Current step number (1-based)      |
| `total_steps` | Total steps in the operation       |

**Queries sent**:

1. `SELECT current_database()` - Get database name
2. `SELECT table_name FROM information_schema.tables...` - List tables
3. `SELECT typname, array_agg(enumlabel...) FROM pg_type/pg_enum...` - List enum types with values
4. For each table:
   - Get columns with types, nullability, defaults, comments
   - Get primary key columns
   - Get foreign key relationships
   - Get non-primary indexes

## Schema Caching

To avoid re-querying the client on every request, the schema is cached in session state.

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
└── schema/
    └── reader.go               # Client-side schema extraction

internal/domain/websocket/
└── tools.go                    # Schema query definitions
```

### QueryExecutor Interface

```go
type QueryExecutor interface {
    ExecuteQuery(name, description, query string, step, totalSteps int, timeout time.Duration) (*QueryResultPayload, error)
}
```

The `ClientQueryExecutor` implementation sends queries via WebSocket to the client and waits for results.

## Security

Database credentials **never** reach the server:

1. Client connects to AloDB via WebSocket (only API key sent)
2. Client maintains its own database connection locally
3. When `read_schema` runs, it sends SQL queries to the client
4. Client executes queries locally and returns results
5. Server/LLM only sees the schema data, never credentials

## Adding New Tools

1. Create implementation in `internal/infrastructure/agent/` (new package or file)
2. Create tool function and handler in `internal/infrastructure/agent/tools.go`
3. Register in `createTools()` function
4. Update agent prompt in `prompts/agent_instruction.md`

## Planned Tools

| Tool              | Purpose                   | Status         |
| ----------------- | ------------------------- | -------------- |
| `read_schema`     | Read database schema      | ✅ Implemented |
| `query_optimizer` | Analyze and optimize SQL  | 🔜 Planned     |
