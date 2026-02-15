package dispatcher

import (
	"github.com/rs/zerolog/log"
)

// PushSender is a mock FCM push notification sender.
type PushSender struct{}

// NewPushSender creates a new PushSender.
func NewPushSender() *PushSender {
	return &PushSender{}
}

// Send simulates sending a push notification by logging the event.
func (s *PushSender) Send(userID, title, body string) error {
	log.Info().
		Str("user_id", userID).
		Str("title", title).
		Str("body", body).
		Msg("push notification sent (mock)")
	return nil
}
