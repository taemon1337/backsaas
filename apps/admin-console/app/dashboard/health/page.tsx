"use client"

import { useState, useEffect } from 'react'
import { useHealthSummary, useHealthServices, useHealthStatus, useApiMutation } from '@/lib/hooks/use-api'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { 
  Activity, 
  Server, 
  Database, 
  RefreshCw, 
  CheckCircle, 
  XCircle, 
  Clock,
  TrendingUp,
  AlertTriangle
} from 'lucide-react'
import { useToast } from '@/components/ui/use-toast'

interface ServiceCoverage {
  name: string
  coverage: number
  lines_covered: number
  lines_total: number
  last_updated: string
  status: string
}

interface SystemSummary {
  timestamp: string
  overall_coverage: number
  services: Record<string, any>
  total_lines: number
  covered_lines: number
  tests_passed: number
  tests_failed: number
  trends: Record<string, any>
}

interface ServiceStatus {
  collecting: boolean
  priority: number
}

interface SystemStatus {
  timestamp: string
  services: Record<string, ServiceStatus>
  uptime: string
}

export default function SystemHealthPage() {
  const { toast } = useToast()
  
  // Use API hooks for data fetching
  const { 
    data: summary, 
    loading: summaryLoading, 
    error: summaryError, 
    refetch: refetchSummary 
  } = useHealthSummary()
  
  const { 
    data: servicesData, 
    loading: servicesLoading, 
    error: servicesError, 
    refetch: refetchServices 
  } = useHealthServices()
  
  const { 
    data: status, 
    loading: statusLoading, 
    error: statusError, 
    refetch: refetchStatus 
  } = useHealthStatus()

  // Mutation hook for triggering collection
  const { mutate: triggerMutation, loading: triggering } = useApiMutation()

  // Combine loading states
  const loading = summaryLoading || servicesLoading || statusLoading
  const refreshing = triggering

  // Convert services data to array format
  const services = Array.isArray(servicesData) ? servicesData : []

  const fetchHealthData = async () => {
    await Promise.all([refetchSummary(), refetchServices(), refetchStatus()])
  }

  const triggerCollection = async () => {
    const { apiClient } = await import('@/lib/api-client')
    
    const result = await triggerMutation(
      () => apiClient.triggerCoverageCollection(),
      undefined
    )
    
    if (result) {
      toast({
        title: "Collection Started",
        description: (result as any).message || "Coverage collection has been triggered for all services",
      })
      // Refresh data after a short delay
      setTimeout(fetchHealthData, 2000)
    }
  }

  // Show errors if any
  useEffect(() => {
    const errors = [summaryError, servicesError, statusError].filter(Boolean)
    if (errors.length > 0) {
      toast({
        title: "Error",
        description: errors[0] || "Failed to fetch system health data",
        variant: "destructive",
      })
    }
  }, [summaryError, servicesError, statusError, toast])

  // Set up auto-refresh every 30 seconds
  useEffect(() => {
    const interval = setInterval(fetchHealthData, 30000)
    return () => clearInterval(interval)
  }, [])

  const getStatusIcon = (serviceName: string) => {
    const serviceStatus = status?.services[serviceName]
    if (serviceStatus?.collecting) {
      return <RefreshCw className="h-4 w-4 animate-spin text-blue-500" />
    }
    return <CheckCircle className="h-4 w-4 text-green-500" />
  }

  const getCoverageColor = (coverage: number) => {
    if (coverage >= 80) return "text-green-600"
    if (coverage >= 60) return "text-yellow-600"
    return "text-red-600"
  }

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <RefreshCw className="h-8 w-8 animate-spin text-blue-500" />
          <span className="ml-2 text-slate-300">Loading system health data...</span>
        </div>
      </div>
    )
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">System Health</h1>
          <p className="text-slate-400">Monitor service health, coverage, and performance</p>
        </div>
        <div className="flex gap-2">
          <Button 
            onClick={fetchHealthData} 
            disabled={refreshing}
            variant="outline"
            size="sm"
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button 
            onClick={triggerCollection}
            size="sm"
          >
            <Activity className="h-4 w-4 mr-2" />
            Collect Coverage
          </Button>
        </div>
      </div>

      {/* Summary Cards */}
      {summary && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card className="bg-slate-800 border-slate-700">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-slate-200">Total Services</CardTitle>
              <Server className="h-4 w-4 text-blue-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-white">{Object.keys(summary.services).length}</div>
              <p className="text-xs text-slate-400">Active services monitored</p>
            </CardContent>
          </Card>

          <Card className="bg-slate-800 border-slate-700">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-slate-200">Average Coverage</CardTitle>
              <TrendingUp className="h-4 w-4 text-green-500" />
            </CardHeader>
            <CardContent>
              <div className={`text-2xl font-bold ${getCoverageColor(summary.overall_coverage)}`}>
                {summary.overall_coverage.toFixed(1)}%
              </div>
              <p className="text-xs text-slate-400">Across all services</p>
            </CardContent>
          </Card>

          <Card className="bg-slate-800 border-slate-700">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-slate-200">Total Lines</CardTitle>
              <Database className="h-4 w-4 text-purple-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-white">{summary.total_lines.toLocaleString()}</div>
              <p className="text-xs text-slate-400">{summary.covered_lines.toLocaleString()} covered</p>
            </CardContent>
          </Card>

          <Card className="bg-slate-800 border-slate-700">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-slate-200">Last Updated</CardTitle>
              <Clock className="h-4 w-4 text-slate-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-white">
                {new Date(summary.timestamp).toLocaleTimeString()}
              </div>
              <p className="text-xs text-slate-400">
                {new Date(summary.timestamp).toLocaleDateString()}
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Services List */}
      <Card className="bg-slate-800 border-slate-700">
        <CardHeader>
          <CardTitle className="text-white">Service Coverage</CardTitle>
          <CardDescription className="text-slate-400">
            Code coverage and health status for each service
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {services.length > 0 ? services.map((service) => (
              <div key={service.name} className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
                <div className="flex items-center space-x-4">
                  {getStatusIcon(service.name)}
                  <div>
                    <h3 className="font-medium text-white">{service.name}</h3>
                    <p className="text-sm text-slate-400">
                      {service.lines_covered.toLocaleString()} / {service.lines_total.toLocaleString()} lines covered
                    </p>
                  </div>
                </div>
                
                <div className="flex items-center space-x-4">
                  <div className="text-right">
                    <div className={`text-lg font-semibold ${getCoverageColor(service.coverage)}`}>
                      {service.coverage.toFixed(1)}%
                    </div>
                    <div className="text-xs text-slate-400">
                      {new Date(service.last_updated).toLocaleString()}
                    </div>
                  </div>
                  
                  <div className="w-24">
                    <Progress 
                      value={service.coverage} 
                      className="h-2"
                    />
                  </div>
                  
                  <Badge 
                    variant={service.status === 'healthy' ? 'default' : 'destructive'}
                    className="min-w-[70px] justify-center"
                  >
                    {service.status}
                  </Badge>
                </div>
              </div>
            )) : (
              <div className="text-center py-8">
                <AlertTriangle className="h-12 w-12 text-yellow-500 mx-auto mb-4" />
                <p className="text-slate-400">No service data available</p>
                <p className="text-sm text-slate-500">Try triggering a coverage collection</p>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
