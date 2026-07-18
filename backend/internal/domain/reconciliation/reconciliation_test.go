package reconciliation_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/payment"
	"github.com/paymentbridge/pcp/internal/domain/reconciliation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) Create(ctx context.Context, r *reconciliation.Record) error {
	return m.Called(ctx, r).Error(0)
}

func (m *mockRepo) GetByPayment(ctx context.Context, paymentID uuid.UUID) (*reconciliation.Record, error) {
	args := m.Called(ctx, paymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reconciliation.Record), args.Error(1)
}

func (m *mockRepo) ListUnmatched(ctx context.Context, offset, limit int) ([]*reconciliation.Record, int, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*reconciliation.Record), args.Int(1), args.Error(2)
}

func TestReconciliation_Matched(t *testing.T) {
	ctx := context.Background()
	repo := new(mockRepo)
	svc := reconciliation.NewService(repo)

	paymentID := uuid.New()
	providerID := uuid.New()

	repo.On("Create", ctx, mock.Anything).Return(nil)

	rec, err := svc.Reconcile(ctx, paymentID, providerID, 1000, payment.StatusCompleted, 1000, "completed")
	assert.NoError(t, err)
	assert.True(t, rec.IsMatched)
	assert.Empty(t, rec.Discrepancy)
}

func TestReconciliation_Discrepancy(t *testing.T) {
	ctx := context.Background()
	repo := new(mockRepo)
	svc := reconciliation.NewService(repo)

	paymentID := uuid.New()
	providerID := uuid.New()

	repo.On("Create", ctx, mock.Anything).Return(nil)

	rec, err := svc.Reconcile(ctx, paymentID, providerID, 1000, payment.StatusCompleted, 900, "completed")
	assert.NoError(t, err)
	assert.False(t, rec.IsMatched)
	assert.Equal(t, "amount_mismatch", rec.Discrepancy)
}
