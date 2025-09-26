"use client"

import { useState } from 'react'
import { useTenants, useApiMutation } from '@/lib/hooks/use-api'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { 
  Users, 
  Plus, 
  Search,
  Building2,
  Calendar,
  AlertCircle
} from 'lucide-react'
import { useToast } from '@/components/ui/use-toast'

export default function TenantsPage() {
  const { toast } = useToast()
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  
  // Fetch tenants using the API hook
  const { 
    data: tenantsResponse, 
    loading, 
    error, 
    refetch 
  } = useTenants({ page, limit: 10, search })

  // Mutation hook for creating/updating tenants
  const { mutate: createTenant, loading: creating } = useApiMutation()

  const handleCreateTenant = async () => {
    const { apiClient } = await import('@/lib/api-client')
    
    const result = await createTenant(
      (data) => apiClient.createTenant(data),
      {
        name: 'New Tenant',
        domain: 'new-tenant.example.com',
        plan: 'starter'
      }
    )
    
    if (result) {
      toast({
        title: "Tenant Created",
        description: "New tenant has been created successfully",
      })
      refetch()
    }
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-white">Tenants</h1>
        </div>
        <div className="flex items-center justify-center py-12">
          <div className="text-slate-400">Loading tenants...</div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-white">Tenants</h1>
        </div>
        <Card className="bg-slate-800 border-slate-700">
          <CardContent className="flex items-center justify-center py-12">
            <div className="text-center">
              <AlertCircle className="h-12 w-12 text-red-500 mx-auto mb-4" />
              <p className="text-slate-400 mb-4">{error}</p>
              <Button onClick={() => refetch()} variant="outline">
                Try Again
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  // Mock data since the API endpoints don't exist yet
  const mockTenants = [
    {
      id: '1',
      name: 'Acme Corp',
      domain: 'acme.example.com',
      plan: 'enterprise',
      status: 'active',
      created_at: '2024-01-15T10:00:00Z',
      user_count: 150
    },
    {
      id: '2', 
      name: 'StartupXYZ',
      domain: 'startup.example.com',
      plan: 'pro',
      status: 'active',
      created_at: '2024-02-01T14:30:00Z',
      user_count: 25
    },
    {
      id: '3',
      name: 'Demo Company',
      domain: 'demo.example.com', 
      plan: 'starter',
      status: 'trial',
      created_at: '2024-03-10T09:15:00Z',
      user_count: 5
    }
  ]

  const tenants = tenantsResponse?.data || mockTenants

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Tenants</h1>
          <p className="text-slate-400">Manage tenant organizations and their configurations</p>
        </div>
        <Button 
          onClick={handleCreateTenant}
          disabled={creating}
          className="bg-blue-600 hover:bg-blue-700"
        >
          <Plus className="h-4 w-4 mr-2" />
          {creating ? 'Creating...' : 'Add Tenant'}
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-200">Total Tenants</CardTitle>
            <Building2 className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">{tenants.length}</div>
            <p className="text-xs text-slate-400">Active organizations</p>
          </CardContent>
        </Card>

        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-200">Active Users</CardTitle>
            <Users className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">
              {tenants.reduce((sum, t) => sum + (t.user_count || 0), 0)}
            </div>
            <p className="text-xs text-slate-400">Across all tenants</p>
          </CardContent>
        </Card>

        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-200">Enterprise Plans</CardTitle>
            <Calendar className="h-4 w-4 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">
              {tenants.filter(t => t.plan === 'enterprise').length}
            </div>
            <p className="text-xs text-slate-400">Premium subscriptions</p>
          </CardContent>
        </Card>

        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-200">Trial Accounts</CardTitle>
            <AlertCircle className="h-4 w-4 text-yellow-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">
              {tenants.filter(t => t.status === 'trial').length}
            </div>
            <p className="text-xs text-slate-400">Need attention</p>
          </CardContent>
        </Card>
      </div>

      {/* Tenants List */}
      <Card className="bg-slate-800 border-slate-700">
        <CardHeader>
          <CardTitle className="text-white">All Tenants</CardTitle>
          <CardDescription className="text-slate-400">
            Manage tenant organizations and their settings
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {tenants.map((tenant) => (
              <div key={tenant.id} className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
                <div className="flex items-center space-x-4">
                  <Building2 className="h-8 w-8 text-blue-500" />
                  <div>
                    <h3 className="font-medium text-white">{tenant.name}</h3>
                    <p className="text-sm text-slate-400">{tenant.domain}</p>
                    <p className="text-xs text-slate-500">
                      Created {new Date(tenant.created_at).toLocaleDateString()}
                    </p>
                  </div>
                </div>
                
                <div className="flex items-center space-x-4">
                  <div className="text-right">
                    <div className="text-sm font-medium text-white">
                      {tenant.user_count} users
                    </div>
                    <div className="text-xs text-slate-400 capitalize">
                      {tenant.plan} plan
                    </div>
                  </div>
                  
                  <Badge 
                    variant={tenant.status === 'active' ? 'default' : 'secondary'}
                    className="min-w-[70px] justify-center"
                  >
                    {tenant.status}
                  </Badge>
                </div>
              </div>
            ))}
            
            {tenants.length === 0 && (
              <div className="text-center py-8">
                <Building2 className="h-12 w-12 text-slate-500 mx-auto mb-4" />
                <p className="text-slate-400">No tenants found</p>
                <p className="text-sm text-slate-500">Create your first tenant to get started</p>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
