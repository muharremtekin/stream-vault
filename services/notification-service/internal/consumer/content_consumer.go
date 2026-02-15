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

var contentTracer = otel.Tracer("notification-service/consumer/content")

// ContentConsumer consumes catalog content events and dispatches notifications.
type ContentConsumer struct {
	base       *baseConsumer
	dispatcher *dispatcher.Dispatcher
	tmpl       *template.Engine
}

// NewContentConsumer creates a new ContentConsumer.
func NewContentConsumer(conn *amqp.Connection, queue string, prefetch int, d *dispatcher.Dispatcher, tmpl *template.Engine) (*ContentConsumer, error) {
	base, err := newBaseConsumer(conn, queue, prefetch)
	if err != nil {
		return nil, err
	}
	return &ContentConsumer{base: base, dispatcher: d, tmpl: tmpl}, nil
}

// Start begins consuming messages with reconnection logic.
func (c *ContentConsumer) Start(ctx context.Context) {
	for {
		if err := c.consumeLoop(ctx); err != nil {
			log.Error().Err(err).Msg("content consumer loop exited with error")
		}

		select {
		case <-c.base.done:
			log.Info().Msg("content consumer stopping (done signal)")
			return
		case <-ctx.Done():
			log.Info().Msg("content consumer stopping (context cancelled)")
			return
		default:
		}

		if !c.base.reconnect(ctx) {
			return
		}
	}
}

// Stop signals the consumer to stop.
func (c *ContentConsumer) Stop() {
	c.base.stop()
}

func (c *ContentConsumer) consumeLoop(ctx context.Context) error {
	msgs, err := c.base.consume("notification-content-consumer")
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Info().Str("queue", c.base.queue).Msg("content consumer started")

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

func (c *ContentConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	ctx = telemetry.ExtractAMQP(ctx, msg.Headers)
	ctx, span := contentTracer.Start(ctx, "rabbitmq.consume",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.source.name", c.base.queue),
		),
	)
	defer span.End()

	// Catalog events use the standard camelCase envelope with a data field.
	var event model.Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal content event envelope")
		msg.Nack(false, false)
		return
	}

	log.Info().
		Str("event_id", event.EventID).
		Str("event_type", event.EventType).
		Str("correlation_id", event.CorrelationID).
		Msg("received content event")

	if event.EventType != "content.created" {
		log.Warn().Str("event_type", event.EventType).Msg("unexpected content event type")
		msg.Nack(false, false)
		return
	}

	var data model.ContentCreatedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		log.Error().Err(err).Str("event_id", event.EventID).Msg("failed to unmarshal content.created data")
		msg.Nack(false, false)
		return
	}

	// Content notifications are broadcast to all connected users via WebSocket
	// and optionally via push. There is no specific target user.
	c.dispatcher.BroadcastDispatch(ctx, &dispatcher.DispatchRequest{
		Type:     "content.created",
		Category: "content",
		Title:    "Yeni İçerik",
		Body:     fmt.Sprintf("%s platformumuza eklendi!", data.Title),
		Icon:     "film",
		Action:   fmt.Sprintf("/content/%s", data.ContentID),
		Channels: []string{"push", "inapp"},
	})

	msg.Ack(false)
}
