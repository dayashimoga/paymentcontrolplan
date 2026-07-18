# Deployment

## Overview

PCP supports multiple deployment targets:

- **Local Development**: Docker Compose (available now)
- **Kubernetes**: Manifests and Helm charts (Sprint 7)
- **Terraform**: Infrastructure as Code (Sprint 7)
- **Cloudflare**: Edge deployment (Sprint 7)

## Docker Image

### Build

```bash
cd backend
docker build -t pcp-api:latest .
```

### Run

```bash
docker run -p 8080:8080 \
  -e PCP_DATABASE_HOST=your-db-host \
  -e PCP_DATABASE_PASSWORD=your-secret \
  -e PCP_JWT_SECRET=your-jwt-secret \
  pcp-api:latest
```

## Environment Configuration

All production settings are controlled via environment variables. See [LOCAL_DEVELOPMENT.md](LOCAL_DEVELOPMENT.md) for the full variable reference.

**Critical production settings**:
- `PCP_JWT_SECRET` — Use a strong, unique secret
- `PCP_DATABASE_PASSWORD` — Use managed secrets
- `PCP_DATABASE_SSL_MODE` — Set to `require` in production
- `PCP_LOG_FORMAT` — Use `json` for structured log aggregation

## Database Migrations

Migrations run automatically in the Docker Compose setup. For production:

```bash
docker run --rm \
  -v ./backend/migrations:/migrations \
  migrate/migrate:v4.17.0 \
  -path=/migrations \
  -database "postgres://user:pass@host:5432/pcp?sslmode=require" \
  up
```

## Health Checks

- **Liveness**: `GET /health` — Is the process running?
- **Readiness**: `GET /ready` — Is the service ready to handle traffic?

Configure your load balancer/orchestrator to use these endpoints.

## Coming in Future Sprints

- Kubernetes Deployment, Service, ConfigMap, Secret manifests
- Helm chart with configurable values
- Terraform modules for AWS/GCP
- Cloudflare Workers deployment
- CI/CD auto-deploy on merge to main
