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

var subscriptionTracer = otel.Tracer("notification-service/consumer/subscription")

// SubscriptionConsumer consumes subscription events and dispatches notifications.
type SubscriptionConsumer struct {
	base       *baseConsumer
	dispatcher *dispatcher.Dispatcher
	tmpl       *template.Engine
}

// NewSubscriptionConsumer creates a new SubscriptionConsumer.
func NewSubscriptionConsumer(conn *amqp.Connection, queue string, prefetch int, d *dispatcher.Dispatcher, tmpl *template.Engine) (*SubscriptionConsumer, error) {
	base, err := newBaseConsumer(conn, queue, prefetch)
	if err != nil {
		return nil, err
	}
	return &SubscriptionConsumer{base: base, dispatcher: d, tmpl: tmpl}, nil
}

// Start begins consuming messages with reconnection logic.
func (c *SubscriptionConsumer) Start(ctx context.Context) {
	for {
		if err := c.consumeLoop(ctx); err != nil {
			log.Error().Err(err).Msg("subscription consumer loop exited with error")
		}

		select {
		case <-c.base.done:
			log.Info().Msg("subscription consumer stopping (done signal)")
			return
		case <-ctx.Done():
			log.Info().Msg("subscription consumer stopping (context cancelled)")
			return
		default:
		}

		if !c.base.reconnect(ctx) {
			return
		}
	}
}

// Stop signals the consumer to stop.
func (c *SubscriptionConsumer) Stop() {
	c.base.stop()
}

func (c *SubscriptionConsumer) consumeLoop(ctx context.Context) error {
	msgs, err := c.base.consume("notification-subscription-consumer")
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Info().Str("queue", c.base.queue).Msg("subscription consumer started")

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

func (c *SubscriptionConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	ctx = telemetry.ExtractAMQP(ctx, msg.Headers)
	ctx, span := subscriptionTracer.Start(ctx, "rabbitmq.consume",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.source.name", c.base.queue),
		),
	)
	defer span.End()

	routingKey := msg.RoutingKey
	log.Info().Str("routing_key", routingKey).Msg("received subscription event")

	switch routingKey {
	case "subscription.created":
		c.handleSubscriptionCreated(ctx, msg)
	case "subscription.cancelled":
		c.handleSubscriptionCancelled(ctx, msg)
	case "plan.changed":
		c.handlePlanChanged(ctx, msg)
	default:
		log.Warn().Str("routing_key", routingKey).Msg("unknown subscription event type")
		msg.Nack(false, false)
	}
}

func (c *SubscriptionConsumer) handleSubscriptionCreated(ctx context.Context, msg amqp.Delivery) {
	var data model.SubscriptionCreatedData
	if err := json.Unmarshal(msg.Body, &data); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal subscription.created event")
		msg.Nack(false, false)
		return
	}

	emailBody, _ := c.tmpl.Render("welcome.html", map[string]string{
		"PlanName":    data.PlanName,
		"PeriodStart": data.PeriodStart,
		"PeriodEnd":   data.PeriodEnd,
	})

	_, err := c.dispatcher.Dispatch(ctx, &dispatcher.DispatchRequest{
		UserID:        data.UserID,
		Type:          "subscription.created",
		Category:      "subscription",
		Title:         "Hoş Geldiniz!",
		Body:          fmt.Sprintf("%s planınız başarıyla aktifleştirildi.", data.PlanName),
		Icon:          "check-circle",
		Channels:      []string{"email", "inapp"},
		EmailSubject:  "StreamVault - Hoş Geldiniz!",
		EmailHTMLBody: emailBody,
	})
	if err != nil {
		log.Error().Err(err).Str("user_id", data.UserID).Msg("failed to dispatch subscription.created notification")
		msg.Nack(false, true)
		return
	}

	msg.Ack(false)
}

func (c *SubscriptionConsumer) handleSubscriptionCancelled(ctx context.Context, msg amqp.Delivery) {
	var data model.SubscriptionCancelledData
	if err := json.Unmarshal(msg.Body, &data); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal subscription.cancelled event")
		msg.Nack(false, false)
		return
	}

	_, err := c.dispatcher.Dispatch(ctx, &dispatcher.DispatchRequest{
		UserID:   data.UserID,
		Type:     "subscription.cancelled",
		Category: "subscription",
		Title:    "Abonelik İptal Edildi",
		Body:     fmt.Sprintf("Aboneliğiniz iptal edildi. %s tarihine kadar erişiminiz devam edecektir.", data.PeriodEnd),
		Icon:     "x-circle",
		Channels: []string{"email", "inapp"},
	})
	if err != nil {
		log.Error().Err(err).Str("user_id", data.UserID).Msg("failed to dispatch subscription.cancelled notification")
		msg.Nack(false, true)
		return
	}

	msg.Ack(false)
}

func (c *SubscriptionConsumer) handlePlanChanged(ctx context.Context, msg amqp.Delivery) {
	var data model.PlanChangedData
	if err := json.Unmarshal(msg.Body, &data); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal plan.changed event")
		msg.Nack(false, false)
		return
	}

	_, err := c.dispatcher.Dispatch(ctx, &dispatcher.DispatchRequest{
		UserID:   data.UserID,
		Type:     "plan.changed",
		Category: "subscription",
		Title:    "Plan Değişikliği",
		Body:     fmt.Sprintf("Planınız %s'den %s'e değiştirildi.", data.OldTier, data.NewTier),
		Icon:     "refresh-cw",
		Channels: []string{"email", "inapp"},
	})
	if err != nil {
		log.Error().Err(err).Str("user_id", data.UserID).Msg("failed to dispatch plan.changed notification")
		msg.Nack(false, true)
		return
	}

	msg.Ack(false)
}
