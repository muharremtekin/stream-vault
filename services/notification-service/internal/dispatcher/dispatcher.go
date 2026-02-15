package dispatcher

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/streamvault/notification-service/internal/store"
)

// Dispatcher routes notifications to the appropriate channels based on user preferences.
type Dispatcher struct {
	notifStore  store.NotificationStore
	prefStore   store.PreferencesStore
	emailSender *EmailSender
	pushSender  *PushSender
	inAppSender *InAppSender
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher(
	notifStore store.NotificationStore,
	prefStore store.PreferencesStore,
	emailSender *EmailSender,
	pushSender *PushSender,
	inAppSender *InAppSender,
) *Dispatcher {
	return &Dispatcher{
		notifStore:  notifStore,
		prefStore:   prefStore,
		emailSender: emailSender,
		pushSender:  pushSender,
		inAppSender: inAppSender,
	}
}

// DispatchRequest describes a notification to be dispatched.
type DispatchRequest struct {
	UserID        string
	Type          string
	Category      string
	Title         string
	Body          string
	Icon          string
	Action        string
	Channels      []string // "email", "push", "inapp"
	EmailSubject  string
	EmailHTMLBody string
}

// Dispatch creates a notification, saves it to the store, and sends it via the requested channels
// filtered by user preferences.
func (d *Dispatcher) Dispatch(ctx context.Context, req *DispatchRequest) (*store.Notification, error) {
	prefs, err := d.prefStore.Get(ctx, req.UserID)
	if err != nil {
		log.Error().Err(err).Str("user_id", req.UserID).Msg("failed to load preferences, using defaults")
		prefs = nil
	}
	if prefs == nil {
		prefs = store.DefaultPreferences(req.UserID)
	}

	now := time.Now().UTC()
	var channelStatuses []store.ChannelStatus

	for _, ch := range req.Channels {
		if !d.isChannelEnabled(prefs, ch, req.Category) {
			channelStatuses = append(channelStatuses, store.ChannelStatus{
				Channel: ch,
				Status:  "skipped",
			})
			continue
		}

		status := "sent"
		switch ch {
		case "email":
			if req.EmailHTMLBody != "" && d.emailSender != nil {
				if err := d.emailSender.Send(req.UserID+"@streamvault.com", req.EmailSubject, req.EmailHTMLBody); err != nil {
					status = "failed"
				}
			} else {
				status = "skipped"
			}
		case "push":
			if d.pushSender != nil {
				if err := d.pushSender.Send(req.UserID, req.Title, req.Body); err != nil {
					status = "failed"
				}
			}
		case "inapp":
			// In-app delivery happens after the notification is saved (below).
			status = "pending"
		}

		cs := store.ChannelStatus{
			Channel: ch,
			Status:  status,
		}
		if status == "sent" {
			cs.SentAt = now
		}
		channelStatuses = append(channelStatuses, cs)
	}

	notification := &store.Notification{
		UserID:   req.UserID,
		Type:     req.Type,
		Category: req.Category,
		Title:    req.Title,
		Body:     req.Body,
		Icon:     req.Icon,
		Action:   req.Action,
		Channels: channelStatuses,
	}

	if err := d.notifStore.Create(ctx, notification); err != nil {
		return nil, err
	}

	// Deliver in-app notification via WebSocket after save (so we have the ID).
	for i, cs := range channelStatuses {
		if cs.Channel == "inapp" && cs.Status == "pending" {
			d.inAppSender.Send(req.UserID, notification)
			channelStatuses[i].Status = "sent"
			channelStatuses[i].SentAt = time.Now().UTC()
		}
	}

	log.Info().
		Str("user_id", req.UserID).
		Str("type", req.Type).
		Str("category", req.Category).
		Int("channels", len(req.Channels)).
		Msg("notification dispatched")

	return notification, nil
}

// BroadcastDispatch creates and broadcasts a notification to all connected users.
// The notification is not saved per-user (only sent via WebSocket).
func (d *Dispatcher) BroadcastDispatch(ctx context.Context, req *DispatchRequest) {
	notification := &store.Notification{
		Type:      req.Type,
		Category:  req.Category,
		Title:     req.Title,
		Body:      req.Body,
		Icon:      req.Icon,
		Action:    req.Action,
		CreatedAt: time.Now().UTC(),
	}

	d.inAppSender.Broadcast(notification)

	log.Info().
		Str("type", req.Type).
		Str("category", req.Category).
		Msg("notification broadcast to all users")
}

func (d *Dispatcher) isChannelEnabled(prefs *store.NotificationPreferences, channel, category string) bool {
	var pref store.ChannelPreference
	switch channel {
	case "email":
		pref = prefs.Email
	case "push":
		pref = prefs.Push
	case "inapp":
		pref = prefs.InApp
	default:
		return false
	}

	if !pref.Enabled {
		return false
	}

	if len(pref.Categories) > 0 {
		for _, c := range pref.Categories {
			if c == category {
				return true
			}
		}
		return false
	}

	return true
}
