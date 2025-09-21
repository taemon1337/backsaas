# BackSaaS Testing Guide

This document describes the comprehensive testing infrastructure for BackSaaS, designed to run all tests within Docker Compose networks without requiring local Go installations.

## 🎯 Overview

The BackSaaS testing infrastructure provides:
- **Isolated Test Environment**: Separate test database and Redis instances
- **Comprehensive Coverage**: Unit, integration, and end-to-end tests
- **Centralized Orchestration**: Single command to run all tests
- **Real-time Reporting**: Web-based test results and coverage reports
- **Docker-First Approach**: No local dependencies required

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Test Environment                         │
├─────────────────────────────────────────────────────────────┤
│  Test Database (postgres:5433)  │  Test Redis (redis:6380)  │
├─────────────────────────────────────────────────────────────┤
│  Platform API Tests  │  Gateway Tests  │  API Tests  │ CLI  │
├─────────────────────────────────────────────────────────────┤
│              Test Orchestrator (Coordinator)                │
├─────────────────────────────────────────────────────────────┤
│           Test Results Server (nginx:8888)                  │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose installed
- No local Go installation required

### Run All Tests
```bash
# Using the enhanced testing Makefile
make -f Makefile.test test-all

# Or using Docker Compose directly
docker compose -f docker-compose.test.yml up --abort-on-container-exit test-orchestrator
```

### View Results
Open http://localhost:8888 to view test results, coverage reports, and logs.

## 📋 Available Commands

### Basic Testing Commands

```bash
# Setup test environment
make -f Makefile.test test-setup

# Run all tests (unit + integration)
make -f Makefile.test test-all

# Run only unit tests
make -f Makefile.test test-unit

# Run only integration tests
make -f Makefile.test test-integration

# Clean test environment
make -f Makefile.test test-clean
```

### Service-Specific Testing

```bash
# Test specific service
make -f Makefile.test test-service SERVICE=platform-api
make -f Makefile.test test-service SERVICE=gateway
make -f Makefile.test test-service SERVICE=api
make -f Makefile.test test-service SERVICE=cli

# Run specific test
make -f Makefile.test test-run SERVICE=platform-api TEST=TestDatabaseOperations
```

### Coverage and Reporting

```bash
# Generate coverage reports
make -f Makefile.test test-coverage

# Start test results server
make -f Makefile.test test-results-server

# Show test status
make -f Makefile.test test-status
```

### Advanced Testing

```bash
# Run tests with race detection
make -f Makefile.test test-race

# Run benchmark tests
make -f Makefile.test test-benchmark

# Quick smoke tests
make -f Makefile.test test-smoke

# Watch mode (requires entr)
make -f Makefile.test test-watch
```

## 🔧 Test Environment Details

### Test Services

#### Test Database (`test-postgres`)
- **Port**: 5433 (isolated from development)
- **Database**: `backsaas_test`
- **Credentials**: postgres/postgres
- **Features**: Pre-configured test schemas and cleanup functions

#### Test Redis (`test-redis`)
- **Port**: 6380 (isolated from development)
- **Purpose**: Caching and session testing
- **Data**: Isolated from development Redis

#### Service Test Containers
Each service has its own test container with:
- Go 1.21 environment
- Service-specific dependencies
- Access to test database and Redis
- Volume mounts for live code updates

### Test Orchestrator
Coordinates test execution across all services:
- Runs tests in parallel where possible
- Generates unified test reports
- Handles test result aggregation
- Provides comprehensive logging

### Test Results Server
Nginx-based server providing:
- **Main Dashboard**: http://localhost:8888
- **Coverage Reports**: http://localhost:8888/coverage/
- **Test Logs**: http://localhost:8888/unit/ and http://localhost:8888/integration/
- **Real-time Updates**: Auto-refresh capabilities

## 📁 Test Structure

```
backsaas/
├── docker-compose.test.yml          # Test environment definition
├── Makefile.test                    # Enhanced testing commands
├── TESTING.md                       # This documentation
├── scripts/
│   ├── test-orchestrator.sh         # Test coordination script
│   ├── init-test-db.sql            # Test database setup
│   └── nginx-test-results.conf     # Test results server config
├── test-results/                    # Generated test artifacts
│   ├── unit/                       # Unit test logs and results
│   ├── integration/                # Integration test logs
│   ├── coverage/                   # Coverage reports (HTML/text)
│   └── reports/                    # Aggregated reports
└── services/
    ├── platform-api/
    │   ├── Dockerfile.test         # Test container definition
    │   └── tests/integration/      # Integration tests
    ├── gateway/
    │   └── Dockerfile.test
    ├── api/
    │   └── Dockerfile.test
    └── cmd/backsaas/
        └── Dockerfile.test
```

## 🧪 Writing Tests

### Unit Tests
Place unit tests alongside your code with `_test.go` suffix:

```go
// services/platform-api/internal/api/database_test.go
func TestDatabaseOperations(t *testing.T) {
    // Test uses TEST_DATABASE_URL environment variable
    dbURL := os.Getenv("TEST_DATABASE_URL")
    // ... test implementation
}
```

### Integration Tests
Place integration tests in dedicated directories:

```go
// services/platform-api/tests/integration/field_mapping_test.go
func TestFieldMappingIntegration(t *testing.T) {
    // Integration test with real database
    // Uses Docker Compose network services
}
```

### Test Environment Variables
Tests automatically receive:
- `TEST_DATABASE_URL`: Connection to test database
- `REDIS_URL`: Connection to test Redis
- `GO_ENV=test`: Environment indicator
- `LOG_LEVEL=debug`: Enhanced logging

## 🔍 Debugging Tests

### View Logs
```bash
# All test logs
make -f Makefile.test test-logs

# Specific service logs
docker compose -f docker-compose.test.yml logs platform-api-tests

# Test orchestrator logs
docker compose -f docker-compose.test.yml logs test-orchestrator
```

### Interactive Debugging
```bash
# Access test container
docker compose -f docker-compose.test.yml exec platform-api-tests sh

# Run tests manually
docker compose -f docker-compose.test.yml exec platform-api-tests go test -v ./internal/api/...
```

### Database Inspection
```bash
# Connect to test database
docker compose -f docker-compose.test.yml exec test-postgres psql -U postgres -d backsaas_test

# Clean test data
docker compose -f docker-compose.test.yml exec test-postgres psql -U postgres -d backsaas_test -c "SELECT testing.clean_test_data();"
```

## 🚀 CI/CD Integration

### GitHub Actions Example
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Tests
        run: make -f Makefile.test test-all
      - name: Upload Coverage
        uses: actions/upload-artifact@v3
        with:
          name: coverage-reports
          path: test-results/coverage/
```

### Local Development Workflow
```bash
# 1. Setup test environment once
make -f Makefile.test test-setup

# 2. During development, run specific tests
make -f Makefile.test test-service SERVICE=platform-api

# 3. Before committing, run all tests
make -f Makefile.test test-all

# 4. Clean up when done
make -f Makefile.test test-clean
```

## 🎯 Best Practices

### Test Organization
1. **Unit Tests**: Test individual functions/methods in isolation
2. **Integration Tests**: Test service interactions with real dependencies
3. **End-to-End Tests**: Test complete user workflows

### Test Data Management
1. Use the test database for all database-dependent tests
2. Clean up test data between test runs
3. Use meaningful test data that reflects real scenarios

### Performance Considerations
1. Run unit tests in parallel where possible
2. Use test database transactions for isolation
3. Mock external dependencies in unit tests

### Debugging Guidelines
1. Use descriptive test names and error messages
2. Log important test state for debugging
3. Use the test results server for comprehensive analysis

## 🔧 Customization

### Adding New Services
1. Create `Dockerfile.test` in your service directory
2. Add service to `docker-compose.test.yml`
3. Update test orchestrator script
4. Add service-specific test commands to `Makefile.test`

### Custom Test Profiles
You can create custom test profiles by:
1. Extending `docker-compose.test.yml`
2. Adding new test orchestrator scripts
3. Creating specialized Makefile targets

### Environment-Specific Testing
```bash
# Test against different database versions
POSTGRES_VERSION=14 make -f Makefile.test test-all

# Test with different Go versions
GO_VERSION=1.22 make -f Makefile.test test-all
```

## 📊 Monitoring and Metrics

### Test Metrics
The test infrastructure tracks:
- Test execution time per service
- Test success/failure rates
- Code coverage percentages
- Test result trends

### Performance Monitoring
- Database query performance in tests
- Memory usage during test execution
- Test execution parallelization efficiency

## 🆘 Troubleshooting

### Common Issues

#### Port Conflicts
```bash
# Check for port conflicts
netstat -tulpn | grep -E ':(5433|6380|8888)'

# Use different ports if needed
TEST_POSTGRES_PORT=5434 make -f Makefile.test test-all
```

#### Database Connection Issues
```bash
# Verify test database is running
docker compose -f docker-compose.test.yml ps test-postgres

# Check database logs
docker compose -f docker-compose.test.yml logs test-postgres
```

#### Test Container Build Issues
```bash
# Rebuild test containers
docker compose -f docker-compose.test.yml build --no-cache

# Check for Go module issues
docker compose -f docker-compose.test.yml exec platform-api-tests go mod tidy
```

### Getting Help
1. Check test logs: `make -f Makefile.test test-logs`
2. Verify configuration: `make -f Makefile.test test-validate`
3. Review test status: `make -f Makefile.test test-status`
4. Access test results: http://localhost:8888

This testing infrastructure ensures that all BackSaaS tests run consistently in Docker containers, providing reliable and reproducible test execution across different development environments.
