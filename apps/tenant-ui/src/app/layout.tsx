import React from 'react'
import './globals.css'
import { TenantProvider } from '@/lib/tenant-context'
import { Providers } from '@/lib/providers'

export const metadata = {
  title: 'BackSaaS - Tenant Dashboard',
  description: 'Manage your business data and workflows',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className="min-h-screen antialiased bg-background text-foreground">
        <Providers>
          <TenantProvider>
            {children}
          </TenantProvider>
        </Providers>
      </body>
    </html>
  )
}
