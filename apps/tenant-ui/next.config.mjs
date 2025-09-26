/** @type {import('next').NextConfig} */
const nextConfig = {
  // Configure base path for gateway routing
  basePath: '/ui',
  
  // Enable experimental features for better performance
  experimental: {
    optimizePackageImports: ['lucide-react', 'recharts'],
  },
  
  // Configure for multi-tenant deployment
  async rewrites() {
    return [
      // Handle tenant subdomain routing in development (when accessed directly)
      {
        source: '/:path*',
        destination: '/:path*',
        has: [
          {
            type: 'host',
            value: '(?<tenant>.*)\\.localhost:3001',
          },
        ],
      },
    ]
  },
  
  // Configure headers for tenant context
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          {
            key: 'X-Frame-Options',
            value: 'SAMEORIGIN',
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
        ],
      },
    ]
  },
  
  // Ensure static assets work properly
  trailingSlash: false,
  
  // Configure for Docker development
  output: 'standalone',
}

export default nextConfig
