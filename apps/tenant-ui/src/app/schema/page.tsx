"use client"

import React from 'react'
import { DashboardLayout } from '@/components/layout/dashboard-layout'
import { SchemaManager } from '@/components/schema/schema-manager'

export default function SchemaPage() {
  return (
    <DashboardLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Schema Designer</h1>
          <p className="mt-1 text-sm text-gray-600">
            Design and manage your data schemas with our intuitive visual editor.
          </p>
        </div>
        
        <SchemaManager />
      </div>
    </DashboardLayout>
  )
}
