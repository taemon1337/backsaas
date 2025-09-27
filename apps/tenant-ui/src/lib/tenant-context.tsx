"use client"

import React, { createContext, useContext, useEffect, useState } from 'react'
import { Tenant, TenantUser, AuthSession } from './types'

interface TenantContextValue {
  tenant: Tenant | null
  user: TenantUser | null
  session: AuthSession | null
  isLoading: boolean
  error: string | null
  
  // Actions
  setTenant: (tenant: Tenant) => void
  setUser: (user: TenantUser) => void
  setSession: (session: AuthSession) => void
  logout: () => void
  retry: () => void
  
  // Utilities
  hasPermission: (permission: string) => boolean
  hasRole: (role: string) => boolean
  canAccessEntity: (entity: string, action: 'read' | 'write' | 'admin') => boolean
}

const TenantContext = createContext<TenantContextValue | undefined>(undefined)

export function useTenant() {
  const context = useContext(TenantContext)
  if (context === undefined) {
    throw new Error('useTenant must be used within a TenantProvider')
  }
  return context
}

interface TenantProviderProps {
  children: React.ReactNode
  initialTenant?: Tenant
  initialUser?: TenantUser
  initialSession?: AuthSession
}

export function TenantProvider({ 
  children, 
  initialTenant, 
  initialUser, 
  initialSession 
}: TenantProviderProps) {
  const [tenant, setTenant] = useState<Tenant | null>(initialTenant || null)
  const [user, setUser] = useState<TenantUser | null>(initialUser || null)
  const [session, setSession] = useState<AuthSession | null>(initialSession || null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [retryCount, setRetryCount] = useState(0)

  // Initialize tenant context from URL or session
  useEffect(() => {
    const initializeTenant = async () => {
      try {
        setIsLoading(true)
        setError(null)

        // Try to get tenant from subdomain or path
        const hostname = window.location.hostname
        const pathname = window.location.pathname
        const tenantSlug = extractTenantFromHostname(hostname) || extractTenantFromPath(pathname)
        
        if (tenantSlug) {
          // Load tenant data
          const tenantData = await loadTenantData(tenantSlug)
          if (tenantData) {
            setTenant(tenantData)
            
            // Apply tenant branding
            applyTenantBranding(tenantData.branding)
          }
        } else {
          // For path-based routing without specific tenant, use a default demo tenant
          const defaultTenant = await loadTenantData('demo')
          if (defaultTenant) {
            setTenant(defaultTenant)
            applyTenantBranding(defaultTenant.branding)
          }
        }

        // Try to restore session from localStorage
        const savedSession = localStorage.getItem('tenant-session')
        const authToken = localStorage.getItem('auth_token')
        
        if (savedSession) {
          try {
            const parsedSession: AuthSession = JSON.parse(savedSession)
            
            // Validate session is not expired
            if (new Date(parsedSession.expiresAt) > new Date()) {
              setSession(parsedSession)
              setUser(parsedSession.user)
              
              // If we don't have tenant data yet, get it from session
              if (!tenant && parsedSession.tenant) {
                setTenant(parsedSession.tenant)
                applyTenantBranding(parsedSession.tenant.branding)
              }
            } else {
              // Session expired, clear it
              localStorage.removeItem('tenant-session')
            }
          } catch (error) {
            console.error('Failed to parse saved session:', error)
            localStorage.removeItem('tenant-session')
          }
        } else if (authToken) {
          // If we have an auth token from the platform login, create a basic session
          // This allows users who logged in through the landing page to access the tenant UI
          try {
            // Decode JWT to get user info (basic decoding, not verification)
            const tokenParts = authToken.split('.')
            if (tokenParts.length === 3) {
              const payload = JSON.parse(atob(tokenParts[1]))
              
              // Create a basic user session
              const basicUser: TenantUser = {
                id: payload.sub || 'unknown',
                email: payload.email || 'unknown@example.com',
                name: `${payload.firstName || 'User'} ${payload.lastName || ''}`.trim(),
                roles: ['user'],
                permissions: ['read'],
                tenantId: payload.tenant_id || tenant?.id || 'default',
                entityAccess: {},
                isActive: true,
                createdAt: new Date().toISOString(),
                updatedAt: new Date().toISOString()
              }
              
              // Create a basic session (expires in 24 hours)
              const basicSession: AuthSession = {
                user: basicUser,
                tenant: tenant || {
                  id: 'default',
                  name: 'Default Tenant',
                  slug: 'default',
                  domain: 'localhost',
                  branding: {
                    primaryColor: '#3B82F6',
                    secondaryColor: '#1E40AF',
                    logo: '',
                    favicon: ''
                  },
                  settings: {},
                  createdAt: new Date().toISOString(),
                  updatedAt: new Date().toISOString()
                },
                token: authToken,
                expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString()
              }
              
              setSession(basicSession)
              setUser(basicUser)
            }
          } catch (error) {
            console.error('Failed to decode auth token:', error)
            localStorage.removeItem('auth_token')
          }
        }
      } catch (error) {
        console.error('Failed to initialize tenant context:', error)
        setError('Failed to load tenant information')
      } finally {
        setIsLoading(false)
      }
    }

    initializeTenant()
  }, [retryCount])

  // Save session to localStorage when it changes
  useEffect(() => {
    if (session) {
      localStorage.setItem('tenant-session', JSON.stringify(session))
    } else {
      localStorage.removeItem('tenant-session')
    }
  }, [session])

  const logout = () => {
    setSession(null)
    setUser(null)
    localStorage.removeItem('tenant-session')
    localStorage.removeItem('auth_token')
    
    // Redirect to login page
    window.location.href = '/login'
  }

  const retry = () => {
    setRetryCount(prev => prev + 1)
    setError(null)
    setIsLoading(true)
  }

  const hasPermission = (permission: string): boolean => {
    if (!user) return false
    return user.permissions.includes(permission) || user.roles.includes('admin')
  }

  const hasRole = (role: string): boolean => {
    if (!user) return false
    return user.roles.includes(role)
  }

  const canAccessEntity = (entity: string, action: 'read' | 'write' | 'admin'): boolean => {
    if (!user) return false
    
    // Admin role can access everything
    if (user.roles.includes('admin')) return true
    
    // Check entity-specific access
    const entityAccess = user.entityAccess[entity]
    if (!entityAccess) return false
    
    // Check access level
    switch (action) {
      case 'read':
        return ['read', 'write', 'admin'].includes(entityAccess)
      case 'write':
        return ['write', 'admin'].includes(entityAccess)
      case 'admin':
        return entityAccess === 'admin'
      default:
        return false
    }
  }

  const value: TenantContextValue = {
    tenant,
    user,
    session,
    isLoading,
    error,
    setTenant,
    setUser,
    setSession,
    logout,
    retry,
    hasPermission,
    hasRole,
    canAccessEntity,
  }
  return (
    <TenantContext.Provider value={value}>
      {children}
    </TenantContext.Provider>
  )
}

// Helper functions
function extractTenantFromHostname(hostname: string): string | null {
  // Handle localhost development
  if (hostname.includes('localhost')) {
    const parts = hostname.split('.')
    if (parts.length > 1 && parts[0] !== 'localhost') {
      return parts[0]
    }
    return null
  }
  
  // Handle production subdomains
  if (hostname.includes('backsaas.dev')) {
    const parts = hostname.split('.')
    if (parts.length > 2) {
      return parts[0]
    }
  }
  
  return null
}

function extractTenantFromPath(pathname: string): string | null {
  // Handle path-based tenant routing like /ui/tenant-slug or /ui/t/tenant-slug
  const pathParts = pathname.split('/').filter(Boolean)
  
  // For /ui path, check if there's a tenant parameter
  if (pathParts[0] === 'ui' && pathParts.length > 1) {
    // Handle /ui/tenant-slug or /ui/t/tenant-slug
    if (pathParts[1] === 't' && pathParts.length > 2) {
      return pathParts[2]
    } else if (pathParts[1] !== 't') {
      return pathParts[1]
    }
  }
  
  // Check for tenant parameter in query string
  const urlParams = new URLSearchParams(window.location.search)
  const tenantParam = urlParams.get('tenant')
  if (tenantParam) {
    return tenantParam
  }
  
  return null
}

async function loadTenantData(tenantSlug: string): Promise<Tenant | null> {
  try {
    // This would make an API call to get tenant data
    // For now, return mock data
    return {
      id: 'tenant-1',
      name: tenantSlug.charAt(0).toUpperCase() + tenantSlug.slice(1),
      slug: tenantSlug,
      branding: {
        colors: {
          primary: '#3b82f6',
          secondary: '#64748b',
          accent: '#f59e0b',
          background: '#ffffff',
          surface: '#f8fafc',
          text: '#1e293b',
        },
        fonts: {
          heading: 'Inter, system-ui, sans-serif',
          body: 'Inter, system-ui, sans-serif',
        },
      },
      settings: {
        timezone: 'UTC',
        dateFormat: 'MM/dd/yyyy',
        currency: 'USD',
        language: 'en',
        features: ['schemas', 'workflows', 'analytics'],
        integrations: {},
      },
      subscription: {
        plan: 'pro',
        status: 'active',
        limits: {
          users: 100,
          storage: 10000,
          apiCalls: 100000,
        },
      },
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }
  } catch (error) {
    console.error('Failed to load tenant data:', error)
    return null
  }
}

function applyTenantBranding(branding: Tenant['branding']) {
  const root = document.documentElement
  
  // Apply color variables
  root.style.setProperty('--tenant-primary', branding.colors.primary)
  root.style.setProperty('--tenant-secondary', branding.colors.secondary)
  root.style.setProperty('--tenant-accent', branding.colors.accent)
  root.style.setProperty('--tenant-background', branding.colors.background)
  root.style.setProperty('--tenant-surface', branding.colors.surface)
  root.style.setProperty('--tenant-text', branding.colors.text)
  
  // Apply font variables
  root.style.setProperty('--tenant-font-heading', branding.fonts.heading)
  root.style.setProperty('--tenant-font-body', branding.fonts.body)
  
  // Apply custom CSS if provided
  if (branding.customCSS) {
    const styleElement = document.createElement('style')
    styleElement.textContent = branding.customCSS
    document.head.appendChild(styleElement)
  }
}
