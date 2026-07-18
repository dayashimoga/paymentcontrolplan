# Payment Control Plane (PCP)

A production-grade payment orchestration platform that sits between merchants and payment providers, providing one unified API with routing, retries, reconciliation, analytics, and provider abstraction.

> **This is NOT a payment gateway.** PCP never processes or stores raw card information.

## Architecture

```
┌─────────────┐     ┌─────────────────────────────────────┐     ┌──────────────┐
│  Merchant   │────▶│      Payment Control Plane (PCP)     │────▶│   Stripe     │
│  App / API  │◀────│                                     │◀────│   PayPal     │
└─────────────┘     │  ┌─────────┐  ┌────────┐  ┌──────┐ │     │   Adyen      │
                    │  │ Routing │  │ Retry  │  │ Auth │ │     │   Razorpay   │
                    │  │ Engine  │  │ Engine │  │      │ │     └──────────────┘
                    │  └─────────┘  └────────┘  └──────┘ │
                    │  ┌─────────┐  ┌────────┐  ┌──────┐ │
                    │  │Analytics│  │Webhook │  │Audit │ │
                    │  │         │  │ Engine │  │ Log  │ │
                    │  └─────────┘  └────────┘  └──────┘ │
                    └─────────────────────────────────────┘
```

## Quick Start

```bash
# Clone and start
git clone <repo> && cd paymentbridge
docker compose -f docker/docker-compose.yml up --build -d

# Verify
curl http://localhost:8080/health

# Create a merchant
curl -X POST http://localhost:8080/api/v1/merchants \
  -H "Content-Type: application/json" \
  -d '{"name":"My Store","webhook_url":"https://mystore.com/webhook"}'

# Use the returned api_key for authenticated requests
curl -X POST http://localhost:8080/api/v1/payments \
  -H "Content-Type: application/json" \
  -H "X-API-Key: pcp_<your_key>" \
  -d '{"amount":5000,"currency":"USD","description":"Order #123"}'
```

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Health check |
| GET | `/ready` | No | Readiness probe |
| POST | `/api/v1/merchants` | No | Create merchant |
| GET | `/api/v1/merchants` | Yes | List merchants |
| GET | `/api/v1/merchants/{id}` | Yes | Get merchant |
| PUT | `/api/v1/merchants/{id}` | Yes | Update merchant |
| DELETE | `/api/v1/merchants/{id}` | Yes | Delete merchant |
| POST | `/api/v1/providers` | Yes | Register provider |
| GET | `/api/v1/providers` | Yes | List providers |
| GET | `/api/v1/providers/{id}` | Yes | Get provider |
| DELETE | `/api/v1/providers/{id}` | Yes | Delete provider |
| POST | `/api/v1/payments` | Yes | Create payment |
| GET | `/api/v1/payments` | Yes | List payments |
| GET | `/api/v1/payments/{id}` | Yes | Get payment |
| POST | `/api/v1/payments/{id}/refund` | Yes | Refund payment |
| GET | `/api/v1/analytics/summary` | Yes | Payment analytics |
| GET | `/api/v1/analytics/providers` | Yes | Provider stats |

## Authentication

PCP supports two authentication methods:

1. **API Key**: Include `X-API-Key: pcp_<key>` header
2. **JWT Bearer Token**: Include `Authorization: Bearer <token>` header

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.23 |
| Router | Chi v5 |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Auth | JWT (HS256) + API Key |
| Migrations | golang-migrate |
| Monitoring | Prometheus + Grafana |
| Container | Docker + Docker Compose |
| IaC | Terraform + Kubernetes |
| CI/CD | GitHub Actions |

## Project Structure

```
backend/
├── cmd/api/              # Application entrypoint
├── internal/
│   ├── domain/           # Business entities, ports, rules
│   │   ├── merchant/     # Merchant aggregate
│   │   ├── provider/     # Provider + Gateway port
│   │   ├── payment/      # Payment aggregate
│   │   ├── routing/      # Routing rules
│   │   ├── webhook/      # Webhook delivery
│   │   ├── audit/        # Audit log
│   │   ├── event/        # Domain events
│   │   ├── reconciliation/ # Transaction matching
│   │   ├── auth/         # Auth types + ports
│   │   └── common/       # Shared errors
│   ├── application/      # Use cases / services
│   ├── infrastructure/   # Adapters (Postgres, Redis, JWT, connectors)
│   └── interfaces/       # HTTP handlers, middleware, DTOs, router
├── migrations/           # SQL migration files
└── Dockerfile
infrastructure/
├── kubernetes/           # K8s manifests
├── terraform/            # AWS IaC
└── prometheus/           # Monitoring config
```

## Documentation

| Document | Description |
|----------|-------------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | System design and patterns |
| [CHANGELOG.md](CHANGELOG.md) | Version history |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contribution guidelines |
| [SECURITY.md](SECURITY.md) | Security policies |
| [docs/API.md](docs/API.md) | API reference |
| [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) | Deployment guide |
| [docs/LOCAL_DEVELOPMENT.md](docs/LOCAL_DEVELOPMENT.md) | Local setup |
| [docs/TESTING.md](docs/TESTING.md) | Testing strategy |

## License

Proprietary. All rights reserved.
