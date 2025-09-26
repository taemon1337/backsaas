"use client"

import React, { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useTenant } from '@/lib/tenant-context'
import { tenantApi } from '@/lib/tenant-api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { LoadingSpinner } from '@/components/ui/loading-spinner'
import { AlertTriangle } from 'lucide-react'

export default function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState('')
  
  const { tenant, setSession, setUser } = useTenant()
  const router = useRouter()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setError('')

    try {
      if (!tenant) {
        throw new Error('Tenant not found')
      }

      // Set tenant context for API calls
      tenantApi.setTenantContext(tenant.id)

      // Attempt login
      const response = await tenantApi.login(email, password, tenant.slug)
      
      // Set session and user in context
      const session = {
        user: response.user,
        tenant: response.tenant,
        token: response.token,
        expiresAt: response.expiresAt,
      }
      
      setSession(session)
      setUser(response.user)
      
      // Redirect to dashboard
      router.push('/')
    } catch (error: any) {
      console.error('Login failed:', error)
      setError(error.message || 'Login failed. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  if (!tenant) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardContent className="p-6 text-center">
            <AlertTriangle className="h-12 w-12 text-red-500 mx-auto mb-4" />
            <h2 className="text-lg font-semibold text-gray-900 mb-2">
              Organization Not Found
            </h2>
            <p className="text-gray-600">
              Unable to identify your organization from the URL. Please check the address and try again.
            </p>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        {/* Header */}
        <div className="text-center">
          {tenant.branding?.logo ? (
            <img
              className="mx-auto h-12 w-auto"
              src={tenant.branding.logo}
              alt={`${tenant.name} logo`}
            />
          ) : (
            <div className="mx-auto h-12 w-12 rounded-lg bg-tenant-primary flex items-center justify-center">
              <span className="text-white font-bold text-xl">
                {tenant.name.charAt(0).toUpperCase()}
              </span>
            </div>
          )}
          <h2 className="mt-6 text-3xl font-bold text-gray-900">
            Sign in to {tenant.name}
          </h2>
          <p className="mt-2 text-sm text-gray-600">
            Access your business dashboard
          </p>
        </div>

        {/* Login form */}
        <Card>
          <CardHeader>
            <CardTitle>Welcome back</CardTitle>
            <CardDescription>
              Enter your credentials to access your account
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              {error && (
                <div className="bg-red-50 border border-red-200 rounded-md p-3">
                  <div className="flex">
                    <AlertTriangle className="h-4 w-4 text-red-400 mt-0.5" />
                    <div className="ml-2">
                      <p className="text-sm text-red-800">{error}</p>
                    </div>
                  </div>
                </div>
              )}

              <div>
                <Label htmlFor="email">Email address</Label>
                <Input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="Enter your email"
                  required
                  disabled={isLoading}
                />
              </div>

              <div>
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Enter your password"
                  required
                  disabled={isLoading}
                />
              </div>

              <Button
                type="submit"
                className="w-full"
                disabled={isLoading}
              >
                {isLoading ? (
                  <>
                    <LoadingSpinner size="sm" className="mr-2" />
                    Signing in...
                  </>
                ) : (
                  'Sign in'
                )}
              </Button>
            </form>

            {/* Demo credentials */}
            <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-md">
              <h4 className="text-sm font-medium text-blue-900 mb-2">
                Demo Credentials
              </h4>
              <div className="text-sm text-blue-800 space-y-1">
                <p><strong>Email:</strong> demo@{tenant.slug}.com</p>
                <p><strong>Password:</strong> demo123</p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Footer */}
        <div className="text-center text-sm text-gray-500">
          <p>
            Need help? Contact your administrator or{' '}
            <a href="#" className="text-tenant-primary hover:underline">
              visit our support center
            </a>
          </p>
        </div>
      </div>
    </div>
  )
}
