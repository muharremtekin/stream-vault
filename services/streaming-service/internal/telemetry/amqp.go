package telemetry

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

// AMQPCarrier implements propagation.TextMapCarrier over amqp.Table headers.
type AMQPCarrier amqp.Table

// Get returns the value for the given key from AMQP headers.
func (c AMQPCarrier) Get(key string) string {
	if val, ok := amqp.Table(c)[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// Set stores a key-value pair in AMQP headers.
func (c AMQPCarrier) Set(key, val string) {
	amqp.Table(c)[key] = val
}

// Keys returns all keys in the AMQP headers.
func (c AMQPCarrier) Keys() []string {
	keys := make([]string, 0, len(amqp.Table(c)))
	for k := range amqp.Table(c) {
		keys = append(keys, k)
	}
	return keys
}

// InjectAMQP injects trace context from ctx into AMQP headers.
// If headers is nil, a new amqp.Table is created.
func InjectAMQP(ctx context.Context, headers amqp.Table) amqp.Table {
	if headers == nil {
		headers = amqp.Table{}
	}
	otel.GetTextMapPropagator().Inject(ctx, AMQPCarrier(headers))
	return headers
}

// ExtractAMQP extracts trace context from AMQP headers into a new context.
func ExtractAMQP(ctx context.Context, headers amqp.Table) context.Context {
	if headers == nil {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, AMQPCarrier(headers))
}
