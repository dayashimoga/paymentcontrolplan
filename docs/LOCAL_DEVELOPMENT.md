# Local Development

## Prerequisites

- **Docker** and **Docker Compose** (only requirement)
- No Go, PostgreSQL, or Redis installation needed locally

## Quick Start

```bash
# Start all services
cd docker
docker compose up

# The following services will start:
# - PostgreSQL 16 on port 5432
# - Redis 7 on port 6379
# - Database migrations (auto-run)
# - PCP API on port 8080
```

## Verify Setup

```bash
# Health check
curl http://localhost:8080/health
# Expected: {"status":"healthy","service":"pcp-api"}

# Readiness check (includes DB)
curl http://localhost:8080/ready
# Expected: {"status":"ready"}

# Create a test merchant
curl -X POST http://localhost:8080/api/v1/merchants \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Merchant","webhook_url":"https://example.com/webhook"}'
```

## Environment Variables

All configuration is via environment variables with the `PCP_` prefix:

| Variable | Default | Description |
|----------|---------|-------------|
| PCP_SERVER_HOST | 0.0.0.0 | API server bind address |
| PCP_SERVER_PORT | 8080 | API server port |
| PCP_DATABASE_HOST | localhost | PostgreSQL host |
| PCP_DATABASE_PORT | 5432 | PostgreSQL port |
| PCP_DATABASE_USER | pcp | PostgreSQL user |
| PCP_DATABASE_PASSWORD | pcp_secret | PostgreSQL password |
| PCP_DATABASE_NAME | pcp | PostgreSQL database |
| PCP_DATABASE_SSL_MODE | disable | SSL mode |
| PCP_REDIS_HOST | localhost | Redis host |
| PCP_REDIS_PORT | 6379 | Redis port |
| PCP_LOG_LEVEL | info | Log level (debug/info/warn/error) |
| PCP_LOG_FORMAT | json | Log format (json/console) |
| PCP_JWT_SECRET | (required) | JWT signing secret |

## Common Commands

```bash
# Start in background
cd docker && docker compose up -d

# View logs
cd docker && docker compose logs -f api

# Stop everything
cd docker && docker compose down

# Stop and clean volumes
cd docker && docker compose down -v

# Rebuild after code changes
cd docker && docker compose up --build

# Run migrations manually
cd backend && make migrate-up
```

## Database Access

```bash
# Connect to PostgreSQL directly
docker exec -it pcp-postgres psql -U pcp -d pcp

# View merchants table
SELECT * FROM merchants;
```

## Troubleshooting

### API won't start
- Check if PostgreSQL is healthy: `docker compose ps`
- Check logs: `docker compose logs postgres`
- Ensure port 5432 is not used by another process

### Migration fails
- Check migration logs: `docker compose logs migrate`
- Ensure PostgreSQL is fully started before migrations run (Docker health checks handle this)

### Port conflicts
- Change ports in `docker-compose.yml` if 5432, 6379, or 8080 are in use
