"use client"

import React from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { 
  Home, 
  Database, 
  Settings, 
  Users, 
  BarChart3, 
  Workflow,
  FileText,
  Layers,
  X
} from 'lucide-react'
import { useTenant } from '@/lib/tenant-context'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import type { NavItem } from '@/lib/types'

interface SidebarProps {
  open: boolean
  onClose: () => void
}

export function Sidebar({ open, onClose }: SidebarProps) {
  const pathname = usePathname()
  const { tenant, user, hasPermission } = useTenant()

  // Define navigation items based on user permissions
  const navigationItems: NavItem[] = [
    {
      name: 'Dashboard',
      href: '/',
      icon: 'Home',
    },
    {
      name: 'Data',
      href: '/data',
      icon: 'Database',
      children: [
        { name: 'All Entities', href: '/data' },
        { name: 'Import/Export', href: '/data/import-export' },
      ],
      permissions: ['data.read'],
    },
    {
      name: 'Schema Designer',
      href: '/schema',
      icon: 'Layers',
      permissions: ['schema.read'],
    },
    {
      name: 'Workflows',
      href: '/workflows',
      icon: 'Workflow',
      permissions: ['workflow.read'],
    },
    {
      name: 'Analytics',
      href: '/analytics',
      icon: 'BarChart3',
      permissions: ['analytics.read'],
    },
    {
      name: 'Reports',
      href: '/reports',
      icon: 'FileText',
      permissions: ['reports.read'],
    },
    {
      name: 'Team',
      href: '/team',
      icon: 'Users',
      permissions: ['users.read'],
    },
    {
      name: 'Settings',
      href: '/settings',
      icon: 'Settings',
    },
  ]

  // Filter navigation items based on permissions
  const filteredNavigation = navigationItems.filter(item => {
    if (!item.permissions) return true
    return item.permissions.some(permission => hasPermission(permission))
  })

  const iconComponents = {
    Home,
    Database,
    Settings,
    Users,
    BarChart3,
    Workflow,
    FileText,
    Layers,
  }

  return (
    <>
      {/* Desktop sidebar */}
      <div className="hidden lg:fixed lg:inset-y-0 lg:z-50 lg:flex lg:w-64 lg:flex-col">
        <div className="flex grow flex-col gap-y-5 overflow-y-auto bg-white border-r border-gray-200 px-6 pb-4">
          {/* Logo and tenant info */}
          <div className="flex h-16 shrink-0 items-center">
            <div className="flex items-center space-x-3">
              {tenant?.branding?.logo ? (
                <img
                  className="h-8 w-auto"
                  src={tenant.branding.logo}
                  alt={`${tenant.name} logo`}
                />
              ) : (
                <div className="h-8 w-8 rounded bg-tenant-primary flex items-center justify-center text-white font-bold text-sm">
                  {tenant?.name?.charAt(0).toUpperCase() || 'T'}
                </div>
              )}
              <div>
                <div className="text-sm font-semibold text-gray-900">
                  {tenant?.name || 'Tenant'}
                </div>
                <div className="text-xs text-gray-500">
                  Dashboard
                </div>
              </div>
            </div>
          </div>

          {/* Navigation */}
          <nav className="flex flex-1 flex-col">
            <ul role="list" className="flex flex-1 flex-col gap-y-7">
              <li>
                <ul role="list" className="-mx-2 space-y-1">
                  {filteredNavigation.map((item) => {
                    const Icon = iconComponents[item.icon as keyof typeof iconComponents]
                    const isActive = pathname === item.href || 
                      (item.children && item.children.some(child => pathname === child.href))

                    return (
                      <li key={item.name}>
                        <Link
                          href={item.href}
                          className={cn(
                            'group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold transition-colors',
                            isActive
                              ? 'bg-tenant-primary text-white'
                              : 'text-gray-700 hover:text-tenant-primary hover:bg-gray-50'
                          )}
                        >
                          {Icon && (
                            <Icon
                              className={cn(
                                'h-5 w-5 shrink-0',
                                isActive ? 'text-white' : 'text-gray-400 group-hover:text-tenant-primary'
                              )}
                            />
                          )}
                          {item.name}
                          {item.badge && (
                            <span className="ml-auto w-5 h-5 text-xs bg-gray-100 text-gray-600 rounded-full flex items-center justify-center">
                              {item.badge}
                            </span>
                          )}
                        </Link>

                        {/* Sub-navigation */}
                        {item.children && isActive && (
                          <ul className="mt-1 px-2 space-y-1">
                            {item.children.map((child) => (
                              <li key={child.name}>
                                <Link
                                  href={child.href}
                                  className={cn(
                                    'block rounded-md py-2 pl-9 pr-2 text-sm leading-6 transition-colors',
                                    pathname === child.href
                                      ? 'text-tenant-primary bg-gray-50'
                                      : 'text-gray-600 hover:text-tenant-primary hover:bg-gray-50'
                                  )}
                                >
                                  {child.name}
                                </Link>
                              </li>
                            ))}
                          </ul>
                        )}
                      </li>
                    )
                  })}
                </ul>
              </li>
            </ul>
          </nav>
        </div>
      </div>

      {/* Mobile sidebar */}
      <div className={cn(
        'relative z-50 lg:hidden',
        open ? 'block' : 'hidden'
      )}>
        <div className="fixed inset-0 flex">
          <div className="relative mr-16 flex w-full max-w-xs flex-1">
            <div className="absolute left-full top-0 flex w-16 justify-center pt-5">
              <Button
                variant="ghost"
                size="icon"
                onClick={onClose}
                className="text-white hover:text-white hover:bg-white/10"
              >
                <X className="h-6 w-6" />
              </Button>
            </div>

            <div className="flex grow flex-col gap-y-5 overflow-y-auto bg-white px-6 pb-4">
              {/* Mobile logo */}
              <div className="flex h-16 shrink-0 items-center">
                <div className="flex items-center space-x-3">
                  {tenant?.branding?.logo ? (
                    <img
                      className="h-8 w-auto"
                      src={tenant.branding.logo}
                      alt={`${tenant.name} logo`}
                    />
                  ) : (
                    <div className="h-8 w-8 rounded bg-tenant-primary flex items-center justify-center text-white font-bold text-sm">
                      {tenant?.name?.charAt(0).toUpperCase() || 'T'}
                    </div>
                  )}
                  <div>
                    <div className="text-sm font-semibold text-gray-900">
                      {tenant?.name || 'Tenant'}
                    </div>
                    <div className="text-xs text-gray-500">
                      Dashboard
                    </div>
                  </div>
                </div>
              </div>

              {/* Mobile navigation */}
              <nav className="flex flex-1 flex-col">
                <ul role="list" className="flex flex-1 flex-col gap-y-7">
                  <li>
                    <ul role="list" className="-mx-2 space-y-1">
                      {filteredNavigation.map((item) => {
                        const Icon = iconComponents[item.icon as keyof typeof iconComponents]
                        const isActive = pathname === item.href

                        return (
                          <li key={item.name}>
                            <Link
                              href={item.href}
                              onClick={onClose}
                              className={cn(
                                'group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold transition-colors',
                                isActive
                                  ? 'bg-tenant-primary text-white'
                                  : 'text-gray-700 hover:text-tenant-primary hover:bg-gray-50'
                              )}
                            >
                              {Icon && (
                                <Icon
                                  className={cn(
                                    'h-5 w-5 shrink-0',
                                    isActive ? 'text-white' : 'text-gray-400 group-hover:text-tenant-primary'
                                  )}
                                />
                              )}
                              {item.name}
                            </Link>
                          </li>
                        )
                      })}
                    </ul>
                  </li>
                </ul>
              </nav>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
