package dispatcher

import (
	"github.com/rs/zerolog/log"

	"github.com/streamvault/notification-service/internal/store"
	ws "github.com/streamvault/notification-service/internal/websocket"
)

// InAppSender delivers notifications via WebSocket to connected users.
type InAppSender struct {
	hub *ws.Hub
}

// NewInAppSender creates a new InAppSender.
func NewInAppSender(hub *ws.Hub) *InAppSender {
	return &InAppSender{hub: hub}
}

// Send delivers a notification to the user via WebSocket if they are connected.
func (s *InAppSender) Send(userID string, n *store.Notification) {
	msg := &ws.Message{
		Type:      n.Type,
		ID:        n.ID.Hex(),
		Category:  n.Category,
		Title:     n.Title,
		Body:      n.Body,
		Icon:      n.Icon,
		Action:    n.Action,
		Read:      n.Read,
		CreatedAt: n.CreatedAt,
	}
	s.hub.SendToUser(userID, msg)
	log.Debug().Str("user_id", userID).Str("notification_id", n.ID.Hex()).Msg("in-app notification sent")
}

// Broadcast delivers a notification to all connected users.
func (s *InAppSender) Broadcast(n *store.Notification) {
	msg := &ws.Message{
		Type:      n.Type,
		ID:        n.ID.Hex(),
		Category:  n.Category,
		Title:     n.Title,
		Body:      n.Body,
		Icon:      n.Icon,
		Action:    n.Action,
		Read:      n.Read,
		CreatedAt: n.CreatedAt,
	}
	s.hub.Broadcast(msg)
	log.Debug().Str("notification_id", n.ID.Hex()).Msg("in-app notification broadcast")
}
