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

### ✅ **True Self-Hosting**
- Platform API uses same schema-driven approach as tenants
- Platform manages itself using `platform.yaml`
- Consistent patterns across all services

### ✅ **Clear Separation of Concerns**
- **Gateway**: Routing, auth, rate limiting
- **Platform API**: Platform management only
- **Tenant API**: Tenant-specific operations only

### ✅ **Independent Scaling**
- Scale gateway for routing capacity
- Scale platform API for tenant provisioning
- Scale tenant APIs per tenant load

### ✅ **Security Isolation**
- Authentication centralized in gateway
- Tenant data completely isolated
- No cross-tenant access possible

## 📋 Implementation Plan

This architecture perfectly supports your insight about the platform acting as a tenant. The **Platform API becomes just another tenant** (with `tenant_id: "system"`) using the same schema-driven patterns, while the **Gateway handles all routing concerns** separately.

Would you like me to start implementing this architecture by creating the gateway service structure?
