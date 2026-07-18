# Testing

## Quick Start

```bash
# Run all tests with coverage
cd backend && make test

# Run unit tests only
cd backend && make test-unit

# Run integration tests only
cd backend && make test-integration
```

## Test Structure

Tests follow the same architecture as the source:

```
backend/internal/
├── domain/merchant/merchant_test.go         # Entity validation, status transitions
├── application/merchant/service_test.go     # Use case logic with mock repos
└── interfaces/http/handler/
    ├── health_test.go                       # Health endpoint tests
    ├── merchant_test.go                     # CRUD handler tests
    └── mock_test.go                         # Shared test mocks
```

## Test Categories

| Type | Location | Description |
|------|----------|-------------|
| Unit | `*_test.go` beside source | Domain logic, service logic |
| Integration | `*_integration_test.go` | Real database (testcontainers) |
| API | `handler/*_test.go` | HTTP handler with httptest |
| Smoke | Docker health checks | Service alive + DB connected |

## Coverage Requirements

- **Minimum**: 60% overall (increasing to 90% as sprints progress)
- **Target**: 90%+ on domain and application layers
- Coverage enforced in CI pipeline

## Running in CI

GitHub Actions runs the full test suite on every PR:
1. Linting (golangci-lint, go vet)
2. Unit tests with race detection
3. Integration tests with PostgreSQL service container
4. Coverage report generation and threshold enforcement

## Writing Tests

### Unit Tests
- Use standard `testing` package
- Mock external dependencies via interfaces
- Test both happy path and error cases
- Table-driven tests for multiple scenarios

### Integration Tests
- Use `testcontainers` for real database
- Tag with `//go:build integration`
- Clean up test data after each test

### API Tests
- Use `httptest.NewRecorder()` and `httptest.NewRequest()`
- Test full HTTP lifecycle (routing, middleware, handler)
- Verify response status codes, headers, and body
