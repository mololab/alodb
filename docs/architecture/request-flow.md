# Request Flow

This document describes how a request flows through the AloDB system via WebSocket.

## High-Level Flow

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  Client  │◄───►│   Hub    │────►│  Handler │────►│ DBAgent  │
│ (query   │     │          │     │          │     │          │
│  runner) │     │          │     │          │     │          │
└──────────┘     └──────────┘     └──────────┘     └──────────┘
     │                                                  │
     │   Executes SQL locally                          │
     │   Returns rows                                  │
     │◄────────────────────────────────────────────────┤
                                              Sends query via WS
```

## Detailed Steps

### 1. Client Connects

```javascript
const ws = new WebSocket('ws://localhost:8080/v1/agent/stream?api_key=...');
```

### 2. Server Creates Session

```json
← { "type": "session_created", "payload": { "session_id": "..." } }
```

### 3. Client Sends Chat Message

```json
→ { "type": "chat", "payload": { "message": "Show me all users" } }
```

### 4. Handler Layer (`websocket/handler.go`)

- Receives chat message
- Gets or creates agent from manager
- Creates query executor wrapper
- Starts streaming chat

### 5. Agent Execution (`agent/stream_agent.go`)

```
┌─────────────────────────────────────────────────────────────────┐
│                     StreamChat() Method                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Get or create session                                        │
│                                                                  │
│  2. Store query executor in context                              │
│                                                                  │
│  3. Run agent                                                    │
│     └── Agent calls read_schema tool                             │
│     └── Tool uses schema.Reader to send queries                  │
│     └── Each query goes to client via WebSocket                  │
│                                                                  │
│  4. Stream LLM response back to client                           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 6. Schema Reading Flow

When the agent calls `read_schema`, the Schema Reader sends queries one by one:

```
Server                              Client
   │                                   │
   │──► query_request ────────────────►│
   │    "Getting database name"        │
   │    step 1/6                       │
   │                                   │ [runs SQL]
   │◄── query_result ─────────────────◄│
   │    rows: [{"current_database":...}]
   │                                   │
   │──► query_request ────────────────►│
   │    "Discovering tables"           │
   │    step 2/6                       │
   │                                   │ [runs SQL]
   │◄── query_result ─────────────────◄│
   │    rows: [{table_name: "users"}...]
   │                                   │
   │  ... more queries for each table ...
```

### 7. Response to Client

```json
← {
  "type": "response_complete",
  "payload": {
    "success": true,
    "queries": [
      {
        "title": "Get all users",
        "query": "SELECT id, name, email FROM users ORDER BY id",
        "description": "Retrieves all users from the users table."
      }
    ]
  }
}
```

## Agent Caching

Agents are cached by `model + hashed(apiKey)` for performance:

```
Request 1 (user A, gemini-2.5-pro):
  └── Cache miss → Create agent → Cache it

Request 2 (user A, gemini-2.5-pro):
  └── Cache hit → Return existing agent

Request 3 (user B, gemini-2.5-pro, different API key):
  └── Different hash → Cache miss → Create new agent
```

## Session Continuity

Sessions persist across chat messages within the same WebSocket connection.
The agent retains conversation history and can reference previous context.
