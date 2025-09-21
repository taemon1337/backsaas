# Admin Console UI

The **Admin Console** is the platform management interface for BackSaas, providing system administrators with tools to manage tenants, schemas, users, and monitor the entire platform.

## 🎯 Purpose

This interface is designed for **platform operators** and **system administrators** who need to:

- Manage tenant lifecycle (create, configure, monitor, billing)
- Design and validate schemas using visual tools
- Monitor system health and performance across all tenants
- Manage platform users and permissions
- View analytics and billing data
- Configure platform-wide settings

## 🏗️ Architecture

### Access & Routing
- **URL**: `admin.backsaas.dev`
- **Authentication**: Requires `platform_admin` role
- **Gateway Route**: Handled by API Gateway with elevated permissions
- **Backend**: Communicates with Platform API (tenant_id: "system")

### Tech Stack
- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS + shadcn/ui components
- **State Management**: React Query for server state
- **Charts**: Recharts for analytics dashboards
- **Forms**: React Hook Form with Zod validation

## 🎨 Design Principles

### Admin-Focused UX
- **Dense Information Display**: Tables, charts, system data optimized for power users
- **Bulk Operations**: Multi-select, batch actions for efficiency
- **Advanced Filtering**: Complex queries and data exploration tools
- **System-Centric Navigation**: Organized around platform management tasks

### Visual Identity
- **Dark Theme**: Professional admin interface aesthetic
- **Consistent Branding**: BackSaas platform branding throughout
- **Information Hierarchy**: Clear visual hierarchy for complex data
- **Responsive Design**: Works on desktop and tablet devices

## 📱 Key Features

### 1. Tenant Management
```typescript
// Tenant CRUD operations
- Create new tenants with schema selection
- Configure tenant settings and limits
- Monitor tenant usage and performance
- Manage tenant billing and subscriptions
- Suspend/activate tenant accounts
```

### 2. Schema Designer
```typescript
// Visual schema management
- Drag-and-drop schema builder
- Entity relationship visualization
- Function configuration interface
- Schema validation and testing
- Version management and rollback
```

### 3. System Monitoring
```typescript
// Platform health dashboard
- Real-time system metrics
- API performance monitoring
- Error tracking and alerting
- Resource usage analytics
- Service health checks
```

### 4. User Management
```typescript
// Cross-tenant user administration
- Platform admin user management
- Role and permission assignment
- User activity monitoring
- Security audit logs
- Access control policies
```

### 5. Analytics & Billing
```typescript
// Business intelligence
- Usage analytics across tenants
- Revenue and billing dashboards
- Performance metrics and trends
- Cost analysis and optimization
- Custom reporting tools
```

## 🔐 Security & Access Control

### Authentication Flow
```typescript
// Admin-specific auth requirements
1. User attempts to access admin.backsaas.dev
2. Gateway checks for valid JWT token
3. Token must contain "platform_admin" role
4. MFA verification required for production
5. Session management with elevated privileges
```

### Permission Levels
- **Super Admin**: Full platform access
- **Platform Admin**: Tenant and schema management
- **Support Admin**: Read-only access for support
- **Billing Admin**: Billing and usage data only

## 🚀 Development Setup

### Prerequisites
```bash
# Ensure these services are running:
- API Gateway (port 8000)
- Platform API (port 8080)
- PostgreSQL (port 5432)
- Redis (port 6379)
```

### Local Development
```bash
# Install dependencies
cd apps/admin-console
npm install

# Set environment variables
cp .env.example .env.local
# Configure API_URL, AUTH_URL, etc.

# Start development server
npm run dev
# Access at http://localhost:3000

# Build for production
npm run build
npm start
```

### Docker Development
```bash
# From project root
make dev-admin-console

# Or with Docker directly
docker run --rm -it \
  -v $(PWD)/apps/admin-console:/app \
  -w /app \
  -p 3000:3000 \
  node:18-alpine \
  npm run dev
```

## 📁 Project Structure

```
apps/admin-console/
├── src/
│   ├── app/                 # Next.js App Router
│   │   ├── (auth)/         # Authentication layouts
│   │   ├── dashboard/      # Main dashboard
│   │   ├── tenants/        # Tenant management
│   │   ├── schemas/        # Schema designer
│   │   ├── users/          # User management
│   │   ├── analytics/      # Analytics & billing
│   │   └── settings/       # Platform settings
│   ├── components/         # Reusable UI components
│   │   ├── ui/            # shadcn/ui components
│   │   ├── forms/         # Form components
│   │   ├── charts/        # Chart components
│   │   └── layout/        # Layout components
│   ├── lib/               # Utilities and configurations
│   │   ├── api.ts         # API client setup
│   │   ├── auth.ts        # Authentication helpers
│   │   ├── utils.ts       # General utilities
│   │   └── validations.ts # Form validation schemas
│   └── types/             # TypeScript type definitions
├── public/                # Static assets
├── package.json
├── tailwind.config.js
├── next.config.js
└── README.md
```

## 🔗 Integration Points

### API Communication
```typescript
// Platform API integration
const apiClient = createApiClient({
  baseURL: process.env.PLATFORM_API_URL,
  auth: () => getAdminToken(),
  tenantId: "system" // Always system tenant
});

// Example API calls
await apiClient.tenants.list();
await apiClient.schemas.create(schemaData);
await apiClient.users.updateRole(userId, role);
```

### Real-time Updates
```typescript
// WebSocket connection for live updates
const wsClient = new WebSocketClient({
  url: process.env.WS_URL,
  auth: getAdminToken(),
  channels: ['system.events', 'tenant.status']
});

// Listen for tenant status changes
wsClient.on('tenant.status.changed', (data) => {
  updateTenantStatus(data.tenantId, data.status);
});
```

## 🧪 Testing Strategy

### Unit Tests
```bash
# Component testing with Jest + React Testing Library
npm run test

# Watch mode for development
npm run test:watch

# Coverage report
npm run test:coverage
```

### Integration Tests
```bash
# E2E testing with Playwright
npm run test:e2e

# Visual regression testing
npm run test:visual
```

### Accessibility Testing
```bash
# a11y compliance testing
npm run test:a11y
```

## 🚀 Deployment

### Production Build
```bash
# Build optimized production bundle
npm run build

# Start production server
npm start
```

### Docker Deployment
```dockerfile
# Multi-stage build for production
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

FROM node:18-alpine AS runner
WORKDIR /app
COPY --from=builder /app/node_modules ./node_modules
COPY . .
RUN npm run build

EXPOSE 3000
CMD ["npm", "start"]
```

## 📊 Monitoring & Analytics

### Performance Monitoring
- **Core Web Vitals**: LCP, FID, CLS tracking
- **User Experience**: Page load times, interaction metrics
- **Error Tracking**: Client-side error monitoring
- **Usage Analytics**: Feature usage and user behavior

### Business Metrics
- **Admin Activity**: Login frequency, feature usage
- **Platform Health**: System status dashboard usage
- **Tenant Management**: Tenant creation/management patterns
- **Support Efficiency**: Support admin tool usage

This admin console provides platform operators with a powerful, secure, and efficient interface for managing the entire BackSaas platform while maintaining clear separation from tenant-facing interfaces.
