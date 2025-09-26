"use client"

import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { apiClient } from '@/lib/api-client'
import { 
  Users, 
  Database, 
  Activity, 
  DollarSign, 
  TrendingUp, 
  AlertTriangle,
  CheckCircle,
  XCircle,
  Zap
} from 'lucide-react'
import { formatNumber, formatCurrency, getStatusColor } from '@/lib/utils'

export default function DashboardPage() {
  const { data: systemHealth } = useQuery({
    queryKey: ['system-health'],
    queryFn: () => apiClient.getSystemHealth(),
    refetchInterval: 30000, // Refresh every 30 seconds
  })

  const { data: tenantsData } = useQuery({
    queryKey: ['tenants-summary'],
    queryFn: () => apiClient.getTenants({ limit: 5 }),
  })

  const { data: analyticsData } = useQuery({
    queryKey: ['analytics-summary'],
    queryFn: () => apiClient.getAnalytics(),
  })

  const { data: gatewayMetrics } = useQuery({
    queryKey: ['gateway-metrics-summary'],
    queryFn: () => apiClient.getGatewayMetrics(),
    refetchInterval: 30000, // Refresh every 30 seconds
  })

  const stats = [
    {
      title: "Total Requests",
      value: formatNumber(gatewayMetrics?.requests?.total || 0),
      change: "Live",
      icon: Activity,
      color: "text-blue-600",
    },
    {
      title: "Success Rate",
      value: gatewayMetrics?.requests ? 
        `${((gatewayMetrics.requests.by_status['200'] || 0) / gatewayMetrics.requests.total * 100).toFixed(1)}%` : 
        "0%",
      change: "Current",
      icon: CheckCircle,
      color: "text-green-600",
    },
    {
      title: "Avg Response",
      value: `${gatewayMetrics?.requests?.average_response_time_ms || 0}ms`,
      change: "Real-time",
      icon: Zap,
      color: "text-purple-600",
    },
    {
      title: "Total Errors",
      value: formatNumber(gatewayMetrics?.errors?.total || 0),
      change: gatewayMetrics?.errors?.total > 0 ? "Monitor" : "Good",
      icon: gatewayMetrics?.errors?.total > 0 ? AlertTriangle : CheckCircle,
      color: gatewayMetrics?.errors?.total > 0 ? "text-yellow-600" : "text-emerald-600",
    },
  ]

  const healthStatus = systemHealth?.data?.status || 'unknown'
  const services = systemHealth?.data?.services || []

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Dashboard</h1>
          <p className="text-slate-400 mt-1">
            Platform overview and system status
          </p>
        </div>
        <Button className="bg-blue-600 hover:bg-blue-700">
          View Reports
        </Button>
      </div>

      {/* System Health Alert */}
      {healthStatus !== 'healthy' && (
        <Card className="border-orange-500 bg-orange-500/10">
          <CardContent className="flex items-center p-4">
            <AlertTriangle className="h-5 w-5 text-orange-500 mr-3" />
            <div>
              <p className="text-sm font-medium text-orange-200">
                System Status: {healthStatus.toUpperCase()}
              </p>
              <p className="text-xs text-orange-300">
                Some services may be experiencing issues. Check system health for details.
              </p>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {stats.map((stat) => (
          <Card key={stat.title} className="bg-slate-800 border-slate-700">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-slate-400">
                    {stat.title}
                  </p>
                  <p className="text-2xl font-bold text-white mt-2">
                    {stat.value}
                  </p>
                  <p className="text-xs text-green-400 mt-1 flex items-center">
                    <TrendingUp className="h-3 w-3 mr-1" />
                    {stat.change} from last month
                  </p>
                </div>
                <div className={`p-3 rounded-full bg-slate-700 ${stat.color}`}>
                  <stat.icon className="h-6 w-6" />
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* System Health */}
        <Card className="bg-slate-800 border-slate-700">
          <CardHeader>
            <CardTitle className="text-white flex items-center justify-between">
              <div className="flex items-center">
                <Activity className="h-5 w-5 mr-2" />
                System Health
              </div>
              <Button 
                variant="ghost" 
                size="sm" 
                className="text-blue-400 hover:text-blue-300"
                onClick={() => window.location.href = '/admin/dashboard/health'}
              >
                View Details →
              </Button>
            </CardTitle>
            <CardDescription>
              Real-time status of platform services
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex items-center justify-between p-3 bg-slate-700 rounded-lg">
                <span className="text-sm font-medium text-slate-200">
                  Overall Status
                </span>
                <div className="flex items-center">
                  {healthStatus === 'healthy' ? (
                    <CheckCircle className="h-4 w-4 text-green-500 mr-2" />
                  ) : (
                    <XCircle className="h-4 w-4 text-red-500 mr-2" />
                  )}
                  <span className={`text-sm font-medium ${
                    healthStatus === 'healthy' ? 'text-green-400' : 'text-red-400'
                  }`}>
                    {healthStatus.toUpperCase()}
                  </span>
                </div>
              </div>
              
              {services.slice(0, 4).map((service) => (
                <div key={service.name} className="flex items-center justify-between">
                  <span className="text-sm text-slate-300">{service.name}</span>
                  <div className="flex items-center">
                    {service.response_time && (
                      <span className="text-xs text-slate-500 mr-2">
                        {service.response_time}ms
                      </span>
                    )}
                    <span className={`text-xs px-2 py-1 rounded-full ${
                      service.status === 'up' 
                        ? 'bg-green-900 text-green-300' 
                        : 'bg-red-900 text-red-300'
                    }`}>
                      {service.status.toUpperCase()}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Recent Tenants */}
        <Card className="bg-slate-800 border-slate-700">
          <CardHeader>
            <CardTitle className="text-white flex items-center">
              <Users className="h-5 w-5 mr-2" />
              Recent Tenants
            </CardTitle>
            <CardDescription>
              Latest tenant registrations
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {tenantsData?.data?.tenants?.slice(0, 5).map((tenant) => (
                <div key={tenant.id} className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-slate-200">
                      {tenant.name}
                    </p>
                    <p className="text-xs text-slate-500">
                      {tenant.slug}
                    </p>
                  </div>
                  <span className={`text-xs px-2 py-1 rounded-full ${getStatusColor(tenant.status)}`}>
                    {tenant.status.toUpperCase()}
                  </span>
                </div>
              )) || (
                <p className="text-sm text-slate-400 text-center py-4">
                  No tenants found
                </p>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions */}
      <Card className="bg-slate-800 border-slate-700">
        <CardHeader>
          <CardTitle className="text-white">Quick Actions</CardTitle>
          <CardDescription>
            Common administrative tasks
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Button variant="outline" className="h-20 flex-col border-slate-600 hover:bg-slate-700">
              <Users className="h-6 w-6 mb-2" />
              <span>Create Tenant</span>
            </Button>
            <Button variant="outline" className="h-20 flex-col border-slate-600 hover:bg-slate-700">
              <Database className="h-6 w-6 mb-2" />
              <span>New Schema</span>
            </Button>
            <Button 
              variant="outline" 
              className="h-20 flex-col border-slate-600 hover:bg-slate-700"
              onClick={() => window.location.href = '/admin/dashboard/analytics'}
            >
              <Activity className="h-6 w-6 mb-2" />
              <span>Analytics</span>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
