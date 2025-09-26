import React from 'react'
import './globals.css'

export const metadata = {
  title: 'BackSaaS - Multi-Tenant Platform',
  description: 'Build and manage your SaaS applications with ease',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className="min-h-screen antialiased bg-background text-foreground">
        {children}
      </body>
    </html>
  )
}
