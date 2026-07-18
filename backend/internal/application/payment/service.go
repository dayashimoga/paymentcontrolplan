// Package payment implements payment processing use cases.
package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	domain "github.com/paymentbridge/pcp/internal/domain/payment"
	"github.com/paymentbridge/pcp/internal/domain/provider"
	"github.com/paymentbridge/pcp/internal/domain/routing"
)

// Service orchestrates payment processing with routing and provider abstraction.
type Service struct {
	paymentRepo   domain.Repository
	providerRepo  provider.Repository
	routingEngine routing.Engine
	gateways      map[provider.Type]provider.Gateway
}

// NewService creates a new payment service.
func NewService(pr domain.Repository, provRepo provider.Repository, re routing.Engine) *Service {
	return &Service{
		paymentRepo:   pr,
		providerRepo:  provRepo,
		routingEngine: re,
		gateways:      make(map[provider.Type]provider.Gateway),
	}
}

// RegisterGateway registers a gateway implementation.
func (s *Service) RegisterGateway(t provider.Type, gw provider.Gateway) {
	s.gateways[t] = gw
}

// CreateInput holds the input for creating a payment.
type CreateInput struct {
	MerchantID     uuid.UUID
	Amount         int64
	Currency       string
	Description    string
	Metadata       map[string]string
	IdempotencyKey string
}

// Create initiates a new payment, routes it to a provider, and processes it.
func (s *Service) Create(ctx context.Context, input CreateInput) (*domain.Payment, error) {
	// Check idempotency
	if input.IdempotencyKey != "" {
		existing, err := s.paymentRepo.GetByIdempotencyKey(ctx, input.MerchantID, input.IdempotencyKey)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	now := time.Now().UTC()
	payment := &domain.Payment{
		ID:             uuid.New(),
		MerchantID:     input.MerchantID,
		Amount:         input.Amount,
		Currency:       input.Currency,
		Status:         domain.StatusPending,
		Description:    input.Description,
		Metadata:       input.Metadata,
		IdempotencyKey: input.IdempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := payment.Validate(); err != nil {
		return nil, err
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	// Route to a provider
	selectedProvider, err := s.routingEngine.SelectProvider(ctx, input.MerchantID, input.Amount, input.Currency)
	if err != nil {
		payment.MarkFailed("no available provider: " + err.Error())
		s.paymentRepo.Update(ctx, payment)
		return payment, nil
	}

	// Get gateway for the provider type
	gw, ok := s.gateways[selectedProvider.Type]
	if !ok {
		payment.MarkFailed(fmt.Sprintf("no gateway for provider type: %s", selectedProvider.Type))
		s.paymentRepo.Update(ctx, payment)
		return payment, nil
	}

	// Process charge
	payment.MarkProcessing(selectedProvider.ID)
	s.paymentRepo.Update(ctx, payment)

	chargeResp, err := gw.Charge(ctx, provider.ChargeRequest{
		Amount:         input.Amount,
		Currency:       input.Currency,
		Description:    input.Description,
		Metadata:       input.Metadata,
		IdempotencyKey: input.IdempotencyKey,
	})
	if err != nil {
		payment.MarkFailed(err.Error())
		s.paymentRepo.Update(ctx, payment)
		return payment, nil
	}

	payment.MarkCompleted(chargeResp.ExternalID)
	s.paymentRepo.Update(ctx, payment)
	return payment, nil
}

// GetByID retrieves a payment by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	return s.paymentRepo.GetByID(ctx, id)
}

// List retrieves payments for a merchant.
func (s *Service) List(ctx context.Context, merchantID uuid.UUID, offset, limit int) ([]*domain.Payment, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.paymentRepo.List(ctx, merchantID, offset, limit)
}

// Refund processes a refund for a completed payment.
func (s *Service) Refund(ctx context.Context, paymentID uuid.UUID) (*domain.Payment, error) {
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if payment.Status != domain.StatusCompleted {
		return nil, fmt.Errorf("can only refund completed payments")
	}

	prov, err := s.providerRepo.GetByID(ctx, payment.ProviderID)
	if err != nil {
		return nil, err
	}

	gw, ok := s.gateways[prov.Type]
	if !ok {
		return nil, provider.ErrProviderUnavailable
	}

	_, err = gw.Refund(ctx, provider.RefundRequest{
		ExternalID: payment.ExternalID,
		Amount:     payment.Amount,
		Reason:     "merchant_requested",
	})
	if err != nil {
		return nil, err
	}

	payment.MarkRefunded()
	s.paymentRepo.Update(ctx, payment)
	return payment, nil
}
