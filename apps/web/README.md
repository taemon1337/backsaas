# apps/web — Control Plane UI (Next.js + shadcn/ui)

Schema management interface that publishes events to orchestrate the data plane.

## Architecture

### Event Publishing
- **Schema Designer**: YAML editor with real-time validation
- **Migration Planner**: Visualizes expand/backfill/contract steps
- **Event Publisher**: Publishes schema events to Redis Streams

### Key Features
- **Schema Validation**: Server-side validation before publishing events
- **Migration Preview**: Dry-run migrations to preview SQL changes
- **Tenant Management**: Pin/unpin schema versions, canary rollouts
- **Real-time Status**: Live updates from migration events

## Dev
```bash
pnpm i
pnpm dev
```
or via Docker (recommended for consistency):
```bash
docker build -f Dockerfile.dev -t backsaas-web-dev .
docker run --rm -it -p 3000:3000 -v $PWD:/app backsaas-web-dev
```

## Notes
- Uses App Router.
- Auth wired for OIDC via NextAuth (placeholders).
- Bring your own shadcn components progressively.
