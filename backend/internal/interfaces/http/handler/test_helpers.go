package handler

import (
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"context"

	appmch "github.com/paymentbridge/pcp/internal/application/merchant"
	"github.com/paymentbridge/pcp/internal/domain/merchant"
	"go.uber.org/zap"
)

// TestableHandler wraps MerchantHandler for integration testing.
type TestableHandler struct {
	*MerchantHandler
}

// SetupTestHandler creates a handler with in-memory repository for testing.
func SetupTestHandler(t *testing.T) *TestableHandler {
	t.Helper()
	repo := newMockMerchantRepo()
	svc := appmch.NewService(repo)
	logger, _ := zap.NewDevelopment()
	h := NewMerchantHandler(svc, nil, logger)
	return &TestableHandler{h}
}

// ServeGet dispatches to Get with a Chi URL param.
func (h *TestableHandler) ServeGet(rr *httptest.ResponseRecorder, req *http.Request, id string) {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	h.Get(rr, req)
}

// ServeDelete dispatches to Delete with a Chi URL param.
func (h *TestableHandler) ServeDelete(rr *httptest.ResponseRecorder, req *http.Request, id string) {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	h.Delete(rr, req)
}

// mockMerchantRepo is an in-memory merchant repository for testing.
type mockMerchantRepo struct {
	merchants map[uuid.UUID]*merchant.Merchant
}

func newMockMerchantRepo() *mockMerchantRepo {
	return &mockMerchantRepo{merchants: make(map[uuid.UUID]*merchant.Merchant)}
}

func (r *mockMerchantRepo) Create(_ context.Context, m *merchant.Merchant) error {
	for _, existing := range r.merchants {
		if existing.Name == m.Name {
			return merchant.ErrDuplicateMerchant
		}
	}
	r.merchants[m.ID] = m
	return nil
}

func (r *mockMerchantRepo) GetByID(_ context.Context, id uuid.UUID) (*merchant.Merchant, error) {
	if m, ok := r.merchants[id]; ok {
		return m, nil
	}
	return nil, merchant.ErrMerchantNotFound
}

func (r *mockMerchantRepo) GetByAPIKey(_ context.Context, apiKey string) (*merchant.Merchant, error) {
	for _, m := range r.merchants {
		if m.APIKey == apiKey {
			return m, nil
		}
	}
	return nil, merchant.ErrMerchantNotFound
}

func (r *mockMerchantRepo) List(_ context.Context, offset, limit int) ([]*merchant.Merchant, int, error) {
	result := make([]*merchant.Merchant, 0)
	for _, m := range r.merchants {
		result = append(result, m)
	}
	total := len(result)
	if offset > len(result) { offset = len(result) }
	end := offset + limit
	if end > len(result) { end = len(result) }
	return result[offset:end], total, nil
}

func (r *mockMerchantRepo) Update(_ context.Context, m *merchant.Merchant) error {
	if _, ok := r.merchants[m.ID]; !ok {
		return merchant.ErrMerchantNotFound
	}
	r.merchants[m.ID] = m
	return nil
}

func (r *mockMerchantRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := r.merchants[id]; !ok {
		return merchant.ErrMerchantNotFound
	}
	delete(r.merchants, id)
	return nil
}
