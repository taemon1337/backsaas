# BackSaas CLI

A powerful command-line interface for managing the BackSaas platform. This CLI provides comprehensive tools for platform administration, tenant management, schema operations, and system monitoring.

## 🎯 Purpose

The BackSaas CLI is designed for:
- **Platform Administrators**: Manage tenants, users, and system configuration
- **DevOps Engineers**: Bootstrap, monitor, and troubleshoot the platform
- **Developers**: Validate schemas, test APIs, and debug issues
- **Support Teams**: Investigate issues and manage user accounts

## 🚀 Quick Start

### Installation

```bash
# Build from source
cd cmd/backsaas
make build

# The binary will be available at bin/backsaas
./bin/backsaas --help

# Or install to GOPATH/bin
make install
backsaas --help
```

### Docker Build

```bash
# Build using Docker (consistent with project pattern)
make docker

# Or run in Docker development environment
make dev-docker
```

### Initial Setup

```bash
# 1. Initialize configuration
backsaas config init

# 2. Check service health
backsaas health check

# 3. Bootstrap the platform
backsaas bootstrap --admin-email=admin@example.com

# 4. Create your first tenant
backsaas tenant create "My Company"
```

## 📋 Commands Overview

### System Management

```bash
# Real-time monitoring dashboard
backsaas dashboard                       # Live platform dashboard (like 'top')
backsaas dashboard --refresh=5           # Custom refresh interval
backsaas dashboard --compact             # Compact display mode

# Health monitoring
backsaas health check                    # Check all services
backsaas health check --service=gateway  # Check specific service

# Platform bootstrap
backsaas bootstrap --admin-email=admin@example.com
backsaas bootstrap --admin-email=admin@example.com --yes  # Skip confirmations

# Configuration management
backsaas config show                     # Show current config
backsaas config init                     # Initialize config file
backsaas config set gateway_url http://localhost:8000
backsaas config validate                 # Validate config and connectivity
```

### Tenant Management

```bash
# List tenants
backsaas tenant list                     # List all tenants
backsaas tenant list --json              # JSON output

# Create tenants
backsaas tenant create "Acme Corp"       # Create new tenant
backsaas tenant create "Acme Corp" --schema=./crm.yaml  # With initial schema
backsaas tenant create "Acme Corp" --plan=pro --domain=acme.example.com

# Tenant details
backsaas tenant show acme-corp           # Show tenant details
backsaas tenant delete acme-corp         # Delete tenant (with confirmation)
backsaas tenant delete acme-corp --force # Force delete without confirmation
```

### Schema Management

```bash
# Schema validation
backsaas schema validate ./schema.yaml   # Validate schema file
backsaas schema validate ./schemas/*.yaml # Validate multiple schemas

# Schema deployment
backsaas schema deploy ./schema.yaml --tenant=acme-corp
backsaas schema deploy ./schema.yaml --tenant=acme-corp --dry-run
backsaas schema deploy ./schema.yaml --tenant=acme-corp --force

# Schema operations
backsaas schema list --tenant=acme-corp  # List tenant schemas
backsaas schema diff old.yaml new.yaml   # Compare schema files
```

### User Management

```bash
# List users
backsaas user list --tenant=acme-corp    # List users for tenant
backsaas user list --all-tenants         # List all users (admin only)

# Create users
backsaas user create john@acme-corp.com --tenant=acme-corp --role=admin
backsaas user create jane@acme-corp.com --tenant=acme-corp --role=user --password=secret

# User operations
backsaas user show user-123              # Show user details
backsaas user roles user-123             # Show user roles
backsaas user roles user-123 --add=admin --remove=user  # Modify roles
backsaas user delete user-123            # Delete user
```

## ⚙️ Configuration

### Configuration File

The CLI uses a YAML configuration file located at `~/.backsaas.yaml`:

```yaml
gateway_url: "http://localhost:8000"
platform_url: "http://localhost:8080"
auth_token: "your-jwt-token"
default_tenant: "acme-corp"
verbose: false
format: "table"  # table, json, yaml
```

### Environment Variables

All configuration can be overridden with environment variables:

```bash
export BACKSAAS_GATEWAY_URL="https://api.backsaas.dev"
export BACKSAAS_PLATFORM_URL="https://platform.backsaas.dev"
export BACKSAAS_AUTH_TOKEN="your-jwt-token"
export BACKSAAS_DEFAULT_TENANT="acme-corp"
export BACKSAAS_VERBOSE="true"
```

### Global Flags

```bash
--config string        # Config file path (default: ~/.backsaas.yaml)
--gateway-url string   # API Gateway URL (default: http://localhost:8000)
--platform-url string # Platform API URL (default: http://localhost:8080)
--verbose, -v          # Verbose output
```

## 🔧 Development

### Project Structure

```
cmd/backsaas/
├── main.go                 # CLI entry point
├── internal/cli/           # CLI implementation
│   ├── root.go            # Root command and configuration
│   ├── health.go          # Health check commands
│   ├── bootstrap.go       # Platform bootstrap
│   ├── tenant.go          # Tenant management
│   ├── schema.go          # Schema operations
│   ├── user.go            # User management
│   └── config.go          # Configuration management
├── go.mod                 # Go module definition
├── Makefile              # Build automation
└── README.md             # This file
```

### Building

```bash
# Local build
make build                 # Build binary to bin/backsaas
make test                  # Run tests
make clean                 # Clean build artifacts

# Docker build (consistent with project)
make docker               # Build using Docker
make test-docker          # Test using Docker

# Cross-platform builds
make release              # Build for multiple platforms
```

### Testing

```bash
# Run all tests
make test

# Test specific functionality
go test -v ./internal/cli/...

# Test with Docker
make test-docker

# Manual testing
make dev                  # Build and run sample commands
```

## 🎨 Output Formats

The CLI supports multiple output formats:

### Table Format (Default)

```bash
backsaas tenant list
```

```
ID         | Name           | Domain                    | Plan | Status    | Users | Schemas | Created
-----------|----------------|---------------------------|------|-----------|-------|---------|----------
system     | BackSaas Platform | system.backsaas.dev    | system | ✅ active | 1     | 1       | 2024-01-01
acme-corp  | Acme Corporation | acme-corp.backsaas.dev  | pro    | ✅ active | 25    | 3       | 2024-01-15
```

### JSON Format

```bash
backsaas tenant list --json
```

```json
[
  {
    "id": "system",
    "name": "BackSaas Platform",
    "slug": "system",
    "plan": "system",
    "status": "active",
    "created_at": "2024-01-01T00:00:00Z",
    "user_count": 1,
    "schema_count": 1
  }
]
```

## 🔐 Authentication

### JWT Token Authentication

```bash
# Set authentication token
backsaas config set auth_token "your-jwt-token"

# Or use environment variable
export BACKSAAS_AUTH_TOKEN="your-jwt-token"

# Token is automatically included in API requests
backsaas tenant list
```

### Bootstrap Admin User

```bash
# Create initial admin user during bootstrap
backsaas bootstrap --admin-email=admin@example.com

# This creates a platform_admin user that can manage the entire system
```

## 🚨 Error Handling

The CLI provides clear error messages and exit codes:

```bash
# Exit codes:
# 0 - Success
# 1 - General error
# 2 - Configuration error
# 3 - Authentication error
# 4 - API error

# Example error handling in scripts:
if ! backsaas health check; then
    echo "Services are not healthy, aborting deployment"
    exit 1
fi
```

## 📊 Monitoring & Debugging

### Health Checks

```bash
# Check all services
backsaas health check

# Check specific service
backsaas health check --service=gateway
backsaas health check --service=platform-api

# Set custom timeout
backsaas health check --timeout=30
```

### Verbose Output

```bash
# Enable verbose output for debugging
backsaas --verbose tenant create "Test Tenant"

# Or set globally
backsaas config set verbose true
```

### Configuration Validation

```bash
# Validate configuration and connectivity
backsaas config validate

# This checks:
# - Configuration file exists and is valid
# - URLs are properly formatted
# - Services are accessible
# - Authentication is working
```

## 🔄 Integration with CI/CD

### Automated Scripts

```bash
#!/bin/bash
# deployment-script.sh

set -e  # Exit on any error

# Validate configuration
backsaas config validate

# Check system health
backsaas health check

# Deploy new schema
backsaas schema validate ./new-schema.yaml
backsaas schema deploy ./new-schema.yaml --tenant=production --dry-run
backsaas schema deploy ./new-schema.yaml --tenant=production

echo "Deployment completed successfully"
```

### Docker Integration

```dockerfile
# Use CLI in Docker containers
FROM golang:1.21-alpine AS builder
COPY cmd/backsaas /app
WORKDIR /app
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/bin/backsaas /usr/local/bin/
CMD ["backsaas"]
```

## 🎯 Use Cases

### Platform Bootstrap

```bash
# Complete platform setup
backsaas config init
backsaas bootstrap --admin-email=admin@company.com
backsaas tenant create "Production" --plan=enterprise
backsaas schema deploy ./production-schema.yaml --tenant=production
```

### Development Workflow

```bash
# Validate schema during development
backsaas schema validate ./schema.yaml

# Deploy to development tenant
backsaas schema deploy ./schema.yaml --tenant=dev --dry-run
backsaas schema deploy ./schema.yaml --tenant=dev

# Test with sample user
backsaas user create test@dev.com --tenant=dev --role=admin
```

### Production Monitoring

```bash
# Regular health checks
backsaas health check

# Monitor tenant status
backsaas tenant list

# Check user activity
backsaas user list --all-tenants
```

## 📊 Real-Time Dashboard

The CLI includes a powerful real-time dashboard similar to the `top` command but designed for BackSaas platform monitoring.

### Dashboard Features

```bash
# Start the dashboard
backsaas dashboard

# Custom refresh rate
backsaas dashboard --refresh=5

# Compact mode for smaller terminals
backsaas dashboard --compact
```

### Dashboard Sections

#### 🏥 Service Health
- Real-time health status of all services
- Response times and error indicators
- Overall system health summary

#### 💻 System Stats
- Platform uptime and version info
- Memory and CPU usage
- Active connection count
- Environment information

#### 🏢 Tenant Stats
- Active vs total tenants and users
- Schema deployment count
- Recent activity feed

#### 📊 Request Stats
- Requests per second (RPS)
- Average response times
- Error rates with color coding
- Top API endpoints by usage

### Dashboard Controls

- **Ctrl+C**: Exit dashboard
- **--refresh**: Set update interval (default: 2 seconds)
- **--compact**: Reduce information density
- **Auto-refresh**: Continuously updates without user input

### Troubleshooting

```bash
# Debug connectivity issues
backsaas config validate

# Check specific tenant
backsaas tenant show problematic-tenant

# Verify schema deployment
backsaas schema list --tenant=problematic-tenant

# Monitor in real-time
backsaas dashboard --refresh=1
```

This CLI provides a comprehensive interface for managing the BackSaas platform, enabling efficient administration, monitoring, and troubleshooting workflows with real-time visibility.
