# Security Documentation

AloDB implements several security measures to protect sensitive data.

## Connection String Protection

The most critical security feature is **never exposing database credentials to the LLM**.

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
│                    └───────────────────┘      to LLM             │
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

### Implementation

**1. Custom Context Key**

Using a custom type prevents key collisions:

```go
// types.go
type contextKey string
const connectionStringKey contextKey = "db_connection_string"
```

**2. Storage in Context**

Connection string is stored server-side only:

```go
// chat.go
ctx = context.WithValue(ctx, connectionStringKey, req.ConnectionString)
```

**3. Tool Access**

Tools read credentials at runtime:

```go
// tools.go
connStr, ok := toolCtx.Value(connectionStringKey).(string)
```

**4. LLM Never Sees It**

Only the user message is sent to the LLM:

```go
content := genai.NewContentFromText(req.Message, genai.RoleUser)
// ^ Only message, NO connection string
```

## What the LLM Sees

| Data | Visible to LLM? |
|------|-----------------|
| User message | ✅ Yes |
| Connection string | ❌ No |
| Database schema | ✅ Yes (via tool result) |
| Query results | ❌ No (not implemented yet) |

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
- Non-empty message

### SQL Injection Prevention

The agent generates SQL queries but **does not execute them**. The client is responsible for:
- Reviewing generated queries
- Using parameterized queries for execution
- Implementing proper access controls

## Best Practices for Deployment

### 1. Use HTTPS

Always deploy behind HTTPS to encrypt:
- Connection strings in transit
- API requests and responses

### 2. Secure Connection Strings

- Use environment variables, not hardcoded strings
- Rotate database passwords regularly
- Use read-only database users when possible

### 3. Network Security

- Run database on private network
- Use firewall rules to restrict access
- Consider VPN for remote access

### 4. API Security (Future)

Planned features:
- API key authentication
- Rate limiting
- Request logging and auditing

## Reporting Security Issues

If you discover a security vulnerability, please report it responsibly:

1. Do NOT open a public issue
2. Email security concerns to the maintainers
3. Provide detailed reproduction steps
4. Allow time for a fix before disclosure
