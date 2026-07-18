// Package analytics provides payment analytics and reporting.
package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Summary holds aggregated payment statistics.
type Summary struct {
	TotalPayments     int     `json:"total_payments"`
	CompletedPayments int     `json:"completed_payments"`
	FailedPayments    int     `json:"failed_payments"`
	TotalAmount       int64   `json:"total_amount"`
	SuccessRate       float64 `json:"success_rate"`
	Currency          string  `json:"currency"`
	Period            string  `json:"period"`
}

// ProviderStats holds per-provider statistics.
type ProviderStats struct {
	ProviderID   uuid.UUID `json:"provider_id"`
	ProviderName string    `json:"provider_name"`
	TotalCharges int       `json:"total_charges"`
	SuccessCount int       `json:"success_count"`
	FailureCount int       `json:"failure_count"`
	TotalAmount  int64     `json:"total_amount"`
	SuccessRate  float64   `json:"success_rate"`
	AvgLatencyMs int64     `json:"avg_latency_ms"`
}

// Repository defines the analytics query port.
type Repository interface {
	GetSummary(ctx context.Context, merchantID uuid.UUID, from, to time.Time) (*Summary, error)
	GetProviderStats(ctx context.Context, merchantID uuid.UUID, from, to time.Time) ([]*ProviderStats, error)
}

// Service provides analytics query capabilities.
type Service struct {
	repo Repository
}

// NewService creates a new analytics service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetSummary retrieves aggregated payment statistics for a merchant.
func (s *Service) GetSummary(ctx context.Context, merchantID uuid.UUID, from, to time.Time) (*Summary, error) {
	return s.repo.GetSummary(ctx, merchantID, from, to)
}

// GetProviderStats retrieves per-provider statistics.
func (s *Service) GetProviderStats(ctx context.Context, merchantID uuid.UUID, from, to time.Time) ([]*ProviderStats, error) {
	return s.repo.GetProviderStats(ctx, merchantID, from, to)
}
