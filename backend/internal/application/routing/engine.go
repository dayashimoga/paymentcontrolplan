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

// SelectCandidateProviders returns an ordered list of candidate active providers for a payment,
// enabling automated cross-provider failover if the primary provider fails.
func (e *DefaultEngine) SelectCandidateProviders(ctx context.Context, merchantID uuid.UUID, amount int64, currency string) ([]*provider.Provider, error) {
	seen := make(map[uuid.UUID]bool)
	var candidates []*provider.Provider

	// 1. Check primary provider via SelectProvider
	if primary, err := e.SelectProvider(ctx, merchantID, amount, currency); err == nil && primary != nil {
		seen[primary.ID] = true
		candidates = append(candidates, primary)
	}

	// 2. Append all other active providers ordered by priority
	allActive, err := e.providerRepo.ListActive(ctx)
	if err == nil && len(allActive) > 0 {
		sort.Slice(allActive, func(i, j int) bool {
			return allActive[i].Priority < allActive[j].Priority
		})
		for _, p := range allActive {
			if !seen[p.ID] {
				seen[p.ID] = true
				candidates = append(candidates, p)
			}
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no active providers available for failover")
	}

	return candidates, nil
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
