# services/migrator — Migration Worker

Event-driven migration orchestrator that handles database schema evolution safely.

## Architecture

### Event-Driven Migration Flow
1. **Listen**: Subscribe to `schema.migration.requested` events
2. **Plan**: Generate expand/backfill/contract steps
3. **Execute**: Apply migration phases sequentially
4. **Report**: Publish status events back to registry

### Migration Phases
```
Phase 1: EXPAND
├─ Add new tables/columns (non-breaking)
├─ Create new indexes concurrently  
└─ Publish: schema.migration.expanded

Phase 2: BACKFILL  
├─ Migrate existing data to new structure
├─ Run custom backfill hooks
└─ Publish: schema.migration.backfilled

Phase 3: CONTRACT
├─ Remove old columns/tables
├─ Update constraints
└─ Publish: schema.migration.completed
```

### Event Types Published
- `schema.migration.started` - Migration began
- `schema.migration.expanded` - Expand phase completed
- `schema.migration.backfilled` - Backfill phase completed  
- `schema.migration.completed` - Full migration success
- `schema.migration.failed` - Migration failed (with rollback)

### Safety Features
- **Concurrent-safe operations**: Uses `ADD COLUMN IF NOT EXISTS`, etc.
- **Rollback capability**: Each phase can be reverted
- **Progress tracking**: Detailed status reporting
- **Timeout handling**: Prevents stuck migrations

## Dev (Docker)
```bash
docker build -f Dockerfile.dev -t backsaas-migrator-dev .
docker run --rm -it --env-file ../../.env -v $PWD:/app backsaas-migrator-dev
```
