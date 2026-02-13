package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"

	"github.com/streamvault/streaming-service/internal/config"
)

type EncodingJob struct {
	JobID        string    `json:"job_id"`
	ContentID    string    `json:"content_id"`
	SourcePath   string    `json:"source_path"`
	SourceBucket string    `json:"source_bucket"`
	RequestedBy  string    `json:"requested_by"`
	CreatedAt    time.Time `json:"created_at"`
}

type Publisher interface {
	PublishEncodingJob(ctx context.Context, job EncodingJob) error
	Close() error
}

type amqpPublisher struct {
	channel    *amqp.Channel
	exchange   string
	routingKey string
}

func NewPublisher(conn *amqp.Connection, cfg config.RabbitMQConfig) (Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("opening rabbitmq channel: %w", err)
	}

	log.Info().Str("exchange", cfg.Exchange).Str("routing_key", cfg.PublishRoutingKey).Msg("rabbitmq publisher ready")

	return &amqpPublisher{
		channel:    ch,
		exchange:   cfg.Exchange,
		routingKey: cfg.PublishRoutingKey,
	}, nil
}

func (p *amqpPublisher) PublishEncodingJob(ctx context.Context, job EncodingJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshalling encoding job: %w", err)
	}

	err = p.channel.PublishWithContext(ctx,
		p.exchange,
		p.routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    job.JobID,
			Timestamp:    job.CreatedAt,
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publishing encoding job: %w", err)
	}

	log.Info().Str("job_id", job.JobID).Str("content_id", job.ContentID).Msg("encoding job published")
	return nil
}

func (p *amqpPublisher) Close() error {
	return p.channel.Close()
}
