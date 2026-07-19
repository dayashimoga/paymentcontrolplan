package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	appmch "github.com/paymentbridge/pcp/internal/application/merchant"
	apppay "github.com/paymentbridge/pcp/internal/application/payment"
	appprov "github.com/paymentbridge/pcp/internal/application/provider"
	"github.com/paymentbridge/pcp/internal/application/analytics"
	approuting "github.com/paymentbridge/pcp/internal/application/routing"
	"github.com/paymentbridge/pcp/internal/domain/audit"
	"github.com/paymentbridge/pcp/internal/domain/provider"
	infraauth "github.com/paymentbridge/pcp/internal/infrastructure/auth"
	"github.com/paymentbridge/pcp/internal/infrastructure/cache"
	"github.com/paymentbridge/pcp/internal/infrastructure/config"
	"github.com/paymentbridge/pcp/internal/infrastructure/connector"
	"github.com/paymentbridge/pcp/internal/infrastructure/logging"
	"github.com/paymentbridge/pcp/internal/infrastructure/observability"
	"github.com/paymentbridge/pcp/internal/infrastructure/persistence"
	"github.com/paymentbridge/pcp/internal/interfaces/http/handler"
	"github.com/paymentbridge/pcp/internal/interfaces/http/router"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	logger, err := logging.NewLogger(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		return fmt.Errorf("initializing logger: %w", err)
	}
	defer func() { _ = logger.Sync() }()

	logger.Info("starting PCP API server", zap.String("host", cfg.Server.Host), zap.Int("port", cfg.Server.Port))

	// Database
	ctx := context.Background()
	pool, err := persistence.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()
	logger.Info("connected to PostgreSQL")

	// Redis (optional, non-fatal if unavailable)
	redisClient, err := cache.NewRedisClient(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Warn("redis unavailable, using in-memory fallbacks", zap.Error(err))
	} else {
		defer redisClient.Close()
		logger.Info("connected to Redis")
	}

	// Prometheus metrics
	metrics := observability.NewMetrics()
	logger.Info("prometheus metrics initialized")

	// JWT Service
	jwtService := infraauth.NewJWTService(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.Expiration)

	// Repositories
	merchantRepo := persistence.NewPostgresMerchantRepository(pool)
	providerRepo := persistence.NewPostgresProviderRepository(pool)
	paymentRepo := persistence.NewPostgresPaymentRepository(pool)
	routingRuleRepo := persistence.NewPostgresRoutingRuleRepository(pool)
	analyticsRepo := persistence.NewPostgresAnalyticsRepository(pool)
	auditRepo := persistence.NewPostgresAuditRepository(pool)
	_ = persistence.NewPostgresWebhookRepository(pool)           // webhook repo ready for Sprint 2
	_ = persistence.NewPostgresReconciliationRepository(pool)    // reconciliation repo ready for Sprint 2

	// Application services
	merchantService := appmch.NewService(merchantRepo)
	providerService := appprov.NewService(providerRepo)
	routingEngine := approuting.NewEngine(routingRuleRepo, providerRepo)
	paymentService := apppay.NewService(paymentRepo, providerRepo, routingEngine)
	analyticsService := analytics.NewService(analyticsRepo)
	auditService := audit.NewService(auditRepo)

	// Register payment gateways
	stripeGw := connector.NewStripeGateway(map[string]string{"api_key": "sk_test_default"})
	paypalGw := connector.NewPayPalGateway(map[string]string{"client_id": "test", "client_secret": "test"})
	providerService.RegisterGateway(provider.TypeStripe, stripeGw)
	providerService.RegisterGateway(provider.TypePayPal, paypalGw)
	paymentService.RegisterGateway(provider.TypeStripe, stripeGw)
	paymentService.RegisterGateway(provider.TypePayPal, paypalGw)

	// HTTP handlers
	healthHandler := handler.NewHealthHandler(pool)
	merchantHandler := handler.NewMerchantHandler(merchantService, auditService, logger)
	providerHandler := handler.NewProviderHandler(providerService, logger)
	paymentHandler := handler.NewPaymentHandler(paymentService, logger)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService, logger)

	// Build router
	r := router.New(logger, jwtService, merchantRepo, healthHandler, merchantHandler, providerHandler, paymentHandler, analyticsHandler, metrics)

	// HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr: addr, Handler: r,
		ReadTimeout: cfg.Server.ReadTimeout, WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout: 2 * time.Minute,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("HTTP server listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	sig := <-quit
	logger.Info("shutdown signal received", zap.String("signal", sig.String()))

	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	logger.Info("server stopped gracefully")
	return nil
}
