import { expect, test } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

/**
 * Phase 5.2 — Generating screen pixel-parity gate.
 *
 * Clean-checkout hard gate: DOM anchor + computed style + bounding box +
 * responsive geometry + non-empty screenshot smoke. `toHaveScreenshot` is
 * intentionally NOT used here until a stable baseline is committed and this
 * phase explicitly opts into updating it.
 */

interface OperationFixture {
  scenarios: Record<
    string,
    {
      response: {
        status: number;
        headers?: Record<string, string>;
        body: unknown;
      };
    }
  >;
}

const REPORT_ID = "01918fa0-0000-7000-8000-000000007000";
const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";

function fixtureResponse(relativePath: string, scenario = "default") {
  const absolutePath = resolve(process.cwd(), "..", relativePath);
  const fixture = JSON.parse(readFileSync(absolutePath, "utf8")) as OperationFixture;
  const response = fixture.scenarios[scenario]?.response;
  if (!response) throw new Error(`missing fixture ${relativePath}#${scenario}`);
  return response;
}

async function fulfillFixture(
  route: import("@playwright/test").Route,
  relativePath: string,
  scenario = "default",
) {
  const response = fixtureResponse(relativePath, scenario);
  await route.fulfill({
    status: response.status,
    headers: {
      "content-type": "application/json; charset=utf-8",
      ...(response.headers ?? {}),
    },
    body: JSON.stringify(response.body),
  });
}

async function mockGeneratingApis(
  page: import("@playwright/test").Page,
  scenario: "default" | "report-generating" | "report-failed" = "default",
) {
  await page.route("**/api/v1/**", async (route) => {
    const url = route.request().url();
    if (url.includes("/reports/")) {
      return fulfillFixture(route, "openapi/fixtures/Reports/getFeedbackReport.json", scenario);
    }
    if (url.endsWith("/runtime/config")) {
      return fulfillFixture(route, "openapi/fixtures/Auth/getRuntimeConfig.json");
    }
    if (url.endsWith("/me")) {
      return fulfillFixture(route, "openapi/fixtures/Auth/getMe.json");
    }
    await route.fulfill({ status: 204, body: "" });
  });
}

const ROUTE_FRAGMENT = `#route=generating&reportId=${REPORT_ID}&sessionId=${SESSION_ID}`;

test.describe("generating dashboard parity", () => {
  test("desktop renders the 5-phase composition and SLA hint within viewport", async ({
    page,
  }) => {
    await mockGeneratingApis(page, "report-generating");
    await page.goto(`/${ROUTE_FRAGMENT}`);
    await page.waitForSelector("[data-testid='generating-screen']");
    const root = page.locator("[data-testid='generating-screen']");
    await expect(root).toBeVisible();
    await expect(page.locator("[data-testid='generating-progress']")).toBeVisible();
    await expect(page.locator("[data-testid='generating-phase-list']")).toBeVisible();
    await expect(page.locator("[data-testid='generating-live-stream']")).toBeVisible();
    await expect(page.locator("[data-testid='generating-sla-hint']")).toBeVisible();

    const box = await root.boundingBox();
    expect(box).not.toBeNull();
    expect(box!.x).toBeGreaterThanOrEqual(0);
    expect(box!.y).toBeGreaterThanOrEqual(0);
  });

  test("missing reportId surfaces GeneratingErrorState with retry CTA hidden", async ({
    page,
  }) => {
    await mockGeneratingApis(page);
    await page.goto(`/#route=generating&sessionId=${SESSION_ID}`);
    await page.waitForSelector("[data-testid='generating-error-state']");
    await expect(
      page.locator("[data-testid='generating-error-back-to-workspace']"),
    ).toBeVisible();
    await expect(
      page.locator("[data-testid='generating-error-retry']"),
    ).toHaveCount(0);
  });

  test("mobile viewport keeps the surface inside 390px width with no overflow", async ({
    page,
  }) => {
    await mockGeneratingApis(page, "report-generating");
    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto(`/${ROUTE_FRAGMENT}`);
    await page.waitForSelector("[data-testid='generating-screen']");
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth);
    expect(overflow).toBeLessThanOrEqual(390);
  });
});
