"use client"

import React from 'react'
import Link from 'next/link'
import { 
  Database, 
  Users, 
  Activity, 
  TrendingUp,
  Plus,
  ArrowRight,
  Layers,
  Workflow
} from 'lucide-react'
import { useTenant } from '@/lib/tenant-context'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export function DashboardOverview() {
  const { tenant, user } = useTenant()

  // Mock data - in real app this would come from API
  const stats = [
    {
      name: 'Total Records',
      value: '12,345',
      change: '+12%',
      changeType: 'positive' as const,
      icon: Database,
    },
    {
      name: 'Active Users',
      value: '23',
      change: '+3',
      changeType: 'positive' as const,
      icon: Users,
    },
    {
      name: 'API Calls',
      value: '45.2K',
      change: '+8%',
      changeType: 'positive' as const,
      icon: Activity,
    },
    {
      name: 'Growth Rate',
      value: '15.3%',
      change: '+2.1%',
      changeType: 'positive' as const,
      icon: TrendingUp,
    },
  ]

  const quickActions = [
    {
      name: 'Design Schema',
      description: 'Create and modify your data schemas',
      href: '/schema',
      icon: Layers,
      color: 'bg-blue-500',
    },
    {
      name: 'Add Data',
      description: 'Import or create new records',
      href: '/data/create',
      icon: Plus,
      color: 'bg-green-500',
    },
    {
      name: 'Create Workflow',
      description: 'Automate your business processes',
      href: '/workflows/create',
      icon: Workflow,
      color: 'bg-purple-500',
    },
    {
      name: 'View Analytics',
      description: 'Analyze your data and trends',
      href: '/analytics',
      icon: TrendingUp,
      color: 'bg-orange-500',
    },
  ]

  const recentActivity = [
    {
      id: 1,
      type: 'schema',
      action: 'created',
      target: 'Customer schema',
      user: 'John Doe',
      timestamp: '2 hours ago',
    },
    {
      id: 2,
      type: 'data',
      action: 'imported',
      target: '150 customer records',
      user: 'Jane Smith',
      timestamp: '4 hours ago',
    },
    {
      id: 3,
      type: 'workflow',
      action: 'activated',
      target: 'Order processing workflow',
      user: 'Mike Johnson',
      timestamp: '6 hours ago',
    },
  ]

  return (
    <div className="space-y-8">
      {/* Welcome section */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">
          Welcome back, {user?.name?.split(' ')[0] || 'there'}!
        </h1>
        <p className="mt-1 text-sm text-gray-600">
          Here's what's happening with {tenant?.name || 'your organization'} today.
        </p>
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => {
          const Icon = stat.icon
          return (
            <Card key={stat.name}>
              <CardContent className="p-6">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <Icon className="h-6 w-6 text-gray-400" />
                  </div>
                  <div className="ml-4 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        {stat.name}
                      </dt>
                      <dd className="flex items-baseline">
                        <div className="text-2xl font-semibold text-gray-900">
                          {stat.value}
                        </div>
                        <div className={`ml-2 flex items-baseline text-sm font-semibold ${
                          stat.changeType === 'positive' ? 'text-green-600' : 'text-red-600'
                        }`}>
                          {stat.change}
                        </div>
                      </dd>
                    </dl>
                  </div>
                </div>
              </CardContent>
            </Card>
          )
        })}
      </div>

      {/* Quick actions */}
      <div>
        <h2 className="text-lg font-medium text-gray-900 mb-4">Quick Actions</h2>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {quickActions.map((action) => {
            const Icon = action.icon
            return (
              <Link key={action.name} href={action.href}>
                <Card className="hover:shadow-md transition-shadow cursor-pointer">
                  <CardContent className="p-6">
                    <div className="flex items-center">
                      <div className={`flex-shrink-0 p-3 rounded-lg ${action.color}`}>
                        <Icon className="h-6 w-6 text-white" />
                      </div>
                      <div className="ml-4">
                        <h3 className="text-sm font-medium text-gray-900">
                          {action.name}
                        </h3>
                        <p className="text-sm text-gray-500">
                          {action.description}
                        </p>
                      </div>
                      <ArrowRight className="ml-auto h-4 w-4 text-gray-400" />
                    </div>
                  </CardContent>
                </Card>
              </Link>
            )
          })}
        </div>
      </div>

      {/* Recent activity and schema overview */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Recent activity */}
        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
            <CardDescription>
              Latest changes in your organization
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {recentActivity.map((activity) => (
                <div key={activity.id} className="flex items-center space-x-4">
                  <div className="flex-shrink-0">
                    <div className="h-8 w-8 rounded-full bg-gray-100 flex items-center justify-center">
                      <Activity className="h-4 w-4 text-gray-600" />
                    </div>
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-gray-900">
                      <span className="font-medium">{activity.user}</span>
                      {' '}{activity.action}{' '}
                      <span className="font-medium">{activity.target}</span>
                    </p>
                    <p className="text-sm text-gray-500">{activity.timestamp}</p>
                  </div>
                </div>
              ))}
            </div>
            <div className="mt-6">
              <Link href="/activity">
                <Button variant="outline" className="w-full">
                  View all activity
                </Button>
              </Link>
            </div>
          </CardContent>
        </Card>

        {/* Schema overview */}
        <Card>
          <CardHeader>
            <CardTitle>Schema Overview</CardTitle>
            <CardDescription>
              Your current data schemas
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-900">Customers</p>
                  <p className="text-sm text-gray-500">1,234 records</p>
                </div>
                <div className="text-sm text-gray-500">Active</div>
              </div>
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-900">Orders</p>
                  <p className="text-sm text-gray-500">5,678 records</p>
                </div>
                <div className="text-sm text-gray-500">Active</div>
              </div>
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-900">Products</p>
                  <p className="text-sm text-gray-500">890 records</p>
                </div>
                <div className="text-sm text-gray-500">Active</div>
              </div>
            </div>
            <div className="mt-6 space-y-2">
              <Link href="/schema">
                <Button variant="outline" className="w-full">
                  Manage Schemas
                </Button>
              </Link>
              <Link href="/schema/create">
                <Button className="w-full">
                  <Plus className="h-4 w-4 mr-2" />
                  Create New Schema
                </Button>
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
