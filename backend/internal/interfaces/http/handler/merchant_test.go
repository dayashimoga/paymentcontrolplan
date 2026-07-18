package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	appmch "github.com/paymentbridge/pcp/internal/application/merchant"
	"github.com/paymentbridge/pcp/internal/interfaces/http/dto"
	"github.com/paymentbridge/pcp/internal/interfaces/http/handler"
	"go.uber.org/zap"
)

// testRouter creates a router with merchant handler backed by the mock repo from service_test.
func testMerchantRouter() (*chi.Mux, *appmch.Service) {
	repo := newTestMockRepo()
	svc := appmch.NewService(repo)
	logger := zap.NewNop()
	h := handler.NewMerchantHandler(svc, logger)

	r := chi.NewRouter()
	r.Post("/api/v1/merchants", h.Create)
	r.Get("/api/v1/merchants", h.List)
	r.Get("/api/v1/merchants/{id}", h.Get)
	r.Put("/api/v1/merchants/{id}", h.Update)
	r.Delete("/api/v1/merchants/{id}", h.Delete)

	return r, svc
}

func TestMerchantHandler_Create(t *testing.T) {
	r, _ := testMerchantRouter()

	body := `{"name":"Test Merchant","webhook_url":"https://example.com/webhook"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/merchants", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp dto.MerchantResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Name != "Test Merchant" {
		t.Errorf("expected name Test Merchant, got %s", resp.Name)
	}
	if resp.APIKey == "" {
		t.Error("expected API key to be set")
	}
}

func TestMerchantHandler_Create_EmptyName(t *testing.T) {
	r, _ := testMerchantRouter()

	body := `{"name":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/merchants", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMerchantHandler_Create_InvalidJSON(t *testing.T) {
	r, _ := testMerchantRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/merchants", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMerchantHandler_Get(t *testing.T) {
	r, svc := testMerchantRouter()

	// Create a merchant first
	m, _ := svc.Create(nil, appmch.CreateInput{Name: "GetTest"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/merchants/"+m.ID.String(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMerchantHandler_Get_InvalidUUID(t *testing.T) {
	r, _ := testMerchantRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/merchants/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMerchantHandler_List(t *testing.T) {
	r, svc := testMerchantRouter()

	svc.Create(nil, appmch.CreateInput{Name: "ListTest1"})
	svc.Create(nil, appmch.CreateInput{Name: "ListTest2"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/merchants?offset=0&limit=10", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp dto.ListMerchantsResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Total < 2 {
		t.Errorf("expected at least 2 merchants, got %d", resp.Total)
	}
}

func TestMerchantHandler_Delete(t *testing.T) {
	r, svc := testMerchantRouter()

	m, _ := svc.Create(nil, appmch.CreateInput{Name: "DeleteTest"})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/merchants/"+m.ID.String(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}
