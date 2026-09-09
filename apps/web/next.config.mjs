/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: 'standalone',
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'images.unsplash.com',
      },
      {
        protocol: 'https',
        hostname: 'res.cloudinary.com',
      },
    ],
  },
  // Proxies the admin API same-origin so the HttpOnly sun_admin_token
  // cookie (host-only, port-agnostic) keeps working in dev and prod without
  // direct cross-origin credentialed requests (decisions.md A5).
  async rewrites() {
    return [
      {
        source: '/api/v1/:path*',
        destination: `${process.env.API_PROXY_TARGET ?? 'http://localhost:8080'}/api/v1/:path*`,
      },
    ];
  },
};

export default nextConfig;
