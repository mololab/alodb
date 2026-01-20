# Security Documentation

AloDB implements several security measures to protect sensitive data.

## API Key Handling

API keys are provided per-request via HTTP headers, not stored on the server.

### How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│                     API KEY FLOW                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   Client Request                                                 │
│   ┌─────────────────────┐                                        │
│   │ Header: X-Gemini-   │                                        │
│   │         Api-Key     │──────┐                                 │
│   └─────────────────────┘      │                                 │
│                                ▼                                 │
│                    ┌───────────────────┐                         │
│                    │     Handler       │                         │
│                    │  Extracts key     │                         │
│                    └───────────────────┘                         │
│                                │                                 │
│                                ▼                                 │
│                    ┌───────────────────┐                         │
│                    │     Manager       │                         │
│                    │  Hashes key for   │                         │
│                    │  cache lookup     │                         │
│                    └───────────────────┘                         │
│                                │                                 │
│                                ▼                                 │
│                    ┌───────────────────┐                         │
│                    │    DBAgent        │                         │
│                    │  Uses key for     │                         │
│                    │  LLM API calls    │                         │
│                    └───────────────────┘                         │
│                                                                  │
│   Raw API key is NEVER:                                          │
│   - Stored in environment variables (server-side)                │
│   - Logged                                                       │
│   - Persisted to disk                                            │
│   - Used as cache key (only hash is used)                        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Agent Caching

Agents are cached for performance using `model + SHA256(apiKey)[:8]`:

- Raw API key is never stored in the cache key
- Different API keys result in different cache entries
- Cache includes `lastUsed` timestamp for future cleanup

## Connection String Protection

Database credentials are never exposed to the LLM.

### How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│                     SECURITY BOUNDARY                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   User Request                                                   │
│   ┌─────────────────────┐                                        │
│   │ connection_string   │──────┐                                 │
│   │ message             │──────┼──┐                              │
│   └─────────────────────┘      │  │                              │
│                                │  │                              │
│                                ▼  │                              │
│                    ┌───────────────────┐                         │
│                    │  context.Context  │  ◄── Server-side only   │
│                    │  (Go runtime)     │      Never serialized   │
│                    └───────────────────┘                         │
│                                │                                 │
│                                │  Tool reads via                 │
│                                │  ctx.Value(key)                 │
│                                ▼                                 │
│                    ┌───────────────────┐                         │
│                    │   Schema Reader   │                         │
│                    │      Tool         │                         │
│                    └───────────────────┘                         │
│                                │                                 │
│                                ▼                                 │
│   LLM sees ONLY:   ┌───────────────────┐                         │
│                    │  "Show me users"  │  ◄── Clean message      │
│                    └───────────────────┘      No credentials     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## What the LLM Sees

| Data              | Visible to LLM? |
| ----------------- | --------------- |
| User message      | ✅ Yes          |
| API key           | ❌ No           |
| Connection string | ❌ No           |
| Database schema   | ✅ Yes (via tool)|
| Query results     | ❌ No           |

## Session Security

### Session IDs

- Generated using UUIDs (cryptographically random)
- Stored in-memory only
- Lost on server restart

### Session Isolation

Each session is isolated:

- Separate conversation history
- No cross-session data access
- Session ID required for follow-up requests

## Input Validation

### Request Validation

The handler validates:

- JSON body structure
- Required fields present
- Required API key header present

### SQL Injection Prevention

The agent generates SQL queries but **does not execute them**. The client is responsible for:

- Reviewing generated queries
- Using parameterized queries for execution
- Implementing proper access controls

## Best Practices for Deployment

### 1. Use HTTPS

Always deploy behind HTTPS to encrypt:

- API keys in transit
- Connection strings in transit
- API requests and responses

### 2. Secure Connection Strings

- Use read-only database users when possible
- Rotate database passwords regularly
- Run database on private network

### 3. Network Security

- Use firewall rules to restrict access
- Consider VPN for remote access
- Deploy behind a reverse proxy

### 4. Client-Side Key Storage

Clients should:

- Store API keys securely (not in source code)
- Use environment variables or secure vaults
- Rotate keys periodically

## Reporting Security Issues

If you discover a security vulnerability:

1. Do NOT open a public issue
2. Email security concerns to the maintainers
3. Provide detailed reproduction steps
4. Allow time for a fix before disclosure
