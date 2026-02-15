package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/streamvault/search-service/internal/config"
	"github.com/streamvault/search-service/internal/elasticsearch"
	"github.com/streamvault/search-service/internal/model"
	"github.com/streamvault/search-service/internal/telemetry"
	"github.com/streamvault/search-service/internal/trending"
)

var watchTracer = otel.Tracer("search-service/consumer/watch")

// WatchConsumer consumes watch events and updates view counts and trending data.
type WatchConsumer struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	indexer     *elasticsearch.Indexer
	redisClient *redis.Client
	trendingSvc *trending.Service
	queueName   string
	prefetch    int
	done        chan struct{}
}

// NewWatchConsumer creates a new WatchConsumer.
func NewWatchConsumer(conn *amqp.Connection, cfg config.RabbitMQConfig, indexer *elasticsearch.Indexer, redisClient *redis.Client, trendingSvc *trending.Service) (*WatchConsumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("opening rabbitmq channel for watch consumer: %w", err)
	}

	if err := ch.Qos(cfg.Prefetch, 0, false); err != nil {
		return nil, fmt.Errorf("setting watch consumer prefetch: %w", err)
	}

	return &WatchConsumer{
		conn:        conn,
		channel:     ch,
		indexer:     indexer,
		redisClient: redisClient,
		trendingSvc: trendingSvc,
		queueName:   cfg.WatchQueue,
		prefetch:    cfg.Prefetch,
		done:        make(chan struct{}),
	}, nil
}

// Start begins consuming messages with reconnection logic.
func (c *WatchConsumer) Start(ctx context.Context) {
	for {
		if err := c.consumeLoop(ctx); err != nil {
			log.Error().Err(err).Msg("watch consumer loop exited with error")
		}

		select {
		case <-c.done:
			log.Info().Msg("watch consumer stopping (done signal)")
			return
		case <-ctx.Done():
			log.Info().Msg("watch consumer stopping (context cancelled)")
			return
		default:
		}

		if !c.reconnect(ctx) {
			return
		}
	}
}

// Stop signals the consumer to stop.
func (c *WatchConsumer) Stop() {
	close(c.done)
	if c.channel != nil {
		c.channel.Close()
	}
}

func (c *WatchConsumer) consumeLoop(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.queueName,
		"search-watch-consumer",
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Info().Str("queue", c.queueName).Msg("watch consumer started")

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

func (c *WatchConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	// Extract trace context from AMQP headers
	ctx = telemetry.ExtractAMQP(ctx, msg.Headers)
	ctx, span := watchTracer.Start(ctx, "rabbitmq.consume.watch",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.destination", c.queueName),
		),
	)
	defer span.End()

	var event model.Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal watch event")
		span.RecordError(err)
		msg.Nack(false, false)
		return
	}

	span.SetAttributes(
		attribute.String("event.id", event.EventID),
		attribute.String("event.type", event.EventType),
	)

	log.Info().
		Str("event_id", event.EventID).
		Str("event_type", event.EventType).
		Str("correlation_id", event.CorrelationID).
		Msg("received watch event")

	var data model.WatchEventData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		log.Error().Err(err).Str("event_id", event.EventID).Msg("failed to unmarshal watch data")
		msg.Nack(false, false)
		return
	}

	if err := c.indexer.IncrementViewCount(ctx, data.ContentID); err != nil {
		log.Error().Err(err).Str("content_id", data.ContentID).Msg("failed to increment view count in ES")
		msg.Nack(false, true)
		return
	}

	c.trendingSvc.IncrementViewCount(ctx, data.ContentID)

	today := time.Now().UTC().Format("2006-01-02")
	dailyKey := fmt.Sprintf("views:daily:%s:%s", today, data.ContentID)
	if err := c.redisClient.Incr(ctx, dailyKey).Err(); err != nil {
		log.Warn().Err(err).Str("key", dailyKey).Msg("failed to increment daily view count")
	} else {
		c.redisClient.Expire(ctx, dailyKey, 48*time.Hour)
	}

	log.Debug().Str("content_id", data.ContentID).Str("user_id", data.UserID).Msg("watch event processed")
	msg.Ack(false)
}

func (c *WatchConsumer) reconnect(ctx context.Context) bool {
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

		log.Info().Dur("backoff", backoff).Msg("attempting watch consumer channel recreation")

		ch, err := c.conn.Channel()
		if err != nil {
			log.Error().Err(err).Msg("failed to recreate watch consumer channel")
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		if err := ch.Qos(c.prefetch, 0, false); err != nil {
			log.Error().Err(err).Msg("failed to set watch consumer prefetch on new channel")
			ch.Close()
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		if c.channel != nil {
			c.channel.Close()
		}
		c.channel = ch
		log.Info().Msg("watch consumer channel recreated successfully")
		return true
	}
}
