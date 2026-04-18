<h1 align="center">AloDB</h1>

<p align="center"><img width="100" height="100" alt="logo" src="https://github.com/user-attachments/assets/4b0b9292-c548-462a-af85-311f33c55d3a" /></p>

<p align="center"><b>AI-powered database assistant</b> - talk to your PostgreSQL databases in plain English. Your credentials and data never leave your machine.</p>

[Website](https://alodb.com) | [API Docs](./docs/api/README.md) | [Architecture](./docs/architecture/README.md)

https://github.com/user-attachments/assets/21d55d7f-bfc5-4e68-9110-15de653c8da0

<p align="center">
<img width="1512" height="1012" alt="alodb chat" src="https://github.com/user-attachments/assets/2a5d62d9-ac01-4021-a8c9-ee94f5cc1fcf" />
</p>

<p align="center">
<img width="1512" height="1012" alt="alodb studio" src="https://github.com/user-attachments/assets/cb559a48-0517-4ed8-9789-a84d55cba12b" />
</p>

---

## How It Works

AloDB uses a **split-execution architecture** - the AI runs on the server, but all database queries execute locally on your machine.

```
┌──────────────────────────────┐
│   Your Machine               │
│  ┌────────────────────────┐  │
│  │  AloDB Desktop App     │  │
│  │  (Electron + React)    │──┼──► PostgreSQL (local or remote)
│  │  + Local SQLite DB     │  │    (credentials stay here)
│  └──────────┬─────────────┘  │
└─────────────┼────────────────┘
              │ WebSocket (API key only)
┌─────────────▼────────────────┐
│   AloDB Server (Go)          │
│   + Google Gemini LLM        │
│   (NO database access)       │
└──────────────────────────────┘
```

1. You ask a question like *"Show me all users who signed up last month"*
2. The app connects to the server via WebSocket
3. The AI agent reads your database schema (by sending SQL queries back to the client for local execution)
4. The AI generates the appropriate SQL query
5. You review and run the query - execution happens entirely on your machine

**The server never touches your database.** It only sees schema structure and your questions - never credentials, connection strings, or row data.

---

## Quick Start

### Prerequisites

- **Go 1.24+**
- A **Gemini API key** ([get one here](https://aistudio.google.com/apikey))

### Setup

```bash
# Clone the repository
git clone https://github.com/mlolab/alodb.git
cd alodb

# Copy and configure environment
cp app.env.example app.env
```

Edit `app.env`:

```env
SERVER_PORT=8080
SERVER_ENV=development        # "development" or "production"
SCHEMA_CACHE_TTL=1h           # Schema cache duration (e.g., "1h", "30m")
```

### Run

```bash
make run
```

The server starts at `http://localhost:8080`.

---

## API

### Health Check

```bash
curl http://localhost:8080/v1/health
```

### Available Models

```bash
curl http://localhost:8080/v1/models
```

Returns all supported LLM models with provider metadata. Currently supports **6 Gemini models** including Gemini 3 Pro, Gemini 2.5 Flash/Pro, and more.

### WebSocket - Chat Streaming

```bash
# Connect (API key via query param)
wscat -c 'ws://localhost:8080/v1/agent/stream?api_key=YOUR_GEMINI_KEY'

# Or via header: X-Gemini-Api-Key
```

Once connected, you'll receive a `session_created` event. Then send chat messages:

```json
{"type": "chat", "payload": {"message": "Show me all users"}}
```

**Server events stream back in real-time:**

| Event               | Description                                          |
| ------------------- | ---------------------------------------------------- |
| `session_created`   | Connection established, returns session ID           |
| `thinking`          | Agent processing started                             |
| `query_request`     | Server asks client to execute a schema query locally |
| `text_delta`        | Streamed text chunks from the AI                     |
| `response_complete` | Final structured response with SQL queries           |
| `error`             | Error details                                        |

The client responds to `query_request` events by executing the SQL locally and returning results via `query_result` messages.

### Delete Session

```bash
curl -X DELETE http://localhost:8080/v1/sessions/{session_id}
```

See the full [WebSocket API docs](./docs/api/websocket.md) for protocol details and examples.

---

## Tech Stack

| Component     | Technology                | Purpose                                          |
| ------------- | ------------------------- | ------------------------------------------------ |
| Language      | **Go 1.24**               | Core server                                      |
| Web Framework | **Gin**                   | HTTP routing, middleware, CORS                   |
| WebSocket     | **Gorilla WebSocket**     | Real-time bidirectional communication            |
| AI Agent      | **Google ADK**            | Agent framework, tool orchestration              |
| LLM           | **Google Gemini** (genai) | Natural language understanding, SQL generation   |
| Logging       | **Zerolog**               | Structured logging (JSON in prod, pretty in dev) |
| Configuration | **Viper**                 | Env file + environment variable management       |

---

## Architecture

The backend follows **Domain-Driven Design (DDD)** with clean layer separation:

```
cmd/
└── main.go                    # Entry point, graceful shutdown

internal/
├── domain/                    # Pure business logic, no external deps
│   ├── agent/                 # Chat types, model registry, query types
│   ├── database/              # Schema type definitions
│   └── websocket/             # Event types, schema query definitions
│
├── application/               # Use cases, orchestration
│   └── agent/                 # Agent service lifecycle
│
└── infrastructure/            # External integrations
    ├── agent/                 # ADK integration, streaming, caching
    │   ├── cache/             # Schema caching with TTL
    │   ├── response/          # LLM response parser
    │   └── schema/            # Schema reader (client-side execution)
    ├── config/                # Viper configuration loading
    ├── websocket/             # Hub, client, handler, message routing
    └── web/                   # HTTP server, handlers, DTOs

pkg/
└── logger/                    # Zerolog singleton logger

prompts/
└── agent_instruction.md       # LLM agent behavior instructions
```

### Key Design Decisions

- **Client-side query execution** - the server never connects to any database. Schema queries are sent to the client via WebSocket, executed locally, and results returned.
- **Agent caching** - agents are cached by `model + SHA256(apiKey)[:8]` hash. Reused across sessions sharing the same model and key without storing raw keys.
- **Schema caching** - database schemas are cached per-session with configurable TTL (default 1h) to avoid redundant introspection.
- **External prompts** - agent instructions live in `prompts/agent_instruction.md`, allowing behavior changes without code changes.
- **Context-driven security** - credentials flow through Go's `context.Context`, never serialized or logged.

---

## Security Model

| Data                        | Leaves user's machine?               |
| --------------------------- | ------------------------------------ |
| Database credentials        | **Never**                            |
| Query results / row data    | **Never**                            |
| Database schema (structure) | Yes - sent to LLM for SQL generation |
| User's question             | Yes - sent to LLM                    |
| API key (Gemini)            | Per-request to LLM provider          |

- API keys are passed per-request via header (`X-Gemini-Api-Key`) or query param - never stored server-side
- All SQL execution happens on the client
- The LLM only sees schema structure and the user's question
- Sessions use cryptographically random UUIDs, stored in-memory only

---

## Commands

| Command      | Description                     |
| ------------ | ------------------------------- |
| `make run`   | Run the server                  |
| `make build` | Build binary to `bin/alodb`     |
| `make test`  | Run all tests                   |
| `make clean` | Remove build artifacts          |
| `make tidy`  | Download/update Go dependencies |

---

## Environment Variables

| Variable           | Required | Default      | Description                                             |
| ------------------ | -------- | ------------ | ------------------------------------------------------- |
| `SERVER_PORT`      | Yes      | -            | HTTP server port                                        |
| `SERVER_ENV`       | No       | `production` | `development` (pretty logs) or `production` (JSON logs) |
| `SCHEMA_CACHE_TTL` | No       | `1h`         | How long to cache database schemas per session          |

---

## Documentation

| Doc                                           | Description                                   |
| --------------------------------------------- | --------------------------------------------- |
| [API](./docs/api/README.md)                   | REST + WebSocket API, endpoints, protocol     |
| [Agent](./docs/agent/README.md)               | LLM agent behavior, tools, prompt engineering |
| [Architecture](./docs/architecture/README.md) | System architecture, request flow, streaming  |
| [Security](./docs/security/README.md)         | API key handling, credential protection       |
| [Development](./docs/development/README.md)   | Setup, debugging, contributing                |

---

<!-- IMAGES: Additional screenshots -->
<!-- ![Query Results](images/query-results.png) -->
<!-- ![Schema Reading Progress](images/schema-reading.png) -->
<!-- ![Model Selection](images/model-selection.png) -->
