# BackSaaS Simplified Makefile
# ============================
# Service-oriented commands that map directly to Docker Compose services

.PHONY: help up down restart logs status build test clean

help: ## Show available commands
	@echo "🏗️  BackSaaS Development Commands"
	@echo "=================================="
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "📦 Services: postgres, redis, platform-api, gateway, health-dashboard"
	@echo "🔧 Profiles: monitoring, test"

# === CORE SERVICES ===
up: ## Start all core services
	@echo "🚀 Starting BackSaaS core services..."
	docker compose up -d postgres redis platform-api gateway admin-console control-plane tenant-ui health-dashboard
	@echo "✅ Services started! Check with 'make status'"

down: ## Stop all services
	@echo "🛑 Stopping all services..."
	docker compose down
	@echo "✅ All services stopped"

restart: ## Restart all core services
	@echo "🔄 Restarting services..."
	docker compose restart postgres redis platform-api gateway admin-console control-plane health-dashboard
	@echo "✅ Services restarted"

# === SERVICE MANAGEMENT ===
restart-api: ## Restart platform-api
	@echo "🔄 Restarting platform-api..."
	docker compose restart platform-api

restart-gateway: ## Restart gateway
	@echo "🔄 Restarting gateway..."
	docker compose restart gateway

restart-admin: ## Restart admin console
	@echo "🔄 Restarting admin console..."
	docker compose restart admin-console

restart-control-plane: ## Restart control plane
	@echo "🔄 Restarting control plane..."
	docker compose restart control-plane

restart-tenant-ui: ## Restart tenant UI
	@echo "🔄 Restarting tenant UI..."
	docker compose restart tenant-ui

restart-dashboard: ## Restart health dashboard
	@echo "🔄 Restarting health dashboard..."
	docker compose restart health-dashboard

restart-db: ## Restart database services
	@echo "🔄 Restarting databases..."
	docker compose restart postgres redis

# === MONITORING ===
logs: ## Show logs from all services
	docker compose logs -f

logs-api: ## Show platform-api logs
	docker compose logs -f platform-api

logs-gateway: ## Show gateway logs
	docker compose logs -f gateway

logs-admin: ## Show admin console logs
	docker compose logs -f admin-console

logs-control-plane: ## Show control plane logs
	docker compose logs -f control-plane

logs-tenant-ui: ## Show tenant UI logs
	docker compose logs -f tenant-ui

logs-dashboard: ## Show health dashboard logs
	docker compose logs -f health-dashboard

logs-db: ## Show database logs
	docker compose logs -f postgres redis

status: ## Show service status
	@echo "📊 BackSaaS Service Status"
	@echo "==========================="
	@docker compose ps
	@echo ""
	@echo "🌐 URLs:"
	@echo "  Platform API:    http://localhost:8080"
	@echo "  Gateway:         http://localhost:8000"
	@echo "  Admin Console:   http://localhost:8000/admin"
	@echo "  Control Plane:   http://localhost:8000/control-plane"
	@echo "  Health Dashboard: http://localhost:8090"

# === BUILDING ===
build: ## Build all service images
	@echo "🔨 Building all images..."
	docker compose build

build-api: ## Build platform-api
	docker compose build platform-api

build-gateway: ## Build gateway
	docker compose build gateway

build-admin: ## Build admin console
	docker compose build admin-console

build-control-plane: ## Build control plane
	docker compose build control-plane

build-tenant-ui: ## Build tenant UI
	docker compose build tenant-ui

# === TESTING ===
test: ## Run complete test suite
	@echo "🚀 Running Complete Test Suite..."
	@echo "================================="
	@echo "📋 Phase 1: Unit Tests + Coverage"
	@$(MAKE) test-coverage
	@echo ""
	@echo "📋 Phase 2: E2E Integration Tests"
	@$(MAKE) test-e2e
	@echo ""
	@echo "📋 Phase 3: Test Report"
	@$(MAKE) test-report
	@echo "================================="
	@echo "✅ Complete Test Suite Finished!"
	@echo "📊 Results: http://localhost:8090"

test-coverage: ## Run unit tests with coverage
	@$(MAKE) -f Makefile.test test-coverage

test-e2e: ## Run E2E integration tests
	@$(MAKE) -f Makefile.test test-e2e

test-quick: ## Quick unit tests (smoke tests)
	@$(MAKE) -f Makefile.test test-smoke

test-report: ## Generate test report
	@echo "📊 Collecting latest test data..."
	@curl -s http://localhost:8090/api/collect -X POST >/dev/null || echo "⚠️  Dashboard collection failed"
	@sleep 2
	@echo "📊 Test Report:"
	@curl -s http://localhost:8090/api/services | jq '{services: keys, avg_coverage: ([.[] | .overall] | add / length | floor)}' 2>/dev/null || echo "Dashboard not available"

# === HEALTH ===
health: ## Check service health
	@echo "🏥 Service Health:"
	@echo -n "Platform API: " && curl -s http://localhost:8080/health | jq -r '.status // "❌ Down"' 2>/dev/null || echo "❌ Down"
	@echo -n "Gateway: " && curl -s http://localhost:8000/health | jq -r '.status // "❌ Down"' 2>/dev/null || echo "❌ Down"
	@echo -n "Health Dashboard: " && curl -s http://localhost:8090/ >/dev/null 2>&1 && echo "✅ Healthy" || echo "❌ Down"
	@echo -n "Database: " && docker compose exec postgres pg_isready -U postgres >/dev/null 2>&1 && echo "✅ Healthy" || echo "❌ Down"

dashboard: ## Open health dashboard
	@echo "📊 BackSaaS Health Dashboard"
	@echo "🌐 Dashboard: http://localhost:8090"
	@echo "💡 Use 'make up' to start all services including dashboard"

admin: ## Open admin console
	@echo "🏗️ BackSaaS Admin Console"
	@echo "🌐 Login: http://localhost:8080/admin/login"
	@echo "👤 Username: admin"
	@echo "🔑 Password: admin123"
	@echo "💡 Use 'make up' to start all services first"

dashboard-update: ## Update dashboard with latest test data
	@echo "🔄 Updating health dashboard with latest data..."
	@curl -s http://localhost:8090/api/collect -X POST >/dev/null && echo "✅ Dashboard updated" || echo "❌ Update failed"

# === DATABASE ===
db-reset: ## Reset database
	@echo "🗄️ Resetting database..."
	docker compose down postgres
	docker volume rm backsaas_postgres_data 2>/dev/null || true
	docker compose up -d postgres
	@echo "⏳ Waiting for database..."
	@sleep 10
	@echo "✅ Database reset"

db-shell: ## Connect to database
	docker compose exec postgres psql -U postgres -d backsaas

# === CLEANUP ===
clean: ## Clean containers and volumes
	@echo "🧹 Cleaning up..."
	docker compose down -v
	docker system prune -f
	@echo "✅ Cleanup complete"

# === DEVELOPMENT HELPERS ===
shell-api: ## Shell into platform-api
	docker compose exec platform-api sh

shell-gateway: ## Shell into gateway
	docker compose exec gateway sh

shell-admin: ## Shell into admin console
	docker compose exec admin-console sh

shell-db: ## Shell into database
	docker compose exec postgres sh

admin: ## Show admin console access info
	@echo "🔐 Admin Console Access"
	@echo "======================="
	@echo "URL: http://localhost:8000/admin"
	@echo "Login: admin@backsaas.dev"
	@echo "Password: admin123"
	@echo ""

control-plane: ## Show control plane access info
	@echo "🎛️ Control Plane Access"
	@echo "======================="
	@echo "URL: http://localhost:8000/control-plane"
	@echo "Schema Designer & Migration Planner"
	@echo ""
# === LEGACY ALIASES ===
dev: up
dev-up: up
dev-down: down
dev-logs: logs
dev-status: status
