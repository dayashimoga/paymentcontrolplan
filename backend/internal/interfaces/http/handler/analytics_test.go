package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	appanalytics "github.com/paymentbridge/pcp/internal/application/analytics"
	"github.com/paymentbridge/pcp/internal/domain/merchant"
	"github.com/paymentbridge/pcp/internal/interfaces/http/handler"
	"github.com/paymentbridge/pcp/internal/interfaces/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockAnalyticsRepo struct {
	mock.Mock
}

func (m *mockAnalyticsRepo) GetSummary(ctx context.Context, merchantID uuid.UUID, from, to time.Time) (*appanalytics.Summary, error) {
	args := m.Called(ctx, merchantID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*appanalytics.Summary), args.Error(1)
}

func (m *mockAnalyticsRepo) GetProviderStats(ctx context.Context, merchantID uuid.UUID, from, to time.Time) ([]*appanalytics.ProviderStats, error) {
	args := m.Called(ctx, merchantID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*appanalytics.ProviderStats), args.Error(1)
}

func testAnalyticsRouter() (*chi.Mux, *mockAnalyticsRepo, uuid.UUID) {
	repo := new(mockAnalyticsRepo)
	svc := appanalytics.NewService(repo)
	h := handler.NewAnalyticsHandler(svc, zap.NewNop())

	merchantID := uuid.New()
	mObj := &merchant.Merchant{ID: merchantID, Name: "Test Merchant", Status: merchant.StatusActive}

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(middleware.WithMerchant(r.Context(), mObj)))
		})
	})
	r.Get("/api/v1/analytics/summary", h.GetSummary)
	r.Get("/api/v1/analytics/providers", h.GetProviderStats)

	return r, repo, merchantID
}

func TestAnalyticsHandler_GetSummary(t *testing.T) {
	r, repo, _ := testAnalyticsRouter()
	repo.On("GetSummary", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&appanalytics.Summary{}, nil)

	req := httptest.NewRequest("GET", "/api/v1/analytics/summary?days=7", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAnalyticsHandler_GetProviderStats(t *testing.T) {
	r, repo, _ := testAnalyticsRouter()
	repo.On("GetProviderStats", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*appanalytics.ProviderStats{}, nil)

	req := httptest.NewRequest("GET", "/api/v1/analytics/providers?days=7", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}
