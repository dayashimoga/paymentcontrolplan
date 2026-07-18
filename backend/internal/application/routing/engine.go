// Package routing implements the payment routing engine.
package routing

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/provider"
	"github.com/paymentbridge/pcp/internal/domain/routing"
)

// DefaultEngine implements the routing.Engine port using rule-based weighted selection.
type DefaultEngine struct {
	ruleRepo     routing.RuleRepository
	providerRepo provider.Repository
}

// NewEngine creates a new routing engine.
func NewEngine(ruleRepo routing.RuleRepository, providerRepo provider.Repository) *DefaultEngine {
	return &DefaultEngine{ruleRepo: ruleRepo, providerRepo: providerRepo}
}

// SelectProvider selects the best provider for a payment based on routing rules.
// Priority: rule priority → weight-based random selection → fallback to highest priority active provider.
func (e *DefaultEngine) SelectProvider(ctx context.Context, merchantID uuid.UUID, amount int64, currency string) (*provider.Provider, error) {
	rules, err := e.ruleRepo.GetByMerchant(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("fetching routing rules: %w", err)
	}

	// Filter matching rules
	var matching []*routing.Rule
	for _, r := range rules {
		if r.Matches(currency, amount) {
			matching = append(matching, r)
		}
	}

	// If rules exist, use weighted selection
	if len(matching) > 0 {
		sort.Slice(matching, func(i, j int) bool {
			return matching[i].Priority < matching[j].Priority
		})

		selected := weightedSelect(matching)
		p, err := e.providerRepo.GetByID(ctx, selected.ProviderID)
		if err == nil && p.Status == provider.StatusActive {
			return p, nil
		}
		// Fall through to try other matching rules
		for _, r := range matching {
			p, err := e.providerRepo.GetByID(ctx, r.ProviderID)
			if err == nil && p.Status == provider.StatusActive {
				return p, nil
			}
		}
	}

	// Fallback: pick highest priority active provider
	providers, err := e.providerRepo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing active providers: %w", err)
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("no active providers available")
	}

	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Priority < providers[j].Priority
	})

	return providers[0], nil
}

// weightedSelect picks a rule using weight-based random selection.
func weightedSelect(rules []*routing.Rule) *routing.Rule {
	totalWeight := 0
	for _, r := range rules {
		w := r.Weight
		if w <= 0 {
			w = 1
		}
		totalWeight += w
	}

	pick := rand.Intn(totalWeight)
	cumulative := 0
	for _, r := range rules {
		w := r.Weight
		if w <= 0 {
			w = 1
		}
		cumulative += w
		if pick < cumulative {
			return r
		}
	}
	return rules[0]
}
