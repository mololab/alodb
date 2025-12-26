package response

import (
	"encoding/json"
	"strings"

	domainAgent "github.com/mololab/alodb/internal/domain/agent"
	"github.com/mololab/alodb/pkg/logger"
)

// AgentResponse represents the expected JSON response structure from the LLM
type AgentResponse struct {
	Message string  `json:"message"`
	Queries []Query `json:"queries"`
}

// Query represents a query in the agent response
type Query struct {
	Title       string `json:"title"`
	Query       string `json:"query"`
	Description string `json:"description"`
}

// Parser handles parsing of agent responses
type Parser struct{}

// NewParser creates a new response parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses the raw LLM response into a structured ChatResponse
func (p *Parser) Parse(sessionID, rawResponse string) (*domainAgent.ChatResponse, error) {
	cleaned := p.cleanJSON(rawResponse)

	var parsed AgentResponse
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		logger.Debug().Err(err).Msg("failed to parse JSON response, returning raw")
		return &domainAgent.ChatResponse{
			SessionID: sessionID,
			Message:   rawResponse,
			Queries:   nil,
		}, nil
	}

	queries := p.convertQueries(parsed.Queries)
	logger.Debug().Int("queries", len(queries)).Msg("parsed response")

	return &domainAgent.ChatResponse{
		SessionID: sessionID,
		Message:   parsed.Message,
		Queries:   queries,
	}, nil
}

// cleanJSON removes markdown code blocks and extra whitespace from the response
func (p *Parser) cleanJSON(rawResponse string) string {
	cleaned := strings.TrimSpace(rawResponse)

	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```JSON")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")

	return strings.TrimSpace(cleaned)
}

// convertQueries converts parsed queries to domain queries
func (p *Parser) convertQueries(parsedQueries []Query) []domainAgent.Query {
	if len(parsedQueries) == 0 {
		return nil
	}

	queries := make([]domainAgent.Query, 0, len(parsedQueries))
	for _, q := range parsedQueries {
		queries = append(queries, domainAgent.Query{
			Title:       q.Title,
			Query:       q.Query,
			Description: q.Description,
		})
	}
	return queries
}

// IsValidJSONResponse checks if the response appears to be valid JSON
func (p *Parser) IsValidJSONResponse(rawResponse string) bool {
	cleaned := p.cleanJSON(rawResponse)
	return strings.HasPrefix(cleaned, "{") && strings.HasSuffix(cleaned, "}")
}
