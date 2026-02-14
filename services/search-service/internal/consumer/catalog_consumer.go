package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"

	"github.com/streamvault/search-service/internal/config"
	"github.com/streamvault/search-service/internal/elasticsearch"
	"github.com/streamvault/search-service/internal/model"
	"github.com/streamvault/search-service/internal/trending"
)

// CatalogConsumer consumes catalog events and syncs content to Elasticsearch.
type CatalogConsumer struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	indexer     *elasticsearch.Indexer
	trendingSvc *trending.Service
	queueName   string
	prefetch    int
	done        chan struct{}
}

// NewCatalogConsumer creates a new CatalogConsumer.
func NewCatalogConsumer(conn *amqp.Connection, cfg config.RabbitMQConfig, indexer *elasticsearch.Indexer, trendingSvc *trending.Service) (*CatalogConsumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("opening rabbitmq channel for catalog consumer: %w", err)
	}

	if err := ch.Qos(cfg.Prefetch, 0, false); err != nil {
		return nil, fmt.Errorf("setting catalog consumer prefetch: %w", err)
	}

	return &CatalogConsumer{
		conn:        conn,
		channel:     ch,
		indexer:     indexer,
		trendingSvc: trendingSvc,
		queueName:   cfg.CatalogQueue,
		prefetch:    cfg.Prefetch,
		done:        make(chan struct{}),
	}, nil
}

// Start begins consuming messages with reconnection logic.
func (c *CatalogConsumer) Start(ctx context.Context) {
	for {
		if err := c.consumeLoop(ctx); err != nil {
			log.Error().Err(err).Msg("catalog consumer loop exited with error")
		}

		select {
		case <-c.done:
			log.Info().Msg("catalog consumer stopping (done signal)")
			return
		case <-ctx.Done():
			log.Info().Msg("catalog consumer stopping (context cancelled)")
			return
		default:
		}

		if !c.reconnect(ctx) {
			return
		}
	}
}

// Stop signals the consumer to stop.
func (c *CatalogConsumer) Stop() {
	close(c.done)
	if c.channel != nil {
		c.channel.Close()
	}
}

func (c *CatalogConsumer) consumeLoop(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.queueName,
		"search-catalog-consumer",
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Info().Str("queue", c.queueName).Msg("catalog consumer started")

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("delivery channel closed")
			}
			c.handleMessage(ctx, msg)
		case <-c.done:
			return nil
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *CatalogConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	var event model.Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal catalog event")
		msg.Nack(false, false)
		return
	}

	log.Info().
		Str("event_id", event.EventID).
		Str("event_type", event.EventType).
		Str("source", event.Source).
		Str("correlation_id", event.CorrelationID).
		Msg("received catalog event")

	switch event.EventType {
	case "catalog.content.created", "catalog.content.updated":
		c.handleContentUpsert(ctx, event, msg)
	case "catalog.content.deleted":
		c.handleContentDelete(ctx, event, msg)
	default:
		log.Warn().Str("event_type", event.EventType).Msg("unknown catalog event type, skipping")
		msg.Ack(false)
	}
}

func (c *CatalogConsumer) handleContentUpsert(ctx context.Context, event model.Event, msg amqp.Delivery) {
	var data model.ContentEventData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		log.Error().Err(err).Str("event_id", event.EventID).Msg("failed to unmarshal content data")
		msg.Nack(false, false)
		return
	}

	doc := data.ToSearchDocument()
	if err := c.indexer.IndexDocument(ctx, doc); err != nil {
		log.Error().Err(err).Str("content_id", data.ID).Msg("failed to index document")
		msg.Nack(false, true)
		return
	}

	c.trendingSvc.SetContentMeta(ctx, data.ID, data.Title, data.ThumbnailURL, data.ContentType, data.ReleaseYear, data.AverageRating)

	log.Info().Str("content_id", data.ID).Str("title", data.Title).Msg("content indexed")
	msg.Ack(false)
}

func (c *CatalogConsumer) handleContentDelete(ctx context.Context, event model.Event, msg amqp.Delivery) {
	var data model.DeleteEventData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		log.Error().Err(err).Str("event_id", event.EventID).Msg("failed to unmarshal delete data")
		msg.Nack(false, false)
		return
	}

	if err := c.indexer.DeleteDocument(ctx, data.ID); err != nil {
		log.Error().Err(err).Str("content_id", data.ID).Msg("failed to delete document")
		msg.Nack(false, true)
		return
	}

	log.Info().Str("content_id", data.ID).Msg("content deleted from index")
	msg.Ack(false)
}

func (c *CatalogConsumer) reconnect(ctx context.Context) bool {
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-c.done:
			return false
		case <-ctx.Done():
			return false
		case <-time.After(backoff):
		}

		log.Info().Dur("backoff", backoff).Msg("attempting catalog consumer channel recreation")

		ch, err := c.conn.Channel()
		if err != nil {
			log.Error().Err(err).Msg("failed to recreate catalog consumer channel")
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		if err := ch.Qos(c.prefetch, 0, false); err != nil {
			log.Error().Err(err).Msg("failed to set catalog consumer prefetch on new channel")
			ch.Close()
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		if c.channel != nil {
			c.channel.Close()
		}
		c.channel = ch
		log.Info().Msg("catalog consumer channel recreated successfully")
		return true
	}
}
