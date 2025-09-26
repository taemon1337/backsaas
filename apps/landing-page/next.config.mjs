/** @type {import('next').NextConfig} */
const nextConfig = {
  // No basePath needed - this serves the root
  
  // Enable experimental features for better performance
  experimental: {
    optimizePackageImports: ['lucide-react'],
  },

  // Configure headers for security
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
