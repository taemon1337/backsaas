"use client"

import { useState } from 'react'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { useToast } from '@/components/ui/use-toast'
import { apiClient } from '@/lib/api-client'
import { AlertTriangle, Loader2, Users, HardDrive, Database } from 'lucide-react'

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
}

interface DeleteTenantDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  tenant: Tenant | null
  onSuccess?: () => void
}

export function DeleteTenantDialog({ open, onOpenChange, tenant, onSuccess }: DeleteTenantDialogProps) {
  const [isDeleting, setIsDeleting] = useState(false)
  const [confirmationText, setConfirmationText] = useState('')
  const { toast } = useToast()

  const handleDelete = async () => {
    if (!tenant) return
    
    // Require exact name confirmation for safety
    if (confirmationText !== tenant.name) {
      toast({
        title: "Confirmation Required",
        description: `Please type "${tenant.name}" exactly to confirm deletion.`,
        variant: "destructive",
      })
      return
    }

    setIsDeleting(true)
    
    try {
      await apiClient.deleteTenant(tenant.id)
      
      toast({
        title: "Tenant Deleted",
        description: `${tenant.name} has been permanently deleted.`,
      })
      
      setConfirmationText('')
      onOpenChange(false)
      onSuccess?.()
      
    } catch (error: any) {
      console.error('Failed to delete tenant:', error)
      
      toast({
        title: "Failed to Delete Tenant",
        description: error.message || "An unexpected error occurred. Please try again.",
        variant: "destructive",
      })
    } finally {
      setIsDeleting(false)
    }
  }

  const handleCancel = () => {
    setConfirmationText('')
    onOpenChange(false)
  }

  const getPlanColor = (plan: string) => {
    switch (plan) {
      case 'enterprise': return 'bg-purple-500'
      case 'pro': return 'bg-blue-500'
      case 'starter': return 'bg-green-500'
      default: return 'bg-gray-500'
    }
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

  if (!tenant) return null

  const isConfirmationValid = confirmationText === tenant.name

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent className="sm:max-w-[500px] bg-slate-800 border-slate-700">
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2 text-red-400">
            <AlertTriangle className="h-5 w-5" />
            Delete Tenant
          </AlertDialogTitle>
          <AlertDialogDescription className="text-slate-300">
            This action cannot be undone. This will permanently delete the tenant and all associated data.
          </AlertDialogDescription>
        </AlertDialogHeader>

        {/* Tenant Information */}
        <div className="bg-slate-700 rounded-lg p-4 space-y-3">
          <div className="flex items-center justify-between">
            <h3 className="font-semibold text-white">{tenant.name}</h3>
            <div className="flex gap-2">
              <Badge className={`${getStatusColor(tenant.status)} text-white text-xs`}>
                {tenant.status}
              </Badge>
              <Badge className={`${getPlanColor(tenant.plan)} text-white text-xs`}>
                {tenant.plan}
              </Badge>
            </div>
          </div>
          
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-slate-400">Domain:</span>
              <p className="text-white">{tenant.domain}</p>
            </div>
            <div>
              <span className="text-slate-400">Owner:</span>
              <p className="text-white">{tenant.owner_name}</p>
            </div>
          </div>

          {/* Data Impact Warning */}
          <div className="border-t border-slate-600 pt-3">
            <p className="text-sm font-medium text-red-400 mb-2">⚠️ Data that will be permanently deleted:</p>
            <div className="grid grid-cols-1 gap-2 text-sm">
              {tenant.user_count && (
                <div className="flex items-center gap-2 text-slate-300">
                  <Users className="h-4 w-4" />
                  <span>{tenant.user_count} user accounts and profiles</span>
                </div>
              )}
              {tenant.storage_used && (
                <div className="flex items-center gap-2 text-slate-300">
                  <HardDrive className="h-4 w-4" />
                  <span>{tenant.storage_used} of stored data and files</span>
                </div>
              )}
              <div className="flex items-center gap-2 text-slate-300">
                <Database className="h-4 w-4" />
                <span>All tenant configurations and settings</span>
              </div>
            </div>
          </div>
        </div>

        {/* Confirmation Input */}
        <div className="space-y-2">
          <Label htmlFor="confirmation" className="text-slate-300">
            Type <span className="font-mono font-bold text-white">{tenant.name}</span> to confirm:
          </Label>
          <Input
            id="confirmation"
            value={confirmationText}
            onChange={(e) => setConfirmationText(e.target.value)}
            placeholder={tenant.name}
            className="bg-slate-700 border-slate-600 text-white"
            autoComplete="off"
          />
        </div>

        <AlertDialogFooter>
          <AlertDialogCancel 
            onClick={handleCancel}
            disabled={isDeleting}
            className="bg-slate-700 border-slate-600 text-white hover:bg-slate-600"
          >
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            disabled={!isConfirmationValid || isDeleting}
            className="bg-red-600 hover:bg-red-700 text-white"
          >
            {isDeleting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Deleting...
              </>
            ) : (
              'Delete Tenant'
            )}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
