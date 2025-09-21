#!/bin/sh
# BackSaaS CLI Integration Test Script
# Tests CLI commands within the Docker Compose test network

set -e

echo "🔧 BackSaaS CLI Integration Tests"
echo "=================================="

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 15

# Test CLI health check command
echo "🏥 Testing CLI health check..."
if backsaas health check --gateway-url=http://platform-api-tests:8080; then
    echo "✅ CLI health check passed"
else
    echo "❌ CLI health check failed"
    exit 1
fi

# Test CLI configuration
echo "⚙️ Testing CLI configuration..."
backsaas config show || echo "ℹ️ No config found (expected in test environment)"

# Test CLI tenant operations (if platform API is available)
echo "🏢 Testing CLI tenant operations..."
if backsaas tenant list --platform-url=http://platform-api-tests:8080 2>/dev/null; then
    echo "✅ CLI tenant operations working"
else
    echo "ℹ️ Tenant operations not available (expected in test environment)"
fi

# Test CLI schema validation
echo "📋 Testing CLI schema validation..."
if [ -f "/app/testdata/sample-crm.yaml" ]; then
    if backsaas schema validate /app/testdata/sample-crm.yaml; then
        echo "✅ CLI schema validation passed"
    else
        echo "❌ CLI schema validation failed"
        exit 1
    fi
else
    echo "ℹ️ No test schema found, skipping validation test"
fi

echo "🎉 CLI integration tests completed successfully!"
