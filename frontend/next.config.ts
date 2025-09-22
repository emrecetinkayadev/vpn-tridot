import type { NextConfig } from "next";

const config: NextConfig = {
  reactStrictMode: true,
  experimental: {
    serverActions: true,
  },
};

export default config;
