package messaging

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

	"github.com/streamvault/streaming-service/internal/config"
	"github.com/streamvault/streaming-service/internal/telemetry"
)

type EncodingResult struct {
	// Event envelope fields (Rule 3.4)
	EventID       string `json:"event_id,omitempty"`
	EventType     string `json:"event_type,omitempty"`
	Timestamp     string `json:"timestamp,omitempty"`
	Source        string `json:"source,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`

	// Data fields
	JobID       string           `json:"job_id"`
	ContentID   string           `json:"content_id"`
	Status      string           `json:"status"`
	Outputs     []EncodingOutput `json:"outputs"`
	Duration    int64            `json:"duration_seconds"`
	Error       string           `json:"error_message,omitempty"`
	CompletedAt string           `json:"completed_at"`
}

type EncodingOutput struct {
	Quality      string `json:"quality"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	BitrateKbps  int    `json:"bitrate_kbps"`
	SegmentCount int    `json:"segment_count"`
	PlaylistPath string `json:"playlist_path"`
}

type Consumer struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	redisClient *redis.Client
	queueName   string
	prefetch    int
	done        chan struct{}
}

func NewConsumer(conn *amqp.Connection, cfg config.RabbitMQConfig, redisClient *redis.Client) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("opening rabbitmq channel for consumer: %w", err)
	}

	if err := ch.Qos(cfg.Prefetch, 0, false); err != nil {
		return nil, fmt.Errorf("setting consumer prefetch: %w", err)
	}

	return &Consumer{
		conn:        conn,
		channel:     ch,
		redisClient: redisClient,
		queueName:   cfg.ResultQueue,
		prefetch:    cfg.Prefetch,
		done:        make(chan struct{}),
	}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	for {
		if err := c.consumeLoop(ctx); err != nil {
			log.Error().Err(err).Msg("consumer loop exited with error")
		}

		select {
		case <-c.done:
			log.Info().Msg("consumer stopping (done signal)")
			return
		case <-ctx.Done():
			log.Info().Msg("consumer stopping (context cancelled)")
			return
		default:
		}

		if !c.reconnect(ctx) {
			return
		}
	}
}

func (c *Consumer) consumeLoop(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.queueName,
		"streaming-consumer",
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Info().Str("queue", c.queueName).Msg("consumer started")

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("delivery channel closed")
			}
			c.handleResult(ctx, msg)
		case <-c.done:
			return nil
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *Consumer) reconnect(ctx context.Context) bool {
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

		log.Info().Dur("backoff", backoff).Msg("attempting consumer channel recreation")

		ch, err := c.conn.Channel()
		if err != nil {
			log.Error().Err(err).Msg("failed to recreate consumer channel")
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		if err := ch.Qos(c.prefetch, 0, false); err != nil {
			log.Error().Err(err).Msg("failed to set consumer prefetch on new channel")
			ch.Close()
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		if c.channel != nil {
			c.channel.Close()
		}
		c.channel = ch
		log.Info().Msg("consumer channel recreated successfully")
		return true
	}
}

func (c *Consumer) Stop() {
	close(c.done)
	if c.channel != nil {
		c.channel.Close()
	}
}

var consumeTracer = otel.Tracer("streaming-service/messaging")

func (c *Consumer) handleResult(ctx context.Context, msg amqp.Delivery) {
	// Extract trace context from AMQP headers.
	ctx = telemetry.ExtractAMQP(ctx, msg.Headers)
	ctx, span := consumeTracer.Start(ctx, "rabbitmq.consume",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.source.name", c.queueName),
			attribute.String("messaging.operation", "receive"),
		),
	)
	defer span.End()

	var result EncodingResult
	if err := json.Unmarshal(msg.Body, &result); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal encoding result")
		span.RecordError(err)
		msg.Nack(false, false)
		return
	}

	span.SetAttributes(
		attribute.String("messaging.message.id", result.EventID),
		attribute.String("job.id", result.JobID),
		attribute.String("content.id", result.ContentID),
	)

	log.Info().
		Str("event_id", result.EventID).
		Str("event_type", result.EventType).
		Str("source", result.Source).
		Str("correlation_id", result.CorrelationID).
		Str("job_id", result.JobID).
		Str("content_id", result.ContentID).
		Str("status", result.Status).
		Msg("received encoding result")

	if result.Status == "completed" {
		key := fmt.Sprintf("stream-info:%s", result.ContentID)
		data, _ := json.Marshal(result)
		if err := c.redisClient.Set(ctx, key, data, 0).Err(); err != nil {
			log.Error().Err(err).Str("content_id", result.ContentID).Msg("failed to cache stream info")
			span.RecordError(err)
			msg.Nack(false, true)
			return
		}
		log.Info().Str("content_id", result.ContentID).Msg("stream info cached")
	}

	msg.Ack(false)
}
