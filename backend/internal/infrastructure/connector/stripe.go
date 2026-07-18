// Package connector provides payment gateway implementations.
package connector

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/provider"
)

// StripeGateway implements the Gateway port for Stripe.
// In production, this would use the Stripe SDK. For PCP, it provides
// the abstraction layer — actual Stripe API calls use provider config credentials.
type StripeGateway struct {
	apiKey string
}

// NewStripeGateway creates a Stripe gateway from provider config.
func NewStripeGateway(config map[string]string) *StripeGateway {
	return &StripeGateway{apiKey: config["api_key"]}
}

func (g *StripeGateway) ProviderType() provider.Type { return provider.TypeStripe }

func (g *StripeGateway) Charge(ctx context.Context, req provider.ChargeRequest) (*provider.ChargeResponse, error) {
	// In production: call Stripe API with g.apiKey
	// stripe.Key = g.apiKey
	// params := &stripe.ChargeParams{Amount: &req.Amount, Currency: &req.Currency, ...}
	// ch, err := charge.New(params)
	externalID := fmt.Sprintf("ch_%s", uuid.New().String()[:24])
	return &provider.ChargeResponse{
		ExternalID:   externalID,
		Status:       "succeeded",
		ProviderType: provider.TypeStripe,
		RawResponse:  map[string]interface{}{"id": externalID, "object": "charge", "amount": req.Amount, "currency": req.Currency},
	}, nil
}

func (g *StripeGateway) Refund(ctx context.Context, req provider.RefundRequest) (*provider.RefundResponse, error) {
	refundID := fmt.Sprintf("re_%s", uuid.New().String()[:24])
	return &provider.RefundResponse{
		ExternalID:  req.ExternalID,
		Status:      "succeeded",
		RefundID:    refundID,
		RawResponse: map[string]interface{}{"id": refundID, "object": "refund", "charge": req.ExternalID},
	}, nil
}

func (g *StripeGateway) GetTransactionStatus(ctx context.Context, externalID string) (*provider.TransactionStatus, error) {
	return &provider.TransactionStatus{
		ExternalID: externalID,
		Status:     "succeeded",
		RawData:    map[string]interface{}{"id": externalID, "status": "succeeded"},
	}, nil
}

func (g *StripeGateway) HealthCheck(ctx context.Context) error {
	if g.apiKey == "" {
		return provider.ErrProviderUnavailable
	}
	return nil
}
