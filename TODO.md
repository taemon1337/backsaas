# TODO — Milestones & Tasks

## M0: Scaffolding ✅
- [x] Monorepo structure (pnpm + turbo)
- [x] docker-compose with Postgres, Redis, API, Migrator, Web
- [x] Dev/Prod Dockerfiles for each component
- [x] System bootstrap schema placeholder
- [x] Fix naming consistency (stellar → backsaas)
- [x] Document event-sourced architecture approach

## M0.5: Bootstrap Platform Schema 🚧
- [x] **Platform schema YAML**: Define complete platform entities (tenants, schemas, users, policies, migrations)
- [x] **Rich RBAC rules**: Implement `self`, `current_user`, role-based access patterns
- [ ] **Schema bootstrap loader**: Parse and migrate platform schema at startup
- [ ] **Self-hosting validation**: Ensure platform can manage its own schema evolution
- [ ] **Policy engine integration**: Connect Casbin with schema-defined RBAC rules

## M0.6: Go-Based Function System 🚧
- [x] **Function system documentation**: Complete guide with secure architecture
- [x] **Function entities in platform.yaml**: Add functions, function_executions, function_tests entities
- [x] **Security architecture**: Eliminate direct SQL access, use curated platform functions
- [x] **Go function registry**: Predefined, high-performance Go functions implemented
- [x] **Function implementations**: Validation, security, communication functions created
- [x] **Platform.yaml integration**: Updated to use Go function calls instead of JavaScript
- [ ] **Function execution engine**: Hook triggers, validation, computed fields integration
- [ ] **Expression language**: Simple expressions for computed fields and conditions
- [ ] **Tenant configuration**: YAML-based function configuration per tenant

## M0.7: Multi-Service Architecture 🚧
- [x] **Architecture documentation**: Service separation design completed
- [x] **Generic schema-driven API engine**: Loads any schema (platform.yaml or tenant schemas)
- [x] **Platform API service structure**: Complete service with main.go, tests, Makefile
- [x] **Comprehensive test suite**: Schema loader tests, function tests, integration tests
- [x] **Function system tests**: All predefined Go functions tested (validation, security, communication)
- [ ] **API Gateway service**: Central routing, auth, rate limiting
- [ ] **Database operations implementation**: Complete CRUD operations with tenant scoping
- [ ] **Function execution integration**: Hook triggers, validation, computed fields
- [ ] **Service communication**: Inter-service communication patterns
- [ ] **Docker compose update**: Multi-service development environment

## M1: Event Infrastructure
- [ ] **Redis Streams setup**: Configure streams for schema events
- [ ] **Postgres LISTEN/NOTIFY**: Setup event publishing from registry
- [ ] **Event schema definitions**: Define event types and payloads
- [ ] **Basic event publisher**: Registry publishes schema events
- [ ] **Basic event subscriber**: API service subscribes to events

## M2: Schema Registry Core
- [ ] **System schema bootstrap**: Parse and migrate system tables
- [ ] **Registry CRUD API**: Tenants, schemas, migrations endpoints
- [ ] **Schema validation**: JSON Schema validation on create/update
- [ ] **Event publishing**: Publish schema.created/updated events
- [ ] **Schema versioning**: Track schema versions per tenant

## M3: API Event-Driven Cache
- [ ] **Schema cache implementation**: In-memory tenant schema cache
- [ ] **Event subscription**: Listen to Redis Streams + Postgres NOTIFY
- [ ] **Compatible updates**: Hot-reload additive schema changes
- [ ] **Breaking change handling**: Graceful restart on incompatible changes
- [ ] **CRUD endpoints**: Generate endpoints from cached schemas

## M4: Migration Orchestrator
- [ ] **Migration event listener**: Subscribe to schema.migration.requested
- [ ] **Diff detection**: Compare schema v→v+1 for changes
- [ ] **Expand phase**: Generate and execute additive SQL changes
- [ ] **Backfill phase**: Data migration with custom hooks
- [ ] **Contract phase**: Remove deprecated columns/tables
- [ ] **Status reporting**: Publish migration progress events

## M5: Control Plane UI
- [ ] **NextAuth OIDC**: Authentication setup
- [ ] **Schema Designer**: Monaco YAML editor with validation
- [ ] **Event publisher UI**: Trigger schema updates via events
- [ ] **Migration planner**: Visualize migration steps before execution
- [ ] **Real-time status**: Live updates from migration events

## M6: UI Power-Ups
- [ ] **ERD view**: React Flow diagram synced with YAML editor
- [ ] **Policy Studio**: RBAC policy builder + simulator
- [ ] **Advanced Migration Planner**: Dry-run with SQL preview
- [ ] **Tenant Management**: Pin/unpin schema versions, canary rollouts
- [ ] **Data Browser**: RBAC-aware data explorer with AG Grid

## M7: Tooling & DX
- [ ] OpenAPI generator in API build
- [ ] TypeScript SDK generator (packages/sdk) from OpenAPI
- [ ] Storybook for packages/ui components
- [ ] CI/CD pipelines (lint, build, test, image publish)
- [ ] E2E (Playwright) happy-path tests

## Event Schema Definitions

### Schema Events
```yaml
schema.created:
  tenant_id: string
  schema_id: string
  version: int
  spec: object

schema.updated.compatible:
  tenant_id: string
  schema_id: string
  old_version: int
  new_version: int
  changes: array<AddField|AddIndex>

schema.updated.breaking:
  tenant_id: string
  schema_id: string
  old_version: int
  new_version: int
  changes: array<RemoveField|ChangeType>
```

### Migration Events
```yaml
schema.migration.requested:
  tenant_id: string
  from_version: int
  to_version: int
  migration_id: string

schema.migration.started:
  migration_id: string
  phase: "expand" | "backfill" | "contract"

schema.migration.completed:
  migration_id: string
  tenant_id: string
  new_version: int
```

## Implementation Phases

### Phase 1: Simple Schema Versioning (M1-M2)
- No hot-reload initially
- Basic CRUD with static schemas
- Event infrastructure setup

### Phase 2: Event-Driven Updates (M3)
- Add hot-reload for compatible changes
- Event subscription in API service
- Schema cache implementation

### Phase 3: Migration Orchestration (M4)
- Add migration worker
- Expand/backfill/contract phases
- Breaking change handling

### Phase 4: Advanced UI (M5-M6)
- Rich schema designer
- Migration planning
- Real-time status updates

## Stretch Goals
- [ ] **GraphQL surface**: Shared authN/Z with REST API
- [ ] **Multi-backend adapters**: Mongo/Firestore/File support
- [ ] **Enterprise features**: SSO multi-IdP, org mapping, billing
