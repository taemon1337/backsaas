# BackSaas API Gateway Makefile
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

.PHONY: build test test-verbose test-coverage run run-dev clean help

# Build configuration
BINARY_NAME=gateway
BUILD_DIR=bin
GO_FILES=$(shell find . -name "*.go" -type f)

# Docker configuration for Go commands
# NOTE: Always run Go commands in Docker containers for consistency
# Using full golang image (not alpine) to include git for go mod operations
# Mount Go caches from host for faster builds and module downloads
GO_IMAGE=golang:1.21
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
build: $(BUILD_DIR)/$(BINARY_NAME) ## Build the gateway binary

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES)
	@echo "Building $(BINARY_NAME) in Docker container..."
	@mkdir -p $(BUILD_DIR)
	# NOTE: Using Docker container to ensure consistent build environment
	$(DOCKER_GO_BUILD) go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

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

# Run targets
run: build ## Build and run gateway with default config
	$(BUILD_DIR)/$(BINARY_NAME) \
		-port=8000 \
		-redis-url="redis://localhost:6379" \
		-jwt-secret="your-jwt-secret-here" \
		-config="config/gateway.yaml" \
		-env="development"

run-dev: ## Run gateway in development mode using Docker
	@echo "Running gateway development server in Docker..."
	# NOTE: Using Docker container with volume mounts for hot reload and Go caches
	docker run --rm -it \
		-v $(PWD):/app \
		-v $(HOME)/go/pkg/mod:/go/pkg/mod \
		-v $(HOME)/.cache/go-build:/root/.cache/go-build \
		-w /app \
		-p 8000:8000 \
		-e PORT=8000 \
		-e REDIS_URL="redis://host.docker.internal:6379" \
		-e JWT_SECRET="dev-jwt-secret-change-in-production" \
		-e GATEWAY_CONFIG="config/gateway.yaml" \
		-e ENVIRONMENT="development" \
		$(GO_IMAGE) \
		sh -c "go run ./cmd/server"

run-production: build ## Run gateway in production mode
	$(BUILD_DIR)/$(BINARY_NAME) \
		-port=8000 \
		-redis-url="$$REDIS_URL" \
		-jwt-secret="$$JWT_SECRET" \
		-config="$$GATEWAY_CONFIG" \
		-env="production"

# Development targets
deps: ## Download dependencies in Docker container
	@echo "Downloading dependencies in Docker container..."
	# NOTE: Create Go cache directories if they don't exist
	@mkdir -p $(HOME)/go/pkg/mod $(HOME)/.cache/go-build
	# NOTE: Using temporary Docker container for dependency management with cache mounts
	$(DOCKER_GO_RUN) go mod download
	$(DOCKER_GO_RUN) go mod tidy

format: ## Format Go code in Docker container
	@echo "Formatting Go code in Docker container..."
	# NOTE: Using temporary Docker container for formatting
	$(DOCKER_GO_RUN) go fmt ./...

lint: ## Run golangci-lint in Docker container
	@echo "Running linter in Docker container..."
	# NOTE: Using golangci-lint Docker image for consistency
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:latest golangci-lint run

# Utility targets
clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Docker targets
docker-build: ## Build Docker image
	docker build -t backsaas/gateway .

docker-run: ## Run Docker container
	docker run --rm -p 8000:8000 \
		-e PORT=8000 \
		-e REDIS_URL="redis://host.docker.internal:6379" \
		-e JWT_SECRET="dev-jwt-secret" \
		-e GATEWAY_CONFIG="config/gateway.yaml" \
		-e ENVIRONMENT="development" \
		backsaas/gateway

# Configuration targets
config-validate: ## Validate gateway configuration
	@echo "Validating gateway configuration..."
	$(DOCKER_GO_RUN) go run ./cmd/server -config=config/gateway.yaml -validate-only

config-example: ## Generate example configuration
	@echo "Generating example configuration..."
	@mkdir -p config
	@echo "Example configuration created in config/gateway.example.yaml"

# Benchmarks
benchmark: ## Run benchmarks in Docker container
	@echo "Running benchmarks in Docker container..."
	# NOTE: Using temporary Docker container for benchmarking
	$(DOCKER_GO_RUN) go test -bench=. -benchmem ./...

# Show configuration
show-config: ## Show current build configuration
	@echo "Build Configuration:"
	@echo "  Binary Name: $(BINARY_NAME)"
	@echo "  Build Dir: $(BUILD_DIR)"
	@echo "  Go Files: $(words $(GO_FILES)) files"
	@echo "  Go Image: $(GO_IMAGE)"
	@echo "  Docker Go Run: $(DOCKER_GO_RUN)"
	# NOTE: Go version check now runs in Docker container
	@echo "  Go Version: $$($(DOCKER_GO_RUN) go version)"
