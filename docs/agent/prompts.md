# Prompt Engineering

The agent's behavior is controlled by the system prompt in `prompts/agent_instruction.md`.

## Prompt Location

```
prompts/
└── agent_instruction.md    # Main agent system prompt
```

## Prompt Structure

The prompt contains these sections:

### 1. Critical Tool Usage

```markdown
## CRITICAL: Tool Usage

You MUST call the `read_schema` tool FIRST before responding. 
Do NOT output any text before calling the tool.

**Wrong behavior:**
- Saying "I will call read_schema..." 
- Explaining what you're about to do

**Correct behavior:**
- Immediately call `read_schema` tool
- Wait for schema data
- Then generate your JSON response
```

**Why this matters**: LLMs tend to "think out loud". Without explicit instructions to NOT output text before tool calls, the agent would say "I will call the tool..." which gets captured as the response instead of the actual result.

### 2. Available Tools

Lists all tools the agent can call with their purposes.

### 3. Workflow

Step-by-step process the agent should follow:

1. Call `read_schema` tool (no text output)
2. Analyze the returned schema
3. Generate SQL query for user's request
4. Return JSON response

### 4. Response Format

Specifies the exact JSON structure required:

```json
{
  "message": "",
  "queries": [
    {
      "title": "Short descriptive title",
      "query": "SELECT ... FROM ... WHERE ...",
      "description": "What this query does and why"
    }
  ]
}
```

### 5. SQL Best Practices

Guidelines for query generation:

- Use table aliases
- Use explicit JOINs
- Select specific columns
- Use foreign keys for joins
- Default to read-only queries

### 6. Examples

Concrete examples showing expected behavior.

### 7. Rules

Mandatory behaviors that override other instructions.

## Editing the Prompt

1. Open `prompts/agent_instruction.md`
2. Make your changes
3. Restart the server (`make run`)

The prompt is loaded at agent initialization, so changes require a restart.

## Best Practices

### Be Explicit About What NOT to Do

```markdown
**Wrong behavior:**
- Saying "I will call read_schema..." 

**Correct behavior:**
- Immediately call `read_schema` tool
```

### Use Examples

Concrete examples are more effective than abstract rules:

```markdown
### Example: User asks "Show me all users"

1. You call: `read_schema` (no text output!)
2. You receive: schema with users table
3. You respond:

{
  "message": "",
  "queries": [...]
}
```

### Keep It Focused

The prompt should be:
- Concise but complete
- Action-oriented
- Example-rich
- Rule-based for critical behaviors

## Common Issues

| Issue | Cause | Fix |
|-------|-------|-----|
| Agent says "I will call..." | LLM thinking out loud | Add explicit "no text before tool" rule |
| Invalid JSON in response | Missing format specification | Add JSON examples |
| Wrong table/column names | Guessing instead of reading schema | Emphasize "ALWAYS call read_schema first" |
| Markdown around JSON | LLM adding code blocks | Add "no markdown formatting" rule |
