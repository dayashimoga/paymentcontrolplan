package analytics_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/application/analytics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAnalyticsRepo struct {
	mock.Mock
}

func (m *mockAnalyticsRepo) GetSummary(ctx context.Context, merchantID uuid.UUID, from, to time.Time) (*analytics.Summary, error) {
	args := m.Called(ctx, merchantID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*analytics.Summary), args.Error(1)
}

func (m *mockAnalyticsRepo) GetProviderStats(ctx context.Context, merchantID uuid.UUID, from, to time.Time) ([]*analytics.ProviderStats, error) {
	args := m.Called(ctx, merchantID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*analytics.ProviderStats), args.Error(1)
}

func TestAnalyticsService(t *testing.T) {
	ctx := context.Background()
	repo := new(mockAnalyticsRepo)
	svc := analytics.NewService(repo)

	merchantID := uuid.New()
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	expectedSummary := &analytics.Summary{
		TotalPayments:     100,
		CompletedPayments: 95,
		FailedPayments:    5,
		TotalAmount:       500000,
		SuccessRate:       95.0,
	}

	expectedStats := []*analytics.ProviderStats{
		{
			ProviderID:   uuid.New(),
			ProviderName: "Stripe",
			TotalCharges: 100,
			SuccessCount: 95,
			FailureCount: 5,
			TotalAmount:  500000,
			SuccessRate:  95.0,
			AvgLatencyMs: 150,
		},
	}

	repo.On("GetSummary", ctx, merchantID, mock.Anything, mock.Anything).Return(expectedSummary, nil)
	repo.On("GetProviderStats", ctx, merchantID, mock.Anything, mock.Anything).Return(expectedStats, nil)

	s, err := svc.GetSummary(ctx, merchantID, from, to)
	assert.NoError(t, err)
	assert.Equal(t, expectedSummary.TotalPayments, s.TotalPayments)

	p, err := svc.GetProviderStats(ctx, merchantID, from, to)
	assert.NoError(t, err)
	assert.Len(t, p, 1)
	assert.Equal(t, "Stripe", p[0].ProviderName)
}
