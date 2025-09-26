"use client"

import { useTenant } from '@/lib/tenant-context'
import { DashboardLayout } from '@/components/layout/dashboard-layout'
import { DashboardOverview } from '@/components/dashboard/dashboard-overview'
import { LoadingSpinner } from '@/components/ui/loading-spinner'
import { ErrorMessage } from '@/components/ui/error-message'

export default function DashboardPage() {
  const { tenant, user, isLoading, error } = useTenant()

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <ErrorMessage 
          title="Failed to load tenant"
          message={error}
          action={{
            label: "Retry",
            onClick: () => window.location.reload()
          }}
        />
      </div>
    )
  }

  if (!tenant) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <ErrorMessage 
          title="Tenant not found"
          message="Unable to identify your organization. Please check the URL and try again."
        />
      </div>
    )
  }

  // If user is not authenticated, redirect to login
  if (!user) {
    window.location.href = '/auth/login'
    return null
  }

  return (
    <DashboardLayout>
      <DashboardOverview />
    </DashboardLayout>
  )
}
