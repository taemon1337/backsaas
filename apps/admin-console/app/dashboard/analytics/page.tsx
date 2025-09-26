"use client"

import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api-client'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { 
  Activity, 
  BarChart3, 
  TrendingUp, 
  TrendingDown,
  RefreshCw, 
  Clock,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Zap,
  Users,
  Globe,
  Server
} from 'lucide-react'

interface GatewayMetrics {
  gateway: {
    uptime_seconds: number
    start_time: string
    last_request_time: string
  }
  requests: {
    total: number
    by_status: Record<string, number>
    by_route: Record<string, number>
    by_tenant: Record<string, number>
    average_response_time_ms: number
  }
  errors: {
    total: number
    by_type: Record<string, number>
  }
  rate_limiting: {
    total_hits: number
    by_tenant: Record<string, number>
  }
  backends: {
    requests: Record<string, number>
    errors: Record<string, number>
    response_times: Record<string, number>
  }
}

export default function AnalyticsPage() {
  const [refreshing, setRefreshing] = useState(false)
  
  // Fetch gateway metrics
  const { 
    data: metrics, 
    isLoading, 
    error, 
    refetch 
  } = useQuery({
    queryKey: ['gateway-metrics'],
    queryFn: () => apiClient.getGatewayMetrics(),
    refetchInterval: 30000, // Refresh every 30 seconds
  })

  const handleRefresh = async () => {
    setRefreshing(true)
    await refetch()
    setRefreshing(false)
  }

  // Calculate derived metrics
  const successRate = metrics?.requests ? 
    ((metrics.requests.by_status['200'] || 0) / metrics.requests.total * 100) : 0
  
  const errorRate = metrics?.errors ? 
    (metrics.errors.total / (metrics?.requests?.total || 1) * 100) : 0

  const formatUptime = (seconds: number) => {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    return `${hours}h ${minutes}m`
  }

  const formatNumber = (num: number) => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`
    return num.toString()
  }

  const getStatusColor = (status: string) => {
    if (status.startsWith('2')) return 'text-green-400'
    if (status.startsWith('3')) return 'text-blue-400'
    if (status.startsWith('4')) return 'text-yellow-400'
    if (status.startsWith('5')) return 'text-red-400'
    return 'text-slate-400'
  }

  if (isLoading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <RefreshCw className="h-8 w-8 animate-spin text-blue-500" />
          <span className="ml-2 text-slate-300">Loading analytics data...</span>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6">
        <Card className="bg-red-900/20 border-red-500">
          <CardContent className="flex items-center p-6">
            <XCircle className="h-5 w-5 text-red-500 mr-3" />
            <div>
              <p className="text-sm font-medium text-red-200">Failed to load analytics data</p>
              <p className="text-xs text-red-300">Please try refreshing the page</p>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center space-x-2 mb-2">
            <Button 
              variant="ghost" 
              size="sm" 
              className="text-slate-400 hover:text-white p-0"
              onClick={() => window.location.href = '/admin/dashboard'}
            >
              ← Back to Dashboard
            </Button>
          </div>
          <h1 className="text-2xl font-bold text-white">Analytics Dashboard</h1>
          <p className="text-slate-400">Gateway traffic, performance, and usage metrics</p>
        </div>
        <Button 
          onClick={handleRefresh}
          disabled={refreshing}
          variant="outline"
          size="sm"
        >
          <RefreshCw className={`h-4 w-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-200">Total Requests</CardTitle>
            <Globe className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">
              {formatNumber(metrics?.requests?.total || 0)}
            </div>
            <p className="text-xs text-slate-400">
              Since {metrics?.gateway ? new Date(metrics.gateway.start_time).toLocaleDateString() : 'startup'}
            </p>
          </CardContent>
        </Card>

        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-200">Success Rate</CardTitle>
            <CheckCircle className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-400">
              {successRate.toFixed(1)}%
            </div>
            <p className="text-xs text-slate-400">
              {metrics?.requests?.by_status['200'] || 0} successful requests
            </p>
          </CardContent>
        </Card>

        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-200">Avg Response Time</CardTitle>
            <Zap className="h-4 w-4 text-yellow-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">
              {metrics?.requests?.average_response_time_ms || 0}ms
            </div>
            <p className="text-xs text-slate-400">Average across all requests</p>
          </CardContent>
        </Card>

        <Card className="bg-slate-800 border-slate-700">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-slate-200">Uptime</CardTitle>
            <Clock className="h-4 w-4 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">
              {formatUptime(metrics?.gateway?.uptime_seconds || 0)}
            </div>
            <p className="text-xs text-slate-400">
              Since {metrics?.gateway ? new Date(metrics.gateway.start_time).toLocaleTimeString() : 'startup'}
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Request Status Breakdown */}
        <Card className="bg-slate-800 border-slate-700">
          <CardHeader>
            <CardTitle className="text-white flex items-center">
              <BarChart3 className="h-5 w-5 mr-2" />
              HTTP Status Codes
            </CardTitle>
            <CardDescription>Request success and error breakdown</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {Object.entries(metrics?.requests?.by_status || {})
                .sort(([a], [b]) => a.localeCompare(b))
                .map(([status, count]) => {
                  const percentage = ((count as number) / (metrics?.requests?.total || 1)) * 100
                  return (
                    <div key={status} className="flex items-center justify-between">
                      <div className="flex items-center space-x-3">
                        <Badge 
                          variant="outline" 
                          className={`min-w-[60px] justify-center ${getStatusColor(status)}`}
                        >
                          {status}
                        </Badge>
                        <span className="text-sm text-slate-300">
                          {count as number} requests
                        </span>
                      </div>
                      <div className="flex items-center space-x-2">
                        <div className="w-20">
                          <Progress value={percentage} className="h-2" />
                        </div>
                        <span className="text-xs text-slate-400 min-w-[40px]">
                          {percentage.toFixed(1)}%
                        </span>
                      </div>
                    </div>
                  )
                })}
            </div>
          </CardContent>
        </Card>

        {/* Route Usage */}
        <Card className="bg-slate-800 border-slate-700">
          <CardHeader>
            <CardTitle className="text-white flex items-center">
              <Activity className="h-5 w-5 mr-2" />
              Route Usage
            </CardTitle>
            <CardDescription>Most active API routes</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {Object.entries(metrics?.requests?.by_route || {})
                .sort(([,a], [,b]) => (b as number) - (a as number))
                .slice(0, 5)
                .map(([route, count]) => {
                  const percentage = ((count as number) / (metrics?.requests?.total || 1)) * 100
                  return (
                    <div key={route} className="space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium text-slate-200 truncate">
                          {route}
                        </span>
                        <span className="text-sm text-slate-400">
                          {count as number} requests
                        </span>
                      </div>
                      <div className="flex items-center space-x-2">
                        <Progress value={percentage} className="h-2 flex-1" />
                        <span className="text-xs text-slate-400 min-w-[40px]">
                          {percentage.toFixed(1)}%
                        </span>
                      </div>
                    </div>
                  )
                })}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Error Analysis */}
      {metrics?.errors && metrics.errors.total > 0 && (
        <Card className="bg-slate-800 border-slate-700">
          <CardHeader>
            <CardTitle className="text-white flex items-center">
              <AlertTriangle className="h-5 w-5 mr-2 text-yellow-500" />
              Error Analysis
            </CardTitle>
            <CardDescription>
              {metrics.errors.total} errors detected ({errorRate.toFixed(1)}% error rate)
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <h4 className="text-sm font-medium text-slate-200 mb-3">Error Types</h4>
                <div className="space-y-2">
                  {Object.entries(metrics.errors.by_type).map(([type, count]) => (
                    <div key={type} className="flex items-center justify-between">
                      <span className="text-sm text-slate-300 capitalize">
                        {type.replace('_', ' ')}
                      </span>
                      <Badge variant="destructive" className="min-w-[60px] justify-center">
                        {count as number}
                      </Badge>
                    </div>
                  ))}
                </div>
              </div>
              
              <div>
                <h4 className="text-sm font-medium text-slate-200 mb-3">Error Distribution</h4>
                <div className="space-y-2">
                  {Object.entries(metrics.errors.by_type).map(([type, count]) => {
                    const percentage = ((count as number) / metrics.errors.total) * 100
                    return (
                      <div key={type} className="space-y-1">
                        <div className="flex justify-between text-sm">
                          <span className="text-slate-300 capitalize">{type.replace('_', ' ')}</span>
                          <span className="text-slate-400">{percentage.toFixed(1)}%</span>
                        </div>
                        <Progress value={percentage} className="h-2" />
                      </div>
                    )
                  })}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* System Status */}
      <Card className="bg-slate-800 border-slate-700">
        <CardHeader>
          <CardTitle className="text-white flex items-center">
            <Server className="h-5 w-5 mr-2" />
            Gateway Status
          </CardTitle>
          <CardDescription>Current gateway health and performance</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="text-center">
              <div className="flex items-center justify-center mb-2">
                <CheckCircle className="h-8 w-8 text-green-500" />
              </div>
              <h4 className="text-lg font-semibold text-white">Operational</h4>
              <p className="text-sm text-slate-400">All systems running normally</p>
            </div>
            
            <div className="text-center">
              <div className="text-2xl font-bold text-white mb-1">
                {formatNumber(metrics?.requests?.total || 0)}
              </div>
              <h4 className="text-sm font-medium text-slate-200">Total Requests</h4>
              <p className="text-xs text-slate-400">Processed successfully</p>
            </div>
            
            <div className="text-center">
              <div className="text-2xl font-bold text-white mb-1">
                {metrics?.requests?.average_response_time_ms || 0}ms
              </div>
              <h4 className="text-sm font-medium text-slate-200">Response Time</h4>
              <p className="text-xs text-slate-400">Average latency</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
