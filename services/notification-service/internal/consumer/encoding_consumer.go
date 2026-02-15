package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/streamvault/notification-service/internal/dispatcher"
	"github.com/streamvault/notification-service/internal/model"
	"github.com/streamvault/notification-service/internal/telemetry"
	"github.com/streamvault/notification-service/internal/template"
)

var encodingTracer = otel.Tracer("notification-service/consumer/encoding")

// EncodingConsumer consumes encoding job result events and dispatches notifications.
type EncodingConsumer struct {
	base       *baseConsumer
	dispatcher *dispatcher.Dispatcher
	tmpl       *template.Engine
}

// NewEncodingConsumer creates a new EncodingConsumer.
func NewEncodingConsumer(conn *amqp.Connection, queue string, prefetch int, d *dispatcher.Dispatcher, tmpl *template.Engine) (*EncodingConsumer, error) {
	base, err := newBaseConsumer(conn, queue, prefetch)
	if err != nil {
		return nil, err
	}
	return &EncodingConsumer{base: base, dispatcher: d, tmpl: tmpl}, nil
}

// Start begins consuming messages with reconnection logic.
func (c *EncodingConsumer) Start(ctx context.Context) {
	for {
		if err := c.consumeLoop(ctx); err != nil {
			log.Error().Err(err).Msg("encoding consumer loop exited with error")
		}

		select {
		case <-c.base.done:
			log.Info().Msg("encoding consumer stopping (done signal)")
			return
		case <-ctx.Done():
			log.Info().Msg("encoding consumer stopping (context cancelled)")
			return
		default:
		}

		if !c.base.reconnect(ctx) {
			return
		}
	}
}

// Stop signals the consumer to stop.
func (c *EncodingConsumer) Stop() {
	c.base.stop()
}

func (c *EncodingConsumer) consumeLoop(ctx context.Context) error {
	msgs, err := c.base.consume("notification-encoding-consumer")
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Info().Str("queue", c.base.queue).Msg("encoding consumer started")

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("delivery channel closed")
			}
			c.handleMessage(ctx, msg)
		case <-c.base.done:
			return nil
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *EncodingConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	ctx = telemetry.ExtractAMQP(ctx, msg.Headers)
	ctx, span := encodingTracer.Start(ctx, "rabbitmq.consume",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.source.name", c.base.queue),
		),
	)
	defer span.End()

	var event model.EncodingResultEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal encoding event")
		msg.Nack(false, false)
		return
	}

	log.Info().
		Str("event_id", event.EventID).
		Str("event_type", event.EventType).
		Str("job_id", event.JobID).
		Str("status", event.Status).
		Msg("received encoding event")

	routingKey := msg.RoutingKey

	switch routingKey {
	case "job.completed":
		c.handleJobCompleted(ctx, msg, &event)
	case "job.failed":
		c.handleJobFailed(ctx, msg, &event)
	default:
		log.Warn().Str("routing_key", routingKey).Msg("unknown encoding event type")
		msg.Nack(false, false)
	}
}

func (c *EncodingConsumer) handleJobCompleted(ctx context.Context, msg amqp.Delivery, event *model.EncodingResultEvent) {
	emailBody, _ := c.tmpl.Render("encoding_complete.html", map[string]string{
		"Status":      "tamamlandı",
		"JobID":       event.JobID,
		"ContentID":   event.ContentID,
		"CompletedAt": event.CompletedAt,
	})

	// Encoding notifications are admin-only. The encoding service doesn't carry a target user,
	// so we send to a well-known admin user ID or broadcast to admins.
	_, err := c.dispatcher.Dispatch(ctx, &dispatcher.DispatchRequest{
		UserID:        "admin",
		Type:          "encoding.completed",
		Category:      "encoding",
		Title:         "Video Hazır",
		Body:          fmt.Sprintf("İçerik %s için encoding tamamlandı.", event.ContentID),
		Icon:          "film",
		Channels:      []string{"inapp"},
		EmailSubject:  "StreamVault - Encoding Tamamlandı",
		EmailHTMLBody: emailBody,
	})
	if err != nil {
		log.Error().Err(err).Str("job_id", event.JobID).Msg("failed to dispatch job.completed notification")
		msg.Nack(false, true)
		return
	}

	msg.Ack(false)
}

func (c *EncodingConsumer) handleJobFailed(ctx context.Context, msg amqp.Delivery, event *model.EncodingResultEvent) {
	emailBody, _ := c.tmpl.Render("encoding_complete.html", map[string]string{
		"Status":       "başarısız",
		"JobID":        event.JobID,
		"ContentID":    event.ContentID,
		"CompletedAt":  event.CompletedAt,
		"ErrorMessage": event.ErrorMessage,
	})

	_, err := c.dispatcher.Dispatch(ctx, &dispatcher.DispatchRequest{
		UserID:        "admin",
		Type:          "encoding.failed",
		Category:      "encoding",
		Title:         "Encoding Hatası",
		Body:          fmt.Sprintf("İçerik %s için encoding başarısız: %s", event.ContentID, event.ErrorMessage),
		Icon:          "alert-triangle",
		Channels:      []string{"email", "inapp"},
		EmailSubject:  "StreamVault - Encoding Hatası",
		EmailHTMLBody: emailBody,
	})
	if err != nil {
		log.Error().Err(err).Str("job_id", event.JobID).Msg("failed to dispatch job.failed notification")
		msg.Nack(false, true)
		return
	}

	msg.Ack(false)
}
