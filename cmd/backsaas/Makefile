# BackSaas CLI Makefile
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

.PHONY: build test clean install dev help format lint deps

# Build configuration
BINARY_NAME=backsaas
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
	@echo "BackSaas CLI Build Commands"
	@echo "=========================="
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the CLI binary
build: $(BUILD_DIR)/$(BINARY_NAME) ## Build the CLI binary in Docker container

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES)
	@echo "🔨 Building BackSaas CLI in Docker container..."
	@mkdir -p $(BUILD_DIR)
	# NOTE: Using Docker container to ensure consistent build environment
	$(DOCKER_GO_BUILD) go build -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go
	@echo "✅ Built: $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test: ## Run all tests in Docker container
	@echo "🧪 Running CLI tests in Docker container..."
	# NOTE: Using temporary Docker container for testing
	$(DOCKER_GO_RUN) go test ./internal/cli/...

test-verbose: ## Run tests with verbose output in Docker container
	@echo "🧪 Running verbose CLI tests in Docker container..."
	$(DOCKER_GO_RUN) go test -v ./internal/cli/...

# Development targets
dev: build ## Build and run with sample commands
	@echo "🚀 Testing CLI..."
	./$(BUILD_DIR)/$(BINARY_NAME) --help
	@echo ""
	@echo "🏥 Testing health check..."
	./$(BUILD_DIR)/$(BINARY_NAME) health check || true

# Format code
format: ## Format Go code in Docker container
	@echo "🎨 Formatting Go code in Docker container..."
	# NOTE: Using temporary Docker container for formatting
	$(DOCKER_GO_RUN) go fmt ./...

# Lint code
lint: ## Run golangci-lint in Docker container
	@echo "🔍 Running linter in Docker container..."
	# NOTE: Using golangci-lint Docker image for consistency
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:latest golangci-lint run

# Download dependencies
deps: ## Download dependencies in Docker container
	@echo "📦 Downloading dependencies in Docker container..."
	# NOTE: Create Go cache directories if they don't exist
	@mkdir -p $(HOME)/go/pkg/mod $(HOME)/.cache/go-build
	# NOTE: Using temporary Docker container for dependency management with cache mounts
	$(DOCKER_GO_RUN) go mod download
	$(DOCKER_GO_RUN) go mod tidy

# Cross-platform release builds
release: ## Build for multiple platforms in Docker container
	@echo "🚀 Building release binaries in Docker container..."
	@mkdir -p $(BUILD_DIR)/release
	
	# Linux AMD64
	$(DOCKER_GO_RUN) sh -c "GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 ./main.go"
	
	# Linux ARM64
	$(DOCKER_GO_RUN) sh -c "GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-arm64 ./main.go"
	
	# macOS AMD64
	$(DOCKER_GO_RUN) sh -c "GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 ./main.go"
	
	# macOS ARM64 (Apple Silicon)
	$(DOCKER_GO_RUN) sh -c "GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64 ./main.go"
	
	# Windows AMD64
	$(DOCKER_GO_RUN) sh -c "GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/release/$(BINARY_NAME)-windows-amd64.exe ./main.go"
	
	@echo "✅ Release binaries built in $(BUILD_DIR)/release/"
	@ls -la $(BUILD_DIR)/release/

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)/

# Install to system (requires local Go)
install: build ## Install CLI to system (requires local Go installation)
	@echo "📦 Installing BackSaas CLI to system..."
	@echo "⚠️  Note: This requires local Go installation"
	go install ./main.go
	@echo "✅ Installed to $(shell go env GOPATH 2>/dev/null || echo 'GOPATH not set')/bin/"

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
	# NOTE: Go version check now runs in Docker container
	@echo "  Go Version: $$($(DOCKER_GO_RUN) go version)"
	@echo "  Git Commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"

# Create bin directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)
