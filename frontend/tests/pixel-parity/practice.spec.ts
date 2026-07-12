import { expect, test } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

interface OperationFixture {
  scenarios: Record<string, { response: { status: number; headers?: Record<string, string>; body: unknown } }>;
}

const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";
const PLAN_ID = "01918fa0-0000-7000-8000-000000004000";
const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";
const RESUME_ID = "01918fa0-0000-7000-8000-000000001000";

function fixtureResponse(relativePath: string, scenario = "default") {
  const fixture = JSON.parse(readFileSync(resolve(process.cwd(), "..", relativePath), "utf8")) as OperationFixture;
  const response = fixture.scenarios[scenario]?.response;
  if (!response) throw new Error(`missing fixture scenario ${relativePath}#${scenario}`);
  return response;
}

async function fulfillFixture(route: import("@playwright/test").Route, relativePath: string, scenario = "default") {
  const response = fixtureResponse(relativePath, scenario);
  await route.fulfill({
    status: response.status,
    headers: { "content-type": "application/json; charset=utf-8", ...(response.headers ?? {}) },
    body: JSON.stringify(response.body),
  });
}

async function mockPracticeApis(page: import("@playwright/test").Page) {
  await page.route("**/api/v1/**", async (route) => {
    const path = new URL(route.request().url()).pathname.replace(/^\/api\/v1/, "");
    if (path === "/runtime-config") return fulfillFixture(route, "openapi/fixtures/Auth/getRuntimeConfig.json");
    if (path === "/me") return fulfillFixture(route, "openapi/fixtures/Auth/getMe.json", "authenticated");
    if (/^\/practice\/sessions\/[^/]+$/.test(path)) return fulfillFixture(route, "openapi/fixtures/PracticeSessions/getPracticeSession.json");
    if (/^\/targets\/[^/]+$/.test(path)) return fulfillFixture(route, "openapi/fixtures/TargetJobs/getTargetJob.json");
    if (/^\/practice\/sessions\/[^/]+\/messages$/.test(path)) return fulfillFixture(route, "openapi/fixtures/PracticeSessions/sendPracticeMessage.json");
    if (/^\/practice\/sessions\/[^/]+\/complete$/.test(path)) return fulfillFixture(route, "openapi/fixtures/PracticeSessions/completePracticeSession.json");
    await route.fulfill({ status: 404, headers: { "content-type": "application/json" }, body: JSON.stringify({ error: { code: "NOT_FOUND", message: path } }) });
  });
}

async function goToPractice(page: import("@playwright/test").Page) {
  await mockPracticeApis(page);
  await page.addInitScript((route) => {
    (window as Window & { __EASYINTERVIEW_INITIAL_ROUTE__?: { name: string; params: Record<string, string> } }).__EASYINTERVIEW_INITIAL_ROUTE__ = route;
  }, {
    name: "practice",
    params: {
      sessionId: SESSION_ID,
      planId: PLAN_ID,
      targetJobId: TARGET_JOB_ID,
      jdId: `jd-${TARGET_JOB_ID}`,
      resumeId: RESUME_ID,
      roundId: "round-technical-1",
      practiceGoal: "baseline",
    },
  });
  await page.goto("/");
  await page.waitForSelector("[data-testid='practice-screen']");
}

test.describe("practice continuous conversation parity", () => {
  test("renders one full-width chat with no structured-question surfaces", async ({ page }) => {
    await goToPractice(page);
    for (const id of [
      "practice-topbar",
      "practice-topbar-phone-toggle",
      "practice-conversation",
      "practice-transcript",
      "practice-input",
      "practice-input-textarea",
      "practice-finish-cta",
    ]) {
      await expect(page.locator(`[data-testid='${id}']`), id).toHaveCount(1);
    }
    for (const stale of ["practice-sessionmap", "practice-question", "practice-question-prompt", "practice-phone-surface"]) {
      await expect(page.locator(`[data-testid='${stale}']`), stale).toHaveCount(0);
    }
    await expect(page.locator("[data-testid='practice-topbar-phone-toggle']")).toBeDisabled();
  });

  test("conversation remains inside desktop and mobile viewports", async ({ page }) => {
    await goToPractice(page);
    const viewport = page.viewportSize();
    expect(viewport).toBeTruthy();
    const geometry = await page.locator("[data-testid='practice-conversation']").evaluate((node) => {
      const rect = node.getBoundingClientRect();
      return { left: rect.left, right: rect.right, width: rect.width, scrollWidth: document.documentElement.scrollWidth };
    });
    expect(geometry.left).toBeGreaterThanOrEqual(-1);
    expect(geometry.right).toBeLessThanOrEqual(viewport!.width + 1);
    expect(geometry.width).toBeGreaterThan(viewport!.width * 0.9);
    expect(geometry.scrollWidth).toBeLessThanOrEqual(viewport!.width);
  });

  test("disabled phone control keeps the prototype geometry", async ({ page }) => {
    await goToPractice(page);
    const style = await page.locator("[data-testid='practice-topbar-phone-toggle']").evaluate((node) => {
      const computed = getComputedStyle(node);
      return { width: computed.width, height: computed.height, borderRadius: computed.borderRadius, cursor: computed.cursor };
    });
    expect(style).toEqual({ width: "34px", height: "34px", borderRadius: "17px", cursor: "not-allowed" });
  });

  test("screenshot smoke is non-empty", async ({ page }) => {
    await goToPractice(page);
    const image = await page.screenshot({ fullPage: false });
    expect(image.length).toBeGreaterThan(10_000);
  });
});
