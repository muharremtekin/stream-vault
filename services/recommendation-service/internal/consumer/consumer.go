package consumer

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

// baseConsumer provides shared reconnection and lifecycle logic for all consumers.
type baseConsumer struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	queue    string
	prefetch int
	done     chan struct{}
}

func newBaseConsumer(conn *amqp.Connection, queue string, prefetch int) (*baseConsumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("opening channel for %s consumer: %w", queue, err)
	}
	if err := ch.Qos(prefetch, 0, false); err != nil {
		ch.Close()
		return nil, fmt.Errorf("setting prefetch for %s consumer: %w", queue, err)
	}
	return &baseConsumer{
		conn:     conn,
		channel:  ch,
		queue:    queue,
		prefetch: prefetch,
		done:     make(chan struct{}),
	}, nil
}

// consume registers as a consumer and returns the delivery channel.
func (b *baseConsumer) consume(tag string) (<-chan amqp.Delivery, error) {
	return b.channel.Consume(
		b.queue,
		tag,
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
}

// reconnect recreates the AMQP channel with exponential backoff.
func (b *baseConsumer) reconnect(ctx context.Context) bool {
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-b.done:
			return false
		case <-ctx.Done():
			return false
		case <-time.After(backoff):
		}

		log.Info().Str("queue", b.queue).Dur("backoff", backoff).Msg("attempting consumer channel recreation")

		ch, err := b.conn.Channel()
		if err != nil {
			log.Error().Err(err).Str("queue", b.queue).Msg("failed to recreate consumer channel")
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		if err := ch.Qos(b.prefetch, 0, false); err != nil {
			log.Error().Err(err).Str("queue", b.queue).Msg("failed to set prefetch on new channel")
			ch.Close()
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		if b.channel != nil {
			b.channel.Close()
		}
		b.channel = ch
		log.Info().Str("queue", b.queue).Msg("consumer channel recreated")
		return true
	}
}

// stop signals the consumer to stop and closes the channel.
func (b *baseConsumer) stop() {
	close(b.done)
	if b.channel != nil {
		b.channel.Close()
	}
}
