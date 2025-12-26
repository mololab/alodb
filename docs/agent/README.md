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
│                        DBAgent                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│   │   Gemini    │    │   Runner    │    │   Session   │         │
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
│                    │  (FunctionTool) │                           │
│                    └─────────────────┘                           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Components

| Component        | Package                                   | Purpose                           |
| ---------------- | ----------------------------------------- | --------------------------------- |
| **Model**        | `google.golang.org/adk/model/gemini`      | Gemini LLM interface              |
| **LLMAgent**     | `google.golang.org/adk/agent/llmagent`    | Agent with instructions and tools |
| **Runner**       | `google.golang.org/adk/runner`            | Executes agent, manages sessions  |
| **Session**      | `google.golang.org/adk/session`           | In-memory session storage         |
| **FunctionTool** | `google.golang.org/adk/tool/functiontool` | Wraps Go functions as LLM tools   |

## Agent Package Structure

```
internal/infrastructure/agent/
├── db_agent.go          # Agent constructor and initialization
├── chat.go              # Chat execution and event handling
├── events.go            # Event processing utilities
├── types.go             # DBAgent struct and context keys
├── tools.go             # Tool creation functions
├── response/
│   └── parser.go        # JSON response parsing
└── tools/
    └── schema_reader.go # Schema reader implementation
```

## Event Handling

The agent produces events during execution:

| Event Type       | Description                   |
| ---------------- | ----------------------------- |
| FunctionCall     | Agent decides to call a tool  |
| FunctionResponse | Tool returns result to agent  |
| Text (Model)     | Agent generates text response |

**Critical**: We only capture the **last model text response** (after all tools complete).

```go
// We iterate through ALL events but only keep the last model response
for event, err := range events {
    if event.Content.Role == "model" {
        text := ExtractTextFromEvent(event)
        if text != "" {
            lastModelResponse = text  // Keep overwriting until the last one
        }
    }
}
```

## Configuration

The agent is configured via:

| Config           | Source                         | Description               |
| ---------------- | ------------------------------ | ------------------------- |
| `GOOGLE_API_KEY` | Environment                    | Gemini API authentication |
| `ModelName`      | Code default                   | `gemini-2.0-flash`        |
| Instruction      | `prompts/agent_instruction.md` | System prompt             |

## Further Reading

- [Tools Documentation](./tools.md) - Available tools and how they work
- [Prompts Documentation](./prompts.md) - Prompt engineering guidelines
