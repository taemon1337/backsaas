import { AuthService } from './auth'

// API Client Configuration
const API_CONFIG = {
  baseURL: '', // Use relative URLs to go through gateway
  timeout: 30000,
  retries: 5, // Increased for rate limiting
  baseDelay: 1000, // Base delay in ms
  maxDelay: 30000, // Max delay in ms
  jitterFactor: 0.1, // Add randomness to prevent thundering herd
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
          const apiError = new ApiError(
            errorData.message || `HTTP ${response.status}`,
            response.status,
            errorData.code,
            errorData
          )
          
          // For rate limiting, add retry-after info if available
          if (response.status === 429) {
            const retryAfter = response.headers.get('retry-after')
            if (retryAfter) {
              apiError.details = { ...apiError.details, retryAfter: parseInt(retryAfter) }
            }
          }
          
          throw apiError
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
        
        // Don't retry on auth errors or most client errors (4xx)
        // BUT DO retry on 429 (Too Many Requests) and 408 (Request Timeout)
        if (error instanceof ApiError && error.status) {
          const shouldRetry = error.status === 429 || error.status === 408 || error.status >= 500
          if (!shouldRetry) {
            throw error
          }
        }

        // Don't retry on last attempt
        if (attempt === this.maxRetries) {
          break
        }

        // Calculate delay with smart backoff
        let delay = this.calculateBackoffDelay(attempt, error as ApiError)
        
        console.log(`API request failed (attempt ${attempt}/${this.maxRetries}), retrying in ${delay}ms...`, {
          url,
          status: (error as ApiError).status,
          message: (error as ApiError).message
        })

        await this.delay(delay)
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

  /**
   * Calculate smart backoff delay for retries
   */
  private calculateBackoffDelay(attempt: number, error?: ApiError): number {
    // If server provides retry-after header (for 429), respect it
    if (error?.status === 429 && error.details?.retryAfter) {
      const retryAfterMs = error.details.retryAfter * 1000
      // Cap at max delay and add some jitter
      const cappedDelay = Math.min(retryAfterMs, API_CONFIG.maxDelay)
      return this.addJitter(cappedDelay)
    }

    // Exponential backoff: baseDelay * (2 ^ (attempt - 1))
    const exponentialDelay = API_CONFIG.baseDelay * Math.pow(2, attempt - 1)
    
    // Cap at maximum delay
    const cappedDelay = Math.min(exponentialDelay, API_CONFIG.maxDelay)
    
    // Add jitter to prevent thundering herd
    return this.addJitter(cappedDelay)
  }

  /**
   * Add jitter to delay to prevent thundering herd
   */
  private addJitter(delay: number): number {
    const jitter = delay * API_CONFIG.jitterFactor * Math.random()
    return Math.floor(delay + jitter)
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

  async getTenants(params?: { page?: number; limit?: number; search?: string }): Promise<PaginatedResponse<any>> {
    const queryParams = new URLSearchParams()
    if (params?.page) queryParams.append('page', params.page.toString())
    if (params?.limit) queryParams.append('limit', params.limit.toString())
    if (params?.search) queryParams.append('search', params.search)
    
    const url = `/api/platform/tenants${queryParams.toString() ? '?' + queryParams.toString() : ''}`
    return this.makeRequest(url, { requireAuth: true })
  }

  async getTenant(id: string): Promise<any> {
    return this.makeRequest(`/api/platform/tenants/${id}`, { requireAuth: true })
  }

  async createTenant(data: any): Promise<any> {
    return this.makeRequest('/api/platform/tenants', {
      method: 'POST',
      body: data,
      requireAuth: true,
    })
  }

  async updateTenant(id: string, data: any): Promise<any> {
    return this.makeRequest(`/api/platform/tenants/${id}`, {
      method: 'PUT',
      body: data,
      requireAuth: true,
    })
  }

  async deleteTenant(id: string): Promise<any> {
    return this.makeRequest(`/api/platform/tenants/${id}`, {
      method: 'DELETE',
      requireAuth: true,
    })
  }

  // ============================================================================
  // SCHEMA MANAGEMENT API
  // ============================================================================

  async getSchemas(params?: {
    page?: number
    limit?: number
    search?: string
    tenantId?: string
  }): Promise<PaginatedResponse<any>> {
    const searchParams = new URLSearchParams()
    if (params?.page) searchParams.set('page', params.page.toString())
    if (params?.limit) searchParams.set('limit', params.limit.toString())
    if (params?.search) searchParams.set('search', params.search)
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
  async getUser(id: string): Promise<any> {
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
