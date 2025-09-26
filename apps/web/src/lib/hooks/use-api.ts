"use client"

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient, UseQueryOptions, UseMutationOptions } from '@tanstack/react-query'
import { apiClient, Schema, ValidationResult, MigrationPlan, PaginatedResponse } from '../api-client'

// ============================================================================
// SCHEMA MANAGEMENT HOOKS
// ============================================================================

export function useSchemas(params?: {
  page?: number
  limit?: number
  search?: string
  tenant_id?: string
}) {
  return useQuery({
    queryKey: ['schemas', params],
    queryFn: () => apiClient.getSchemas(params),
    staleTime: 30000, // 30 seconds
  })
}

export function useSchema(id: string, enabled = true) {
  return useQuery({
    queryKey: ['schema', id],
    queryFn: () => apiClient.getSchema(id),
    enabled: enabled && !!id,
    staleTime: 60000, // 1 minute
  })
}

export function useCreateSchema() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: (schema: Omit<Schema, 'id' | 'metadata'>) => 
      apiClient.createSchema(schema),
    onSuccess: () => {
      // Invalidate schemas list to refetch
      queryClient.invalidateQueries({ queryKey: ['schemas'] })
    },
  })
}

export function useUpdateSchema() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: ({ id, schema }: { id: string; schema: Partial<Schema> }) =>
      apiClient.updateSchema(id, schema),
    onSuccess: (data, variables) => {
      // Update the specific schema in cache
      queryClient.setQueryData(['schema', variables.id], data)
      // Invalidate schemas list
      queryClient.invalidateQueries({ queryKey: ['schemas'] })
    },
  })
}

export function useDeleteSchema() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: (id: string) => apiClient.deleteSchema(id),
    onSuccess: (_, id) => {
      // Remove from cache
      queryClient.removeQueries({ queryKey: ['schema', id] })
      // Invalidate schemas list
      queryClient.invalidateQueries({ queryKey: ['schemas'] })
    },
  })
}

export function usePublishSchema() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: (id: string) => apiClient.publishSchema(id),
    onSuccess: (_, id) => {
      // Invalidate the specific schema to refetch updated status
      queryClient.invalidateQueries({ queryKey: ['schema', id] })
      // Invalidate schemas list
      queryClient.invalidateQueries({ queryKey: ['schemas'] })
    },
  })
}

// ============================================================================
// SCHEMA VALIDATION HOOKS
// ============================================================================

export function useValidateSchema() {
  return useMutation({
    mutationFn: (schema: Schema) => apiClient.validateSchema(schema),
  })
}

export function useValidateSchemaYaml() {
  return useMutation({
    mutationFn: (yamlContent: string) => apiClient.validateSchemaYaml(yamlContent),
  })
}

// ============================================================================
// MIGRATION PLANNING HOOKS
// ============================================================================

export function useGenerateMigrationPlan() {
  return useMutation({
    mutationFn: ({ fromSchemaId, toSchema }: { fromSchemaId: string; toSchema: Schema }) =>
      apiClient.generateMigrationPlan(fromSchemaId, toSchema),
  })
}

export function useMigrationPlan(planId: string, enabled = true) {
  return useQuery({
    queryKey: ['migration-plan', planId],
    queryFn: () => apiClient.getMigrationPlan(planId),
    enabled: enabled && !!planId,
    refetchInterval: 5000, // Refetch every 5 seconds for status updates
  })
}

export function useExecuteMigrationPlan() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: (planId: string) => apiClient.executeMigrationPlan(planId),
    onSuccess: (_, planId) => {
      // Invalidate migration plan to get updated status
      queryClient.invalidateQueries({ queryKey: ['migration-plan', planId] })
    },
  })
}

export function useMigrationStatus(planId: string, enabled = true) {
  return useQuery({
    queryKey: ['migration-status', planId],
    queryFn: () => apiClient.getMigrationStatus(planId),
    enabled: enabled && !!planId,
    refetchInterval: 2000, // Frequent updates during execution
  })
}

export function useRollbackMigration() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: (planId: string) => apiClient.rollbackMigration(planId),
    onSuccess: (_, planId) => {
      // Invalidate migration plan to get updated status
      queryClient.invalidateQueries({ queryKey: ['migration-plan', planId] })
    },
  })
}

// ============================================================================
// SQL GENERATION HOOKS
// ============================================================================

export function useGenerateSql() {
  return useMutation({
    mutationFn: (schema: Schema) => apiClient.generateSql(schema),
  })
}

export function usePreviewSql() {
  return useMutation({
    mutationFn: (schema: Schema) => apiClient.previewSql(schema),
  })
}

// ============================================================================
// EVENT STREAMING HOOKS
// ============================================================================

export function useSchemaEvents(params?: {
  schema_id?: string
  event_type?: string
  since?: string
  limit?: number
}) {
  return useQuery({
    queryKey: ['schema-events', params],
    queryFn: () => apiClient.getSchemaEvents(params),
    refetchInterval: 10000, // Refetch every 10 seconds
  })
}

// ============================================================================
// SYSTEM HOOKS
// ============================================================================

export function useHealthCheck() {
  return useQuery({
    queryKey: ['health-check'],
    queryFn: () => apiClient.healthCheck(),
    refetchInterval: 30000, // Check every 30 seconds
  })
}

export function useSystemInfo() {
  return useQuery({
    queryKey: ['system-info'],
    queryFn: () => apiClient.getSystemInfo(),
    staleTime: 300000, // 5 minutes
  })
}

// ============================================================================
// UTILITY HOOKS
// ============================================================================

// Hook for real-time schema validation
export function useRealtimeValidation(schema: Schema | null, debounceMs = 1000) {
  const validateMutation = useValidateSchema()
  
  // Debounced validation effect would go here
  // This is a simplified version
  const validate = () => {
    if (schema) {
      validateMutation.mutate(schema)
    }
  }
  
  return {
    validate,
    validationResult: validateMutation.data,
    isValidating: validateMutation.isPending,
    validationError: validateMutation.error,
  }
}

// Hook for managing schema drafts
export function useSchemaDraft(initialSchema?: Schema) {
  const [draft, setDraft] = useState<Schema | null>(initialSchema || null)
  const [hasChanges, setHasChanges] = useState(false)
  
  const updateDraft = (updates: Partial<Schema>) => {
    setDraft(prev => prev ? { ...prev, ...updates } : null)
    setHasChanges(true)
  }
  
  const resetDraft = () => {
    setDraft(initialSchema || null)
    setHasChanges(false)
  }
  
  const saveDraft = () => {
    // Save to localStorage or session storage
    if (draft) {
      localStorage.setItem('schema-draft', JSON.stringify(draft))
    }
  }
  
  const loadDraft = () => {
    const saved = localStorage.getItem('schema-draft')
    if (saved) {
      try {
        const parsed = JSON.parse(saved)
        setDraft(parsed)
        setHasChanges(true)
      } catch (error) {
        console.error('Failed to load draft:', error)
      }
    }
  }
  
  return {
    draft,
    hasChanges,
    updateDraft,
    resetDraft,
    saveDraft,
    loadDraft,
  }
}
