"use client"

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { 
  Building2, 
  Globe, 
  Database, 
  ShoppingCart, 
  Users, 
  FileText,
  ArrowRight,
  Check,
  AlertCircle
} from 'lucide-react'

const INDUSTRY_TEMPLATES = [
  {
    id: 'crm',
    name: 'Customer Relationship Management',
    description: 'Manage contacts, leads, deals, and customer interactions',
    icon: <Users className="h-6 w-6" />,
    features: ['Contact Management', 'Lead Tracking', 'Deal Pipeline', 'Activity History'],
    color: 'bg-blue-500'
  },
  {
    id: 'ecommerce',
    name: 'E-commerce Platform',
    description: 'Build online stores with products, orders, and inventory',
    icon: <ShoppingCart className="h-6 w-6" />,
    features: ['Product Catalog', 'Order Management', 'Inventory Tracking', 'Customer Accounts'],
    color: 'bg-green-500'
  },
  {
    id: 'content',
    name: 'Content Management',
    description: 'Create and manage content, blogs, and documentation',
    icon: <FileText className="h-6 w-6" />,
    features: ['Content Editor', 'Media Library', 'Publishing Workflow', 'SEO Tools'],
    color: 'bg-purple-500'
  },
  {
    id: 'custom',
    name: 'Custom Application',
    description: 'Start with a blank slate and build your own schema',
    icon: <Database className="h-6 w-6" />,
    features: ['Flexible Schema', 'Custom Fields', 'API Access', 'Full Control'],
    color: 'bg-gray-500'
  }
]

export default function CreateTenantPage() {
  const [formData, setFormData] = useState({
    companyName: '',
    tenantSlug: '',
    template: '',
    description: ''
  })
  const [isLoading, setIsLoading] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [slugAvailable, setSlugAvailable] = useState<boolean | null>(null)

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: value
    }))
    
    // Auto-generate slug from company name
    if (name === 'companyName') {
      const slug = value
        .toLowerCase()
        .replace(/[^a-z0-9\s-]/g, '')
        .replace(/\s+/g, '-')
        .replace(/-+/g, '-')
        .trim()
      setFormData(prev => ({ ...prev, tenantSlug: slug }))
      setSlugAvailable(null)
    }

    // Clear error when user starts typing
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: '' }))
    }
  }

  const handleTemplateSelect = (templateId: string) => {
    setFormData(prev => ({ ...prev, template: templateId }))
    if (errors.template) {
      setErrors(prev => ({ ...prev, template: '' }))
    }
  }

  const checkSlugAvailability = async () => {
    if (!formData.tenantSlug) return

    try {
      const token = localStorage.getItem('auth_token')
      const response = await fetch(`/api/platform/tenants/check-slug?slug=${formData.tenantSlug}`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      const data = await response.json()
      setSlugAvailable(data.available)
    } catch (error) {
      console.error('Error checking slug availability:', error)
    }
  }

  const validateForm = () => {
    const newErrors: Record<string, string> = {}

    if (!formData.companyName.trim()) {
      newErrors.companyName = 'Company name is required'
    }
    if (!formData.tenantSlug.trim()) {
      newErrors.tenantSlug = 'URL identifier is required'
    } else if (!/^[a-z0-9-]+$/.test(formData.tenantSlug)) {
      newErrors.tenantSlug = 'URL identifier can only contain lowercase letters, numbers, and hyphens'
    } else if (formData.tenantSlug.length < 3) {
      newErrors.tenantSlug = 'URL identifier must be at least 3 characters'
    }
    if (!formData.template) {
      newErrors.template = 'Please select a template to get started'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!validateForm()) {
      return
    }

    setIsLoading(true)
    
    try {
      const token = localStorage.getItem('auth_token')
      const response = await fetch('/api/platform/tenants', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          name: formData.companyName,
          slug: formData.tenantSlug,
          template: formData.template,
          description: formData.description,
        }),
      })

      if (response.ok) {
        const tenant = await response.json()
        // Redirect to tenant dashboard
        window.location.href = `/ui?tenant=${tenant.slug}`
      } else {
        const error = await response.json()
        setErrors({ submit: error.message || 'Failed to create tenant. Please try again.' })
      }
    } catch (error) {
      setErrors({ submit: 'Network error. Please check your connection and try again.' })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="text-center mb-12">
          <div className="flex items-center justify-center mb-6">
            <Building2 className="h-8 w-8 text-blue-600" />
            <span className="ml-2 text-2xl font-bold text-gray-900">BackSaaS</span>
          </div>
          <h1 className="text-4xl font-bold text-gray-900 mb-4">Create Your First Tenant</h1>
          <p className="text-xl text-gray-600 max-w-2xl mx-auto">
            Set up your organization and choose a template to get started quickly
          </p>
        </div>

        <div className="bg-white rounded-lg shadow-lg p-8">
          <form onSubmit={handleSubmit} className="space-y-8">
            {/* Company Information */}
            <div className="space-y-6">
              <h2 className="text-2xl font-semibold text-gray-900 border-b pb-2">
                Company Information
              </h2>
              
              <div>
                <label htmlFor="companyName" className="block text-sm font-medium text-gray-700 mb-2">
                  Company Name
                </label>
                <input
                  type="text"
                  id="companyName"
                  name="companyName"
                  value={formData.companyName}
                  onChange={handleInputChange}
                  className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
                    errors.companyName ? 'border-red-500' : 'border-gray-300'
                  }`}
                  placeholder="Acme Corporation"
                />
                {errors.companyName && (
                  <p className="mt-1 text-sm text-red-600">{errors.companyName}</p>
                )}
              </div>

              <div>
                <label htmlFor="tenantSlug" className="block text-sm font-medium text-gray-700 mb-2">
                  URL Identifier
                </label>
                <div className="flex items-center">
                  <span className="inline-flex items-center px-3 py-3 border border-r-0 border-gray-300 bg-gray-50 text-gray-500 text-sm rounded-l-lg">
                    <Globe className="h-4 w-4 mr-2" />
                    https://
                  </span>
                  <input
                    type="text"
                    id="tenantSlug"
                    name="tenantSlug"
                    value={formData.tenantSlug}
                    onChange={handleInputChange}
                    onBlur={checkSlugAvailability}
                    className={`flex-1 px-4 py-3 border border-l-0 border-r-0 focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
                      errors.tenantSlug ? 'border-red-500' : 'border-gray-300'
                    }`}
                    placeholder="acme-corp"
                  />
                  <span className="inline-flex items-center px-3 py-3 border border-l-0 border-gray-300 bg-gray-50 text-gray-500 text-sm rounded-r-lg">
                    .backsaas.dev
                  </span>
                </div>
                {slugAvailable === true && (
                  <p className="mt-1 text-sm text-green-600 flex items-center">
                    <Check className="h-4 w-4 mr-1" />
                    Available
                  </p>
                )}
                {slugAvailable === false && (
                  <p className="mt-1 text-sm text-red-600 flex items-center">
                    <AlertCircle className="h-4 w-4 mr-1" />
                    Not available
                  </p>
                )}
                {errors.tenantSlug && (
                  <p className="mt-1 text-sm text-red-600">{errors.tenantSlug}</p>
                )}
              </div>

              <div>
                <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-2">
                  Description (Optional)
                </label>
                <textarea
                  id="description"
                  name="description"
                  value={formData.description}
                  onChange={handleInputChange}
                  rows={3}
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="Brief description of your company or use case..."
                />
              </div>
            </div>

            {/* Template Selection */}
            <div className="space-y-6">
              <h2 className="text-2xl font-semibold text-gray-900 border-b pb-2">
                Choose Your Template
              </h2>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {INDUSTRY_TEMPLATES.map((template) => (
                  <div
                    key={template.id}
                    className={`relative border-2 rounded-lg p-6 cursor-pointer transition-all hover:shadow-md ${
                      formData.template === template.id
                        ? 'border-blue-500 bg-blue-50'
                        : 'border-gray-200 hover:border-gray-300'
                    }`}
                    onClick={() => handleTemplateSelect(template.id)}
                  >
                    <div className="flex items-start space-x-4">
                      <div className={`${template.color} text-white p-3 rounded-lg flex-shrink-0`}>
                        {template.icon}
                      </div>
                      <div className="flex-1">
                        <h3 className="text-lg font-semibold text-gray-900 mb-2">
                          {template.name}
                        </h3>
                        <p className="text-sm text-gray-600 mb-3">
                          {template.description}
                        </p>
                        <ul className="space-y-1">
                          {template.features.map((feature, index) => (
                            <li key={index} className="text-xs text-gray-500 flex items-center">
                              <Check className="h-3 w-3 mr-1 text-green-500" />
                              {feature}
                            </li>
                          ))}
                        </ul>
                      </div>
                    </div>
                    {formData.template === template.id && (
                      <div className="absolute top-4 right-4">
                        <div className="bg-blue-500 text-white rounded-full p-1">
                          <Check className="h-4 w-4" />
                        </div>
                      </div>
                    )}
                  </div>
                ))}
              </div>
              
              {errors.template && (
                <p className="text-sm text-red-600">{errors.template}</p>
              )}
            </div>

            {/* Submit Error */}
            {errors.submit && (
              <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                <p className="text-sm text-red-600">{errors.submit}</p>
              </div>
            )}

            {/* Submit Button */}
            <div className="flex justify-end">
              <Button
                type="submit"
                disabled={isLoading}
                size="lg"
                className="px-8 py-3 text-lg font-medium"
              >
                {isLoading ? (
                  'Creating Tenant...'
                ) : (
                  <>
                    Create Tenant
                    <ArrowRight className="ml-2 h-5 w-5" />
                  </>
                )}
              </Button>
            </div>
          </form>
        </div>

        {/* Next Steps Preview */}
        <div className="mt-8 bg-blue-50 border border-blue-200 rounded-lg p-6">
          <h3 className="text-lg font-semibold text-blue-900 mb-3">What happens next?</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
            <div className="flex items-start space-x-3">
              <div className="bg-blue-500 text-white rounded-full p-1 flex-shrink-0 mt-0.5">
                <span className="block w-4 h-4 text-xs font-bold text-center leading-4">1</span>
              </div>
              <div>
                <p className="font-medium text-blue-900">Tenant Created</p>
                <p className="text-blue-700">Your organization space is set up</p>
              </div>
            </div>
            <div className="flex items-start space-x-3">
              <div className="bg-blue-500 text-white rounded-full p-1 flex-shrink-0 mt-0.5">
                <span className="block w-4 h-4 text-xs font-bold text-center leading-4">2</span>
              </div>
              <div>
                <p className="font-medium text-blue-900">Schema Initialized</p>
                <p className="text-blue-700">Template schema is deployed</p>
              </div>
            </div>
            <div className="flex items-start space-x-3">
              <div className="bg-blue-500 text-white rounded-full p-1 flex-shrink-0 mt-0.5">
                <span className="block w-4 h-4 text-xs font-bold text-center leading-4">3</span>
              </div>
              <div>
                <p className="font-medium text-blue-900">Ready to Use</p>
                <p className="text-blue-700">Access your dashboard and start building</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
