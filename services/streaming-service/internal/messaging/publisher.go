package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/streamvault/streaming-service/internal/config"
	"github.com/streamvault/streaming-service/internal/telemetry"
)

type EncodingJob struct {
	JobID        string    `json:"job_id"`
	ContentID    string    `json:"content_id"`
	SourcePath   string    `json:"source_path"`
	SourceBucket string    `json:"source_bucket"`
	RequestedBy  string    `json:"requested_by"`
	CreatedAt    time.Time `json:"created_at"`
}

type WatchCompletedEvent struct {
	EventID       string             `json:"eventId"`
	EventType     string             `json:"eventType"`
	Timestamp     string             `json:"timestamp"`
	Source        string             `json:"source"`
	CorrelationID string             `json:"correlationId"`
	Data          WatchCompletedData `json:"data"`
}

type WatchCompletedData struct {
	UserID               string  `json:"userId"`
	ContentID            string  `json:"contentId"`
	CompletionPercentage float64 `json:"completionPercentage"`
	WatchedAt            string  `json:"watchedAt"`
}

type Publisher interface {
	PublishEncodingJob(ctx context.Context, job EncodingJob) error
	PublishWatchCompleted(ctx context.Context, event WatchCompletedEvent) error
	Close() error
}

type amqpPublisher struct {
	mu              sync.Mutex
	conn            *amqp.Connection
	channel         *amqp.Channel
	exchange        string
	routingKey      string
	watchExchange   string
	watchRoutingKey string
}

func NewPublisher(conn *amqp.Connection, cfg config.RabbitMQConfig) (Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("opening rabbitmq channel: %w", err)
	}

	log.Info().Str("exchange", cfg.Exchange).Str("routing_key", cfg.PublishRoutingKey).Msg("rabbitmq publisher ready")

	return &amqpPublisher{
		conn:            conn,
		channel:         ch,
		exchange:        cfg.Exchange,
		routingKey:      cfg.PublishRoutingKey,
		watchExchange:   cfg.WatchExchange,
		watchRoutingKey: cfg.WatchRoutingKey,
	}, nil
}

func (p *amqpPublisher) recreateChannel() error {
	if p.channel != nil {
		p.channel.Close()
	}
	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("recreating rabbitmq channel: %w", err)
	}
	p.channel = ch
	log.Info().Msg("rabbitmq publisher channel recreated")
	return nil
}

var publishTracer = otel.Tracer("streaming-service/messaging")

func (p *amqpPublisher) publish(ctx context.Context, exchange, routingKey string, messageID string, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	ctx, span := publishTracer.Start(ctx, "rabbitmq.publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.destination.name", exchange),
			attribute.String("messaging.rabbitmq.routing_key", routingKey),
			attribute.String("messaging.message.id", messageID),
		),
	)
	defer span.End()

	headers := telemetry.InjectAMQP(ctx, nil)

	publishing := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    messageID,
		Timestamp:    time.Now().UTC(),
		Headers:      headers,
		Body:         body,
	}

	err := p.channel.PublishWithContext(ctx, exchange, routingKey, false, false, publishing)
	if err != nil {
		log.Warn().Err(err).Msg("publish failed, attempting channel recreation")
		if recreateErr := p.recreateChannel(); recreateErr != nil {
			return fmt.Errorf("publish (channel recreation failed): %w", recreateErr)
		}
		err = p.channel.PublishWithContext(ctx, exchange, routingKey, false, false, publishing)
		if err != nil {
			return fmt.Errorf("publish (retry failed): %w", err)
		}
	}

	return nil
}

func (p *amqpPublisher) PublishEncodingJob(ctx context.Context, job EncodingJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshalling encoding job: %w", err)
	}

	if err := p.publish(ctx, p.exchange, p.routingKey, job.JobID, body); err != nil {
		return fmt.Errorf("publishing encoding job: %w", err)
	}

	log.Info().Str("job_id", job.JobID).Str("content_id", job.ContentID).Msg("encoding job published")
	return nil
}

func (p *amqpPublisher) PublishWatchCompleted(ctx context.Context, event WatchCompletedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling watch completed event: %w", err)
	}

	if err := p.publish(ctx, p.watchExchange, p.watchRoutingKey, event.EventID, body); err != nil {
		return fmt.Errorf("publishing watch completed event: %w", err)
	}

	log.Info().
		Str("userId", event.Data.UserID).
		Str("contentId", event.Data.ContentID).
		Float64("percentage", event.Data.CompletionPercentage).
		Msg("watch completed event published")
	return nil
}

func (p *amqpPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.channel != nil {
		return p.channel.Close()
	}
	return nil
}
