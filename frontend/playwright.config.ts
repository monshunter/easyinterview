import { defineConfig, devices } from "@playwright/test";

/**
 * Playwright configuration for the D2 follow-up pixel parity gate.
 *
 * Truth source: docs/spec/frontend-shell/plans/003-ui-design-pixel-parity-
 * gate/plan.md §4 (Phase 1.2). Configures two viewport projects (desktop
 * 1440×900 / mobile 390×844), a Node-only static server fixture
 * `scripts/serve-pixel-parity.mjs` that mounts both `frontend/dist/` and
 * `ui-design/`, and outputDir scoped to `.playwright-output/`.
 */
export default defineConfig({
  testDir: "./tests/pixel-parity",
  outputDir: "./.playwright-output",
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  reporter: [["list"]],
  use: {
    baseURL: "http://127.0.0.1:4173",
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },
  projects: [
    {
      name: "desktop",
      use: {
        ...devices["Desktop Chrome"],
        viewport: { width: 1440, height: 900 },
      },
    },
    {
      // Mobile project uses the same chromium engine as `desktop` so the
      // gate is about layout / responsive behaviour at 390×844, not browser
      // engine differences. Keeps the install footprint to a single browser
      // (`pnpm exec playwright install chromium`).
      name: "mobile",
      use: {
        ...devices["Desktop Chrome"],
        viewport: { width: 390, height: 844 },
        deviceScaleFactor: 3,
        isMobile: false,
        hasTouch: true,
      },
    },
  ],
  webServer: {
    command: "node ./scripts/serve-pixel-parity.mjs",
    url: "http://127.0.0.1:4173/health",
    reuseExistingServer: !process.env.CI,
    timeout: 30_000,
    stdout: "pipe",
    stderr: "pipe",
  },
});
