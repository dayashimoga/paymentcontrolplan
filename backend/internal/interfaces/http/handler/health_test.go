package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paymentbridge/pcp/internal/interfaces/http/handler"
)

func TestHealth_ReturnsOK(t *testing.T) {
	h := handler.NewHealthHandler(nil) // nil pool — health doesn't need DB

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "healthy" {
		t.Errorf("expected status healthy, got %s", body["status"])
	}
	if body["service"] != "pcp-api" {
		t.Errorf("expected service pcp-api, got %s", body["service"])
	}
}
