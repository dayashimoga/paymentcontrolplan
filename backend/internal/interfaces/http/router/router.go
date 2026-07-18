package router

import (
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/paymentbridge/pcp/internal/domain/auth"
	"github.com/paymentbridge/pcp/internal/domain/merchant"
	"github.com/paymentbridge/pcp/internal/interfaces/http/handler"
	"github.com/paymentbridge/pcp/internal/interfaces/http/middleware"
	"go.uber.org/zap"
)

// New creates and configures the Chi router with all middleware and route groups.
func New(
	logger *zap.Logger,
	tokenSvc auth.TokenService,
	merchantRepo merchant.Repository,
	healthHandler *handler.HealthHandler,
	merchantHandler *handler.MerchantHandler,
	providerHandler *handler.ProviderHandler,
	paymentHandler *handler.PaymentHandler,
	analyticsHandler *handler.AnalyticsHandler,
) *chi.Mux {
	r := chi.NewRouter()

	// Rate limiter: 100 requests per minute per key
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	// Idempotency store: 24h TTL
	idempotencyStore := middleware.NewIdempotencyStore(24 * time.Hour)

	// Global middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logging(logger))
	r.Use(chimw.RealIP)
	r.Use(chimw.Compress(5))
	r.Use(rateLimiter.RateLimit)

	// Health and readiness probes (no auth required)
	r.Get("/health", healthHandler.Health)
	r.Get("/ready", healthHandler.Ready)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(chimw.AllowContentType("application/json"))
		r.Use(idempotencyStore.Idempotency)

		// Public merchant management (no auth for create)
		r.Post("/merchants", merchantHandler.Create)

		// Auth token generation (public)
		r.Post("/auth/token", merchantHandler.GenerateToken)

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(tokenSvc, merchantRepo))

			// Merchants
			r.Get("/merchants", merchantHandler.List)
			r.Get("/merchants/{id}", merchantHandler.Get)
			r.Put("/merchants/{id}", merchantHandler.Update)
			r.Delete("/merchants/{id}", merchantHandler.Delete)

			// Providers
			r.Post("/providers", providerHandler.Create)
			r.Get("/providers", providerHandler.List)
			r.Get("/providers/{id}", providerHandler.Get)
			r.Delete("/providers/{id}", providerHandler.Delete)

			// Payments
			r.Post("/payments", paymentHandler.Create)
			r.Get("/payments", paymentHandler.List)
			r.Get("/payments/{id}", paymentHandler.Get)
			r.Post("/payments/{id}/refund", paymentHandler.Refund)

			// Analytics
			r.Get("/analytics/summary", analyticsHandler.GetSummary)
			r.Get("/analytics/providers", analyticsHandler.GetProviderStats)
		})
	})

	return r
}
