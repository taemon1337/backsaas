# BackSaaS — Complete Multi-Tenant SaaS Platform 🚀

A **production-ready, fully functional multi-tenant SaaS platform** with complete user authentication, tenant management, and rich dashboard interfaces.

## ✨ **What's Built & Working**

- ✅ **Complete User Journey**: Landing page → Registration → Tenant creation → Dashboard
- ✅ **JWT Authentication**: Secure login/registration with token storage
- ✅ **Multi-Tenant Architecture**: Full tenant isolation and management
- ✅ **Rich Dashboard UI**: Modern interface with metrics, actions, and management
- ✅ **API Gateway**: Routing, authentication, rate limiting, CORS
- ✅ **Error Handling**: Comprehensive error boundaries and user feedback
- ✅ **Automated Testing**: Complete test suites for all user flows
- ✅ **Docker Deployment**: One-command setup with all services

## 🚀 **Quick Start**

```bash
# Start the entire platform
docker compose up -d

# Access the platform
open http://localhost:8000

# Run tests to verify everything works
./scripts/test-complete-ux.sh
```

## 🌐 **Platform URLs**

- **Main Platform**: http://localhost:8000
- **Admin Console**: http://localhost:8000/admin  
- **Tenant Dashboard**: http://localhost:8000/ui
- **Health Monitoring**: http://localhost:8000/dashboard

## 🧪 **Testing Suite**

```bash
# Complete user flow test
./scripts/test-user-flow.sh

# Full UX validation with content checks  
./scripts/test-complete-ux.sh

# Error handling and security tests
./scripts/test-error-handling.sh
```

## 📊 **System Status**

All core features are **✅ COMPLETE** and **🧪 TESTED**:
- User registration and authentication
- Tenant creation and management  
- Rich dashboard with business metrics
- API gateway with security
- Comprehensive error handling
See per‑folder **README.md** files for details.

## Monorepo Layout

### User Interfaces
- `apps/admin-console` — Next.js platform management UI (admin.backsaas.dev)
- `apps/tenant-ui` — Next.js business interface UI ({tenant}.backsaas.dev)

### Backend Services  
- `services/gateway` — Go API gateway (routing, auth, rate limiting)
- `services/platform-api` — Go platform management API (tenant_id: "system")
- `services/tenant-api` — Go tenant-specific data APIs (dynamic tenant_id)
- `services/migrator` — Go migration worker

### Shared Packages
- `packages/ui` — shadcn‑based design system
- `packages/sdk` — generated OpenAPI client (placeholder)
- `packages/config` — shared configs
- `infra/` — db seed & optional policy bundles

## Architecture Philosophy

BackSaas uses an **event-sourced schema management** approach to handle the complexity of multi-tenancy + hot-reload + migrations:

### Core Principles
- **Single Source of Truth**: Schema registry maintains authoritative schema state
- **Event-Driven Updates**: Schema changes are published as events, not direct cache updates
- **Separation of Concerns**: Hot-reload ≠ Migrations (different event types, different handlers)
- **Tenant Isolation**: Each tenant locked to specific schema version until migration completes

### Event Flow
```
Schema Registry → Event Stream → API Instances
     ↓              ↓              ↓
  Publishes     Redis Streams   Subscribe &
  Events        + Postgres      Update Cache
                LISTEN/NOTIFY
```

### Change Types
- **Compatible Changes** (additive): Hot-reloaded via events
- **Breaking Changes**: Require coordinated migration + API deployment
- **Migration Events**: Handled by dedicated migrator service

### Benefits
- **Atomic Updates**: Schema changes are events, ensuring consistency
- **Replay Capability**: API instances can rebuild state from events  
- **Audit Trail**: Full history of schema changes
- **Graceful Degradation**: API instances restart gracefully on incompatible changes

## Custom Business Logic

BackSaas provides a **Go-native function system** for implementing secure, high-performance business logic within schemas:

### Function Types
- **Validation Functions**: Real-time data validation with predefined Go functions
- **Business Logic Hooks**: before_create, after_update, field_change triggers
- **Computed Fields**: Dynamic field calculation using expression language
- **Event-Driven Workflows**: Async processing triggered by data changes
- **External Integrations**: HTTP calls and webhook functions

### Execution Environment
- **Native Go Performance**: No runtime overhead, compile-time safety
- **Predefined Function Registry**: Curated, secure functions implemented in Go
- **YAML Configuration**: Tenant-configurable rules and parameters
- **Expression Language**: Simple expressions for computed fields and conditions
- **Tenant Isolation**: All functions automatically scoped to tenant data

### Security & Performance
- **Compile-Time Safety**: No arbitrary code execution, all functions predefined
- **High Performance**: Native Go speed, no JavaScript runtime overhead
- **Resource Efficiency**: Lower memory usage, better concurrency
- **Type Safety**: All function parameters validated at compile time

### Self-Hosting Example
The platform itself uses Go functions for user registration, tenant provisioning, schema validation, and RBAC enforcement - demonstrating enterprise-grade performance and security.

## Purpose

- Dogfood architecture: the registry (tenants, schemas, policies, migrations) is managed via the same API pattern used for customer data.
- Event-sourced schema updates with hot-reload for compatible changes only.
- Safe relational evolution via **Expand → Backfill → Contract**.
