// Package event defines domain event types for the event-driven architecture.
package event

import (
	"time"

	"github.com/google/uuid"
)

// Type represents the kind of domain event.
type Type string

const (
	PaymentCreated   Type = "payment.created"
	PaymentCompleted Type = "payment.completed"
	PaymentFailed    Type = "payment.failed"
	PaymentRefunded  Type = "payment.refunded"
	MerchantCreated  Type = "merchant.created"
	MerchantUpdated  Type = "merchant.updated"
	ProviderHealthChanged Type = "provider.health_changed"
	WebhookDelivered Type = "webhook.delivered"
	WebhookFailed    Type = "webhook.failed"
)

// Event is the base domain event structure.
type Event struct {
	ID        uuid.UUID              `json:"id"`
	Type      Type                   `json:"type"`
	Source    string                  `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// New creates a new domain event.
func New(eventType Type, source string, data map[string]interface{}) Event {
	return Event{
		ID:        uuid.New(),
		Type:      eventType,
		Source:    source,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
}
