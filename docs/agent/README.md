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
│                    └────────┬────────┘                           │
│                             │                                    │
│                             ▼                                    │
│                    ┌─────────────────┐                           │
│                    │ QueryExecutor   │ ◄─── WebSocket to Client  │
│                    │ (client-side)   │                           │
│                    └─────────────────┘                           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Components

| Component         | Package                                   | Purpose                            |
| ----------------- | ----------------------------------------- | ---------------------------------- |
| **Manager**       | `internal/infrastructure/agent`           | Agent caching and lifecycle        |
| **Model**         | `google.golang.org/adk/model/gemini`      | Gemini LLM interface               |
| **LLMAgent**      | `google.golang.org/adk/agent/llmagent`    | Agent with instructions and tools  |
| **Runner**        | `google.golang.org/adk/runner`            | Executes agent, manages sessions   |
| **Session**       | `google.golang.org/adk/session`           | In-memory session storage          |
| **FunctionTool**  | `google.golang.org/adk/tool/functiontool` | Wraps Go functions as LLM tools    |
| **QueryExecutor** | `internal/infrastructure/agent/schema`    | Interface for client-side queries  |
| **SchemaReader**  | `internal/infrastructure/agent/schema`    | Extracts schema via query executor |

## Agent Package Structure

```
internal/infrastructure/agent/
├── manager.go           # Agent caching by model+apiKeyHash
├── db_agent.go          # Agent constructor
├── model_fetcher.go     # Dynamic model discovery from provider APIs
├── stream_agent.go      # Streaming chat with callbacks
├── events.go            # Event processing
├── types.go             # Structs and context keys
├── constants.go         # Application constants
├── tools.go             # Tool creation and handlers
├── cache/
│   └── schema_cache.go  # Schema caching
├── response/
│   └── parser.go        # JSON response parsing
└── schema/
    └── reader.go        # Client-side schema extraction
```

## Multi-Model Support

Available models are **fetched dynamically** from the provider's API at runtime - no hardcoded model lists. Provider metadata (name, header key) is defined in `internal/domain/agent/models.go`, while model discovery is handled by `internal/infrastructure/agent/model_fetcher.go`.

```go
// Provider metadata only - models are fetched from Google's API
var ProviderRegistry = map[Provider]ProviderConfig{
    ProviderGoogle: {
        HeaderKey: "X-Gemini-Api-Key",
        Metadata:  ProviderMetadata{Name: "Google Gemini", ...},
    },
}
```

The `ModelFetcher` uses the `google.golang.org/genai` SDK's `client.Models.All()` to list all models, filters for those supporting `generateContent`, and caches results for 1 hour.

Adding a new provider requires:
1. Add provider constant to `ProviderRegistry` with header key
2. Implement model fetching in `model_fetcher.go` (or a new fetcher)
3. Add the provider case in `application/agent/service.go` `GetModels()`
4. Implement model creation in `db_agent.go`

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
