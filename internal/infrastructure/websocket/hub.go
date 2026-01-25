package websocket

import (
	"sync"

	"github.com/mololab/alodb/pkg/logger"
)

type Hub struct {
	clients    map[string]*Client
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			logger.Debug().
				Str("client_id", client.ID).
				Str("session_id", client.SessionID).
				Msg("client registered")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, exists := h.clients[client.ID]; exists {
				delete(h.clients, client.ID)
				close(client.send)
				logger.Debug().
					Str("client_id", client.ID).
					Msg("client unregistered")
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) GetClient(clientID string) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client, exists := h.clients[clientID]
	return client, exists
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
