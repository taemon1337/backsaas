"use client"

import React, { useState } from 'react'
import { 
  ArrowLeft, 
  ArrowRight, 
  CheckCircle, 
  AlertTriangle,
  Database,
  Clock,
  Shield,
  Play,
  Pause,
  RotateCcw
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { cn } from '@/lib/utils'

interface MigrationWizardProps {
  schemaId: string
  onBack: () => void
  onComplete: () => void
}

type MigrationStep = 'review' | 'plan' | 'execute' | 'complete'

interface MigrationPlanStep {
  id: string
  type: 'expand' | 'backfill' | 'contract'
  operation: string
  description: string
  sql: string
  estimatedDuration: number
  riskLevel: 'low' | 'medium' | 'high'
  status: 'pending' | 'running' | 'completed' | 'failed'
}

export function MigrationWizard({ schemaId, onBack, onComplete }: MigrationWizardProps) {
  const [currentStep, setCurrentStep] = useState<MigrationStep>('review')
  const [isExecuting, setIsExecuting] = useState(false)
  const [executionProgress, setExecutionProgress] = useState(0)

  // Mock migration plan - in real app this would come from API
  const migrationPlan: MigrationPlanStep[] = [
    {
      id: '1',
      type: 'expand',
      operation: 'ADD COLUMN',
      description: 'Add new email_verified column',
      sql: 'ALTER TABLE customers ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;',
      estimatedDuration: 30,
      riskLevel: 'low',
      status: 'pending'
    },
    {
      id: '2',
      type: 'expand',
      operation: 'CREATE INDEX',
      description: 'Create index on email_verified column',
      sql: 'CREATE INDEX CONCURRENTLY idx_customers_email_verified ON customers(email_verified);',
      estimatedDuration: 120,
      riskLevel: 'low',
      status: 'pending'
    },
    {
      id: '3',
      type: 'backfill',
      operation: 'UPDATE DATA',
      description: 'Backfill email_verified values based on existing data',
      sql: 'UPDATE customers SET email_verified = TRUE WHERE email IS NOT NULL AND email != \'\';',
      estimatedDuration: 300,
      riskLevel: 'medium',
      status: 'pending'
    },
    {
      id: '4',
      type: 'contract',
      operation: 'ADD CONSTRAINT',
      description: 'Add NOT NULL constraint to email_verified',
      sql: 'ALTER TABLE customers ALTER COLUMN email_verified SET NOT NULL;',
      estimatedDuration: 60,
      riskLevel: 'medium',
      status: 'pending'
    }
  ]

  const totalEstimatedDuration = migrationPlan.reduce((sum, step) => sum + step.estimatedDuration, 0)
  const riskLevel = migrationPlan.some(step => step.riskLevel === 'high') ? 'high' :
                   migrationPlan.some(step => step.riskLevel === 'medium') ? 'medium' : 'low'

  const getRiskColor = (risk: string) => {
    switch (risk) {
      case 'high': return 'text-red-600 bg-red-100'
      case 'medium': return 'text-yellow-600 bg-yellow-100'
      case 'low': return 'text-green-600 bg-green-100'
      default: return 'text-gray-600 bg-gray-100'
    }
  }

  const getStepIcon = (type: string) => {
    switch (type) {
      case 'expand': return <Database className="h-4 w-4 text-blue-600" />
      case 'backfill': return <Clock className="h-4 w-4 text-yellow-600" />
      case 'contract': return <Shield className="h-4 w-4 text-green-600" />
      default: return <Database className="h-4 w-4" />
    }
  }

  const handleExecute = async () => {
    setIsExecuting(true)
    setCurrentStep('execute')
    
    // Simulate migration execution
    for (let i = 0; i <= 100; i += 10) {
      await new Promise(resolve => setTimeout(resolve, 500))
      setExecutionProgress(i)
    }
    
    setIsExecuting(false)
    setCurrentStep('complete')
  }

  const steps = [
    { id: 'review', name: 'Review Changes', description: 'Review the schema changes' },
    { id: 'plan', name: 'Migration Plan', description: 'Review the migration strategy' },
    { id: 'execute', name: 'Execute', description: 'Run the migration' },
    { id: 'complete', name: 'Complete', description: 'Migration finished' },
  ]

  const currentStepIndex = steps.findIndex(step => step.id === currentStep)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button variant="ghost" onClick={onBack}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Schemas
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Migration Wizard</h1>
            <p className="text-sm text-gray-600">
              Deploy your schema changes safely with zero downtime
            </p>
          </div>
        </div>
      </div>

      {/* Progress steps */}
      <div className="flex items-center justify-between">
        {steps.map((step, index) => (
          <div key={step.id} className="flex items-center">
            <div className="flex items-center">
              <div className={cn(
                'flex items-center justify-center w-8 h-8 rounded-full text-sm font-medium',
                index <= currentStepIndex
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-200 text-gray-600'
              )}>
                {index < currentStepIndex ? (
                  <CheckCircle className="h-4 w-4" />
                ) : (
                  index + 1
                )}
              </div>
              <div className="ml-3">
                <div className="text-sm font-medium text-gray-900">{step.name}</div>
                <div className="text-xs text-gray-500">{step.description}</div>
              </div>
            </div>
            {index < steps.length - 1 && (
              <div className={cn(
                'w-16 h-0.5 mx-4',
                index < currentStepIndex ? 'bg-blue-600' : 'bg-gray-200'
              )} />
            )}
          </div>
        ))}
      </div>

      {/* Step content */}
      {currentStep === 'review' && (
        <Card>
          <CardHeader>
            <CardTitle>Schema Changes Review</CardTitle>
            <CardDescription>
              Review the changes that will be applied to your schema
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <h4 className="font-medium text-blue-900 mb-2">Changes Summary</h4>
              <ul className="text-sm text-blue-800 space-y-1">
                <li>• Add email_verified field (boolean)</li>
                <li>• Create index on email_verified column</li>
                <li>• Backfill existing records</li>
                <li>• Add NOT NULL constraint</li>
              </ul>
            </div>
            
            <div className="flex justify-end">
              <Button onClick={() => setCurrentStep('plan')}>
                Continue to Migration Plan
                <ArrowRight className="h-4 w-4 ml-2" />
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {currentStep === 'plan' && (
        <div className="space-y-6">
          {/* Migration overview */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Card>
              <CardContent className="p-4">
                <div className="flex items-center space-x-2">
                  <Clock className="h-5 w-5 text-blue-600" />
                  <div>
                    <div className="text-sm font-medium">Estimated Duration</div>
                    <div className="text-lg font-semibold">{Math.ceil(totalEstimatedDuration / 60)} min</div>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-4">
                <div className="flex items-center space-x-2">
                  <AlertTriangle className="h-5 w-5 text-yellow-600" />
                  <div>
                    <div className="text-sm font-medium">Risk Level</div>
                    <div className={cn('text-lg font-semibold capitalize', getRiskColor(riskLevel))}>
                      {riskLevel}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-4">
                <div className="flex items-center space-x-2">
                  <Shield className="h-5 w-5 text-green-600" />
                  <div>
                    <div className="text-sm font-medium">Zero Downtime</div>
                    <div className="text-lg font-semibold text-green-600">Guaranteed</div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Migration steps */}
          <Card>
            <CardHeader>
              <CardTitle>Migration Plan</CardTitle>
              <CardDescription>
                Expand-Backfill-Contract strategy for zero-downtime deployment
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {migrationPlan.map((step, index) => (
                  <div key={step.id} className="flex items-start space-x-4 p-4 border rounded-lg">
                    <div className="flex-shrink-0">
                      {getStepIcon(step.type)}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center space-x-2">
                        <h4 className="text-sm font-medium text-gray-900">
                          Step {index + 1}: {step.operation}
                        </h4>
                        <span className={cn('px-2 py-1 text-xs rounded-full', getRiskColor(step.riskLevel))}>
                          {step.riskLevel} risk
                        </span>
                        <span className="text-xs text-gray-500">
                          ~{step.estimatedDuration}s
                        </span>
                      </div>
                      <p className="text-sm text-gray-600 mt-1">{step.description}</p>
                      <details className="mt-2">
                        <summary className="text-xs text-blue-600 cursor-pointer">View SQL</summary>
                        <pre className="mt-2 text-xs bg-gray-100 p-2 rounded overflow-x-auto">
                          {step.sql}
                        </pre>
                      </details>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          <div className="flex justify-between">
            <Button variant="outline" onClick={() => setCurrentStep('review')}>
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Review
            </Button>
            <Button onClick={handleExecute}>
              <Play className="h-4 w-4 mr-2" />
              Execute Migration
            </Button>
          </div>
        </div>
      )}

      {currentStep === 'execute' && (
        <Card>
          <CardHeader>
            <CardTitle>Executing Migration</CardTitle>
            <CardDescription>
              Please wait while we apply your schema changes...
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span>Progress</span>
                <span>{executionProgress}%</span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div 
                  className="bg-blue-600 h-2 rounded-full transition-all duration-500"
                  style={{ width: `${executionProgress}%` }}
                />
              </div>
            </div>

            <div className="text-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4" />
              <p className="text-sm text-gray-600">
                Applying changes... This may take a few minutes.
              </p>
            </div>

            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
              <div className="flex items-start space-x-2">
                <AlertTriangle className="h-4 w-4 text-yellow-600 mt-0.5" />
                <div className="text-sm text-yellow-800">
                  <strong>Important:</strong> Do not close this window or navigate away during migration.
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {currentStep === 'complete' && (
        <Card>
          <CardHeader className="text-center">
            <div className="mx-auto flex items-center justify-center w-12 h-12 rounded-full bg-green-100 mb-4">
              <CheckCircle className="w-6 h-6 text-green-600" />
            </div>
            <CardTitle>Migration Complete!</CardTitle>
            <CardDescription>
              Your schema has been successfully updated with zero downtime
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="bg-green-50 border border-green-200 rounded-lg p-4">
              <h4 className="font-medium text-green-900 mb-2">Migration Summary</h4>
              <ul className="text-sm text-green-800 space-y-1">
                <li>✓ Added email_verified column</li>
                <li>✓ Created performance index</li>
                <li>✓ Backfilled existing records</li>
                <li>✓ Applied constraints</li>
              </ul>
            </div>

            <div className="text-center space-y-4">
              <p className="text-sm text-gray-600">
                Your schema is now live and ready to use. All existing data has been preserved.
              </p>
              
              <div className="flex justify-center space-x-4">
                <Button variant="outline" onClick={() => window.open('/data', '_blank')}>
                  View Data
                </Button>
                <Button onClick={onComplete}>
                  Back to Schemas
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
