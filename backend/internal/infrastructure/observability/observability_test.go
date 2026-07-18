package observability_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paymentbridge/pcp/internal/infrastructure/observability"
	"github.com/stretchr/testify/assert"
)

func TestMetricsMiddleware(t *testing.T) {
	metrics := observability.NewMetrics()
	assert.NotNil(t, metrics)

	mw := metrics.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/payments", nil)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
