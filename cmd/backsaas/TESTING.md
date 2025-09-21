# BackSaaS CLI Testing Commands

The BackSaaS CLI includes comprehensive end-to-end testing capabilities designed to validate platform functionality from a user's perspective. These tests are ideal for post-deployment verification, continuous integration, and platform health monitoring.

## 🎯 Overview

The CLI testing framework provides:
- **Complete Platform Validation**: Tests the entire stack end-to-end
- **Tenant Lifecycle Testing**: Full tenant creation to deletion workflows
- **Real User Scenarios**: Tests mirror actual user interactions
- **CI/CD Integration**: Perfect for automated deployment verification
- **Performance Validation**: Basic load and performance testing
- **Security Verification**: Access controls and data isolation testing

## 📋 Available Test Commands

### Platform Tests
```bash
# Run comprehensive platform tests
backsaas test platform

# Platform tests with custom configuration
backsaas test platform \
  --test-tenant-prefix="ci-test" \
  --timeout=15m \
  --concurrent-tenants=3 \
  --cleanup=true \
  --verbose=true
```

### Tenant Lifecycle Tests
```bash
# Test complete tenant lifecycle
backsaas test tenant-lifecycle

# Tenant lifecycle with custom schema
backsaas test tenant-lifecycle \
  --tenant-name="test-tenant-123" \
  --schema-file="./schemas/test-crm.yaml" \
  --keep-tenant=false \
  --timeout=10m
```

### API Tests
```bash
# Test API functionality
backsaas test api --tenant="existing-tenant"

# API tests with load testing
backsaas test api \
  --tenant="test-tenant" \
  --requests-per-endpoint=50 \
  --load-test=true \
  --request-timeout=30s
```

### Schema Tests
```bash
# Test schema operations
backsaas test schema --tenant="test-tenant"

# Schema tests with migration testing
backsaas test schema \
  --schema-dir="./test-schemas" \
  --tenant="test-tenant" \
  --test-migrations=true
```

## 🚀 Platform Test Suite

The platform test suite (`backsaas test platform`) runs comprehensive tests across all platform components:

### Test Phases

1. **Platform Health Check**
   - API endpoint availability
   - Database connectivity
   - Redis connectivity
   - Service health verification

2. **Authentication & Authorization**
   - Admin authentication
   - Role-based access control
   - JWT token validation
   - Permission verification

3. **Tenant Management**
   - Tenant creation/deletion
   - Tenant configuration
   - Multi-tenant isolation
   - Concurrent tenant operations

4. **Schema Operations**
   - Schema validation
   - Schema deployment
   - Schema updates/migrations
   - Backward compatibility

5. **User Management**
   - User creation/deletion
   - Role assignment
   - Permission testing
   - User authentication

6. **API Operations**
   - CRUD operations
   - Data validation
   - Error handling
   - Rate limiting

7. **Data Consistency**
   - Transaction consistency
   - Data isolation
   - Referential integrity
   - Concurrent access

8. **Performance Validation**
   - Response time testing
   - Load handling
   - Concurrent operations
   - Resource utilization

9. **Cleanup Verification**
   - Resource cleanup
   - Data purging
   - Tenant isolation verification

### Example Output
```
🚀 Starting BackSaaS Platform End-to-End Tests
===============================================

🔧 Test Configuration:
  • Test Prefix: e2e-test
  • Timeout: 10m0s
  • Concurrent Tenants: 1
  • Cleanup: true
  • Verbose: false

📋 Phase 1/9: Platform Health Check
─────────────────────────────────────
🏥 Checking platform health...
  ✓ platform-api/health responding
  ✓ Database connection established
  ✓ Redis connection established
✅ Platform health checks passed
✅ Phase completed in 2.3s

📋 Phase 2/9: Authentication & Authorization
─────────────────────────────────────
🔐 Testing authentication and authorization...
  ✓ Admin authentication working
  ✓ Role-based access control working
✅ Authentication tests passed
✅ Phase completed in 1.8s

... [additional phases] ...

🎯 Platform Test Summary
========================
✅ All 9 test phases passed
⏱️  Total duration: 2m34s
🏢 Tenants tested: 1

🎉 All platform tests passed successfully!
```

## 🏢 Tenant Lifecycle Test Suite

The tenant lifecycle test suite (`backsaas test tenant-lifecycle`) validates complete tenant workflows:

### Lifecycle Phases

1. **Pre-flight Checks**
   - Platform connectivity
   - Authentication verification
   - Schema file validation
   - Tenant name availability

2. **Tenant Creation**
   - Tenant creation with configuration
   - Creation verification
   - Status validation

3. **Initial Configuration**
   - Settings configuration
   - Quota setup
   - Feature enablement

4. **Schema Deployment**
   - Schema validation
   - Deployment process
   - Functionality verification

5. **User Management**
   - User creation (multiple roles)
   - Permission verification
   - Role-based access testing

6. **Data Operations**
   - CRUD operations
   - Data validation
   - Bulk operations
   - Query testing

7. **API Validation**
   - Endpoint testing
   - Error handling
   - Rate limiting
   - Authentication

8. **Schema Updates**
   - Schema migration
   - Data migration
   - Backward compatibility

9. **Backup & Restore**
   - Backup creation
   - Backup verification
   - Restore testing

10. **Performance Testing**
    - Load testing
    - Concurrent operations
    - Large dataset handling

11. **Security Validation**
    - Data isolation
    - Access controls
    - Audit logging

12. **Cleanup & Deletion**
    - Tenant deletion
    - Resource cleanup
    - Verification

### Example Usage Scenarios

#### Post-Deployment Verification
```bash
# Verify platform after deployment
backsaas test platform --timeout=15m --verbose=true

# Test with multiple concurrent tenants
backsaas test platform --concurrent-tenants=5
```

#### CI/CD Pipeline Integration
```bash
#!/bin/bash
# deployment-verification.sh

# Wait for platform to be ready
sleep 30

# Run comprehensive platform tests
if ! backsaas test platform --timeout=10m; then
    echo "❌ Platform tests failed - deployment verification failed"
    exit 1
fi

# Run tenant lifecycle test
if ! backsaas test tenant-lifecycle --timeout=5m; then
    echo "❌ Tenant lifecycle test failed"
    exit 1
fi

echo "✅ Deployment verification completed successfully"
```

#### Development Testing
```bash
# Quick smoke test during development
backsaas test platform --timeout=2m --cleanup=true

# Test specific tenant with custom schema
backsaas test tenant-lifecycle \
  --tenant-name="dev-test-$(date +%s)" \
  --schema-file="./my-schema.yaml" \
  --keep-tenant=true
```

#### Load Testing
```bash
# Test platform under load
backsaas test platform \
  --concurrent-tenants=10 \
  --timeout=30m

# API load testing
backsaas test api \
  --tenant="load-test-tenant" \
  --requests-per-endpoint=100 \
  --load-test=true
```

## 🔧 Configuration Options

### Global Flags
- `--gateway-url`: API Gateway URL (default: http://localhost:8000)
- `--platform-url`: Platform API URL (default: http://localhost:8080)
- `--config`: Configuration file path
- `--verbose`: Enable verbose output

### Platform Test Flags
- `--test-tenant-prefix`: Prefix for test tenant names (default: "e2e-test")
- `--timeout`: Test timeout duration (default: 10m)
- `--cleanup`: Clean up test resources (default: true)
- `--verbose`: Enable verbose test output (default: false)
- `--test-schema`: Path to test schema file
- `--concurrent-tenants`: Number of concurrent tenant tests (default: 1)

### Tenant Lifecycle Test Flags
- `--tenant-name`: Specific tenant name (generates random if empty)
- `--schema-file`: Schema file for testing
- `--keep-tenant`: Keep tenant after test (default: false)
- `--timeout`: Test timeout duration (default: 5m)

### API Test Flags
- `--tenant`: Tenant to run API tests against (required)
- `--requests-per-endpoint`: Requests per endpoint (default: 10)
- `--request-timeout`: Individual request timeout (default: 30s)
- `--load-test`: Enable load testing scenarios (default: false)

### Schema Test Flags
- `--schema-dir`: Directory with test schemas (default: "./schemas")
- `--tenant`: Tenant for schema deployment
- `--test-migrations`: Test schema migrations (default: true)

## 📊 Test Results and Reporting

### Exit Codes
- `0`: All tests passed
- `1`: Tests failed
- `2`: Configuration error
- `3`: Authentication error
- `4`: Timeout error

### Verbose Output
Use `--verbose=true` for detailed test output including:
- Individual test step results
- Timing information
- Resource creation/deletion details
- API request/response details

### Integration with Monitoring
The test commands can be integrated with monitoring systems:

```bash
# Run tests and send results to monitoring
if backsaas test platform --timeout=5m; then
    curl -X POST "https://monitoring.example.com/api/metrics" \
         -d '{"test":"platform","status":"pass","timestamp":"'$(date -u +%s)'"}'
else
    curl -X POST "https://monitoring.example.com/api/metrics" \
         -d '{"test":"platform","status":"fail","timestamp":"'$(date -u +%s)'"}'
fi
```

## 🎯 Best Practices

### Test Environment Setup
1. **Dedicated Test Environment**: Run tests against dedicated test environments
2. **Clean State**: Ensure clean state before running tests
3. **Resource Cleanup**: Always enable cleanup in CI/CD environments
4. **Timeout Configuration**: Set appropriate timeouts for your environment

### CI/CD Integration
1. **Post-Deployment**: Run tests after each deployment
2. **Parallel Execution**: Use concurrent tenant testing for faster feedback
3. **Failure Handling**: Implement proper failure handling and notifications
4. **Test Data**: Use realistic test data that mirrors production scenarios

### Development Workflow
1. **Local Testing**: Run quick tests during development
2. **Schema Validation**: Test schema changes before deployment
3. **Performance Monitoring**: Regular performance validation
4. **Security Testing**: Include security tests in regular workflows

### Monitoring and Alerting
1. **Regular Health Checks**: Schedule regular platform tests
2. **Performance Baselines**: Establish performance baselines
3. **Alert Thresholds**: Set up alerts for test failures
4. **Trend Analysis**: Monitor test execution trends over time

## 🚨 Troubleshooting

### Common Issues

#### Connection Errors
```bash
# Check platform connectivity
backsaas health check

# Test with custom URLs
backsaas test platform \
  --gateway-url="https://api.example.com" \
  --platform-url="https://platform.example.com"
```

#### Authentication Failures
```bash
# Verify authentication configuration
backsaas config show

# Test with explicit authentication
export BACKSAAS_AUTH_TOKEN="your-token"
backsaas test platform
```

#### Timeout Issues
```bash
# Increase timeout for slower environments
backsaas test platform --timeout=30m

# Run individual test phases
backsaas test api --tenant="existing-tenant"
```

#### Resource Cleanup Issues
```bash
# Manual cleanup if needed
backsaas tenant list | grep "e2e-test" | xargs -I {} backsaas tenant delete {}

# Disable cleanup for debugging
backsaas test tenant-lifecycle --keep-tenant=true
```

This CLI testing framework provides comprehensive validation of your BackSaaS platform, ensuring that all components work together correctly from a user's perspective. It's an essential tool for maintaining platform quality and reliability.
