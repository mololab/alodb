package agent

import (
	"time"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
)

type contextKey string

const (
	connectionStringKey contextKey = "db_connection_string"
	schemaCacheTTLKey   contextKey = "schema_cache_ttl"
)

type DBAgent struct {
	agent          agent.Agent
	runner         *runner.Runner
	sessionService session.Service
	modelSlug      string
	schemaCacheTTL time.Duration
}
