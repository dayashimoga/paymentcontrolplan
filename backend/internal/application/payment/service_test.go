package payment_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	apppay "github.com/paymentbridge/pcp/internal/application/payment"
	"github.com/paymentbridge/pcp/internal/domain/payment"
	"github.com/paymentbridge/pcp/internal/domain/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestPaymentService_Create(t *testing.T) {
	ctx := context.Background()
	payRepo := new(mockPaymentRepo)
	provRepo := new(mockProviderRepo)
	eng := new(mockEngine)

	svc := apppay.NewService(payRepo, provRepo, eng)

	gw := new(mockGateway)
	svc.RegisterGateway(provider.TypeStripe, gw)

	merchantID := uuid.New()
	pObj := &provider.Provider{ID: uuid.New(), Name: "Stripe", Type: provider.TypeStripe, Status: provider.StatusActive, Priority: 1}

	payRepo.On("GetByIdempotencyKey", ctx, merchantID, "idem_123").Return(nil, payment.ErrPaymentNotFound)
	eng.On("SelectProvider", ctx, merchantID, int64(1000), "USD").Return(pObj, nil)
	payRepo.On("Create", ctx, mock.Anything).Return(nil)
	payRepo.On("Update", ctx, mock.Anything).Return(nil)

	gw.On("Charge", ctx, mock.Anything).Return(&provider.ChargeResponse{
		ExternalID:   "ch_123",
		Status:       "succeeded",
		ProviderType: provider.TypeStripe,
	}, nil)

	p, err := svc.Create(ctx, apppay.CreateInput{
		MerchantID:     merchantID,
		Amount:         1000,
		Currency:       "USD",
		Description:    "Test charge",
		IdempotencyKey: "idem_123",
	})
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, payment.StatusCompleted, p.Status)
	assert.Equal(t, "ch_123", p.ExternalID)
}

func TestPaymentService_Refund(t *testing.T) {
	ctx := context.Background()
	payRepo := new(mockPaymentRepo)
	provRepo := new(mockProviderRepo)
	eng := new(mockEngine)

	svc := apppay.NewService(payRepo, provRepo, eng)

	gw := new(mockGateway)
	svc.RegisterGateway(provider.TypeStripe, gw)

	merchantID := uuid.New()
	provID := uuid.New()
	pObj := &provider.Provider{ID: provID, Name: "Stripe", Type: provider.TypeStripe, Status: provider.StatusActive, Priority: 1}

	payObj := &payment.Payment{
		ID:          uuid.New(),
		MerchantID:  merchantID,
		ProviderID:  provID,
		Amount:      1000,
		Currency:    "USD",
		Status:      payment.StatusCompleted,
		ExternalID:  "ch_123",
	}

	payRepo.On("GetByID", ctx, payObj.ID).Return(payObj, nil)
	provRepo.On("GetByID", ctx, provID).Return(pObj, nil)

	gw.On("Refund", ctx, mock.Anything).Return(&provider.RefundResponse{
		ExternalID: "ch_123",
		Status:     "succeeded",
		RefundID:   "re_123",
	}, nil)

	payRepo.On("Update", ctx, mock.Anything).Return(nil)

	p, err := svc.Refund(ctx, payObj.ID)
	assert.NoError(t, err)
	assert.Equal(t, payment.StatusRefunded, p.Status)
}
