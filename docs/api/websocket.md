# WebSocket API

AloDB uses WebSocket for real-time streaming with **client-side query execution**. The client is a simple query runner - it just executes SQL sent by the server and returns results.

## Why This Architecture?

1. **Security**: Your database credentials and data never touch the server
2. **Firewall-friendly**: Works with databases behind firewalls
3. **Simple client**: Client just runs SQL queries, no schema logic needed
4. **Progress tracking**: Each query has a name for smooth UX

## Connection

### Endpoint

```
ws://localhost:{SERVER_PORT}/v1/agent/stream
```

### Authentication

API key can be provided via **query parameter** or **header**:

| Method      | Format                       |
| ----------- | ---------------------------- |
| Query param | `?api_key=your-key`          |
| Header      | `X-Gemini-Api-Key: your-key` |

### Query Parameters

| Parameter | Required | Description                              |
| --------- | -------- | ---------------------------------------- |
| `api_key` | No*      | LLM provider API key (*or use header)    |
| `model`   | No       | Model slug (defaults to first available) |

### Example Connections

**Via query parameter:**
```javascript
const ws = new WebSocket(
  'ws://localhost:8080/v1/agent/stream?api_key=your-gemini-key'
);
```

**Note:** Browser WebSocket API doesn't support custom headers. Use query param from browsers, or use a library like `ws` in Node.js for header support.

## Protocol

### Server → Client Events

```json
{
  "type": "event_type",
  "session_id": "uuid",
  "timestamp": "2024-01-15T10:30:00Z",
  "payload": {}
}
```

| Type                | Description                                |
| ------------------- | ------------------------------------------ |
| `session_created`   | Connection established                     |
| `thinking`          | Agent processing status                    |
| `query_request`     | Server sends SQL query for client to run   |
| `text_delta`        | Streaming text from agent                  |
| `response_complete` | Final structured response with SQL queries |
| `error`             | Error occurred                             |

### Client → Server Messages

```json
{
  "type": "message_type",
  "payload": {}
}
```

| Type           | Description                      |
| -------------- | -------------------------------- |
| `chat`         | Send a message to the agent      |
| `query_result` | Return result of query execution |

## Complete Flow Example

### 1. Connect and Receive Session

```json
← {
  "type": "session_created",
  "session_id": "550e8400-...",
  "payload": { "session_id": "550e8400-..." }
}
```

### 2. Send Chat Message

```json
→ {
  "type": "chat",
  "payload": { "message": "Show me all users" }
}
```

### 3. Receive Thinking Status

```json
← {
  "type": "thinking",
  "payload": { "status": "started" }
}
```

### 4. Receive Query Requests (One by One)

The server sends SQL queries with human-readable names. Client just executes and returns results.

```json
← {
  "type": "query_request",
  "payload": {
    "request_id": "req-001",
    "name": "Getting database name",
    "description": "Retrieves the current database name",
    "query": "SELECT current_database()",
    "step": 1,
    "total_steps": 6
  }
}
```

### 5. Execute Query and Return Result

```json
→ {
  "type": "query_result",
  "payload": {
    "request_id": "req-001",
    "success": true,
    "rows": [{ "current_database": "mydb" }]
  }
}
```

### 6. More Query Requests Follow...

```json
← {
  "type": "query_request",
  "payload": {
    "request_id": "req-002",
    "name": "Discovering tables",
    "description": "Lists all tables in the public schema",
    "query": "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'...",
    "step": 2,
    "total_steps": 6
  }
}
```

```json
→ {
  "type": "query_result",
  "payload": {
    "request_id": "req-002",
    "success": true,
    "rows": [
      { "table_name": "users" },
      { "table_name": "orders" }
    ]
  }
}
```

### 7. Receive Final Response

```json
← {
  "type": "response_complete",
  "payload": {
    "success": true,
    "message": "",
    "queries": [
      {
        "title": "Get all users",
        "query": "SELECT id, name, email FROM users ORDER BY id",
        "description": "Retrieves all users from the database."
      }
    ]
  }
}
```

## Query Request Details

Each `query_request` includes:

| Field         | Description                         |
| ------------- | ----------------------------------- |
| `request_id`  | Unique ID to match with result      |
| `name`        | Human-readable name for UI progress |
| `description` | What this query does                |
| `query`       | The actual SQL to execute           |
| `step`        | Current step number (e.g., 3)       |
| `total_steps` | Total steps in operation (e.g., 10) |

**Example progress display:**
```
[3/10] Reading columns for users...
```

## Client Implementation

### JavaScript Example

```javascript
class AloDBClient {
  constructor(apiKey, connectionString, model = null) {
    this.apiKey = apiKey;
    this.connectionString = connectionString;
    this.model = model;
    this.ws = null;
    this.db = null; // Your PostgreSQL client
  }

  async connect() {
    // Connect to your local PostgreSQL
    this.db = await connectToPostgres(this.connectionString);

    // Connect to AloDB server
    const params = new URLSearchParams({
      api_key: this.apiKey,
      ...(this.model && { model: this.model })
    });

    this.ws = new WebSocket(`ws://localhost:8080/v1/agent/stream?${params}`);
    this.ws.onmessage = (e) => this.handleMessage(JSON.parse(e.data));
  }

  handleMessage(event) {
    switch (event.type) {
      case 'session_created':
        console.log('Connected:', event.payload.session_id);
        break;

      case 'query_request':
        this.executeQuery(event.payload);
        break;

      case 'response_complete':
        console.log('Generated queries:', event.payload.queries);
        break;

      case 'error':
        console.error('Error:', event.payload.message);
        break;
    }
  }

  async executeQuery(payload) {
    const { request_id, name, query, step, total_steps } = payload;
    
    // Show progress to user
    console.log(`[${step}/${total_steps}] ${name}`);

    try {
      // Execute query locally
      const result = await this.db.query(query);
      
      this.send({
        type: 'query_result',
        payload: {
          request_id,
          success: true,
          rows: result.rows
        }
      });
    } catch (error) {
      this.send({
        type: 'query_result',
        payload: {
          request_id,
          success: false,
          error: error.message
        }
      });
    }
  }

  chat(message) {
    this.send({ type: 'chat', payload: { message } });
  }

  send(data) {
    this.ws.send(JSON.stringify(data));
  }
}

// Usage
const client = new AloDBClient(
  'your-gemini-api-key',
  'postgres://user:pass@localhost:5432/mydb'
);
await client.connect();
client.chat('Show me all users with their orders');
```

## Error Handling

```json
{
  "type": "error",
  "payload": {
    "code": "error_code",
    "message": "Human readable message"
  }
}
```

| Code               | Description                   |
| ------------------ | ----------------------------- |
| `invalid_message`  | Malformed message format      |
| `invalid_payload`  | Invalid payload structure     |
| `agent_error`      | Agent initialization failed   |
| `processing_error` | Error during chat processing  |
| `query_timeout`    | Client didn't respond in time |

## Timeouts

| Operation       | Timeout |
| --------------- | ------- |
| Query execution | 30 sec  |
| WebSocket ping  | 60 sec  |

---

## Request & Response Reference

This section provides complete TypeScript type definitions for all WebSocket messages.

### Base Message Types

#### ServerEvent (Server → Client)

All messages from server to client follow this envelope:

```typescript
interface ServerEvent {
  type: ServerEventType;
  session_id: string;
  timestamp: string; // ISO 8601 format
  payload?: object;
}

type ServerEventType =
  | "session_created"
  | "thinking"
  | "query_request"
  | "text_delta"
  | "response_complete"
  | "error";
```

#### ClientMessage (Client → Server)

All messages from client to server follow this envelope:

```typescript
interface ClientMessage {
  type: ClientMessageType;
  session_id?: string; // Optional, server tracks session
  payload: object;
}

type ClientMessageType = "chat" | "query_result";
```

---

### Server → Client Payloads

#### `session_created`

Sent immediately when WebSocket connection is established.

```typescript
interface SessionCreatedPayload {
  session_id: string;
}
```

**Example:**
```json
{
  "type": "session_created",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-01-15T10:30:00Z",
  "payload": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

---

#### `thinking`

Indicates agent processing state changes.

```typescript
interface ThinkingPayload {
  status: "started" | "reading_schema" | "generating";
}
```

| Status           | Description                          |
| ---------------- | ------------------------------------ |
| `started`        | Agent began processing the request   |
| `reading_schema` | Agent is reading database schema     |
| `generating`     | Agent is generating the SQL response |

**Example:**
```json
{
  "type": "thinking",
  "session_id": "550e8400-...",
  "timestamp": "2024-01-15T10:30:01Z",
  "payload": {
    "status": "reading_schema"
  }
}
```

---

#### `query_request`

Server sends a SQL query for the client to execute locally.

```typescript
interface QueryRequestPayload {
  request_id: string;  // Unique ID to match with result
  name: string;        // Human-readable name for UI progress
  description: string; // What this query does
  query: string;       // The actual SQL to execute
  step: number;        // Current step number (1-based)
  total_steps: number; // Total steps in operation
}
```

**Example:**
```json
{
  "type": "query_request",
  "session_id": "550e8400-...",
  "timestamp": "2024-01-15T10:30:02Z",
  "payload": {
    "request_id": "req-001",
    "name": "Discovering tables",
    "description": "Lists all tables in the public schema",
    "query": "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'",
    "step": 2,
    "total_steps": 6
  }
}
```

---

#### `text_delta`

Streaming text output from the LLM (for real-time display).

```typescript
interface TextDeltaPayload {
  delta: string; // Incremental text chunk
}
```

**Example:**
```json
{
  "type": "text_delta",
  "session_id": "550e8400-...",
  "timestamp": "2024-01-15T10:30:05Z",
  "payload": {
    "delta": "Based on your database schema, "
  }
}
```

**Usage:** Concatenate all `delta` values to build the complete response text.

---

#### `response_complete`

Sent when agent finishes processing. Contains the final structured response.

```typescript
interface ResponseCompletePayload {
  success: boolean;
  message?: string;          // Optional message/explanation
  queries?: GeneratedSQL[];  // Generated SQL queries
}

interface GeneratedSQL {
  title: string;       // Human-readable query title
  query: string;       // The SQL query
  description: string; // What the query does
}
```

**Example (success):**
```json
{
  "type": "response_complete",
  "session_id": "550e8400-...",
  "timestamp": "2024-01-15T10:30:10Z",
  "payload": {
    "success": true,
    "message": "",
    "queries": [
      {
        "title": "Get all users",
        "query": "SELECT id, name, email FROM users ORDER BY id",
        "description": "Retrieves all users from the database."
      },
      {
        "title": "Count active users",
        "query": "SELECT COUNT(*) FROM users WHERE active = true",
        "description": "Returns the number of active users."
      }
    ]
  }
}
```

**Example (no queries needed):**
```json
{
  "type": "response_complete",
  "session_id": "550e8400-...",
  "timestamp": "2024-01-15T10:30:10Z",
  "payload": {
    "success": true,
    "message": "I can help you with SQL queries. What would you like to know about your database?",
    "queries": []
  }
}
```

---

#### `error`

Sent when an error occurs.

```typescript
interface ErrorPayload {
  code: string;    // Error code for programmatic handling
  message: string; // Human-readable error message
}
```

| Code               | Description                            |
| ------------------ | -------------------------------------- |
| `invalid_message`  | Malformed JSON or message format       |
| `invalid_payload`  | Invalid payload structure for the type |
| `agent_error`      | Agent initialization or runtime error  |
| `processing_error` | Error during chat/request processing   |
| `query_timeout`    | Client didn't respond to query in time |

**Example:**
```json
{
  "type": "error",
  "session_id": "550e8400-...",
  "timestamp": "2024-01-15T10:30:15Z",
  "payload": {
    "code": "query_timeout",
    "message": "Query execution timed out after 30 seconds"
  }
}
```

---

### Client → Server Payloads

#### `chat`

Send a message to the agent.

```typescript
interface ChatPayload {
  message: string;  // The user's message/question
  model?: string;   // Optional: override model for this request
}
```

**Example:**
```json
{
  "type": "chat",
  "payload": {
    "message": "Show me all users who signed up last month"
  }
}
```

**Example with model override:**
```json
{
  "type": "chat",
  "payload": {
    "message": "Show me all users",
    "model": "gemini-2.5-flash"
  }
}
```

---

#### `query_result`

Return the result of a query execution. Send this after receiving a `query_request`.

```typescript
interface QueryResultPayload {
  request_id: string; // Must match the request_id from query_request
  success: boolean;   // Whether query executed successfully
  rows?: any[];       // Query result rows (on success)
  error?: string;     // Error message (on failure)
}
```

**Example (success):**
```json
{
  "type": "query_result",
  "payload": {
    "request_id": "req-001",
    "success": true,
    "rows": [
      { "id": 1, "name": "Alice", "email": "alice@example.com" },
      { "id": 2, "name": "Bob", "email": "bob@example.com" }
    ]
  }
}
```

**Example (failure):**
```json
{
  "type": "query_result",
  "payload": {
    "request_id": "req-001",
    "success": false,
    "error": "relation \"users\" does not exist"
  }
}
```

---

## Complete Type Definitions

For quick copy-paste into your TypeScript project:

```typescript
// ============================================
// Server → Client
// ============================================

type ServerEventType =
  | "session_created"
  | "thinking"
  | "query_request"
  | "text_delta"
  | "response_complete"
  | "error";

interface ServerEvent {
  type: ServerEventType;
  session_id: string;
  timestamp: string;
  payload?: object;
}

interface SessionCreatedPayload {
  session_id: string;
}

interface ThinkingPayload {
  status: "started" | "reading_schema" | "generating";
}

interface QueryRequestPayload {
  request_id: string;
  name: string;
  description: string;
  query: string;
  step: number;
  total_steps: number;
}

interface TextDeltaPayload {
  delta: string;
}

interface ResponseCompletePayload {
  success: boolean;
  message?: string;
  queries?: GeneratedSQL[];
}

interface GeneratedSQL {
  title: string;
  query: string;
  description: string;
}

interface ErrorPayload {
  code: string;
  message: string;
}

// ============================================
// Client → Server
// ============================================

type ClientMessageType = "chat" | "query_result";

interface ClientMessage {
  type: ClientMessageType;
  session_id?: string;
  payload: object;
}

interface ChatPayload {
  message: string;
  model?: string;
}

interface QueryResultPayload {
  request_id: string;
  success: boolean;
  rows?: any[];
  error?: string;
}
```
