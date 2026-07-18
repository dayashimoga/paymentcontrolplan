# Roadmap

## Sprint 1: Foundation ✅
- [x] Project structure (Clean/Hexagonal/DDD)
- [x] Merchant Management CRUD
- [x] PostgreSQL + migrations
- [x] Docker Compose local dev
- [x] GitHub Actions CI
- [x] Documentation suite (12 files)

## Sprint 2: Auth, Providers, Swagger ✅
- [x] JWT authentication
- [x] API key authentication
- [x] Rate limiting (in-memory, Redis-ready)
- [x] OpenAPI 3.0 specification
- [x] Provider abstraction layer (Gateway port)
- [x] Stripe connector
- [x] PayPal connector
- [x] Integration tests (httptest + mock repo)

## Sprint 3: Payments & Routing ✅
- [x] Payment domain and API (create, get, list, refund)
- [x] gRPC protobuf definitions + server scaffold
- [x] Routing engine (priority + weighted random)
- [x] Idempotency middleware

## Sprint 4: Resilience & Events ✅
- [x] Retry engine (exponential backoff + jitter)
- [x] Circuit breaker (closed/open/half-open state machine)
- [x] Webhook engine (domain + delivery tracking)
- [x] Kafka producer/consumer (segmentio/kafka-go)
- [x] Domain events + in-memory publisher
- [x] Messaging abstraction (Publisher/Subscriber ports)

## Sprint 5: Reconciliation & Analytics ✅
- [x] Reconciliation engine (cross-provider matching)
- [x] Analytics data API (summary + provider breakdown)
- [x] Audit log system (immutable entries)
- [x] Reporting endpoints

## Sprint 6: Observability ✅
- [x] OpenTelemetry distributed tracing
- [x] Prometheus metrics (/metrics endpoint + middleware)
- [x] Grafana dashboard (JSON provisioning)
- [x] Tracing middleware (span injection)

## Sprint 7: Infrastructure ✅
- [x] Kubernetes manifests (Deployment, Service, Ingress)
- [x] Helm charts (Chart.yaml, values.yaml, templates with HPA)
- [x] Terraform modules (VPC, RDS, ElastiCache, EKS)
- [x] Cloudflare WAF/CDN configuration
- [x] Image signing (Cosign) + SBOM (Syft) in CI

## Future Enhancements
- [ ] Adyen connector
- [ ] Razorpay connector
- [ ] OAuth2 provider
- [ ] Admin dashboard UI
- [ ] SDK generation (Go, Python, JS)
- [ ] Load testing (k6)
- [ ] Multi-tenancy
