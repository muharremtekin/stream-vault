package websocket

import "time"

// Message represents a notification delivered over WebSocket.
type Message struct {
	Type      string    `json:"type"`
	ID        string    `json:"id"`
	Category  string    `json:"category"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Icon      string    `json:"icon,omitempty"`
	Action    string    `json:"action,omitempty"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"createdAt"`
}
