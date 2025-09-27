import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { 
  Tenant, 
  TenantUser, 
  EntitySchema, 
  EntityRecord, 
  QueryOptions, 
  QueryResult,
  DashboardConfig,
  Workflow,
  ApiResponse,
  ApiError 
} from './types'

class TenantApiClient {
  private client: AxiosInstance
  private tenantId: string | null = null
  private token: string | null = null

  constructor() {
    const baseURL = process.env.NEXT_PUBLIC_GATEWAY_API_URL || 'http://localhost:8000'
    
    this.client = axios.create({
      baseURL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Request interceptor
    this.client.interceptors.request.use(
      (config) => {
        // Add tenant context headers
        if (this.tenantId) {
          config.headers['X-Tenant-ID'] = this.tenantId
        }
        
        // Add authentication
        if (this.token) {
          config.headers['Authorization'] = `Bearer ${this.token}`
        }
        
        return config
      },
      (error) => Promise.reject(error)
    )

    // Response interceptor
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        const apiError: ApiError = {
          code: error.response?.data?.code || error.code || 'UNKNOWN_ERROR',
          message: error.response?.data?.message || error.message || 'An error occurred',
          details: error.response?.data?.details || error.response?.data,
          status: error.response?.status,
        }
        
        // Handle authentication errors
        if (error.response?.status === 401) {
          // Clear session and redirect to login
          localStorage.removeItem('tenant-session')
          window.location.href = '/login'
        }
        
        return Promise.reject(apiError)
      }
    )
  }

  // Set tenant context
  setTenantContext(tenantId: string, token?: string) {
    this.tenantId = tenantId
    if (token) {
      this.token = token
    }
  }

  // Clear tenant context
  clearTenantContext() {
    this.tenantId = null
    this.token = null
  }

  private async makeRequest<T>(
    endpoint: string,
    options: AxiosRequestConfig = {}
  ): Promise<T> {
    try {
      const response: AxiosResponse<ApiResponse<T>> = await this.client.request({
        url: endpoint,
        ...options,
      })
      
      if (response.data.success === false) {
        throw new Error(response.data.error?.message || 'API request failed')
      }
      
      return response.data.data as T
    } catch (error) {
      throw error
    }
  }

  // ============================================================================
  // TENANT MANAGEMENT
  // ============================================================================

  async getTenant(tenantSlug: string): Promise<Tenant> {
    return this.makeRequest(`/api/tenants/by-slug/${tenantSlug}`)
  }

  async updateTenant(tenantId: string, updates: Partial<Tenant>): Promise<Tenant> {
    return this.makeRequest(`/api/tenants/${tenantId}`, {
      method: 'PUT',
      data: updates,
    })
  }

  // ============================================================================
  // AUTHENTICATION
  // ============================================================================

  async login(email: string, password: string, tenantSlug: string): Promise<{
    user: TenantUser
    tenant: Tenant
    token: string
    expiresAt: string
  }> {
    return this.makeRequest('/api/platform/auth/login', {
      method: 'POST',
      data: { email, password },
    })
  }

  async logout(): Promise<void> {
    return this.makeRequest('/api/auth/logout', {
      method: 'POST',
    })
  }

  async refreshToken(): Promise<{ token: string; expiresAt: string }> {
    return this.makeRequest('/api/auth/refresh', {
      method: 'POST',
    })
  }

  async getCurrentUser(): Promise<TenantUser> {
    return this.makeRequest('/api/auth/me')
  }

  // ============================================================================
  // SCHEMA MANAGEMENT
  // ============================================================================

  async getSchemas(): Promise<EntitySchema[]> {
    return this.makeRequest('/api/tenant/schemas')
  }

  async getSchema(schemaId: string): Promise<EntitySchema> {
    return this.makeRequest(`/api/tenant/schemas/${schemaId}`)
  }

  async createSchema(schema: Omit<EntitySchema, 'id' | 'metadata'>): Promise<EntitySchema> {
    return this.makeRequest('/api/tenant/schemas', {
      method: 'POST',
      data: schema,
    })
  }

  async updateSchema(schemaId: string, updates: Partial<EntitySchema>): Promise<EntitySchema> {
    return this.makeRequest(`/api/tenant/schemas/${schemaId}`, {
      method: 'PUT',
      data: updates,
    })
  }

  async deleteSchema(schemaId: string): Promise<void> {
    return this.makeRequest(`/api/tenant/schemas/${schemaId}`, {
      method: 'DELETE',
    })
  }

  async deploySchema(schemaId: string): Promise<{ success: boolean; message: string }> {
    return this.makeRequest(`/api/tenant/schemas/${schemaId}/deploy`, {
      method: 'POST',
    })
  }

  // ============================================================================
  // ENTITY DATA MANAGEMENT
  // ============================================================================

  async getEntityRecords(
    entity: string, 
    options: QueryOptions = {}
  ): Promise<QueryResult<EntityRecord>> {
    const params = new URLSearchParams()
    
    if (options.page) params.append('page', options.page.toString())
    if (options.limit) params.append('limit', options.limit.toString())
    if (options.sort) params.append('sort', options.sort)
    if (options.order) params.append('order', options.order)
    if (options.filters) {
      Object.entries(options.filters).forEach(([key, value]) => {
        params.append(`filter[${key}]`, value.toString())
      })
    }
    if (options.include) {
      params.append('include', options.include.join(','))
    }

    return this.makeRequest(`/api/tenant/entities/${entity}?${params.toString()}`)
  }

  async getEntityRecord(entity: string, recordId: string): Promise<EntityRecord> {
    return this.makeRequest(`/api/tenant/entities/${entity}/${recordId}`)
  }

  async createEntityRecord(entity: string, data: Record<string, any>): Promise<EntityRecord> {
    return this.makeRequest(`/api/tenant/entities/${entity}`, {
      method: 'POST',
      data,
    })
  }

  async updateEntityRecord(
    entity: string, 
    recordId: string, 
    data: Record<string, any>
  ): Promise<EntityRecord> {
    return this.makeRequest(`/api/tenant/entities/${entity}/${recordId}`, {
      method: 'PUT',
      data,
    })
  }

  async deleteEntityRecord(entity: string, recordId: string): Promise<void> {
    return this.makeRequest(`/api/tenant/entities/${entity}/${recordId}`, {
      method: 'DELETE',
    })
  }

  async bulkCreateEntityRecords(
    entity: string, 
    records: Record<string, any>[]
  ): Promise<EntityRecord[]> {
    return this.makeRequest(`/api/tenant/entities/${entity}/bulk`, {
      method: 'POST',
      data: { records },
    })
  }

  async bulkUpdateEntityRecords(
    entity: string, 
    updates: { id: string; data: Record<string, any> }[]
  ): Promise<EntityRecord[]> {
    return this.makeRequest(`/api/tenant/entities/${entity}/bulk`, {
      method: 'PUT',
      data: { updates },
    })
  }

  async bulkDeleteEntityRecords(entity: string, recordIds: string[]): Promise<void> {
    return this.makeRequest(`/api/tenant/entities/${entity}/bulk`, {
      method: 'DELETE',
      data: { ids: recordIds },
    })
  }

  // ============================================================================
  // DASHBOARD MANAGEMENT
  // ============================================================================

  async getDashboards(): Promise<DashboardConfig[]> {
    return this.makeRequest('/api/tenant/dashboards')
  }

  async getDashboard(dashboardId: string): Promise<DashboardConfig> {
    return this.makeRequest(`/api/tenant/dashboards/${dashboardId}`)
  }

  async createDashboard(dashboard: Omit<DashboardConfig, 'id' | 'tenantId'>): Promise<DashboardConfig> {
    return this.makeRequest('/api/tenant/dashboards', {
      method: 'POST',
      data: dashboard,
    })
  }

  async updateDashboard(
    dashboardId: string, 
    updates: Partial<DashboardConfig>
  ): Promise<DashboardConfig> {
    return this.makeRequest(`/api/tenant/dashboards/${dashboardId}`, {
      method: 'PUT',
      data: updates,
    })
  }

  async deleteDashboard(dashboardId: string): Promise<void> {
    return this.makeRequest(`/api/tenant/dashboards/${dashboardId}`, {
      method: 'DELETE',
    })
  }

  // ============================================================================
  // WORKFLOW MANAGEMENT
  // ============================================================================

  async getWorkflows(): Promise<Workflow[]> {
    return this.makeRequest('/api/tenant/workflows')
  }

  async getWorkflow(workflowId: string): Promise<Workflow> {
    return this.makeRequest(`/api/tenant/workflows/${workflowId}`)
  }

  async createWorkflow(workflow: Omit<Workflow, 'id' | 'tenantId'>): Promise<Workflow> {
    return this.makeRequest('/api/tenant/workflows', {
      method: 'POST',
      data: workflow,
    })
  }

  async updateWorkflow(workflowId: string, updates: Partial<Workflow>): Promise<Workflow> {
    return this.makeRequest(`/api/tenant/workflows/${workflowId}`, {
      method: 'PUT',
      data: updates,
    })
  }

  async deleteWorkflow(workflowId: string): Promise<void> {
    return this.makeRequest(`/api/tenant/workflows/${workflowId}`, {
      method: 'DELETE',
    })
  }

  async executeWorkflow(workflowId: string, data?: Record<string, any>): Promise<{
    executionId: string
    status: string
  }> {
    return this.makeRequest(`/api/tenant/workflows/${workflowId}/execute`, {
      method: 'POST',
      data,
    })
  }

  // ============================================================================
  // USER MANAGEMENT
  // ============================================================================

  async getTenantUsers(): Promise<TenantUser[]> {
    return this.makeRequest('/api/tenant/users')
  }

  async getTenantUser(userId: string): Promise<TenantUser> {
    return this.makeRequest(`/api/tenant/users/${userId}`)
  }

  async inviteUser(email: string, roles: string[]): Promise<{ success: boolean; message: string }> {
    return this.makeRequest('/api/tenant/users/invite', {
      method: 'POST',
      data: { email, roles },
    })
  }

  async updateUserRoles(userId: string, roles: string[]): Promise<TenantUser> {
    return this.makeRequest(`/api/tenant/users/${userId}/roles`, {
      method: 'PUT',
      data: { roles },
    })
  }

  async deactivateUser(userId: string): Promise<void> {
    return this.makeRequest(`/api/tenant/users/${userId}/deactivate`, {
      method: 'POST',
    })
  }

  // ============================================================================
  // ANALYTICS
  // ============================================================================

  async getAnalytics(params: {
    entity?: string
    metric: string
    period: 'day' | 'week' | 'month' | 'year'
    startDate?: string
    endDate?: string
  }): Promise<{
    data: Array<{ date: string; value: number }>
    summary: { total: number; change: number; changePercent: number }
  }> {
    const queryParams = new URLSearchParams()
    Object.entries(params).forEach(([key, value]) => {
      if (value) queryParams.append(key, value.toString())
    })

    return this.makeRequest(`/api/tenant/analytics?${queryParams.toString()}`)
  }

  // ============================================================================
  // FILE MANAGEMENT
  // ============================================================================

  async uploadFile(file: File, entity?: string, recordId?: string): Promise<{
    id: string
    url: string
    filename: string
    size: number
    mimeType: string
  }> {
    const formData = new FormData()
    formData.append('file', file)
    if (entity) formData.append('entity', entity)
    if (recordId) formData.append('recordId', recordId)

    return this.makeRequest('/api/tenant/files/upload', {
      method: 'POST',
      data: formData,
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  }

  async deleteFile(fileId: string): Promise<void> {
    return this.makeRequest(`/api/tenant/files/${fileId}`, {
      method: 'DELETE',
    })
  }
}

// Create singleton instance
export const tenantApi = new TenantApiClient()

// Export the class for testing
export { TenantApiClient }
