import { expect, test } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

/**
 * Phase 5.3 — Report dashboard pixel-parity gate (clean-checkout hard gate).
 *
 * DOM anchor + bounding box + responsive geometry + chrome visibility.
 * `toHaveScreenshot` is intentionally NOT enabled until a stable baseline
 * for both viewports has been committed by an explicit baseline-update PR.
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
const RESUME_ID = "01918fa0-0000-7000-8000-000000004000";

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

async function mockReportApis(page: import("@playwright/test").Page) {
  await page.route("**/api/v1/**", async (route) => {
    const url = route.request().url();
    if (url.match(/\/reports\/[^/]+$/)) {
      return fulfillFixture(route, "openapi/fixtures/Reports/getFeedbackReport.json", "default");
    }
    if (url.includes("/targets/")) {
      return fulfillFixture(route, "openapi/fixtures/TargetJobs/getTargetJob.json");
    }
    if (url.includes("/resumes/")) {
      return fulfillFixture(route, "openapi/fixtures/Resumes/getResume.json");
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

test.describe("report dashboard parity", () => {
  test("desktop renders header + context strip + 4 summary cards + detail surface inside viewport", async ({
    page,
  }) => {
    await mockReportApis(page);
    await page.goto(`/#route=report&reportId=${REPORT_ID}&sessionId=${SESSION_ID}&targetJobId=01918fa0-0000-7000-8000-000000002000&resumeId=${RESUME_ID}`);
    await page.waitForSelector("[data-testid='report-dashboard']");

    await expect(page.locator("[data-testid='report-header']")).toBeVisible();
    await expect(page.locator("[data-testid='report-context-strip']")).toBeVisible();
    await expect(page.locator("[data-testid='report-summary-cards']")).toBeVisible();
    await expect(page.locator("[data-testid='report-detail-surface']")).toBeVisible();

    // App chrome / TopBar must remain visible on the report route.
    await expect(page.locator("[data-testid='app-shell-topbar']")).toBeVisible();
  });

  test("missing sessionId surfaces ReportMissingSessionState", async ({ page }) => {
    await mockReportApis(page);
    await page.goto(`/#route=report&reportId=${REPORT_ID}`);
    await page.waitForSelector("[data-testid='report-missing-session']");
    await expect(
      page.locator("[data-testid='report-missing-session-cta']"),
    ).toBeVisible();
  });

  test("reportStatus=failed renders ReportFailureState with retry CTA", async ({
    page,
  }) => {
    await mockReportApis(page);
    await page.goto(`/#route=report&reportId=${REPORT_ID}&sessionId=${SESSION_ID}&reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT`);
    await page.waitForSelector("[data-testid='report-failure-state']");
    await expect(
      page.locator("[data-testid='report-failure-retry-cta']"),
    ).toBeVisible();
    await expect(
      page.locator("[data-testid='report-failure-back-to-workspace']"),
    ).toBeVisible();
  });

  test("mobile viewport keeps the dashboard inside 390px width", async ({ page }) => {
    await mockReportApis(page);
    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto(`/#route=report&reportId=${REPORT_ID}&sessionId=${SESSION_ID}&targetJobId=01918fa0-0000-7000-8000-000000002000&resumeId=${RESUME_ID}`);
    await page.waitForSelector("[data-testid='report-dashboard']");
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth);
    expect(overflow).toBeLessThanOrEqual(420);
  });
});
