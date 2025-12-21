# Request Flow

This document describes how a request flows through the AloDB system.

## High-Level Flow

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  Client  │────▶│  Handler │────▶│  Service │────▶│ DBAgent  │────▶│   LLM    │
└──────────┘     └──────────┘     └──────────┘     └──────────┘     └──────────┘
     │                │                │                │                │
     │   HTTP POST    │    DTO to      │   Domain to    │   Run Agent    │
     │   /v1/agent/   │    Domain      │   Agent Call   │   with Tools   │
     │   chat         │                │                │                │
```

## Detailed Steps

### 1. Client Request

```json
{
  "message": "Show me all token activity of user with email test@example.com",
  "connection_string": "postgres://user:pass@localhost:5432/mydb"
}
```

### 2. Handler Layer (`web/handlers/agent_handler.go`)

- Receives HTTP POST request
- Validates JSON body
- Converts DTO to domain object
- Passes to service layer

### 3. Service Layer (`application/agent/service.go`)

- Receives domain `ChatRequest`
- Calls `DBAgent.Chat()`
- Returns domain `ChatResponse`

### 4. Agent Execution (`infrastructure/agent/chat.go`)

```
┌─────────────────────────────────────────────────────────────────┐
│                        Chat() Method                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Get or create session (UUID)                                 │
│     └── New session: generates UUID                              │
│     └── Existing: retrieves from session service                 │
│                                                                  │
│  2. Store connection string in context                           │
│     └── ctx = context.WithValue(ctx, connectionStringKey, ...)   │
│     └── SECURITY: Never sent to LLM                              │
│                                                                  │
│  3. Run agent to completion                                      │
│     └── Sends ONLY message to LLM                                │
│     └── Iterates through all events                              │
│     └── Captures LAST model response (after tools complete)      │
│                                                                  │
│  4. Parse response                                               │
│     └── Extracts JSON from LLM response                          │
│     └── Converts to domain ChatResponse                          │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 5. Agent Event Flow

When the agent runs, it produces multiple events:

```
Event 1: Model decides to call read_schema
         ├── Type: FunctionCall
         └── Content: {name: "read_schema", args: {}}

Event 2: Tool executes and returns schema
         ├── Type: FunctionResponse
         └── Content: {schema: {...tables...}}

Event 3: Model generates final response
         ├── Type: Text (Role: Model)
         └── Content: {"message": "", "queries": [...]}
```

**Important**: We only capture Event 3 (the last model text response).

### 6. Response to Client

```json
{
  "success": true,
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "",
  "queries": [
    {
      "title": "Token activity for user",
      "query": "SELECT t.* FROM tokens t JOIN users u ON t.user_id = u.id WHERE u.email = 'test@example.com'",
      "description": "Retrieves all token activity for the specified user."
    }
  ]
}
```

## Session Continuity

For follow-up requests, include the `session_id`:

```json
{
  "message": "Now show only active tokens",
  "connection_string": "postgres://...",
  "session_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

The agent retains conversation history and can reference previous context.
