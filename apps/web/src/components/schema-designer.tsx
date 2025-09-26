"use client"

import { useState, useCallback, useEffect } from 'react'
import { useValidateSchema, useCreateSchema, useUpdateSchema } from '@/lib/hooks/use-api'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { 
  Save, 
  Play, 
  FileText, 
  Database, 
  Settings, 
  Plus,
  Trash2,
  Eye,
  Code
} from 'lucide-react'
import { MonacoEditor } from './monaco-editor'
import { SchemaValidator } from './schema-validator'
import { MigrationPlanner } from './migration-planner'

interface Entity {
  name: string
  fields: Field[]
  relationships: Relationship[]
  functions: Function[]
}

interface Field {
  name: string
  type: string
  required: boolean
  unique: boolean
  default?: any
  validation?: any
}

interface Relationship {
  name: string
  type: 'one-to-one' | 'one-to-many' | 'many-to-many'
  target: string
}

interface Function {
  name: string
  type: 'validation' | 'transformation' | 'hook'
  trigger: string
  code: string
}

interface Schema {
  name: string
  version: string
  entities: Entity[]
  metadata: {
    created_at: string
    updated_at: string
    author: string
  }
}

const defaultSchema: Schema = {
  name: "example-schema",
  version: "1.0.0",
  entities: [
    {
      name: "users",
      fields: [
        { name: "id", type: "uuid", required: true, unique: true },
        { name: "email", type: "string", required: true, unique: true },
        { name: "name", type: "string", required: true, unique: false },
        { name: "created_at", type: "timestamp", required: true, unique: false }
      ],
      relationships: [],
      functions: []
    }
  ],
  metadata: {
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    author: "admin"
  }
}

export function SchemaDesigner() {
  const [schema, setSchema] = useState<Schema>(defaultSchema)
  const [yamlContent, setYamlContent] = useState('')
  const [activeTab, setActiveTab] = useState<'visual' | 'yaml' | 'preview' | 'migration'>('visual')
  const [validationErrors, setValidationErrors] = useState<string[]>([])
  const [isValidating, setIsValidating] = useState(false)
  const [isSaving, setIsSaving] = useState(false)

  const handleSchemaChange = useCallback((newSchema: Schema) => {
    setSchema(newSchema)
    // Convert to YAML when schema changes
    // This would use js-yaml in real implementation
    setYamlContent(JSON.stringify(newSchema, null, 2))
  }, [])

  const handleYamlChange = useCallback((newYaml: string) => {
    setYamlContent(newYaml)
    try {
      // Parse YAML and update schema
      // This would use js-yaml in real implementation
      const parsedSchema = JSON.parse(newYaml)
      setSchema(parsedSchema)
      setValidationErrors([])
    } catch (error) {
      setValidationErrors([`Invalid YAML: ${error.message}`])
    }
  }, [])

  const validateSchema = async () => {
    setIsValidating(true)
    try {
      // Simulate validation API call
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      const errors: string[] = []
      
      // Basic validation
      if (!schema.name) errors.push('Schema name is required')
      if (!schema.version) errors.push('Schema version is required')
      if (schema.entities.length === 0) errors.push('At least one entity is required')
      
      // Entity validation
      schema.entities.forEach((entity, index) => {
        if (!entity.name) errors.push(`Entity ${index + 1}: Name is required`)
        if (entity.fields.length === 0) errors.push(`Entity ${entity.name}: At least one field is required`)
        
        entity.fields.forEach((field, fieldIndex) => {
          if (!field.name) errors.push(`Entity ${entity.name}, Field ${fieldIndex + 1}: Name is required`)
          if (!field.type) errors.push(`Entity ${entity.name}, Field ${field.name}: Type is required`)
        })
      })
      
      setValidationErrors(errors)
    } finally {
      setIsValidating(false)
    }
  }

  const saveSchema = async () => {
    setIsSaving(true)
    try {
      // Simulate save API call
      await new Promise(resolve => setTimeout(resolve, 1500))
      
      // Update metadata
      const updatedSchema = {
        ...schema,
        metadata: {
          ...schema.metadata,
          updated_at: new Date().toISOString()
        }
      }
      setSchema(updatedSchema)
      
      // In real implementation, this would publish to Redis Streams
      console.log('Schema saved and published to event stream:', updatedSchema)
    } finally {
      setIsSaving(false)
    }
  }

  const addEntity = () => {
    const newEntity: Entity = {
      name: `entity_${schema.entities.length + 1}`,
      fields: [
        { name: "id", type: "uuid", required: true, unique: true }
      ],
      relationships: [],
      functions: []
    }
    
    handleSchemaChange({
      ...schema,
      entities: [...schema.entities, newEntity]
    })
  }

  const removeEntity = (index: number) => {
    const newEntities = schema.entities.filter((_, i) => i !== index)
    handleSchemaChange({
      ...schema,
      entities: newEntities
    })
  }

  const updateEntity = (index: number, updatedEntity: Entity) => {
    const newEntities = [...schema.entities]
    newEntities[index] = updatedEntity
    handleSchemaChange({
      ...schema,
      entities: newEntities
    })
  }

  return (
    <div className="h-screen flex flex-col">
      {/* Header */}
      <div className="border-b bg-background p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <h1 className="text-2xl font-bold">Schema Designer</h1>
            <Badge variant="outline">{schema.name} v{schema.version}</Badge>
          </div>
          
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              size="sm"
              onClick={validateSchema}
              disabled={isValidating}
            >
              <Settings className="h-4 w-4 mr-2" />
              {isValidating ? 'Validating...' : 'Validate'}
            </Button>
            
            <Button
              size="sm"
              onClick={saveSchema}
              disabled={isSaving || validationErrors.length > 0}
            >
              <Save className="h-4 w-4 mr-2" />
              {isSaving ? 'Publishing...' : 'Publish Schema'}
            </Button>
          </div>
        </div>
        
        {/* Validation Errors */}
        {validationErrors.length > 0 && (
          <div className="mt-4 p-3 bg-destructive/10 border border-destructive/20 rounded-md">
            <h4 className="text-sm font-medium text-destructive mb-2">Validation Errors:</h4>
            <ul className="text-sm text-destructive space-y-1">
              {validationErrors.map((error, index) => (
                <li key={index}>• {error}</li>
              ))}
            </ul>
          </div>
        )}
      </div>

      {/* Tab Navigation */}
      <div className="border-b bg-muted/30">
        <div className="flex space-x-1 p-1">
          {[
            { id: 'visual', label: 'Visual Designer', icon: Database },
            { id: 'yaml', label: 'YAML Editor', icon: Code },
            { id: 'preview', label: 'Preview', icon: Eye },
            { id: 'migration', label: 'Migration Plan', icon: Play }
          ].map(({ id, label, icon: Icon }) => (
            <Button
              key={id}
              variant={activeTab === id ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setActiveTab(id as any)}
              className="flex items-center space-x-2"
            >
              <Icon className="h-4 w-4" />
              <span>{label}</span>
            </Button>
          ))}
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-hidden">
        {activeTab === 'visual' && (
          <div className="h-full p-6 overflow-auto">
            <div className="space-y-6">
              {/* Schema Metadata */}
              <Card>
                <CardHeader>
                  <CardTitle>Schema Information</CardTitle>
                  <CardDescription>Basic schema metadata and configuration</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="text-sm font-medium">Schema Name</label>
                      <Input
                        value={schema.name}
                        onChange={(e) => handleSchemaChange({ ...schema, name: e.target.value })}
                        placeholder="my-schema"
                      />
                    </div>
                    <div>
                      <label className="text-sm font-medium">Version</label>
                      <Input
                        value={schema.version}
                        onChange={(e) => handleSchemaChange({ ...schema, version: e.target.value })}
                        placeholder="1.0.0"
                      />
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Entities */}
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <h2 className="text-xl font-semibold">Entities</h2>
                  <Button onClick={addEntity} size="sm">
                    <Plus className="h-4 w-4 mr-2" />
                    Add Entity
                  </Button>
                </div>

                {schema.entities.map((entity, index) => (
                  <Card key={index}>
                    <CardHeader>
                      <div className="flex items-center justify-between">
                        <CardTitle className="text-lg">{entity.name}</CardTitle>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => removeEntity(index)}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-4">
                        <div>
                          <label className="text-sm font-medium">Entity Name</label>
                          <Input
                            value={entity.name}
                            onChange={(e) => {
                              const updatedEntity = { ...entity, name: e.target.value }
                              updateEntity(index, updatedEntity)
                            }}
                            placeholder="entity_name"
                          />
                        </div>
                        
                        <div>
                          <h4 className="text-sm font-medium mb-2">Fields ({entity.fields.length})</h4>
                          <div className="space-y-2">
                            {entity.fields.map((field, fieldIndex) => (
                              <div key={fieldIndex} className="flex items-center space-x-2 p-2 border rounded">
                                <Input
                                  value={field.name}
                                  onChange={(e) => {
                                    const updatedFields = [...entity.fields]
                                    updatedFields[fieldIndex] = { ...field, name: e.target.value }
                                    updateEntity(index, { ...entity, fields: updatedFields })
                                  }}
                                  placeholder="field_name"
                                  className="flex-1"
                                />
                                <Input
                                  value={field.type}
                                  onChange={(e) => {
                                    const updatedFields = [...entity.fields]
                                    updatedFields[fieldIndex] = { ...field, type: e.target.value }
                                    updateEntity(index, { ...entity, fields: updatedFields })
                                  }}
                                  placeholder="string"
                                  className="w-32"
                                />
                                <div className="flex space-x-1">
                                  <Badge variant={field.required ? "default" : "outline"}>
                                    {field.required ? "Required" : "Optional"}
                                  </Badge>
                                  <Badge variant={field.unique ? "default" : "outline"}>
                                    {field.unique ? "Unique" : "Non-unique"}
                                  </Badge>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </div>
          </div>
        )}

        {activeTab === 'yaml' && (
          <div className="h-full">
            <MonacoEditor
              value={yamlContent}
              onChange={handleYamlChange}
              language="yaml"
            />
          </div>
        )}

        {activeTab === 'preview' && (
          <div className="h-full p-6 overflow-auto">
            <SchemaValidator schema={schema} />
          </div>
        )}

        {activeTab === 'migration' && (
          <div className="h-full p-6 overflow-auto">
            <MigrationPlanner schema={schema} />
          </div>
        )}
      </div>
    </div>
  )
}
