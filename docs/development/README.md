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

| Command | Description |
|---------|-------------|
| `make run` | Run the application |
| `make build` | Build binary to `bin/alodb` |
| `make tidy` | Download dependencies |
| `make test` | Run tests |
| `make clean` | Remove build artifacts |

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `GOOGLE_API_KEY` | Gemini API key | Yes |
| `SERVER_PORT` | HTTP server port | Yes |

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

Log output shows the request flow:

```
[CHAT] Processing request: message=Show me users
[CHAT] Running agent...
[TOOL] read_schema called
[EVENT #1] author=model hasFuncCall=true
[EVENT #2] author=tool hasFuncResp=true
[EVENT #3] author=model hasText=true
[CHAT] Agent completed with 3 events
```

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| "No response generated" | LLM didn't produce text | Check API key and prompt |
| "Connection refused" | Database not running | Start PostgreSQL |
| "No database connection" | Missing connection string | Include in request |
