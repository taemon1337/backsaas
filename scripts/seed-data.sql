-- BackSaas Seed Data Script
-- This script populates the database with initial development data

\c backsaas;

-- Insert system tenant (for platform administration)
INSERT INTO tenants (tenant_id, name, domain, plan, status, settings) VALUES
('system', 'BackSaas Platform', 'system.backsaas.dev', 'system', 'active', '{"is_system": true}')
ON CONFLICT (tenant_id) DO NOTHING;

-- Insert test tenant for development
INSERT INTO tenants (tenant_id, name, domain, plan, status, settings) VALUES
('test-tenant', 'Test Company', 'test-tenant.backsaas.dev', 'pro', 'active', '{"features": ["api_access", "custom_schemas", "advanced_analytics"]}')
ON CONFLICT (tenant_id) DO NOTHING;

-- Insert demo tenant
INSERT INTO tenants (tenant_id, name, domain, plan, status, settings) VALUES
('acme-corp', 'Acme Corporation', 'acme-corp.backsaas.dev', 'enterprise', 'active', '{"features": ["api_access", "custom_schemas", "advanced_analytics", "white_label"]}')
ON CONFLICT (tenant_id) DO NOTHING;

-- Insert system admin user
INSERT INTO users (user_id, tenant_id, email, password_hash, first_name, last_name, role, status) VALUES
('00000000-0000-0000-0000-000000000001', 'system', 'admin@backsaas.dev', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'System', 'Admin', 'platform_admin', 'active')
ON CONFLICT (tenant_id, email) DO NOTHING;

-- Insert test tenant users
INSERT INTO users (user_id, tenant_id, email, password_hash, first_name, last_name, role, status) VALUES
('00000000-0000-0000-0000-000000000002', 'test-tenant', 'admin@test-tenant.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Test', 'Admin', 'admin', 'active'),
('00000000-0000-0000-0000-000000000003', 'test-tenant', 'user@test-tenant.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Test', 'User', 'user', 'active')
ON CONFLICT (tenant_id, email) DO NOTHING;

-- Insert Acme Corp users
INSERT INTO users (user_id, tenant_id, email, password_hash, first_name, last_name, role, status) VALUES
('00000000-0000-0000-0000-000000000004', 'acme-corp', 'admin@acme-corp.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'John', 'Smith', 'admin', 'active'),
('00000000-0000-0000-0000-000000000005', 'acme-corp', 'sales@acme-corp.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Jane', 'Doe', 'user', 'active'),
('00000000-0000-0000-0000-000000000006', 'acme-corp', 'support@acme-corp.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Bob', 'Johnson', 'user', 'active')
ON CONFLICT (tenant_id, email) DO NOTHING;

-- Insert platform schema (system tenant)
INSERT INTO schemas (schema_id, tenant_id, name, version, schema_definition, status, deployed_at, created_by) VALUES
('00000000-0000-0000-0000-000000000001', 'system', 'platform', 1, '{
  "version": 1,
  "service": {
    "name": "platform",
    "description": "BackSaas Platform Management API"
  },
  "entities": {
    "tenants": {
      "key": "tenant_id",
      "schema": {
        "type": "object",
        "required": ["tenant_id", "name"],
        "properties": {
          "tenant_id": {"type": "string"},
          "name": {"type": "string", "maxLength": 255},
          "domain": {"type": "string", "format": "hostname"},
          "plan": {"type": "string", "enum": ["free", "pro", "enterprise", "system"]},
          "status": {"type": "string", "enum": ["active", "inactive", "suspended"]},
          "settings": {"type": "object"}
        }
      }
    },
    "users": {
      "key": "user_id",
      "schema": {
        "type": "object",
        "required": ["user_id", "email", "tenant_id"],
        "properties": {
          "user_id": {"type": "string"},
          "tenant_id": {"type": "string"},
          "email": {"type": "string", "format": "email"},
          "first_name": {"type": "string", "maxLength": 100},
          "last_name": {"type": "string", "maxLength": 100},
          "role": {"type": "string", "enum": ["platform_admin", "admin", "user"]},
          "status": {"type": "string", "enum": ["active", "inactive", "pending"]}
        }
      }
    }
  }
}', 'deployed', CURRENT_TIMESTAMP, '00000000-0000-0000-0000-000000000001')
ON CONFLICT (tenant_id, name, version) DO NOTHING;

-- Insert CRM schema for test tenant
INSERT INTO schemas (schema_id, tenant_id, name, version, schema_definition, status, deployed_at, created_by) VALUES
('00000000-0000-0000-0000-000000000002', 'test-tenant', 'crm', 1, '{
  "version": 1,
  "service": {
    "name": "sample-crm",
    "description": "Sample CRM system for testing database operations"
  },
  "entities": {
    "contacts": {
      "key": "contact_id",
      "schema": {
        "type": "object",
        "required": ["contact_id", "email", "first_name", "last_name"],
        "properties": {
          "contact_id": {"type": "string"},
          "email": {"type": "string", "format": "email"},
          "first_name": {"type": "string", "minLength": 1, "maxLength": 50},
          "last_name": {"type": "string", "minLength": 1, "maxLength": 50},
          "phone": {"type": "string", "pattern": "^[+]?[1-9]?[0-9]{7,15}$"},
          "company": {"type": "string", "maxLength": 100},
          "status": {"type": "string", "enum": ["lead", "prospect", "customer", "inactive"], "default": "lead"},
          "tags": {"type": "array", "items": {"type": "string"}},
          "metadata": {"type": "object"}
        }
      }
    },
    "companies": {
      "key": "company_id",
      "schema": {
        "type": "object",
        "required": ["company_id", "name"],
        "properties": {
          "company_id": {"type": "string"},
          "name": {"type": "string", "minLength": 1, "maxLength": 100},
          "industry": {"type": "string", "maxLength": 50},
          "size": {"type": "string", "enum": ["startup", "small", "medium", "large", "enterprise"]},
          "website": {"type": "string", "format": "uri"},
          "address": {"type": "object"},
          "annual_revenue": {"type": "number", "minimum": 0},
          "active": {"type": "boolean", "default": true}
        }
      }
    },
    "deals": {
      "key": "deal_id",
      "schema": {
        "type": "object",
        "required": ["deal_id", "title", "contact_id", "amount", "stage"],
        "properties": {
          "deal_id": {"type": "string"},
          "title": {"type": "string", "minLength": 1, "maxLength": 200},
          "contact_id": {"type": "string"},
          "company_id": {"type": "string"},
          "amount": {"type": "number", "minimum": 0},
          "currency": {"type": "string", "enum": ["USD", "EUR", "GBP", "CAD"], "default": "USD"},
          "stage": {"type": "string", "enum": ["prospecting", "qualification", "proposal", "negotiation", "closed_won", "closed_lost"]},
          "probability": {"type": "integer", "minimum": 0, "maximum": 100},
          "expected_close_date": {"type": "string", "format": "date"},
          "notes": {"type": "string"}
        }
      }
    }
  }
}', 'deployed', CURRENT_TIMESTAMP, '00000000-0000-0000-0000-000000000002')
ON CONFLICT (tenant_id, name, version) DO NOTHING;

-- Insert sample API keys for development
INSERT INTO api_keys (key_id, tenant_id, user_id, key_hash, name, permissions, expires_at) VALUES
('00000000-0000-0000-0000-000000000001', 'system', '00000000-0000-0000-0000-000000000001', 'sk_test_system_admin_key_hash', 'System Admin Key', '["platform:admin"]', NULL),
('00000000-0000-0000-0000-000000000002', 'test-tenant', '00000000-0000-0000-0000-000000000002', 'sk_test_tenant_admin_key_hash', 'Test Tenant Admin Key', '["tenant:admin"]', NULL),
('00000000-0000-0000-0000-000000000003', 'acme-corp', '00000000-0000-0000-0000-000000000004', 'sk_test_acme_admin_key_hash', 'Acme Admin Key', '["tenant:admin"]', NULL)
ON CONFLICT (key_hash) DO NOTHING;

-- Insert some sample audit log entries
INSERT INTO audit_log (tenant_id, user_id, action, resource_type, resource_id, details) VALUES
('system', '00000000-0000-0000-0000-000000000001', 'tenant.created', 'tenant', 'test-tenant', '{"name": "Test Company"}'),
('system', '00000000-0000-0000-0000-000000000001', 'tenant.created', 'tenant', 'acme-corp', '{"name": "Acme Corporation"}'),
('test-tenant', '00000000-0000-0000-0000-000000000002', 'schema.deployed', 'schema', 'crm', '{"version": 1}'),
('test-tenant', '00000000-0000-0000-0000-000000000002', 'user.created', 'user', '00000000-0000-0000-0000-000000000003', '{"email": "user@test-tenant.com"}');

-- Create some sample data for the test tenant CRM
-- Note: These will be created by the application when it starts up and creates the tables
