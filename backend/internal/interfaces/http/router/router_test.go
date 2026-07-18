package router_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	appprov "github.com/paymentbridge/pcp/internal/application/provider"
	"github.com/paymentbridge/pcp/internal/domain/auth"
	"github.com/paymentbridge/pcp/internal/interfaces/http/handler"
	"github.com/paymentbridge/pcp/internal/interfaces/http/router"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type dummyTokenSvc struct{}

func (d *dummyTokenSvc) GenerateToken(ctx context.Context, merchantID uuid.UUID, merchantName string) (string, error) {
	return "token", nil
}
func (d *dummyTokenSvc) ValidateToken(ctx context.Context, token string) (*auth.Claims, error) {
	return &auth.Claims{MerchantID: uuid.New()}, nil
}

func TestNewRouter(t *testing.T) {
	logger := zap.NewNop()
	tokenSvc := &dummyTokenSvc{}

	healthH := handler.NewHealthHandler(nil)
	merchantH := handler.NewMerchantHandler(nil, logger)
	provH := handler.NewProviderHandler(appprov.NewService(nil), logger)
	paymentH := handler.NewPaymentHandler(nil, logger)
	analyticsH := handler.NewAnalyticsHandler(nil, logger)

	r := router.New(
		logger,
		tokenSvc,
		nil,
		healthH,
		merchantH,
		provH,
		paymentH,
		analyticsH,
	)

	assert.NotNil(t, r)

	// Test health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
