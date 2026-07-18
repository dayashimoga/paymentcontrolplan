package messaging_test

import (
	"testing"

	"github.com/paymentbridge/pcp/internal/infrastructure/messaging"
	"github.com/stretchr/testify/assert"
)

func TestNewKafkaPublisher(t *testing.T) {
	pub := messaging.NewKafkaPublisher([]string{"localhost:9092"})
	assert.NotNil(t, pub)

	err := pub.Close()
	assert.NoError(t, err)
}

func TestNewKafkaSubscriber(t *testing.T) {
	sub := messaging.NewKafkaSubscriber([]string{"localhost:9092"}, "group-1")
	assert.NotNil(t, sub)

	err := sub.Close()
	assert.NoError(t, err)
}
