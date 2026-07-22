package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/reconciliation"
)

// ReconciliationResponse represents a reconciliation record in API responses.
type ReconciliationResponse struct {
	ID             uuid.UUID `json:"id"`
	PaymentID      uuid.UUID `json:"payment_id"`
	ProviderID     uuid.UUID `json:"provider_id"`
	InternalAmount int64     `json:"internal_amount"`
	ExternalAmount int64     `json:"external_amount"`
	InternalStatus string    `json:"internal_status"`
	ExternalStatus string    `json:"external_status"`
	IsMatched      bool      `json:"is_matched"`
	Discrepancy    string    `json:"discrepancy,omitempty"`
	ReconciledAt   time.Time `json:"reconciled_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// ListReconciliationResponse represents a paginated list of reconciliation records.
type ListReconciliationResponse struct {
	Data   []ReconciliationResponse `json:"data"`
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Limit  int                      `json:"limit"`
}

// ToReconciliationResponse converts a domain Record to a DTO response.
func ToReconciliationResponse(r *reconciliation.Record) ReconciliationResponse {
	return ReconciliationResponse{
		ID:             r.ID,
		PaymentID:      r.PaymentID,
		ProviderID:     r.ProviderID,
		InternalAmount: r.InternalAmount,
		ExternalAmount: r.ExternalAmount,
		InternalStatus: string(r.InternalStatus),
		ExternalStatus: r.ExternalStatus,
		IsMatched:      r.IsMatched,
		Discrepancy:    r.Discrepancy,
		ReconciledAt:   r.ReconciledAt,
		CreatedAt:      r.CreatedAt,
	}
}

// ToReconciliationListResponse converts a slice of domain records to a list response.
func ToReconciliationListResponse(records []*reconciliation.Record, total, offset, limit int) ListReconciliationResponse {
	data := make([]ReconciliationResponse, 0, len(records))
	for _, r := range records {
		data = append(data, ToReconciliationResponse(r))
	}
	return ListReconciliationResponse{
		Data:   data,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}
}
