# Development Guide

## Prerequisites

- Go 1.21 or later
- PostgreSQL (for testing)

## Setup

```bash
# Clone
git clone <repo>
cd alodb

# Install dependencies
make tidy

# Configure
cp app.env.example app.env
# Edit app.env with SERVER_PORT

# Run
make run
```

## Commands

| Command      | Description                 |
| ------------ | --------------------------- |
| `make run`   | Run the application         |
| `make build` | Build binary to `bin/alodb` |
| `make tidy`  | Download dependencies       |
| `make test`  | Run tests                   |
| `make clean` | Remove build artifacts      |

## Environment Variables

| Variable           | Description                                      | Required | Default      |
| ------------------ | ------------------------------------------------ | -------- | ------------ |
| `SERVER_PORT`      | HTTP server port                                 | Yes      | -            |
| `SERVER_ENV`       | Environment mode (`development` or `production`) | No       | `production` |
| `SCHEMA_CACHE_TTL` | Schema cache duration (e.g., `1h`, `30m`)        | No       | `1h`         |

API keys are provided by clients via request headers, not environment variables.

## Testing

```bash
# Run tests
make test

# With coverage
go test -cover ./...

# Test API (requires API key header)
curl http://localhost:8080/v1/health

curl -X POST http://localhost:8080/v1/agent/chat \
  -H "Content-Type: application/json" \
  -H "X-Gemini-Api-Key: your-api-key" \
  -d '{"message": "Show tables", "connection_string": "postgres://..."}'
```

## Contributing

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting

### Adding a New Provider

1. Add provider constant in `internal/domain/agent/models.go`
2. Add models and header key to `ProviderRegistry`
3. Implement model creation in `internal/infrastructure/agent/db_agent.go`

### Adding a New Tool

1. Create implementation in `internal/infrastructure/agent/tools/`
2. Create wrapper in `internal/infrastructure/agent/tools.go`
3. Register in `createTools()` function
4. Update `prompts/agent_instruction.md`

### Adding a New Endpoint

1. Define DTO in `internal/infrastructure/web/dto/`
2. Create handler in `internal/infrastructure/web/handlers/`
3. Register route in `internal/infrastructure/web/server.go`
4. Update [API docs](../api/README.md)

## Debugging

### Logging

The application uses [zerolog](https://github.com/rs/zerolog) for structured logging.

**Development mode** (`SERVER_ENV=development`): Pretty console output

```
12:34:56 INF server starting port=8080
12:34:57 DBG processing chat request session_id=abc123 model=gemini-2.5-pro
12:34:58 INF initializing agent model=gemini-2.5-pro provider=google
```

**Production mode** (`SERVER_ENV=production`): JSON structured logging

```json
{"level":"info","port":"8080","time":1234567890,"message":"server starting"}
```

### Common Issues

| Issue                    | Cause                     | Solution                       |
| ------------------------ | ------------------------- | ------------------------------ |
| "missing required header"| API key not in header     | Add `X-Gemini-Api-Key` header  |
| "No response generated"  | LLM didn't produce text   | Check API key validity         |
| "Connection refused"     | Database not running      | Start PostgreSQL               |
| "No database connection" | Missing connection string | Include in request body        |
