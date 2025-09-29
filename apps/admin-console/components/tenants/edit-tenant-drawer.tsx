"use client"

import { useState, useEffect } from 'react'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { useToast } from '@/components/ui/use-toast'
import { apiClient } from '@/lib/api-client'
import { Building2, Loader2, Calendar, Users, HardDrive } from 'lucide-react'

interface Tenant {
  id: string
  name: string
  slug: string
  domain: string
  plan: string
  status: string
  owner_email: string
  owner_name: string
  user_count?: number
  storage_used?: string
  created_at?: string
  last_activity?: string
}

interface EditTenantFormData {
  name: string
  slug: string
  domain: string
  plan: string
  status: string
  owner_email: string
  owner_name: string
}

interface EditTenantDrawerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  tenant: Tenant | null
  onSuccess?: () => void
}

export function EditTenantDrawer({ open, onOpenChange, tenant, onSuccess }: EditTenantDrawerProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [formData, setFormData] = useState<EditTenantFormData>({
    name: '',
    slug: '',
    domain: '',
    plan: 'starter',
    status: 'active',
    owner_email: '',
    owner_name: '',
  })
  const { toast } = useToast()

  // Update form data when tenant changes
  useEffect(() => {
    if (tenant) {
      setFormData({
        name: tenant.name,
        slug: tenant.slug,
        domain: tenant.domain,
        plan: tenant.plan,
        status: tenant.status,
        owner_email: tenant.owner_email,
        owner_name: tenant.owner_name,
      })
    }
  }, [tenant])

  const handleInputChange = (field: keyof EditTenantFormData, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }))
  }

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!tenant) return
    
    setIsSubmitting(true)
    
    try {
      // Basic validation
      if (!formData.name || !formData.slug || !formData.domain || !formData.owner_email || !formData.owner_name) {
        throw new Error('Please fill in all required fields')
      }

      // Transform form data to API format
      const tenantData = {
        name: formData.name,
        slug: formData.slug,
        domain: formData.domain,
        plan: formData.plan,
        status: formData.status,
        owner_email: formData.owner_email,
        owner_name: formData.owner_name,
      }

      await apiClient.updateTenant(tenant.id, tenantData)
      
      toast({
        title: "Tenant Updated Successfully",
        description: `${formData.name} has been updated.`,
      })
      
      onOpenChange(false)
      onSuccess?.()
      
    } catch (error: any) {
      console.error('Failed to update tenant:', error)
      
      toast({
        title: "Failed to Update Tenant",
        description: error.message || "An unexpected error occurred. Please try again.",
        variant: "destructive",
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleCancel = () => {
    if (tenant) {
      setFormData({
        name: tenant.name,
        slug: tenant.slug,
        domain: tenant.domain,
        plan: tenant.plan,
        status: tenant.status,
        owner_email: tenant.owner_email,
        owner_name: tenant.owner_name,
      })
    }
    onOpenChange(false)
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-green-500'
      case 'trial': return 'bg-blue-500'
      case 'suspended': return 'bg-red-500'
      case 'inactive': return 'bg-gray-500'
      default: return 'bg-gray-500'
    }
  }

  const getPlanColor = (plan: string) => {
    switch (plan) {
      case 'enterprise': return 'bg-purple-500'
      case 'pro': return 'bg-blue-500'
      case 'starter': return 'bg-green-500'
      default: return 'bg-gray-500'
    }
  }

  if (!tenant) return null

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="sm:max-w-[600px] w-full">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5 text-blue-500" />
            Edit Tenant: {tenant.name}
          </SheetTitle>
          <SheetDescription>
            Update tenant information and settings. Changes will be applied immediately.
          </SheetDescription>
        </SheetHeader>

        {/* Tenant Overview */}
        <div className="py-4 border-b border-slate-700">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            <div className="flex items-center gap-2">
              <Badge className={`${getStatusColor(tenant.status)} text-white`}>
                {tenant.status}
              </Badge>
              <span className="text-slate-400">Status</span>
            </div>
            <div className="flex items-center gap-2">
              <Badge className={`${getPlanColor(tenant.plan)} text-white`}>
                {tenant.plan}
              </Badge>
              <span className="text-slate-400">Plan</span>
            </div>
            {tenant.user_count && (
              <div className="flex items-center gap-2">
                <Users className="h-4 w-4 text-slate-400" />
                <span className="text-white">{tenant.user_count}</span>
                <span className="text-slate-400">Users</span>
              </div>
            )}
            {tenant.storage_used && (
              <div className="flex items-center gap-2">
                <HardDrive className="h-4 w-4 text-slate-400" />
                <span className="text-white">{tenant.storage_used}</span>
                <span className="text-slate-400">Storage</span>
              </div>
            )}
          </div>
          {tenant.created_at && (
            <div className="mt-2 flex items-center gap-2 text-sm text-slate-400">
              <Calendar className="h-4 w-4" />
              <span>Created: {new Date(tenant.created_at).toLocaleDateString()}</span>
              {tenant.last_activity && (
                <span className="ml-4">Last Activity: {new Date(tenant.last_activity).toLocaleDateString()}</span>
              )}
            </div>
          )}
        </div>

        <form onSubmit={onSubmit} className="space-y-6 py-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Organization Name */}
            <div className="space-y-2">
              <Label htmlFor="edit-name">Organization Name</Label>
              <Input
                id="edit-name"
                value={formData.name}
                onChange={(e) => handleInputChange('name', e.target.value)}
                className="bg-slate-700 border-slate-600 text-white"
                required
              />
            </div>

            {/* Slug */}
            <div className="space-y-2">
              <Label htmlFor="edit-slug">Slug</Label>
              <Input
                id="edit-slug"
                value={formData.slug}
                onChange={(e) => handleInputChange('slug', e.target.value)}
                className="bg-slate-700 border-slate-600 text-white"
                required
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Domain */}
            <div className="space-y-2">
              <Label htmlFor="edit-domain">Domain</Label>
              <Input
                id="edit-domain"
                value={formData.domain}
                onChange={(e) => handleInputChange('domain', e.target.value)}
                className="bg-slate-700 border-slate-600 text-white"
                required
              />
            </div>

            {/* Plan */}
            <div className="space-y-2">
              <Label htmlFor="edit-plan">Plan</Label>
              <Select value={formData.plan} onValueChange={(value: string) => handleInputChange('plan', value)}>
                <SelectTrigger className="bg-slate-700 border-slate-600 text-white">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="starter">Starter - $29/month</SelectItem>
                  <SelectItem value="pro">Pro - $99/month</SelectItem>
                  <SelectItem value="enterprise">Enterprise - $299/month</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Status */}
            <div className="space-y-2">
              <Label htmlFor="edit-status">Status</Label>
              <Select value={formData.status} onValueChange={(value: string) => handleInputChange('status', value)}>
                <SelectTrigger className="bg-slate-700 border-slate-600 text-white">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="active">Active</SelectItem>
                  <SelectItem value="trial">Trial</SelectItem>
                  <SelectItem value="suspended">Suspended</SelectItem>
                  <SelectItem value="inactive">Inactive</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Owner Name */}
            <div className="space-y-2">
              <Label htmlFor="edit-owner-name">Owner Name</Label>
              <Input
                id="edit-owner-name"
                value={formData.owner_name}
                onChange={(e) => handleInputChange('owner_name', e.target.value)}
                className="bg-slate-700 border-slate-600 text-white"
                required
              />
            </div>

            {/* Owner Email */}
            <div className="space-y-2">
              <Label htmlFor="edit-owner-email">Owner Email</Label>
              <Input
                id="edit-owner-email"
                type="email"
                value={formData.owner_email}
                onChange={(e) => handleInputChange('owner_email', e.target.value)}
                className="bg-slate-700 border-slate-600 text-white"
                required
              />
            </div>
          </div>

          <SheetFooter>
            <Button
              type="button"
              variant="outline"
              onClick={handleCancel}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isSubmitting}
              className="bg-blue-600 hover:bg-blue-700"
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Updating...
                </>
              ) : (
                'Update Tenant'
              )}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
