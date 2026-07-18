# Architecture

## Overview

Payment Control Plane (PCP) is a **neutral payment orchestration platform** that sits between merchants and payment providers (Stripe, PayPal, Adyen, Razorpay), providing one unified API while handling routing, retries, reconciliation, analytics, and provider abstraction.

**This is NOT a payment gateway.** PCP never processes or stores raw card information. It orchestrates tokenized payment operations through provider APIs.

The system is designed using three complementary architectural patterns:

| Pattern | Purpose |
|---------|---------|
| **Clean Architecture** | Layer isolation with strict dependency rule (inner layers never depend on outer) |
| **Hexagonal Architecture** | Port/adapter separation — domain defines interfaces, infrastructure implements them |
| **Domain-Driven Design** | Bounded contexts, aggregate roots, value objects, ubiquitous language |

---

## System Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│                            Merchant Applications                         │
│                    (E-commerce, Mobile, SaaS Platforms)                   │
└────────────────────────────────┬─────────────────────────────────────────┘
                                 │  REST API / gRPC
                                 ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                        PAYMENT CONTROL PLANE (PCP)                       │
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────────┐ │
│  │                      Interfaces Layer                               │ │
│  │  ┌──────────┐ ┌──────────┐ ┌───────────────────────────────────┐  │ │
│  │  │   REST   │ │   gRPC   │ │        Middleware Stack            │  │ │
│  │  │ Handlers │ │  Server  │ │ RequestID → Recovery → Logging    │  │ │
│  │  │ (Chi v5) │ │ (grpc-go)│ │ → Auth → RateLimit → Idempotency │  │ │
│  │  └────┬─────┘ └────┬─────┘ │ → Metrics → Tracing → Compress   │  │ │
│  │       └─────────────┘       └───────────────────────────────────┘  │ │
│  ├─────────────────────────────────────────────────────────────────────┤ │
│  │                     Application Layer                               │ │
│  │  ┌────────────┐ ┌──────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐│ │
│  │  │  Payment   │ │ Routing  │ │  Retry  │ │Analytics│ │  Audit  ││ │
│  │  │  Service   │ │  Engine  │ │  Engine │ │ Service │ │ Service ││ │
│  │  └────────────┘ └──────────┘ └─────────┘ └─────────┘ └─────────┘│ │
│  ├─────────────────────────────────────────────────────────────────────┤ │
│  │                       Domain Layer (Core)                           │ │
│  │  ┌──────────┐ ┌──────────┐ ┌─────────┐ ┌─────────┐ ┌───────────┐│ │
│  │  │ Merchant │ │ Payment  │ │Provider │ │ Routing │ │  Webhook  ││ │
│  │  │Aggregate │ │Aggregate │ │+Gateway │ │  Rules  │ │  Domain   ││ │
│  │  └──────────┘ └──────────┘ └─────────┘ └─────────┘ └───────────┘│ │
│  │  ┌──────────┐ ┌──────────┐ ┌─────────┐ ┌─────────────────────┐  │ │
│  │  │  Audit   │ │  Event   │ │  Recon  │ │  Common (Errors,    │  │ │
│  │  │   Log    │ │  Types   │ │ Records │ │   Auth Ports)       │  │ │
│  │  └──────────┘ └──────────┘ └─────────┘ └─────────────────────┘  │ │
│  ├─────────────────────────────────────────────────────────────────────┤ │
│  │                  Infrastructure Layer (Adapters)                    │ │
│  │  ┌──────────┐ ┌──────┐ ┌───────┐ ┌─────────┐ ┌────────────────┐ │ │
│  │  │PostgreSQL│ │Redis │ │ Kafka │ │   JWT   │ │   Connectors   │ │ │
│  │  │  Repos   │ │Cache │ │Pub/Sub│ │ Service │ │ Stripe, PayPal │ │ │
│  │  └──────────┘ └──────┘ └───────┘ └─────────┘ └────────────────┘ │ │
│  │  ┌──────────────┐ ┌────────────┐ ┌──────────────────────┐       │ │
│  │  │Circuit Breaker│ │Prometheus  │ │   OpenTelemetry      │       │ │
│  │  │              │ │  Metrics   │ │     Tracing          │       │ │
│  │  └──────────────┘ └────────────┘ └──────────────────────┘       │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼  Provider Gateway Port
┌──────────────────────────────────────────────────────────────────────────┐
│                         Payment Providers                                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐               │
│  │  Stripe  │  │  PayPal  │  │  Adyen   │  │ Razorpay │               │
│  │ (Active) │  │ (Active) │  │ (Planned)│  │ (Planned)│               │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘               │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## Design Principles

### 1. Dependency Rule

Dependencies flow strictly inward. The domain layer has **zero** external dependencies — no frameworks, no database drivers, no HTTP libraries.

```
Interfaces → Application → Domain ← Infrastructure
     ↓            ↓           ↑           ↑
  (depends on) (depends on) (defines)  (implements)
```

### 2. Port & Adapter Pattern

Every boundary crossing uses an interface (port) defined by the consuming layer:

```
[Driving Adapter]  →  [Port (Interface)]  →  [Domain Logic]  →  [Port (Interface)]  →  [Driven Adapter]
   HTTP Handler         Service Input           Business Rules    Repository Interface    PostgreSQL
   gRPC Server                                                    Gateway Interface       Stripe API
   CLI Command                                                    Publisher Interface      Kafka
```

**Driving adapters** (left side) invoke the domain — they are owned by the interfaces layer.
**Driven adapters** (right side) are invoked by the domain — they implement domain-defined ports.

### 3. Aggregate Root Pattern

Each bounded context has an aggregate root that enforces invariants:

| Context | Aggregate Root | Key Invariants |
|---------|---------------|----------------|
| Merchant | `Merchant` | Name required, max 255 chars; status must be valid enum |
| Payment | `Payment` | Amount > 0; currency is 3-letter ISO; merchant ID required |
| Provider | `Provider` | Name unique; type must be recognized; gateway must implement port |
| Routing | `Rule` | Must reference valid provider and merchant; priority/weight non-negative |

---

## Bounded Contexts

### Merchant Context
**Responsibility**: Merchant lifecycle management and API key provisioning.

```
domain/merchant/
├── merchant.go    — Aggregate root (entity + validation + status transitions)
├── errors.go      — Domain errors (ErrMerchantNotFound, ErrDuplicateMerchant, etc.)
└── merchant_test.go — 8 unit tests

application/merchant/
├── service.go     — Use cases: Create, Get, List, Update, Delete, GetByAPIKey
└── service_test.go — 11 unit tests (with mock repository)
```

### Payment Context
**Responsibility**: Payment orchestration — routing, charging, tracking, and refunding.

```
domain/payment/
├── payment.go     — Aggregate root with 6-state lifecycle
├── errors.go      — ErrPaymentNotFound, ErrInvalidAmount, ErrInvalidCurrency
└── payment_test.go — 7 unit tests (validation + all status transitions)

application/payment/
└── service.go     — Orchestrates: validate → check idempotency → route → charge → persist
```

### Provider Context
**Responsibility**: Provider abstraction with the Gateway port pattern.

```
domain/provider/
├── provider.go    — Entity + Type enum + Status enum + Gateway port interface
├── errors.go      — ErrProviderNotFound, ErrDuplicateProvider
└── provider_test.go — 4 unit tests

application/provider/
└── service.go     — CRUD + gateway registry (RegisterGateway, GetGateway)

infrastructure/connector/
├── stripe.go      — Gateway implementation for Stripe
├── paypal.go      — Gateway implementation for PayPal
└── factory.go     — NewGateway(type, config) factory
```

### Routing Context
**Responsibility**: Intelligent payment routing based on configurable rules.

```
domain/routing/
├── rule.go        — Rule entity with priority, weight, currency/amount filters
├── rule_test.go   — 7 unit tests for rule matching logic

application/routing/
└── engine.go      — Route(merchantID, currency, amount) → selected provider
                     Strategies: priority-based, weighted random
```

### Webhook Context
**Responsibility**: Asynchronous webhook delivery with retry scheduling.

```
domain/webhook/
└── webhook.go     — Entity with attempts tracking, max retries, next retry scheduling
```

### Audit Context
**Responsibility**: Immutable audit trail for compliance and debugging.

```
domain/audit/
└── audit.go       — AuditLog entity + Service (Log, ListByEntity, ListByActor)
```

### Reconciliation Context
**Responsibility**: Cross-provider transaction matching and discrepancy detection.

```
domain/reconciliation/
└── reconciliation.go — Record entity + Service (Reconcile, ListUnmatched)
```

### Event Context
**Responsibility**: Domain event definitions for event-driven architecture.

```
domain/event/
└── event.go       — 9 event types (payment.*, merchant.*, provider.*, webhook.*)
```

---

## Data Flow

### Payment Processing Flow

```
1. HTTP POST /api/v1/payments
   ├── Middleware: RequestID → Recovery → Logging → Auth → RateLimit → Idempotency
   │
2. PaymentHandler.Create()
   ├── Deserialize JSON → CreatePaymentRequest DTO
   ├── Extract merchant from context (set by Auth middleware)
   │
3. PaymentService.Create()
   ├── Check idempotency (same merchant + idempotency_key → return cached)
   ├── Create Payment entity (status: pending)
   ├── Persist to PostgreSQL
   │
4. RoutingEngine.Route(merchantID, currency, amount)
   ├── Load routing rules for merchant
   ├── Filter by currency, amount range, active flag
   ├── Select provider (priority-based or weighted random)
   │
5. Payment.MarkProcessing(providerID)
   ├── Status: pending → processing
   ├── Increment AttemptCount
   │
6. Gateway.Charge(ctx, amount, currency, metadata)
   ├── Call provider API (Stripe/PayPal)
   ├── Return GatewayResponse (success, transactionID, error)
   │   ├── On success → Payment.MarkCompleted(externalID)
   │   └── On failure → RetryEngine.Execute() or Payment.MarkFailed(reason)
   │
7. Persist final state to PostgreSQL
8. Return PaymentResponse DTO as JSON
```

### Authentication Flow

```
Request arrives with either:
  A) Authorization: Bearer <jwt_token>
     → JWTService.ValidateToken() → extract merchantID → repo.GetByID() → inject merchant
  
  B) X-API-Key: pcp_<hex_key>
     → repo.GetByAPIKey() → inject merchant into context

Neither present → 401 Unauthorized
```

---

## Infrastructure Patterns

### Resilience

| Pattern | Implementation | Purpose |
|---------|---------------|---------|
| **Retry** | Exponential backoff + jitter | Transient failure recovery |
| **Circuit Breaker** | 3-state machine (closed/open/half-open) | Cascading failure prevention |
| **Idempotency** | Key-based response caching (24h TTL) | Duplicate request safety |
| **Rate Limiting** | Per-key sliding window | Abuse prevention |

### Circuit Breaker State Machine

```
                    failure_count >= threshold
         ┌──────────────────────────────────────┐
         │                                      ▼
    ┌─────────┐                           ┌──────────┐
    │ CLOSED  │                           │   OPEN   │
    │ (normal)│                           │ (blocked)│
    └─────────┘                           └────┬─────┘
         ▲                                     │
         │  success_count >= threshold         │ timeout expired
         │                                     ▼
         │                              ┌────────────┐
         └──────────────────────────────│ HALF-OPEN  │
                                        │  (testing) │
                    failure ───────────▶└────────────┘
                                              │
                                              └──▶ OPEN (on failure)
```

### Observability Stack

```
┌─────────┐     scrape /metrics      ┌────────────┐     query      ┌─────────┐
│ PCP API │◄─────────────────────────│ Prometheus │◄──────────────│ Grafana │
│         │                          │            │               │(9 panels)│
│ Metrics:│                          └────────────┘               └─────────┘
│ • HTTP  │
│ • Pay   │     traces (OTLP)        ┌────────────┐
│ • Prov  │─────────────────────────▶│   Jaeger/  │
│ • CB    │                          │   Tempo    │
└─────────┘                          └────────────┘
```

**Metrics collected:**
- `pcp_http_requests_total` — request count by method/path/status
- `pcp_http_request_duration_seconds` — latency histogram
- `pcp_payments_total` — payment count by status/provider/currency
- `pcp_payment_amount` — amount distribution histogram
- `pcp_provider_latency_seconds` — provider API latency
- `pcp_circuit_breaker_state` — circuit breaker state gauge
- `pcp_active_connections` — current active connection count

---

## Directory Structure

```
paymentbridge/
├── backend/
│   ├── cmd/api/main.go                    # Composition root (wires all dependencies)
│   ├── api/proto/v1/pcp.proto             # gRPC protobuf definitions
│   ├── internal/
│   │   ├── domain/                        # === DOMAIN LAYER ===
│   │   │   ├── merchant/                  # Merchant aggregate (entity, errors, port)
│   │   │   ├── payment/                   # Payment aggregate (6-state lifecycle)
│   │   │   ├── provider/                  # Provider entity + Gateway port
│   │   │   ├── routing/                   # Routing rules entity
│   │   │   ├── webhook/                   # Webhook delivery tracking
│   │   │   ├── audit/                     # Immutable audit log
│   │   │   ├── event/                     # Domain event types
│   │   │   ├── reconciliation/            # Transaction matching
│   │   │   ├── auth/                      # Auth token port
│   │   │   └── common/                    # Shared errors
│   │   │
│   │   ├── application/                   # === APPLICATION LAYER ===
│   │   │   ├── merchant/service.go        # Merchant use cases
│   │   │   ├── payment/service.go         # Payment orchestration
│   │   │   ├── provider/service.go        # Provider CRUD + gateway registry
│   │   │   ├── routing/engine.go          # Route selection engine
│   │   │   ├── retry/engine.go            # Retry with backoff + jitter
│   │   │   └── analytics/service.go       # Analytics queries
│   │   │
│   │   ├── infrastructure/                # === INFRASTRUCTURE LAYER ===
│   │   │   ├── persistence/               # PostgreSQL repositories (6 files)
│   │   │   ├── auth/jwt.go                # JWT token service (HS256)
│   │   │   ├── cache/redis.go             # Redis client wrapper
│   │   │   ├── config/config.go           # Viper configuration
│   │   │   ├── logging/logger.go          # Zap structured logging
│   │   │   ├── connector/                 # Provider adapters (Stripe, PayPal, factory)
│   │   │   ├── messaging/                 # Kafka + in-memory publishers
│   │   │   ├── circuitbreaker/            # Circuit breaker implementation
│   │   │   └── observability/             # Prometheus metrics + OTEL tracing
│   │   │
│   │   └── interfaces/                    # === INTERFACES LAYER ===
│   │       ├── http/
│   │       │   ├── handler/               # REST handlers (merchant, payment, provider, analytics, health, auth)
│   │       │   ├── middleware/            # Auth, rate limit, idempotency, logging, recovery, request ID
│   │       │   ├── router/router.go       # Chi router with route groups
│   │       │   └── dto/                   # Request/response DTOs
│   │       └── grpc/server.go             # gRPC server with health + reflection
│   │
│   ├── migrations/                        # SQL migrations (4 up/down pairs)
│   ├── tests/integration/                 # Integration test suite
│   ├── Dockerfile                         # Multi-stage build
│   ├── Makefile                           # Build automation
│   ├── go.mod                             # Go module (15 direct dependencies)
│   └── go.sum                             # Dependency checksums
│
├── docker/
│   └── docker-compose.yml                 # 6 services: API, Postgres, Redis, Migrate, Prometheus, Grafana
│
├── infrastructure/
│   ├── kubernetes/manifests.yaml          # Deployment, Service, ConfigMap, Ingress, Namespace
│   ├── helm/pcp/                          # Helm chart (Chart.yaml, values.yaml, templates/)
│   ├── terraform/main.tf                  # AWS: VPC, RDS, ElastiCache, EKS
│   ├── prometheus/prometheus.yml          # Scrape configuration
│   ├── grafana/dashboards/                # Dashboard JSON provisioning
│   └── cloudflare/config.md              # WAF, SSL, rate limiting config
│
├── docs/
│   ├── openapi.yaml                       # OpenAPI 3.0 specification (17 endpoints)
│   ├── ARCHITECTURE.md                    # This document
│   ├── API.md                             # API reference
│   ├── LOCAL_DEVELOPMENT.md              # Setup guide
│   ├── DEPLOYMENT.md                      # Deployment guide
│   ├── TESTING.md                         # Testing strategy
│   ├── DECISIONS.md                       # Architecture decision records
│   └── ROADMAP.md                         # Feature roadmap
│
├── .github/workflows/ci.yml              # CI: lint → test → build → Docker → security → SBOM → sign
├── README.md                              # Project overview + quick start
├── CHANGELOG.md                           # Detailed version history
├── CONTRIBUTING.md                        # Contribution guidelines
├── SECURITY.md                            # Security policies
├── STATUS_REPORT.md                       # Live project status
└── TODO.md                                # Task tracking
```

---

## Technology Stack

| Component | Technology | Version | Rationale |
|-----------|-----------|---------|-----------|
| Language | Go | 1.23 | Performance, concurrency, strong typing, single binary |
| HTTP Router | Chi | v5.2 | Lightweight, stdlib-compatible, middleware-first |
| gRPC | grpc-go | v1.68 | High-performance internal communication |
| Database | PostgreSQL | 16 | ACID, JSONB, partial indexes, mature ecosystem |
| DB Driver | pgx | v5.7 | Fastest pure-Go Postgres driver, connection pooling |
| Cache | Redis | 7 | Rate limiting, session cache, pub/sub |
| Redis Client | go-redis | v9.7 | Full Redis API, pipelining, cluster support |
| Messaging | Kafka | via kafka-go | Event-driven architecture, at-least-once delivery |
| Auth | JWT | HS256 | Stateless authentication, standard claims |
| Config | Viper | v1.20 | Env vars, config files, 12-factor compliant |
| Logging | Zap | v1.27 | Structured, zero-allocation, high performance |
| Metrics | Prometheus | v1.20 | Industry standard, pull-based, PromQL |
| Tracing | OpenTelemetry | v1.31 | Vendor-neutral distributed tracing |
| Migrations | golang-migrate | v4.17 | Docker-friendly, SQL-based, up/down support |
| Testing | stdlib + testify | v1.10 | Assertions, mocking, table-driven tests |
| Container | Docker | Multi-stage | Minimal Alpine runtime, non-root user |
| Orchestration | Kubernetes | 1.30 | Production scaling, health management |
| IaC | Terraform | ≥1.5 | AWS infrastructure automation |
| Package Manager | Helm | v3 | Kubernetes templating and release management |
| CI/CD | GitHub Actions | — | Automated lint, test, build, security, SBOM |
| Security | Trivy + Nancy | — | Container and dependency vulnerability scanning |
| Supply Chain | Cosign + Syft | — | Image signing and SBOM generation |

---

## Deployment Architecture

### Local Development

```
docker compose -f docker/docker-compose.yml up --build -d
```

Starts 6 services with automatic migration and health checking.

### Production (AWS)

```
┌─────────────────────────────────────────────────────┐
│                    AWS Cloud                         │
│                                                     │
│  ┌─────────────┐    ┌───────────────────────────┐  │
│  │ CloudFlare  │    │        VPC (10.0.0.0/16)  │  │
│  │  CDN + WAF  │───▶│                           │  │
│  └─────────────┘    │  ┌─────────────────────┐  │  │
│                     │  │    Public Subnets    │  │  │
│                     │  │    (NAT Gateway)     │  │  │
│                     │  └──────────┬──────────┘  │  │
│                     │             │              │  │
│                     │  ┌──────────▼──────────┐  │  │
│                     │  │   Private Subnets   │  │  │
│                     │  │                     │  │  │
│                     │  │  ┌──────┐ ┌──────┐  │  │  │
│                     │  │  │ EKS  │ │ EKS  │  │  │  │
│                     │  │  │Node 1│ │Node 2│  │  │  │
│                     │  │  └──┬───┘ └──┬───┘  │  │  │
│                     │  │     │        │      │  │  │
│                     │  │  ┌──▼────────▼──┐   │  │  │
│                     │  │  │  RDS Postgres │   │  │  │
│                     │  │  │  (Multi-AZ)   │   │  │  │
│                     │  │  └──────────────┘   │  │  │
│                     │  │  ┌──────────────┐   │  │  │
│                     │  │  │  ElastiCache  │   │  │  │
│                     │  │  │    Redis     │   │  │  │
│                     │  │  └──────────────┘   │  │  │
│                     │  └─────────────────────┘  │  │
│                     └───────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

---

## Security Architecture

| Control | Implementation |
|---------|---------------|
| **Authentication** | Dual: JWT (Bearer) + API Key (X-API-Key) |
| **Authorization** | Merchant-scoped — each merchant sees only their own data |
| **Rate Limiting** | Per-key sliding window (100 req/min default) |
| **Idempotency** | Key-based deduplication with 24h TTL |
| **Transport** | TLS 1.2+ enforced via Ingress/Cloudflare |
| **Secrets** | Kubernetes Secrets for DB password, JWT secret |
| **API Keys** | Cryptographically random, `pcp_` prefixed, 256-bit entropy |
| **SQL Injection** | Parameterized queries only (pgx prepared statements) |
| **Dependency Scanning** | Trivy (containers) + Nancy (Go modules) in CI |
| **Image Integrity** | Cosign keyless signing + Syft SBOM (SPDX) |
| **No Raw Cards** | PCP never sees PAN/CVV — providers handle card data |

---

## Database Schema

```sql
merchants          — API key-authenticated merchant accounts
providers          — Payment provider configurations (Stripe, PayPal, etc.)
payments           — Payment transaction records with status lifecycle
routing_rules      — Per-merchant provider routing configuration
webhooks           — Webhook delivery tracking with retry scheduling
audit_logs         — Immutable audit trail (entity changes, actor tracking)
reconciliation_records — Cross-provider transaction matching results
```

All tables use UUID primary keys, TIMESTAMPTZ for temporal fields, and JSONB for flexible metadata.

---

## Technology Decisions

See [docs/DECISIONS.md](docs/DECISIONS.md) for Architecture Decision Records explaining each technology choice.
