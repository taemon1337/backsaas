import React from 'react'

export const metadata = { title: 'BackSaas Console' }
export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen antialiased bg-background text-foreground">
        <div className="max-w-6xl mx-auto p-6">
          <header className="py-4 border-b mb-6">
            <h1 className="text-2xl font-semibold">BackSaas Console</h1>
          </header>
          {children}
        </div>
      </body>
    </html>
  )
}
