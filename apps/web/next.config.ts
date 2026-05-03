import path from "node:path"
import type { NextConfig } from "next"

const nextConfig: NextConfig = {
  turbopack: {
    root: path.join(__dirname, "../.."),
  },
  async rewrites() {
    const apiBaseURL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080"

    return [
      {
        source: "/api/:path*",
        destination: `${apiBaseURL}/api/:path*`,
      },
      {
        source: "/health",
        destination: `${apiBaseURL}/health`,
      },
    ]
  },
}

export default nextConfig
