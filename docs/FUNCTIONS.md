# Custom Business Logic Functions

BackSaas provides a powerful, declarative function system that allows you to embed custom business logic directly into your schemas. Functions are executed in a secure JavaScript sandbox with full access to database, HTTP, email, and event publishing capabilities.

## Function Types

### 1. Validation Functions
Real-time validation during API operations:

```yaml
entities:
  users:
    schema:
      properties:
        email: { type: string, format: email }
    functions:
      validate_email:
        type: validation
        trigger: "before_create,before_update"
        field: "email"
        code: |
          async function validate(value, record, context) {
            // Use secure platform function for email validation
            const isValid = await context.platform.validateEmail(value);
            if (!isValid) {
              throw new ValidationError('Invalid email format');
            }
            
            // Custom domain validation
            const allowedDomains = ['company.com', 'partner.com'];
            const domain = value.split('@')[1];
            
            if (!allowedDomains.includes(domain)) {
              throw new ValidationError('Email domain not allowed');
            }
            
            // Check uniqueness using secure data API
            const existing = await context.data.findMany('users', {
              where: { 
                email: { eq: value },
                id: { ne: record.id || '' }
              }
            });
            
            if (existing.length > 0) {
              throw new ValidationError('Email already exists');
            }
            
            return value.toLowerCase(); // Normalize
          }
```

### 2. Business Logic Hooks
Execute custom logic at specific lifecycle events:

```yaml
entities:
  tenants:
    functions:
      setup_tenant:
        type: hook
        trigger: "after_create"
        async: true
        code: |
          async function setup(record, context) {
            // Create default schema for new tenant using secure data API
            const defaultSchema = {
              tenant_id: record.id,
              name: 'default',
              version: 1,
              spec: {
                entities: {
                  users: {
                    key: 'id',
                    schema: {
                      type: 'object',
                      required: ['id', 'email'],
                      properties: {
                        id: { type: 'string', format: 'uuid' },
                        email: { type: 'string', format: 'email' },
                        name: { type: 'string' }
                      }
                    }
                  }
                }
              }
            };
            
            // Validate schema using platform function
            const validation = await context.platform.validateSchema(defaultSchema.spec);
            if (!validation.valid) {
              throw new ValidationError('Invalid default schema: ' + validation.errors.join(', '));
            }
            
            // Create schema record using secure API
            await context.data.create('schemas', {
              tenant_id: defaultSchema.tenant_id,
              name: defaultSchema.name,
              version: defaultSchema.version,
              spec: JSON.stringify(defaultSchema.spec),
              status: 'active'
            });
            
            // Send welcome email using platform function
            await context.platform.sendEmail('tenant_welcome', record.owner_email, {
              tenant_name: record.name,
              tenant_slug: record.slug
            });
            
            // Publish event
            await context.events.publish('tenant.provisioned', {
              tenant_id: record.id,
              owner_id: record.owner_id
            });
          }
```

### 3. Computed Fields
Dynamic field calculation:

```yaml
entities:
  users:
    schema:
      properties:
        # ... other fields
    computed_fields:
      full_name:
        dependencies: ["first_name", "last_name"]
        code: |
          function compute(record, context) {
            return `${record.first_name} ${record.last_name}`.trim();
          }
      
      tenant_role:
        code: |
          async function compute(record, context) {
            const memberships = await context.data.findMany('tenant_memberships', {
              where: {
                user_id: { eq: record.id },
                tenant_id: { eq: context.tenant_id },
                status: { eq: 'active' }
              },
              limit: 1
            });
            return memberships.length > 0 ? memberships[0].role : 'none';
          }
```

### 4. Event-Driven Workflows
Async processing triggered by data changes:

```yaml
entities:
  schemas:
    functions:
      validate_schema_change:
        type: workflow
        trigger: "before_update"
        condition: "field_changed('spec')"
        code: |
          async function validate(record, context) {
            // Parse and validate schema
            try {
              const spec = JSON.parse(record.spec);
              
              // Check for breaking changes
              const oldSpec = context.query('SELECT spec FROM schemas WHERE id = ?', [record.id])[0];
              const breakingChanges = detectBreakingChanges(oldSpec.spec, spec);
              
              if (breakingChanges.length > 0) {
                // Require migration
                await context.query(
                  'INSERT INTO migrations (tenant_id, schema_id, from_version, to_version, status) VALUES (?, ?, ?, ?, ?)',
                  [record.tenant_id, record.id, record.version, record.version + 1, 'pending']
                );
                
                // Publish migration event
                await context.events.publish('schema.migration.required', {
                  tenant_id: record.tenant_id,
                  schema_id: record.id,
                  breaking_changes: breakingChanges
                });
              }
              
            } catch (error) {
              throw new ValidationError(`Invalid schema: ${error.message}`);
            }
          }
```

### 5. External Integrations
Call external services for complex logic:

```yaml
entities:
  api_keys:
    functions:
      generate_key:
        type: hook
        trigger: "before_create"
        code: |
          async function generate(record, context) {
            // Generate secure API key
            const response = await context.http.post('https://crypto-service.internal/generate', {
              tenant_id: record.tenant_id,
              permissions: record.permissions
            });
            
            record.key_hash = response.hash;
            record.key_prefix = response.prefix;
            
            return record;
          }
      
      check_rate_limit:
        type: external
        endpoint: "https://rate-limiter.internal/check"
        timeout: "2s"
        fallback: "allow"
        trigger: "before_api_call"
```

## Secure Function Context API

Functions receive a **secure, curated context** with predefined platform functions instead of direct SQL access:

```javascript
interface FunctionContext {
  // Secure data access via predefined platform functions
  data: {
    // Entity operations (automatically tenant-scoped)
    findById(entity: string, id: string): Promise<Record>;
    findMany(entity: string, filters: FilterOptions): Promise<Record[]>;
    create(entity: string, data: object): Promise<Record>;
    update(entity: string, id: string, data: object): Promise<Record>;
    delete(entity: string, id: string): Promise<boolean>;
    
    // Aggregation functions
    count(entity: string, filters?: FilterOptions): Promise<number>;
    sum(entity: string, field: string, filters?: FilterOptions): Promise<number>;
    avg(entity: string, field: string, filters?: FilterOptions): Promise<number>;
    
    // Relationship queries
    getRelated(entity: string, id: string, relation: string): Promise<Record[]>;
    
    // Search and pagination
    search(entity: string, query: string, options?: SearchOptions): Promise<SearchResult>;
    paginate(entity: string, filters: FilterOptions, page: number, limit: number): Promise<PaginatedResult>;
  };
  
  // Predefined platform functions (curated and secure)
  platform: {
    // User management
    validateEmail(email: string): Promise<boolean>;
    hashPassword(password: string): Promise<string>;
    generateApiKey(): Promise<string>;
    
    // Tenant operations
    getTenantSettings(tenant_id?: string): Promise<TenantSettings>;
    updateTenantUsage(metric: string, value: number): Promise<void>;
    
    // Schema operations
    validateSchema(spec: object): Promise<ValidationResult>;
    getSchemaVersion(tenant_id: string): Promise<number>;
    
    // Notification functions
    sendEmail(template: string, to: string, data: object): Promise<void>;
    sendWebhook(url: string, payload: object): Promise<void>;
    
    // Utility functions
    formatCurrency(amount: number, currency: string): string;
    parseDate(dateString: string): Date;
    generateSlug(text: string): string;
  };
  
  // HTTP client (restricted to allowed domains)
  http: {
    get(url: string, options?: RequestOptions): Promise<Response>;
    post(url: string, data: any, options?: RequestOptions): Promise<Response>;
  };
  
  // Event publishing
  events: {
    publish(event: string, data: any): Promise<void>;
    schedule(event: string, data: any, delay: string): Promise<void>;
  };
  
  // Current context (read-only)
  readonly user: User;
  readonly tenant_id: string;
  readonly request_id: string;
  readonly entity: string;  // Current entity being operated on
  readonly operation: 'create' | 'update' | 'delete' | 'read';
  
  // Logging
  log: {
    info(message: string, data?: any): void;
    warn(message: string, data?: any): void;
    error(message: string, error?: Error): void;
  };
}

// Secure filter options with validation
interface FilterOptions {
  where?: {
    [field: string]: any | {
      eq?: any;
      ne?: any;
      gt?: number;
      gte?: number;
      lt?: number;
      lte?: number;
      in?: any[];
      contains?: string;
      startsWith?: string;
      endsWith?: string;
    };
  };
  orderBy?: {
    field: string;
    direction: 'asc' | 'desc';
  }[];
  limit?: number;  // Max 1000
}
```

## Security Architecture

### Secure Function Design
BackSaas uses a **curated platform function approach** instead of direct SQL access to eliminate security risks:

#### ✅ Security Benefits
- **No SQL Injection**: Tenants cannot write arbitrary SQL queries
- **Automatic Tenant Isolation**: All data operations are automatically scoped to tenant
- **Validated Operations**: All data access goes through validated, secure platform functions
- **Resource Control**: Built-in limits on query complexity and execution time
- **Audit Trail**: All function calls are logged and monitored

#### ✅ Performance Benefits  
- **Optimized Queries**: Platform functions use optimized, indexed queries
- **Connection Pooling**: Efficient database connection management
- **Query Planning**: Pre-planned execution paths for common operations
- **Resource Limits**: Automatic limits prevent expensive operations

#### ❌ What Tenants Cannot Do
```javascript
// ❌ Direct SQL access (security risk)
context.query('SELECT * FROM users WHERE tenant_id = ?', [tenant_id]);

// ❌ Cross-tenant data access
context.query('SELECT * FROM other_tenant_data');

// ❌ Schema inspection
context.query('SELECT * FROM information_schema.tables');

// ❌ Expensive operations
context.query('SELECT * FROM large_table ORDER BY random()');
```

#### ✅ What Tenants Can Do
```javascript
// ✅ Secure, tenant-scoped data access
const users = await context.data.findMany('users', {
  where: { status: { eq: 'active' } },
  limit: 100
});

// ✅ Curated platform functions
const isValid = await context.platform.validateEmail(email);
await context.platform.sendEmail('welcome', email, data);

// ✅ Safe aggregations with limits
const count = await context.data.count('orders', {
  where: { created_at: { gte: startDate } }
});
```

### Function Execution Sandbox

#### JavaScript Runtime Security
- **V8 Isolate**: Each function runs in isolated JavaScript context
- **Memory Limits**: Maximum 128MB per function execution
- **CPU Limits**: Maximum 5 seconds CPU time
- **No File System Access**: Functions cannot read/write files
- **No Network Access**: Except through controlled HTTP client

#### Data Access Security
- **Automatic Tenant Scoping**: All queries automatically filtered by tenant_id
- **Field-Level Permissions**: RBAC rules applied to all data operations
- **Query Validation**: All filters validated against entity schema
- **Result Limits**: Maximum 1000 records per query

## Security & Sandboxing

### Resource Limits
```yaml
function_limits:
  max_execution_time: "30s"
  max_memory: "128MB"
  max_cpu_time: "5s"
  max_http_requests: 10
  max_db_queries: 50
  max_email_sends: 5
  allowed_domains: 
    - "api.company.com"
    - "webhook.stripe.com"
    - "*.internal"
```

### Sandbox Features
- **Memory Isolation**: Each function runs in isolated V8 context
- **Network Restrictions**: Only allowed domains accessible
- **Database Scoping**: Automatic tenant_id filtering
- **Execution Timeouts**: Prevent infinite loops
- **Error Handling**: Graceful error capture and logging

## Error Handling

```javascript
// Built-in error types
throw new ValidationError('Field is required');
throw new AuthorizationError('Access denied');
throw new BusinessLogicError('Insufficient inventory');
throw new ExternalServiceError('Payment service unavailable');

// Error context is automatically captured
try {
  await context.http.post('https://external-api.com/endpoint', data);
} catch (error) {
  context.log.error('External API call failed', { 
    url: 'https://external-api.com/endpoint',
    error: error.message,
    tenant_id: context.tenant_id,
    user_id: context.user.id
  });
  throw new ExternalServiceError('Unable to process request');
}
```

## Testing & Debugging

### Function Testing
```yaml
functions:
  validate_email:
    # ... function definition
    tests:
      - name: "valid_email"
        input: { email: "user@company.com" }
        expect: "success"
      
      - name: "invalid_domain"
        input: { email: "user@invalid.com" }
        expect: "ValidationError: Email domain not allowed"
      
      - name: "duplicate_email"
        setup: |
          INSERT INTO users (email) VALUES ('existing@company.com');
        input: { email: "existing@company.com" }
        expect: "ValidationError: Email already exists"
```

### Debug Mode
```javascript
// Debug logging automatically available
function validate(value, record, context) {
  context.log.info('Validating email', { 
    email: value, 
    user_id: record.id 
  });
  
  // ... validation logic
  
  context.log.info('Email validation passed');
  return value;
}
```

## Performance Considerations

### Function Caching
- **Code Compilation**: JavaScript functions compiled once per schema version
- **Context Reuse**: Database connections and HTTP clients pooled
- **Result Caching**: Computed fields cached with dependency tracking

### Async Execution
- **Hook Functions**: Run asynchronously to avoid blocking API responses
- **Workflow Functions**: Queued for background processing
- **External Calls**: Timeout and retry logic built-in

### Monitoring
- **Execution Metrics**: Duration, memory usage, success/failure rates
- **Error Tracking**: Automatic error aggregation and alerting
- **Performance Profiling**: Identify slow functions and optimize

## Best Practices

1. **Keep Functions Small**: Single responsibility, easy to test
2. **Handle Errors Gracefully**: Always provide meaningful error messages
3. **Use Async for Heavy Operations**: Don't block API responses
4. **Validate Inputs**: Never trust external data
5. **Log Important Events**: Aid debugging and monitoring
6. **Test Thoroughly**: Use built-in testing framework
7. **Document Function Purpose**: Clear descriptions and examples

## Migration Strategy

When updating function code:

1. **Compatible Changes**: Hot-reloaded like schema changes
2. **Breaking Changes**: Versioned with schema migrations
3. **Rollback Support**: Previous function versions preserved
4. **Gradual Rollout**: Test with subset of tenants first

This function system provides the flexibility of full programming while maintaining BackSaas's declarative, schema-driven philosophy.
