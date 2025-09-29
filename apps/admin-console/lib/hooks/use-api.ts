import { useState, useEffect, useCallback, useMemo } from 'react'
import { apiClient, ApiError } from '../api-client'

// Generic API hook for data fetching
export function useApi<T>(
  apiCall: () => Promise<T>,
  dependencies: any[] = []
) {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    
    const fetchData = async () => {
      try {
        setLoading(true)
        setError(null)
        const result = await apiCall()
        
        if (!cancelled) {
          setData(result)
        }
      } catch (err) {
        if (!cancelled) {
          let errorMessage = 'An error occurred'
          
          if (err instanceof ApiError) {
            if (err.status === 429) {
              errorMessage = 'Too many requests. The system is automatically retrying with smart backoff.'
            } else {
              errorMessage = err.message
            }
          }
          
          setError(errorMessage)
          console.error('API call failed:', err)
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    fetchData()

    return () => {
      cancelled = true
    }
  }, dependencies)

  const refetch = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const result = await apiCall()
      setData(result)
    } catch (err) {
      let errorMessage = 'An error occurred'
      
      if (err instanceof ApiError) {
        if (err.status === 429) {
          errorMessage = 'Too many requests. The system is automatically retrying with smart backoff.'
        } else {
          errorMessage = err.message
        }
      }
      
      setError(errorMessage)
      console.error('API call failed:', err)
    } finally {
      setLoading(false)
    }
  }, [apiCall])

  return {
    data,
    loading,
    error,
    refetch,
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
  // Extract individual values to avoid object reference issues
  const page = params?.page
  const limit = params?.limit
  const search = params?.search
  
  // Stabilize the API call function with individual dependencies
  const apiCall = useCallback(() => {
    return apiClient.getTenants({ page, limit, search })
  }, [page, limit, search])
  
  return useApi(apiCall, [page, limit, search])
}

// Alternative hook with individual parameters to avoid object reference issues
export function useTenantsWithParams(page?: number, limit?: number, search?: string) {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    
    const fetchData = async () => {
      try {
        setLoading(true)
        setError(null)
        const result = await apiClient.getTenants({ page, limit, search })
        
        if (!cancelled) {
          setData(result)
        }
      } catch (err) {
        if (!cancelled) {
          let errorMessage = 'An error occurred'
          
          if (err instanceof ApiError) {
            if (err.status === 429) {
              errorMessage = 'Too many requests. The system is automatically retrying with smart backoff.'
            } else {
              errorMessage = err.message
            }
          }
          
          setError(errorMessage)
          console.error('API call failed:', err)
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    fetchData()

    return () => {
      cancelled = true
    }
  }, [page, limit, search]) // Simple primitive dependencies

  const refetch = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const result = await apiClient.getTenants({ page, limit, search })
      setData(result)
    } catch (err) {
      let errorMessage = 'An error occurred'
      
      if (err instanceof ApiError) {
        if (err.status === 429) {
          errorMessage = 'Too many requests. The system is automatically retrying with smart backoff.'
        } else {
          errorMessage = err.message
        }
      }
      
      setError(errorMessage)
      console.error('API call failed:', err)
    } finally {
      setLoading(false)
    }
  }, [page, limit, search])

  return {
    data,
    loading,
    error,
    refetch,
  }
}

// Debounced version of useTenants for search to reduce API calls
export function useTenantsDebounced(
  params?: { page?: number; limit?: number; search?: string },
  delay: number = 500
) {
  const [debouncedParams, setDebouncedParams] = useState(params)

  // Use individual values to avoid object reference issues
  const page = params?.page
  const limit = params?.limit
  const search = params?.search

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedParams({ page, limit, search })
    }, delay)

    return () => clearTimeout(timer)
  }, [page, limit, search, delay])

  // Memoize the API call to prevent unnecessary re-renders
  const stableParams = useMemo(() => debouncedParams, [
    debouncedParams?.page,
    debouncedParams?.limit, 
    debouncedParams?.search
  ])

  return useApi(() => apiClient.getTenants(stableParams), [stableParams])
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

// Alternative hook for schemas with individual parameters
export function useSchemasWithParams(page?: number, limit?: number, search?: string) {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    
    const fetchData = async () => {
      try {
        setLoading(true)
        setError(null)
        const result = await apiClient.getSchemas({ page, limit, search })
        
        if (!cancelled) {
          setData(result)
        }
      } catch (err) {
        if (!cancelled) {
          let errorMessage = 'An error occurred'
          
          if (err instanceof ApiError) {
            if (err.status === 429) {
              errorMessage = 'Too many requests. The system is automatically retrying with smart backoff.'
            } else {
              errorMessage = err.message
            }
          }
          
          setError(errorMessage)
          console.error('API call failed:', err)
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    fetchData()

    return () => {
      cancelled = true
    }
  }, [page, limit, search]) // Simple primitive dependencies

  const refetch = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const result = await apiClient.getSchemas({ page, limit, search })
      setData(result)
    } catch (err) {
      let errorMessage = 'An error occurred'
      
      if (err instanceof ApiError) {
        if (err.status === 429) {
          errorMessage = 'Too many requests. The system is automatically retrying with smart backoff.'
        } else {
          errorMessage = err.message
        }
      }
      
      setError(errorMessage)
      console.error('API call failed:', err)
    } finally {
      setLoading(false)
    }
  }, [page, limit, search])

  return {
    data,
    loading,
    error,
    refetch,
  }
}
