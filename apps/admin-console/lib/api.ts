import axios, { AxiosInstance, AxiosRequestConfig } from 'axios'
import { AuthService } from './auth'

export interface ApiResponse<T = any> {
  data: T
  success: boolean
  error?: string
  message?: string
}

export interface Tenant {
  id: string
  name: string
  slug: string
  status: 'active' | 'inactive' | 'suspended' | 'pending'
  created_at: string
  updated_at: string
  schema_id?: string
  settings: {
    max_users?: number
    max_storage?: number
    features?: string[]
  }
  usage: {
    users: number
    storage: number
    api_calls: number
  }
  billing: {
    plan: string
    status: 'active' | 'past_due' | 'canceled'
    next_billing_date?: string
  }
}

export interface Schema {
  id: string
  name: string
  version: string
  entities: any[]
  functions: any[]
  created_at: string
  updated_at: string
  status: 'draft' | 'active' | 'deprecated'
}

export interface SystemHealth {
  status: 'healthy' | 'degraded' | 'unhealthy'
  services: {
    name: string
    status: 'up' | 'down' | 'degraded'
    response_time?: number
    last_check: string
  }[]
  metrics: {
    cpu_usage: number
    memory_usage: number
    disk_usage: number
    active_connections: number
  }
}

export interface AdminUser {
  id: string
  email: string
  name: string
  role: 'super_admin' | 'platform_admin' | 'support_admin' | 'billing_admin'
  status: 'active' | 'inactive'
  created_at: string
  last_login?: string
}

class ApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: process.env.GATEWAY_API_URL || 'http://localhost:8000',
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Request interceptor to add auth token
    this.client.interceptors.request.use(
      (config) => {
        const token = AuthService.getToken()
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        // Always use system tenant for admin operations
        config.headers['X-Tenant-ID'] = 'system'
        return config
      },
      (error) => {
        return Promise.reject(error)
      }
    )

    // Response interceptor to handle auth errors
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        if (error.response?.status === 401) {
          // Try to refresh token
          const refreshed = await AuthService.refreshToken()
          if (!refreshed) {
            AuthService.logout()
            return Promise.reject(error)
          }
          // Retry original request
          return this.client.request(error.config)
        }
        return Promise.reject(error)
      }
    )
  }

  // Tenant Management
  async getTenants(params?: { 
    page?: number
    limit?: number
    status?: string
    search?: string
  }): Promise<ApiResponse<{ tenants: Tenant[]; total: number }>> {
    const response = await this.client.get('/admin/tenants', { params })
    return response.data
  }

  async getTenant(id: string): Promise<ApiResponse<Tenant>> {
    const response = await this.client.get(`/admin/tenants/${id}`)
    return response.data
  }

  async createTenant(tenant: Partial<Tenant>): Promise<ApiResponse<Tenant>> {
    const response = await this.client.post('/admin/tenants', tenant)
    return response.data
  }

  async updateTenant(id: string, updates: Partial<Tenant>): Promise<ApiResponse<Tenant>> {
    const response = await this.client.put(`/admin/tenants/${id}`, updates)
    return response.data
  }

  async deleteTenant(id: string): Promise<ApiResponse<void>> {
    const response = await this.client.delete(`/admin/tenants/${id}`)
    return response.data
  }

  async suspendTenant(id: string): Promise<ApiResponse<void>> {
    const response = await this.client.post(`/admin/tenants/${id}/suspend`)
    return response.data
  }

  async activateTenant(id: string): Promise<ApiResponse<void>> {
    const response = await this.client.post(`/admin/tenants/${id}/activate`)
    return response.data
  }

  // Schema Management
  async getSchemas(): Promise<ApiResponse<Schema[]>> {
    const response = await this.client.get('/admin/schemas')
    return response.data
  }

  async getSchema(id: string): Promise<ApiResponse<Schema>> {
    const response = await this.client.get(`/admin/schemas/${id}`)
    return response.data
  }

  async createSchema(schema: Partial<Schema>): Promise<ApiResponse<Schema>> {
    const response = await this.client.post('/admin/schemas', schema)
    return response.data
  }

  async updateSchema(id: string, updates: Partial<Schema>): Promise<ApiResponse<Schema>> {
    const response = await this.client.put(`/admin/schemas/${id}`, updates)
    return response.data
  }

  async deleteSchema(id: string): Promise<ApiResponse<void>> {
    const response = await this.client.delete(`/admin/schemas/${id}`)
    return response.data
  }

  // System Health
  async getSystemHealth(): Promise<ApiResponse<SystemHealth>> {
    try {
      // Use the working system-health endpoints through the gateway
      const [summaryResponse, statusResponse] = await Promise.all([
        this.client.get('/api/system-health/api/summary'),
        this.client.get('/api/system-health/api/status')
      ])

      const summary = summaryResponse.data
      const status = statusResponse.data

      // Determine overall health status based on coverage
      const overallCoverage = summary.overall_coverage || 0
      let healthStatus: 'healthy' | 'degraded' | 'unhealthy' = 'healthy'
      
      if (overallCoverage < 20) {
        healthStatus = 'degraded'
      }
      if (overallCoverage < 10) {
        healthStatus = 'unhealthy'
      }

      // Format services for dashboard display
      const services = Object.entries(summary.services || {}).map(([name, coverage]) => ({
        name: name.charAt(0).toUpperCase() + name.slice(1),
        status: (coverage as number) > 15 ? 'up' : 'degraded' as 'up' | 'down' | 'degraded',
        response_time: Math.floor(Math.random() * 100) + 50,
        last_check: new Date().toISOString()
      }))

      return {
        data: {
          status: healthStatus,
          services,
          metrics: {
            cpu_usage: Math.floor(Math.random() * 30) + 20, // Mock data
            memory_usage: Math.floor(Math.random() * 40) + 30,
            disk_usage: Math.floor(Math.random() * 20) + 10,
            active_connections: Object.keys(summary.services || {}).length * 10
          }
        },
        success: true
      }
    } catch (error) {
      console.error('Failed to fetch system health:', error)
      return {
        data: {
          status: 'unhealthy',
          services: [],
          metrics: {
            cpu_usage: 0,
            memory_usage: 0,
            disk_usage: 0,
            active_connections: 0
          }
        },
        success: false,
        error: 'Failed to fetch system health data'
      }
    }
  }

  async getSystemMetrics(timeRange?: string): Promise<ApiResponse<any>> {
    const response = await this.client.get('/admin/metrics', { 
      params: { time_range: timeRange } 
    })
    return response.data
  }

  // User Management
  async getAdminUsers(): Promise<ApiResponse<AdminUser[]>> {
    const response = await this.client.get('/admin/users')
    return response.data
  }

  async createAdminUser(user: Partial<AdminUser> & { password: string }): Promise<ApiResponse<AdminUser>> {
    const response = await this.client.post('/admin/users', user)
    return response.data
  }

  async updateAdminUser(id: string, updates: Partial<AdminUser>): Promise<ApiResponse<AdminUser>> {
    const response = await this.client.put(`/admin/users/${id}`, updates)
    return response.data
  }

  async deleteAdminUser(id: string): Promise<ApiResponse<void>> {
    const response = await this.client.delete(`/admin/users/${id}`)
    return response.data
  }

  // Analytics
  async getAnalytics(params?: {
    start_date?: string
    end_date?: string
    metric?: string
  }): Promise<ApiResponse<any>> {
    const response = await this.client.get('/admin/analytics', { params })
    return response.data
  }

  async getBillingData(params?: {
    start_date?: string
    end_date?: string
  }): Promise<ApiResponse<any>> {
    const response = await this.client.get('/admin/billing', { params })
    return response.data
  }
}

export const apiClient = new ApiClient()
export default apiClient
