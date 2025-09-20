# BackSaas — Backend Schema-Driven Data API Platform

A platform that turns **declarative schemas** into a **policy‑enforced, multi‑tenant data API**.

- **Control Plane UI**: Next.js + shadcn/ui for schema/policy design, migration planning, and tenant ops.
- **Data Plane API**: Go server that hot‑reloads tenant schemas from a central registry (Postgres) and serves a best‑practice REST API to pluggable backends.
- **Migrator**: Go worker that performs expand/backfill/contract database migrations in response to schema updates.

## Quickstart (Docker)

```bash
cp .env.example .env
docker compose up --build
# UI: http://localhost:3000
# API: http://localhost:8080
# Postgres: localhost:5432 (postgres/postgres) DB: backsaas
# Redis: localhost:6379
```
See per‑folder **README.md** files for details.

## Monorepo Layout

- `apps/web` — Next.js control‑plane UI
- `services/api` — Go data‑plane API server
- `services/migrator` — Go migration worker
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
