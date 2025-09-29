"use client"

import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useToast } from '@/components/ui/use-toast'
import { apiClient } from '@/lib/api-client'
import { Building2, Loader2 } from 'lucide-react'

interface CreateTenantFormData {
  name: string
  slug: string
  domain: string
  plan: string
  ownerEmail: string
  ownerName: string
}

interface CreateTenantModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

export function CreateTenantModal({ open, onOpenChange, onSuccess }: CreateTenantModalProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [formData, setFormData] = useState<CreateTenantFormData>({
    name: '',
    slug: '',
    domain: '',
    plan: 'starter',
    ownerEmail: '',
    ownerName: '',
  })
  const { toast } = useToast()

  // Auto-generate slug from name
  const handleNameChange = (name: string) => {
    const slug = name
      .toLowerCase()
      .replace(/[^a-z0-9\s\-]/g, '') // Remove invalid characters
      .replace(/\s+/g, '-') // Replace spaces with hyphens
      .replace(/-+/g, '-') // Replace multiple hyphens with single
      .replace(/^-|-$/g, '') // Remove leading/trailing hyphens
    
    setFormData(prev => ({
      ...prev,
      name,
      slug,
      domain: slug ? `${slug}.yourdomain.com` : ''
    }))
  }

  const handleInputChange = (field: keyof CreateTenantFormData, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }))
  }

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    
    try {
      // Basic validation
      if (!formData.name || !formData.slug || !formData.domain || !formData.ownerEmail || !formData.ownerName) {
        throw new Error('Please fill in all required fields')
      }

      // Transform form data to API format
      const tenantData = {
        name: formData.name,
        slug: formData.slug,
        domain: formData.domain,
        plan: formData.plan,
        owner_email: formData.ownerEmail,
        owner_name: formData.ownerName,
        status: 'active',
      }

      await apiClient.createTenant(tenantData)
      
      toast({
        title: "Tenant Created Successfully",
        description: `${formData.name} has been created and is ready to use.`,
      })
      
      // Reset form and close modal
      setFormData({
        name: '',
        slug: '',
        domain: '',
        plan: 'starter',
        ownerEmail: '',
        ownerName: '',
      })
      onOpenChange(false)
      onSuccess?.()
      
    } catch (error: any) {
      console.error('Failed to create tenant:', error)
      
      toast({
        title: "Failed to Create Tenant",
        description: error.message || "An unexpected error occurred. Please try again.",
        variant: "destructive",
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleCancel = () => {
    setFormData({
      name: '',
      slug: '',
      domain: '',
      plan: 'starter',
      ownerEmail: '',
      ownerName: '',
    })
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5 text-blue-500" />
            Create New Tenant
          </DialogTitle>
          <DialogDescription>
            Add a new tenant organization to your platform. All fields are required.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={onSubmit} className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Organization Name */}
            <div className="space-y-2">
              <Label htmlFor="name">Organization Name</Label>
              <Input
                id="name"
                placeholder="Acme Corporation"
                value={formData.name}
                onChange={(e) => handleNameChange(e.target.value)}
                className="bg-slate-700 border-slate-600 text-white"
                required
              />
              <p className="text-sm text-slate-400">The display name for this organization</p>
            </div>

            {/* Slug */}
            <div className="space-y-2">
              <Label htmlFor="slug">Slug</Label>
              <Input
                id="slug"
                placeholder="acme-corporation"
                value={formData.slug}
                onChange={(e) => handleInputChange('slug', e.target.value)}
                className="bg-slate-700 border-slate-600 text-white"
                required
              />
              <p className="text-sm text-slate-400">URL-friendly identifier (auto-generated)</p>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Domain */}
            <div className="space-y-2">
              <Label htmlFor="domain">Domain</Label>
              <Input
                id="domain"
                placeholder="acme.yourdomain.com"
                value={formData.domain}
                onChange={(e) => handleInputChange('domain', e.target.value)}
                className="bg-slate-700 border-slate-600 text-white"
                required
              />
              <p className="text-sm text-slate-400">Custom domain for this tenant</p>
            </div>

            {/* Plan */}
            <div className="space-y-2">
              <Label htmlFor="plan">Plan</Label>
              <Select value={formData.plan} onValueChange={(value: string) => handleInputChange('plan', value)}>
                <SelectTrigger className="bg-slate-700 border-slate-600 text-white">
                  <SelectValue placeholder="Select a plan" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="starter">Starter - $29/month</SelectItem>
                  <SelectItem value="pro">Pro - $99/month</SelectItem>
                  <SelectItem value="enterprise">Enterprise - $299/month</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-sm text-slate-400">Choose the subscription plan</p>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Owner Name */}
            <div className="space-y-2">
              <Label htmlFor="ownerName">Owner Name</Label>
              <Input
                id="ownerName"
                placeholder="John Smith"
                value={formData.ownerName}
                onChange={(e) => handleInputChange('ownerName', e.target.value)}
                className="bg-slate-700 border-slate-600 text-white"
                required
              />
              <p className="text-sm text-slate-400">Full name of the organization owner</p>
            </div>

            {/* Owner Email */}
            <div className="space-y-2">
              <Label htmlFor="ownerEmail">Owner Email</Label>
              <Input
                id="ownerEmail"
                type="email"
                placeholder="john@acme.com"
                value={formData.ownerEmail}
                onChange={(e) => handleInputChange('ownerEmail', e.target.value)}
                className="bg-slate-700 border-slate-600 text-white"
                required
              />
              <p className="text-sm text-slate-400">Primary contact email for this tenant</p>
            </div>
          </div>

          <DialogFooter>
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
                  Creating...
                </>
              ) : (
                'Create Tenant'
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
