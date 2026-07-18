package provider_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	appprov "github.com/paymentbridge/pcp/internal/application/provider"
	"github.com/paymentbridge/pcp/internal/domain/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestProviderService_Create(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProviderRepo)
	svc := appprov.NewService(repo)

	repo.On("Create", ctx, mock.Anything).Return(nil)

	p, err := svc.Create(ctx, appprov.CreateInput{
		Name:     "Stripe Main",
		Type:     provider.TypeStripe,
		Config:   map[string]string{"api_key": "sk_test_123"},
		Priority: 1,
	})
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "Stripe Main", p.Name)
	assert.Equal(t, provider.TypeStripe, p.Type)
}

func TestProviderService_Get_List_Delete(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProviderRepo)
	svc := appprov.NewService(repo)

	pObj := &provider.Provider{
		ID:       uuid.New(),
		Name:     "PayPal Main",
		Type:     provider.TypePayPal,
		Status:   provider.StatusActive,
		Priority: 2,
	}

	repo.On("GetByID", ctx, pObj.ID).Return(pObj, nil)
	repo.On("List", ctx, 0, 10).Return([]*provider.Provider{pObj}, 1, nil)
	repo.On("Delete", ctx, pObj.ID).Return(nil)

	res, err := svc.GetByID(ctx, pObj.ID)
	assert.NoError(t, err)
	assert.Equal(t, pObj.ID, res.ID)

	list, total, err := svc.List(ctx, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, list, 1)

	err = svc.Delete(ctx, pObj.ID)
	assert.NoError(t, err)
}
