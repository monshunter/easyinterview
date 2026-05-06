import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    // Default to node so file-system / fixture-projection tests keep working;
    // React component tests opt into jsdom via `// @vitest-environment jsdom`.
    environment: "node",
    setupFiles: ["./src/test/setup.ts"],
    css: false,
    include: ["src/**/*.test.ts", "src/**/*.test.tsx"],
  },
});
