# ============================================================================
# PCP Backend - Multi-stage Docker Build
# ============================================================================
# Stage 1: Build the Go binary
# Stage 2: Minimal runtime image (distroless)
# ============================================================================

FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy dependency files first for layer caching
COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true

# Copy source code
COPY . .

# Resolve dependencies (in case go.sum doesn't exist yet)
RUN go mod tidy

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always 2>/dev/null || echo 'dev')" \
    -o /build/pcp-api \
    ./cmd/api

# ============================================================================
# Stage 2: Minimal runtime
# ============================================================================
FROM alpine:3.20 AS runtime

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -S pcp && adduser -S pcp -G pcp

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/pcp-api .

# Copy migrations for runtime migration support
COPY --from=builder /build/migrations ./migrations

# Use non-root user
USER pcp

# Expose API port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["./pcp-api"]
