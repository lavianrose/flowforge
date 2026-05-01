import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Enable standalone output for minimal Docker image
  output: 'standalone',

  // Disable telemetry
  experimental: {
    instrumentationHook: false,
  },
};

export default nextConfig;
