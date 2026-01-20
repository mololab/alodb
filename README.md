# AloDB

AI-powered database assistant that helps users interact with PostgreSQL databases using natural language.

## Quick Start

```bash
# Setup
cp app.env.example app.env
# Edit app.env with SERVER_PORT

# Run
make run
```

## Usage

```bash
# Get available models
curl http://localhost:8080/v1/models

# Chat with agent (requires API key header)
curl -X POST http://localhost:8080/v1/agent/chat \
  -H "Content-Type: application/json" \
  -H "X-Gemini-Api-Key: your-api-key" \
  -d '{
    "message": "Show me all users with their orders",
    "connection_string": "postgres://user:pass@localhost:5432/mydb"
  }'
```

## Documentation

| Doc | Description |
|-----|-------------|
| [API](./docs/api/README.md) | REST endpoints, authentication, examples |
| [Architecture](./docs/architecture/README.md) | System design, DDD layers, request flow |
| [Agent](./docs/agent/README.md) | LLM agent, tools, prompt engineering |
| [Security](./docs/security/README.md) | API key handling, connection string protection |
| [Development](./docs/development/README.md) | Setup, building, contributing |

## Tech Stack

- **Go 1.21+** with [Google ADK](https://google.github.io/adk-docs/)
- **Gemini** LLM (API key via request header)
- **PostgreSQL** support
- **Domain-Driven Design** architecture

## Commands

| Command | Description |
|---------|-------------|
| `make run` | Run the server |
| `make build` | Build binary |
| `make test` | Run tests |
| `make tidy` | Install dependencies |
