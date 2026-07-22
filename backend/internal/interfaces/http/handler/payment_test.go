package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	apppay "github.com/paymentbridge/pcp/internal/application/payment"
	"github.com/paymentbridge/pcp/internal/domain/merchant"
	"github.com/paymentbridge/pcp/internal/domain/payment"
	"github.com/paymentbridge/pcp/internal/domain/provider"
	"github.com/paymentbridge/pcp/internal/interfaces/http/handler"
	"github.com/paymentbridge/pcp/internal/interfaces/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockPaymentRepo struct {
	mock.Mock
}

func (m *mockPaymentRepo) Create(ctx context.Context, p *payment.Payment) error {
	return m.Called(ctx, p).Error(0)
}

func (m *mockPaymentRepo) GetByID(ctx context.Context, id uuid.UUID) (*payment.Payment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payment.Payment), args.Error(1)
}

func (m *mockPaymentRepo) GetByIdempotencyKey(ctx context.Context, merchantID uuid.UUID, key string) (*payment.Payment, error) {
	args := m.Called(ctx, merchantID, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payment.Payment), args.Error(1)
}

func (m *mockPaymentRepo) List(ctx context.Context, merchantID uuid.UUID, offset, limit int) ([]*payment.Payment, int, error) {
	args := m.Called(ctx, merchantID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*payment.Payment), args.Int(1), args.Error(2)
}

func (m *mockPaymentRepo) Update(ctx context.Context, p *payment.Payment) error {
	return m.Called(ctx, p).Error(0)
}

type mockProviderRepo struct {
	mock.Mock
}

func (m *mockProviderRepo) Create(ctx context.Context, pr *provider.Provider) error {
	return m.Called(ctx, pr).Error(0)
}

func (m *mockProviderRepo) GetByID(ctx context.Context, id uuid.UUID) (*provider.Provider, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Provider), args.Error(1)
}

func (m *mockProviderRepo) List(ctx context.Context, offset, limit int) ([]*provider.Provider, int, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*provider.Provider), args.Int(1), args.Error(2)
}

func (m *mockProviderRepo) ListActive(ctx context.Context) ([]*provider.Provider, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*provider.Provider), args.Error(1)
}

func (m *mockProviderRepo) Update(ctx context.Context, pr *provider.Provider) error {
	return m.Called(ctx, pr).Error(0)
}

func (m *mockProviderRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type mockGateway struct {
	mock.Mock
}

func (m *mockGateway) ProviderType() provider.Type { return provider.TypeStripe }

func (m *mockGateway) Charge(ctx context.Context, req provider.ChargeRequest) (*provider.ChargeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.ChargeResponse), args.Error(1)
}

func (m *mockGateway) Refund(ctx context.Context, req provider.RefundRequest) (*provider.RefundResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.RefundResponse), args.Error(1)
}

func (m *mockGateway) GetTransactionStatus(ctx context.Context, externalID string) (*provider.TransactionStatus, error) {
	args := m.Called(ctx, externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.TransactionStatus), args.Error(1)
}

func (m *mockGateway) HealthCheck(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type mockEngine struct {
	mock.Mock
}

func (m *mockEngine) SelectProvider(ctx context.Context, merchantID uuid.UUID, amount int64, currency string) (*provider.Provider, error) {
	args := m.Called(ctx, merchantID, amount, currency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Provider), args.Error(1)
}

func (m *mockEngine) SelectCandidateProviders(ctx context.Context, merchantID uuid.UUID, amount int64, currency string) ([]*provider.Provider, error) {
	p, err := m.SelectProvider(ctx, merchantID, amount, currency)
	if err != nil {
		return nil, err
	}
	return []*provider.Provider{p}, nil
}

func testPaymentRouter() (*chi.Mux, *mockPaymentRepo, *mockEngine, *mockGateway, uuid.UUID) {
	payRepo := new(mockPaymentRepo)
	provRepo := new(mockProviderRepo)
	eng := new(mockEngine)

	svc := apppay.NewService(payRepo, provRepo, eng)
	gw := new(mockGateway)
	svc.RegisterGateway(provider.TypeStripe, gw)

	h := handler.NewPaymentHandler(svc, nil, nil, zap.NewNop())
	merchantID := uuid.New()
	mObj := &merchant.Merchant{ID: merchantID, Name: "Test Merchant", Status: merchant.StatusActive}

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(middleware.WithMerchant(r.Context(), mObj)))
		})
	})
	r.Post("/api/v1/payments", h.Create)
	r.Get("/api/v1/payments", h.List)
	r.Get("/api/v1/payments/{id}", h.Get)
	r.Post("/api/v1/payments/{id}/refund", h.Refund)

	return r, payRepo, eng, gw, merchantID
}

func TestPaymentHandler_Create(t *testing.T) {
	r, payRepo, eng, gw, _ := testPaymentRouter()

	payRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	payRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
	eng.On("SelectProvider", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&provider.Provider{ID: uuid.New(), Type: provider.TypeStripe}, nil)
	gw.On("Charge", mock.Anything, mock.Anything).Return(&provider.ChargeResponse{ExternalID: "ext_123", Status: "succeeded"}, nil)

	body := `{"amount":1000,"currency":"USD","description":"Test Payment"}`
	req := httptest.NewRequest("POST", "/api/v1/payments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestPaymentHandler_Get_InvalidUUID(t *testing.T) {
	r, _, _, _, _ := testPaymentRouter()

	req := httptest.NewRequest("GET", "/api/v1/payments/invalid-uuid", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
