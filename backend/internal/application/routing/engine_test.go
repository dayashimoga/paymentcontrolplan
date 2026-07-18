package routing_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	approuting "github.com/paymentbridge/pcp/internal/application/routing"
	"github.com/paymentbridge/pcp/internal/domain/provider"
	"github.com/paymentbridge/pcp/internal/domain/routing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRoutingRuleRepo struct {
	mock.Mock
}

func (m *mockRoutingRuleRepo) GetByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*routing.Rule, error) {
	args := m.Called(ctx, merchantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*routing.Rule), args.Error(1)
}

func (m *mockRoutingRuleRepo) Create(ctx context.Context, rule *routing.Rule) error {
	return m.Called(ctx, rule).Error(0)
}

func (m *mockRoutingRuleRepo) Update(ctx context.Context, rule *routing.Rule) error {
	return m.Called(ctx, rule).Error(0)
}

func (m *mockRoutingRuleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
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

func TestRoutingEngine_Route_Priority(t *testing.T) {
	ctx := context.Background()
	ruleRepo := new(mockRoutingRuleRepo)
	provRepo := new(mockProviderRepo)
	engine := approuting.NewEngine(ruleRepo, provRepo)

	merchantID := uuid.New()
	pStripe := &provider.Provider{ID: uuid.New(), Name: "Stripe", Type: provider.TypeStripe, Status: provider.StatusActive, Priority: 1}
	pPayPal := &provider.Provider{ID: uuid.New(), Name: "PayPal", Type: provider.TypePayPal, Status: provider.StatusActive, Priority: 2}

	rule1 := &routing.Rule{ID: uuid.New(), MerchantID: merchantID, ProviderID: pStripe.ID, Priority: 1, Weight: 100, Currency: "USD", MinAmount: 0, MaxAmount: 10000}
	rule2 := &routing.Rule{ID: uuid.New(), MerchantID: merchantID, ProviderID: pPayPal.ID, Priority: 2, Weight: 50, Currency: "USD", MinAmount: 0, MaxAmount: 10000}

	ruleRepo.On("GetByMerchant", ctx, merchantID).Return([]*routing.Rule{rule1, rule2}, nil)
	provRepo.On("GetByID", ctx, pStripe.ID).Return(pStripe, nil)
	provRepo.On("ListActive", ctx).Return([]*provider.Provider{pStripe, pPayPal}, nil)

	selected, err := engine.SelectProvider(ctx, merchantID, 5000, "USD")
	assert.NoError(t, err)
	assert.Equal(t, pStripe.ID, selected.ID)
}

func TestRoutingEngine_Route_FallbackToAvailableProviders(t *testing.T) {
	ctx := context.Background()
	ruleRepo := new(mockRoutingRuleRepo)
	provRepo := new(mockProviderRepo)
	engine := approuting.NewEngine(ruleRepo, provRepo)

	merchantID := uuid.New()
	pDefault := &provider.Provider{ID: uuid.New(), Name: "Default Stripe", Type: provider.TypeStripe, Status: provider.StatusActive, Priority: 1}

	// No matching rules
	ruleRepo.On("GetByMerchant", ctx, merchantID).Return([]*routing.Rule{}, nil)
	provRepo.On("ListActive", ctx).Return([]*provider.Provider{pDefault}, nil)

	selected, err := engine.SelectProvider(ctx, merchantID, 1000, "EUR")
	assert.NoError(t, err)
	assert.Equal(t, pDefault.ID, selected.ID)
}
