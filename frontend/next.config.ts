import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Enable standalone output for minimal Docker image
  output: 'standalone',
};

export default nextConfig;
