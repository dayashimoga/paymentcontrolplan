// Package provider defines the Provider bounded context for payment gateway abstraction.
package provider

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Type represents the payment provider type.
type Type string

const (
	TypeStripe   Type = "stripe"
	TypePayPal   Type = "paypal"
	TypeAdyen    Type = "adyen"
	TypeRazorpay Type = "razorpay"
)

// Status represents the provider operational status.
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusDegraded Status = "degraded"
)

// Provider is the aggregate root representing a payment provider configuration.
type Provider struct {
	ID        uuid.UUID         `json:"id"`
	Name      string            `json:"name"`
	Type      Type              `json:"type"`
	Config    map[string]string `json:"config"`
	Status    Status            `json:"status"`
	Priority  int               `json:"priority"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// Validate enforces domain invariants.
func (p *Provider) Validate() error {
	if p.Name == "" {
		return ErrInvalidProviderName
	}
	if !p.Type.IsValid() {
		return ErrInvalidProviderType
	}
	return nil
}

// IsValid checks if the provider type is recognized.
func (t Type) IsValid() bool {
	switch t {
	case TypeStripe, TypePayPal, TypeAdyen, TypeRazorpay:
		return true
	}
	return false
}

// ChargeRequest represents a request to charge a payment method.
type ChargeRequest struct {
	Amount         int64
	Currency       string
	Description    string
	CustomerEmail  string
	Metadata       map[string]string
	IdempotencyKey string
}

// ChargeResponse represents the result of a charge operation.
type ChargeResponse struct {
	ExternalID   string
	Status       string
	ProviderType Type
	RawResponse  map[string]interface{}
}

// RefundRequest represents a request to refund a transaction.
type RefundRequest struct {
	ExternalID string
	Amount     int64
	Reason     string
}

// RefundResponse represents the result of a refund operation.
type RefundResponse struct {
	ExternalID  string
	Status      string
	RefundID    string
	RawResponse map[string]interface{}
}

// TransactionStatus represents the status of a transaction at the provider.
type TransactionStatus struct {
	ExternalID string
	Status     string
	Amount     int64
	Currency   string
	RawData    map[string]interface{}
}

// Gateway is the port interface that all payment provider connectors must implement.
type Gateway interface {
	Charge(ctx context.Context, req ChargeRequest) (*ChargeResponse, error)
	Refund(ctx context.Context, req RefundRequest) (*RefundResponse, error)
	GetTransactionStatus(ctx context.Context, externalID string) (*TransactionStatus, error)
	HealthCheck(ctx context.Context) error
	ProviderType() Type
}

// Repository defines the port for provider persistence.
type Repository interface {
	Create(ctx context.Context, p *Provider) error
	GetByID(ctx context.Context, id uuid.UUID) (*Provider, error)
	List(ctx context.Context, offset, limit int) ([]*Provider, int, error)
	ListActive(ctx context.Context) ([]*Provider, error)
	Update(ctx context.Context, p *Provider) error
	Delete(ctx context.Context, id uuid.UUID) error
}
