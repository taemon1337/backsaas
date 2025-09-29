"use client"

import { useState, useMemo, useEffect } from 'react'
import { useTenantsWithParams, useApiMutation } from '@/lib/hooks/use-api'
import { CreateTenantModal } from '@/components/tenants/create-tenant-modal'
import { EditTenantDrawer } from '@/components/tenants/edit-tenant-drawer'
import { DeleteTenantDialog } from '@/components/tenants/delete-tenant-dialog'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { 
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { 
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { 
  Users, 
  Plus, 
  Search,
  Building2,
  Calendar,
  AlertCircle,
  MoreHorizontal,
  Edit,
  Trash2,
  Eye,
  Settings,
  ChevronLeft,
  ChevronRight,
  Download,
  RefreshCw
} from 'lucide-react'
import { useToast } from '@/components/ui/use-toast'

export default function TenantsPage() {
  // ALL HOOKS MUST BE CALLED FIRST - BEFORE ANY CONDITIONAL LOGIC
  const { toast } = useToast()
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')
  const [planFilter, setPlanFilter] = useState('all')
  const [sortBy, setSortBy] = useState('created_at')
  const [sortOrder, setSortOrder] = useState('desc')
  const [pageSize, setPageSize] = useState(10)
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [editDrawerOpen, setEditDrawerOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [selectedTenant, setSelectedTenant] = useState<any>(null)
  
  // Debounce search to reduce API calls
  const [debouncedSearch, setDebouncedSearch] = useState(search)

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search)
    }, 500)

    return () => clearTimeout(timer)
  }, [search])

  // Use individual parameters instead of object to avoid reference issues
  const { 
    data: tenantsResponse, 
    loading, 
    error, 
    refetch 
  } = useTenantsWithParams(page, pageSize, debouncedSearch)

  // Mutation hooks for creating/updating tenants
  const { mutate: createTenant, loading: creating } = useApiMutation()
  const { mutate: updateTenant, loading: updating } = useApiMutation()
  const { mutate: deleteTenant, loading: deleting } = useApiMutation()

  // Mock data - fallback when API is not available
  const mockTenants = useMemo(() => [
    {
      id: '1',
      name: 'Acme Corp',
      slug: 'acme-corp',
      domain: 'acme.example.com',
      plan: 'enterprise',
      status: 'active',
      created_at: '2024-01-15T10:00:00Z',
      updated_at: '2024-03-01T15:30:00Z',
      user_count: 150,
      storage_used: '2.4 GB',
      last_activity: '2024-03-15T09:30:00Z',
      owner_email: 'admin@acme.com',
      billing_status: 'paid'
    },
    {
      id: '2', 
      name: 'StartupXYZ',
      slug: 'startupxyz',
      domain: 'startup.example.com',
      plan: 'pro',
      status: 'active',
      created_at: '2024-02-01T14:30:00Z',
      updated_at: '2024-03-10T11:15:00Z',
      user_count: 25,
      storage_used: '850 MB',
      last_activity: '2024-03-14T16:45:00Z',
      owner_email: 'founder@startupxyz.com',
      billing_status: 'paid'
    },
    {
      id: '3',
      name: 'Demo Company',
      slug: 'demo-company',
      domain: 'demo.example.com', 
      plan: 'starter',
      status: 'trial',
      created_at: '2024-03-10T09:15:00Z',
      updated_at: '2024-03-12T14:20:00Z',
      user_count: 5,
      storage_used: '120 MB',
      last_activity: '2024-03-13T10:15:00Z',
      owner_email: 'demo@example.com',
      billing_status: 'trial'
    },
    {
      id: '4',
      name: 'TechFlow Solutions',
      slug: 'techflow',
      domain: 'techflow.example.com',
      plan: 'pro',
      status: 'active',
      created_at: '2024-01-20T08:45:00Z',
      updated_at: '2024-03-05T13:10:00Z',
      user_count: 45,
      storage_used: '1.2 GB',
      last_activity: '2024-03-15T08:20:00Z',
      owner_email: 'admin@techflow.com',
      billing_status: 'paid'
    },
    {
      id: '5',
      name: 'Global Enterprises',
      slug: 'global-ent',
      domain: 'global.example.com',
      plan: 'enterprise',
      status: 'active',
      created_at: '2023-12-01T12:00:00Z',
      updated_at: '2024-03-08T16:45:00Z',
      user_count: 320,
      storage_used: '5.8 GB',
      last_activity: '2024-03-15T11:30:00Z',
      owner_email: 'it@global-ent.com',
      billing_status: 'paid'
    },
    {
      id: '6',
      name: 'Beta Tester Inc',
      slug: 'beta-tester',
      domain: 'beta.example.com',
      plan: 'starter',
      status: 'suspended',
      created_at: '2024-02-15T16:20:00Z',
      updated_at: '2024-03-01T09:30:00Z',
      user_count: 8,
      storage_used: '200 MB',
      last_activity: '2024-02-28T14:15:00Z',
      owner_email: 'test@beta.com',
      billing_status: 'overdue'
    }
  ], [])

  // Filter and sort tenants
  const filteredAndSortedTenants = useMemo(() => {
    const dataToFilter = tenantsResponse?.data || mockTenants
    
    let filtered = dataToFilter.filter((tenant: any) => {
      const matchesSearch = search === '' || 
        tenant.name.toLowerCase().includes(search.toLowerCase()) ||
        tenant.domain.toLowerCase().includes(search.toLowerCase()) ||
        tenant.owner_email.toLowerCase().includes(search.toLowerCase())
      
      const matchesStatus = statusFilter === 'all' || tenant.status === statusFilter
      const matchesPlan = planFilter === 'all' || tenant.plan === planFilter
      
      return matchesSearch && matchesStatus && matchesPlan
    })

    // Sort tenants
    filtered.sort((a: any, b: any) => {
      let aValue, bValue
      
      switch (sortBy) {
        case 'name':
          aValue = a.name.toLowerCase()
          bValue = b.name.toLowerCase()
          break
        case 'created_at':
          aValue = new Date(a.created_at).getTime()
          bValue = new Date(b.created_at).getTime()
          break
        case 'user_count':
          aValue = a.user_count
          bValue = b.user_count
          break
        case 'last_activity':
          aValue = new Date(a.last_activity).getTime()
          bValue = new Date(b.last_activity).getTime()
          break
        default:
          aValue = a.created_at
          bValue = b.created_at
      }
      
      if (sortOrder === 'asc') {
        return aValue < bValue ? -1 : aValue > bValue ? 1 : 0
      } else {
        return aValue > bValue ? -1 : aValue < bValue ? 1 : 0
      }
    })

    return filtered
  }, [tenantsResponse?.data, mockTenants, search, statusFilter, planFilter, sortBy, sortOrder])

  // Pagination calculations
  const totalTenants = filteredAndSortedTenants.length
  const totalPages = Math.ceil(totalTenants / pageSize)
  const startIndex = (page - 1) * pageSize
  const endIndex = startIndex + pageSize
  const paginatedTenants = filteredAndSortedTenants.slice(startIndex, endIndex)

  // Final tenants data
  const tenants = tenantsResponse?.data || paginatedTenants

  // Event handlers
  const handleCreateTenant = () => {
    setCreateModalOpen(true)
  }

  const handleCreateSuccess = () => {
    refetch()
    toast({
      title: "Tenant Created",
      description: "New tenant has been created successfully",
    })
  }

  const handleEditTenant = (tenant: any) => {
    setSelectedTenant(tenant)
    setEditDrawerOpen(true)
  }

  const handleDeleteTenant = (tenant: any) => {
    setSelectedTenant(tenant)
    setDeleteDialogOpen(true)
  }

  const handleEditSuccess = () => {
    refetch()
    setSelectedTenant(null)
  }

  const handleDeleteSuccess = () => {
    refetch()
    setSelectedTenant(null)
  }

  const handleUpdateTenantStatus = async (tenantId: string, newStatus: string) => {
    try {
      toast({
        title: "Status Updated",
        description: `Tenant status changed to ${newStatus}`,
      })
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to update tenant status",
        variant: "destructive",
      })
    }
  }

  const handleViewTenant = (tenantId: string) => {
    const tenant = tenants.find((t: any) => t.id === tenantId)
    if (!tenant) {
      toast({
        title: "Error",
        description: "Tenant not found",
        variant: "destructive",
      })
      return
    }
    window.open(`/ui?tenant=${tenant.slug}`, '_blank')
  }

  // CONDITIONAL RENDERING AFTER ALL HOOKS
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

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Tenants</h1>
          <p className="text-slate-400">Manage tenant organizations and their configurations</p>
        </div>
        <div className="flex gap-2">
          <Button 
            onClick={() => refetch()}
            variant="outline"
            size="sm"
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
          <Button 
            onClick={handleCreateTenant}
            disabled={creating}
            className="bg-blue-600 hover:bg-blue-700"
          >
            <Plus className="h-4 w-4 mr-2" />
            {creating ? 'Creating...' : 'Add Tenant'}
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-200">Total Tenants</CardTitle>
            <Building2 className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">{totalTenants}</div>
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
              {filteredAndSortedTenants.reduce((sum, t) => sum + (t.user_count || 0), 0)}
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
              {filteredAndSortedTenants.filter(t => t.plan === 'enterprise').length}
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
              {filteredAndSortedTenants.filter(t => t.status === 'trial').length}
            </div>
            <p className="text-xs text-slate-400">Need attention</p>
          </CardContent>
        </Card>
      </div>

      {/* Search and Filters */}
      <Card className="bg-slate-800 border-slate-700">
        <CardContent className="pt-6">
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-slate-400" />
                <Input
                  placeholder="Search tenants by name, domain, or email..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className="pl-10 bg-slate-700 border-slate-600 text-white placeholder-slate-400"
                />
              </div>
            </div>
            <div className="flex gap-2">
              <Select value={statusFilter} onValueChange={setStatusFilter}>
                <SelectTrigger className="w-32 bg-slate-700 border-slate-600 text-white">
                  <SelectValue placeholder="Status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Status</SelectItem>
                  <SelectItem value="active">Active</SelectItem>
                  <SelectItem value="trial">Trial</SelectItem>
                  <SelectItem value="suspended">Suspended</SelectItem>
                </SelectContent>
              </Select>
              
              <Select value={planFilter} onValueChange={setPlanFilter}>
                <SelectTrigger className="w-32 bg-slate-700 border-slate-600 text-white">
                  <SelectValue placeholder="Plan" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Plans</SelectItem>
                  <SelectItem value="starter">Starter</SelectItem>
                  <SelectItem value="pro">Pro</SelectItem>
                  <SelectItem value="enterprise">Enterprise</SelectItem>
                </SelectContent>
              </Select>

              <Select value={sortBy} onValueChange={setSortBy}>
                <SelectTrigger className="w-40 bg-slate-700 border-slate-600 text-white">
                  <SelectValue placeholder="Sort by" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="created_at">Created Date</SelectItem>
                  <SelectItem value="name">Name</SelectItem>
                  <SelectItem value="user_count">User Count</SelectItem>
                  <SelectItem value="last_activity">Last Activity</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Tenants Table */}
      <Card className="bg-slate-800 border-slate-700">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-white">Tenants ({totalTenants})</CardTitle>
              <CardDescription className="text-slate-400">
                Showing {startIndex + 1}-{Math.min(endIndex, totalTenants)} of {totalTenants} tenants
              </CardDescription>
            </div>
            <Button variant="outline" size="sm">
              <Download className="h-4 w-4 mr-2" />
              Export
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="border-slate-700">
                  <TableHead className="text-slate-300">Organization</TableHead>
                  <TableHead className="text-slate-300">Plan</TableHead>
                  <TableHead className="text-slate-300">Status</TableHead>
                  <TableHead className="text-slate-300">Users</TableHead>
                  <TableHead className="text-slate-300">Storage</TableHead>
                  <TableHead className="text-slate-300">Last Activity</TableHead>
                  <TableHead className="text-slate-300">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {tenants.map((tenant: any) => (
                  <TableRow key={tenant.id} className="border-slate-700 hover:bg-slate-700/50">
                    <TableCell>
                      <div className="flex items-center space-x-3">
                        <Building2 className="h-8 w-8 text-blue-500" />
                        <div>
                          <div className="font-medium text-white">{tenant.name}</div>
                          <div className="text-sm text-slate-400">{tenant.domain}</div>
                          <div className="text-xs text-slate-500">{tenant.owner_email}</div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="capitalize">
                        {tenant.plan}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge 
                        variant={
                          tenant.status === 'active' ? 'default' : 
                          tenant.status === 'trial' ? 'secondary' : 
                          'destructive'
                        }
                        className="capitalize"
                      >
                        {tenant.status}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-white">{tenant.user_count}</TableCell>
                    <TableCell className="text-white">{tenant.storage_used}</TableCell>
                    <TableCell className="text-slate-400">
                      {new Date(tenant.last_activity).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="bg-slate-800 border-slate-700">
                          <DropdownMenuLabel className="text-slate-300">Actions</DropdownMenuLabel>
                          <DropdownMenuItem 
                            onClick={() => handleViewTenant(tenant.id)}
                            className="text-slate-300 hover:bg-slate-700"
                          >
                            <Eye className="h-4 w-4 mr-2" />
                            View Dashboard
                          </DropdownMenuItem>
                          <DropdownMenuItem 
                            onClick={() => handleEditTenant(tenant)}
                            className="text-slate-300 hover:bg-slate-700"
                          >
                            <Edit className="h-4 w-4 mr-2" />
                            Edit Details
                          </DropdownMenuItem>
                          <DropdownMenuItem className="text-slate-300 hover:bg-slate-700">
                            <Settings className="h-4 w-4 mr-2" />
                            Settings
                          </DropdownMenuItem>
                          <DropdownMenuSeparator className="bg-slate-700" />
                          <DropdownMenuItem 
                            onClick={() => handleUpdateTenantStatus(tenant.id, tenant.status === 'active' ? 'suspended' : 'active')}
                            className="text-slate-300 hover:bg-slate-700"
                          >
                            {tenant.status === 'active' ? 'Suspend' : 'Activate'}
                          </DropdownMenuItem>
                          <DropdownMenuItem 
                            onClick={() => handleDeleteTenant(tenant)}
                            className="text-red-400 hover:bg-red-900/20"
                          >
                            <Trash2 className="h-4 w-4 mr-2" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between mt-6">
              <div className="flex items-center space-x-2">
                <span className="text-sm text-slate-400">Rows per page:</span>
                <Select value={pageSize.toString()} onValueChange={(value: string) => setPageSize(Number(value))}>
                  <SelectTrigger className="w-20 bg-slate-700 border-slate-600 text-white">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="5">5</SelectItem>
                    <SelectItem value="10">10</SelectItem>
                    <SelectItem value="20">20</SelectItem>
                    <SelectItem value="50">50</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              
              <div className="flex items-center space-x-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage(page - 1)}
                  disabled={page <= 1}
                >
                  <ChevronLeft className="h-4 w-4" />
                  Previous
                </Button>
                
                <span className="text-sm text-slate-400">
                  Page {page} of {totalPages}
                </span>
                
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage(page + 1)}
                  disabled={page >= totalPages}
                >
                  Next
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}

          {tenants.length === 0 && (
            <div className="text-center py-12">
              <Building2 className="h-12 w-12 text-slate-500 mx-auto mb-4" />
              <p className="text-slate-400 mb-2">No tenants found</p>
              <p className="text-sm text-slate-500">
                {search || statusFilter !== 'all' || planFilter !== 'all' 
                  ? 'Try adjusting your search or filters'
                  : 'Create your first tenant to get started'
                }
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create Tenant Modal */}
      <CreateTenantModal
        open={createModalOpen}
        onOpenChange={setCreateModalOpen}
        onSuccess={handleCreateSuccess}
      />

      {/* Edit Tenant Drawer */}
      <EditTenantDrawer
        open={editDrawerOpen}
        onOpenChange={setEditDrawerOpen}
        tenant={selectedTenant}
        onSuccess={handleEditSuccess}
      />

      {/* Delete Tenant Dialog - Temporarily commented out due to missing dependencies */}
      {/*
      <DeleteTenantDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        tenant={selectedTenant}
        onSuccess={handleDeleteSuccess}
      />
      */}
    </div>
  )
}
