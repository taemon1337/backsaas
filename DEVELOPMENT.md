# BackSaas Development Guide

This guide covers the complete development environment setup and workflow for BackSaas.

## 🚀 Quick Start

```bash
# 1. Check prerequisites
make check-tools

# 2. Start development environment
make dev-up

# 3. Check status
make dev-status

# 4. View logs
make dev-logs
```

## 📋 Prerequisites

- **Docker** (required)
- **Docker Compose** (required)
- **Make** (recommended)

**No local installations required** - Go, Node.js, PostgreSQL, Redis all run in containers!

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │    │  Platform API   │    │ Test Tenant API │
│   Port: 8000    │────│   Port: 8080    │    │   Port: 8081    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌─────────────────┐    ┌─────────────────┐
         │   PostgreSQL    │    │     Redis       │
         │   Port: 5432    │    │   Port: 6379    │
         └─────────────────┘    └─────────────────┘
```

## 🛠️ Development Services

### Core Services
- **Platform API** (`http://localhost:8080`) - System tenant management
- **API Gateway** (`http://localhost:8000`) - Routing and authentication
- **Test Tenant API** (`http://localhost:8081`) - Sample CRM for testing
- **PostgreSQL** (`localhost:5432`) - Primary database
- **Redis** (`localhost:6379`) - Caching and sessions

### Development Tools
- **Adminer** (`http://localhost:8082`) - Database administration
- **Prometheus** (`http://localhost:9090`) - Metrics collection (optional)
- **Grafana** (`http://localhost:3001`) - Metrics visualization (optional)

## 🔧 Development Commands

### Environment Management
```bash
make dev-up          # Start all services
make dev-down        # Stop all services
make dev-status      # Show service status
make dev-logs        # Show logs from all services
make dev-monitoring  # Start with monitoring stack
make dev-db-only     # Start only database services
```

### Building and Testing
```bash
make build           # Build all services
make test            # Run all tests
make clean           # Clean build artifacts
```

### Individual Services
```bash
make build-platform-api    # Build Platform API
make test-platform-api     # Test Platform API
make build-gateway         # Build API Gateway
make test-gateway          # Test API Gateway
```

## 🗄️ Database Management

### Connection Details
- **Host**: `localhost:5432`
- **Username**: `postgres`
- **Password**: `postgres`
- **Database**: `backsaas`

### Database Structure
```sql
-- Platform tables (automatically created)
tenants         -- Tenant information
users           -- Platform and tenant users
schemas         -- Schema definitions
api_keys        -- API authentication keys
audit_log       -- Audit trail

-- Tenant-specific tables (created dynamically)
contacts        -- CRM contacts (test-tenant)
companies       -- CRM companies (test-tenant)
deals           -- CRM deals (test-tenant)
```

### Access Database
```bash
# Via Adminer (Web UI)
open http://localhost:8082

# Via CLI
docker exec -it backsaas-postgres psql -U postgres -d backsaas
```

## 🧪 Testing the APIs

### Health Checks
```bash
curl http://localhost:8080/health  # Platform API
curl http://localhost:8000/health  # API Gateway
curl http://localhost:8081/health  # Test Tenant API
```

### Platform API (System Tenant)
```bash
# List tenants
curl http://localhost:8080/api/tenants

# Create tenant
curl -X POST http://localhost:8080/api/tenants \
  -H "Content-Type: application/json" \
  -d '{"tenant_id": "new-tenant", "name": "New Tenant"}'

# List users
curl http://localhost:8080/api/users
```

### Test Tenant API (CRM)
```bash
# List contacts
curl http://localhost:8081/api/contacts

# Create contact
curl -X POST http://localhost:8081/api/contacts \
  -H "Content-Type: application/json" \
  -d '{
    "contact_id": "contact-1",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "status": "lead"
  }'

# Get contact
curl http://localhost:8081/api/contacts/contact-1

# List companies
curl http://localhost:8081/api/companies

# List deals
curl http://localhost:8081/api/deals
```

### API Gateway (Routing)
```bash
# Platform routes (system tenant)
curl http://localhost:8000/platform/api/tenants

# Tenant routes (test-tenant)
curl http://localhost:8000/tenant/test-tenant/api/contacts
```

## 🔍 CLI Tools

### BackSaas CLI
```bash
# Build CLI
cd cmd/backsaas && make build

# Health check
./bin/backsaas health check

# Dashboard (real-time monitoring)
./bin/backsaas dashboard

# Tenant management
./bin/backsaas tenant list
./bin/backsaas tenant create "New Company"

# Schema operations
./bin/backsaas schema validate ./testdata/sample-crm.yaml
./bin/backsaas schema deploy ./testdata/sample-crm.yaml --tenant=test-tenant
```

## 📊 Monitoring and Debugging

### View Logs
```bash
# All services
make dev-logs

# Specific service
docker-compose logs -f platform-api
docker-compose logs -f api-gateway
docker-compose logs -f postgres
```

### Metrics (with monitoring profile)
```bash
# Start with monitoring
make dev-monitoring

# Access Prometheus
open http://localhost:9090

# Access Grafana (admin/admin)
open http://localhost:3001
```

### Debug Database
```bash
# Connect to database
docker exec -it backsaas-postgres psql -U postgres -d backsaas

# Check tables
\dt

# Query tenants
SELECT * FROM tenants;

# Query users
SELECT * FROM users;

# Check tenant-specific data
SELECT * FROM contacts WHERE tenant_id = 'test-tenant';
```

## 🔄 Development Workflow

### 1. Start Development Environment
```bash
make dev-up
```

### 2. Make Code Changes
Edit files in your IDE - changes are automatically reflected via volume mounts.

### 3. Test Changes
```bash
# Run tests
make test

# Test specific API
curl http://localhost:8081/api/contacts
```

### 4. View Logs
```bash
make dev-logs
```

### 5. Stop Environment
```bash
make dev-down
```

## 🐛 Troubleshooting

### Services Won't Start
```bash
# Check Docker
docker --version
docker-compose --version

# Check ports
lsof -i :8080  # Platform API
lsof -i :8000  # API Gateway
lsof -i :5432  # PostgreSQL

# Restart everything
make dev-down
make dev-up
```

### Database Issues
```bash
# Reset database
make dev-down
docker volume rm backsaas_postgres_data
make dev-up
```

### Build Issues
```bash
# Clean and rebuild
make clean
make build
```

### Permission Issues
```bash
# Fix Docker permissions
sudo chown -R $USER:$USER .
```

## 📁 Project Structure

```
backsaas/
├── docker-compose.yml          # Main development environment
├── Makefile                    # Development commands
├── DEVELOPMENT.md              # This file
├── scripts/
│   ├── init-db.sql            # Database initialization
│   └── seed-data.sql          # Sample data
├── schemas/
│   └── platform.yaml          # Platform schema
├── config/
│   └── prometheus.yml         # Monitoring configuration
├── services/
│   ├── platform-api/          # Platform API service
│   ├── api-gateway/           # API Gateway service
│   └── migrator/              # Database migrations
├── cmd/
│   └── backsaas/              # CLI tool
└── apps/
    ├── admin-console/         # Admin UI (future)
    └── tenant-ui/             # Tenant UI (future)
```

## 🎯 Next Steps

1. **Test the APIs** - Use the curl examples above
2. **Try the CLI** - Build and test the CLI tools
3. **Explore the Database** - Use Adminer to browse data
4. **Monitor Services** - Use the dashboard and logs
5. **Start Building** - Begin developing new features!

## 💡 Tips

- **Use make commands** - They handle Docker complexity
- **Check logs frequently** - `make dev-logs` shows everything
- **Test incrementally** - Use curl to test API changes
- **Monitor resources** - `docker stats` shows container usage
- **Clean regularly** - `make clean` removes build artifacts

## 🆘 Getting Help

- Check service logs: `make dev-logs`
- Verify service status: `make dev-status`
- Test health endpoints: `curl http://localhost:8080/health`
- Reset environment: `make dev-down && make dev-up`

Happy coding! 🚀
