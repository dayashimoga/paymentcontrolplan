// Package observability provides Prometheus metrics and OpenTelemetry tracing.
package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metric collectors for PCP.
type Metrics struct {
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	PaymentsTotal        *prometheus.CounterVec
	PaymentAmount        *prometheus.HistogramVec
	ProviderLatency      *prometheus.HistogramVec
	CircuitBreakerState  *prometheus.GaugeVec
	ActiveConnections    prometheus.Gauge
}

// NewMetrics registers and returns all PCP Prometheus metrics.
func NewMetrics() *Metrics {
	return &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "pcp_http_requests_total",
			Help: "Total HTTP requests by method, path, and status code",
		}, []string{"method", "path", "status"}),

		HTTPRequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "pcp_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path"}),

		PaymentsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "pcp_payments_total",
			Help: "Total payment transactions by status and provider",
		}, []string{"status", "provider", "currency"}),

		PaymentAmount: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "pcp_payment_amount",
			Help:    "Payment amounts in cents",
			Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000},
		}, []string{"currency"}),

		ProviderLatency: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "pcp_provider_latency_seconds",
			Help:    "Payment provider response latency",
			Buckets: prometheus.DefBuckets,
		}, []string{"provider", "operation"}),

		CircuitBreakerState: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pcp_circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
		}, []string{"name"}),

		ActiveConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "pcp_active_connections",
			Help: "Number of active HTTP connections",
		}),
	}
}

// Handler returns the Prometheus metrics HTTP handler.
func Handler() http.Handler {
	return promhttp.Handler()
}

// MetricsMiddleware records HTTP request metrics.
func (m *Metrics) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		m.ActiveConnections.Inc()
		defer m.ActiveConnections.Dec()

		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		duration := time.Since(start).Seconds()
		path := r.URL.Path
		method := r.Method
		status := strconv.Itoa(rec.statusCode)

		m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}
