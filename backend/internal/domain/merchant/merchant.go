// Package merchant defines the Merchant bounded context including the
// aggregate root entity, value objects, repository port, and domain rules.
package merchant

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Status represents the lifecycle state of a merchant.
type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusInactive  Status = "inactive"
)

// IsValid returns true if the status is a recognized value.
func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusSuspended, StatusInactive:
		return true
	}
	return false
}

// String implements the Stringer interface.
func (s Status) String() string {
	return string(s)
}

// Merchant is the aggregate root for the merchant bounded context.
// It represents a business entity that uses PCP to orchestrate payments.
type Merchant struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	APIKey     string    `json:"api_key"`
	WebhookURL string    `json:"webhook_url"`
	Status     Status    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Validate enforces domain invariants on the merchant entity.
func (m *Merchant) Validate() error {
	if m.Name == "" {
		return ErrInvalidName
	}
	if len(m.Name) > 255 {
		return ErrInvalidName
	}
	if !m.Status.IsValid() {
		return ErrInvalidStatus
	}
	return nil
}

// Activate transitions the merchant to active status.
func (m *Merchant) Activate() {
	m.Status = StatusActive
	m.UpdatedAt = time.Now().UTC()
}

// Suspend transitions the merchant to suspended status.
func (m *Merchant) Suspend() {
	m.Status = StatusSuspended
	m.UpdatedAt = time.Now().UTC()
}

// Deactivate transitions the merchant to inactive status.
func (m *Merchant) Deactivate() {
	m.Status = StatusInactive
	m.UpdatedAt = time.Now().UTC()
}

// Repository defines the port for merchant persistence.
// This is the primary port in our hexagonal architecture —
// the domain declares what it needs, and infrastructure adapters implement it.
type Repository interface {
	// Create persists a new merchant. Returns ErrDuplicateMerchant if the name or API key conflicts.
	Create(ctx context.Context, m *Merchant) error

	// GetByID retrieves a merchant by its unique identifier.
	// Returns ErrMerchantNotFound if no merchant exists with the given ID.
	GetByID(ctx context.Context, id uuid.UUID) (*Merchant, error)

	// GetByAPIKey retrieves a merchant by its API key.
	// Returns ErrMerchantNotFound if no merchant exists with the given key.
	GetByAPIKey(ctx context.Context, apiKey string) (*Merchant, error)

	// List retrieves a paginated list of merchants.
	// Returns the merchants slice, total count, and any error.
	List(ctx context.Context, offset, limit int) ([]*Merchant, int, error)

	// Update persists changes to an existing merchant.
	// Returns ErrMerchantNotFound if the merchant does not exist.
	Update(ctx context.Context, m *Merchant) error

	// Delete removes a merchant by ID.
	// Returns ErrMerchantNotFound if the merchant does not exist.
	Delete(ctx context.Context, id uuid.UUID) error
}
