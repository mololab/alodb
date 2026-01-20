# Agent Documentation

AloDB uses Google's Agent Development Kit (ADK) to power its AI capabilities.

## Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Tools](./tools.md)
- [Prompts](./prompts.md)

## Overview

The agent is an LLM-powered assistant that:

1. Understands natural language database queries
2. Reads database schema using tools
3. Generates accurate SQL queries
4. Returns structured JSON responses

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Manager                                   │
│              (Caches agents by model + apiKeyHash)               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│   │   LLM       │    │   Runner    │    │   Session   │         │
│   │   Model     │    │             │    │   Service   │         │
│   └─────────────┘    └─────────────┘    └─────────────┘         │
│          │                  │                  │                 │
│          └──────────────────┼──────────────────┘                 │
│                             │                                    │
│                             ▼                                    │
│                    ┌─────────────────┐                           │
│                    │    LLMAgent     │                           │
│                    │  (alodb_agent)  │                           │
│                    └─────────────────┘                           │
│                             │                                    │
│                             │ Tools                              │
│                             ▼                                    │
│                    ┌─────────────────┐                           │
│                    │  read_schema    │                           │
│                    └─────────────────┘                           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Components

| Component        | Package                                   | Purpose                           |
| ---------------- | ----------------------------------------- | --------------------------------- |
| **Manager**      | `internal/infrastructure/agent`           | Agent caching and lifecycle       |
| **Model**        | `google.golang.org/adk/model/gemini`      | Gemini LLM interface              |
| **LLMAgent**     | `google.golang.org/adk/agent/llmagent`    | Agent with instructions and tools |
| **Runner**       | `google.golang.org/adk/runner`            | Executes agent, manages sessions  |
| **Session**      | `google.golang.org/adk/session`           | In-memory session storage         |
| **FunctionTool** | `google.golang.org/adk/tool/functiontool` | Wraps Go functions as LLM tools   |

## Agent Package Structure

```
internal/infrastructure/agent/
├── manager.go           # Agent caching by model+apiKeyHash
├── db_agent.go          # Agent constructor
├── chat.go              # Chat execution
├── events.go            # Event processing
├── types.go             # Structs and context keys
├── tools.go             # Tool creation
├── cache/
│   └── schema_cache.go  # Schema caching
├── response/
│   └── parser.go        # JSON response parsing
└── tools/
    └── schema_reader.go # Schema reader implementation
```

## Multi-Model Support

Models are defined in `internal/domain/agent/models.go`:

```go
var ProviderRegistry = map[Provider]ProviderConfig{
    ProviderGoogle: {
        HeaderKey: "X-Gemini-Api-Key",
        Models:    GoogleModels,
    },
    ProviderOpenAI: {
        HeaderKey: "X-Openai-Api-Key",
        Models:    OpenAIModels,
    },
}
```

Adding a new provider requires:
1. Add provider constant and models
2. Add to `ProviderRegistry` with header key
3. Implement model creation in `db_agent.go`

## Event Handling

The agent produces events during execution:

| Event Type       | Description                   |
| ---------------- | ----------------------------- |
| FunctionCall     | Agent decides to call a tool  |
| FunctionResponse | Tool returns result to agent  |
| Text (Model)     | Agent generates text response |

We only capture the **last model text response** (after all tools complete).

## Configuration

| Config             | Source                         | Description           |
| ------------------ | ------------------------------ | --------------------- |
| API Key            | Request header                 | Per-request auth      |
| Model              | Request body or default        | Which model to use    |
| Instruction        | `prompts/agent_instruction.md` | System prompt         |
| Schema Cache TTL   | Environment                    | How long to cache     |

## Further Reading

- [Tools Documentation](./tools.md)
- [Prompts Documentation](./prompts.md)
