/** @type {import('next').NextConfig} */
const nextConfig = {
  basePath: '/admin',
  
  env: {
    PLATFORM_API_URL: process.env.PLATFORM_API_URL || 'http://localhost:8080',
    GATEWAY_API_URL: process.env.GATEWAY_API_URL || 'http://localhost:8000',
    JWT_SECRET: process.env.JWT_SECRET || 'your-jwt-secret-key',
  },
  
  // Enable standalone output for production Docker builds
  output: 'standalone',
};

module.exports = nextConfig;
