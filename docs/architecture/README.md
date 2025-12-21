# Architecture

AloDB follows Domain-Driven Design (DDD) with clean separation of concerns.

## Layers

```
┌─────────────────────────────────────────────────────────────────┐
│                        PRESENTATION                              │
│                   (HTTP API / Handlers)                          │
├─────────────────────────────────────────────────────────────────┤
│                        APPLICATION                               │
│                    (Services / Use Cases)                        │
├─────────────────────────────────────────────────────────────────┤
│                          DOMAIN                                  │
│              (Business Logic / Entities / Types)                 │
├─────────────────────────────────────────────────────────────────┤
│                       INFRASTRUCTURE                             │
│         (ADK Agent / Database / External Services)               │
└─────────────────────────────────────────────────────────────────┘
```

### Domain Layer (`internal/domain/`)

Pure business objects with no external dependencies.

| Package | Purpose |
|---------|---------|
| `agent/types.go` | Chat request/response models |
| `database/types.go` | Database schema types |

### Application Layer (`internal/application/`)

Orchestrates domain objects and infrastructure.

| Package | Purpose |
|---------|---------|
| `agent/service.go` | Agent service - lifecycle and chat |

### Infrastructure Layer (`internal/infrastructure/`)

External systems and implementations.

| Package | Purpose |
|---------|---------|
| `agent/` | Google ADK agent |
| `config/` | Configuration (Viper) |
| `web/` | HTTP server, handlers |

## Project Structure

```
alodb/
├── cmd/
│   └── main.go                 # Entry point
├── docs/                       # Documentation
├── internal/
│   ├── application/
│   │   └── agent/
│   │       └── service.go
│   ├── domain/
│   │   ├── agent/
│   │   │   └── types.go
│   │   └── database/
│   │       └── types.go
│   └── infrastructure/
│       ├── agent/
│       │   ├── db_agent.go     # Agent constructor
│       │   ├── chat.go         # Chat handling
│       │   ├── events.go       # Event utilities
│       │   ├── tools.go        # Tool creation
│       │   ├── types.go
│       │   ├── response/
│       │   │   └── parser.go
│       │   └── tools/
│       │       └── schema_reader.go
│       ├── config/
│       │   └── config.go
│       └── web/
│           ├── server.go
│           ├── handlers/
│           │   └── agent_handler.go
│           └── dto/
│               └── agent.go
├── prompts/
│   └── agent_instruction.md
├── go.mod
└── makefile
```

## Request Flow

See [request-flow.md](./request-flow.md) for detailed flow.

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| DDD Layers | Test business logic without infrastructure |
| Context-based credentials | Connection strings never reach LLM |
| External prompts | Modify agent behavior without code changes |
| Modular agent package | Each file has single responsibility |
