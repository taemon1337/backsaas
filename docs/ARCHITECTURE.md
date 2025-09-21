# BackSaas Multi-Tenant Architecture

BackSaas uses a **layered, service-oriented architecture** with clear separation between routing, platform management, tenant data operations, and user interfaces.

## 🏗️ Service Architecture

### Overview
```
┌─────────────────────────────────────────────────────────────┐
│                    API Gateway (Port 8000)                  │
│  Routes: /admin/*, /app/*, /{tenant}.api.backsaas.dev      │
└─────────────────────────────────────────────────────────────┘
                              │
                ┌─────────────┴─────────────┐
                │                           │
┌───────────────▼────────────┐   ┌─────────▼──────────────┐
│     Admin Console UI        │   │    Tenant UI           │
│     (Port 3000)            │   │    (Port 3001)         │
│                            │   │                        │
│ • Platform Management      │   │ • Business Workflows   │
│ • Tenant Administration    │   │ • Schema-driven Forms  │
│ • System Monitoring        │   │ • Custom Dashboards    │
│ • Schema Designer          │   │ • Tenant Branding      │
│ • Analytics & Billing      │   │ • User Management      │
└────────────────────────────┘   └────────────────────────┘
                │                           │
                └─────────────┬─────────────┘
                              │
┌─────────────────────────────▼─────────────────────────────┐
│                Platform API (Port 8080)                   │
│              + Tenant APIs (Port 8081+)                   │
└───────────────────────────────────────────────────────────┘
```

### 1. API Gateway (`services/gateway`)
**Purpose**: Central routing, authentication, and request handling

```
┌─────────────────────────────────────────┐
│              API Gateway                │
│  ┌─────────────┐  ┌─────────────────┐   │
│  │   Router    │  │  Authentication │   │
│  │ • Routing   │  │ • JWT tokens    │   │
│  │ • CORS      │  │ • Sessions      │   │
│  │ • Rate      │  │ • OAuth         │   │
│  │   limiting  │  │ • API keys      │   │
│  └─────────────┘  └─────────────────┘   │
└─────────────────────────────────────────┘
```

**Routing Rules:**
```yaml
routes:
  # Admin Console UI
  - host: "admin.backsaas.dev"
    target: "admin-console:3000"
    auth: { required_roles: ["platform_admin"] }
    
  # Tenant UI (by subdomain)
  - host: "*.backsaas.dev"
    target: "tenant-ui:3001"
    auth: { required: true }
    
  # Tenant UI (by path)
  - path: "/app/*"
    target: "tenant-ui:3001"
    auth: { required: true }
    
  # Platform API
  - path: "/api/platform/*"
    target: "platform-api:8080"
    
  # Tenant APIs
  - path: "/api/tenants/{slug}/*"
    target: "tenant-api-{slug}:8080"
    
  # Authentication
  - path: "/auth/*"
    target: "auth-service:8080"
```

### 2. Platform API (`services/platform-api`)
**Purpose**: Platform management using self-hosted schema (tenant_id: "system")

```
┌─────────────────────────────────────────┐
│            Platform API                 │
│         (tenant_id: "system")           │
│  ┌─────────────┐  ┌─────────────────┐   │
│  │ platform.   │  │   Functions     │   │
│  │ yaml        │  │ • User mgmt     │   │
│  │ • Users     │  │ • Tenant        │   │
│  │ • Tenants   │  │   provisioning  │   │
│  │ • Schemas   │  │ • Schema        │   │
│  │ • Functions │  │   validation    │   │
│  └─────────────┘  └─────────────────┘   │
└─────────────────────────────────────────┘
```

### 3. Admin Console UI (`apps/admin-console`)
**Purpose**: Platform administration and tenant management interface

```
┌─────────────────────────────────────────┐
│           Admin Console UI              │
│         (admin.backsaas.dev)            │
│  ┌─────────────┐  ┌─────────────────┐   │
│  │ Platform    │  │   Management    │   │
│  │ Management  │  │ • Tenant CRUD   │   │
│  │ • Tenants   │  │ • User mgmt     │   │
│  │ • Schemas   │  │ • Billing       │   │
│  │ • Users     │  │ • Analytics     │   │
│  │ • Billing   │  │ • Monitoring    │   │
│  └─────────────┘  └─────────────────┘   │
└─────────────────────────────────────────┘
```

**Tech Stack:**
- Next.js 14 (App Router)
- TypeScript + Tailwind CSS
- React Query for API state
- Recharts for analytics
- shadcn/ui components

### 4. Tenant UI (`apps/tenant-ui`)
**Purpose**: Schema-driven business interface for tenant users

```
┌─────────────────────────────────────────┐
│             Tenant UI                   │
│      ({tenant}.backsaas.dev)            │
│  ┌─────────────┐  ┌─────────────────┐   │
│  │ Dynamic     │  │   Business      │   │
│  │ Interface   │  │ • Dashboards    │   │
│  │ • Forms     │  │ • Reports       │   │
│  │ • Tables    │  │ • Workflows     │   │
│  │ • Charts    │  │ • Custom views  │   │
│  │ • Branding  │  │ • User mgmt     │   │
│  └─────────────┘  └─────────────────┘   │
└─────────────────────────────────────────┘
```

**Tech Stack:**
- Next.js 14 (App Router)
- TypeScript + Tailwind CSS
- React Query for API state
- Dynamic form generation
- Tenant-customizable themes

### 5. Tenant API (`services/tenant-api`)
**Purpose**: Schema-driven data operations for individual tenants

```
┌─────────────────────────────────────────┐
│             Tenant API                  │
│        (tenant_id: dynamic)             │
│  ┌─────────────┐  ┌─────────────────┐   │
│  │ Dynamic     │  │   Functions     │   │
│  │ Schema      │  │ • Validation    │   │
│  │ • Custom    │  │ • Hooks         │   │
│  │   entities  │  │ • Computed      │   │
│  │ • Business  │  │   fields        │   │
│  │   rules     │  │ • Workflows     │   │
│  └─────────────┘  └─────────────────┘   │
└─────────────────────────────────────────┘
```

## 🔄 Request Flow Examples

### Admin Console Access
```
GET admin.backsaas.dev/tenants
     ↓
Gateway: Authentication (platform_admin role required)
     ↓
Admin Console UI (React app)
     ↓
API calls to Platform API
     ↓
Return tenant management interface
```

### Tenant UI Access
```
GET acme-corp.backsaas.dev/dashboard
     ↓
Gateway: Authentication & tenant resolution
     ↓
Tenant UI (React app with tenant branding)
     ↓
API calls to Tenant API (tenant_id: "acme-corp")
     ↓
Return tenant-specific dashboard
```

### Platform Management Request
```
POST /api/platform/tenants
     ↓
Gateway: Authentication & routing
     ↓
Platform API (tenant_id: "system")
     ↓
Execute platform functions (validate_tenant_slug, provision_tenant)
     ↓
Return response
```

### Tenant Data Request
```
GET /api/tenants/acme-corp/users
     ↓
Gateway: Authentication & tenant resolution
     ↓
Tenant API (tenant_id: "acme-corp")
     ↓
Load schema from registry → Execute tenant functions
     ↓
Return response
```

## 🚀 Benefits

### ✅ **Separated User Experiences**
- **Admin Console**: Platform operators get specialized management tools
- **Tenant UI**: End users get business-focused, branded interfaces
- **Clear role separation**: No confusion between platform and tenant operations

### ✅ **True Self-Hosting**
- Platform API uses same schema-driven approach as tenants
- Platform manages itself using `platform.yaml`
- Consistent patterns across all services

### ✅ **Clear Separation of Concerns**
- **Gateway**: Routing, auth, rate limiting
- **Admin Console**: Platform management interface
- **Tenant UI**: Business workflow interface
- **Platform API**: Platform management logic
- **Tenant API**: Tenant-specific operations

### ✅ **Independent Development & Scaling**
- UI teams can work independently on admin vs tenant experiences
- Scale gateway for routing capacity
- Scale UIs based on user load patterns
- Scale APIs per operational requirements

### ✅ **Security & Compliance**
- Authentication centralized in gateway with role-based routing
- Admin operations isolated from tenant operations
- Tenant data completely isolated with no cross-tenant access
- Audit trails separated by user type

### ✅ **Customization & Branding**
- **Admin Console**: Consistent platform branding and UX
- **Tenant UI**: Per-tenant branding, themes, and customization
- **Independent deployment**: Update admin features without affecting tenants

## 📋 Implementation Status

### ✅ Completed (M0.7)
- **API Gateway**: Complete with routing, auth, rate limiting
- **Platform API**: Schema-driven platform management
- **Generic API Engine**: Reusable for both platform and tenant APIs

### 🚧 Next Steps (M1)
- **Admin Console UI**: Platform management interface
- **Tenant UI**: Schema-driven business interface
- **Authentication Service**: JWT token management
- **Event Infrastructure**: Real-time updates and notifications

This architecture creates a **true multi-tenant SaaS platform** where platform operators and tenant users have completely different, optimized experiences while sharing the same underlying schema-driven infrastructure.
