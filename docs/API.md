# API Reference

## Base URL

```
http://localhost:8080
```

## Authentication

> **Sprint 1**: No authentication enforced. JWT and API key auth will be added in Sprint 2.

---

## Health Endpoints

### GET /health

Liveness probe. Returns 200 if the service process is running.

**Response** `200 OK`:
```json
{
  "status": "healthy",
  "service": "pcp-api"
}
```

### GET /ready

Readiness probe. Returns 200 if the service and all dependencies are healthy.

**Response** `200 OK`:
```json
{
  "status": "ready"
}
```

**Response** `503 Service Unavailable`:
```json
{
  "error": "service_unavailable",
  "message": "database connection failed",
  "code": 503
}
```

---

## Merchant Endpoints

### POST /api/v1/merchants

Create a new merchant. An API key is auto-generated.

**Request Body**:
```json
{
  "name": "Acme Corp",
  "webhook_url": "https://acme.com/webhooks/pcp"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | Yes | Unique merchant name (1-255 chars) |
| webhook_url | string | No | URL for webhook notifications |

**Response** `201 Created`:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Acme Corp",
  "api_key": "pcp_a1b2c3d4e5f6...",
  "webhook_url": "https://acme.com/webhooks/pcp",
  "status": "active",
  "created_at": "2026-07-18T08:00:00Z",
  "updated_at": "2026-07-18T08:00:00Z"
}
```

**Errors**:
- `400` — Invalid request body or missing name
- `409` — Merchant with same name already exists

---

### GET /api/v1/merchants

List all merchants with pagination.

**Query Parameters**:

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| offset | int | 0 | Pagination offset |
| limit | int | 20 | Page size (max 100) |

**Response** `200 OK`:
```json
{
  "data": [
    {
      "id": "...",
      "name": "Acme Corp",
      "api_key": "pcp_...",
      "webhook_url": "...",
      "status": "active",
      "created_at": "...",
      "updated_at": "..."
    }
  ],
  "total": 42,
  "offset": 0,
  "limit": 20
}
```

---

### GET /api/v1/merchants/{id}

Get a merchant by UUID.

**Response** `200 OK`: Single merchant object.

**Errors**:
- `400` — Invalid UUID format
- `404` — Merchant not found

---

### PUT /api/v1/merchants/{id}

Update a merchant. Only provided fields are changed.

**Request Body**:
```json
{
  "name": "Acme Corp Updated",
  "webhook_url": "https://new-url.com/webhook",
  "status": "suspended"
}
```

All fields are optional. Valid status values: `active`, `suspended`, `inactive`.

**Response** `200 OK`: Updated merchant object.

**Errors**:
- `400` — Invalid request
- `404` — Merchant not found
- `409` — Name conflict

---

### DELETE /api/v1/merchants/{id}

Delete a merchant.

**Response** `204 No Content`

**Errors**:
- `400` — Invalid UUID
- `404` — Merchant not found

---

## Error Response Format

All errors follow a consistent format:

```json
{
  "error": "error_type",
  "message": "Human-readable description",
  "code": 400
}
```

## Headers

| Header | Description |
|--------|-------------|
| X-Request-ID | Unique request ID (auto-generated if not provided) |
| Content-Type | `application/json` for all responses |
