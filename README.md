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

AloDB uses WebSocket for real-time streaming with **client-side query execution** - your database credentials never leave your machine.

```bash
# Get available models
curl http://localhost:8080/v1/models

# Connect via WebSocket (use wscat, websocat, or your app)
wscat -c 'ws://localhost:8080/v1/agent/stream?api_key=your-gemini-key'

# Then send chat messages:
# {"type": "chat", "payload": {"message": "Show me all users"}}
```

See [WebSocket API docs](./docs/api/websocket.md) for the full protocol.

## Documentation

| Doc | Description |
|-----|-------------|
| [API](./docs/api/README.md) | WebSocket API, endpoints, examples |
| [Agent](./docs/agent/README.md) | LLM agent, tools, prompt engineering |
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
