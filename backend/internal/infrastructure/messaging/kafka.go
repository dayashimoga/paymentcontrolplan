package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paymentbridge/pcp/internal/domain/event"
	"github.com/segmentio/kafka-go"
)

// KafkaPublisher publishes domain events to Kafka topics.
type KafkaPublisher struct {
	writers map[string]*kafka.Writer
	brokers []string
}

// NewKafkaPublisher creates a Kafka publisher connected to the given brokers.
func NewKafkaPublisher(brokers []string) *KafkaPublisher {
	return &KafkaPublisher{
		writers: make(map[string]*kafka.Writer),
		brokers: brokers,
	}
}

func (p *KafkaPublisher) getWriter(topic string) *kafka.Writer {
	if w, ok := p.writers[topic]; ok {
		return w
	}
	w := &kafka.Writer{
		Addr:     kafka.TCP(p.brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	p.writers[topic] = w
	return w
}

// Publish sends a domain event to the specified Kafka topic.
func (p *KafkaPublisher) Publish(ctx context.Context, topic string, evt event.Event) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshaling event: %w", err)
	}
	w := p.getWriter(topic)
	return w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(evt.ID.String()),
		Value: data,
	})
}

// Close closes all Kafka writers.
func (p *KafkaPublisher) Close() error {
	for _, w := range p.writers {
		if err := w.Close(); err != nil {
			return err
		}
	}
	return nil
}

// KafkaSubscriber consumes domain events from Kafka topics.
type KafkaSubscriber struct {
	readers []*kafka.Reader
	brokers []string
	groupID string
}

// NewKafkaSubscriber creates a Kafka subscriber with the given consumer group.
func NewKafkaSubscriber(brokers []string, groupID string) *KafkaSubscriber {
	return &KafkaSubscriber{brokers: brokers, groupID: groupID}
}

// Subscribe starts consuming from a topic and calls handler for each event.
func (s *KafkaSubscriber) Subscribe(ctx context.Context, topic string, handler func(event.Event) error) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: s.brokers,
		Topic:   topic,
		GroupID: s.groupID,
	})
	s.readers = append(s.readers, reader)

	go func() {
		for {
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				continue
			}
			var evt event.Event
			if err := json.Unmarshal(msg.Value, &evt); err != nil {
				continue
			}
			_ = handler(evt)
		}
	}()

	return nil
}

// Close closes all Kafka readers.
func (s *KafkaSubscriber) Close() error {
	for _, r := range s.readers {
		if err := r.Close(); err != nil {
			return err
		}
	}
	return nil
}
