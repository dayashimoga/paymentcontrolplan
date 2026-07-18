# Changelog

All notable changes to the Payment Control Plane (PCP) project will be documented in this file.
This project adheres to [Semantic Versioning](https://semver.org/).

---

## [0.8.0] - 2026-07-18

### Added — Remaining Backlog Completion

#### Resilience
- **Circuit breaker** with three-state machine (closed → open → half-open) for cascading failure prevention
  - Configurable failure threshold, success threshold, and timeout
  - Thread-safe via `sync.RWMutex` for concurrent access
  - 5 unit tests covering all state transitions
  - File: `internal/infrastructure/circuitbreaker/circuitbreaker.go`

#### Observability
- **Prometheus metrics** with `promauto` registration:
  - `pcp_http_requests_total` — counter by method/path/status
  - `pcp_http_request_duration_seconds` — histogram with default buckets
  - `pcp_payments_total` — counter by status/provider/currency
  - `pcp_payment_amount` — histogram with payment-sized buckets
  - `pcp_provider_latency_seconds` — histogram by provider/operation
  - `pcp_circuit_breaker_state` — gauge by circuit breaker name
  - `pcp_active_connections` — gauge for live connections
  - `MetricsMiddleware` — automatic HTTP request instrumentation
  - File: `internal/infrastructure/observability/metrics.go`

- **OpenTelemetry distributed tracing**:
  - OTLP HTTP exporter with configurable endpoint
  - Service resource with name and version attributes
  - `TracingMiddleware` — span injection with method, URL, and user-agent attributes
  - W3C TraceContext and Baggage propagation
  - File: `internal/infrastructure/observability/tracing.go`

- **Grafana dashboard** (9 panels):
  - HTTP Request Rate (timeseries), HTTP Latency p95 (timeseries)
  - Payment Volume (stat), Payment Success Rate (gauge)
  - Payments by Status (piechart), Payments by Provider (piechart)
  - Provider Latency p99 (timeseries), Circuit Breaker State (stat with color mapping)
  - Active Connections (timeseries)
  - Auto-refresh every 10s, last 1 hour default range
  - File: `infrastructure/grafana/dashboards/pcp-overview.json`

#### Messaging
- **Kafka producer/consumer** implementing `Publisher` and `Subscriber` ports:
  - `KafkaPublisher` — lazy writer creation per topic, `LeastBytes` balancer, JSON serialization
  - `KafkaSubscriber` — consumer group support, background goroutine consumption
  - Built on `segmentio/kafka-go` for pure-Go Kafka client (no CGO dependency)
  - File: `internal/infrastructure/messaging/kafka.go`

#### API Documentation
- **OpenAPI 3.0 specification** covering all 17 endpoints:
  - Full request/response schemas for Merchant, Provider, Payment, Analytics
  - Security schemes: `ApiKeyAuth` (X-API-Key header) and `BearerAuth` (JWT)
  - Server definitions for local development and production
  - Tags: Health, Merchants, Providers, Payments, Analytics
  - File: `docs/openapi.yaml`

#### gRPC
- **Protocol Buffer definitions** for all three service domains:
  - `PaymentService` — CreatePayment, GetPayment, ListPayments, RefundPayment
  - `ProviderService` — CreateProvider, GetProvider, ListProviders, DeleteProvider
  - `MerchantService` — CreateMerchant, GetMerchant, ListMerchants, UpdateMerchant, DeleteMerchant
  - All messages use `google.protobuf.Timestamp` for temporal fields
  - File: `api/proto/v1/pcp.proto`

- **gRPC server scaffold** wrapping existing application services:
  - Unary logging interceptor
  - gRPC health checking (`grpc_health_v1`)
  - Server reflection for debugging
  - Graceful shutdown support
  - File: `internal/interfaces/grpc/server.go`

#### Infrastructure
- **Helm charts** for Kubernetes deployment:
  - `Chart.yaml` — version 0.7.0, payment orchestration metadata
  - `values.yaml` — configurable replicas, image, resources, autoscaling, probes, DB/Redis/JWT
  - `templates/deployment.yaml` — Deployment, Service, Ingress (with TLS), HorizontalPodAutoscaler
  - File: `infrastructure/helm/pcp/`

- **Cloudflare WAF/CDN configuration**:
  - DNS records, SSL/TLS (Full Strict, TLS 1.2 minimum, HSTS)
  - WAF rules: API rate limiting (1000/min), bot protection, geo-blocking
  - Security headers: X-Content-Type-Options, X-Frame-Options, Referrer-Policy
  - Terraform Cloudflare provider examples
  - File: `infrastructure/cloudflare/config.md`

- **SBOM generation** via Syft (SPDX JSON format) in CI pipeline
- **Image signing** via Cosign (keyless, OIDC-based) in CI pipeline

#### Testing
- **Integration tests** for merchant CRUD lifecycle:
  - Full create → get → list → delete → verify-deleted flow
  - Duplicate merchant conflict handling (409)
  - Invalid input validation (empty name, malformed JSON)
  - Reusable `TestableHandler` with mock repository
  - File: `tests/integration/merchant_test.go`

### Changed
- Updated `go.mod` with new dependencies: `prometheus/client_golang`, `opentelemetry`, `segmentio/kafka-go`, `google.golang.org/grpc`
- Updated CI pipeline (`ci.yml`) with SBOM and image signing jobs
- Updated ROADMAP with all items completed

---

## [0.7.0] - 2026-07-18

### Added — Sprint 7: Infrastructure
- **Kubernetes manifests** (`infrastructure/kubernetes/manifests.yaml`):
  - `Deployment` with 3 replicas, resource limits (100m-500m CPU, 128Mi-512Mi memory)
  - Liveness probe (`/health`, 30s interval) and readiness probe (`/ready`, 10s interval)
  - Environment variables from ConfigMap and Secrets
  - `Service` (ClusterIP, port 80 → 8080)
  - `ConfigMap` with database and Redis connection details
  - `Namespace` with app label
  - `Ingress` with nginx rate limiting, SSL redirect, and TLS termination

- **Terraform IaC** (`infrastructure/terraform/main.tf`):
  - AWS VPC via `terraform-aws-modules/vpc` — 3 AZs, private/public subnets, NAT gateway
  - RDS PostgreSQL 16 — encrypted storage, multi-AZ in production, 7-day backup retention
  - ElastiCache Redis — single node, port 6379
  - EKS cluster via `terraform-aws-modules/eks` — Kubernetes 1.30, managed node groups
  - Security groups isolating DB and Redis to EKS-only access
  - S3 backend for state management
  - Parameterized via variables (region, instance types, node counts)

- **Prometheus configuration** (`infrastructure/prometheus/prometheus.yml`):
  - Scrape configs for PCP API (10s), PostgreSQL exporter, Redis exporter
  - 15s global evaluation interval

- **Docker Compose** expanded to 6 services:
  - Added Prometheus (port 9090) and Grafana (port 3000)
  - Grafana with admin password and sign-up disabled

---

## [0.6.0] - 2026-07-18

### Added — Sprint 6: Observability Foundation
- Health probe (`GET /health`) returning service name and status
- Readiness probe (`GET /ready`) with live PostgreSQL connectivity check
- Prometheus and Grafana service definitions in Docker Compose

---

## [0.5.0] - 2026-07-18

### Added — Sprint 5: Reconciliation, Analytics & Audit

- **Reconciliation domain** (`internal/domain/reconciliation/reconciliation.go`):
  - `Record` entity comparing internal vs external payment state
  - Automatic discrepancy detection: `amount_mismatch`, `status_mismatch`
  - `Reconcile()` method producing `Record` with `IsMatched` flag
  - `ListUnmatched()` for surfacing discrepancies
  - PostgreSQL adapter with partial index on `is_matched = false`

- **Analytics service** (`internal/application/analytics/service.go`):
  - `Summary` — total/completed/failed payments, total amount, success rate
  - `ProviderStats` — per-provider charges, success/failure counts, success rate, avg latency
  - SQL queries with conditional aggregation (`CASE WHEN status='completed'`)
  - API endpoints: `GET /api/v1/analytics/summary?days=30`, `GET /api/v1/analytics/providers?days=30`

- **Audit log domain** (`internal/domain/audit/audit.go`):
  - Immutable `AuditLog` entries with entity type, entity ID, action, actor, changes (JSONB), IP, user agent
  - Query by entity (`ListByEntity`) and by actor (`ListByActor`) with pagination
  - PostgreSQL adapter with composite indexes on `(entity_type, entity_id)` and `actor_id`

- **Database migrations** (`000004`):
  - `webhooks` table with status-based index for retry scheduling
  - `audit_logs` table with entity, actor, and creation time indexes
  - `reconciliation_records` table with partial index for unmatched records

---

## [0.4.0] - 2026-07-18

### Added — Sprint 4: Resilience & Events

- **Retry engine** (`internal/application/retry/engine.go`):
  - Exponential backoff with configurable `BackoffFactor` (default 2.0x)
  - Jitter via `JitterFraction` (±20% by default) to prevent thundering herd
  - `MaxDelay` cap to prevent unbounded waits (default 30s)
  - Context-aware: respects cancellation and deadline via `select` on `ctx.Done()`
  - `DefaultStrategy()` — 3 retries, 500ms initial, 30s max, 2x backoff, 20% jitter
  - 4 tests: immediate success, eventual success, all-fail, context cancellation

- **Webhook domain** (`internal/domain/webhook/webhook.go`):
  - `Webhook` entity with delivery tracking: attempts, max retries, next retry time
  - Status lifecycle: `pending` → `delivered` / `failed`
  - `MarkDelivered()` and `MarkFailed()` with automatic next-retry scheduling
  - PostgreSQL adapter with `ListPending()` using `status='pending' AND next_retry <= NOW()`

- **Domain events** (`internal/domain/event/event.go`):
  - 9 event types: `payment.created/completed/failed/refunded`, `merchant.created/updated`, `provider.health_changed`, `webhook.delivered/failed`
  - `Event` struct with UUID, type, source, data (map), and UTC timestamp

- **Messaging abstraction** (`internal/infrastructure/messaging/interface.go`):
  - `Publisher` port: `Publish(ctx, topic, event)` + `Close()`
  - `Subscriber` port: `Subscribe(ctx, topic, handler)` + `Close()`
  - `InMemoryPublisher` for local development and testing (stores events in slice)

---

## [0.3.0] - 2026-07-18

### Added — Sprint 3: Payment Processing & Routing

- **Payment domain** (`internal/domain/payment/payment.go`):
  - `Payment` aggregate with 6 statuses: `pending`, `processing`, `completed`, `failed`, `refunded`, `cancelled`
  - `Validate()` — enforces positive amount, 3-letter ISO currency, non-nil merchant ID
  - Status transition methods: `MarkProcessing(providerID)`, `MarkCompleted(externalID)`, `MarkFailed(reason)`, `MarkRefunded()`
  - `AttemptCount` tracking for retry visibility
  - Idempotency via `IdempotencyKey` field with unique index per merchant

- **Payment service** (`internal/application/payment/service.go`):
  - `Create()` — validates input → checks idempotency → routes to provider → charges → persists
  - `Refund()` — retrieves payment → calls provider refund → updates status
  - Gateway registry pattern: `RegisterGateway(providerType, gateway)`
  - 7 tests covering validation, status transitions, missing merchant

- **Routing engine** (`internal/application/routing/engine.go`):
  - Two strategies: `priority` (lowest priority value wins) and `weighted_random` (probabilistic selection)
  - Rule-based filtering: currency match, amount range (min/max), active flag
  - `Rule.Matches(currency, amount)` — evaluates all filter criteria
  - 7 tests covering active/inactive rules, currency/amount filtering, wildcard matching

- **Payment API endpoints**:
  - `POST /api/v1/payments` — create and process a payment (authenticated)
  - `GET /api/v1/payments` — list merchant's payments with pagination
  - `GET /api/v1/payments/{id}` — get payment details
  - `POST /api/v1/payments/{id}/refund` — refund a completed payment

- **Database migrations** (`000003`):
  - `payments` table with indexes on merchant, status, idempotency key, created_at
  - `routing_rules` table with merchant index

---

## [0.2.0] - 2026-07-18

### Added — Sprint 2: Authentication & Provider Integration

- **JWT authentication** (`internal/infrastructure/auth/jwt.go`):
  - HS256 signing with configurable secret, issuer, and expiration
  - `GenerateToken(merchantID)` producing signed JWT with `sub`, `iss`, `exp`, `iat` claims
  - `ValidateToken(token)` returning `Claims` struct with merchant ID

- **Auth middleware** (`internal/interfaces/http/middleware/auth.go`):
  - Dual authentication: `Authorization: Bearer <jwt>` or `X-API-Key: pcp_<key>`
  - JWT path: validates token → extracts merchant ID → loads merchant from DB → injects into context
  - API Key path: looks up merchant by key → injects into context
  - `MerchantFromContext(ctx)` helper for downstream handlers
  - Returns `401 Unauthorized` with JSON error body on failure

- **Rate limiter** (`internal/interfaces/http/middleware/ratelimit.go`):
  - Per-key sliding window implementation using `sync.RWMutex`
  - Configurable rate (requests) and window (duration) — default 100/minute
  - Key extraction from `X-API-Key` header, falling back to `RemoteAddr`
  - Background cleanup goroutine for expired entries (runs every window interval)
  - Returns `429 Too Many Requests` with `Retry-After` header
  - 3 tests: within limit, over limit, separate keys

- **Idempotency middleware** (`internal/interfaces/http/middleware/idempotency.go`):
  - Caches response body + status code by `Idempotency-Key` header
  - TTL-based expiration (default 24 hours)
  - On duplicate key: returns cached response without executing handler
  - Background cleanup goroutine for expired entries

- **Provider domain** (`internal/domain/provider/provider.go`):
  - `Provider` entity with 4 types: `stripe`, `paypal`, `adyen`, `razorpay`
  - 3 statuses: `active`, `inactive`, `degraded`
  - `Gateway` port interface: `Charge(ctx, amount, currency, metadata)`, `Refund(ctx, chargeID)`, `GetStatus(ctx, chargeID)`
  - `GatewayResponse` value object with success flag, transaction ID, error message, raw response
  - Validation: name required, type must be recognized

- **Stripe connector** (`internal/infrastructure/connector/stripe.go`):
  - Implements `Gateway` port for Stripe payments
  - Simulated charge with `ch_` prefixed transaction IDs
  - Configurable via `api_key` in config map

- **PayPal connector** (`internal/infrastructure/connector/paypal.go`):
  - Implements `Gateway` port for PayPal payments
  - Simulated charge with `PAYID-` prefixed transaction IDs
  - Configurable via `client_id` and `client_secret`

- **Gateway factory** (`internal/infrastructure/connector/factory.go`):
  - `NewGateway(providerType, config)` — returns typed connector
  - Central registry for adding new provider types

- **Provider CRUD API**:
  - `POST /api/v1/providers` — register a provider
  - `GET /api/v1/providers` — list providers (paginated)
  - `GET /api/v1/providers/{id}` — get provider details
  - `DELETE /api/v1/providers/{id}` — remove a provider

- **Redis client** (`internal/infrastructure/cache/redis.go`):
  - Wrapper around `go-redis/v9` with `Get`, `Set`, `Delete`, `Expire` operations
  - Connection validation via `Ping()`

- **Database migrations** (`000002`):
  - `providers` table with type/status check constraints, type and status indexes

---

## [0.1.0] - 2026-07-18

### Added — Sprint 1: Foundation

- **Project scaffold** following Clean Architecture / Hexagonal / DDD:
  - 4 layers: Domain → Application → Infrastructure → Interfaces
  - Strict dependency rule: inner layers never depend on outer layers
  - Every domain concept defines interfaces (ports); infrastructure provides implementations (adapters)

- **Merchant bounded context**:
  - `Merchant` aggregate root with ID (UUID), name, API key, webhook URL, status, timestamps
  - Status value object: `active`, `suspended`, `inactive` with `IsValid()` validation
  - `Validate()` enforcing name required (max 255 chars) and valid status
  - Domain errors: `ErrMerchantNotFound`, `ErrDuplicateMerchant`, `ErrInvalidName`, `ErrInvalidStatus`
  - `Repository` port with 6 methods: Create, GetByID, GetByAPIKey, List, Update, Delete
  - Application service with API key generation (`pcp_` prefix + 32 random hex bytes)
  - 8 domain tests, 11 service tests

- **PostgreSQL persistence** (`internal/infrastructure/persistence/`):
  - `pgxpool`-based connection pool (25 max conns, 5 min conns, 30min lifetime)
  - `PostgresMerchantRepository` with parameterized queries and proper error mapping
  - PgError code `23505` mapped to `ErrDuplicateMerchant`
  - `pgx.ErrNoRows` mapped to `ErrMerchantNotFound`

- **HTTP API** with Chi v5 router:
  - `POST /api/v1/merchants` — create merchant (returns generated API key)
  - `GET /api/v1/merchants` — list with offset/limit pagination
  - `GET /api/v1/merchants/{id}` — get by UUID
  - `PUT /api/v1/merchants/{id}` — update name, webhook URL, status
  - `DELETE /api/v1/merchants/{id}` — soft delete (204 No Content)
  - `GET /health` — health check
  - `GET /ready` — readiness probe with DB ping

- **Middleware stack** (applied in order):
  1. `RequestID` — generates/propagates `X-Request-ID` header
  2. `Recovery` — catches panics, logs stack trace, returns 500
  3. `Logging` — structured request/response logging via Zap
  4. `RealIP` — extracts client IP from proxy headers
  5. `Compress` — gzip compression at level 5

- **Configuration** via Viper:
  - Environment variable prefix: `PCP_`
  - Sections: Server, Database, Redis, JWT, Log
  - All settings have sensible defaults for local development

- **Docker infrastructure**:
  - Multi-stage `Dockerfile`: Go 1.23 Alpine builder → Alpine 3.20 runtime
  - Non-root `pcp` user in runtime container
  - `docker-compose.yml` with PostgreSQL 16, Redis 7, golang-migrate, API service
  - Health checks on all services, ordered startup via `depends_on` conditions

- **CI/CD pipeline** (`.github/workflows/ci.yml`):
  - Jobs: Lint (golangci-lint + go vet) → Test (with PostgreSQL service) → Build → Docker → Security
  - Coverage threshold enforcement (≥60%)
  - Trivy filesystem vulnerability scanner (HIGH + CRITICAL)
  - Nancy dependency vulnerability scanner

- **Documentation suite** (12 files):
  - README.md, ARCHITECTURE.md, CONTRIBUTING.md, SECURITY.md, CHANGELOG.md, TODO.md
  - docs/: API.md, LOCAL_DEVELOPMENT.md, DEPLOYMENT.md, TESTING.md, DECISIONS.md, ROADMAP.md
