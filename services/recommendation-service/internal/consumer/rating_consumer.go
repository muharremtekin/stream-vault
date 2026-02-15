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

	"github.com/streamvault/recommendation-service/internal/cache"
	"github.com/streamvault/recommendation-service/internal/config"
	"github.com/streamvault/recommendation-service/internal/model"
	"github.com/streamvault/recommendation-service/internal/repository"
	"github.com/streamvault/recommendation-service/internal/telemetry"
)

// RatingConsumer consumes content.rated events and records rating interactions.
type RatingConsumer struct {
	base  *baseConsumer
	repo  *repository.Repository
	cache *cache.Cache
}

// NewRatingConsumer creates a new RatingConsumer.
func NewRatingConsumer(conn *amqp.Connection, cfg config.RabbitMQConfig, repo *repository.Repository, cache *cache.Cache) (*RatingConsumer, error) {
	base, err := newBaseConsumer(conn, cfg.RatingsQueue, cfg.Prefetch)
	if err != nil {
		return nil, err
	}
	return &RatingConsumer{base: base, repo: repo, cache: cache}, nil
}

// Start begins consuming messages with reconnection logic.
func (c *RatingConsumer) Start(ctx context.Context) {
	for {
		if err := c.consumeLoop(ctx); err != nil {
			log.Error().Err(err).Msg("rating consumer loop exited with error")
		}

		select {
		case <-c.base.done:
			log.Info().Msg("rating consumer stopping (done signal)")
			return
		case <-ctx.Done():
			log.Info().Msg("rating consumer stopping (context cancelled)")
			return
		default:
		}

		if !c.base.reconnect(ctx) {
			return
		}
	}
}

// Stop signals the consumer to stop.
func (c *RatingConsumer) Stop() {
	c.base.stop()
}

func (c *RatingConsumer) consumeLoop(ctx context.Context) error {
	msgs, err := c.base.consume("recommendation-rating-consumer")
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Info().Str("queue", c.base.queue).Msg("rating consumer started")

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

var ratingTracer = otel.Tracer("recommendation-service/consumer/rating")

func (c *RatingConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	ctx = telemetry.ExtractAMQP(ctx, msg.Headers)
	ctx, span := ratingTracer.Start(ctx, "rabbitmq.consume",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.source.name", c.base.queue),
			attribute.String("messaging.operation", "receive"),
		),
	)
	defer span.End()

	var event model.Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal rating event envelope")
		msg.Nack(false, false)
		return
	}

	log.Info().
		Str("event_id", event.EventID).
		Str("event_type", event.EventType).
		Str("correlation_id", event.CorrelationID).
		Msg("received rating event")

	var data model.RatingEventData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		log.Error().Err(err).Str("event_id", event.EventID).Msg("failed to unmarshal rating event data")
		msg.Nack(false, false)
		return
	}

	rating := data.Rating
	interaction := &model.Interaction{
		UserID:          data.UserID,
		ContentID:       data.ContentID,
		InteractionType: model.InteractionRate,
		Rating:          &rating,
	}

	if err := c.repo.CreateInteraction(ctx, interaction); err != nil {
		log.Error().Err(err).
			Str("user_id", data.UserID).
			Str("content_id", data.ContentID).
			Msg("failed to create rating interaction")
		msg.Nack(false, true)
		return
	}

	// Refresh user profile
	refreshUserProfile(ctx, c.repo, c.cache, data.UserID)

	// Invalidate cached recommendations
	c.cache.InvalidateRecommendations(ctx, data.UserID)

	log.Info().
		Str("user_id", data.UserID).
		Str("content_id", data.ContentID).
		Float64("rating", data.Rating).
		Msg("rating interaction recorded")

	msg.Ack(false)
}
