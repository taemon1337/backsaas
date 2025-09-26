"use client"

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { 
  Play, 
  Pause, 
  CheckCircle, 
  Clock, 
  AlertTriangle,
  ArrowRight,
  Database,
  FileText,
  RefreshCw
} from 'lucide-react'

interface Schema {
  name: string
  version: string
  entities: any[]
  metadata: any
}

interface MigrationStep {
  id: string
  type: 'expand' | 'backfill' | 'contract'
  operation: string
  description: string
  sql: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  duration?: number
  error?: string
}

interface MigrationPlan {
  id: string
  fromVersion: string
  toVersion: string
  steps: MigrationStep[]
  estimatedDuration: number
  riskLevel: 'low' | 'medium' | 'high'
  canRollback: boolean
}

interface MigrationPlannerProps {
  schema: Schema
}

export function MigrationPlanner({ schema }: MigrationPlannerProps) {
  const [migrationPlan, setMigrationPlan] = useState<MigrationPlan | null>(null)
  const [isGenerating, setIsGenerating] = useState(false)
  const [isExecuting, setIsExecuting] = useState(false)
  const [currentStep, setCurrentStep] = useState<number>(-1)

  useEffect(() => {
    generateMigrationPlan()
  }, [schema])

  const generateMigrationPlan = async () => {
    setIsGenerating(true)
    
    try {
      // Simulate migration plan generation
      await new Promise(resolve => setTimeout(resolve, 1500))
      
      const steps: MigrationStep[] = []
      
      // Generate steps based on schema changes
      schema.entities.forEach((entity, index) => {
        // Expand phase - Add new columns/tables
        steps.push({
          id: `expand-${index}`,
          type: 'expand',
          operation: 'CREATE_TABLE',
          description: `Create table ${entity.name}`,
          sql: generateCreateTableSQL(entity),
          status: 'pending'
        })
        
        // Add indexes for unique fields
        entity.fields?.forEach((field: any, fieldIndex: number) => {
          if (field.unique && field.name !== 'id') {
            steps.push({
              id: `expand-index-${index}-${fieldIndex}`,
              type: 'expand',
              operation: 'CREATE_INDEX',
              description: `Create unique index on ${entity.name}.${field.name}`,
              sql: `CREATE UNIQUE INDEX idx_${entity.name}_${field.name} ON ${entity.name} (${field.name});`,
              status: 'pending'
            })
          }
        })
      })
      
      // Backfill phase - Migrate data
      if (schema.entities.length > 0) {
        steps.push({
          id: 'backfill-data',
          type: 'backfill',
          operation: 'MIGRATE_DATA',
          description: 'Migrate existing data to new schema',
          sql: '-- Data migration scripts would be generated here\n-- Based on the differences between old and new schema',
          status: 'pending'
        })
      }
      
      // Contract phase - Remove old structures (if any)
      steps.push({
        id: 'contract-cleanup',
        type: 'contract',
        operation: 'CLEANUP',
        description: 'Clean up temporary migration artifacts',
        sql: '-- Cleanup temporary tables and indexes\n-- Remove old unused columns',
        status: 'pending'
      })
      
      const plan: MigrationPlan = {
        id: `migration-${Date.now()}`,
        fromVersion: '0.0.0', // Would be fetched from current schema
        toVersion: schema.version,
        steps,
        estimatedDuration: steps.length * 2, // 2 seconds per step estimate
        riskLevel: steps.length > 5 ? 'high' : steps.length > 2 ? 'medium' : 'low',
        canRollback: true
      }
      
      setMigrationPlan(plan)
    } finally {
      setIsGenerating(false)
    }
  }
  
  const generateCreateTableSQL = (entity: any): string => {
    let sql = `CREATE TABLE ${entity.name} (\n`
    
    const fieldSql = entity.fields?.map((field: any) => {
      let fieldDef = `  ${field.name} `
      
      switch (field.type) {
        case 'uuid':
          fieldDef += 'UUID'
          break
        case 'string':
          fieldDef += 'VARCHAR(255)'
          break
        case 'text':
          fieldDef += 'TEXT'
          break
        case 'integer':
          fieldDef += 'INTEGER'
          break
        case 'float':
          fieldDef += 'DECIMAL'
          break
        case 'boolean':
          fieldDef += 'BOOLEAN'
          break
        case 'timestamp':
          fieldDef += 'TIMESTAMP DEFAULT CURRENT_TIMESTAMP'
          break
        case 'json':
          fieldDef += 'JSONB'
          break
        default:
          fieldDef += 'VARCHAR(255)'
      }
      
      if (field.required) fieldDef += ' NOT NULL'
      if (field.unique) fieldDef += ' UNIQUE'
      
      return fieldDef
    }).join(',\n') || '  id UUID PRIMARY KEY'
    
    sql += fieldSql
    
    // Add primary key if id field exists
    const hasId = entity.fields?.some((f: any) => f.name === 'id')
    if (hasId) {
      sql += ',\n  PRIMARY KEY (id)'
    }
    
    sql += '\n);'
    return sql
  }

  const executeMigration = async () => {
    if (!migrationPlan) return
    
    setIsExecuting(true)
    setCurrentStep(0)
    
    try {
      for (let i = 0; i < migrationPlan.steps.length; i++) {
        setCurrentStep(i)
        
        // Update step status to running
        const updatedPlan = { ...migrationPlan }
        updatedPlan.steps[i].status = 'running'
        setMigrationPlan(updatedPlan)
        
        // Simulate step execution
        const stepDuration = Math.random() * 2000 + 1000 // 1-3 seconds
        await new Promise(resolve => setTimeout(resolve, stepDuration))
        
        // Update step status to completed
        updatedPlan.steps[i].status = 'completed'
        updatedPlan.steps[i].duration = Math.round(stepDuration)
        setMigrationPlan({ ...updatedPlan })
      }
      
      setCurrentStep(-1)
    } catch (error) {
      // Handle migration failure
      if (migrationPlan && currentStep >= 0) {
        const updatedPlan = { ...migrationPlan }
        updatedPlan.steps[currentStep].status = 'failed'
        updatedPlan.steps[currentStep].error = 'Migration step failed'
        setMigrationPlan(updatedPlan)
      }
    } finally {
      setIsExecuting(false)
    }
  }

  const getStepIcon = (step: MigrationStep, index: number) => {
    if (step.status === 'running' || (isExecuting && index === currentStep)) {
      return <RefreshCw className="h-4 w-4 animate-spin text-blue-500" />
    }
    if (step.status === 'completed') {
      return <CheckCircle className="h-4 w-4 text-green-500" />
    }
    if (step.status === 'failed') {
      return <AlertTriangle className="h-4 w-4 text-red-500" />
    }
    return <Clock className="h-4 w-4 text-gray-400" />
  }

  const getStepColor = (step: MigrationStep) => {
    switch (step.type) {
      case 'expand':
        return 'bg-blue-50 border-blue-200'
      case 'backfill':
        return 'bg-yellow-50 border-yellow-200'
      case 'contract':
        return 'bg-green-50 border-green-200'
      default:
        return 'bg-gray-50 border-gray-200'
    }
  }

  if (isGenerating) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="flex items-center space-x-2">
          <RefreshCw className="h-5 w-5 animate-spin" />
          <span>Generating migration plan...</span>
        </div>
      </div>
    )
  }

  if (!migrationPlan) {
    return (
      <div className="flex items-center justify-center h-64">
        <Button onClick={generateMigrationPlan}>
          <FileText className="h-4 w-4 mr-2" />
          Generate Migration Plan
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Migration Overview */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center space-x-2">
              <Database className="h-5 w-5" />
              <span>Migration Plan</span>
            </CardTitle>
            <div className="flex items-center space-x-2">
              <Button variant="outline" size="sm" onClick={generateMigrationPlan}>
                <RefreshCw className="h-4 w-4 mr-2" />
                Regenerate
              </Button>
              <Button 
                size="sm" 
                onClick={executeMigration}
                disabled={isExecuting}
              >
                <Play className="h-4 w-4 mr-2" />
                {isExecuting ? 'Executing...' : 'Execute Migration'}
              </Button>
            </div>
          </div>
          <CardDescription>
            Migration from v{migrationPlan.fromVersion} to v{migrationPlan.toVersion}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="text-center">
              <div className="text-2xl font-bold">{migrationPlan.steps.length}</div>
              <div className="text-sm text-muted-foreground">Steps</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold">{migrationPlan.estimatedDuration}s</div>
              <div className="text-sm text-muted-foreground">Est. Duration</div>
            </div>
            <div className="text-center">
              <Badge variant={
                migrationPlan.riskLevel === 'low' ? 'default' :
                migrationPlan.riskLevel === 'medium' ? 'outline' : 'destructive'
              }>
                {migrationPlan.riskLevel.toUpperCase()} RISK
              </Badge>
            </div>
            <div className="text-center">
              <Badge variant={migrationPlan.canRollback ? 'default' : 'destructive'}>
                {migrationPlan.canRollback ? 'ROLLBACK OK' : 'NO ROLLBACK'}
              </Badge>
            </div>
          </div>
          
          {migrationPlan.riskLevel === 'high' && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-md mb-4">
              <div className="flex items-center space-x-2">
                <AlertTriangle className="h-4 w-4 text-red-500" />
                <span className="text-sm font-medium text-red-800">High Risk Migration</span>
              </div>
              <p className="text-sm text-red-600 mt-1">
                This migration involves significant schema changes. Consider running in maintenance mode.
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Migration Steps */}
      <Card>
        <CardHeader>
          <CardTitle>Migration Steps</CardTitle>
          <CardDescription>
            Expand-Backfill-Contract pattern for zero-downtime migrations
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {migrationPlan.steps.map((step, index) => (
              <div key={step.id} className={`p-4 border rounded-lg ${getStepColor(step)}`}>
                <div className="flex items-start space-x-3">
                  <div className="flex-shrink-0 mt-1">
                    {getStepIcon(step, index)}
                  </div>
                  
                  <div className="flex-1">
                    <div className="flex items-center space-x-2 mb-2">
                      <Badge variant="outline" className="text-xs">
                        {step.type.toUpperCase()}
                      </Badge>
                      <span className="font-medium">{step.operation}</span>
                      {step.duration && (
                        <span className="text-xs text-muted-foreground">
                          ({step.duration}ms)
                        </span>
                      )}
                    </div>
                    
                    <p className="text-sm text-muted-foreground mb-3">
                      {step.description}
                    </p>
                    
                    <details className="text-sm">
                      <summary className="cursor-pointer text-blue-600 hover:text-blue-800">
                        View SQL
                      </summary>
                      <pre className="mt-2 p-2 bg-white border rounded text-xs overflow-x-auto">
                        <code>{step.sql}</code>
                      </pre>
                    </details>
                    
                    {step.error && (
                      <div className="mt-2 p-2 bg-red-100 border border-red-200 rounded text-sm text-red-600">
                        Error: {step.error}
                      </div>
                    )}
                  </div>
                  
                  {index < migrationPlan.steps.length - 1 && (
                    <ArrowRight className="h-4 w-4 text-muted-foreground mt-6" />
                  )}
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Migration Phases Explanation */}
      <Card>
        <CardHeader>
          <CardTitle>Migration Strategy</CardTitle>
          <CardDescription>
            Understanding the Expand-Backfill-Contract pattern
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
              <h4 className="font-medium text-blue-800 mb-2">1. Expand Phase</h4>
              <p className="text-sm text-blue-600">
                Add new tables, columns, and indexes. The application can run with both old and new schema.
              </p>
            </div>
            
            <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
              <h4 className="font-medium text-yellow-800 mb-2">2. Backfill Phase</h4>
              <p className="text-sm text-yellow-600">
                Migrate existing data to the new schema structure. This can be done gradually.
              </p>
            </div>
            
            <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
              <h4 className="font-medium text-green-800 mb-2">3. Contract Phase</h4>
              <p className="text-sm text-green-600">
                Remove old unused columns and tables. Clean up temporary migration artifacts.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
