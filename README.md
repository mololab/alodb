# AloDB

AI-powered database assistant that helps users interact with PostgreSQL databases using natural language.

## Quick Start

```bash
# Setup
cp app.env.example app.env
# Edit app.env with your GOOGLE_API_KEY and SERVER_PORT

# Run
make run
```

## Usage

```bash
curl -X POST http://localhost:8080/v1/agent/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Show me all users with their orders",
    "connection_string": "postgres://user:pass@localhost:5432/mydb"
  }'
```

## Documentation

| Doc | Description |
|-----|-------------|
| [Architecture](./docs/architecture/README.md) | System design, DDD layers, request flow |
| [Agent](./docs/agent/README.md) | LLM agent, tools, prompt engineering |
| [API](./docs/api/README.md) | REST endpoints and examples |
| [Security](./docs/security/README.md) | Connection string protection |
| [Development](./docs/development/README.md) | Setup, building, contributing |

## Tech Stack

- **Go 1.21+** with [Google ADK](https://google.github.io/adk-docs/)
- **Gemini 2.0 Flash** LLM
- **PostgreSQL** support
- **Domain-Driven Design** architecture

## Commands

| Command | Description |
|---------|-------------|
| `make run` | Run the server |
| `make build` | Build binary |
| `make test` | Run tests |
| `make tidy` | Install dependencies |
