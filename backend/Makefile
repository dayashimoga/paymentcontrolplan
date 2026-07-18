# ============================================================================
# PCP Backend Makefile
# ============================================================================

.PHONY: all build test lint clean run migrate-up migrate-down tidy docker-build docker-up docker-down

# Variables
APP_NAME    := pcp-api
BUILD_DIR   := ./bin
MAIN_PATH   := ./cmd/api
COVERAGE    := ./coverage

# Go commands
GOTEST      := go test
GOBUILD     := go build
GOVET       := go vet

all: lint test build

# ----------------------------- Build -----------------------------------

build:
	@echo "==> Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

run: build
	@$(BUILD_DIR)/$(APP_NAME)

clean:
	@rm -rf $(BUILD_DIR) $(COVERAGE)

# ----------------------------- Dependencies ----------------------------

tidy:
	@echo "==> Tidying Go modules..."
	go mod tidy

# ----------------------------- Testing ---------------------------------

test:
	@echo "==> Running tests..."
	@mkdir -p $(COVERAGE)
	$(GOTEST) -v -race -coverprofile=$(COVERAGE)/coverage.out -covermode=atomic ./...
	@go tool cover -func=$(COVERAGE)/coverage.out
	@go tool cover -html=$(COVERAGE)/coverage.out -o $(COVERAGE)/coverage.html
	@echo "==> Coverage report: $(COVERAGE)/coverage.html"

test-unit:
	@echo "==> Running unit tests..."
	$(GOTEST) -v -short -race ./...

test-integration:
	@echo "==> Running integration tests..."
	$(GOTEST) -v -run Integration -race ./...

# ----------------------------- Linting ---------------------------------

lint:
	@echo "==> Running linter..."
	$(GOVET) ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

# ----------------------------- Migrations ------------------------------

migrate-up:
	@echo "==> Running migrations up..."
	@docker run --rm --network host \
		-v $(CURDIR)/migrations:/migrations \
		migrate/migrate \
		-path=/migrations -database "postgres://pcp:pcp_secret@localhost:5432/pcp?sslmode=disable" up

migrate-down:
	@echo "==> Running migrations down..."
	@docker run --rm --network host \
		-v $(CURDIR)/migrations:/migrations \
		migrate/migrate \
		-path=/migrations -database "postgres://pcp:pcp_secret@localhost:5432/pcp?sslmode=disable" down 1

# ----------------------------- Docker ----------------------------------

docker-build:
	@echo "==> Building Docker image..."
	docker build -t $(APP_NAME):latest .

docker-up:
	@echo "==> Starting services..."
	docker compose -f ../docker/docker-compose.yml up --build -d

docker-down:
	@echo "==> Stopping services..."
	docker compose -f ../docker/docker-compose.yml down

docker-logs:
	docker compose -f ../docker/docker-compose.yml logs -f

# ----------------------------- Help ------------------------------------

help:
	@echo "Available targets:"
	@echo "  all              - Lint, test, and build"
	@echo "  build            - Build the binary"
	@echo "  run              - Build and run locally"
	@echo "  test             - Run all tests with coverage"
	@echo "  test-unit        - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  lint             - Run linters"
	@echo "  tidy             - Tidy Go modules"
	@echo "  migrate-up       - Run database migrations"
	@echo "  migrate-down     - Rollback last migration"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-up        - Start all services"
	@echo "  docker-down      - Stop all services"
	@echo "  clean            - Remove build artifacts"
