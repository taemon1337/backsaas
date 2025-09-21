#!/bin/sh
# BackSaaS Test Orchestrator
# Coordinates execution of all tests across services

set -e

echo "🚀 BackSaaS Test Orchestrator Starting..."
echo "=========================================="

# Create test results directory structure
mkdir -p /test-results/{unit,integration,coverage,reports}

# Function to run tests in a service container
run_service_tests() {
    local service=$1
    local test_type=$2
    local test_command=$3
    
    echo "🧪 Running $test_type tests for $service..."
    
    # Execute tests in the service container
    if docker exec backsaas-${service}-tests sh -c "$test_command"; then
        echo "✅ $service $test_type tests passed"
        echo "PASS" > /test-results/${test_type}/${service}.status
    else
        echo "❌ $service $test_type tests failed"
        echo "FAIL" > /test-results/${test_type}/${service}.status
        return 1
    fi
}

# Function to generate test report
generate_report() {
    echo "📊 Generating test report..."
    
    cat > /test-results/index.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>BackSaaS Test Results</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; }
        .service { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .pass { background: #d4edda; border-color: #c3e6cb; }
        .fail { background: #f8d7da; border-color: #f5c6cb; }
        .timestamp { color: #666; font-size: 0.9em; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🧪 BackSaaS Test Results</h1>
        <p class="timestamp">Generated: $(date)</p>
    </div>
EOF

    # Add service results
    for service in platform-api gateway api cli; do
        if [ -f "/test-results/unit/${service}.status" ]; then
            status=$(cat /test-results/unit/${service}.status)
            class=$([ "$status" = "PASS" ] && echo "pass" || echo "fail")
            echo "    <div class=\"service $class\">" >> /test-results/index.html
            echo "        <h3>$service Service</h3>" >> /test-results/index.html
            echo "        <p>Unit Tests: $status</p>" >> /test-results/index.html
            echo "    </div>" >> /test-results/index.html
        fi
    done

    echo "</body></html>" >> /test-results/index.html
    echo "📊 Test report generated at /test-results/index.html"
}

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 10

# Install dependencies if needed
apk add --no-cache curl jq

# Run unit tests for each service
echo "🔬 Starting Unit Tests..."
echo "========================"

run_service_tests "platform-api" "unit" "cd /app && go test -v -coverprofile=/test-results/coverage/platform-api.out ./... > /test-results/unit/platform-api.log 2>&1"
run_service_tests "gateway" "unit" "cd /app && go test -v -coverprofile=/test-results/coverage/gateway.out ./... > /test-results/unit/gateway.log 2>&1"
run_service_tests "api" "unit" "cd /app && go test -v -coverprofile=/test-results/coverage/api.out ./... > /test-results/unit/api.log 2>&1"
run_service_tests "cli" "unit" "cd /app && go test -v ./internal/cli/... > /test-results/unit/cli.log 2>&1"

# Run integration tests
echo "🔗 Starting Integration Tests..."
echo "================================"

run_service_tests "platform-api" "integration" "cd /app && go test -v ./tests/integration/... > /test-results/integration/platform-api.log 2>&1"

# Generate coverage reports
echo "📈 Generating Coverage Reports..."
echo "================================="

for service in platform-api gateway api; do
    if [ -f "/test-results/coverage/${service}.out" ]; then
        echo "📊 Generating HTML coverage for $service..."
        docker exec backsaas-${service}-tests sh -c "cd /app && go tool cover -html=/test-results/coverage/${service}.out -o /test-results/coverage/${service}.html"
    fi
done

# Generate final report
generate_report

# Summary
echo ""
echo "🎯 Test Summary"
echo "==============="

total_tests=0
passed_tests=0

for service in platform-api gateway api cli; do
    if [ -f "/test-results/unit/${service}.status" ]; then
        total_tests=$((total_tests + 1))
        status=$(cat /test-results/unit/${service}.status)
        if [ "$status" = "PASS" ]; then
            passed_tests=$((passed_tests + 1))
            echo "✅ $service: PASSED"
        else
            echo "❌ $service: FAILED"
        fi
    fi
done

echo ""
echo "📊 Results: $passed_tests/$total_tests tests passed"

if [ $passed_tests -eq $total_tests ]; then
    echo "🎉 All tests passed!"
    exit 0
else
    echo "💥 Some tests failed!"
    exit 1
fi
