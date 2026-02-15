package websocket

import (
	"encoding/json"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/streamvault/notification-service/internal/metrics"
)

// Hub manages WebSocket client connections grouped by user ID.
type Hub struct {
	clients    map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	maxPerUser int
}

// NewHub creates a new Hub.
func NewHub(maxConnectionsPerUser int) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		maxPerUser: maxConnectionsPerUser,
	}
}

// Register returns the registration channel for adding clients.
func (h *Hub) Register() chan<- *Client {
	return h.register
}

// Run processes register/unregister events. Must be run as a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			if len(h.clients[client.UserID]) >= h.maxPerUser {
				h.mu.Unlock()
				log.Warn().
					Str("user_id", client.UserID).
					Int("max", h.maxPerUser).
					Msg("max websocket connections per user reached, closing new connection")
				close(client.send)
				client.conn.Close()
				continue
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()
			metrics.NotificationWSActiveConnections.Inc()
			log.Debug().
				Str("user_id", client.UserID).
				Int("total", h.ClientCount()).
				Msg("websocket client registered")

		case client := <-h.unregister:
			h.mu.Lock()
			if conns, ok := h.clients[client.UserID]; ok {
				if _, exists := conns[client]; exists {
					delete(conns, client)
					close(client.send)
					if len(conns) == 0 {
						delete(h.clients, client.UserID)
					}
					metrics.NotificationWSActiveConnections.Dec()
				}
			}
			h.mu.Unlock()
			log.Debug().
				Str("user_id", client.UserID).
				Int("total", h.ClientCount()).
				Msg("websocket client unregistered")
		}
	}
}

// SendToUser sends a message to all connections of a specific user.
func (h *Hub) SendToUser(userID string, msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("failed to marshal ws message")
		return
	}

	h.mu.RLock()
	conns := h.clients[userID]
	h.mu.RUnlock()

	for client := range conns {
		select {
		case client.send <- data:
			metrics.NotificationWSMessagesSentTotal.Inc()
		default:
			h.mu.Lock()
			if _, ok := h.clients[userID]; ok {
				delete(h.clients[userID], client)
				close(client.send)
				if len(h.clients[userID]) == 0 {
					delete(h.clients, userID)
				}
			}
			h.mu.Unlock()
		}
	}
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal broadcast ws message")
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, conns := range h.clients {
		for client := range conns {
			select {
			case client.send <- data:
				metrics.NotificationWSMessagesSentTotal.Inc()
			default:
			}
		}
	}
}

// ClientCount returns the total number of active connections.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for _, conns := range h.clients {
		count += len(conns)
	}
	return count
}

// UserClientCount returns the number of active connections for a user.
func (h *Hub) UserClientCount(userID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[userID])
}
