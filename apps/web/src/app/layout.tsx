import React from 'react'
import './globals.css'
import { Providers } from '@/lib/providers'

export const metadata = { 
  title: 'BackSaaS Control Plane',
  description: 'Schema management and control plane for BackSaaS platform'
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen antialiased bg-background text-foreground">
        <Providers>
          {children}
        </Providers>
      </body>
    </html>
  )
}
