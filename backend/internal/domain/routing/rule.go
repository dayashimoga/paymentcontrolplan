// Package routing defines the routing engine for provider selection.
package routing

import (
	"context"

	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/provider"
)

// Rule defines a routing rule that maps merchants to providers with conditions.
type Rule struct {
	ID         uuid.UUID `json:"id"`
	MerchantID uuid.UUID `json:"merchant_id"`
	ProviderID uuid.UUID `json:"provider_id"`
	Priority   int       `json:"priority"`
	Weight     int       `json:"weight"`
	Currency   string    `json:"currency,omitempty"`
	MinAmount  int64     `json:"min_amount,omitempty"`
	MaxAmount  int64     `json:"max_amount,omitempty"`
	IsActive   bool      `json:"is_active"`
}

// Matches returns true if this rule applies to the given payment parameters.
func (r *Rule) Matches(currency string, amount int64) bool {
	if !r.IsActive {
		return false
	}
	if r.Currency != "" && r.Currency != currency {
		return false
	}
	if r.MinAmount > 0 && amount < r.MinAmount {
		return false
	}
	if r.MaxAmount > 0 && amount > r.MaxAmount {
		return false
	}
	return true
}

// Engine is the port for the routing engine that selects payment providers.
type Engine interface {
	SelectProvider(ctx context.Context, merchantID uuid.UUID, amount int64, currency string) (*provider.Provider, error)
	SelectCandidateProviders(ctx context.Context, merchantID uuid.UUID, amount int64, currency string) ([]*provider.Provider, error)
}

// RuleRepository defines the port for routing rule persistence.
type RuleRepository interface {
	GetByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*Rule, error)
	Create(ctx context.Context, rule *Rule) error
	Update(ctx context.Context, rule *Rule) error
	Delete(ctx context.Context, id uuid.UUID) error
}
