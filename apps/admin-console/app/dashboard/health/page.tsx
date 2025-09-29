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
  Server, 
  Database, 
  RefreshCw, 
  CheckCircle, 
  XCircle, 
  Clock,
  TrendingUp,
  AlertTriangle,
  ChevronDown,
  ChevronRight,
  TestTube,
  FileText,
  Target,
  Zap
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
  uptime: string
}

export default function SystemHealthPage() {
  const { toast } = useToast()
  const [isCollecting, setIsCollecting] = useState(false)
  const [showTestDetails, setShowTestDetails] = useState(false)
  
  // Use React Query for data fetching
  const { 
    data: summary, 
    isLoading: summaryLoading, 
    error: summaryError, 
    refetch: refetchSummary 
  } = useQuery({
    queryKey: ['health-summary'],
    queryFn: () => apiClient.getHealthSummary(),
    refetchInterval: 30000,
  })
  
  const { 
    data: servicesData, 
    isLoading: servicesLoading, 
    error: servicesError, 
    refetch: refetchServices 
  } = useQuery({
    queryKey: ['health-services'],
    queryFn: () => apiClient.getHealthServices(),
    refetchInterval: 30000,
  })
  
  const { 
    data: status, 
    isLoading: statusLoading, 
    error: statusError, 
    refetch: refetchStatus 
  } = useQuery({
    queryKey: ['health-status'],
    queryFn: () => apiClient.getHealthStatus(),
    refetchInterval: 30000,
  })

  const { 
    data: systemTests, 
    isLoading: testsLoading, 
    error: testsError, 
    refetch: refetchTests 
  } = useQuery({
    queryKey: ['system-tests'],
    queryFn: () => apiClient.getSystemTests(),
    refetchInterval: 60000, // Refresh every minute
  })

  // Combine loading states
  const loading = summaryLoading || servicesLoading || statusLoading || testsLoading
  const refreshing = isCollecting

  // Convert services data to array format for display
  const services = summary?.services ? Object.entries(summary.services).map(([name, coverage]) => ({
    name: name.charAt(0).toUpperCase() + name.slice(1),
    coverage: coverage as number,
    lines_covered: Math.floor((coverage as number) * summary.total_lines / 100 / Object.keys(summary.services).length),
    lines_total: Math.floor(summary.total_lines / Object.keys(summary.services).length),
    last_updated: summary.timestamp,
    status: (coverage as number) > 15 ? 'healthy' : 'warning'
  })) : []

  const fetchHealthData = async () => {
    await Promise.all([refetchSummary(), refetchServices(), refetchStatus(), refetchTests()])
  }

  const runSystemTests = async () => {
    setIsCollecting(true)
    try {
      const result = await apiClient.runSystemTests()
      
      toast({
        title: "Tests Completed",
        description: `System tests completed with ${result.summary?.success_rate?.toFixed(1)}% success rate`,
      })
      
      // Refresh test data
      setTimeout(refetchTests, 2000)
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to run system tests",
        variant: "destructive",
      })
    } finally {
      setIsCollecting(false)
    }
  }

  const triggerCollection = async () => {
    setIsCollecting(true)
    try {
      const result = await apiClient.triggerCoverageCollection()
      
      toast({
        title: "Collection Started",
        description: result.message || "Coverage collection has been triggered for all services",
      })
      
      // Refresh data after a short delay
      setTimeout(fetchHealthData, 2000)
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to trigger coverage collection",
        variant: "destructive",
      })
    } finally {
      setIsCollecting(false)
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
            onClick={runSystemTests}
            disabled={refreshing}
            variant="outline"
            size="sm"
          >
            <CheckCircle className="h-4 w-4 mr-2" />
            Run Tests
          </Button>
          <Button 
            onClick={triggerCollection}
            disabled={refreshing}
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

      {/* Service Status Overview */}
      {status && (
        <Card className="bg-slate-800 border-slate-700">
          <CardHeader>
            <CardTitle className="text-white">Service Status</CardTitle>
            <CardDescription className="text-slate-400">
              Real-time status and collection activity
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              {Object.entries(status.services).map(([name, serviceStatus]) => (
                <div key={name} className="p-4 bg-slate-700/50 rounded-lg">
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-medium text-white capitalize">{name}</h3>
                    {serviceStatus.collecting ? (
                      <RefreshCw className="h-4 w-4 animate-spin text-blue-500" />
                    ) : (
                      <CheckCircle className="h-4 w-4 text-green-500" />
                    )}
                  </div>
                  <div className="space-y-1">
                    <div className="flex justify-between text-sm">
                      <span className="text-slate-400">Status:</span>
                      <span className={serviceStatus.collecting ? 'text-blue-400' : 'text-green-400'}>
                        {serviceStatus.collecting ? 'Collecting' : 'Ready'}
                      </span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-slate-400">Priority:</span>
                      <span className="text-slate-300">{serviceStatus.priority}</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
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

      {/* System Tests */}
      {systemTests && (
        <Card className="bg-slate-800 border-slate-700">
          <CardHeader>
            <CardTitle className="text-white flex items-center space-x-2">
              <CheckCircle className="h-5 w-5 text-green-500" />
              <span>System Tests</span>
            </CardTitle>
            <CardDescription className="text-slate-400">
              Automated test results for user flows, error handling, and system validation
            </CardDescription>
          </CardHeader>
          <CardContent>
            {/* Test Summary */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
              <div className="p-4 bg-slate-700/50 rounded-lg">
                <div className="text-2xl font-bold text-white">{systemTests.summary?.total_tests || 0}</div>
                <div className="text-sm text-slate-400">Total Tests</div>
              </div>
              <div className="p-4 bg-slate-700/50 rounded-lg">
                <div className="text-2xl font-bold text-green-400">{systemTests.summary?.passed_tests || 0}</div>
                <div className="text-sm text-slate-400">Passed</div>
              </div>
              <div className="p-4 bg-slate-700/50 rounded-lg">
                <div className="text-2xl font-bold text-red-400">{systemTests.summary?.failed_tests || 0}</div>
                <div className="text-sm text-slate-400">Failed</div>
              </div>
              <div className="p-4 bg-slate-700/50 rounded-lg">
                <div className={`text-2xl font-bold ${systemTests.summary?.success_rate >= 90 ? 'text-green-400' : systemTests.summary?.success_rate >= 70 ? 'text-yellow-400' : 'text-red-400'}`}>
                  {systemTests.summary?.success_rate?.toFixed(1) || 0}%
                </div>
                <div className="text-sm text-slate-400">Success Rate</div>
              </div>
            </div>

            {/* Test Suites */}
            <div className="space-y-4">
              {systemTests.suites?.map((suite: any, index: number) => (
                <div key={index} className="p-4 bg-slate-700/50 rounded-lg">
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center space-x-3">
                      {suite.status === 'pass' ? (
                        <CheckCircle className="h-5 w-5 text-green-500" />
                      ) : suite.status === 'fail' ? (
                        <XCircle className="h-5 w-5 text-red-500" />
                      ) : (
                        <AlertTriangle className="h-5 w-5 text-yellow-500" />
                      )}
                      <div>
                        <h3 className="font-medium text-white">{suite.name}</h3>
                        <p className="text-sm text-slate-400">{suite.description}</p>
                      </div>
                    </div>
                    <div className="text-right">
                      <Badge 
                        variant={suite.status === 'pass' ? 'default' : suite.status === 'fail' ? 'destructive' : 'secondary'}
                        className="mb-1"
                      >
                        {suite.status.toUpperCase()}
                      </Badge>
                      <div className="text-xs text-slate-400">
                        {suite.duration}ms
                      </div>
                    </div>
                  </div>
                  
                  <div className="flex items-center space-x-4 text-sm">
                    <span className="text-green-400">✅ {suite.passed_tests} passed</span>
                    {suite.failed_tests > 0 && (
                      <span className="text-red-400">❌ {suite.failed_tests} failed</span>
                    )}
                    {suite.warning_tests > 0 && (
                      <span className="text-yellow-400">⚠️ {suite.warning_tests} warnings</span>
                    )}
                    <span className="text-slate-400 ml-auto">
                      Last run: {new Date(suite.timestamp).toLocaleString()}
                    </span>
                  </div>
                </div>
              )) || (
                <div className="text-center py-8">
                  <CheckCircle className="h-12 w-12 text-slate-500 mx-auto mb-4" />
                  <p className="text-slate-400">No test results available</p>
                  <p className="text-sm text-slate-500">Click "Run Tests" to execute system tests</p>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Detailed Test Results Table */}
      <Card className="bg-slate-800 border-slate-700">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <TestTube className="h-5 w-5 text-blue-500" />
              <CardTitle className="text-white">Detailed Test Results</CardTitle>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowTestDetails(!showTestDetails)}
              className="text-slate-400 hover:text-white"
            >
              {showTestDetails ? (
                <>
                  <ChevronDown className="h-4 w-4 mr-2" />
                  Hide Details
                </>
              ) : (
                <>
                  <ChevronRight className="h-4 w-4 mr-2" />
                  Show Details
                </>
              )}
            </Button>
          </div>
          <CardDescription className="text-slate-400">
            Comprehensive test coverage and results for all platform services
          </CardDescription>
        </CardHeader>
        
        {showTestDetails && (
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-slate-700">
                    <th className="text-left py-3 px-4 text-slate-300 font-medium">Service</th>
                    <th className="text-left py-3 px-4 text-slate-300 font-medium">Health</th>
                    <th className="text-left py-3 px-4 text-slate-300 font-medium">Coverage</th>
                    <th className="text-left py-3 px-4 text-slate-300 font-medium">Unit Tests</th>
                    <th className="text-left py-3 px-4 text-slate-300 font-medium">Integration</th>
                    <th className="text-left py-3 px-4 text-slate-300 font-medium">E2E Tests</th>
                    <th className="text-left py-3 px-4 text-slate-300 font-medium">Components</th>
                    <th className="text-left py-3 px-4 text-slate-300 font-medium">Last Updated</th>
                  </tr>
                </thead>
                <tbody>
                  {servicesData && Object.entries(servicesData).map(([serviceName, serviceData]: [string, any]) => (
                    <tr key={serviceName} className="border-b border-slate-700/50 hover:bg-slate-700/30">
                      <td className="py-4 px-4">
                        <div className="flex items-center space-x-2">
                          <Server className="h-4 w-4 text-blue-500" />
                          <span className="font-medium text-white capitalize">{serviceName}</span>
                        </div>
                      </td>
                      <td className="py-4 px-4">
                        <div className="flex items-center space-x-2">
                          {serviceData.service_health?.status === 'healthy' ? (
                            <CheckCircle className="h-4 w-4 text-green-500" />
                          ) : serviceData.service_health?.status === 'unhealthy' ? (
                            <XCircle className="h-4 w-4 text-red-500" />
                          ) : (
                            <Clock className="h-4 w-4 text-yellow-500" />
                          )}
                          <span className={`text-sm ${
                            serviceData.service_health?.status === 'healthy' ? 'text-green-400' :
                            serviceData.service_health?.status === 'unhealthy' ? 'text-red-400' :
                            'text-yellow-400'
                          }`}>
                            {serviceData.service_health?.status || 'unknown'}
                          </span>
                          {serviceData.service_health?.response_time_ms && (
                            <span className="text-xs text-slate-500">
                              ({serviceData.service_health.response_time_ms}ms)
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="py-4 px-4">
                        <div className="flex items-center space-x-2">
                          <div className="w-16 bg-slate-700 rounded-full h-2">
                            <div 
                              className={`h-2 rounded-full ${
                                serviceData.overall >= 80 ? 'bg-green-500' :
                                serviceData.overall >= 60 ? 'bg-yellow-500' :
                                'bg-red-500'
                              }`}
                              style={{ width: `${Math.min(serviceData.overall || 0, 100)}%` }}
                            />
                          </div>
                          <span className="text-sm text-slate-300 font-mono">
                            {(serviceData.overall || 0).toFixed(1)}%
                          </span>
                        </div>
                        <div className="text-xs text-slate-500 mt-1">
                          {serviceData.covered_lines || 0}/{serviceData.lines || 0} lines
                        </div>
                      </td>
                      <td className="py-4 px-4">
                        <div className="flex items-center space-x-2">
                          {serviceData.test_status?.unit_tests?.exists ? (
                            <CheckCircle className="h-4 w-4 text-green-500" />
                          ) : (
                            <XCircle className="h-4 w-4 text-red-500" />
                          )}
                          <span className="text-sm text-slate-300">
                            {serviceData.test_status?.unit_tests?.count || 0} tests
                          </span>
                        </div>
                        <div className="text-xs text-slate-500">
                          Status: {serviceData.test_status?.unit_tests?.status || 'unknown'}
                        </div>
                      </td>
                      <td className="py-4 px-4">
                        <div className="flex items-center space-x-2">
                          {serviceData.test_status?.integration_tests?.exists ? (
                            <CheckCircle className="h-4 w-4 text-green-500" />
                          ) : (
                            <XCircle className="h-4 w-4 text-red-500" />
                          )}
                          <span className="text-sm text-slate-300">
                            {serviceData.test_status?.integration_tests?.count || 0} tests
                          </span>
                        </div>
                        <div className="text-xs text-slate-500">
                          Status: {serviceData.test_status?.integration_tests?.status || 'none'}
                        </div>
                      </td>
                      <td className="py-4 px-4">
                        <div className="flex items-center space-x-2">
                          {serviceData.test_status?.e2e_tests?.exists ? (
                            <CheckCircle className="h-4 w-4 text-green-500" />
                          ) : (
                            <XCircle className="h-4 w-4 text-red-500" />
                          )}
                          <span className="text-sm text-slate-300">
                            {serviceData.test_status?.e2e_tests?.count || 0} tests
                          </span>
                        </div>
                        <div className="text-xs text-slate-500">
                          Status: {serviceData.test_status?.e2e_tests?.status || 'none'}
                        </div>
                      </td>
                      <td className="py-4 px-4">
                        <div className="flex items-center space-x-2">
                          <Target className="h-4 w-4 text-purple-500" />
                          <span className="text-sm text-slate-300">
                            {serviceData.test_status?.tested_components || 0}/
                            {serviceData.test_status?.total_components || 0}
                          </span>
                        </div>
                        <div className="text-xs text-slate-500">
                          {(serviceData.test_status?.test_coverage_percent || 0).toFixed(1)}% covered
                        </div>
                      </td>
                      <td className="py-4 px-4">
                        <div className="text-sm text-slate-300">
                          {serviceData.timestamp ? new Date(serviceData.timestamp).toLocaleString() : 'Never'}
                        </div>
                        <div className="text-xs text-slate-500">
                          {serviceData.timestamp ? 
                            `${Math.round((Date.now() - new Date(serviceData.timestamp).getTime()) / 60000)}m ago` : 
                            'No data'
                          }
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Test Summary Cards */}
            <div className="mt-6 grid grid-cols-1 md:grid-cols-4 gap-4">
              <div className="p-4 bg-slate-700/50 rounded-lg">
                <div className="flex items-center space-x-2 mb-2">
                  <FileText className="h-4 w-4 text-blue-500" />
                  <span className="text-sm font-medium text-slate-300">Total Tests</span>
                </div>
                <div className="text-2xl font-bold text-white">
                  {servicesData ? Object.values(servicesData).reduce((sum: number, service: any) => 
                    sum + (service.test_status?.unit_tests?.count || 0) + 
                          (service.test_status?.integration_tests?.count || 0) + 
                          (service.test_status?.e2e_tests?.count || 0), 0
                  ) : 0}
                </div>
              </div>
              
              <div className="p-4 bg-slate-700/50 rounded-lg">
                <div className="flex items-center space-x-2 mb-2">
                  <Zap className="h-4 w-4 text-green-500" />
                  <span className="text-sm font-medium text-slate-300">Avg Coverage</span>
                </div>
                <div className="text-2xl font-bold text-white">
                  {servicesData ? (
                    Object.values(servicesData).reduce((sum: number, service: any) => sum + (service.overall || 0), 0) / 
                    Object.keys(servicesData).length
                  ).toFixed(1) : 0}%
                </div>
              </div>
              
              <div className="p-4 bg-slate-700/50 rounded-lg">
                <div className="flex items-center space-x-2 mb-2">
                  <CheckCircle className="h-4 w-4 text-green-500" />
                  <span className="text-sm font-medium text-slate-300">Services Healthy</span>
                </div>
                <div className="text-2xl font-bold text-white">
                  {servicesData ? Object.values(servicesData).filter((service: any) => 
                    service.service_health?.status === 'healthy'
                  ).length : 0}/{servicesData ? Object.keys(servicesData).length : 0}
                </div>
              </div>
              
              <div className="p-4 bg-slate-700/50 rounded-lg">
                <div className="flex items-center space-x-2 mb-2">
                  <Target className="h-4 w-4 text-purple-500" />
                  <span className="text-sm font-medium text-slate-300">Components Tested</span>
                </div>
                <div className="text-2xl font-bold text-white">
                  {servicesData ? Object.values(servicesData).reduce((sum: number, service: any) => 
                    sum + (service.test_status?.tested_components || 0), 0
                  ) : 0}/{servicesData ? Object.values(servicesData).reduce((sum: number, service: any) => 
                    sum + (service.test_status?.total_components || 0), 0
                  ) : 0}
                </div>
              </div>
            </div>
          </CardContent>
        )}
      </Card>
    </div>
  )
}
