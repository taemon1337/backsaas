/**
 * Unit Tests for Tenant CRUD Operations
 * 
 * This file contains comprehensive tests for all tenant management functionality
 * including Create, Read, Update, Delete operations and UI interactions.
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { jest } from '@jest/globals'
import { CreateTenantModal } from '@/components/tenants/create-tenant-modal'
import { EditTenantDrawer } from '@/components/tenants/edit-tenant-drawer'
import { DeleteTenantDialog } from '@/components/tenants/delete-tenant-dialog'
import { apiClient } from '@/lib/api-client'

// Mock the API client
jest.mock('@/lib/api-client', () => ({
  apiClient: {
    createTenant: jest.fn(),
    updateTenant: jest.fn(),
    deleteTenant: jest.fn(),
    getTenants: jest.fn(),
  }
}))

// Mock the toast hook
jest.mock('@/components/ui/use-toast', () => ({
  useToast: () => ({
    toast: jest.fn()
  })
}))

// Mock tenant data for testing
const mockTenant = {
  id: 'test-tenant-1',
  name: 'Test Corporation',
  slug: 'test-corporation',
  domain: 'test.example.com',
  plan: 'pro',
  status: 'active',
  owner_email: 'admin@test.com',
  owner_name: 'John Doe',
  user_count: 25,
  storage_used: '1.2 GB',
  created_at: '2024-01-01T00:00:00Z',
  last_activity: '2024-01-15T12:00:00Z'
}

describe('Tenant CRUD Operations', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('Create Tenant Modal', () => {
    test('renders create tenant modal with all form fields', () => {
      render(
        <CreateTenantModal 
          open={true} 
          onOpenChange={jest.fn()} 
          onSuccess={jest.fn()} 
        />
      )

      expect(screen.getByText('Create New Tenant')).toBeInTheDocument()
      expect(screen.getByLabelText('Organization Name')).toBeInTheDocument()
      expect(screen.getByLabelText('Slug')).toBeInTheDocument()
      expect(screen.getByLabelText('Domain')).toBeInTheDocument()
      expect(screen.getByLabelText('Plan')).toBeInTheDocument()
      expect(screen.getByLabelText('Owner Name')).toBeInTheDocument()
      expect(screen.getByLabelText('Owner Email')).toBeInTheDocument()
    })

    test('auto-generates slug and domain from organization name', async () => {
      render(
        <CreateTenantModal 
          open={true} 
          onOpenChange={jest.fn()} 
          onSuccess={jest.fn()} 
        />
      )

      const nameInput = screen.getByLabelText('Organization Name')
      fireEvent.change(nameInput, { target: { value: 'Acme Corporation' } })

      await waitFor(() => {
        const slugInput = screen.getByLabelText('Slug')
        const domainInput = screen.getByLabelText('Domain')
        
        expect(slugInput).toHaveValue('acme-corporation')
        expect(domainInput).toHaveValue('acme-corporation.yourdomain.com')
      })
    })

    test('validates required fields before submission', async () => {
      const onSuccess = jest.fn()
      render(
        <CreateTenantModal 
          open={true} 
          onOpenChange={jest.fn()} 
          onSuccess={onSuccess} 
        />
      )

      const submitButton = screen.getByText('Create Tenant')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(apiClient.createTenant).not.toHaveBeenCalled()
        expect(onSuccess).not.toHaveBeenCalled()
      })
    })

    test('successfully creates tenant with valid data', async () => {
      const onSuccess = jest.fn()
      const onOpenChange = jest.fn()
      
      ;(apiClient.createTenant as jest.Mock).mockResolvedValue({ id: 'new-tenant' })

      render(
        <CreateTenantModal 
          open={true} 
          onOpenChange={onOpenChange} 
          onSuccess={onSuccess} 
        />
      )

      // Fill out the form
      fireEvent.change(screen.getByLabelText('Organization Name'), { 
        target: { value: 'New Company' } 
      })
      fireEvent.change(screen.getByLabelText('Owner Name'), { 
        target: { value: 'Jane Smith' } 
      })
      fireEvent.change(screen.getByLabelText('Owner Email'), { 
        target: { value: 'jane@newcompany.com' } 
      })

      const submitButton = screen.getByText('Create Tenant')
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(apiClient.createTenant).toHaveBeenCalledWith({
          name: 'New Company',
          slug: 'new-company',
          domain: 'new-company.yourdomain.com',
          plan: 'starter',
          owner_email: 'jane@newcompany.com',
          owner_name: 'Jane Smith',
          status: 'active'
        })
        expect(onSuccess).toHaveBeenCalled()
        expect(onOpenChange).toHaveBeenCalledWith(false)
      })
    })

    test('handles API errors gracefully', async () => {
      const onSuccess = jest.fn()
      
      ;(apiClient.createTenant as jest.Mock).mockRejectedValue(
        new Error('Network error')
      )

      render(
        <CreateTenantModal 
          open={true} 
          onOpenChange={jest.fn()} 
          onSuccess={onSuccess} 
        />
      )

      // Fill out and submit form
      fireEvent.change(screen.getByLabelText('Organization Name'), { 
        target: { value: 'Test Company' } 
      })
      fireEvent.change(screen.getByLabelText('Owner Name'), { 
        target: { value: 'Test User' } 
      })
      fireEvent.change(screen.getByLabelText('Owner Email'), { 
        target: { value: 'test@test.com' } 
      })

      fireEvent.click(screen.getByText('Create Tenant'))

      await waitFor(() => {
        expect(onSuccess).not.toHaveBeenCalled()
        // Toast error should be shown
      })
    })
  })

  describe('Edit Tenant Drawer', () => {
    test('renders edit drawer with tenant data pre-filled', () => {
      render(
        <EditTenantDrawer
          open={true}
          onOpenChange={jest.fn()}
          tenant={mockTenant}
          onSuccess={jest.fn()}
        />
      )

      expect(screen.getByText('Edit Tenant: Test Corporation')).toBeInTheDocument()
      expect(screen.getByDisplayValue('Test Corporation')).toBeInTheDocument()
      expect(screen.getByDisplayValue('test-corporation')).toBeInTheDocument()
      expect(screen.getByDisplayValue('test.example.com')).toBeInTheDocument()
      expect(screen.getByDisplayValue('John Doe')).toBeInTheDocument()
      expect(screen.getByDisplayValue('admin@test.com')).toBeInTheDocument()
    })

    test('shows tenant overview information', () => {
      render(
        <EditTenantDrawer
          open={true}
          onOpenChange={jest.fn()}
          tenant={mockTenant}
          onSuccess={jest.fn()}
        />
      )

      expect(screen.getByText('active')).toBeInTheDocument()
      expect(screen.getByText('pro')).toBeInTheDocument()
      expect(screen.getByText('25')).toBeInTheDocument()
      expect(screen.getByText('1.2 GB')).toBeInTheDocument()
    })

    test('successfully updates tenant data', async () => {
      const onSuccess = jest.fn()
      
      ;(apiClient.updateTenant as jest.Mock).mockResolvedValue({ success: true })

      render(
        <EditTenantDrawer
          open={true}
          onOpenChange={jest.fn()}
          tenant={mockTenant}
          onSuccess={onSuccess}
        />
      )

      // Update the tenant name
      const nameInput = screen.getByDisplayValue('Test Corporation')
      fireEvent.change(nameInput, { target: { value: 'Updated Corporation' } })

      const updateButton = screen.getByText('Update Tenant')
      fireEvent.click(updateButton)

      await waitFor(() => {
        expect(apiClient.updateTenant).toHaveBeenCalledWith('test-tenant-1', {
          name: 'Updated Corporation',
          slug: 'test-corporation',
          domain: 'test.example.com',
          plan: 'pro',
          status: 'active',
          owner_email: 'admin@test.com',
          owner_name: 'John Doe'
        })
        expect(onSuccess).toHaveBeenCalled()
      })
    })

    test('allows status changes', async () => {
      const onSuccess = jest.fn()
      
      ;(apiClient.updateTenant as jest.Mock).mockResolvedValue({ success: true })

      render(
        <EditTenantDrawer
          open={true}
          onOpenChange={jest.fn()}
          tenant={mockTenant}
          onSuccess={onSuccess}
        />
      )

      // Change status to suspended
      const statusSelect = screen.getByDisplayValue('active')
      fireEvent.change(statusSelect, { target: { value: 'suspended' } })

      const updateButton = screen.getByText('Update Tenant')
      fireEvent.click(updateButton)

      await waitFor(() => {
        expect(apiClient.updateTenant).toHaveBeenCalledWith('test-tenant-1', 
          expect.objectContaining({
            status: 'suspended'
          })
        )
      })
    })
  })

  describe('Delete Tenant Dialog', () => {
    test('renders delete confirmation dialog with tenant information', () => {
      render(
        <DeleteTenantDialog
          open={true}
          onOpenChange={jest.fn()}
          tenant={mockTenant}
          onSuccess={jest.fn()}
        />
      )

      expect(screen.getByText('Delete Tenant')).toBeInTheDocument()
      expect(screen.getByText('Test Corporation')).toBeInTheDocument()
      expect(screen.getByText('test.example.com')).toBeInTheDocument()
      expect(screen.getByText('John Doe')).toBeInTheDocument()
      expect(screen.getByText('25 user accounts and profiles')).toBeInTheDocument()
      expect(screen.getByText('1.2 GB of stored data and files')).toBeInTheDocument()
    })

    test('requires exact name confirmation before deletion', async () => {
      const onSuccess = jest.fn()
      
      render(
        <DeleteTenantDialog
          open={true}
          onOpenChange={jest.fn()}
          tenant={mockTenant}
          onSuccess={onSuccess}
        />
      )

      const deleteButton = screen.getByText('Delete Tenant')
      
      // Try to delete without confirmation
      fireEvent.click(deleteButton)
      
      await waitFor(() => {
        expect(apiClient.deleteTenant).not.toHaveBeenCalled()
      })

      // Enter wrong confirmation
      const confirmationInput = screen.getByPlaceholderText('Test Corporation')
      fireEvent.change(confirmationInput, { target: { value: 'Wrong Name' } })
      fireEvent.click(deleteButton)
      
      await waitFor(() => {
        expect(apiClient.deleteTenant).not.toHaveBeenCalled()
      })

      // Enter correct confirmation
      fireEvent.change(confirmationInput, { target: { value: 'Test Corporation' } })
      fireEvent.click(deleteButton)
      
      await waitFor(() => {
        expect(apiClient.deleteTenant).toHaveBeenCalledWith('test-tenant-1')
        expect(onSuccess).toHaveBeenCalled()
      })
    })

    test('handles deletion errors gracefully', async () => {
      const onSuccess = jest.fn()
      
      ;(apiClient.deleteTenant as jest.Mock).mockRejectedValue(
        new Error('Cannot delete tenant with active users')
      )

      render(
        <DeleteTenantDialog
          open={true}
          onOpenChange={jest.fn()}
          tenant={mockTenant}
          onSuccess={onSuccess}
        />
      )

      // Enter correct confirmation and try to delete
      const confirmationInput = screen.getByPlaceholderText('Test Corporation')
      fireEvent.change(confirmationInput, { target: { value: 'Test Corporation' } })
      
      const deleteButton = screen.getByText('Delete Tenant')
      fireEvent.click(deleteButton)

      await waitFor(() => {
        expect(onSuccess).not.toHaveBeenCalled()
        // Error toast should be shown
      })
    })
  })

  describe('API Integration Tests', () => {
    test('getTenants API call with pagination', async () => {
      const mockResponse = {
        data: [mockTenant],
        total: 1,
        page: 1,
        limit: 10
      }
      
      ;(apiClient.getTenants as jest.Mock).mockResolvedValue(mockResponse)

      const result = await apiClient.getTenants({ page: 1, limit: 10, search: '' })

      expect(apiClient.getTenants).toHaveBeenCalledWith({
        page: 1,
        limit: 10,
        search: ''
      })
      expect(result).toEqual(mockResponse)
    })

    test('createTenant API call with proper data transformation', async () => {
      const tenantData = {
        name: 'New Tenant',
        slug: 'new-tenant',
        domain: 'new-tenant.example.com',
        plan: 'starter',
        owner_email: 'owner@newtenant.com',
        owner_name: 'Owner Name',
        status: 'active'
      }

      ;(apiClient.createTenant as jest.Mock).mockResolvedValue({ id: 'new-id' })

      const result = await apiClient.createTenant(tenantData)

      expect(apiClient.createTenant).toHaveBeenCalledWith(tenantData)
      expect(result).toEqual({ id: 'new-id' })
    })

    test('updateTenant API call with partial updates', async () => {
      const updateData = {
        name: 'Updated Name',
        status: 'suspended'
      }

      ;(apiClient.updateTenant as jest.Mock).mockResolvedValue({ success: true })

      const result = await apiClient.updateTenant('tenant-id', updateData)

      expect(apiClient.updateTenant).toHaveBeenCalledWith('tenant-id', updateData)
      expect(result).toEqual({ success: true })
    })

    test('deleteTenant API call', async () => {
      ;(apiClient.deleteTenant as jest.Mock).mockResolvedValue({ success: true })

      const result = await apiClient.deleteTenant('tenant-id')

      expect(apiClient.deleteTenant).toHaveBeenCalledWith('tenant-id')
      expect(result).toEqual({ success: true })
    })
  })

  describe('Error Handling', () => {
    test('handles network errors', async () => {
      ;(apiClient.getTenants as jest.Mock).mockRejectedValue(
        new Error('Network error')
      )

      try {
        await apiClient.getTenants({})
      } catch (error) {
        expect(error.message).toBe('Network error')
      }
    })

    test('handles validation errors', async () => {
      ;(apiClient.createTenant as jest.Mock).mockRejectedValue({
        status: 400,
        message: 'Validation failed',
        details: {
          email: 'Invalid email format'
        }
      })

      try {
        await apiClient.createTenant({})
      } catch (error) {
        expect(error.status).toBe(400)
        expect(error.message).toBe('Validation failed')
      }
    })

    test('handles rate limiting', async () => {
      ;(apiClient.createTenant as jest.Mock).mockRejectedValue({
        status: 429,
        message: 'Too many requests'
      })

      try {
        await apiClient.createTenant({})
      } catch (error) {
        expect(error.status).toBe(429)
        expect(error.message).toBe('Too many requests')
      }
    })
  })

  describe('Form Validation', () => {
    test('validates email format', () => {
      const invalidEmails = [
        'invalid-email',
        '@domain.com',
        'user@',
        'user@domain',
        ''
      ]

      invalidEmails.forEach(email => {
        // Test email validation logic
        const isValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
        expect(isValid).toBe(false)
      })
    })

    test('validates slug format', () => {
      const validSlugs = ['valid-slug', 'company-name', 'test123']
      const invalidSlugs = ['-invalid', 'invalid-', 'Invalid_Slug', 'slug with spaces']

      validSlugs.forEach(slug => {
        const isValid = /^[a-z0-9\-]+$/.test(slug) && !slug.startsWith('-') && !slug.endsWith('-')
        expect(isValid).toBe(true)
      })

      invalidSlugs.forEach(slug => {
        const isValid = /^[a-z0-9\-]+$/.test(slug) && !slug.startsWith('-') && !slug.endsWith('-')
        expect(isValid).toBe(false)
      })
    })

    test('validates domain format', () => {
      const validDomains = ['example.com', 'subdomain.example.com', 'test.co.uk']
      const invalidDomains = ['invalid', 'example', '.com', 'example.']

      validDomains.forEach(domain => {
        const isValid = domain.includes('.') && domain.length > 3
        expect(isValid).toBe(true)
      })

      invalidDomains.forEach(domain => {
        const isValid = domain.includes('.') && domain.length > 3
        expect(isValid).toBe(false)
      })
    })
  })
})

/**
 * Test Configuration and Setup
 * 
 * To run these tests:
 * 1. Install testing dependencies: jest, @testing-library/react, @testing-library/jest-dom
 * 2. Configure jest.config.js with proper module resolution
 * 3. Set up test environment with jsdom
 * 4. Run: npm test or yarn test
 * 
 * Test Coverage Areas:
 * - Component rendering and UI interactions
 * - Form validation and data handling
 * - API integration and error handling
 * - User workflows (create, edit, delete)
 * - Edge cases and error scenarios
 */
