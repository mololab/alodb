package agent

import (
	"sync"
	"time"

	"github.com/mololab/alodb/internal/domain/database"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
)

type contextKey string

const (
	schemaCacheTTLKey contextKey = "schema_cache_ttl"
)

// SchemaHolder allows the tool handler to pass the read schema back to StreamChat.
type SchemaHolder struct {
	Schema *database.DatabaseSchema
}

type DBAgent struct {
	agent          agent.Agent
	runner         *runner.Runner
	sessionService session.Service
	modelSlug      string
	schemaCacheTTL time.Duration
	schemaStore    sync.Map // sessionID -> *database.DatabaseSchema
}
