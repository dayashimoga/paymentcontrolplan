package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paymentbridge/pcp/internal/interfaces/http/middleware"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRequestIDMiddleware(t *testing.T) {
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetRequestID(r.Context())
		assert.NotEmpty(t, reqID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
}

func TestRecoveryMiddleware(t *testing.T) {
	logger := zap.NewNop()
	recoveryMw := middleware.Recovery(logger)
	handler := recoveryMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
