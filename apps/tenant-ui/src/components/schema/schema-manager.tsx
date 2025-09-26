"use client"

import React, { useState } from 'react'
import { 
  Plus, 
  Edit, 
  Trash2, 
  Play, 
  Database,
  Layers,
  FileText,
  Settings
} from 'lucide-react'
import { useTenant } from '@/lib/tenant-context'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { SchemaBuilder } from './schema-builder'
import { SchemaList } from './schema-list'
import { MigrationWizard } from './migration-wizard'

type ViewMode = 'list' | 'builder' | 'migration'

export function SchemaManager() {
  const [viewMode, setViewMode] = useState<ViewMode>('list')
  const [selectedSchema, setSelectedSchema] = useState<string | null>(null)
  const { tenant, hasPermission } = useTenant()

  // Mock schemas data - in real app this would come from API
  const schemas = [
    {
      id: '1',
      name: 'customers',
      displayName: 'Customers',
      description: 'Customer information and contact details',
      recordCount: 1234,
      status: 'active' as const,
      lastModified: '2024-01-15T10:30:00Z',
      version: '1.2.0',
    },
    {
      id: '2',
      name: 'orders',
      displayName: 'Orders',
      description: 'Customer orders and transaction history',
      recordCount: 5678,
      status: 'active' as const,
      lastModified: '2024-01-14T15:45:00Z',
      version: '1.1.0',
    },
    {
      id: '3',
      name: 'products',
      displayName: 'Products',
      description: 'Product catalog and inventory',
      recordCount: 890,
      status: 'draft' as const,
      lastModified: '2024-01-13T09:15:00Z',
      version: '0.9.0',
    },
  ]

  const handleCreateSchema = () => {
    setSelectedSchema(null)
    setViewMode('builder')
  }

  const handleEditSchema = (schemaId: string) => {
    setSelectedSchema(schemaId)
    setViewMode('builder')
  }

  const handlePlanMigration = (schemaId: string) => {
    setSelectedSchema(schemaId)
    setViewMode('migration')
  }

  const handleBackToList = () => {
    setSelectedSchema(null)
    setViewMode('list')
  }

  if (viewMode === 'builder') {
    return (
      <SchemaBuilder
        schemaId={selectedSchema}
        onBack={handleBackToList}
        onSave={handleBackToList}
      />
    )
  }

  if (viewMode === 'migration') {
    return (
      <MigrationWizard
        schemaId={selectedSchema!}
        onBack={handleBackToList}
        onComplete={handleBackToList}
      />
    )
  }

  return (
    <div className="space-y-6">
      {/* Quick stats */}
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-3">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <Layers className="h-6 w-6 text-blue-600" />
              </div>
              <div className="ml-4">
                <div className="text-2xl font-semibold text-gray-900">
                  {schemas.length}
                </div>
                <div className="text-sm text-gray-500">Total Schemas</div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <Database className="h-6 w-6 text-green-600" />
              </div>
              <div className="ml-4">
                <div className="text-2xl font-semibold text-gray-900">
                  {schemas.reduce((sum, schema) => sum + schema.recordCount, 0).toLocaleString()}
                </div>
                <div className="text-sm text-gray-500">Total Records</div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <Play className="h-6 w-6 text-purple-600" />
              </div>
              <div className="ml-4">
                <div className="text-2xl font-semibold text-gray-900">
                  {schemas.filter(s => s.status === 'active').length}
                </div>
                <div className="text-sm text-gray-500">Active Schemas</div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Actions */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-lg font-medium text-gray-900">Your Schemas</h2>
          <p className="text-sm text-gray-600">
            Manage your data structures and relationships
          </p>
        </div>
        
        {hasPermission('schema.create') && (
          <Button onClick={handleCreateSchema}>
            <Plus className="h-4 w-4 mr-2" />
            Create Schema
          </Button>
        )}
      </div>

      {/* Schema list */}
      <SchemaList
        schemas={schemas}
        onEdit={handleEditSchema}
        onPlanMigration={handlePlanMigration}
        canEdit={hasPermission('schema.update')}
        canDelete={hasPermission('schema.delete')}
      />

      {/* Getting started guide for empty state */}
      {schemas.length === 0 && (
        <Card>
          <CardHeader className="text-center">
            <div className="mx-auto flex items-center justify-center w-12 h-12 rounded-full bg-blue-100 mb-4">
              <Layers className="w-6 h-6 text-blue-600" />
            </div>
            <CardTitle>Get Started with Schemas</CardTitle>
            <CardDescription>
              Create your first data schema to start managing your business data
            </CardDescription>
          </CardHeader>
          <CardContent className="text-center space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-left">
              <div className="flex items-start space-x-3">
                <div className="flex-shrink-0 w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                  <span className="text-sm font-medium text-blue-600">1</span>
                </div>
                <div>
                  <h4 className="text-sm font-medium text-gray-900">Design Your Schema</h4>
                  <p className="text-sm text-gray-600">
                    Use our visual builder to define your data structure
                  </p>
                </div>
              </div>
              
              <div className="flex items-start space-x-3">
                <div className="flex-shrink-0 w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                  <span className="text-sm font-medium text-blue-600">2</span>
                </div>
                <div>
                  <h4 className="text-sm font-medium text-gray-900">Validate & Preview</h4>
                  <p className="text-sm text-gray-600">
                    Review your schema and see the generated database structure
                  </p>
                </div>
              </div>
              
              <div className="flex items-start space-x-3">
                <div className="flex-shrink-0 w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                  <span className="text-sm font-medium text-blue-600">3</span>
                </div>
                <div>
                  <h4 className="text-sm font-medium text-gray-900">Deploy Safely</h4>
                  <p className="text-sm text-gray-600">
                    Use our migration wizard for zero-downtime deployment
                  </p>
                </div>
              </div>
            </div>
            
            <div className="pt-4">
              <Button onClick={handleCreateSchema} size="lg">
                <Plus className="h-4 w-4 mr-2" />
                Create Your First Schema
              </Button>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
