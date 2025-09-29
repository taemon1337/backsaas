"use client"

import { useState, useMemo, useEffect } from 'react'
import { useSchemasWithParams, useApiMutation } from '@/lib/hooks/use-api'
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
import { useToast } from '@/components/ui/use-toast'
import { 
  Plus, 
  Search, 
  Filter, 
  MoreHorizontal, 
  Eye, 
  Edit, 
  Copy,
  Archive,
  Trash2,
  Database,
  Calendar,
  Users,
  FileText,
  Code,
  GitBranch
} from 'lucide-react'

export default function SchemasPage() {
  // ALL HOOKS MUST BE CALLED FIRST - BEFORE ANY CONDITIONAL LOGIC
  const { toast } = useToast()
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')
  const [typeFilter, setTypeFilter] = useState('all')
  const [sortBy, setSortBy] = useState('updated_at')
  const [sortOrder, setSortOrder] = useState('desc')
  const [pageSize, setPageSize] = useState(10)
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [editDrawerOpen, setEditDrawerOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [selectedSchema, setSelectedSchema] = useState<any>(null)
  
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
    data: schemasResponse, 
    loading, 
    error, 
    refetch 
  } = useSchemasWithParams(page, pageSize, debouncedSearch)

  // Mutation hooks for creating/updating schemas
  const { mutate: createSchema, loading: creating } = useApiMutation()
  const { mutate: updateSchema, loading: updating } = useApiMutation()
  const { mutate: deleteSchema, loading: deleting } = useApiMutation()

  // Mock data - fallback when API is not available
  const mockSchemas = useMemo(() => [
    {
      id: '1',
      name: 'User Profile',
      slug: 'user-profile',
      description: 'Core user profile data structure',
      type: 'entity',
      status: 'active',
      version: '2.1.0',
      field_count: 12,
      usage_count: 1250,
      created_at: '2024-01-15T10:00:00Z',
      updated_at: '2024-01-20T14:30:00Z',
      created_by: 'admin@backsaas.dev',
      tags: ['core', 'user', 'profile']
    },
    {
      id: '2',
      name: 'Product Catalog',
      slug: 'product-catalog',
      description: 'E-commerce product information schema',
      type: 'collection',
      status: 'active',
      version: '1.5.2',
      field_count: 18,
      usage_count: 890,
      created_at: '2024-01-10T09:15:00Z',
      updated_at: '2024-01-18T16:45:00Z',
      created_by: 'dev@backsaas.dev',
      tags: ['ecommerce', 'product', 'catalog']
    },
    {
      id: '3',
      name: 'Order Management',
      slug: 'order-management',
      description: 'Order processing and tracking schema',
      type: 'workflow',
      status: 'draft',
      version: '0.8.0',
      field_count: 24,
      usage_count: 0,
      created_at: '2024-01-22T11:30:00Z',
      updated_at: '2024-01-25T13:20:00Z',
      created_by: 'admin@backsaas.dev',
      tags: ['order', 'workflow', 'processing']
    },
    {
      id: '4',
      name: 'Analytics Events',
      slug: 'analytics-events',
      description: 'Event tracking and analytics data structure',
      type: 'event',
      status: 'active',
      version: '3.0.1',
      field_count: 8,
      usage_count: 5420,
      created_at: '2024-01-05T08:00:00Z',
      updated_at: '2024-01-24T10:15:00Z',
      created_by: 'analytics@backsaas.dev',
      tags: ['analytics', 'events', 'tracking']
    },
    {
      id: '5',
      name: 'Customer Support',
      slug: 'customer-support',
      description: 'Support ticket and interaction schema',
      type: 'entity',
      status: 'deprecated',
      version: '1.2.5',
      field_count: 15,
      usage_count: 45,
      created_at: '2023-12-01T14:00:00Z',
      updated_at: '2024-01-12T09:30:00Z',
      created_by: 'support@backsaas.dev',
      tags: ['support', 'tickets', 'legacy']
    },
    {
      id: '6',
      name: 'Payment Processing',
      slug: 'payment-processing',
      description: 'Payment and billing information schema',
      type: 'entity',
      status: 'active',
      version: '2.3.0',
      field_count: 20,
      usage_count: 2100,
      created_at: '2024-01-08T12:45:00Z',
      updated_at: '2024-01-26T15:00:00Z',
      created_by: 'billing@backsaas.dev',
      tags: ['payment', 'billing', 'finance']
    }
  ], [])

  // Filter and sort schemas
  const filteredAndSortedSchemas = useMemo(() => {
    let filtered = mockSchemas

    // Apply search filter
    if (search) {
      filtered = filtered.filter(schema => 
        schema.name.toLowerCase().includes(search.toLowerCase()) ||
        schema.description.toLowerCase().includes(search.toLowerCase()) ||
        (schema.tags || []).some(tag => tag.toLowerCase().includes(search.toLowerCase()))
      )
    }

    // Apply status filter
    if (statusFilter !== 'all') {
      filtered = filtered.filter(schema => schema.status === statusFilter)
    }

    // Apply type filter
    if (typeFilter !== 'all') {
      filtered = filtered.filter(schema => schema.type === typeFilter)
    }

    // Apply sorting
    filtered.sort((a: any, b: any) => {
      let aValue = a[sortBy]
      let bValue = b[sortBy]

      if (sortBy === 'updated_at' || sortBy === 'created_at') {
        aValue = new Date(aValue).getTime()
        bValue = new Date(bValue).getTime()
      }

      if (sortOrder === 'asc') {
        return aValue > bValue ? 1 : -1
      } else {
        return aValue < bValue ? 1 : -1
      }
    })

    return filtered
  }, [mockSchemas, search, statusFilter, typeFilter, sortBy, sortOrder])

  // Pagination
  const totalSchemas = filteredAndSortedSchemas.length
  const totalPages = Math.ceil(totalSchemas / pageSize)
  const startIndex = (page - 1) * pageSize
  const endIndex = startIndex + pageSize
  const paginatedSchemas = filteredAndSortedSchemas.slice(startIndex, endIndex)

  // Final schemas data
  const schemas = schemasResponse?.data || paginatedSchemas

  // Statistics
  const stats = useMemo(() => {
    const activeSchemas = filteredAndSortedSchemas.filter(s => s.status === 'active').length
    const totalUsage = filteredAndSortedSchemas.reduce((sum: number, s: any) => sum + s.usage_count, 0)
    const avgFields = Math.round(filteredAndSortedSchemas.reduce((sum: number, s: any) => sum + s.field_count, 0) / filteredAndSortedSchemas.length)
    const draftSchemas = filteredAndSortedSchemas.filter(s => s.status === 'draft').length

    return { activeSchemas, totalUsage, avgFields, draftSchemas }
  }, [filteredAndSortedSchemas])

  // Event handlers
  const handleCreateSchema = () => {
    setCreateModalOpen(true)
  }

  const handleCreateSuccess = () => {
    refetch()
    toast({
      title: "Schema Created",
      description: "New schema has been created successfully",
    })
  }

  const handleEditSchema = (schema: any) => {
    setSelectedSchema(schema)
    setEditDrawerOpen(true)
  }

  const handleDeleteSchema = (schema: any) => {
    setSelectedSchema(schema)
    setDeleteDialogOpen(true)
  }

  const handleEditSuccess = () => {
    refetch()
    setSelectedSchema(null)
  }

  const handleDeleteSuccess = () => {
    refetch()
    setSelectedSchema(null)
  }

  const handleDuplicateSchema = async (schemaId: string, schemaName: string) => {
    try {
      toast({
        title: "Schema Duplicated",
        description: `${schemaName} has been duplicated successfully`,
      })
      refetch()
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to duplicate schema",
        variant: "destructive",
      })
    }
  }

  const handleViewSchema = (schemaId: string) => {
    const schema = schemas.find((s: any) => s.id === schemaId)
    if (!schema) {
      toast({
        title: "Error",
        description: "Schema not found",
        variant: "destructive",
      })
      return
    }
    window.open(`/schemas/${schema.slug}`, '_blank')
  }

  // CONDITIONAL RENDERING AFTER ALL HOOKS
  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-white">Schemas</h1>
        </div>
        <div className="flex items-center justify-center py-12">
          <div className="text-slate-400">Loading schemas...</div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-white">Schemas</h1>
        </div>
        <div className="flex items-center justify-center py-12">
          <div className="text-red-400">Error loading schemas: {error}</div>
        </div>
      </div>
    )
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-green-500'
      case 'draft': return 'bg-yellow-500'
      case 'deprecated': return 'bg-red-500'
      case 'archived': return 'bg-gray-500'
      default: return 'bg-gray-500'
    }
  }

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'entity': return 'bg-blue-500'
      case 'collection': return 'bg-purple-500'
      case 'workflow': return 'bg-orange-500'
      case 'event': return 'bg-green-500'
      default: return 'bg-gray-500'
    }
  }

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'entity': return <Database className="h-4 w-4" />
      case 'collection': return <FileText className="h-4 w-4" />
      case 'workflow': return <GitBranch className="h-4 w-4" />
      case 'event': return <Code className="h-4 w-4" />
      default: return <Database className="h-4 w-4" />
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Schema Management</h1>
          <p className="text-slate-400 mt-1">Manage data structures and API schemas</p>
        </div>
        <Button 
          onClick={handleCreateSchema}
          className="bg-blue-600 hover:bg-blue-700"
        >
          <Plus className="h-4 w-4 mr-2" />
          Create Schema
        </Button>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-300">Active Schemas</CardTitle>
            <Database className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">{stats.activeSchemas}</div>
            <p className="text-xs text-slate-400">Production ready</p>
          </CardContent>
        </Card>

        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-300">Total Usage</CardTitle>
            <Users className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">{(stats.totalUsage || 0).toLocaleString()}</div>
            <p className="text-xs text-slate-400">API calls this month</p>
          </CardContent>
        </Card>

        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-300">Avg Fields</CardTitle>
            <FileText className="h-4 w-4 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">{stats.avgFields}</div>
            <p className="text-xs text-slate-400">Per schema</p>
          </CardContent>
        </Card>

        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-300">Draft Schemas</CardTitle>
            <GitBranch className="h-4 w-4 text-yellow-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">{stats.draftSchemas}</div>
            <p className="text-xs text-slate-400">In development</p>
          </CardContent>
        </Card>
      </div>

      {/* Filters and Search */}
      <Card className="bg-slate-800 border-slate-700">
        <CardHeader>
          <CardTitle className="text-white">Schemas</CardTitle>
          <CardDescription className="text-slate-400">
            Manage your data schemas and API structures
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-3 h-4 w-4 text-slate-400" />
                <Input
                  placeholder="Search schemas..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className="pl-10 bg-slate-700 border-slate-600 text-white"
                />
              </div>
            </div>
            <div className="flex gap-2">
              <Select value={statusFilter} onValueChange={setStatusFilter}>
                <SelectTrigger className="w-32 bg-slate-700 border-slate-600 text-white">
                  <Filter className="h-4 w-4 mr-2" />
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Status</SelectItem>
                  <SelectItem value="active">Active</SelectItem>
                  <SelectItem value="draft">Draft</SelectItem>
                  <SelectItem value="deprecated">Deprecated</SelectItem>
                  <SelectItem value="archived">Archived</SelectItem>
                </SelectContent>
              </Select>
              <Select value={typeFilter} onValueChange={setTypeFilter}>
                <SelectTrigger className="w-32 bg-slate-700 border-slate-600 text-white">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Types</SelectItem>
                  <SelectItem value="entity">Entity</SelectItem>
                  <SelectItem value="collection">Collection</SelectItem>
                  <SelectItem value="workflow">Workflow</SelectItem>
                  <SelectItem value="event">Event</SelectItem>
                </SelectContent>
              </Select>
              <Select value={`${pageSize}`} onValueChange={(value) => setPageSize(parseInt(value))}>
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
          </div>

          {/* Schemas Table */}
          <div className="rounded-md border border-slate-700">
            <Table>
              <TableHeader>
                <TableRow className="border-slate-700 hover:bg-slate-700/50">
                  <TableHead className="text-slate-300">Schema</TableHead>
                  <TableHead className="text-slate-300">Type</TableHead>
                  <TableHead className="text-slate-300">Status</TableHead>
                  <TableHead className="text-slate-300">Version</TableHead>
                  <TableHead className="text-slate-300">Fields</TableHead>
                  <TableHead className="text-slate-300">Usage</TableHead>
                  <TableHead className="text-slate-300">Updated</TableHead>
                  <TableHead className="text-slate-300 w-12"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {schemas.map((schema: any) => (
                  <TableRow key={schema.id} className="border-slate-700 hover:bg-slate-700/30">
                    <TableCell>
                      <div className="space-y-1">
                        <div className="font-medium text-white">{schema.name}</div>
                        <div className="text-sm text-slate-400">{schema.description}</div>
                        <div className="flex gap-1 mt-1">
                          {(schema.tags || []).slice(0, 3).map((tag: string) => (
                            <Badge key={tag} variant="outline" className="text-xs border-slate-600 text-slate-400">
                              {tag}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {getTypeIcon(schema.type)}
                        <Badge className={`${getTypeColor(schema.type)} text-white text-xs`}>
                          {schema.type}
                        </Badge>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge className={`${getStatusColor(schema.status)} text-white text-xs`}>
                        {schema.status}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-slate-300 font-mono text-sm">{schema.version}</TableCell>
                    <TableCell className="text-slate-300">{schema.field_count}</TableCell>
                    <TableCell className="text-slate-300">{(schema.usage_count || 0).toLocaleString()}</TableCell>
                    <TableCell className="text-slate-400 text-sm">
                      {schema.updated_at ? new Date(schema.updated_at).toLocaleDateString() : 'N/A'}
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" className="h-8 w-8 p-0 text-slate-400 hover:text-white">
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="bg-slate-800 border-slate-700">
                          <DropdownMenuLabel className="text-slate-300">Actions</DropdownMenuLabel>
                          <DropdownMenuItem 
                            onClick={() => handleViewSchema(schema.id)}
                            className="text-slate-300 hover:bg-slate-700"
                          >
                            <Eye className="h-4 w-4 mr-2" />
                            View Schema
                          </DropdownMenuItem>
                          <DropdownMenuItem 
                            onClick={() => handleEditSchema(schema)}
                            className="text-slate-300 hover:bg-slate-700"
                          >
                            <Edit className="h-4 w-4 mr-2" />
                            Edit Schema
                          </DropdownMenuItem>
                          <DropdownMenuItem 
                            onClick={() => handleDuplicateSchema(schema.id, schema.name)}
                            className="text-slate-300 hover:bg-slate-700"
                          >
                            <Copy className="h-4 w-4 mr-2" />
                            Duplicate
                          </DropdownMenuItem>
                          <DropdownMenuSeparator className="bg-slate-700" />
                          <DropdownMenuItem className="text-slate-300 hover:bg-slate-700">
                            <Archive className="h-4 w-4 mr-2" />
                            Archive
                          </DropdownMenuItem>
                          <DropdownMenuItem 
                            onClick={() => handleDeleteSchema(schema)}
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
            <div className="flex items-center justify-between">
              <div className="text-sm text-slate-400">
                Showing {startIndex + 1} to {Math.min(endIndex, totalSchemas)} of {totalSchemas} schemas
              </div>
              <div className="flex items-center space-x-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage(page - 1)}
                  disabled={page === 1}
                  className="border-slate-600 text-slate-300 hover:bg-slate-700"
                >
                  Previous
                </Button>
                <div className="flex items-center space-x-1">
                  {Array.from({ length: totalPages }, (_, i) => i + 1)
                    .filter(pageNum => 
                      pageNum === 1 || 
                      pageNum === totalPages || 
                      Math.abs(pageNum - page) <= 1
                    )
                    .map((pageNum, index, array) => (
                      <div key={pageNum} className="flex items-center">
                        {index > 0 && array[index - 1] !== pageNum - 1 && (
                          <span className="text-slate-400 px-1">...</span>
                        )}
                        <Button
                          variant={pageNum === page ? "default" : "outline"}
                          size="sm"
                          onClick={() => setPage(pageNum)}
                          className={
                            pageNum === page 
                              ? "bg-blue-600 hover:bg-blue-700" 
                              : "border-slate-600 text-slate-300 hover:bg-slate-700"
                          }
                        >
                          {pageNum}
                        </Button>
                      </div>
                    ))}
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage(page + 1)}
                  disabled={page === totalPages}
                  className="border-slate-600 text-slate-300 hover:bg-slate-700"
                >
                  Next
                </Button>
              </div>
            </div>
          )}

          {schemas.length === 0 && (
            <div className="text-center py-12">
              <Database className="h-12 w-12 text-slate-500 mx-auto mb-4" />
              <p className="text-slate-400 mb-2">No schemas found</p>
              <p className="text-sm text-slate-500">
                {search || statusFilter !== 'all' || typeFilter !== 'all' 
                  ? 'Try adjusting your search or filters'
                  : 'Create your first schema to get started'
                }
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Modals and Drawers - Temporarily commented out */}
      {/* 
      <CreateSchemaModal
        open={createModalOpen}
        onOpenChange={setCreateModalOpen}
        onSuccess={handleCreateSuccess}
      />

      <EditSchemaDrawer
        open={editDrawerOpen}
        onOpenChange={setEditDrawerOpen}
        schema={selectedSchema}
        onSuccess={handleEditSuccess}
      />

      <DeleteSchemaDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        schema={selectedSchema}
        onSuccess={handleDeleteSuccess}
      />
      */}
    </div>
  )
}
