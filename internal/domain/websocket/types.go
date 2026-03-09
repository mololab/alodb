package websocket

import "time"

// server -> client event types
type ServerEventType string

const (
	EventSessionCreated   ServerEventType = "session_created"
	EventThinking         ServerEventType = "thinking"
	EventQueryRequest     ServerEventType = "query_request"
	EventTextDelta        ServerEventType = "text_delta"
	EventResponseComplete ServerEventType = "response_complete"
	EventError            ServerEventType = "error"
)

// client -> server message types
type ClientMessageType string

const (
	MessageChat        ClientMessageType = "chat"
	MessageQueryResult ClientMessageType = "query_result"
)

type ServerEvent struct {
	Type      ServerEventType `json:"type"`
	SessionID string          `json:"session_id"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   any             `json:"payload,omitempty"`
}

type ClientMessage struct {
	Type      ClientMessageType `json:"type"`
	SessionID string            `json:"session_id,omitempty"`
	Payload   any               `json:"payload"`
}

type SessionCreatedPayload struct {
	SessionID string `json:"session_id"`
}

type ThinkingPayload struct {
	Status string `json:"status"` // started, reading_schema, generating
}

type QueryRequestPayload struct {
	RequestID   string `json:"request_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Query       string `json:"query"`
	Step        int    `json:"step"`
	TotalSteps  int    `json:"total_steps"`
}

type TextDeltaPayload struct {
	Delta string `json:"delta"`
}

type ResponseCompletePayload struct {
	Success bool           `json:"success"`
	Message string         `json:"message,omitempty"`
	Queries []GeneratedSQL `json:"queries,omitempty"`
}

type GeneratedSQL struct {
	Title       string `json:"title"`
	Query       string `json:"query"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ChatPayload struct {
	Message string `json:"message"`
	Model   string `json:"model,omitempty"`
}

type QueryResultPayload struct {
	RequestID string `json:"request_id"`
	Success   bool   `json:"success"`
	Rows      []any  `json:"rows,omitempty"`
	Error     string `json:"error,omitempty"`
}
