# API Documentation

AloDB uses a WebSocket-based API where the **client executes all database queries locally**.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                           YOUR MACHINE                               │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐         │
│  │     UI       │◄───►│   Client     │◄───►│  PostgreSQL  │         │
│  │              │     │  (executes   │     │  (local or   │         │
│  │              │     │   queries)   │     │   remote)    │         │
│  └──────────────┘     └──────┬───────┘     └──────────────┘         │
│                              │                                       │
└──────────────────────────────┼───────────────────────────────────────┘
                               │ WebSocket (only API key goes to server)
                               │
┌──────────────────────────────┼───────────────────────────────────────┐
│                              │                    ALODB SERVER        │
│                       ┌──────▼───────┐                               │
│                       │    Agent     │◄──► LLM (Gemini)              │
│                       │  (sends SQL  │                               │
│                       │   queries)   │                               │
│                       └──────────────┘                               │
│                                                                      │
│                    ⚠️  NO DATABASE ACCESS HERE                        │
└──────────────────────────────────────────────────────────────────────┘
```

## Why Client-Side Execution?

1. **Security**: Database credentials never leave your machine
2. **Firewall-friendly**: Works with databases behind firewalls
3. **Simple client**: Just a query runner - server sends the SQL
4. **Progress tracking**: Each query has a name for smooth UX

## Endpoints

### GET /v1/agent/stream (WebSocket)

Main endpoint for all agent interaction. See [WebSocket API](./websocket.md) for full protocol.

### DELETE /v1/sessions/:session_id

Deletes a session to free memory. Call this when a user closes their chat.

```bash
curl -X DELETE http://localhost:8080/v1/sessions/your-session-id
```

```json
{ "deleted": true, "session_id": "your-session-id" }
```

### GET /v1/providers

Returns available AI providers. **No authentication required.** Use this to discover which providers are available and what API key header each one expects.

```bash
curl http://localhost:8080/v1/providers
```

```json
{
  "providers": [
    {
      "provider": "google",
      "header_key": "X-Gemini-Api-Key",
      "metadata": {
        "name": "Google Gemini",
        "description": "Google's Gemini family of multimodal AI models"
      }
    }
  ]
}
```

### GET /v1/providers/:provider/models

Returns available models for a specific provider. **Requires an API key** - models are fetched dynamically from the provider's API and cached for 1 hour.

Pass the API key via `api_key` query param or the provider's header (e.g. `X-Gemini-Api-Key` for Google).

```bash
curl -H "X-Gemini-Api-Key: your-key" http://localhost:8080/v1/providers/google/models
```

```json
{
  "models": [
    { "slug": "gemini-2.5-flash", "name": "Gemini 2.5 Flash", "provider": "google" },
    { "slug": "gemini-2.5-pro", "name": "Gemini 2.5 Pro", "provider": "google" },
    ...
  ]
}
```

Only models that support `generateContent` are returned.

### GET /v1/health

Health check endpoint.

```bash
curl http://localhost:8080/v1/health
```

```json
{ "status": "healthy" }
```

## Quick Start

1. Get a Gemini API key from [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Start AloDB server
3. Call `GET /v1/providers` to discover available providers
4. Call `GET /v1/providers/google/models` with your API key to list models
5. Connect via WebSocket with your API key and chosen model
6. Send chat messages, execute queries locally when prompted
7. Receive generated SQL queries

See [WebSocket API](./websocket.md) for the complete protocol, client implementation examples, and full type definitions.

## Message Types Quick Reference

### Server → Client Events

| Event Type          | Payload Type              | Description                       |
| ------------------- | ------------------------- | --------------------------------- |
| `session_created`   | `SessionCreatedPayload`   | Connection established            |
| `thinking`          | `ThinkingPayload`         | Agent processing status           |
| `query_request`     | `QueryRequestPayload`     | SQL query for client to execute   |
| `text_delta`        | `TextDeltaPayload`        | Streaming text from LLM           |
| `response_complete` | `ResponseCompletePayload` | Final response with generated SQL |
| `error`             | `ErrorPayload`            | Error occurred                    |

### Client → Server Messages

| Message Type   | Payload Type         | Description                   |
| -------------- | -------------------- | ----------------------------- |
| `chat`         | `ChatPayload`        | Send message to agent         |
| `query_result` | `QueryResultPayload` | Return query execution result |

See [WebSocket API - Request & Response Reference](./websocket.md#request--response-reference) for complete TypeScript definitions.
