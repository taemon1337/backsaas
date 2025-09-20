# services/api — Go Data Plane

Event-driven API server that subscribes to schema changes and serves tenant-specific CRUD endpoints.

## Architecture

### Event-Sourced Schema Management
- **Schema Cache**: In-memory cache of tenant schemas, updated via events
- **Event Subscription**: Listens to Redis Streams + Postgres LISTEN/NOTIFY for schema events
- **Hot-Reload**: Compatible schema changes applied without restart
- **Graceful Restart**: Breaking changes trigger graceful shutdown/restart

### Event Types Handled
- `schema.created` - New tenant schema registered
- `schema.updated.compatible` - Additive changes (new fields, indexes)
- `schema.updated.breaking` - Breaking changes (requires restart)
- `schema.migrated` - Migration completed, switch to new version

### Cache Strategy
```go
type SchemaCache struct {
    tenants map[string]*TenantSchema
    version int64
    mu      sync.RWMutex
}

// Hot-reload compatible changes
func (c *SchemaCache) ApplyCompatibleUpdate(event SchemaEvent)

// Graceful restart for breaking changes  
func (c *SchemaCache) HandleBreakingChange(event SchemaEvent)
```

## Dev (Docker)
```bash
docker build -f Dockerfile.dev -t backsaas-api-dev .
docker run --rm -it -p 8080:8080 --env-file ../../.env -v $PWD:/app backsaas-api-dev
```
