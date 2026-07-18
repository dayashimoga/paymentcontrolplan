// Package payment defines the Payment bounded context.
package payment

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Status represents the payment lifecycle state.
type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusRefunded   Status = "refunded"
	StatusCancelled  Status = "cancelled"
)

// Payment is the aggregate root for payment transactions.
type Payment struct {
	ID             uuid.UUID         `json:"id"`
	MerchantID     uuid.UUID         `json:"merchant_id"`
	ProviderID     uuid.UUID         `json:"provider_id"`
	Amount         int64             `json:"amount"`
	Currency       string            `json:"currency"`
	Status         Status            `json:"status"`
	ExternalID     string            `json:"external_id"`
	IdempotencyKey string            `json:"idempotency_key"`
	Description    string            `json:"description"`
	Metadata       map[string]string `json:"metadata"`
	ErrorMessage   string            `json:"error_message,omitempty"`
	AttemptCount   int               `json:"attempt_count"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// Validate enforces domain invariants.
func (p *Payment) Validate() error {
	if p.Amount <= 0 {
		return ErrInvalidAmount
	}
	if p.Currency == "" || len(p.Currency) != 3 {
		return ErrInvalidCurrency
	}
	if p.MerchantID == uuid.Nil {
		return ErrInvalidMerchantID
	}
	return nil
}

// MarkProcessing transitions payment to processing.
func (p *Payment) MarkProcessing(providerID uuid.UUID) {
	p.Status = StatusProcessing
	p.ProviderID = providerID
	p.AttemptCount++
	p.UpdatedAt = time.Now().UTC()
}

// MarkCompleted transitions payment to completed.
func (p *Payment) MarkCompleted(externalID string) {
	p.Status = StatusCompleted
	p.ExternalID = externalID
	p.UpdatedAt = time.Now().UTC()
}

// MarkFailed transitions payment to failed.
func (p *Payment) MarkFailed(errMsg string) {
	p.Status = StatusFailed
	p.ErrorMessage = errMsg
	p.UpdatedAt = time.Now().UTC()
}

// MarkRefunded transitions payment to refunded.
func (p *Payment) MarkRefunded() {
	p.Status = StatusRefunded
	p.UpdatedAt = time.Now().UTC()
}

// Repository defines the port for payment persistence.
type Repository interface {
	Create(ctx context.Context, p *Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*Payment, error)
	GetByIdempotencyKey(ctx context.Context, merchantID uuid.UUID, key string) (*Payment, error)
	List(ctx context.Context, merchantID uuid.UUID, offset, limit int) ([]*Payment, int, error)
	Update(ctx context.Context, p *Payment) error
}
