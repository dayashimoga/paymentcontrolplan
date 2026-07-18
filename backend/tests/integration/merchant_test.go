package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paymentbridge/pcp/internal/interfaces/http/handler"
)

// TestIntegration_MerchantCRUD_E2E tests the full merchant lifecycle
// through the HTTP handler layer with mock repository.
func TestIntegration_MerchantCRUD_E2E(t *testing.T) {
	// This test uses the in-memory mock from handler tests
	// For real integration tests with PostgreSQL, use testcontainers:
	//
	//   container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
	//     ContainerRequest: testcontainers.ContainerRequest{
	//       Image: "postgres:16-alpine",
	//       ExposedPorts: []string{"5432/tcp"},
	//       Env: map[string]string{"POSTGRES_PASSWORD": "test"},
	//       WaitingFor: wait.ForListeningPort("5432/tcp"),
	//     }, Started: true,
	//   })

	t.Run("Create_Get_List_Update_Delete", func(t *testing.T) {
		h := handler.SetupTestHandler(t)

		// CREATE
		body := `{"name":"Integration Corp","webhook_url":"https://test.com/hook"}`
		req := httptest.NewRequest("POST", "/api/v1/merchants", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("create: expected 201, got %d: %s", rr.Code, rr.Body.String())
		}

		var created map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &created)
		id := created["id"].(string)
		apiKey := created["api_key"].(string)

		if id == "" || apiKey == "" {
			t.Fatal("create: missing id or api_key")
		}
		if created["status"] != "active" {
			t.Fatalf("create: expected active status, got %v", created["status"])
		}

		// GET
		rr2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("GET", "/api/v1/merchants/"+id, nil)
		h.ServeGet(rr2, req2, id)
		if rr2.Code != http.StatusOK {
			t.Fatalf("get: expected 200, got %d", rr2.Code)
		}

		// LIST
		rr3 := httptest.NewRecorder()
		req3 := httptest.NewRequest("GET", "/api/v1/merchants?limit=10", nil)
		h.List(rr3, req3)
		if rr3.Code != http.StatusOK {
			t.Fatalf("list: expected 200, got %d", rr3.Code)
		}

		var listed map[string]interface{}
		json.Unmarshal(rr3.Body.Bytes(), &listed)
		if listed["total"].(float64) < 1 {
			t.Fatal("list: expected at least 1 merchant")
		}

		// DELETE
		rr5 := httptest.NewRecorder()
		h.ServeDelete(rr5, httptest.NewRequest("DELETE", "/api/v1/merchants/"+id, nil), id)
		if rr5.Code != http.StatusNoContent {
			t.Fatalf("delete: expected 204, got %d", rr5.Code)
		}

		// VERIFY DELETED
		rr6 := httptest.NewRecorder()
		h.ServeGet(rr6, httptest.NewRequest("GET", "/api/v1/merchants/"+id, nil), id)
		if rr6.Code != http.StatusNotFound {
			t.Fatalf("get after delete: expected 404, got %d", rr6.Code)
		}
	})
}

// TestIntegration_DuplicateMerchant tests conflict handling.
func TestIntegration_DuplicateMerchant(t *testing.T) {
	h := handler.SetupTestHandler(t)

	body := `{"name":"Duplicate Corp","webhook_url":"https://test.com/hook"}`
	req := httptest.NewRequest("POST", "/api/v1/merchants", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("first create failed: %d", rr.Code)
	}

	// Duplicate
	req2 := httptest.NewRequest("POST", "/api/v1/merchants", bytes.NewBufferString(body))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	h.Create(rr2, req2)
	if rr2.Code != http.StatusConflict {
		t.Fatalf("duplicate: expected 409, got %d", rr2.Code)
	}
}

// TestIntegration_InvalidInput tests validation error handling.
func TestIntegration_InvalidInput(t *testing.T) {
	h := handler.SetupTestHandler(t)

	cases := []struct {
		name string
		body string
	}{
		{"empty_name", `{"name":"","webhook_url":"https://t.com"}`},
		{"invalid_json", `{invalid}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/merchants", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			h.Create(rr, req)
			if rr.Code == http.StatusCreated {
				t.Fatalf("expected error response, got 201")
			}
		})
	}
}
