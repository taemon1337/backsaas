"use client"

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { AuthService } from '@/lib/auth'
import { useToast } from '@/components/ui/use-toast'
import { Loader2, Shield } from 'lucide-react'
import { NoSSR } from '@/components/no-ssr'

export default function LoginPage() {
  const [email, setEmail] = useState('admin@backsaas.dev')
  const [password, setPassword] = useState('admin123')
  const [isLoading, setIsLoading] = useState(false)
  const router = useRouter()
  const { toast } = useToast()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)

    try {
      const result = await AuthService.login(email, password)
      
      if (result.success) {
        toast({
          title: "Login successful",
          description: "Welcome to BackSaaS Admin Console",
        })
        router.push('/dashboard')
      } else {
        toast({
          title: "Login failed",
          description: result.error || "Invalid credentials",
          variant: "destructive",
        })
      }
    } catch (error) {
      toast({
        title: "Login error",
        description: "An unexpected error occurred",
        variant: "destructive",
      })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="flex items-center justify-center mb-4">
            <Shield className="h-12 w-12 text-blue-500" />
          </div>
          <h1 className="text-3xl font-bold text-white mb-2">BackSaaS</h1>
          <p className="text-slate-300">Admin Console</p>
        </div>

        <Card className="border-slate-700 bg-slate-800/50 backdrop-blur">
          <CardHeader className="text-center">
            <CardTitle className="text-white">Sign In</CardTitle>
            <CardDescription className="text-slate-300">
              Access the platform administration interface
            </CardDescription>
          </CardHeader>
          <CardContent>
            <NoSSR fallback={<div className="space-y-4 animate-pulse">
              <div className="h-4 bg-slate-700 rounded w-16"></div>
              <div className="h-10 bg-slate-700 rounded"></div>
              <div className="h-4 bg-slate-700 rounded w-20"></div>
              <div className="h-10 bg-slate-700 rounded"></div>
              <div className="h-10 bg-slate-700 rounded"></div>
            </div>}>
              <form onSubmit={handleSubmit} className="space-y-4" suppressHydrationWarning>
                <div className="space-y-2">
                  <Label htmlFor="email" className="text-slate-200">Email</Label>
                  <Input
                    id="email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="admin@backsaas.dev"
                    required
                    className="bg-slate-700 border-slate-600 text-white placeholder:text-slate-400"
                    suppressHydrationWarning
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="password" className="text-slate-200">Password</Label>
                  <Input
                    id="password"
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Enter your password"
                    required
                    className="bg-slate-700 border-slate-600 text-white placeholder:text-slate-400"
                    suppressHydrationWarning
                  />
                </div>
                <Button 
                  type="submit" 
                  className="w-full bg-blue-600 hover:bg-blue-700" 
                  disabled={isLoading}
                >
                  {isLoading ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Signing in...
                    </>
                  ) : (
                    'Sign In'
                  )}
                </Button>
              </form>
            </NoSSR>

            <div className="mt-6 p-4 bg-slate-700/50 rounded-lg">
              <p className="text-xs text-slate-300 mb-2">Demo Credentials:</p>
              <p className="text-xs text-slate-400">Email: admin@backsaas.dev</p>
              <p className="text-xs text-slate-400">Password: admin123</p>
            </div>
          </CardContent>
        </Card>

        <div className="text-center mt-6">
          <p className="text-xs text-slate-400">
            Secure platform administration • BackSaaS v1.0
          </p>
        </div>
      </div>
    </div>
  )
}
