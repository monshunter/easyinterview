import path from "node:path";
import { fileURLToPath } from "node:url";

import { defineConfig, devices } from "@playwright/test";

const configDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(configDir, "..");
const artifactRoot = path.join(
  repoRoot,
  ".test-output",
  "e2e",
  "p0-101-auth-email-code-profile-setup",
);

const frontendOrigin =
  process.env.EI_AUTH_EMAIL_CODE_FRONTEND_ORIGIN ?? "http://127.0.0.1:5173";
const outputDir =
  process.env.EI_PLAYWRIGHT_OUTPUT_DIR ?? path.join(artifactRoot, "playwright");

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
});
