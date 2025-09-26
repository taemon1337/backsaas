import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'

// Types for API responses
export interface ApiError {
  message: string
  code?: string
  details?: any
}

export interface PaginatedResponse<T> {
  data: T[]
  pagination: {
    page: number
    limit: number
    total: number
    totalPages: number
  }
}

export interface Schema {
  id?: string
  name: string
  version: string
  entities: Entity[]
  metadata: {
    created_at: string
    updated_at: string
    author: string
    tenant_id?: string
  }
}

export interface Entity {
  name: string
  fields: Field[]
  relationships: Relationship[]
  functions: Function[]
}

export interface Field {
  name: string
  type: string
  required: boolean
  unique: boolean
  default?: any
  validation?: any
}

export interface Relationship {
  name: string
  type: 'one-to-one' | 'one-to-many' | 'many-to-many'
  target: string
}

export interface Function {
  name: string
  type: 'validation' | 'transformation' | 'hook'
  trigger: string
  code: string
}

export interface ValidationResult {
  valid: boolean
  errors: ValidationError[]
  warnings: ValidationWarning[]
  summary: {
    entities: number
    fields: number
    relationships: number
    functions: number
  }
}

export interface ValidationError {
  type: 'schema' | 'entity' | 'field' | 'relationship' | 'function'
  path: string
  message: string
  severity: 'error' | 'warning'
}

export interface ValidationWarning {
  type: string
  path: string
  message: string
  suggestion?: string
}

export interface MigrationPlan {
  id: string
  fromVersion: string
  toVersion: string
  steps: MigrationStep[]
  estimatedDuration: number
  riskLevel: 'low' | 'medium' | 'high'
  canRollback: boolean
}

export interface MigrationStep {
  id: string
  type: 'expand' | 'backfill' | 'contract'
  operation: string
  description: string
  sql: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  duration?: number
  error?: string
}

export interface EventStreamMessage {
  id: string
  type: 'schema.created' | 'schema.updated' | 'schema.deleted' | 'migration.started' | 'migration.completed'
  timestamp: string
  data: any
}

class ControlPlaneApiClient {
  private client: AxiosInstance
  private baseURL: string

  constructor() {
    // Use environment variables or fallback to defaults
    this.baseURL = process.env.NEXT_PUBLIC_GATEWAY_API_URL || 'http://localhost:8000'
    
    this.client = axios.create({
      baseURL: this.baseURL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Request interceptor
    this.client.interceptors.request.use(
      (config) => {
        // Add any auth headers if needed
        // const token = getAuthToken()
        // if (token) {
        //   config.headers.Authorization = `Bearer ${token}`
        // }
        return config
      },
      (error) => Promise.reject(error)
    )

    // Response interceptor
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        const apiError: ApiError = {
          message: error.response?.data?.message || error.message || 'An error occurred',
          code: error.response?.data?.code || error.code,
          details: error.response?.data?.details || error.response?.data,
        }
        return Promise.reject(apiError)
      }
    )
  }

  private async makeRequest<T>(
    endpoint: string,
    options: AxiosRequestConfig = {}
  ): Promise<T> {
    try {
      const response: AxiosResponse<T> = await this.client.request({
        url: endpoint,
        ...options,
      })
      return response.data
    } catch (error) {
      throw error
    }
  }

  // ============================================================================
  // SCHEMA MANAGEMENT API
  // ============================================================================

  async getSchemas(params?: {
    page?: number
    limit?: number
    search?: string
    tenant_id?: string
  }): Promise<PaginatedResponse<Schema>> {
    const queryParams = new URLSearchParams()
    if (params?.page) queryParams.append('page', params.page.toString())
    if (params?.limit) queryParams.append('limit', params.limit.toString())
    if (params?.search) queryParams.append('search', params.search)
    if (params?.tenant_id) queryParams.append('tenant_id', params.tenant_id)

    return this.makeRequest(`/api/platform/schemas?${queryParams.toString()}`)
  }

  async getSchema(id: string): Promise<Schema> {
    return this.makeRequest(`/api/platform/schemas/${id}`)
  }

  async createSchema(schema: Omit<Schema, 'id' | 'metadata'>): Promise<Schema> {
    return this.makeRequest('/api/platform/schemas', {
      method: 'POST',
      data: schema,
    })
  }

  async updateSchema(id: string, schema: Partial<Schema>): Promise<Schema> {
    return this.makeRequest(`/api/platform/schemas/${id}`, {
      method: 'PUT',
      data: schema,
    })
  }

  async deleteSchema(id: string): Promise<void> {
    return this.makeRequest(`/api/platform/schemas/${id}`, {
      method: 'DELETE',
    })
  }

  async publishSchema(id: string): Promise<{ success: boolean; message: string }> {
    return this.makeRequest(`/api/platform/schemas/${id}/publish`, {
      method: 'POST',
    })
  }

  // ============================================================================
  // SCHEMA VALIDATION API
  // ============================================================================

  async validateSchema(schema: Schema): Promise<ValidationResult> {
    return this.makeRequest('/api/platform/schemas/validate', {
      method: 'POST',
      data: schema,
    })
  }

  async validateSchemaYaml(yamlContent: string): Promise<ValidationResult> {
    return this.makeRequest('/api/platform/schemas/validate/yaml', {
      method: 'POST',
      data: { yaml: yamlContent },
    })
  }

  // ============================================================================
  // MIGRATION PLANNING API
  // ============================================================================

  async generateMigrationPlan(
    fromSchemaId: string,
    toSchema: Schema
  ): Promise<MigrationPlan> {
    return this.makeRequest('/api/platform/migrations/plan', {
      method: 'POST',
      data: {
        from_schema_id: fromSchemaId,
        to_schema: toSchema,
      },
    })
  }

  async getMigrationPlan(planId: string): Promise<MigrationPlan> {
    return this.makeRequest(`/api/platform/migrations/plans/${planId}`)
  }

  async executeMigrationPlan(planId: string): Promise<{ success: boolean; message: string }> {
    return this.makeRequest(`/api/platform/migrations/plans/${planId}/execute`, {
      method: 'POST',
    })
  }

  async getMigrationStatus(planId: string): Promise<MigrationPlan> {
    return this.makeRequest(`/api/platform/migrations/plans/${planId}/status`)
  }

  async rollbackMigration(planId: string): Promise<{ success: boolean; message: string }> {
    return this.makeRequest(`/api/platform/migrations/plans/${planId}/rollback`, {
      method: 'POST',
    })
  }

  // ============================================================================
  // SQL GENERATION API
  // ============================================================================

  async generateSql(schema: Schema): Promise<{ sql: string; statements: string[] }> {
    return this.makeRequest('/api/platform/schemas/generate-sql', {
      method: 'POST',
      data: schema,
    })
  }

  async previewSql(schema: Schema): Promise<{ sql: string; statements: string[] }> {
    return this.makeRequest('/api/platform/schemas/preview-sql', {
      method: 'POST',
      data: schema,
    })
  }

  // ============================================================================
  // EVENT STREAMING API
  // ============================================================================

  async getSchemaEvents(params?: {
    schema_id?: string
    event_type?: string
    since?: string
    limit?: number
  }): Promise<EventStreamMessage[]> {
    const queryParams = new URLSearchParams()
    if (params?.schema_id) queryParams.append('schema_id', params.schema_id)
    if (params?.event_type) queryParams.append('event_type', params.event_type)
    if (params?.since) queryParams.append('since', params.since)
    if (params?.limit) queryParams.append('limit', params.limit.toString())

    return this.makeRequest(`/api/platform/events/schemas?${queryParams.toString()}`)
  }

  // ============================================================================
  // UTILITY METHODS
  // ============================================================================

  async healthCheck(): Promise<{ status: string; timestamp: string }> {
    return this.makeRequest('/api/platform/health')
  }

  async getSystemInfo(): Promise<{
    version: string
    environment: string
    database_status: string
    redis_status: string
  }> {
    return this.makeRequest('/api/platform/system/info')
  }

  // Convert schema to YAML format
  schemaToYaml(schema: Schema): string {
    // This would use js-yaml library
    return JSON.stringify(schema, null, 2) // Fallback to JSON for now
  }

  // Parse YAML to schema format
  yamlToSchema(yamlContent: string): Schema {
    try {
      // This would use js-yaml library
      return JSON.parse(yamlContent) // Fallback to JSON for now
    } catch (error) {
      throw new Error(`Invalid YAML format: ${error.message}`)
    }
  }
}

// Create singleton instance
export const apiClient = new ControlPlaneApiClient()

// Export the class for testing
export { ControlPlaneApiClient }
