package connector_test

import (
	"context"
	"testing"

	"github.com/paymentbridge/pcp/internal/domain/provider"
	"github.com/paymentbridge/pcp/internal/infrastructure/connector"
	"github.com/stretchr/testify/assert"
)

func TestPayPalGateway(t *testing.T) {
	ctx := context.Background()
	gw := connector.NewPayPalGateway(map[string]string{
		"client_id":     "test_client",
		"client_secret": "test_secret",
	})

	assert.Equal(t, provider.TypePayPal, gw.ProviderType())

	// HealthCheck
	err := gw.HealthCheck(ctx)
	assert.NoError(t, err)

	// Charge
	req := provider.ChargeRequest{
		Amount:   2500,
		Currency: "USD",
	}
	res, err := gw.Charge(ctx, req)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.ExternalID)
	assert.Equal(t, "COMPLETED", res.Status)

	// Refund
	refReq := provider.RefundRequest{
		ExternalID: res.ExternalID,
		Amount:     2500,
	}
	refRes, err := gw.Refund(ctx, refReq)
	assert.NoError(t, err)
	assert.Equal(t, "COMPLETED", refRes.Status)

	// Transaction Status
	status, err := gw.GetTransactionStatus(ctx, res.ExternalID)
	assert.NoError(t, err)
	assert.Equal(t, "COMPLETED", status.Status)
}
