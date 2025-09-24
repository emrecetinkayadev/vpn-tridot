import { defineConfig } from "vitest/config";
import path from "node:path";

export default defineConfig({
  test: {
    include: ["tests/unit/**/*.test.{ts,tsx}", "tests/unit/**/*.spec.{ts,tsx}"],
    exclude: ["tests/e2e/**"],
    environment: "happy-dom",
    setupFiles: ["./tests/setupTests.ts"],
    globals: true,
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "."),
    },
  },
});
