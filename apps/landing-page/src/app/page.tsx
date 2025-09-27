"use client"

import { Button } from '@/components/ui/button'
import { 
  Building2, 
  Database, 
  Workflow, 
  BarChart3, 
  Shield, 
  Zap, 
  Users, 
  Settings,
  ArrowRight,
  CheckCircle
} from 'lucide-react'

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <div className="flex items-center">
              <Building2 className="h-8 w-8 text-blue-600" />
              <span className="ml-2 text-2xl font-bold text-gray-900">BackSaaS</span>
            </div>
            <nav className="hidden md:flex space-x-8">
              <a href="#features" className="text-gray-600 hover:text-gray-900">Features</a>
              <a href="#pricing" className="text-gray-600 hover:text-gray-900">Pricing</a>
              <a href="#docs" className="text-gray-600 hover:text-gray-900">Docs</a>
            </nav>
            <div className="flex items-center space-x-4">
              <Button variant="outline" onClick={() => window.location.href = '/admin'}>
                Admin Console
              </Button>
              <Button onClick={() => window.location.href = '/register'}>
                Get Started
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h1 className="text-5xl font-bold text-gray-900 mb-6">
            Build Multi-Tenant SaaS
            <span className="text-blue-600"> Applications</span>
          </h1>
          <p className="text-xl text-gray-600 mb-8 max-w-3xl mx-auto">
            BackSaaS provides a complete platform for building, deploying, and managing 
            multi-tenant SaaS applications with built-in schema management, workflows, 
            and analytics.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Button size="lg" className="text-lg px-8 py-4" onClick={() => window.location.href = '/register'}>
              Create Your Account
              <ArrowRight className="ml-2 h-5 w-5" />
            </Button>
            <Button variant="outline" size="lg" className="text-lg px-8 py-4">
              View Demo
            </Button>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section id="features" className="py-20 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-bold text-gray-900 mb-4">
              Everything You Need to Build SaaS
            </h2>
            <p className="text-lg text-gray-600 max-w-2xl mx-auto">
              From tenant management to advanced analytics, BackSaaS provides all the tools 
              you need to build and scale your multi-tenant application.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            <FeatureCard
              icon={<Database className="h-8 w-8 text-blue-600" />}
              title="Dynamic Schema Management"
              description="Create and modify database schemas on-the-fly with our visual schema builder and migration wizard."
            />
            <FeatureCard
              icon={<Workflow className="h-8 w-8 text-green-600" />}
              title="Workflow Automation"
              description="Build complex business processes with our drag-and-drop workflow designer and automation engine."
            />
            <FeatureCard
              icon={<BarChart3 className="h-8 w-8 text-purple-600" />}
              title="Advanced Analytics"
              description="Get insights into your data with customizable dashboards, reports, and real-time analytics."
            />
            <FeatureCard
              icon={<Shield className="h-8 w-8 text-red-600" />}
              title="Enterprise Security"
              description="Built-in authentication, authorization, and data isolation for enterprise-grade security."
            />
            <FeatureCard
              icon={<Users className="h-8 w-8 text-indigo-600" />}
              title="Team Management"
              description="Manage users, roles, and permissions across multiple tenants with fine-grained access control."
            />
            <FeatureCard
              icon={<Zap className="h-8 w-8 text-yellow-600" />}
              title="API-First Design"
              description="RESTful APIs and GraphQL endpoints for seamless integration with your existing systems."
            />
          </div>
        </div>
      </section>

      {/* Quick Access Section */}
      <section className="py-20 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-bold text-gray-900 mb-4">
              Quick Access
            </h2>
            <p className="text-lg text-gray-600">
              Jump right into the platform with these quick access links
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <QuickAccessCard
              title="Tenant Dashboard"
              description="Access your business dashboard and manage your data"
              href="/ui"
              icon={<Building2 className="h-6 w-6" />}
              color="blue"
            />
            <QuickAccessCard
              title="Admin Console"
              description="Platform administration and system management"
              href="/admin"
              icon={<Settings className="h-6 w-6" />}
              color="green"
            />
            <QuickAccessCard
              title="Control Plane"
              description="Schema management and database operations"
              href="/control-plane"
              icon={<Database className="h-6 w-6" />}
              color="purple"
            />
            <QuickAccessCard
              title="System Health"
              description="Monitor system status and performance metrics"
              href="/dashboard"
              icon={<BarChart3 className="h-6 w-6" />}
              color="red"
            />
          </div>
        </div>
      </section>

      {/* Getting Started Section */}
      <section className="py-20 bg-blue-600">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h2 className="text-3xl font-bold text-white mb-8">
            Ready to Get Started?
          </h2>
          <p className="text-xl text-blue-100 mb-8 max-w-2xl mx-auto">
            Create your first tenant and start building your SaaS application today.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Button 
              size="lg" 
              variant="secondary" 
              className="text-lg px-8 py-4"
              onClick={() => window.location.href = '/register'}
            >
              Get Started
            </Button>
            <Button 
              size="lg" 
              variant="outline" 
              className="text-lg px-8 py-4 text-white border-white hover:bg-white hover:text-blue-600"
              onClick={() => window.location.href = '/admin'}
            >
              Admin Access
            </Button>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gray-900 text-white py-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
            <div>
              <div className="flex items-center mb-4">
                <Building2 className="h-6 w-6 text-blue-400" />
                <span className="ml-2 text-xl font-bold">BackSaaS</span>
              </div>
              <p className="text-gray-400">
                The complete platform for building multi-tenant SaaS applications.
              </p>
            </div>
            <div>
              <h3 className="font-semibold mb-4">Platform</h3>
              <ul className="space-y-2 text-gray-400">
                <li><a href="/ui" className="hover:text-white">Tenant Dashboard</a></li>
                <li><a href="/admin" className="hover:text-white">Admin Console</a></li>
                <li><a href="/control-plane" className="hover:text-white">Control Plane</a></li>
                <li><a href="/dashboard" className="hover:text-white">System Health</a></li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold mb-4">Resources</h3>
              <ul className="space-y-2 text-gray-400">
                <li><a href="#" className="hover:text-white">Documentation</a></li>
                <li><a href="#" className="hover:text-white">API Reference</a></li>
                <li><a href="#" className="hover:text-white">Tutorials</a></li>
                <li><a href="#" className="hover:text-white">Support</a></li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold mb-4">Company</h3>
              <ul className="space-y-2 text-gray-400">
                <li><a href="#" className="hover:text-white">About</a></li>
                <li><a href="#" className="hover:text-white">Blog</a></li>
                <li><a href="#" className="hover:text-white">Careers</a></li>
                <li><a href="#" className="hover:text-white">Contact</a></li>
              </ul>
            </div>
          </div>
          <div className="border-t border-gray-800 mt-8 pt-8 text-center text-gray-400">
            <p>&copy; 2024 BackSaaS. All rights reserved.</p>
          </div>
        </div>
      </footer>
    </div>
  )
}

function FeatureCard({ icon, title, description }: {
  icon: React.ReactNode
  title: string
  description: string
}) {
  return (
    <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200 hover:shadow-md transition-shadow">
      <div className="mb-4">{icon}</div>
      <h3 className="text-xl font-semibold text-gray-900 mb-2">{title}</h3>
      <p className="text-gray-600">{description}</p>
    </div>
  )
}

function QuickAccessCard({ title, description, href, icon, color }: {
  title: string
  description: string
  href: string
  icon: React.ReactNode
  color: 'blue' | 'green' | 'purple' | 'red'
}) {
  const colorClasses = {
    blue: 'bg-blue-500 hover:bg-blue-600',
    green: 'bg-green-500 hover:bg-green-600',
    purple: 'bg-purple-500 hover:bg-purple-600',
    red: 'bg-red-500 hover:bg-red-600',
  }

  return (
    <div 
      className="bg-white p-6 rounded-lg shadow-sm border border-gray-200 hover:shadow-md transition-all cursor-pointer group"
      onClick={() => window.location.href = href}
    >
      <div className={`inline-flex p-3 rounded-lg text-white mb-4 ${colorClasses[color]} group-hover:scale-110 transition-transform`}>
        {icon}
      </div>
      <h3 className="text-lg font-semibold text-gray-900 mb-2">{title}</h3>
      <p className="text-gray-600 text-sm">{description}</p>
      <div className="mt-4 flex items-center text-sm font-medium text-blue-600 group-hover:text-blue-700">
        Access Now
        <ArrowRight className="ml-1 h-4 w-4 group-hover:translate-x-1 transition-transform" />
      </div>
    </div>
  )
}
