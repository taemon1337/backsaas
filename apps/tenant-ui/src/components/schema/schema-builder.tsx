"use client"

import React, { useState } from 'react'
import { 
  ArrowLeft, 
  Plus, 
  Trash2, 
  Save,
  Eye,
  Database,
  Type,
  Hash,
  Calendar,
  Mail,
  Phone,
  Link,
  FileText,
  Image,
  ToggleLeft
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { Textarea } from '@/components/ui/textarea'
import { cn } from '@/lib/utils'
import type { EntitySchema, SchemaField, FieldType } from '@/lib/types'

interface SchemaBuilderProps {
  schemaId: string | null
  onBack: () => void
  onSave: (schema: EntitySchema) => void
}

export function SchemaBuilder({ schemaId, onBack, onSave }: SchemaBuilderProps) {
  const [schema, setSchema] = useState<Partial<EntitySchema>>({
    name: '',
    displayName: '',
    description: '',
    fields: [
      {
        name: 'id',
        displayName: 'ID',
        type: 'uuid',
        required: true,
        unique: true,
        ui: { readonly: true, hidden: false }
      }
    ],
    relationships: [],
    permissions: [],
    functions: []
  })

  const [activeTab, setActiveTab] = useState<'basic' | 'fields' | 'preview'>('basic')

  const fieldTypes: { value: FieldType; label: string; icon: React.ReactNode; description: string }[] = [
    { value: 'string', label: 'Text', icon: <Type className="h-4 w-4" />, description: 'Short text up to 255 characters' },
    { value: 'text', label: 'Long Text', icon: <FileText className="h-4 w-4" />, description: 'Long text content' },
    { value: 'integer', label: 'Number', icon: <Hash className="h-4 w-4" />, description: 'Whole numbers' },
    { value: 'decimal', label: 'Decimal', icon: <Hash className="h-4 w-4" />, description: 'Numbers with decimals' },
    { value: 'boolean', label: 'True/False', icon: <ToggleLeft className="h-4 w-4" />, description: 'Yes/No values' },
    { value: 'date', label: 'Date', icon: <Calendar className="h-4 w-4" />, description: 'Date only' },
    { value: 'datetime', label: 'Date & Time', icon: <Calendar className="h-4 w-4" />, description: 'Date and time' },
    { value: 'email', label: 'Email', icon: <Mail className="h-4 w-4" />, description: 'Email addresses' },
    { value: 'phone', label: 'Phone', icon: <Phone className="h-4 w-4" />, description: 'Phone numbers' },
    { value: 'url', label: 'Website', icon: <Link className="h-4 w-4" />, description: 'Web URLs' },
    { value: 'image', label: 'Image', icon: <Image className="h-4 w-4" />, description: 'Image files' },
    { value: 'file', label: 'File', icon: <FileText className="h-4 w-4" />, description: 'Any file type' },
  ]

  const addField = () => {
    const newField: SchemaField = {
      name: '',
      displayName: '',
      type: 'string',
      required: false,
      unique: false,
      ui: { hidden: false, readonly: false }
    }
    
    setSchema(prev => ({
      ...prev,
      fields: [...(prev.fields || []), newField]
    }))
  }

  const updateField = (index: number, updates: Partial<SchemaField>) => {
    setSchema(prev => ({
      ...prev,
      fields: prev.fields?.map((field, i) => 
        i === index ? { ...field, ...updates } : field
      ) || []
    }))
  }

  const removeField = (index: number) => {
    setSchema(prev => ({
      ...prev,
      fields: prev.fields?.filter((_, i) => i !== index) || []
    }))
  }

  const handleSave = () => {
    // Basic validation
    if (!schema.name || !schema.displayName) {
      alert('Please fill in the schema name and display name')
      return
    }

    const completeSchema: EntitySchema = {
      id: schemaId || undefined,
      name: schema.name,
      displayName: schema.displayName,
      description: schema.description || '',
      fields: schema.fields || [],
      relationships: schema.relationships || [],
      permissions: schema.permissions || [],
      functions: schema.functions || [],
      metadata: {
        tenantId: 'current-tenant', // This would come from context
        version: '1.0.0',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        author: 'current-user' // This would come from context
      }
    }

    onSave(completeSchema)
  }

  const generatePreviewSQL = () => {
    if (!schema.fields?.length) return ''

    const tableName = schema.name
    const columns = schema.fields.map(field => {
      let sqlType = 'TEXT'
      
      switch (field.type) {
        case 'string':
          sqlType = 'VARCHAR(255)'
          break
        case 'text':
          sqlType = 'TEXT'
          break
        case 'integer':
          sqlType = 'INTEGER'
          break
        case 'decimal':
          sqlType = 'DECIMAL(10,2)'
          break
        case 'boolean':
          sqlType = 'BOOLEAN'
          break
        case 'date':
          sqlType = 'DATE'
          break
        case 'datetime':
          sqlType = 'TIMESTAMP'
          break
        case 'uuid':
          sqlType = 'UUID'
          break
        case 'email':
        case 'phone':
        case 'url':
          sqlType = 'VARCHAR(255)'
          break
        case 'json':
          sqlType = 'JSONB'
          break
      }

      const constraints = []
      if (field.required) constraints.push('NOT NULL')
      if (field.unique) constraints.push('UNIQUE')
      if (field.name === 'id') constraints.push('PRIMARY KEY')

      return `  ${field.name} ${sqlType}${constraints.length ? ' ' + constraints.join(' ') : ''}`
    }).join(',\n')

    return `CREATE TABLE ${tableName} (\n${columns}\n);`
  }

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
            <h1 className="text-2xl font-bold text-gray-900">
              {schemaId ? 'Edit Schema' : 'Create New Schema'}
            </h1>
            <p className="text-sm text-gray-600">
              Design your data structure with our visual builder
            </p>
          </div>
        </div>
        
        <div className="flex items-center space-x-2">
          <Button variant="outline" onClick={() => setActiveTab('preview')}>
            <Eye className="h-4 w-4 mr-2" />
            Preview
          </Button>
          <Button onClick={handleSave}>
            <Save className="h-4 w-4 mr-2" />
            Save Schema
          </Button>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-8">
          {[
            { id: 'basic', name: 'Basic Info', icon: Database },
            { id: 'fields', name: 'Fields', icon: Type },
            { id: 'preview', name: 'Preview', icon: Eye },
          ].map((tab) => {
            const Icon = tab.icon
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as any)}
                className={cn(
                  'flex items-center space-x-2 py-2 px-1 border-b-2 font-medium text-sm',
                  activeTab === tab.id
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                )}
              >
                <Icon className="h-4 w-4" />
                <span>{tab.name}</span>
              </button>
            )
          })}
        </nav>
      </div>

      {/* Tab content */}
      {activeTab === 'basic' && (
        <Card>
          <CardHeader>
            <CardTitle>Schema Information</CardTitle>
            <CardDescription>
              Basic details about your data schema
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <Label htmlFor="displayName">Display Name *</Label>
                <Input
                  id="displayName"
                  placeholder="e.g., Customers"
                  value={schema.displayName || ''}
                  onChange={(e) => setSchema(prev => ({ ...prev, displayName: e.target.value }))}
                />
                <p className="text-xs text-gray-500 mt-1">
                  Human-readable name shown in the interface
                </p>
              </div>
              
              <div>
                <Label htmlFor="name">Technical Name *</Label>
                <Input
                  id="name"
                  placeholder="e.g., customers"
                  value={schema.name || ''}
                  onChange={(e) => setSchema(prev => ({ ...prev, name: e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, '_') }))}
                />
                <p className="text-xs text-gray-500 mt-1">
                  Database table name (lowercase, underscores only)
                </p>
              </div>
            </div>
            
            <div>
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                placeholder="Describe what this schema represents..."
                value={schema.description || ''}
                onChange={(e) => setSchema(prev => ({ ...prev, description: e.target.value }))}
                rows={3}
              />
            </div>
          </CardContent>
        </Card>
      )}

      {activeTab === 'fields' && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-lg font-medium text-gray-900">Fields</h3>
              <p className="text-sm text-gray-600">
                Define the data fields for your schema
              </p>
            </div>
            <Button onClick={addField}>
              <Plus className="h-4 w-4 mr-2" />
              Add Field
            </Button>
          </div>

          <div className="space-y-4">
            {schema.fields?.map((field, index) => (
              <Card key={index}>
                <CardContent className="p-4">
                  <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                    <div>
                      <Label>Display Name</Label>
                      <Input
                        placeholder="e.g., Full Name"
                        value={field.displayName}
                        onChange={(e) => updateField(index, { displayName: e.target.value })}
                      />
                    </div>
                    
                    <div>
                      <Label>Field Name</Label>
                      <Input
                        placeholder="e.g., full_name"
                        value={field.name}
                        onChange={(e) => updateField(index, { 
                          name: e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, '_') 
                        })}
                        disabled={field.name === 'id'}
                      />
                    </div>
                    
                    <div>
                      <Label>Type</Label>
                      <Select
                        value={field.type}
                        onValueChange={(value: FieldType) => updateField(index, { type: value })}
                        disabled={field.name === 'id'}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {fieldTypes.map((type) => (
                            <SelectItem key={type.value} value={type.value}>
                              <div className="flex items-center space-x-2">
                                {type.icon}
                                <span>{type.label}</span>
                              </div>
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    
                    <div className="flex items-center space-x-4">
                      <div className="flex items-center space-x-2">
                        <Checkbox
                          id={`required-${index}`}
                          checked={field.required}
                          onCheckedChange={(checked) => updateField(index, { required: !!checked })}
                          disabled={field.name === 'id'}
                        />
                        <Label htmlFor={`required-${index}`} className="text-sm">Required</Label>
                      </div>
                      
                      {field.name !== 'id' && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => removeField(index)}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      )}

      {activeTab === 'preview' && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Schema Preview</CardTitle>
              <CardDescription>
                Review your schema before saving
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <h4 className="font-medium text-gray-900">Schema: {schema.displayName}</h4>
                  <p className="text-sm text-gray-600">{schema.description}</p>
                </div>
                
                <div>
                  <h5 className="font-medium text-gray-900 mb-2">Fields ({schema.fields?.length || 0})</h5>
                  <div className="space-y-2">
                    {schema.fields?.map((field, index) => (
                      <div key={index} className="flex items-center justify-between p-2 bg-gray-50 rounded">
                        <div className="flex items-center space-x-3">
                          <span className="font-medium">{field.displayName}</span>
                          <span className="text-sm text-gray-500">({field.name})</span>
                          <span className="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">
                            {field.type}
                          </span>
                        </div>
                        <div className="flex items-center space-x-2 text-xs text-gray-500">
                          {field.required && <span className="bg-red-100 text-red-800 px-2 py-1 rounded">Required</span>}
                          {field.unique && <span className="bg-green-100 text-green-800 px-2 py-1 rounded">Unique</span>}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Generated SQL</CardTitle>
              <CardDescription>
                Database table that will be created
              </CardDescription>
            </CardHeader>
            <CardContent>
              <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
                {generatePreviewSQL()}
              </pre>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}
