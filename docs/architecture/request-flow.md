# Request Flow

This document describes how a request flows through the AloDB system.

## High-Level Flow

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  Client  │────▶│  Handler │────▶│  Service │────▶│  Manager │────▶│ DBAgent  │
└──────────┘     └──────────┘     └──────────┘     └──────────┘     └──────────┘
     │                │                │                │                │
     │   HTTP POST    │   Extract      │   Get/Create   │   Run Agent    │
     │   + API Key    │   API Key      │   Agent        │   with Tools   │
     │   Header       │   from Header  │                │                │
```

## Detailed Steps

### 1. Client Request

```bash
curl -X POST http://localhost:8080/v1/agent/chat \
  -H "Content-Type: application/json" \
  -H "X-Gemini-Api-Key: your-api-key" \
  -d '{
    "message": "Show me all users",
    "connection_string": "postgres://user:pass@localhost:5432/mydb"
  }'
```

### 2. Handler Layer (`web/handlers/agent_handler.go`)

- Receives HTTP POST request
- Validates JSON body
- Gets required header key from service based on model
- Extracts API key from request header
- Converts DTO to domain object with API key
- Passes to service layer

### 3. Service Layer (`application/agent/service.go`)

- Receives domain `ChatRequest` with API key
- Calls `Manager.GetAgent()` with model and API key
- Calls `DBAgent.Chat()`
- Returns domain `ChatResponse`

### 4. Manager Layer (`infrastructure/agent/manager.go`)

- Creates cache key from model + hashed API key
- Returns cached agent if exists, updates `lastUsed`
- Otherwise creates new agent with provided API key
- Caches agent for future requests

### 5. Agent Execution (`infrastructure/agent/chat.go`)

```
┌─────────────────────────────────────────────────────────────────┐
│                        Chat() Method                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Get or create session (UUID)                                 │
│                                                                  │
│  2. Store connection string in context                           │
│     └── SECURITY: Never sent to LLM                              │
│                                                                  │
│  3. Run agent to completion                                      │
│     └── Sends ONLY message to LLM                                │
│     └── Captures LAST model response                             │
│                                                                  │
│  4. Parse response                                               │
│     └── Extracts JSON from LLM response                          │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 6. Agent Event Flow

When the agent runs, it produces multiple events:

```
Event 1: Model decides to call read_schema
         └── Type: FunctionCall

Event 2: Tool executes and returns schema
         └── Type: FunctionResponse

Event 3: Model generates final response
         └── Type: Text (Role: Model)
```

We only capture Event 3 (the last model text response).

### 7. Response to Client

```json
{
  "success": true,
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "",
  "queries": [
    {
      "title": "Get all users",
      "query": "SELECT id, name, email FROM users ORDER BY id",
      "description": "Retrieves all users from the users table."
    }
  ]
}
```

## Agent Caching

Agents are cached by `model + hashed(apiKey)` for performance:

```
Request 1 (user A, gemini-2.5-pro):
  └── Cache miss → Create agent → Cache it

Request 2 (user A, gemini-2.5-pro):
  └── Cache hit → Return existing agent, update lastUsed

Request 3 (user B, gemini-2.5-pro, different API key):
  └── Different hash → Cache miss → Create new agent
```

## Session Continuity

For follow-up requests, include the `session_id`:

```json
{
  "message": "Now show only active users",
  "connection_string": "postgres://...",
  "session_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

The agent retains conversation history and can reference previous context.
