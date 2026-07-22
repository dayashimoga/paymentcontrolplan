package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	appprov "github.com/paymentbridge/pcp/internal/application/provider"
	"github.com/paymentbridge/pcp/internal/interfaces/http/handler"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func testProviderRouter() (*chi.Mux, *appprov.Service) {
	provRepo := new(mockProviderRepo)
	svc := appprov.NewService(provRepo)
	h := handler.NewProviderHandler(svc, nil, zap.NewNop())

	r := chi.NewRouter()
	r.Post("/api/v1/providers", h.Create)
	r.Get("/api/v1/providers", h.List)
	r.Get("/api/v1/providers/{id}", h.Get)
	r.Delete("/api/v1/providers/{id}", h.Delete)

	return r, svc
}

func TestProviderHandler_Create_InvalidJSON(t *testing.T) {
	r, _ := testProviderRouter()

	req := httptest.NewRequest("POST", "/api/v1/providers", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestProviderHandler_Get_InvalidUUID(t *testing.T) {
	r, _ := testProviderRouter()

	req := httptest.NewRequest("GET", "/api/v1/providers/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestProviderHandler_Delete_InvalidUUID(t *testing.T) {
	r, _ := testProviderRouter()

	req := httptest.NewRequest("DELETE", "/api/v1/providers/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
