# BackSaas Platform API Makefile
# 
# IMPORTANT: This Makefile runs ALL Go commands inside Docker containers
# to ensure consistent development environments across different machines.
# 
# Requirements:
# - Docker must be installed and running
# - No local Go installation required
# 
# All Go operations (build, test, format, lint) use temporary Docker containers
# that are automatically removed after execution (--rm flag).

.PHONY: build build-test-server test test-verbose test-coverage run run-platform run-tenant run-test-server clean help

# Build configuration
BINARY_NAME=platform-api
BUILD_DIR=bin
GO_FILES=$(shell find . -name "*.go" -type f)

# Docker configuration for Go commands
# NOTE: Always run Go commands in Docker containers for consistency
# Using full golang image (not alpine) to include git for go mod operations
# Mount Go caches from host for faster builds and module downloads
GO_IMAGE=golang:1.25
DOCKER_GO_RUN=docker run --rm \
	-v $(PWD):/app \
	-v $(HOME)/go/pkg/mod:/go/pkg/mod \
	-v $(HOME)/.cache/go-build:/root/.cache/go-build \
	-w /app \
	$(GO_IMAGE)
DOCKER_GO_BUILD=docker run --rm \
	-v $(PWD):/app \
	-v $(PWD)/$(BUILD_DIR):/app/$(BUILD_DIR) \
	-v $(HOME)/go/pkg/mod:/go/pkg/mod \
	-v $(HOME)/.cache/go-build:/root/.cache/go-build \
	-w /app \
	$(GO_IMAGE)

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: $(BUILD_DIR)/$(BINARY_NAME) ## Build the platform API binary

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES)
	@echo "Building $(BINARY_NAME) in Docker container..."
	@mkdir -p $(BUILD_DIR)
	# NOTE: Using Docker container to ensure consistent build environment
	$(DOCKER_GO_BUILD) go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

build-test-server: ## Build the test server binary
	@echo "Building test-server in Docker container..."
	@mkdir -p $(BUILD_DIR)
	$(DOCKER_GO_BUILD) go build -o $(BUILD_DIR)/test-server ./cmd/test-server

# Test targets
test: ## Run all tests in Docker container
	@echo "Running tests in Docker container..."
	# NOTE: Using temporary Docker container for testing
	$(DOCKER_GO_RUN) go test ./...

test-verbose: ## Run tests with verbose output in Docker container
	@echo "Running verbose tests in Docker container..."
	$(DOCKER_GO_RUN) go test -v ./...

test-coverage: ## Run tests with coverage report in Docker container
	@echo "Running coverage tests in Docker container..."
	# NOTE: Coverage files are written to host filesystem via volume mount
	$(DOCKER_GO_RUN) go test -coverprofile=coverage.out ./...
	$(DOCKER_GO_RUN) go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-functions: ## Run function tests specifically in Docker container
	@echo "Running function tests in Docker container..."
	$(DOCKER_GO_RUN) go test -v ./internal/functions/...

test-schema: ## Run schema tests specifically in Docker container
	@echo "Running schema tests in Docker container..."
	$(DOCKER_GO_RUN) go test -v ./internal/schema/...

test-api: ## Run API engine tests in Docker container (requires TEST_DATABASE_URL)
	@echo "Running API tests in Docker container..."
	@if [ -z "$(TEST_DATABASE_URL)" ]; then \
		echo "Warning: TEST_DATABASE_URL not set, skipping integration tests"; \
		$(DOCKER_GO_RUN) go test -v ./internal/api/... -short; \
	else \
		$(DOCKER_GO_RUN) sh -c "TEST_DATABASE_URL='$(TEST_DATABASE_URL)' go test -v ./internal/api/..."; \
	fi

test-docker: ## Run tests using Docker Compose network
	@echo "🧪 Running tests within Docker Compose network..."
	@echo "📋 Building and starting test runner service..."
	@cd ../.. && docker compose --profile test build test-runner
	@cd ../.. && docker compose --profile test up -d test-runner
	@echo "🔬 Executing database tests..."
	@cd ../.. && docker compose exec test-runner go test -v ./internal/api/... -run="TestDatabaseOperations|TestFieldMappingConsistency|TestCRMSchemaFieldMapping"
	@echo "🧹 Cleaning up test runner..."
	@cd ../.. && docker compose --profile test down

test-docker-all: ## Run all tests using Docker Compose network
	@echo "🧪 Running all tests within Docker Compose network..."
	@echo "📋 Building and starting test runner service..."
	@cd ../.. && docker compose --profile test build test-runner
	@cd ../.. && docker compose --profile test up -d test-runner
	@echo "🔬 Executing all tests..."
	@cd ../.. && docker compose exec test-runner go test -v ./...
	@echo "🧹 Cleaning up test runner..."
	@cd ../.. && docker compose --profile test down

test-field-mapping: ## Run specific field mapping tests
	@echo "🧪 Running field mapping tests..."
	@cd ../.. && docker compose --profile test build test-runner
	@cd ../.. && docker compose --profile test up -d test-runner
	@echo "🔬 Executing field mapping tests..."
	@cd ../.. && docker compose exec test-runner go test -v ./internal/api/... -run="FieldMapping"
	@echo "🧹 Cleaning up test runner..."
	@cd ../.. && docker compose --profile test down

test-integration: ## Run integration tests
	@echo "🧪 Running integration tests..."
	@cd ../.. && docker compose --profile test build test-runner
	@cd ../.. && docker compose --profile test up -d test-runner
	@echo "🔬 Executing integration tests..."
	@cd ../.. && docker compose exec test-runner go test -v ./tests/integration/...
	@echo "🧹 Cleaning up test runner..."
	@cd ../.. && docker compose --profile test down

# Run targets
run: build ## Build and run with test schema
	$(BUILD_DIR)/$(BINARY_NAME) \
		-tenant-id=test-tenant \
		-schema-source=file \
		-schema-path=testdata/test-schema.yaml \
		-database-url="postgres://postgres:postgres@localhost:5432/backsaas_test?sslmode=disable" \
		-port=8080

run-platform: build ## Build and run as platform API (system tenant)
	$(BUILD_DIR)/$(BINARY_NAME) \
		-tenant-id=system \
		-schema-source=file \
		-schema-path=../api/system/schema/platform.yaml \
		-database-url="postgres://postgres:postgres@localhost:5432/backsaas_platform?sslmode=disable" \
		-port=8080

run-tenant: build ## Build and run as tenant API (requires TENANT_ID and SCHEMA_PATH)
	@if [ -z "$(TENANT_ID)" ]; then \
		echo "Error: TENANT_ID is required"; \
		echo "Usage: make run-tenant TENANT_ID=acme-corp SCHEMA_PATH=path/to/schema.yaml"; \
		exit 1; \
	fi
	@if [ -z "$(SCHEMA_PATH)" ]; then \
		echo "Error: SCHEMA_PATH is required"; \
		echo "Usage: make run-tenant TENANT_ID=acme-corp SCHEMA_PATH=path/to/schema.yaml"; \
		exit 1; \
	fi

run-test-server: build-test-server ## Build and run test server with sample CRM schema
	@echo "🚀 Starting test server with sample CRM schema..."
	@echo "📊 Available at: http://localhost:8081"
	@echo "💡 Use Ctrl+C to stop"
	$(BUILD_DIR)/test-server

# Development targets
dev: ## Run in development mode with hot reload using Docker
	@echo "Running development server with hot reload in Docker..."
	# NOTE: Using Docker container with volume mounts for hot reload and Go caches
	docker run --rm -it \
		-v $(PWD):/app \
		-v $(HOME)/go/pkg/mod:/go/pkg/mod \
		-v $(HOME)/.cache/go-build:/root/.cache/go-build \
		-w /app \
		-p 8080:8080 \
		-e TENANT_ID=test-tenant \
		-e SCHEMA_SOURCE=file \
		-e SCHEMA_PATH=testdata/test-schema.yaml \
		-e DATABASE_URL="postgres://postgres:postgres@host.docker.internal:5432/backsaas_test?sslmode=disable" \
		$(GO_IMAGE) \
		sh -c "go run ./cmd/server"

format: ## Format Go code in Docker container
	@echo "Formatting Go code in Docker container..."
	# NOTE: Using temporary Docker container for formatting
	$(DOCKER_GO_RUN) go fmt ./...

lint: ## Run golangci-lint in Docker container
	@echo "Running linter in Docker container..."
	# NOTE: Using golangci-lint Docker image for consistency
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:latest golangci-lint run

# Database targets
db-setup: ## Setup test databases
	@echo "Setting up test databases..."
	@docker run --name backsaas-test-db -d \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=backsaas_test \
		-p 5432:5432 \
		postgres:15-alpine || echo "Database container already exists"
	@sleep 2
	@docker exec backsaas-test-db psql -U postgres -c "CREATE DATABASE backsaas_platform;" || echo "Platform DB already exists"

db-clean: ## Clean up test databases
	@echo "Cleaning up test databases..."
	@docker stop backsaas-test-db || true
	@docker rm backsaas-test-db || true

# Docker targets
docker-build: ## Build Docker image
	docker build -t backsaas/platform-api .

docker-run: ## Run Docker container
	docker run --rm -p 8080:8080 \
		-e TENANT_ID=test-tenant \
		-e SCHEMA_SOURCE=file \
		-e SCHEMA_PATH=testdata/test-schema.yaml \
		-e DATABASE_URL="postgres://postgres:postgres@host.docker.internal:5432/backsaas_test?sslmode=disable" \
		backsaas/platform-api

# Utility targets
clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

deps: ## Download dependencies in Docker container
	@echo "Downloading dependencies in Docker container..."
	# NOTE: Create Go cache directories if they don't exist
	@mkdir -p $(HOME)/go/pkg/mod $(HOME)/.cache/go-build
	# NOTE: Using temporary Docker container for dependency management with cache mounts
	$(DOCKER_GO_RUN) go mod download
	$(DOCKER_GO_RUN) go mod tidy

# Benchmarks
benchmark: ## Run benchmarks in Docker container
	@echo "Running benchmarks in Docker container..."
	# NOTE: Using temporary Docker container for benchmarking
	$(DOCKER_GO_RUN) go test -bench=. -benchmem ./...

# Integration tests with different schemas
test-integration: ## Run integration tests with multiple schemas in Docker
	@echo "Running integration tests in Docker container..."
	# NOTE: Using temporary Docker container for integration testing
	@$(DOCKER_GO_RUN) sh -c "TEST_DATABASE_URL='postgres://postgres:postgres@host.docker.internal:5432/backsaas_test?sslmode=disable' go test -v ./internal/api/..."
	@echo "Testing schema loading in Docker container..."
	@$(DOCKER_GO_RUN) go test -v ./internal/schema/...

# Generate test data
generate-test-data: ## Generate additional test schemas using Docker
	@echo "Generating test data in Docker container..."
	@mkdir -p testdata/generated
	# NOTE: Using temporary Docker container for test data generation
	@$(DOCKER_GO_RUN) sh -c "echo '# Generated test schema' > testdata/generated/example.yaml"

# Check if all required tools are installed
check-tools: ## Check if required development tools are installed
	@echo "Checking required tools..."
	@command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed"; exit 1; }
	@echo "Docker is installed - Go will run in containers"
	@echo "All required tools are available!"

# Show current configuration
show-config: ## Show current build configuration
	@echo "Build Configuration:"
	@echo "  Binary Name: $(BINARY_NAME)"
	@echo "  Build Dir: $(BUILD_DIR)"
	@echo "  Go Files: $(words $(GO_FILES)) files"
	@echo "  Go Image: $(GO_IMAGE)"
	@echo "  Docker Go Run: $(DOCKER_GO_RUN)"
	# NOTE: Go version check now runs in Docker container
	@echo "  Go Version: $$($(DOCKER_GO_RUN) go version)"
	@echo "  Git Commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
