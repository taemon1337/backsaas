import { useState, useEffect } from 'react'
import { apiClient, ApiError } from '../api-client'

// Generic API hook for data fetching
export function useApi<T>(
  apiCall: () => Promise<T>,
  dependencies: any[] = []
) {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchData = async () => {
    try {
      setLoading(true)
      setError(null)
      const result = await apiCall()
      setData(result)
    } catch (err) {
      const errorMessage = err instanceof ApiError ? err.message : 'An error occurred'
      setError(errorMessage)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, dependencies)

  return {
    data,
    loading,
    error,
    refetch: fetchData,
  }
}

// Specific hooks for common API calls
export function useHealthSummary() {
  return useApi(() => apiClient.getHealthSummary())
}

export function useHealthServices() {
  return useApi(() => apiClient.getHealthServices())
}

export function useHealthStatus() {
  return useApi(() => apiClient.getHealthStatus())
}

export function useTenants(params?: { page?: number; limit?: number; search?: string }) {
  return useApi(() => apiClient.getTenants(params), [params])
}

export function useTenant(id: string) {
  return useApi(() => apiClient.getTenant(id), [id])
}

export function useSchemas(params?: { page?: number; limit?: number; tenantId?: string }) {
  return useApi(() => apiClient.getSchemas(params), [params])
}

export function useUsers(params?: { page?: number; limit?: number; tenantId?: string }) {
  return useApi(() => apiClient.getUsers(params), [params])
}

export function useAnalytics(params?: { timeRange?: string; tenantId?: string; metric?: string }) {
  return useApi(() => apiClient.getAnalytics(params), [params])
}

export function useSettings() {
  return useApi(() => apiClient.getSettings())
}

// Mutation hook for API calls that modify data
export function useApiMutation<TData, TVariables = void>() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const mutate = async (
    apiCall: (variables: TVariables) => Promise<TData>,
    variables: TVariables
  ): Promise<TData | null> => {
    try {
      setLoading(true)
      setError(null)
      const result = await apiCall(variables)
      return result
    } catch (err) {
      const errorMessage = err instanceof ApiError ? err.message : 'An error occurred'
      setError(errorMessage)
      return null
    } finally {
      setLoading(false)
    }
  }

  return {
    mutate,
    loading,
    error,
  }
}
