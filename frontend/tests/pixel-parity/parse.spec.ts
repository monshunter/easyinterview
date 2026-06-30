import { expect, test } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

/**
 * Phase 6.2 — Parse screen DOM anchor and loading state parity.
 *
 * Truth source: ui-design/src/screens-p0-complete.jsx::ParseScreen,
 * docs/spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-
 * parse/plan.md §4 Phase 6.
 *
 * The parse screen requires a targetJobId param to load. In Playwright
 * we can only test the initial loading state when navigated to via the
 * home import flow (paste JD -> submit). Without mock transport, the
 * import call will fail; the test asserts the DOM anchors that are
 * reachable in the SPA flow.
 *
 * Full e2e with fixture-backed transport is deferred to the scenario
 * gate (E2E.P0.015 / E2E.P0.016 under test/scenarios/e2e/).
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

function fixtureResponse(relativePath: string, scenario = "default") {
  const absolutePath = resolve(process.cwd(), "..", relativePath);
  const fixture = JSON.parse(readFileSync(absolutePath, "utf8")) as OperationFixture;
  const response = fixture.scenarios[scenario]?.response;
  if (!response) throw new Error(`missing fixture scenario ${relativePath}#${scenario}`);
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

async function mockParseReadyApis(page: import("@playwright/test").Page): Promise<void> {
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    const path = url.pathname.replace(/^\/api\/v1/, "");
    const method = route.request().method();
    if (path === "/runtime-config") {
      await fulfillFixture(route, "openapi/fixtures/Auth/getRuntimeConfig.json");
      return;
    }
    if (path === "/me") {
      await fulfillFixture(route, "openapi/fixtures/Auth/getMe.json", "authenticated");
      return;
    }
    if (method === "GET" && path.startsWith("/targets/")) {
      await fulfillFixture(route, "openapi/fixtures/TargetJobs/getTargetJob.json");
      return;
    }
    if (method === "GET" && path === "/resumes") {
      await fulfillFixture(route, "openapi/fixtures/Resumes/listResumes.json");
      return;
    }
    await route.fulfill({
      status: 404,
      headers: { "content-type": "application/json; charset=utf-8" },
      body: JSON.stringify({ error: { code: "NOT_FOUND", message: `No fixture for ${path}` } }),
    });
  });
}

async function mockParseConfirmApis(
  page: import("@playwright/test").Page,
  onUpdateTargetJob: (request: import("@playwright/test").Request) => Promise<void>,
): Promise<void> {
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    const path = url.pathname.replace(/^\/api\/v1/, "");
    const method = route.request().method();
    if (path === "/runtime-config") {
      await fulfillFixture(route, "openapi/fixtures/Auth/getRuntimeConfig.json");
      return;
    }
    if (path === "/me") {
      await fulfillFixture(route, "openapi/fixtures/Auth/getMe.json", "authenticated");
      return;
    }
    if (path === "/targets") {
      await fulfillFixture(route, "openapi/fixtures/TargetJobs/listTargetJobs.json");
      return;
    }
    if (method === "GET" && path.startsWith("/targets/")) {
      await fulfillFixture(route, "openapi/fixtures/TargetJobs/getTargetJob.json");
      return;
    }
    if (method === "GET" && path === "/resumes") {
      await fulfillFixture(route, "openapi/fixtures/Resumes/listResumes.json");
      return;
    }
    if (method === "PATCH" && path.startsWith("/targets/")) {
      await onUpdateTargetJob(route.request());
      await fulfillFixture(route, "openapi/fixtures/TargetJobs/updateTargetJob.json");
      return;
    }
    if (method === "POST" && path === "/practice/plans") {
      await fulfillFixture(route, "openapi/fixtures/PracticePlans/createPracticePlan.json");
      return;
    }
    if (method === "POST" && path === "/practice/sessions") {
      await fulfillFixture(route, "openapi/fixtures/PracticeSessions/startPracticeSession.json");
      return;
    }
    if (method === "GET" && path.startsWith("/practice/sessions/")) {
      await fulfillFixture(route, "openapi/fixtures/PracticeSessions/getPracticeSession.json");
      return;
    }
    if (path.startsWith("/resumes/")) {
      await fulfillFixture(route, "openapi/fixtures/Resumes/getResume.json");
      return;
    }
    if (path.startsWith("/practice/plans/")) {
      await fulfillFixture(route, "openapi/fixtures/PracticePlans/getPracticePlan.json");
      return;
    }
    await route.fulfill({
      status: 404,
      headers: { "content-type": "application/json; charset=utf-8" },
      body: JSON.stringify({ error: { code: "NOT_FOUND", message: `No fixture for ${path}` } }),
    });
  });
}

async function freezeVisualAnimations(page: import("@playwright/test").Page): Promise<void> {
  await page.addStyleTag({
    content: `
      *, *::before, *::after {
        animation: none !important;
        transition: none !important;
        caret-color: transparent !important;
      }
    `,
  });
  await page.evaluate(async () => {
    if (document.fonts && typeof document.fonts.ready?.then === "function") {
      await document.fonts.ready;
    }
  });
}

test.describe("parse screen DOM anchor parity", () => {
  test("home screen renders parse entry points (textarea + submit)", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-jd-textarea']");

    await expect(page.locator("[data-testid='home-jd-textarea']")).toBeEnabled();
    await expect(page.locator("[data-testid='home-jd-submit']")).toBeVisible();

    // Submit should be disabled when textarea is empty
    await expect(page.locator("[data-testid='home-jd-submit']")).toBeDisabled();
  });

  test("home jd textarea accepts input and submit enables", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-jd-textarea']");

    await page.fill(
      "[data-testid='home-jd-textarea']",
      "Senior Frontend Engineer needed at Acme Corp",
    );

    // Submit button should become enabled
    await expect(page.locator("[data-testid='home-jd-submit']")).toBeEnabled();
  });

  test("upload modal opens and closes", async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-jd-textarea']");

    // Click upload button (or upload link)
    const uploadTrigger = page.locator("[data-testid='home-jd-upload-trigger']");
    if ((await uploadTrigger.count()) > 0) {
      await uploadTrigger.click();
      await expect(
        page.locator("[data-testid='home-modal-upload-dropzone']"),
      ).toBeVisible();

      // Close with X
      await page.click("[data-testid='home-modal-upload-close']");
      await expect(
        page.locator("[data-testid='home-modal-upload-dropzone']"),
      ).toHaveCount(0);
    }
  });

  test("ready target job response keeps ui-design loading demo before preview", async ({
    page,
  }, testInfo) => {
    await mockParseReadyApis(page);

    await page.goto("/parse?targetJobId=01918fa0-0000-7000-8000-000000002000");
    await page.waitForSelector("[data-testid='parse-loading-step-0']");
    await expect(page.locator("[data-testid='route-parse']")).toHaveCount(1);
    await expect(page.locator("[data-testid='parse-loading-step-0']")).toBeVisible();
    await expect(page.locator("[data-testid='parse-loading-step-1']")).toBeVisible();
    await expect(page.locator("[data-testid='parse-loading-step-2']")).toBeVisible();
    await expect(page.locator("[data-testid='parse-loading-step-3']")).toBeVisible();
    await expect(page.locator("[data-testid='parse-loading-footer']")).toBeVisible();
    await expect(page.locator("[data-testid='parse-basics-title']")).toHaveCount(0);

    await freezeVisualAnimations(page);
    const screenshot = await page.locator("[data-testid='route-parse']").screenshot();
    await testInfo.attach("parse-ready-response-loading-demo", {
      body: screenshot,
      contentType: "image/png",
    });
    expect(screenshot.length).toBeGreaterThan(10_000);
    console.log(
      `E2E.P0.015 ready-response loading browser gate screenshotBytes=${screenshot.length}`,
    );

    await page.waitForTimeout(1_000);
    await expect(page.locator("[data-testid='parse-loading-step-0']")).toBeVisible();
    await expect(page.locator("[data-testid='parse-basics-title']")).toHaveCount(0);

    await page.waitForTimeout(2_600);
    await expect(page.locator("[data-testid='parse-basics-title']")).toBeVisible();
  });

  test("save plan navigates to workspace with bound resume context", async ({
    page,
  }, testInfo) => {
    const updateCalls: Array<{
      body: unknown;
      idempotencyKey: string | null;
    }> = [];
    await mockParseConfirmApis(page, async (request) => {
      updateCalls.push({
        body: request.postDataJSON(),
        idempotencyKey: request.headers()["idempotency-key"] ?? null,
      });
    });

    await page.goto("/parse?targetJobId=01918fa0-0000-7000-8000-000000002000");
    await page.waitForSelector("[data-testid='parse-basics-title']", { timeout: 5_000 });
    await expect(page.locator("[data-testid='parse-resume-binding']")).toContainText(
      "Choose the resume for this interview",
    );
    await expect(page.locator("[data-testid='parse-action-save-plan']")).toBeDisabled();
    await page.click(
      "[data-testid='parse-resume-option-01918fa0-0000-7000-8000-000000001000']",
    );
    await expect(page.locator("[data-testid='parse-resume-binding']")).toContainText(
      "Alice Example - Senior Frontend Engineer",
    );
    await expect(page.locator("[data-testid='parse-action-save-plan']")).toBeEnabled();
    await page.click("[data-testid='parse-action-save-plan']");
    await page.waitForURL(/\/workspace\?/);
    await expect(page.locator("[data-testid='workspace-missing-resume']")).toHaveCount(0);
    await expect(page.locator("[data-testid='workspace-cta-start']")).toBeVisible();

    expect(updateCalls).toHaveLength(1);
    expect(updateCalls[0]?.idempotencyKey).toBeTruthy();
    expect(updateCalls[0]?.body).toMatchObject({
      titleHint: "Senior Frontend Engineer",
      companyNameHint: "Acme",
      locationText: "Shanghai · Hybrid",
    });
    expect(updateCalls[0]?.body).not.toHaveProperty("level");
    expect(updateCalls[0]?.body).not.toHaveProperty("language");

    const url = new URL(page.url());
    const expectedParams = {
      targetJobId: "01918fa0-0000-7000-8000-000000002000",
      jobId: "01918fa0-0000-7000-8000-000000002000",
      jdId: "jd-01918fa0-0000-7000-8000-000000002000",
      planId: "plan-01918fa0-0000-7000-8000-000000002000",
      resumeId: "01918fa0-0000-7000-8000-000000001000",
      roundId: "round-technical-1",
      roundName: "Technical Round 1",
    };
    expect(url.pathname).toBe("/workspace");
    for (const [key, value] of Object.entries(expectedParams)) {
      expect(url.searchParams.get(key), key).toBe(value);
    }
    expect(url.searchParams.get("rawText")).toBeNull();
    expect(url.searchParams.get("sourceUrl")).toBeNull();
    expect(url.search).not.toContain("resume-unbound");

    await freezeVisualAnimations(page);
    const screenshot = await page.locator("[data-testid='workspace-launcher']").screenshot();
    await testInfo.attach("parse-save-plan-workspace-bound-resume", {
      body: screenshot,
      contentType: "image/png",
    });
    expect(screenshot.length).toBeGreaterThan(10_000);
    console.log(
      `E2E.P0.016 parse save-plan workspace browser gate contextKeys=${Object.keys(expectedParams).join(",")} resumeId=${expectedParams.resumeId} screenshotBytes=${screenshot.length}`,
    );
  });

  test("start interview hands off through workspace autoStart with bound resume", async ({
    page,
  }) => {
    const updateCalls: Array<{
      body: unknown;
      idempotencyKey: string | null;
    }> = [];
    await mockParseConfirmApis(page, async (request) => {
      updateCalls.push({
        body: request.postDataJSON(),
        idempotencyKey: request.headers()["idempotency-key"] ?? null,
      });
    });

    await page.goto("/parse?targetJobId=01918fa0-0000-7000-8000-000000002000");
    await page.waitForSelector("[data-testid='parse-basics-title']", { timeout: 5_000 });
    await expect(
      page.locator("[data-testid='parse-action-start-interview']"),
    ).toBeDisabled();
    await page.click(
      "[data-testid='parse-resume-option-01918fa0-0000-7000-8000-000000001000']",
    );
    await expect(page.locator("[data-testid='parse-action-start-interview']")).toBeEnabled();
    await page.click("[data-testid='parse-action-start-interview']");

    await expect(page.locator("[data-testid='practice-screen']")).toBeVisible({
      timeout: 10_000,
    });
    expect(updateCalls).toHaveLength(1);
    expect(updateCalls[0]?.idempotencyKey).toBeTruthy();

    const url = new URL(page.url());
    expect(url.pathname).toBe("/practice");
    expect(url.searchParams.get("targetJobId")).toBe(
      "01918fa0-0000-7000-8000-000000002000",
    );
    expect(url.searchParams.get("resumeId")).toBe(
      "01918fa0-0000-7000-8000-000000001000",
    );
    expect(url.search).not.toContain("resume-unbound");
    console.log(
      "E2E.P0.016 parse start-interview autoStart browser gate resumeId=01918fa0-0000-7000-8000-000000001000 route=practice",
    );
  });
});
