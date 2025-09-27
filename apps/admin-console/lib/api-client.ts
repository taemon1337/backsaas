import { AuthService } from './auth'

// API Client Configuration
const API_CONFIG = {
  baseURL: '', // Use relative URLs to go through gateway
  timeout: 30000,
  retries: 3,
}

// API Response Types
export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
  message?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  limit: number
  hasNext: boolean
  hasPrev: boolean
}

// Error Types
export class ApiError extends Error {
  constructor(
    message: string,
    public status?: number,
    public code?: string,
    public details?: any
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

// HTTP Methods
type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'

interface RequestOptions {
  method?: HttpMethod
  headers?: Record<string, string>
  body?: any
  timeout?: number
  requireAuth?: boolean
}

class ApiClient {
  private baseURL: string
  private defaultTimeout: number
  private maxRetries: number

  constructor() {
    this.baseURL = API_CONFIG.baseURL
    this.defaultTimeout = API_CONFIG.timeout
    this.maxRetries = API_CONFIG.retries
  }

  /**
   * Make an authenticated API request
   */
  private async makeRequest<T>(
    endpoint: string,
    options: RequestOptions = {}
  ): Promise<T> {
    const {
      method = 'GET',
      headers = {},
      body,
      timeout = this.defaultTimeout,
      requireAuth = true,
    } = options

    // Build full URL
    const url = `${this.baseURL}${endpoint}`

    // Prepare headers
    const requestHeaders: Record<string, string> = {
      'Content-Type': 'application/json',
      ...headers,
    }

    // Add authentication if required
    if (requireAuth) {
      const token = AuthService.getToken()
      if (token) {
        requestHeaders.Authorization = `Bearer ${token}`
      } else {
        throw new ApiError('Authentication required', 401, 'AUTH_REQUIRED')
      }
    }

    // Prepare request options
    const requestOptions: RequestInit = {
      method,
      headers: requestHeaders,
      signal: AbortSignal.timeout(timeout),
    }

    // Add body for non-GET requests
    if (body && method !== 'GET') {
      requestOptions.body = JSON.stringify(body)
    }

    // Make request with retries
    let lastError: Error
    for (let attempt = 1; attempt <= this.maxRetries; attempt++) {
      try {
        const response = await fetch(url, requestOptions)
        
        // Handle response
        if (!response.ok) {
          const errorData = await this.parseErrorResponse(response)
          throw new ApiError(
            errorData.message || `HTTP ${response.status}`,
            response.status,
            errorData.code,
            errorData
          )
        }

        // Parse successful response
        const contentType = response.headers.get('content-type')
        if (contentType?.includes('application/json')) {
          return await response.json()
        } else {
          return response.text() as T
        }
      } catch (error) {
        lastError = error as Error
        
        // Don't retry on auth errors or client errors (4xx)
        if (error instanceof ApiError && error.status && error.status < 500) {
          throw error
        }

        // Don't retry on last attempt
        if (attempt === this.maxRetries) {
          break
        }

        // Wait before retry (exponential backoff)
        await this.delay(Math.pow(2, attempt - 1) * 1000)
      }
    }

    throw lastError!
  }

  /**
   * Parse error response
   */
  private async parseErrorResponse(response: Response): Promise<any> {
    try {
      const contentType = response.headers.get('content-type')
      if (contentType?.includes('application/json')) {
        return await response.json()
      } else {
        return { message: await response.text() }
      }
    } catch {
      return { message: `HTTP ${response.status} ${response.statusText}` }
    }
  }

  /**
   * Delay helper for retries
   */
  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms))
  }

  // ============================================================================
  // AUTH API
  // ============================================================================

  async login(email: string, password: string): Promise<{ token: string; user: any }> {
    return this.makeRequest('/api/platform/admin/login', {
      method: 'POST',
      body: { email, password },
      requireAuth: false,
    })
  }

  async refreshToken(): Promise<{ token: string }> {
    return this.makeRequest('/api/platform/admin/refresh', {
      method: 'POST',
    })
  }

  async logout(): Promise<void> {
    return this.makeRequest('/api/platform/admin/logout', {
      method: 'POST',
    })
  }

  // ============================================================================
  // SYSTEM HEALTH API
  // ============================================================================

  async getHealthSummary(): Promise<any> {
    return this.makeRequest('/api/system-health/api/summary', {
      requireAuth: false, // Temporarily false for testing
    })
  }

  async getHealthServices(): Promise<any> {
    return this.makeRequest('/api/system-health/api/services', {
      requireAuth: false, // Temporarily false for testing
    })
  }

  async getHealthStatus(): Promise<any> {
    return this.makeRequest('/api/system-health/api/status', {
      requireAuth: false, // Temporarily false for testing
    })
  }

  async triggerCoverageCollection(): Promise<{ message: string }> {
    return this.makeRequest('/api/system-health/api/collect', {
      method: 'POST',
      requireAuth: false, // Temporarily false for testing
    })
  }

  // ============================================================================
  // SYSTEM TESTING API
  // ============================================================================

  async getSystemTests(runTests: boolean = false): Promise<any> {
    const url = runTests ? '/api/platform/health/tests?run=true' : '/api/platform/health/tests'
    return this.makeRequest(url, {
      requireAuth: true, // Requires admin authentication
      timeout: runTests ? 60000 : 10000, // Longer timeout for running tests
    })
  }

  async runSystemTests(): Promise<any> {
    return this.getSystemTests(true)
  }

  async getSystemHealth(): Promise<any> {
    try {
      const [summary, services, status] = await Promise.all([
        this.getHealthSummary(),
        this.getHealthServices(), 
        this.getHealthStatus()
      ])

      // Determine overall health status based on coverage and service status
      const overallCoverage = summary.overall_coverage || 0
      const serviceStatuses = Object.values(status.services || {}) as any[]
      const allServicesHealthy = serviceStatuses.every(service => !service.collecting)
      
      let healthStatus = 'healthy'
      if (overallCoverage < 20) {
        healthStatus = 'warning'
      }
      if (overallCoverage < 10) {
        healthStatus = 'critical'
      }

      // Format services for dashboard display
      const formattedServices = Object.entries(summary.services || {}).map(([name, coverage]) => ({
        name: name.charAt(0).toUpperCase() + name.slice(1),
        status: (coverage as number) > 15 ? 'up' : 'warning',
        coverage: `${(coverage as number).toFixed(1)}%`,
        response_time: Math.floor(Math.random() * 100) + 50 // Mock response time
      }))

      return {
        data: {
          status: healthStatus,
          overall_coverage: overallCoverage,
          services: formattedServices,
          summary,
          raw_status: status
        }
      }
    } catch (error) {
      console.error('Failed to fetch system health:', error)
      return {
        data: {
          status: 'unknown',
          services: [],
          overall_coverage: 0
        }
      }
    }
  }

  // ============================================================================
  // GATEWAY METRICS API
  // ============================================================================

  async getGatewayMetrics(): Promise<any> {
    return this.makeRequest('/metrics', {
      requireAuth: false, // Metrics endpoint is public
    })
  }

  // ============================================================================
  // TENANT MANAGEMENT API
  // ============================================================================

  async getTenants(params?: {
    page?: number
    limit?: number
    search?: string
  }): Promise<PaginatedResponse<any>> {
    const searchParams = new URLSearchParams()
    if (params?.page) searchParams.set('page', params.page.toString())
    if (params?.limit) searchParams.set('limit', params.limit.toString())
    if (params?.search) searchParams.set('search', params.search)
    
    const query = searchParams.toString()
    const endpoint = `/api/platform/tenants${query ? `?${query}` : ''}`
    
    return this.makeRequest(endpoint)
  }

  async getTenant(id: string): Promise<any> {
    return this.makeRequest(`/api/platform/tenants/${id}`)
  }

  async createTenant(data: any): Promise<any> {
    return this.makeRequest('/api/platform/tenants', {
      method: 'POST',
      body: data,
    })
  }

  async updateTenant(id: string, data: any): Promise<any> {
    return this.makeRequest(`/api/platform/tenants/${id}`, {
      method: 'PUT',
      body: data,
    })
  }

  async deleteTenant(id: string): Promise<void> {
    return this.makeRequest(`/api/platform/tenants/${id}`, {
      method: 'DELETE',
    })
  }

  // ============================================================================
  // SCHEMA MANAGEMENT API
  // ============================================================================

  async getSchemas(params?: {
    page?: number
    limit?: number
    tenantId?: string
  }): Promise<PaginatedResponse<any>> {
    const searchParams = new URLSearchParams()
    if (params?.page) searchParams.set('page', params.page.toString())
    if (params?.limit) searchParams.set('limit', params.limit.toString())
    if (params?.tenantId) searchParams.set('tenant_id', params.tenantId)
    
    const query = searchParams.toString()
    const endpoint = `/api/platform/schemas${query ? `?${query}` : ''}`
    
    return this.makeRequest(endpoint)
  }

  async getSchema(id: string): Promise<any> {
    return this.makeRequest(`/api/platform/schemas/${id}`)
  }

  async createSchema(data: any): Promise<any> {
    return this.makeRequest('/api/platform/schemas', {
      method: 'POST',
      body: data,
    })
  }

  async updateSchema(id: string, data: any): Promise<any> {
    return this.makeRequest(`/api/platform/schemas/${id}`, {
      method: 'PUT',
      body: data,
    })
  }

  async deleteSchema(id: string): Promise<void> {
    return this.makeRequest(`/api/platform/schemas/${id}`, {
      method: 'DELETE',
    })
  }

  // ============================================================================
  // USER MANAGEMENT API
  // ============================================================================

  async getUsers(params?: {
    page?: number
    limit?: number
    tenantId?: string
  }): Promise<PaginatedResponse<any>> {
    const searchParams = new URLSearchParams()
    if (params?.page) searchParams.set('page', params.page.toString())
    if (params?.limit) searchParams.set('limit', params.limit.toString())
    if (params?.tenantId) searchParams.set('tenant_id', params.tenantId)
    
    const query = searchParams.toString()
    const endpoint = `/api/platform/users${query ? `?${query}` : ''}`
    
    return this.makeRequest(endpoint)
  }

  async getUser(id: string): Promise<any> {
    return this.makeRequest(`/api/platform/users/${id}`)
  }

  async createUser(data: any): Promise<any> {
    return this.makeRequest('/api/platform/users', {
      method: 'POST',
      body: data,
    })
  }

  async updateUser(id: string, data: any): Promise<any> {
    return this.makeRequest(`/api/platform/users/${id}`, {
      method: 'PUT',
      body: data,
    })
  }

  async deleteUser(id: string): Promise<void> {
    return this.makeRequest(`/api/platform/users/${id}`, {
      method: 'DELETE',
    })
  }

  // ============================================================================
  // ANALYTICS API
  // ============================================================================

  async getAnalytics(params?: {
    timeRange?: string
    tenantId?: string
    metric?: string
  }): Promise<any> {
    const searchParams = new URLSearchParams()
    if (params?.timeRange) searchParams.set('time_range', params.timeRange)
    if (params?.tenantId) searchParams.set('tenant_id', params.tenantId)
    if (params?.metric) searchParams.set('metric', params.metric)
    
    const query = searchParams.toString()
    const endpoint = `/api/platform/analytics${query ? `?${query}` : ''}`
    
    return this.makeRequest(endpoint)
  }

  // ============================================================================
  // SYSTEM SETTINGS API
  // ============================================================================

  async getSettings(): Promise<any> {
    return this.makeRequest('/api/platform/settings')
  }

  async updateSettings(data: any): Promise<any> {
    return this.makeRequest('/api/platform/settings', {
      method: 'PUT',
      body: data,
    })
  }
}

// Create singleton instance
export const apiClient = new ApiClient()

// Export types (already exported above, no need to re-export)
