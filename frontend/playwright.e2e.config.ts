import path from "node:path";
import { fileURLToPath } from "node:url";

import { defineConfig, devices } from "@playwright/test";

const configDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(configDir, "..");
const artifactRoot = path.join(
  repoRoot,
  ".test-output",
  "e2e",
  "p0-099-full-funnel-fullstack-ui-journey",
);

const backendPort = process.env.EI_E2E_BACKEND_PORT ?? "18099";
const frontendPort = process.env.EI_E2E_FRONTEND_PORT ?? "4174";
const frontendOrigin =
  process.env.EI_E2E_FRONTEND_ORIGIN ?? `http://127.0.0.1:${frontendPort}`;
const apiBaseUrl = `http://127.0.0.1:${backendPort}/api/v1`;
const statePath = process.env.EI_E2E_STATE_PATH ?? path.join(artifactRoot, "state.json");
const outputDir =
  process.env.EI_PLAYWRIGHT_OUTPUT_DIR ?? path.join(artifactRoot, "playwright");
const databaseUrl =
  process.env.DATABASE_URL ??
  "postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable";

function q(value: string): string {
  return `'${value.replace(/'/g, "'\\''")}'`;
}

export default defineConfig({
  testDir: "./tests/e2e",
  outputDir,
  fullyParallel: false,
  workers: 1,
  forbidOnly: !!process.env.CI,
  retries: 0,
  reporter: [["list"]],
  use: {
    baseURL: frontendOrigin,
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: {
        ...devices["Desktop Chrome"],
        viewport: { width: 1440, height: 900 },
      },
    },
  ],
  webServer: [
    {
      command: [
        "cd ../backend &&",
        `${fullEnv("EI_E2E_P0_099_SERVER", "1")}`,
        `${fullEnv("EI_E2E_BACKEND_PORT", backendPort)}`,
        `${fullEnv("EI_E2E_FRONTEND_ORIGIN", frontendOrigin)}`,
        `${fullEnv("EI_E2E_STATE_PATH", statePath)}`,
        `${fullEnv("DATABASE_URL", databaseUrl)}`,
        "go test -v ./cmd/api -run '^TestE2EP0099ScenarioBackendServer$' -count=1 -timeout=30m",
      ].join(" "),
      url: `http://127.0.0.1:${backendPort}/__e2e/health`,
      reuseExistingServer: false,
      timeout: 120_000,
      stdout: "pipe",
      stderr: "pipe",
    },
    {
      command: [
        `${fullEnv("VITE_EI_API_MODE", "real")}`,
        `${fullEnv("VITE_EI_API_BASE_URL", apiBaseUrl)}`,
        "pnpm build",
        "&&",
        `${fullEnv("VITE_EI_API_MODE", "real")}`,
        `${fullEnv("VITE_EI_API_BASE_URL", apiBaseUrl)}`,
        `pnpm exec vite preview --host 127.0.0.1 --port ${frontendPort}`,
      ].join(" "),
      url: frontendOrigin,
      reuseExistingServer: false,
      timeout: 120_000,
      stdout: "pipe",
      stderr: "pipe",
    },
  ],
});

function fullEnv(key: string, value: string): string {
  return `${key}=${q(value)}`;
}
