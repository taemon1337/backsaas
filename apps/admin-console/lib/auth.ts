import jwt from 'jsonwebtoken'
import Cookies from 'js-cookie'

export interface AdminUser {
  id: string
  email: string
  role: 'super_admin' | 'platform_admin' | 'support_admin' | 'billing_admin'
  name: string
}

export interface AuthToken {
  user_id: string
  email: string
  name: string
  role: 'super_admin' | 'platform_admin' | 'support_admin' | 'billing_admin'
  exp: number
  iat: number
  sub: string
  iss: string
}

const TOKEN_COOKIE_NAME = 'admin_token'

export class AuthService {
  static setToken(token: string) {
    Cookies.set(TOKEN_COOKIE_NAME, token, {
      expires: 1, // 1 day
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'strict'
    })
  }

  static getToken(): string | null {
    return Cookies.get(TOKEN_COOKIE_NAME) || null
  }

  static removeToken() {
    Cookies.remove(TOKEN_COOKIE_NAME)
  }

  static decodeToken(token: string): AuthToken | null {
    try {
      // In the browser, we just decode without verification
      // Token verification should happen on the server side
      const decoded = jwt.decode(token)
      return decoded as AuthToken
    } catch (error) {
      console.error('Token decode error:', error)
      return null
    }
  }

  static getCurrentUser(): AdminUser | null {
    const token = this.getToken()
    if (!token) return null

    const decoded = this.decodeToken(token)
    if (!decoded) return null

    // Check if token is expired
    if (decoded.exp * 1000 < Date.now()) {
      this.removeToken()
      return null
    }

    const user: AdminUser = {
      id: decoded.user_id,
      email: decoded.email,
      name: decoded.name,
      role: decoded.role
    }
    return user
  }

  static isAuthenticated(): boolean {
    return this.getCurrentUser() !== null
  }

  static hasRole(requiredRole: AdminUser['role']): boolean {
    const user = this.getCurrentUser()
    if (!user) return false

    // Role hierarchy: super_admin > platform_admin > support_admin > billing_admin
    const roleHierarchy = {
      'super_admin': 4,
      'platform_admin': 3,
      'support_admin': 2,
      'billing_admin': 1
    }

    return roleHierarchy[user.role] >= roleHierarchy[requiredRole]
  }

  static async login(email: string, password: string): Promise<{ success: boolean; error?: string }> {
    try {
      const { apiClient } = await import('./api-client')
      const data = await apiClient.login(email, password)
      
      if (data.token) {
        this.setToken(data.token)
        return { success: true }
      } else {
        return { success: false, error: 'Login failed' }
      }
    } catch (error: any) {
      console.error('Login error:', error)
      return { 
        success: false, 
        error: error.message || 'Network error' 
      }
    }
  }

  static logout() {
    this.removeToken()
    // Redirect to login page
    window.location.href = '/login'
  }

  static async refreshToken(): Promise<boolean> {
    try {
      const currentToken = this.getToken()
      if (!currentToken) return false

      const { apiClient } = await import('./api-client')
      const data = await apiClient.refreshToken()

      if (data.token) {
        this.setToken(data.token)
        return true
      } else {
        this.removeToken()
        return false
      }
    } catch (error) {
      console.error('Token refresh error:', error)
      this.removeToken()
      return false
    }
  }
}

// Auto-refresh token before expiry
if (typeof window !== 'undefined') {
  setInterval(async () => {
    const user = AuthService.getCurrentUser()
    if (user) {
      const token = AuthService.getToken()
      if (token) {
        const decoded = AuthService.decodeToken(token)
        if (decoded) {
          // Refresh if token expires in less than 5 minutes
          const timeUntilExpiry = decoded.exp * 1000 - Date.now()
          if (timeUntilExpiry < 5 * 60 * 1000) {
            await AuthService.refreshToken()
          }
        }
      }
    }
  }, 60000) // Check every minute
}
