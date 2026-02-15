package model

import "encoding/json"

// Event is the standard camelCase envelope used by catalog-service events.
type Event struct {
	EventID       string          `json:"eventId"`
	EventType     string          `json:"eventType"`
	Timestamp     string          `json:"timestamp"`
	Source        string          `json:"source"`
	CorrelationID string          `json:"correlationId"`
	Data          json.RawMessage `json:"data"`
}

// ContentCreatedData is the data payload inside a catalog content.created event.
type ContentCreatedData struct {
	ContentID   string   `json:"contentId"`
	ContentType string   `json:"contentType"`
	Title       string   `json:"title"`
	Genres      []string `json:"genres"`
	ReleaseYear int      `json:"releaseYear"`
	Director    string   `json:"director,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// Subscription events are PascalCase flat records (no envelope).

// SubscriptionCreatedData is the payload for subscription.created events.
type SubscriptionCreatedData struct {
	SubscriptionID string `json:"SubscriptionId"`
	UserID         string `json:"UserId"`
	PlanID         string `json:"PlanId"`
	PlanName       string `json:"PlanName"`
	Tier           string `json:"Tier"`
	PeriodStart    string `json:"PeriodStart"`
	PeriodEnd      string `json:"PeriodEnd"`
}

// SubscriptionCancelledData is the payload for subscription.cancelled events.
type SubscriptionCancelledData struct {
	SubscriptionID string `json:"SubscriptionId"`
	UserID         string `json:"UserId"`
	CancelledAt    string `json:"CancelledAt"`
	PeriodEnd      string `json:"PeriodEnd"`
}

// PlanChangedData is the payload for plan.changed events.
type PlanChangedData struct {
	SubscriptionID string `json:"SubscriptionId"`
	UserID         string `json:"UserId"`
	OldPlanID      string `json:"OldPlanId"`
	OldTier        string `json:"OldTier"`
	NewPlanID      string `json:"NewPlanId"`
	NewTier        string `json:"NewTier"`
	ChangedAt      string `json:"ChangedAt"`
}

// Encoding events are snake_case flat records (envelope + data merged).

// EncodingResultEvent is the payload for job.completed and job.failed events.
type EncodingResultEvent struct {
	EventID       string `json:"event_id"`
	EventType     string `json:"event_type"`
	Timestamp     string `json:"timestamp"`
	Source        string `json:"source"`
	CorrelationID string `json:"correlation_id"`
	JobID         string `json:"job_id"`
	ContentID     string `json:"content_id"`
	Status        string `json:"status"`
	ErrorMessage  string `json:"error_message,omitempty"`
	CompletedAt   string `json:"completed_at"`
}
