# BackSaas Monorepo Makefile
# 
# IMPORTANT: This Makefile runs ALL commands inside Docker containers
# to ensure consistent development environments across different machines.
# 
# Requirements:
# - Docker must be installed and running
# - No local tool installations required (Go, Node.js, etc.)
# 
# All operations use temporary Docker containers that are automatically
# removed after execution (--rm flag).

.PHONY: help build test clean dev setup dev-up dev-down dev-logs dev-status test-all test-unit test-integration test-setup test-clean mod-tidy

# Docker images for different tools
# NOTE: Always use specific versions for reproducible builds
# Using full golang image (not alpine) to include git for go mod operations
GO_IMAGE=golang:1.23
NODE_IMAGE=node:18-alpine
POSTGRES_IMAGE=postgres:15-alpine

# Docker run commands
# NOTE: Using temporary containers with volume mounts
# Mount Go caches from host for faster builds and module downloads
DOCKER_GO=docker run --rm \
	-v $(PWD):/app \
	-v $(HOME)/go/pkg/mod:/go/pkg/mod \
	-v $(HOME)/.cache/go-build:/root/.cache/go-build \
	-w /app \
	$(GO_IMAGE)
DOCKER_NODE=docker run --rm -v $(PWD):/app -w /app $(NODE_IMAGE)

# Default target
help: ## Show this help message
	@echo "BackSaas Development Commands (Docker-based)"
	@echo "============================================="
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "NOTE: All commands run in Docker containers - no local tools required!"

# Setup targets
setup: ## Setup development environment with Docker containers
	@echo "Setting up BackSaas development environment..."
	@echo "NOTE: Using Docker containers for all services"
	$(MAKE) setup-db
	$(MAKE) setup-services
	@echo "Setup complete! Use 'make dev' to start development servers."

setup-db: ## Setup PostgreSQL database container
	@echo "Setting up PostgreSQL database..."
	@docker run --rm --name backsaas-db -d \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=backsaas \
		-p 5432:5432 \
		$(POSTGRES_IMAGE) || echo "Database container already exists"
	@sleep 3
	@docker exec backsaas-db psql -U postgres -c "CREATE DATABASE backsaas_platform;" || echo "Platform DB exists"
	@docker exec backsaas-db psql -U postgres -c "CREATE DATABASE backsaas_test;" || echo "Test DB exists"

setup-services: ## Install dependencies for all services using Docker
	@echo "📦 Downloading dependencies in Docker container..."
	# NOTE: Create Go cache directories if they don't exist
	@mkdir -p $(HOME)/go/pkg/mod $(HOME)/.cache/go-build
	# NOTE: Platform API dependencies (Go modules) with cache mounts
	@echo "Installing Platform API Go dependencies..."
	cd services/platform-api && docker run --rm \
		-v $(PWD)/services/platform-api:/app \
		-v $(HOME)/go/pkg/mod:/go/pkg/mod \
		-v $(HOME)/.cache/go-build:/root/.cache/go-build \
		-w /app \
		$(GO_IMAGE) go mod download
	# NOTE: Gateway dependencies (Go modules) with cache mounts
	@echo "Installing Gateway Go dependencies..."
	cd services/gateway && docker run --rm \
		-v $(PWD)/services/gateway:/app \
		-v $(HOME)/go/pkg/mod:/go/pkg/mod \
		-v $(HOME)/.cache/go-build:/root/.cache/go-build \
		-w /app \
		$(GO_IMAGE) go mod download
	# NOTE: Web UI dependencies (npm packages) - skip if apps/web doesn't exist
	@if [ -d "apps/web" ]; then \
		echo "Installing Node.js dependencies..."; \
		cd apps/web && docker run --rm -v $(PWD)/apps/web:/app -w /app $(NODE_IMAGE) npm install; \
	else \
		echo "Skipping Web UI setup (apps/web not found)"; \
	fi

mod-tidy: ## Run go mod tidy on all Go modules using Docker
	@echo "🧹 Running go mod tidy on all Go modules in Docker containers..."
	@mkdir -p $(HOME)/go/pkg/mod $(HOME)/.cache/go-build
	# NOTE: Platform API
	@echo "Tidying Platform API Go modules..."
	cd services/platform-api && $(DOCKER_GO) go mod tidy
	# NOTE: Gateway
	@echo "Tidying Gateway Go modules..."
	cd services/gateway && $(DOCKER_GO) go mod tidy
	# NOTE: API Service
	@echo "Tidying API Service Go modules..."
	cd services/api && $(DOCKER_GO) go mod tidy
	# NOTE: Migrator
	@echo "Tidying Migrator Go modules..."
	cd services/migrator && $(DOCKER_GO) go mod tidy
	# NOTE: CLI
	@echo "Tidying CLI Go modules..."
	cd cmd/backsaas && $(DOCKER_GO) go mod tidy
	@echo "✅ All Go modules tidied"

# Build targets
build: ## Build all services using Docker containers
	@echo "Building all services in Docker containers..."
	$(MAKE) build-platform-api
	$(MAKE) build-gateway
	$(MAKE) build-web

build-platform-api: ## Build Platform API service
	@echo "Building Platform API in Docker container..."
	# NOTE: Using Docker container for Go build
	cd services/platform-api && $(MAKE) build

build-gateway: ## Build Gateway service
	@echo "Building Gateway in Docker container..."
	# NOTE: Using Docker container for Go build
	cd services/gateway && $(MAKE) build

build-web: ## Build Web UI
	@echo "Building Web UI in Docker container..."
	# NOTE: Using Docker container for Node.js build - skip if doesn't exist
	@if [ -d "apps/web" ]; then \
		cd apps/web && docker run --rm -v $(PWD)/apps/web:/app -w /app $(NODE_IMAGE) npm run build; \
	else \
		echo "Skipping Web UI build (apps/web not found)"; \
	fi

# Test targets
test: test-all ## Run all tests using enhanced testing infrastructure (alias for test-all)

# Enhanced testing infrastructure
test-all: ## Run comprehensive tests in isolated Docker environment
	@echo "🚀 Running comprehensive BackSaaS tests..."
	@echo "📋 Using enhanced testing infrastructure with isolated environment"
	@$(MAKE) -f Makefile.test test-all

test-unit: ## Run unit tests only
	@echo "🧪 Running unit tests in isolated environment..."
	@$(MAKE) -f Makefile.test test-unit

test-integration: ## Run integration tests only
	@echo "🔗 Running integration tests in isolated environment..."
	@$(MAKE) -f Makefile.test test-integration

test-setup: ## Setup isolated test environment
	@echo "🔧 Setting up isolated test environment..."
	@$(MAKE) -f Makefile.test test-setup

test-clean: ## Clean test environment and results
	@echo "🧹 Cleaning test environment..."
	@$(MAKE) -f Makefile.test test-clean

test-coverage: ## Generate test coverage reports
	@echo "📊 Generating coverage reports..."
	@$(MAKE) -f Makefile.test test-coverage

test-results: ## Start test results server
	@echo "🌐 Starting test results server..."
	@$(MAKE) -f Makefile.test test-results-server
	@echo "📊 View results at: http://localhost:8888"

test-e2e: ## Run CLI-based end-to-end platform tests
	@echo "🎯 Running CLI-based end-to-end platform tests..."
	@$(MAKE) -f Makefile.test test-e2e

test-cli-platform: ## Run CLI platform tests within Docker network
	@echo "🏗️  Running CLI platform tests..."
	@$(MAKE) -f Makefile.test test-cli-platform

test-cli-tenant-lifecycle: ## Run CLI tenant lifecycle tests
	@echo "🏢 Running CLI tenant lifecycle tests..."
	@$(MAKE) -f Makefile.test test-cli-tenant-lifecycle

# Legacy test targets (for backward compatibility)
test-platform-api: ## Test Platform API service (legacy)
	@echo "⚠️  Using legacy test method. Consider using 'make test-unit' instead."
	cd services/platform-api && $(MAKE) test

test-gateway: ## Test Gateway service (legacy)
	@echo "⚠️  Using legacy test method. Consider using 'make test-unit' instead."
	cd services/gateway && $(MAKE) test

test-web: ## Test Web UI (legacy)
	@echo "⚠️  Using legacy test method. Consider using 'make test-unit' instead."
	@if [ -d "apps/web" ]; then \
		cd apps/web && docker run --rm -v $(PWD)/apps/web:/app -w /app $(NODE_IMAGE) npm test; \
	else \
		echo "Skipping Web UI tests (apps/web not found)"; \
	fi

# Development targets
dev: dev-up ## Start development environment with Docker Compose (alias for dev-up)

dev-up: ## Start complete development environment with Docker Compose
	@echo "🚀 Starting BackSaas development environment..."
	@echo "📊 Services will be available at:"
	@echo "  - Platform API: http://localhost:8080"
	@echo "  - API Gateway: http://localhost:8000"
	@echo "  - Test Tenant API: http://localhost:8081"
	@echo "  - Database Admin: http://localhost:8082"
	@echo "  - Prometheus: http://localhost:9090 (with --profile monitoring)"
	@echo "  - Grafana: http://localhost:3001 (with --profile monitoring)"
	@echo ""
	@echo "💡 Use 'make dev-logs' to see logs, 'make dev-down' to stop"
	docker compose up --build -d

dev-down: ## Stop development environment
	@echo "🛑 Stopping BackSaas development environment..."
	docker compose down

dev-logs: ## Show logs from development environment
	@echo "📋 Showing logs from all services (Ctrl+C to exit)..."
	docker compose logs -f

dev-status: ## Show status of development services
	@echo "📊 BackSaas Development Environment Status:"
	@echo "==========================================="
	docker compose ps

dev-monitoring: ## Start development environment with monitoring
	@echo "🚀 Starting BackSaas with monitoring (Prometheus + Grafana)..."
	docker compose --profile monitoring up --build -d

dev-db-only: ## Start only database services
	@echo "🗄️ Starting only database services..."
	docker compose up postgres redis adminer -d

dev-platform-api: ## Start Platform API in development mode
	@echo "Starting Platform API development server..."
	cd services/platform-api && $(MAKE) dev

dev-gateway: ## Start Gateway in development mode
	@echo "Starting Gateway development server..."
	cd services/gateway && $(MAKE) run-dev

dev-web: ## Start Web UI in development mode
	@echo "Starting Web UI development server..."
	@if [ -d "apps/web" ]; then \
		cd apps/web && docker run --rm -v $(PWD)/apps/web:/app -w /app $(NODE_IMAGE) npm run dev; \
	else \
		echo "Skipping Web UI dev (apps/web not found)"; \
	fi

# Utility targets
clean: ## Clean all build artifacts
	@echo "Cleaning build artifacts..."
	cd services/platform-api && make clean
	cd services/gateway && make clean
	@if [ -d "apps/web" ]; then \
		cd apps/web && $(DOCKER_NODE) npm run clean || true; \
	fi
	docker system prune -f

clean-db: ## Clean up database containers
	@echo "Cleaning up database containers..."
	@docker stop backsaas-db || true
	@docker rm backsaas-db || true

format: ## Format all code using Docker containers
	@echo "Formatting all code in Docker containers..."
	# NOTE: Go formatting in Docker
	cd services/platform-api && $(DOCKER_GO) make format
	# NOTE: Node.js formatting in Docker
	cd apps/web && $(DOCKER_NODE) npm run format || true

lint: ## Lint all code using Docker containers
	@echo "Linting all code in Docker containers..."
	# NOTE: Go linting in Docker
	cd services/platform-api && $(DOCKER_GO) make lint
	# NOTE: Node.js linting in Docker
	cd apps/web && $(DOCKER_NODE) npm run lint || true

# Docker targets
docker-build: ## Build all Docker images
	@echo "Building all Docker images..."
	docker build -t backsaas/platform-api ./services/platform-api
	docker build -t backsaas/web ./apps/web

docker-up: ## Start all services with Docker Compose
	@echo "Starting all services with Docker Compose..."
	docker compose up --build -d

docker-down: ## Stop all services
	@echo "Stopping all services..."
	docker compose down

docker-logs: ## Show logs from all services
	docker compose logs -f

# Database management
db-migrate: ## Run database migrations using Docker
	@echo "Running database migrations in Docker container..."
	# NOTE: Using Docker container for database operations
	$(DOCKER_GO) sh -c "cd services/migrator && go run . migrate"

db-seed: ## Seed database with test data using Docker
	@echo "Seeding database with test data..."
	$(DOCKER_GO) sh -c "cd services/migrator && go run . seed"

db-reset: ## Reset database (clean + setup)
	$(MAKE) clean-db
	$(MAKE) setup-db

# Monitoring and debugging
logs: ## Show logs from development environment
	docker compose logs -f

ps: ## Show running containers
	docker compose ps

# Check tools
check-tools: ## Check if required tools are installed
	@echo "Checking required tools..."
	@command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed"; exit 1; }
	@docker compose version >/dev/null 2>&1 || { echo "Docker Compose is required but not installed"; exit 1; }
	@echo "All required tools are installed!"
	@echo "NOTE: Go, Node.js, and other tools will run in Docker containers"

# Show configuration
show-config: ## Show current configuration
	@echo "BackSaas Configuration:"
	@echo "======================"
	@echo "  Go Image: $(GO_IMAGE)"
	@echo "  Node Image: $(NODE_IMAGE)"
	@echo "  Postgres Image: $(POSTGRES_IMAGE)"
	@echo "  Docker Go Command: $(DOCKER_GO)"
	@echo "  Docker Node Command: $(DOCKER_NODE)"
	@echo ""
	@echo "Services:"
	@echo "  - Platform API: services/platform-api"
	@echo "  - Web UI: apps/web"
	@echo "  - Migrator: services/migrator"
	@echo ""
	@echo "NOTE: All development happens in Docker containers!"

# Development workflow helpers
quick-start: ## Quick start for new developers
	@echo "BackSaas Quick Start Guide"
	@echo "========================="
	@echo ""
	@echo "1. Check tools: make check-tools"
	@echo "2. Setup environment: make setup"
	@echo "3. Start development: make dev"
	@echo ""
	@echo "NOTE: Everything runs in Docker - no local installs needed!"

# Reset everything
reset: ## Reset entire development environment
	@echo "Resetting entire development environment..."
	$(MAKE) clean
	$(MAKE) clean-db
	$(MAKE) setup
	@echo "Environment reset complete!"
