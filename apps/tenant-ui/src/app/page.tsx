"use client"

import { useTenant } from '@/lib/tenant-context'
import { DashboardLayout } from '@/components/layout/dashboard-layout'
import { DashboardOverview } from '@/components/dashboard/dashboard-overview'
import { ErrorMessage } from '@/components/ui/error-message'
import { LoadingPage } from '@/components/ui/loading'
import { DefaultErrorFallback } from '@/components/ui/error-boundary'
import { Button } from '@/components/ui/button'

export default function TenantDashboard() {
  const { tenant, user, isLoading, error, retry } = useTenant()

  if (isLoading) {
    return <LoadingPage message="Loading your dashboard..." />
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8 text-center">
          <h1 className="text-xl font-semibold text-gray-900 mb-2">
            Unable to load dashboard
          </h1>
          <p className="text-gray-600 mb-6">{error}</p>
          <div className="space-y-3">
            <Button onClick={retry} className="w-full">
              Try Again
            </Button>
            <Button 
              variant="outline" 
              onClick={() => window.location.href = '/'}
              className="w-full"
            >
              Go to Home
            </Button>
          </div>
        </div>
      </div>
    )
  }

  if (!tenant) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
        <ErrorMessage 
          title="Tenant not found"
          message="Unable to identify your organization. Please check the URL and try again."
        />
      </div>
    )
  }

  // If user is not authenticated, redirect to login
  if (!user) {
    window.location.href = '/login'
    return null
  }

  return (
    <DashboardLayout>
      <DashboardOverview />
    </DashboardLayout>
  )
}
