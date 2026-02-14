package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"

	"github.com/streamvault/recommendation-service/internal/cache"
	"github.com/streamvault/recommendation-service/internal/config"
	"github.com/streamvault/recommendation-service/internal/model"
	"github.com/streamvault/recommendation-service/internal/repository"
)

// WatchConsumer consumes watch.completed events and records interactions.
type WatchConsumer struct {
	base  *baseConsumer
	repo  *repository.Repository
	cache *cache.Cache
}

// NewWatchConsumer creates a new WatchConsumer.
func NewWatchConsumer(conn *amqp.Connection, cfg config.RabbitMQConfig, repo *repository.Repository, cache *cache.Cache) (*WatchConsumer, error) {
	base, err := newBaseConsumer(conn, cfg.WatchQueue, cfg.Prefetch)
	if err != nil {
		return nil, err
	}
	return &WatchConsumer{base: base, repo: repo, cache: cache}, nil
}

// Start begins consuming messages with reconnection logic.
func (c *WatchConsumer) Start(ctx context.Context) {
	for {
		if err := c.consumeLoop(ctx); err != nil {
			log.Error().Err(err).Msg("watch consumer loop exited with error")
		}

		select {
		case <-c.base.done:
			log.Info().Msg("watch consumer stopping (done signal)")
			return
		case <-ctx.Done():
			log.Info().Msg("watch consumer stopping (context cancelled)")
			return
		default:
		}

		if !c.base.reconnect(ctx) {
			return
		}
	}
}

// Stop signals the consumer to stop.
func (c *WatchConsumer) Stop() {
	c.base.stop()
}

func (c *WatchConsumer) consumeLoop(ctx context.Context) error {
	msgs, err := c.base.consume("recommendation-watch-consumer")
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Info().Str("queue", c.base.queue).Msg("watch consumer started")

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

func (c *WatchConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	var event model.Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal watch event envelope")
		msg.Nack(false, false)
		return
	}

	log.Info().
		Str("event_id", event.EventID).
		Str("event_type", event.EventType).
		Str("correlation_id", event.CorrelationID).
		Msg("received watch event")

	var data model.WatchEventData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		log.Error().Err(err).Str("event_id", event.EventID).Msg("failed to unmarshal watch event data")
		msg.Nack(false, false)
		return
	}

	// Determine interaction type based on completion percentage
	interactionType := model.InteractionView
	completionPct := data.CompletionPct
	if completionPct >= 90 {
		interactionType = model.InteractionComplete
		completionPct = 100
	}

	interaction := &model.Interaction{
		UserID:          data.UserID,
		ContentID:       data.ContentID,
		InteractionType: interactionType,
		CompletionPct:   completionPct,
	}

	if err := c.repo.CreateInteraction(ctx, interaction); err != nil {
		log.Error().Err(err).
			Str("user_id", data.UserID).
			Str("content_id", data.ContentID).
			Msg("failed to create watch interaction")
		msg.Nack(false, true)
		return
	}

	// Refresh user profile asynchronously (best-effort)
	refreshUserProfile(ctx, c.repo, c.cache, data.UserID)

	// Invalidate cached recommendations
	c.cache.InvalidateRecommendations(ctx, data.UserID)

	log.Info().
		Str("user_id", data.UserID).
		Str("content_id", data.ContentID).
		Str("type", string(interactionType)).
		Float64("completion_pct", completionPct).
		Msg("watch interaction recorded")

	msg.Ack(false)
}
