// Core tenant types
export interface Tenant {
  id: string
  name: string
  slug: string
  domain?: string
  branding: TenantBranding
  settings: TenantSettings
  subscription: TenantSubscription
  createdAt: string
  updatedAt: string
}

export interface TenantBranding {
  logo?: string
  favicon?: string
  colors: {
    primary: string
    secondary: string
    accent: string
    background: string
    surface: string
    text: string
  }
  fonts: {
    heading: string
    body: string
  }
  customCSS?: string
}

export interface TenantSettings {
  timezone: string
  dateFormat: string
  currency: string
  language: string
  features: string[]
  integrations: Record<string, any>
}

export interface TenantSubscription {
  plan: string
  status: 'active' | 'inactive' | 'trial' | 'suspended'
  expiresAt?: string
  limits: {
    users: number
    storage: number
    apiCalls: number
  }
}

// User and authentication types
export interface TenantUser {
  id: string
  tenantId: string
  email: string
  name: string
  avatar?: string
  roles: string[]
  permissions: string[]
  entityAccess: Record<string, 'read' | 'write' | 'admin'>
  isActive: boolean
  lastLoginAt?: string
  createdAt: string
  updatedAt: string
}

export interface AuthSession {
  user: TenantUser
  tenant: Tenant
  token: string
  expiresAt: string
}

// Schema and entity types (from control plane)
export interface EntitySchema {
  id?: string
  name: string
  displayName: string
  description?: string
  fields: SchemaField[]
  relationships: SchemaRelationship[]
  permissions: SchemaPermission[]
  functions: SchemaFunction[]
  metadata: {
    tenantId: string
    version: string
    createdAt: string
    updatedAt: string
    author: string
  }
}

export interface SchemaField {
  name: string
  displayName: string
  type: FieldType
  required: boolean
  unique: boolean
  default?: any
  validation?: FieldValidation
  ui?: FieldUIConfig
}

export type FieldType = 
  | 'string' 
  | 'text' 
  | 'integer' 
  | 'decimal' 
  | 'boolean' 
  | 'date' 
  | 'datetime' 
  | 'time'
  | 'email' 
  | 'phone' 
  | 'url' 
  | 'uuid' 
  | 'json' 
  | 'file' 
  | 'image'
  | 'select'
  | 'multiselect'

export interface FieldValidation {
  min?: number
  max?: number
  pattern?: string
  options?: string[]
  custom?: string
}

export interface FieldUIConfig {
  widget?: 'input' | 'textarea' | 'select' | 'checkbox' | 'radio' | 'file' | 'date' | 'datetime'
  placeholder?: string
  helpText?: string
  hidden?: boolean
  readonly?: boolean
  group?: string
}

export interface SchemaRelationship {
  name: string
  type: 'one-to-one' | 'one-to-many' | 'many-to-many'
  target: string
  foreignKey?: string
  displayField?: string
}

export interface SchemaPermission {
  role: string
  actions: ('create' | 'read' | 'update' | 'delete')[]
  conditions?: Record<string, any>
}

export interface SchemaFunction {
  name: string
  type: 'validation' | 'transformation' | 'hook'
  trigger: 'before_create' | 'after_create' | 'before_update' | 'after_update' | 'before_delete' | 'after_delete'
  code: string
}

// Data management types
export interface EntityRecord {
  id: string
  [key: string]: any
  _metadata: {
    createdAt: string
    updatedAt: string
    createdBy: string
    updatedBy: string
    version: number
  }
}

export interface QueryOptions {
  page?: number
  limit?: number
  sort?: string
  order?: 'asc' | 'desc'
  filters?: Record<string, any>
  include?: string[]
}

export interface QueryResult<T = EntityRecord> {
  data: T[]
  pagination: {
    page: number
    limit: number
    total: number
    totalPages: number
  }
  meta?: Record<string, any>
}

// Dashboard and widget types
export interface DashboardConfig {
  id: string
  tenantId: string
  name: string
  layout: 'grid' | 'masonry' | 'flex'
  widgets: DashboardWidget[]
  theme: string
  isDefault: boolean
}

export interface DashboardWidget {
  id: string
  type: 'chart' | 'table' | 'metric' | 'list' | 'custom'
  title: string
  entity?: string
  query?: QueryConfig
  chartConfig?: ChartConfig
  size: 'small' | 'medium' | 'large' | 'full'
  position: { x: number; y: number; w: number; h: number }
  refreshInterval?: number
}

export interface QueryConfig {
  entity: string
  filters?: Record<string, any>
  aggregations?: Record<string, any>
  groupBy?: string[]
  orderBy?: string
  limit?: number
}

export interface ChartConfig {
  type: 'line' | 'bar' | 'pie' | 'area' | 'scatter'
  xAxis: string
  yAxis: string | string[]
  colors?: string[]
  options?: Record<string, any>
}

// Workflow types
export interface Workflow {
  id: string
  tenantId: string
  name: string
  description?: string
  entity: string
  trigger: WorkflowTrigger
  steps: WorkflowStep[]
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface WorkflowTrigger {
  type: 'manual' | 'automatic' | 'scheduled' | 'webhook'
  conditions?: Record<string, any>
  schedule?: string
}

export interface WorkflowStep {
  id: string
  type: 'action' | 'condition' | 'approval' | 'notification'
  name: string
  config: Record<string, any>
  nextSteps: string[]
}

// API response types
export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
    details?: any
  }
  meta?: Record<string, any>
}

export interface ApiError {
  code: string
  message: string
  details?: any
  status?: number
}

// Form types
export interface FormField {
  name: string
  label: string
  type: FieldType
  required?: boolean
  validation?: FieldValidation
  ui?: FieldUIConfig
  value?: any
  error?: string
}

export interface FormConfig {
  entity: string
  fields: FormField[]
  layout?: 'single' | 'two-column' | 'tabs'
  submitText?: string
  cancelText?: string
  showReset?: boolean
}

// Navigation types
export interface NavItem {
  name: string
  href: string
  icon?: string
  badge?: string | number
  children?: NavItem[]
  permissions?: string[]
}

export interface BreadcrumbItem {
  name: string
  href?: string
  current?: boolean
}
