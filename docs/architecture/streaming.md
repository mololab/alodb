# Streaming Architecture

AloDB uses WebSocket for real-time communication. The **client is a simple query runner** - the server sends SQL queries, the client executes them locally.

## Design Principles

1. **Client = Query Runner** - Client just executes SQL, no schema logic needed
2. **Server sends queries** - All schema extraction queries come from server
3. **Progress tracking** - Each query has a name for smooth UX
4. **Zero trust** - Database credentials never touch the server

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT                                          │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                      │
│  │    UI       │    │  WebSocket  │    │  PostgreSQL │                      │
│  │ (progress)  │◄──►│   Client    │◄──►│   Client    │                      │
│  └─────────────┘    └──────┬──────┘    └─────────────┘                      │
│                            │                                                 │
│   [2/10] Reading columns   │  Just runs SQL, returns rows                    │
│   for users...             │                                                 │
└────────────────────────────┼────────────────────────────────────────────────┘
                             │ WebSocket
┌────────────────────────────┼────────────────────────────────────────────────┐
│                            │                              SERVER             │
│                     ┌──────▼──────┐                                         │
│                     │   Handler   │                                         │
│                     └──────┬──────┘                                         │
│                            │                                                │
│  ┌─────────────────────────▼─────────────────────────────────────┐          │
│  │                      Agent + Schema Reader                     │          │
│  │                                                                │          │
│  │  Sends queries like:                                          │          │
│  │  • "SELECT current_database()"                                │          │
│  │  • "SELECT table_name FROM information_schema.tables..."      │          │
│  │  • "SELECT column_name, data_type FROM... WHERE table='X'"    │          │
│  └────────────────────────────────────────────────────────────────┘          │
│                                                                              │
│                    ⚠️  NO DATABASE ACCESS ON SERVER                          │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Event Flow

```
Client                          Server                           LLM
  │                               │                               │
  │──── WebSocket Connect ───────►│                               │
  │◄─── session_created ─────────│                               │
  │                               │                               │
  │──── chat{message} ───────────►│                               │
  │                               │──── Run Agent ───────────────►│
  │◄─── thinking{started} ───────│                               │
  │                               │◄─── FunctionCall: read_schema─│
  │                               │                               │
  │◄─── query_request ───────────│  "Getting database name"       │
  │     {query: "SELECT..."}      │  step 1/10                    │
  │                               │                               │
  │  [Execute SQL locally]        │                               │
  │                               │                               │
  │──── query_result{rows} ──────►│                               │
  │                               │                               │
  │◄─── query_request ───────────│  "Discovering tables"          │
  │     {query: "SELECT..."}      │  step 2/10                    │
  │                               │                               │
  │  ... more queries ...         │                               │
  │                               │                               │
  │                               │──── Schema complete ─────────►│
  │                               │◄─── Text response ────────────│
  │◄─── text_delta{chunk} ───────│                               │
  │◄─── response_complete ───────│                               │
```

## Components

### Domain Layer (`internal/domain/websocket/`)

| File       | Purpose                                          |
| ---------- | ------------------------------------------------ |
| `types.go` | Event/message types (query_request, etc.)        |
| `tools.go` | Schema query definitions with names/descriptions |

### Infrastructure Layer (`internal/infrastructure/websocket/`)

| File         | Purpose                                  |
| ------------ | ---------------------------------------- |
| `hub.go`     | Client registry and lifecycle management |
| `client.go`  | WebSocket client, query execution        |
| `handler.go` | Message routing, agent orchestration     |
| `errors.go`  | Error definitions                        |

### Agent Integration (`internal/infrastructure/agent/`)

| File               | Purpose                          |
| ------------------ | -------------------------------- |
| `stream_agent.go`  | Streaming chat with callbacks    |
| `schema/reader.go` | Sends schema queries to client   |
| `tools.go`         | ADK tool that uses schema reader |

## Schema Extraction Queries

The server sends these queries to the client (defined in `domain/websocket/tools.go`):

| Query Name                 | Purpose                      |
| -------------------------- | ---------------------------- |
| Getting database name      | `SELECT current_database()`  |
| Discovering tables         | List tables in public schema |
| Reading columns for X      | Get columns for each table   |
| Finding primary key for X  | Get PK columns               |
| Finding foreign keys for X | Get FK relationships         |
| Reading indexes for X      | Get index definitions        |

## Security Benefits

| Aspect            | This Architecture           |
| ----------------- | --------------------------- |
| Connection string | Never leaves client         |
| Query execution   | Always on client            |
| Firewall support  | Works behind any firewall   |
| Data exposure     | Zero - never touches server |
| Client complexity | Minimal - just runs SQL     |

## Timeout Configuration

| Operation           | Timeout    |
| ------------------- | ---------- |
| Single query        | 30 seconds |
| WebSocket ping/pong | 60 seconds |
| Write deadline      | 10 seconds |
