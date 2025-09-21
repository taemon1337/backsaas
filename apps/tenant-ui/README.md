# Tenant UI

The **Tenant UI** is the business interface for BackSaas tenants, providing end users with schema-driven forms, dashboards, and workflows tailored to their specific business needs.

## 🎯 Purpose

This interface is designed for **tenant end users** who need to:

- Work with their business data through schema-driven interfaces
- Access custom dashboards and reports
- Manage users within their tenant organization
- Use tenant-specific workflows and business processes
- Experience branded, customized interfaces

## 🏗️ Architecture

### Access & Routing
- **URL Patterns**: 
  - `{tenant}.backsaas.dev` (subdomain routing)
  - `backsaas.dev/app/{tenant}` (path-based routing)
- **Authentication**: Requires valid tenant user credentials
- **Gateway Route**: Handled by API Gateway with tenant resolution
- **Backend**: Communicates with Tenant APIs (dynamic tenant_id)

### Tech Stack
- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS with tenant-customizable themes
- **State Management**: React Query for server state
- **Forms**: Dynamic form generation from schemas
- **Charts**: Recharts for business dashboards
- **Theming**: CSS variables for tenant branding

## 🎨 Design Principles

### Business-Focused UX
- **Schema-Driven Interface**: Forms and views generated from tenant schemas
- **User-Friendly Workflows**: Guided processes with help text and validation
- **Business Context**: Navigation and features aligned with tenant's domain
- **Responsive Design**: Mobile-first design for field workers

### Tenant Branding
- **Custom Themes**: Per-tenant color schemes, logos, and styling
- **White-Label Experience**: Fully branded as tenant's application
- **Flexible Layouts**: Customizable dashboard layouts per tenant
- **Brand Consistency**: Maintains tenant brand throughout experience

## 📱 Key Features

### 1. Schema-Driven Forms
```typescript
// Dynamic form generation
- Auto-generated forms from entity schemas
- Field validation based on schema rules
- Conditional fields and business logic
- File uploads and rich media support
- Multi-step workflows and wizards
```

### 2. Custom Dashboards
```typescript
// Business intelligence for tenants
- Configurable dashboard widgets
- Real-time data visualization
- KPI tracking and alerts
- Custom reports and exports
- Drill-down analytics
```

### 3. Data Management
```typescript
// CRUD operations for business entities
- List views with filtering and sorting
- Detail views with related data
- Bulk operations and batch updates
- Data import/export capabilities
- Audit trails and change history
```

### 4. User Management
```typescript
// Tenant-scoped user administration
- Invite and manage team members
- Role assignment within tenant
- Permission management
- User activity monitoring
- Single sign-on integration
```

### 5. Workflow Engine
```typescript
// Business process automation
- Custom workflow definitions
- Task assignment and tracking
- Approval processes
- Notification management
- Integration triggers
```

## 🎨 Theming & Branding

### Theme System
```typescript
// CSS custom properties for tenant branding
:root {
  --tenant-primary: #3b82f6;
  --tenant-secondary: #64748b;
  --tenant-accent: #f59e0b;
  --tenant-background: #ffffff;
  --tenant-surface: #f8fafc;
  --tenant-text: #1e293b;
}

// Component theming
.btn-primary {
  background-color: var(--tenant-primary);
  color: var(--tenant-text);
}
```

### Brand Configuration
```typescript
// Tenant branding settings
interface TenantBranding {
  logo: string;
  favicon: string;
  colors: {
    primary: string;
    secondary: string;
    accent: string;
  };
  fonts: {
    heading: string;
    body: string;
  };
  customCSS?: string;
}
```

## 🔐 Security & Multi-Tenancy

### Tenant Isolation
```typescript
// Automatic tenant scoping
1. User accesses {tenant}.backsaas.dev
2. Gateway extracts tenant ID from subdomain
3. All API calls automatically scoped to tenant
4. No cross-tenant data access possible
5. UI elements filtered by tenant permissions
```

### Authentication Flow
```typescript
// Tenant-specific authentication
1. User visits tenant subdomain
2. Redirected to tenant-branded login page
3. Credentials validated against tenant user base
4. JWT token includes tenant_id and user roles
5. Session maintained with tenant context
```

### Permission System
```typescript
// Role-based access within tenant
interface TenantUser {
  id: string;
  tenantId: string;
  roles: string[];
  permissions: string[];
  entityAccess: Record<string, 'read' | 'write' | 'admin'>;
}
```

## 🚀 Development Setup

### Prerequisites
```bash
# Ensure these services are running:
- API Gateway (port 8000)
- Tenant API (port 8081)
- PostgreSQL (port 5432)
- Redis (port 6379)
```

### Local Development
```bash
# Install dependencies
cd apps/tenant-ui
npm install

# Set environment variables
cp .env.example .env.local
# Configure API_URL, AUTH_URL, etc.

# Start development server
npm run dev
# Access at http://localhost:3001

# Test with tenant context
# Add "127.0.0.1 acme-corp.localhost" to /etc/hosts
# Access at http://acme-corp.localhost:3001

# Build for production
npm run build
npm start
```

### Docker Development
```bash
# From project root
make dev-tenant-ui

# Or with Docker directly
docker run --rm -it \
  -v $(PWD)/apps/tenant-ui:/app \
  -w /app \
  -p 3001:3001 \
  node:18-alpine \
  npm run dev
```

## 📁 Project Structure

```
apps/tenant-ui/
├── src/
│   ├── app/                 # Next.js App Router
│   │   ├── (auth)/         # Authentication layouts
│   │   ├── dashboard/      # Main dashboard
│   │   ├── entities/       # Dynamic entity pages
│   │   │   └── [entity]/   # Schema-driven CRUD
│   │   ├── workflows/      # Business workflows
│   │   ├── reports/        # Custom reports
│   │   ├── settings/       # Tenant settings
│   │   └── users/          # User management
│   ├── components/         # Reusable UI components
│   │   ├── ui/            # Base UI components
│   │   ├── forms/         # Dynamic form components
│   │   │   ├── SchemaForm.tsx
│   │   │   ├── FieldRenderer.tsx
│   │   │   └── ValidationProvider.tsx
│   │   ├── tables/        # Data table components
│   │   ├── charts/        # Chart components
│   │   ├── layout/        # Layout components
│   │   └── branding/      # Tenant branding components
│   ├── lib/               # Utilities and configurations
│   │   ├── api.ts         # Tenant API client
│   │   ├── auth.ts        # Authentication helpers
│   │   ├── schema.ts      # Schema processing utilities
│   │   ├── theming.ts     # Tenant theming system
│   │   └── utils.ts       # General utilities
│   ├── hooks/             # Custom React hooks
│   │   ├── useSchema.ts   # Schema data hooks
│   │   ├── useTenant.ts   # Tenant context hooks
│   │   └── useAuth.ts     # Authentication hooks
│   └── types/             # TypeScript type definitions
├── public/                # Static assets
├── styles/               # Global styles and themes
├── package.json
├── tailwind.config.js
├── next.config.js
└── README.md
```

## 🔗 Integration Points

### Schema-Driven Components
```typescript
// Dynamic form generation from schema
interface EntitySchema {
  name: string;
  fields: Field[];
  relationships: Relationship[];
  permissions: Permission[];
  functions: Function[];
}

// Auto-generated form component
const SchemaForm: React.FC<{ schema: EntitySchema }> = ({ schema }) => {
  return (
    <Form>
      {schema.fields.map(field => (
        <FieldRenderer key={field.name} field={field} />
      ))}
    </Form>
  );
};
```

### Tenant API Integration
```typescript
// Tenant-scoped API client
const createTenantApiClient = (tenantId: string) => ({
  baseURL: process.env.TENANT_API_URL,
  headers: {
    'X-Tenant-ID': tenantId,
    'Authorization': `Bearer ${getToken()}`
  }
});

// Example API calls
const api = createTenantApiClient(tenant.id);
await api.get('/entities/users');
await api.post('/entities/orders', orderData);
```

### Real-time Updates
```typescript
// WebSocket connection for tenant updates
const wsClient = new WebSocketClient({
  url: process.env.WS_URL,
  auth: getToken(),
  channels: [`tenant.${tenantId}.updates`]
});

// Listen for data changes
wsClient.on('entity.updated', (data) => {
  queryClient.invalidateQueries(['entities', data.entityType]);
});
```

## 🧪 Testing Strategy

### Component Testing
```bash
# Schema-driven component tests
npm run test

# Test form generation
npm run test -- --testNamePattern="SchemaForm"

# Test tenant theming
npm run test -- --testNamePattern="theming"
```

### Multi-Tenant Testing
```bash
# E2E tests with different tenant contexts
npm run test:e2e -- --project=tenant-a
npm run test:e2e -- --project=tenant-b

# Visual regression testing per tenant
npm run test:visual -- --tenant=acme-corp
```

## 🎨 Customization Examples

### Custom Dashboard Layout
```typescript
// Tenant-specific dashboard configuration
interface DashboardConfig {
  layout: 'grid' | 'masonry' | 'flex';
  widgets: Widget[];
  theme: TenantTheme;
}

// Widget configuration
interface Widget {
  type: 'chart' | 'table' | 'metric' | 'custom';
  entity?: string;
  query?: QueryConfig;
  size: 'small' | 'medium' | 'large';
  position: { x: number; y: number };
}
```

### Custom Field Types
```typescript
// Extensible field renderer system
const fieldRenderers = {
  'text': TextFieldRenderer,
  'email': EmailFieldRenderer,
  'phone': PhoneFieldRenderer,
  'address': AddressFieldRenderer,
  'signature': SignatureFieldRenderer,
  'file-upload': FileUploadRenderer,
  // Tenant-specific custom fields
  'product-selector': ProductSelectorRenderer,
  'customer-lookup': CustomerLookupRenderer
};
```

## 🚀 Deployment

### Multi-Tenant Deployment
```dockerfile
# Single deployment serves all tenants
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

FROM node:18-alpine AS runner
WORKDIR /app
COPY --from=builder /app/node_modules ./node_modules
COPY . .
RUN npm run build

# Environment variables for tenant resolution
ENV TENANT_RESOLUTION=subdomain
ENV API_GATEWAY_URL=http://gateway:8000

EXPOSE 3001
CMD ["npm", "start"]
```

### CDN & Asset Management
```typescript
// Tenant-specific asset handling
const getAssetUrl = (path: string, tenantId: string) => {
  return `${CDN_URL}/tenants/${tenantId}/assets/${path}`;
};

// Tenant logo and branding assets
const TenantLogo = ({ tenant }: { tenant: Tenant }) => (
  <img 
    src={getAssetUrl(tenant.branding.logo, tenant.id)}
    alt={`${tenant.name} logo`}
  />
);
```

## 📊 Monitoring & Analytics

### Tenant Usage Analytics
- **Feature Usage**: Track which features are used by each tenant
- **Performance**: Monitor page load times per tenant
- **User Engagement**: Track user activity and session duration
- **Error Tracking**: Tenant-specific error monitoring

### Business Metrics
- **User Adoption**: New user registrations per tenant
- **Feature Utilization**: Most/least used features per tenant
- **Support Requests**: Tenant-specific support ticket patterns
- **Custom Events**: Business-specific event tracking

This tenant UI provides a flexible, branded, and secure interface that adapts to each tenant's specific business needs while maintaining consistent performance and user experience across all tenants.
