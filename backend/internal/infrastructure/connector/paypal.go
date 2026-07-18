package connector

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/provider"
)

// PayPalGateway implements the Gateway port for PayPal.
type PayPalGateway struct {
	clientID     string
	clientSecret string
}

// NewPayPalGateway creates a PayPal gateway from provider config.
func NewPayPalGateway(config map[string]string) *PayPalGateway {
	return &PayPalGateway{clientID: config["client_id"], clientSecret: config["client_secret"]}
}

func (g *PayPalGateway) ProviderType() provider.Type { return provider.TypePayPal }

func (g *PayPalGateway) Charge(ctx context.Context, req provider.ChargeRequest) (*provider.ChargeResponse, error) {
	// In production: call PayPal Orders API
	externalID := fmt.Sprintf("PAYID-%s", uuid.New().String()[:20])
	return &provider.ChargeResponse{
		ExternalID:   externalID,
		Status:       "COMPLETED",
		ProviderType: provider.TypePayPal,
		RawResponse:  map[string]interface{}{"id": externalID, "status": "COMPLETED", "intent": "CAPTURE"},
	}, nil
}

func (g *PayPalGateway) Refund(ctx context.Context, req provider.RefundRequest) (*provider.RefundResponse, error) {
	refundID := fmt.Sprintf("REF-%s", uuid.New().String()[:20])
	return &provider.RefundResponse{
		ExternalID: req.ExternalID, Status: "COMPLETED", RefundID: refundID,
		RawResponse: map[string]interface{}{"id": refundID, "status": "COMPLETED"},
	}, nil
}

func (g *PayPalGateway) GetTransactionStatus(ctx context.Context, externalID string) (*provider.TransactionStatus, error) {
	return &provider.TransactionStatus{ExternalID: externalID, Status: "COMPLETED",
		RawData: map[string]interface{}{"id": externalID, "status": "COMPLETED"},
	}, nil
}

func (g *PayPalGateway) HealthCheck(ctx context.Context) error {
	if g.clientID == "" {
		return provider.ErrProviderUnavailable
	}
	return nil
}
