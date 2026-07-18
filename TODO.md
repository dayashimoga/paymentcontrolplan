# TODO — Payment Control Plane (PCP)

## ✅ All Sprints Complete

### Sprint 1: Foundation ✅
- [x] Project scaffold (Clean/Hexagonal/DDD)
- [x] Merchant domain (entity, repository, service, handler)
- [x] PostgreSQL persistence adapters
- [x] HTTP API with Chi router
- [x] Middleware (logging, recovery, request ID)
- [x] Docker (multi-stage Dockerfile, docker-compose)
- [x] CI/CD pipeline (GitHub Actions)
- [x] Database migrations (golang-migrate)
- [x] Comprehensive documentation (12 files)
- [x] Unit tests for domain + application + handler layers

### Sprint 2: Authentication & Provider Integration ✅
- [x] JWT authentication middleware
- [x] API Key authentication middleware
- [x] Rate limiting (per-key sliding window)
- [x] Idempotency middleware
- [x] Provider domain + Gateway port
- [x] Stripe connector
- [x] PayPal connector
- [x] Provider CRUD API
- [x] Redis client wrapper
- [x] OpenAPI 3.0 specification (17 endpoints)
- [x] Integration tests (httptest + mock repo)

### Sprint 3: Payment Processing & Routing ✅
- [x] Payment domain (entity, status transitions)
- [x] Payment service (create, charge, refund)
- [x] Routing engine (priority, weighted random, currency/amount filters)
- [x] Payment API endpoints (create, get, list, refund)
- [x] gRPC protobuf definitions + server scaffold

### Sprint 4: Resilience & Events ✅
- [x] Retry engine (exponential backoff, jitter)
- [x] Circuit breaker (closed/open/half-open)
- [x] Webhook domain (delivery tracking)
- [x] Domain event types
- [x] Messaging abstraction (Publisher/Subscriber ports)
- [x] Kafka producer/consumer implementation
- [x] In-memory publisher for local dev

### Sprint 5: Reconciliation, Analytics & Audit ✅
- [x] Reconciliation domain
- [x] Analytics service + API endpoints
- [x] Audit log domain
- [x] PostgreSQL adapters for all domains

### Sprint 6: Observability ✅
- [x] Prometheus metrics + /metrics endpoint + middleware
- [x] OpenTelemetry distributed tracing
- [x] Grafana dashboard (9 panels)
- [x] Health and readiness probes

### Sprint 7: Infrastructure ✅
- [x] Kubernetes manifests
- [x] Helm charts (with HPA autoscaling)
- [x] Terraform IaC (VPC, RDS, EKS, ElastiCache)
- [x] Cloudflare WAF/CDN configuration
- [x] Image signing (Cosign) + SBOM (Syft) in CI
- [x] Docker Compose (6 services)

## 🔮 Future Enhancements
- [ ] Adyen connector implementation
- [ ] Razorpay connector implementation
- [ ] OAuth2 provider
- [ ] Admin dashboard UI
- [ ] SDK generation (Go, Python, JS)
- [ ] Load testing with k6
- [ ] Multi-tenancy support
