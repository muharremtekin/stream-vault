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

// CatalogConsumer consumes catalog events and updates content features.
type CatalogConsumer struct {
	base  *baseConsumer
	repo  *repository.Repository
	cache *cache.Cache
}

// NewCatalogConsumer creates a new CatalogConsumer.
func NewCatalogConsumer(conn *amqp.Connection, cfg config.RabbitMQConfig, repo *repository.Repository, cache *cache.Cache) (*CatalogConsumer, error) {
	base, err := newBaseConsumer(conn, cfg.CatalogQueue, cfg.Prefetch)
	if err != nil {
		return nil, err
	}
	return &CatalogConsumer{base: base, repo: repo, cache: cache}, nil
}

// Start begins consuming messages with reconnection logic.
func (c *CatalogConsumer) Start(ctx context.Context) {
	for {
		if err := c.consumeLoop(ctx); err != nil {
			log.Error().Err(err).Msg("catalog consumer loop exited with error")
		}

		select {
		case <-c.base.done:
			log.Info().Msg("catalog consumer stopping (done signal)")
			return
		case <-ctx.Done():
			log.Info().Msg("catalog consumer stopping (context cancelled)")
			return
		default:
		}

		if !c.base.reconnect(ctx) {
			return
		}
	}
}

// Stop signals the consumer to stop.
func (c *CatalogConsumer) Stop() {
	c.base.stop()
}

func (c *CatalogConsumer) consumeLoop(ctx context.Context) error {
	msgs, err := c.base.consume("recommendation-catalog-consumer")
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Info().Str("queue", c.base.queue).Msg("catalog consumer started")

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

func (c *CatalogConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	var event model.Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal catalog event envelope")
		msg.Nack(false, false)
		return
	}

	log.Info().
		Str("event_id", event.EventID).
		Str("event_type", event.EventType).
		Str("correlation_id", event.CorrelationID).
		Msg("received catalog event")

	var data model.ContentEventData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		log.Error().Err(err).Str("event_id", event.EventID).Msg("failed to unmarshal content event data")
		msg.Nack(false, false)
		return
	}

	// Build content features with computed feature vector
	cf := buildContentFeaturesFromEvent(&data)

	if err := c.repo.UpsertContentFeatures(ctx, cf); err != nil {
		log.Error().Err(err).
			Str("content_id", data.ID).
			Str("title", data.Title).
			Msg("failed to upsert content features")
		msg.Nack(false, true)
		return
	}

	// Invalidate cached features for this content
	c.cache.InvalidateContentFeatures(ctx, data.ID)

	log.Info().
		Str("content_id", data.ID).
		Str("title", data.Title).
		Int("genres", len(data.Genres)).
		Msg("content features updated from catalog event")

	msg.Ack(false)
}
