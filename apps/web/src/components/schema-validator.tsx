"use client"

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { 
  CheckCircle, 
  XCircle, 
  AlertTriangle, 
  Database, 
  FileText,
  Code,
  RefreshCw
} from 'lucide-react'

interface Schema {
  name: string
  version: string
  entities: any[]
  metadata: any
}

interface ValidationResult {
  valid: boolean
  errors: ValidationError[]
  warnings: ValidationWarning[]
  summary: {
    entities: number
    fields: number
    relationships: number
    functions: number
  }
}

interface ValidationError {
  type: 'schema' | 'entity' | 'field' | 'relationship' | 'function'
  path: string
  message: string
  severity: 'error' | 'warning'
}

interface ValidationWarning {
  type: string
  path: string
  message: string
  suggestion?: string
}

interface SchemaValidatorProps {
  schema: Schema
}

export function SchemaValidator({ schema }: SchemaValidatorProps) {
  const [validationResult, setValidationResult] = useState<ValidationResult | null>(null)
  const [isValidating, setIsValidating] = useState(false)
  const [sqlPreview, setSqlPreview] = useState<string>('')

  useEffect(() => {
    validateSchema()
  }, [schema])

  const validateSchema = async () => {
    setIsValidating(true)
    
    try {
      // Simulate validation API call
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      const errors: ValidationError[] = []
      const warnings: ValidationWarning[] = []
      
      // Schema-level validation
      if (!schema.name) {
        errors.push({
          type: 'schema',
          path: 'name',
          message: 'Schema name is required',
          severity: 'error'
        })
      } else if (!/^[a-z][a-z0-9_-]*$/.test(schema.name)) {
        errors.push({
          type: 'schema',
          path: 'name',
          message: 'Schema name must start with a letter and contain only lowercase letters, numbers, hyphens, and underscores',
          severity: 'error'
        })
      }
      
      if (!schema.version) {
        errors.push({
          type: 'schema',
          path: 'version',
          message: 'Schema version is required',
          severity: 'error'
        })
      } else if (!/^\d+\.\d+\.\d+$/.test(schema.version)) {
        warnings.push({
          type: 'schema',
          path: 'version',
          message: 'Version should follow semantic versioning (e.g., 1.0.0)',
          suggestion: 'Use format: MAJOR.MINOR.PATCH'
        })
      }
      
      if (schema.entities.length === 0) {
        errors.push({
          type: 'schema',
          path: 'entities',
          message: 'At least one entity is required',
          severity: 'error'
        })
      }
      
      // Entity-level validation
      const entityNames = new Set<string>()
      let totalFields = 0
      let totalRelationships = 0
      let totalFunctions = 0
      
      schema.entities.forEach((entity, entityIndex) => {
        const entityPath = `entities[${entityIndex}]`
        
        if (!entity.name) {
          errors.push({
            type: 'entity',
            path: `${entityPath}.name`,
            message: 'Entity name is required',
            severity: 'error'
          })
        } else {
          if (entityNames.has(entity.name)) {
            errors.push({
              type: 'entity',
              path: `${entityPath}.name`,
              message: `Duplicate entity name: ${entity.name}`,
              severity: 'error'
            })
          }
          entityNames.add(entity.name)
          
          if (!/^[a-z][a-z0-9_]*$/.test(entity.name)) {
            errors.push({
              type: 'entity',
              path: `${entityPath}.name`,
              message: 'Entity name must start with a letter and contain only lowercase letters, numbers, and underscores',
              severity: 'error'
            })
          }
        }
        
        // Field validation
        if (!entity.fields || entity.fields.length === 0) {
          errors.push({
            type: 'entity',
            path: `${entityPath}.fields`,
            message: 'Entity must have at least one field',
            severity: 'error'
          })
        } else {
          const fieldNames = new Set<string>()
          let hasIdField = false
          
          entity.fields.forEach((field: any, fieldIndex: number) => {
            const fieldPath = `${entityPath}.fields[${fieldIndex}]`
            totalFields++
            
            if (!field.name) {
              errors.push({
                type: 'field',
                path: `${fieldPath}.name`,
                message: 'Field name is required',
                severity: 'error'
              })
            } else {
              if (fieldNames.has(field.name)) {
                errors.push({
                  type: 'field',
                  path: `${fieldPath}.name`,
                  message: `Duplicate field name: ${field.name}`,
                  severity: 'error'
                })
              }
              fieldNames.add(field.name)
              
              if (field.name === 'id') {
                hasIdField = true
                if (field.type !== 'uuid' && field.type !== 'integer') {
                  warnings.push({
                    type: 'field',
                    path: `${fieldPath}.type`,
                    message: 'ID field should typically be uuid or integer type',
                    suggestion: 'Consider using uuid for distributed systems'
                  })
                }
              }
            }
            
            if (!field.type) {
              errors.push({
                type: 'field',
                path: `${fieldPath}.type`,
                message: 'Field type is required',
                severity: 'error'
              })
            } else {
              const validTypes = ['string', 'integer', 'float', 'boolean', 'timestamp', 'uuid', 'json', 'text']
              if (!validTypes.includes(field.type)) {
                warnings.push({
                  type: 'field',
                  path: `${fieldPath}.type`,
                  message: `Unknown field type: ${field.type}`,
                  suggestion: `Valid types: ${validTypes.join(', ')}`
                })
              }
            }
          })
          
          if (!hasIdField) {
            warnings.push({
              type: 'entity',
              path: `${entityPath}`,
              message: 'Entity should have an id field',
              suggestion: 'Add an id field with type uuid or integer'
            })
          }
        }
        
        // Count relationships and functions
        if (entity.relationships) {
          totalRelationships += entity.relationships.length
        }
        if (entity.functions) {
          totalFunctions += entity.functions.length
        }
      })
      
      const result: ValidationResult = {
        valid: errors.length === 0,
        errors,
        warnings,
        summary: {
          entities: schema.entities.length,
          fields: totalFields,
          relationships: totalRelationships,
          functions: totalFunctions
        }
      }
      
      setValidationResult(result)
      
      // Generate SQL preview
      generateSqlPreview()
      
    } finally {
      setIsValidating(false)
    }
  }
  
  const generateSqlPreview = () => {
    let sql = '-- Generated SQL for schema: ' + schema.name + '\n\n'
    
    schema.entities.forEach(entity => {
      sql += `CREATE TABLE ${entity.name} (\n`
      
      const fieldSql = entity.fields.map((field: any) => {
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
            fieldDef += 'TIMESTAMP'
            break
          case 'json':
            fieldDef += 'JSONB'
            break
          default:
            fieldDef += 'VARCHAR(255)'
        }
        
        if (field.required) fieldDef += ' NOT NULL'
        if (field.unique) fieldDef += ' UNIQUE'
        if (field.default !== undefined) fieldDef += ` DEFAULT ${field.default}`
        
        return fieldDef
      }).join(',\n')
      
      sql += fieldSql
      
      // Add primary key if id field exists
      const hasId = entity.fields.some((f: any) => f.name === 'id')
      if (hasId) {
        sql += ',\n  PRIMARY KEY (id)'
      }
      
      sql += '\n);\n\n'
    })
    
    setSqlPreview(sql)
  }

  if (isValidating) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="flex items-center space-x-2">
          <RefreshCw className="h-5 w-5 animate-spin" />
          <span>Validating schema...</span>
        </div>
      </div>
    )
  }

  if (!validationResult) {
    return (
      <div className="flex items-center justify-center h-64">
        <Button onClick={validateSchema}>
          <CheckCircle className="h-4 w-4 mr-2" />
          Validate Schema
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Validation Summary */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center space-x-2">
              {validationResult.valid ? (
                <CheckCircle className="h-5 w-5 text-green-500" />
              ) : (
                <XCircle className="h-5 w-5 text-red-500" />
              )}
              <span>Validation Results</span>
            </CardTitle>
            <Button variant="outline" size="sm" onClick={validateSchema}>
              <RefreshCw className="h-4 w-4 mr-2" />
              Re-validate
            </Button>
          </div>
          <CardDescription>
            Schema validation status and summary
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="text-center">
              <div className="text-2xl font-bold">{validationResult.summary.entities}</div>
              <div className="text-sm text-muted-foreground">Entities</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold">{validationResult.summary.fields}</div>
              <div className="text-sm text-muted-foreground">Fields</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold">{validationResult.summary.relationships}</div>
              <div className="text-sm text-muted-foreground">Relationships</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold">{validationResult.summary.functions}</div>
              <div className="text-sm text-muted-foreground">Functions</div>
            </div>
          </div>
          
          <div className="flex items-center space-x-4">
            <Badge variant={validationResult.valid ? "default" : "destructive"}>
              {validationResult.valid ? "Valid" : "Invalid"}
            </Badge>
            {validationResult.errors.length > 0 && (
              <Badge variant="destructive">
                {validationResult.errors.length} Error{validationResult.errors.length !== 1 ? 's' : ''}
              </Badge>
            )}
            {validationResult.warnings.length > 0 && (
              <Badge variant="outline">
                {validationResult.warnings.length} Warning{validationResult.warnings.length !== 1 ? 's' : ''}
              </Badge>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Errors */}
      {validationResult.errors.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2 text-red-600">
              <XCircle className="h-5 w-5" />
              <span>Validation Errors</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {validationResult.errors.map((error, index) => (
                <div key={index} className="flex items-start space-x-3 p-3 bg-red-50 border border-red-200 rounded-md">
                  <XCircle className="h-4 w-4 text-red-500 mt-0.5 flex-shrink-0" />
                  <div className="flex-1">
                    <div className="font-medium text-red-800">{error.path}</div>
                    <div className="text-sm text-red-600">{error.message}</div>
                  </div>
                  <Badge variant="outline" className="text-xs">
                    {error.type}
                  </Badge>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Warnings */}
      {validationResult.warnings.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2 text-yellow-600">
              <AlertTriangle className="h-5 w-5" />
              <span>Warnings & Suggestions</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {validationResult.warnings.map((warning, index) => (
                <div key={index} className="flex items-start space-x-3 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
                  <AlertTriangle className="h-4 w-4 text-yellow-500 mt-0.5 flex-shrink-0" />
                  <div className="flex-1">
                    <div className="font-medium text-yellow-800">{warning.path}</div>
                    <div className="text-sm text-yellow-600">{warning.message}</div>
                    {warning.suggestion && (
                      <div className="text-sm text-yellow-500 mt-1">
                        💡 {warning.suggestion}
                      </div>
                    )}
                  </div>
                  <Badge variant="outline" className="text-xs">
                    {warning.type}
                  </Badge>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* SQL Preview */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Database className="h-5 w-5" />
            <span>SQL Preview</span>
          </CardTitle>
          <CardDescription>
            Generated SQL statements for this schema
          </CardDescription>
        </CardHeader>
        <CardContent>
          <pre className="bg-muted p-4 rounded-md text-sm overflow-x-auto">
            <code>{sqlPreview}</code>
          </pre>
        </CardContent>
      </Card>
    </div>
  )
}
