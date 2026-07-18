// Package reconciliation defines the Reconciliation bounded context.
package reconciliation

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/payment"
)

// Status is an alias for payment.Status used in reconciliation context.
type Status = payment.Status

// Record represents a reconciliation check between internal and provider state.
type Record struct {
	ID             uuid.UUID `json:"id"`
	PaymentID      uuid.UUID `json:"payment_id"`
	ProviderID     uuid.UUID `json:"provider_id"`
	InternalAmount int64     `json:"internal_amount"`
	ExternalAmount int64     `json:"external_amount"`
	InternalStatus Status    `json:"internal_status"`
	ExternalStatus string    `json:"external_status"`
	IsMatched      bool      `json:"is_matched"`
	Discrepancy    string    `json:"discrepancy,omitempty"`
	ReconciledAt   time.Time `json:"reconciled_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// Repository defines the port for reconciliation persistence.
type Repository interface {
	Create(ctx context.Context, r *Record) error
	GetByPayment(ctx context.Context, paymentID uuid.UUID) (*Record, error)
	ListUnmatched(ctx context.Context, offset, limit int) ([]*Record, int, error)
}

// Service provides reconciliation capabilities.
type Service struct {
	repo Repository
}

// NewService creates a new reconciliation service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Reconcile checks a payment against its provider status and records the result.
func (s *Service) Reconcile(ctx context.Context, paymentID, providerID uuid.UUID, internalAmt int64, internalStatus payment.Status, externalAmt int64, externalStatus string) (*Record, error) {
	isMatched := internalAmt == externalAmt && string(internalStatus) == externalStatus
	discrepancy := ""
	if !isMatched {
		if internalAmt != externalAmt {
			discrepancy = "amount_mismatch"
		} else {
			discrepancy = "status_mismatch"
		}
	}
	now := time.Now().UTC()
	rec := &Record{
		ID: uuid.New(), PaymentID: paymentID, ProviderID: providerID,
		InternalAmount: internalAmt, ExternalAmount: externalAmt,
		InternalStatus: internalStatus, ExternalStatus: externalStatus,
		IsMatched: isMatched, Discrepancy: discrepancy,
		ReconciledAt: now, CreatedAt: now,
	}
	if err := s.repo.Create(ctx, rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// ListUnmatched retrieves reconciliation records with discrepancies.
func (s *Service) ListUnmatched(ctx context.Context, offset, limit int) ([]*Record, int, error) {
	return s.repo.ListUnmatched(ctx, offset, limit)
}
