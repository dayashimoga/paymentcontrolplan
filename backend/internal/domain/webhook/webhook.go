// Package webhook defines the Webhook bounded context.
package webhook

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Status represents webhook delivery status.
type Status string

const (
	StatusPending   Status = "pending"
	StatusDelivered Status = "delivered"
	StatusFailed    Status = "failed"
)

// Webhook represents an outbound webhook notification.
type Webhook struct {
	ID         uuid.UUID `json:"id"`
	MerchantID uuid.UUID `json:"merchant_id"`
	PaymentID  uuid.UUID `json:"payment_id"`
	URL        string    `json:"url"`
	EventType  string    `json:"event_type"`
	Payload    string    `json:"payload"`
	Status     Status    `json:"status"`
	Attempts   int       `json:"attempts"`
	MaxRetries int       `json:"max_retries"`
	NextRetry  time.Time `json:"next_retry"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Repository defines the port for webhook persistence.
type Repository interface {
	Create(ctx context.Context, w *Webhook) error
	GetByID(ctx context.Context, id uuid.UUID) (*Webhook, error)
	ListPending(ctx context.Context, limit int) ([]*Webhook, error)
	ListByPayment(ctx context.Context, paymentID uuid.UUID) ([]*Webhook, error)
	Update(ctx context.Context, w *Webhook) error
}
