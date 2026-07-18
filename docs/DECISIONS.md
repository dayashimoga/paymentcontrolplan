# Architecture Decision Records

## ADR-001: Go as Backend Language

**Status**: Accepted  
**Date**: 2026-07-18

**Context**: Need a language suitable for high-performance payment orchestration with strong concurrency, type safety, and operational simplicity.

**Decision**: Go 1.23 — strong standard library, excellent concurrency primitives, fast compilation, single binary deployment, and strong ecosystem for payment/fintech.

**Consequences**: Team must know Go. No generics-heavy patterns. Testing requires Go-specific tooling.

---

## ADR-002: Chi over Gin for HTTP Router

**Status**: Accepted  
**Date**: 2026-07-18

**Context**: Need a production-grade HTTP router that's lightweight, performant, and idiomatic.

**Decision**: Chi — `net/http` compatible, composable middleware, stdlib patterns, no framework lock-in.

**Alternatives Considered**: Gin (more popular but heavier, custom context), stdlib (too minimal for routing), Echo (similar to Gin).

---

## ADR-003: pgx over database/sql for PostgreSQL

**Status**: Accepted  
**Date**: 2026-07-18

**Context**: Need high-performance PostgreSQL connectivity with full feature support.

**Decision**: pgx v5 — native PostgreSQL protocol, connection pooling, batch queries, LISTEN/NOTIFY support, significantly faster than database/sql.

---

## ADR-004: Clean + Hexagonal Architecture

**Status**: Accepted  
**Date**: 2026-07-18

**Context**: Payment orchestration requires high testability, provider abstraction, and long-term maintainability.

**Decision**: Combine Clean Architecture (layer separation) with Hexagonal Architecture (ports & adapters) and DDD (bounded contexts).

**Consequences**: More files/directories than a flat structure. Clear boundaries enable independent testing and provider swapping.

---

## ADR-005: golang-migrate for Database Migrations

**Status**: Accepted  
**Date**: 2026-07-18

**Context**: Need version-controlled, reproducible database schema evolution.

**Decision**: golang-migrate — SQL-based migrations, Docker-friendly, CLI and library modes, supports PostgreSQL natively.

**Alternatives Considered**: goose (less Docker support), Atlas (newer, less battle-tested), custom scripts (maintenance burden).

---

## ADR-006: Viper for Configuration

**Status**: Accepted  
**Date**: 2026-07-18

**Context**: Need 12-Factor App compliant configuration supporting env vars, files, and defaults.

**Decision**: Viper — industry standard for Go, supports env vars, YAML, JSON, and nested keys with automatic binding.

---

## ADR-007: Zap for Structured Logging

**Status**: Accepted  
**Date**: 2026-07-18

**Context**: Need high-performance structured logging for production observability.

**Decision**: Zap — zero-allocation JSON logging, configurable levels, standard in high-throughput Go services.
