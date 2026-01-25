package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	ws "github.com/mololab/alodb/internal/domain/websocket"
	"github.com/mololab/alodb/pkg/logger"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB
)

type Client struct {
	ID        string
	SessionID string
	APIKey    string
	Model     string

	conn    *websocket.Conn
	send    chan []byte
	hub     *Hub
	handler MessageHandler

	closed   bool
	closedMu sync.RWMutex

	// query result channels - keyed by request id
	queryResults   map[string]chan *ws.QueryResultPayload
	queryResultsMu sync.RWMutex
}

type MessageHandler interface {
	HandleChat(client *Client, payload *ws.ChatPayload)
}

func NewClient(conn *websocket.Conn, hub *Hub, handler MessageHandler) *Client {
	return &Client{
		ID:           uuid.New().String(),
		SessionID:    uuid.New().String(),
		conn:         conn,
		send:         make(chan []byte, 256),
		hub:          hub,
		handler:      handler,
		closed:       false,
		queryResults: make(map[string]chan *ws.QueryResultPayload),
	}
}

// Close marks the client as closed and closes the send channel
func (c *Client) Close() {
	c.closedMu.Lock()
	defer c.closedMu.Unlock()
	if !c.closed {
		c.closed = true
		close(c.send)
	}
}

// SendEvent sends a typed event to the client
func (c *Client) SendEvent(eventType ws.ServerEventType, payload any) {
	// check if client is closed before attempting to send
	c.closedMu.RLock()
	if c.closed {
		c.closedMu.RUnlock()
		return
	}
	c.closedMu.RUnlock()

	event := ws.ServerEvent{
		Type:      eventType,
		SessionID: c.SessionID,
		Timestamp: time.Now(),
		Payload:   payload,
	}

	data, err := json.Marshal(event)
	if err != nil {
		logger.Error().Err(err).Msg("failed to marshal event")
		return
	}

	// additional safety net in case of race condition
	defer func() {
		if r := recover(); r != nil {
			logger.Debug().Str("client_id", c.ID).Msg("send attempted on closed client")
		}
	}()

	select {
	case c.send <- data:
	default:
		logger.Warn().Str("client_id", c.ID).Msg("send buffer full, dropping message")
	}
}

// ExecuteQuery sends a query to client for execution and waits for result
func (c *Client) ExecuteQuery(name, description, query string, step, totalSteps int, timeout time.Duration) (*ws.QueryResultPayload, error) {
	requestID := uuid.New().String()

	resultCh := make(chan *ws.QueryResultPayload, 1)
	c.queryResultsMu.Lock()
	c.queryResults[requestID] = resultCh
	c.queryResultsMu.Unlock()

	defer func() {
		c.queryResultsMu.Lock()
		delete(c.queryResults, requestID)
		c.queryResultsMu.Unlock()
	}()

	c.SendEvent(ws.EventQueryRequest, ws.QueryRequestPayload{
		RequestID:   requestID,
		Name:        name,
		Description: description,
		Query:       query,
		Step:        step,
		TotalSteps:  totalSteps,
	})

	select {
	case result := <-resultCh:
		return result, nil
	case <-time.After(timeout):
		return nil, ErrQueryTimeout
	}
}

// DeliverQueryResult delivers a query result to the waiting goroutine
func (c *Client) DeliverQueryResult(result *ws.QueryResultPayload) {
	c.queryResultsMu.RLock()
	ch, exists := c.queryResults[result.RequestID]
	c.queryResultsMu.RUnlock()

	if exists {
		select {
		case ch <- result:
		default:
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Warn().Err(err).Str("client_id", c.ID).Msg("websocket read error")
			}
			return
		}

		c.handleMessage(data)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(data []byte) {
	var msg ws.ClientMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Warn().Err(err).Msg("invalid message format")
		c.SendEvent(ws.EventError, ws.ErrorPayload{
			Code:    "invalid_message",
			Message: "Invalid message format",
		})
		return
	}

	switch msg.Type {
	case ws.MessageChat:
		var payload ws.ChatPayload
		if err := remarshal(msg.Payload, &payload); err != nil {
			c.SendEvent(ws.EventError, ws.ErrorPayload{
				Code:    "invalid_payload",
				Message: "Invalid chat payload",
			})
			return
		}
		c.handler.HandleChat(c, &payload)

	case ws.MessageQueryResult:
		var payload ws.QueryResultPayload
		if err := remarshal(msg.Payload, &payload); err != nil {
			c.SendEvent(ws.EventError, ws.ErrorPayload{
				Code:    "invalid_payload",
				Message: "Invalid query result payload",
			})
			return
		}
		c.DeliverQueryResult(&payload)

	default:
		c.SendEvent(ws.EventError, ws.ErrorPayload{
			Code:    "unknown_message_type",
			Message: "Unknown message type",
		})
	}
}

// remarshal converts interface{} to a typed struct
func remarshal(src, dst any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}
