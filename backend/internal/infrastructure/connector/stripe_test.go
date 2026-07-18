package connector_test

import (
	"context"
	"testing"

	"github.com/paymentbridge/pcp/internal/domain/provider"
	"github.com/paymentbridge/pcp/internal/infrastructure/connector"
	"github.com/stretchr/testify/assert"
)

func TestStripeGateway(t *testing.T) {
	ctx := context.Background()
	gw := connector.NewStripeGateway(map[string]string{"api_key": "sk_test_123"})

	assert.Equal(t, provider.TypeStripe, gw.ProviderType())

	// HealthCheck
	err := gw.HealthCheck(ctx)
	assert.NoError(t, err)

	// Charge
	req := provider.ChargeRequest{
		Amount:   1000,
		Currency: "USD",
	}
	res, err := gw.Charge(ctx, req)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.ExternalID)
	assert.Equal(t, "succeeded", res.Status)

	// Refund
	refReq := provider.RefundRequest{
		ExternalID: res.ExternalID,
		Amount:     1000,
	}
	refRes, err := gw.Refund(ctx, refReq)
	assert.NoError(t, err)
	assert.Equal(t, "succeeded", refRes.Status)

	// Transaction Status
	status, err := gw.GetTransactionStatus(ctx, res.ExternalID)
	assert.NoError(t, err)
	assert.Equal(t, "succeeded", status.Status)
}
