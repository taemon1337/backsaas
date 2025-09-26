"use client"

import React from 'react'
import { 
  Edit, 
  Trash2, 
  Play, 
  Pause,
  Database,
  Calendar,
  MoreVertical,
  GitBranch
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { formatDate } from '@/lib/utils'

interface Schema {
  id: string
  name: string
  displayName: string
  description: string
  recordCount: number
  status: 'active' | 'draft' | 'deprecated'
  lastModified: string
  version: string
}

interface SchemaListProps {
  schemas: Schema[]
  onEdit: (schemaId: string) => void
  onPlanMigration: (schemaId: string) => void
  canEdit: boolean
  canDelete: boolean
}

export function SchemaList({ 
  schemas, 
  onEdit, 
  onPlanMigration, 
  canEdit, 
  canDelete 
}: SchemaListProps) {
  const getStatusColor = (status: Schema['status']) => {
    switch (status) {
      case 'active':
        return 'bg-green-100 text-green-800'
      case 'draft':
        return 'bg-yellow-100 text-yellow-800'
      case 'deprecated':
        return 'bg-red-100 text-red-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const getStatusIcon = (status: Schema['status']) => {
    switch (status) {
      case 'active':
        return <Play className="h-3 w-3" />
      case 'draft':
        return <Pause className="h-3 w-3" />
      case 'deprecated':
        return <Pause className="h-3 w-3" />
      default:
        return <Pause className="h-3 w-3" />
    }
  }

  return (
    <div className="space-y-4">
      {schemas.map((schema) => (
        <Card key={schema.id} className="hover:shadow-md transition-shadow">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              {/* Schema info */}
              <div className="flex-1 min-w-0">
                <div className="flex items-center space-x-3">
                  <div className="flex-shrink-0">
                    <Database className="h-5 w-5 text-gray-400" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center space-x-2">
                      <h3 className="text-lg font-medium text-gray-900 truncate">
                        {schema.displayName}
                      </h3>
                      <span className="text-sm text-gray-500">
                        ({schema.name})
                      </span>
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(schema.status)}`}>
                        {getStatusIcon(schema.status)}
                        <span className="ml-1 capitalize">{schema.status}</span>
                      </span>
                    </div>
                    <p className="text-sm text-gray-600 mt-1">
                      {schema.description}
                    </p>
                  </div>
                </div>

                {/* Schema stats */}
                <div className="mt-4 flex items-center space-x-6 text-sm text-gray-500">
                  <div className="flex items-center space-x-1">
                    <Database className="h-4 w-4" />
                    <span>{schema.recordCount.toLocaleString()} records</span>
                  </div>
                  <div className="flex items-center space-x-1">
                    <GitBranch className="h-4 w-4" />
                    <span>v{schema.version}</span>
                  </div>
                  <div className="flex items-center space-x-1">
                    <Calendar className="h-4 w-4" />
                    <span>Updated {formatDate(schema.lastModified, { 
                      month: 'short', 
                      day: 'numeric',
                      hour: '2-digit',
                      minute: '2-digit'
                    })}</span>
                  </div>
                </div>
              </div>

              {/* Actions */}
              <div className="flex items-center space-x-2 ml-4">
                {canEdit && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onEdit(schema.id)}
                  >
                    <Edit className="h-4 w-4 mr-1" />
                    Edit
                  </Button>
                )}
                
                {schema.status === 'active' && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onPlanMigration(schema.id)}
                  >
                    <GitBranch className="h-4 w-4 mr-1" />
                    Migrate
                  </Button>
                )}

                {/* More actions dropdown (simplified for now) */}
                <Button variant="ghost" size="sm">
                  <MoreVertical className="h-4 w-4" />
                </Button>
              </div>
            </div>

            {/* Quick actions bar */}
            <div className="mt-4 pt-4 border-t border-gray-200">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-4 text-sm">
                  <button className="text-blue-600 hover:text-blue-800 font-medium">
                    View Data
                  </button>
                  <button className="text-blue-600 hover:text-blue-800 font-medium">
                    Export Schema
                  </button>
                  <button className="text-blue-600 hover:text-blue-800 font-medium">
                    API Docs
                  </button>
                </div>
                
                <div className="text-xs text-gray-500">
                  Last deployed: {formatDate(schema.lastModified, { 
                    month: 'short', 
                    day: 'numeric' 
                  })}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
