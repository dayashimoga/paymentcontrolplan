// Package messaging defines the messaging abstraction allowing Kafka/RabbitMQ/NATS swap.
package messaging

import (
	"context"

	"github.com/paymentbridge/pcp/internal/domain/event"
)

// Publisher publishes domain events to a message broker.
type Publisher interface {
	Publish(ctx context.Context, topic string, evt event.Event) error
	Close() error
}

// Subscriber consumes domain events from a message broker.
type Subscriber interface {
	Subscribe(ctx context.Context, topic string, handler func(event.Event) error) error
	Close() error
}

// InMemoryPublisher is a no-op publisher for local development and testing.
type InMemoryPublisher struct {
	events []event.Event
}

// NewInMemoryPublisher creates an in-memory publisher (no external broker needed).
func NewInMemoryPublisher() *InMemoryPublisher {
	return &InMemoryPublisher{events: make([]event.Event, 0)}
}

func (p *InMemoryPublisher) Publish(_ context.Context, _ string, evt event.Event) error {
	p.events = append(p.events, evt)
	return nil
}

func (p *InMemoryPublisher) Close() error { return nil }

// Events returns all published events (for testing).
func (p *InMemoryPublisher) Events() []event.Event { return p.events }
