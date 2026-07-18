package event_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/event"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	evtType := event.PaymentCreated
	source := "payment_service"
	data := map[string]interface{}{"amount": 1000, "currency": "USD"}

	evt := event.New(evtType, source, data)

	assert.NotEqual(t, uuid.Nil, evt.ID)
	assert.Equal(t, evtType, evt.Type)
	assert.Equal(t, source, evt.Source)
	assert.Equal(t, data, evt.Data)
	assert.WithinDuration(t, time.Now(), evt.Timestamp, time.Second)
}
