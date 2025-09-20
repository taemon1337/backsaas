# Platform Bootstrap Schema

This directory contains the **bootstrap schema** that defines BackSaas platform entities and RBAC rules.

## Self-Hosting Architecture

BackSaas uses the same schema-driven approach for managing itself as it provides to tenants:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  platform.yaml  │───▶│  Bootstrap       │───▶│  Self-Managing  │
│  (Static)       │    │  (Startup)       │    │  Platform       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Bootstrap Process

1. **Startup**: API service loads `platform.yaml`
2. **Migration**: Creates/updates platform tables in Postgres
3. **Self-Management**: Platform manages its own schema evolution via API

## Key Features

### Rich RBAC System
- **Custom Rules**: `self`, `current_user`, `tenant_member`, etc.
- **Role Hierarchy**: admin → tenant_owner → tenant_admin → tenant_developer → tenant_viewer
- **Field-Level Access**: Users can update limited fields on their own records
- **Context-Aware**: Rules can reference current user, tenant membership, etc.

### Platform Entities
- **users**: Authentication and user management
- **tenants**: Organizations/workspaces
- **tenant_memberships**: User-tenant relationships with roles
- **schemas**: Schema definitions (including this bootstrap schema)
- **migrations**: Migration tracking and status
- **policies**: Custom RBAC policies
- **api_keys**: Programmatic access tokens

### Access Rule Examples

```yaml
# Users can read their own record
read:
  - rule: "self"  # current_user.id = resource.id

# Only tenant admins can modify tenant settings
write:
  - rule: "tenant_admin AND tenant_id = resource.tenant_id"

# Users can update limited fields on their own profile
write:
  - rule: "self AND field IN ['name', 'updated_at']"
```

### Event Integration
- All platform operations publish events (user.created, schema.updated, etc.)
- Platform schema changes trigger the same migration process as tenant schemas
- Real-time updates via event streams

## Files

- `platform.yaml` - Complete platform schema with RBAC

## Implementation Notes

The platform schema demonstrates the full capabilities of the BackSaas system:
- Complex RBAC with custom rules
- Multi-tenant access patterns
- Event-driven updates
- Self-hosting capability

This serves as both the bootstrap configuration and a reference implementation for tenant schemas.

## Platform Business Logic

The platform itself demonstrates the full power of the function system with real business logic:

### User Management Functions
- **validate_user_email**: Email normalization, format validation, uniqueness checking
- **setup_new_user**: Welcome email, event publishing, user onboarding

### Tenant Provisioning Functions  
- **validate_tenant_slug**: Format validation, reserved word checking, uniqueness
- **provision_tenant**: Membership creation, default schema setup, welcome emails

### Schema Management Functions
- **validate_schema_spec**: JSON parsing, entity validation, schema structure checks
- **detect_schema_changes**: Breaking change detection, migration triggering, hot-reload events

### Security Functions
- **generate_api_key**: Secure key generation, hashing, prefix creation

### Self-Hosting Benefits

1. **Dogfooding**: Platform uses its own schema-driven system
2. **Consistency**: Same patterns available to all tenants
3. **Testing**: Platform functions validate the system works at scale
4. **Documentation**: Real examples of complex business logic
5. **Evolution**: Platform can evolve its own schema using migrations

This creates a truly self-hosting architecture where BackSaas manages itself using the same powerful capabilities it provides to tenants.
