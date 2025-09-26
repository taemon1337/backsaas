/** @type {import('next').NextConfig} */
const nextConfig = {
  basePath: '/control-plane',
  assetPrefix: '/control-plane',
  // Ensure static assets work properly
  trailingSlash: false,
  // Configure for Docker development
  output: 'standalone',
}

export default nextConfig
