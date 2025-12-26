# Development Guide

## Prerequisites

- Go 1.21 or later
- PostgreSQL (for testing)
- Google API Key (for Gemini)

## Setup

```bash
# Clone
git clone <repo>
cd alodb

# Install dependencies
make tidy

# Configure
cp app.env.example app.env
# Edit app.env with GOOGLE_API_KEY and SERVER_PORT

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

| Variable         | Description                                      | Required | Default      |
| ---------------- | ------------------------------------------------ | -------- | ------------ |
| `GOOGLE_API_KEY` | Gemini API key                                   | Yes      | -            |
| `SERVER_PORT`    | HTTP server port                                 | Yes      | -            |
| `SERVER_ENV`     | Environment mode (`development` or `production`) | No       | `production` |

## Testing

```bash
# Run tests
make test

# With coverage
go test -cover ./...

# Test API
curl http://localhost:8080/v1/health
```

## Contributing

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Add comments for exported functions

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

**Development mode** (`SERVER_ENV=development`): Pretty console output with colors

```
12:34:56 INF server starting port=8080
12:34:57 DBG processing chat request message="Show me users" has_connection=true
12:34:57 DBG read_schema tool called
12:34:58 DBG agent event event=1 author=model func_call=true func=read_schema
12:34:58 INF schema extracted tables=5
12:34:58 DBG agent completed event_count=3
```

**Production mode** (`SERVER_ENV=production`): JSON structured logging

```json
{
  "level": "info",
  "port": "8080",
  "time": 1234567890,
  "message": "server starting"
}
```

### Common Issues

| Issue                    | Cause                     | Solution                 |
| ------------------------ | ------------------------- | ------------------------ |
| "No response generated"  | LLM didn't produce text   | Check API key and prompt |
| "Connection refused"     | Database not running      | Start PostgreSQL         |
| "No database connection" | Missing connection string | Include in request       |
