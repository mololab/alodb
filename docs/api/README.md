# API Documentation

AloDB exposes a REST API for interacting with the database agent.

## Base URL

```
http://localhost:{SERVER_PORT}/v1
```

## Endpoints

### GET /v1/models

Returns available AI models based on configured API keys.

#### Request

```bash
curl http://localhost:8080/v1/models
```

#### Response

**Success (200):**

```json
{
  "models": [
    {
      "slug": "gemini-2.5-pro-preview-06-05",
      "name": "Gemini 2.5 Pro",
      "provider": "google"
    },
    {
      "slug": "gemini-2.5-flash-preview-05-20",
      "name": "Gemini 2.5 Flash",
      "provider": "google"
    }
  ]
}
```

| Field               | Type   | Description                              |
| ------------------- | ------ | ---------------------------------------- |
| `models`            | array  | List of available models                 |
| `models[].slug`     | string | Model identifier to use in chat requests |
| `models[].name`     | string | Human-readable model name                |
| `models[].provider` | string | Provider name (google, openai)           |

---

### POST /v1/agent/chat

Chat with the database agent to generate SQL queries.

#### Request

**Headers:**

```
Content-Type: application/json
```

**Body (new conversation):**

```json
{
  "message": "Show me all users with their orders",
  "connection_string": "postgres://user:pass@localhost:5432/mydb"
}
```

**Body (with specific model):**

```json
{
  "message": "Show me all users with their orders",
  "connection_string": "postgres://user:pass@localhost:5432/mydb",
  "model": "gemini-2.5-flash-preview-05-20"
}
```

**Body (continue session):**

```json
{
  "message": "Now filter by active users only",
  "connection_string": "postgres://user:pass@localhost:5432/mydb",
  "session_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Body (switch model mid-conversation):**

```json
{
  "message": "Now filter by active users only",
  "connection_string": "postgres://user:pass@localhost:5432/mydb",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "model": "gemini-2.0-flash"
}
```

| Field               | Type   | Required | Description                                                           |
| ------------------- | ------ | -------- | --------------------------------------------------------------------- |
| `message`           | string | Yes      | Natural language query                                                |
| `connection_string` | string | Yes      | PostgreSQL connection URL                                             |
| `session_id`        | string | No       | UUID from previous response to continue conversation                  |
| `model`             | string | No       | Model slug from /v1/models (defaults to gemini-2.5-pro-preview-06-05) |

#### Response

**Success (200):**

```json
{
  "success": true,
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "",
  "queries": [
    {
      "title": "Get users with their orders",
      "query": "SELECT u.id, u.name, u.email, o.id AS order_id, o.total FROM users u LEFT JOIN orders o ON u.id = o.user_id ORDER BY u.id",
      "description": "This query joins the users table with orders using LEFT JOIN to include users without orders."
    }
  ]
}
```

| Field                   | Type    | Description                        |
| ----------------------- | ------- | ---------------------------------- |
| `success`               | boolean | Whether the request succeeded      |
| `session_id`            | string  | UUID to use for follow-up requests |
| `message`               | string  | Optional message or explanation    |
| `queries`               | array   | Array of generated SQL queries     |
| `queries[].title`       | string  | Short descriptive title            |
| `queries[].query`       | string  | The SQL query                      |
| `queries[].description` | string  | Detailed explanation               |

**Error (400/500):**

```json
{
  "success": false,
  "error": "Error message here"
}
```

#### Examples

**Example 1: Simple Query**

Request:

```bash
curl -X POST http://localhost:8080/v1/agent/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Show me all users",
    "connection_string": "postgres://root:secret@localhost:5432/mydb"
  }'
```

Response:

```json
{
  "success": true,
  "session_id": "abc123...",
  "message": "",
  "queries": [
    {
      "title": "Get all users",
      "query": "SELECT id, name, email, created_at FROM users ORDER BY created_at DESC",
      "description": "Retrieves all users ordered by creation date."
    }
  ]
}
```

**Example 2: Follow-up Query**

Request:

```bash
curl -X POST http://localhost:8080/v1/agent/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Now only show active users",
    "connection_string": "postgres://root:secret@localhost:5432/mydb",
    "session_id": "abc123..."
  }'
```

Response:

```json
{
  "success": true,
  "session_id": "abc123...",
  "message": "",
  "queries": [
    {
      "title": "Get active users",
      "query": "SELECT id, name, email, created_at FROM users WHERE status = 'active' ORDER BY created_at DESC",
      "description": "Filters the previous query to show only active users."
    }
  ]
}
```

---

### GET /v1/health

Health check endpoint.

#### Request

```bash
curl http://localhost:8080/v1/health
```

#### Response

**Success (200):**

```json
{
  "status": "healthy"
}
```

---

## Connection String Format

PostgreSQL connection strings follow this format:

```
postgres://username:password@host:port/database?sslmode=disable
```

| Component  | Description                              |
| ---------- | ---------------------------------------- |
| `username` | Database user                            |
| `password` | Database password                        |
| `host`     | Database server hostname                 |
| `port`     | Database server port (default: 5432)     |
| `database` | Database name                            |
| `sslmode`  | SSL mode (disable, require, verify-full) |

**Examples:**

```
postgres://root:secret@localhost:5432/mydb?sslmode=disable
postgres://admin:pass123@db.example.com:5432/production?sslmode=require
```

## Error Codes

| Status | Meaning                                    |
| ------ | ------------------------------------------ |
| 200    | Success                                    |
| 400    | Bad request (invalid JSON, missing fields) |
| 500    | Internal server error                      |

## Rate Limiting

Currently no rate limiting is implemented. This is planned for future releases.
