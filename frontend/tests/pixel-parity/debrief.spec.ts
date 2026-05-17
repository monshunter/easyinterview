import { expect, test } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

/**
 * frontend-debrief/001 Playwright parity gate.
 *
 * Clean-checkout evidence uses DOM anchors, computed style, bounding boxes,
 * responsive geometry, theme/custom-accent smoke, and non-empty screenshots.
 * It deliberately avoids snapshot baselines until a stable debrief baseline is
 * maintained in a separate update.
 */

interface Rect {
  left: number;
  top: number;
  right: number;
  bottom: number;
  width: number;
  height: number;
}

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

const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";
const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";
const RESUME_VERSION_ID = "0195f2d0-0001-7000-8000-000000000202";

function fixtureResponse(relativePath: string, scenario = "default") {
  const absolutePath = resolve(process.cwd(), "..", relativePath);
  const fixture = JSON.parse(
    readFileSync(absolutePath, "utf8"),
  ) as OperationFixture;
  const response = fixture.scenarios[scenario]?.response;
  if (!response) {
    throw new Error(`missing fixture scenario ${relativePath}#${scenario}`);
  }
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

async function mockDebriefApis(page: import("@playwright/test").Page) {
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    const path = url.pathname.replace(/^\/api\/v1/, "");
    if (path === "/runtime-config") {
      await fulfillFixture(route, "openapi/fixtures/Auth/getRuntimeConfig.json");
      return;
    }
    if (path === "/me") {
      await fulfillFixture(
        route,
        "openapi/fixtures/Auth/getMe.json",
        "authenticated",
      );
      return;
    }
    if (/^\/targets\/[^/]+$/.test(path)) {
      await fulfillFixture(route, "openapi/fixtures/TargetJobs/getTargetJob.json");
      return;
    }
    if (/^\/practice\/sessions\/[^/]+$/.test(path)) {
      await fulfillFixture(
        route,
        "openapi/fixtures/PracticeSessions/getPracticeSession.json",
      );
      return;
    }
    if (/^\/resume-versions\/[^/]+$/.test(path)) {
      await fulfillFixture(
        route,
        "openapi/fixtures/Resumes/getResumeVersion.json",
      );
      return;
    }
    if (path === "/debriefs/question-suggestions") {
      await fulfillFixture(
        route,
        "openapi/fixtures/Debriefs/suggestDebriefQuestions.json",
      );
      return;
    }
    if (path === "/debriefs") {
      await fulfillFixture(route, "openapi/fixtures/Debriefs/createDebrief.json");
      return;
    }
    if (/^\/debriefs\/[^/]+$/.test(path)) {
      await fulfillFixture(route, "openapi/fixtures/Debriefs/getDebrief.json");
      return;
    }
    if (/^\/jobs\/[^/]+$/.test(path)) {
      await route.fulfill({
        status: 200,
        headers: { "content-type": "application/json; charset=utf-8" },
        body: JSON.stringify({
          id: "01918fa0-0000-7000-8000-00000000d200",
          jobType: "debrief_generate",
          status: "succeeded",
          resourceType: "debrief",
          resourceId: "01918fa0-0000-7000-8000-00000000d100",
          errorCode: null,
          createdAt: "2026-05-17T00:00:00Z",
          updatedAt: "2026-05-17T00:00:01Z",
        }),
      });
      return;
    }
    await route.fulfill({
      status: 404,
      headers: { "content-type": "application/json; charset=utf-8" },
      body: JSON.stringify({
        error: { code: "NOT_FOUND", message: `No fixture for ${path}` },
      }),
    });
  });
}

async function freezeAnimations(page: import("@playwright/test").Page) {
  await page.addStyleTag({
    content: `
      *, *::before, *::after {
        animation: none !important;
        animation-duration: 0s !important;
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

async function goToDebrief(
  page: import("@playwright/test").Page,
  routeName = "debrief",
) {
  await mockDebriefApis(page);
  await page.goto(
    `/#route=${routeName}&targetJobId=${TARGET_JOB_ID}&sessionId=${SESSION_ID}&resumeVersionId=${RESUME_VERSION_ID}`,
  );
  await page.waitForSelector("[data-testid='route-debrief']");
  await page.waitForSelector("[data-testid='debrief-guided-current']");
  await freezeAnimations(page);
}

async function rectOf(
  page: import("@playwright/test").Page,
  selector: string,
): Promise<Rect> {
  return page.evaluate(({ selector }) => {
    const el = document.querySelector(selector) as HTMLElement | null;
    if (!el) throw new Error(`selector not found: ${selector}`);
    const r = el.getBoundingClientRect();
    return {
      left: r.left,
      top: r.top,
      right: r.right,
      bottom: r.bottom,
      width: r.width,
      height: r.height,
    };
  }, { selector });
}

test.describe("debrief screen DOM and geometry parity", () => {
  test("debrief_full hash input normalizes into the formal debrief route", async ({
    page,
  }) => {
    await goToDebrief(page, "debrief_full");
    await expect(page.locator("[data-testid='route-debrief']")).toHaveAttribute(
      "data-route-name",
      "debrief",
    );
    await expect(page.locator("[data-testid='app-shell-topbar']")).toBeVisible();
    await expect(page.locator("[data-testid='topbar-nav-debrief']")).toHaveAttribute(
      "aria-current",
      "page",
    );
  });

  test("renders the source-level Step 0 anchor set", async ({ page }) => {
    await goToDebrief(page);
    for (const id of [
      "debrief-header",
      "debrief-context-strip",
      "debrief-context-card-targetJob",
      "debrief-context-card-mockSession",
      "debrief-context-card-resume",
      "debrief-stepper",
      "debrief-step-panel-0",
      "debrief-record-summary",
      "debrief-mode-toggle",
      "debrief-record-workspace",
      "debrief-vibe-check",
      "debrief-guided-record",
      "debrief-guided-current",
      "debrief-guided-card-list",
      "debrief-guided-active-card",
      "debrief-guided-entries",
      "debrief-submit-btn",
    ]) {
      await expect(page.locator(`[data-testid='${id}']`), id).toHaveCount(1);
    }
  });

  test("text mode mirrors the prototype record workspace hierarchy", async ({
    page,
  }, testInfo) => {
    await goToDebrief(page);
    const workspace = await rectOf(page, "[data-testid='debrief-record-workspace']");
    const guided = await rectOf(page, "[data-testid='debrief-guided-record']");
    const vibe = await rectOf(page, "[data-testid='debrief-vibe-check']");
    expect(guided.left).toBeGreaterThanOrEqual(workspace.left - 1);
    if (testInfo.project.name === "mobile") {
      expect(vibe.top).toBeGreaterThan(guided.bottom - 4);
    } else {
      expect(vibe.left).toBeGreaterThan(guided.right - 4);
    }
    expect(vibe.right).toBeLessThanOrEqual(workspace.right + 1);
    if (testInfo.project.name !== "mobile") {
      await expect(page.locator("[data-testid='debrief-guided-current-icon']")).toBeVisible();
    }
    await expect(page.locator("[data-testid='debrief-guided-card-list']")).toBeVisible();
    await expect(page.locator("[data-testid='debrief-guided-active-card']")).toBeVisible();
    await expect(page.locator("[data-testid='debrief-vibe-mood']")).toBeVisible();
  });

  test("voice mode shows the source prototype intro card instead of a flat placeholder", async ({
    page,
  }, testInfo) => {
    await goToDebrief(page);
    await page.locator("[data-testid='debrief-mode-toggle-voice']").click();
    await expect(page.locator("[data-testid='debrief-voice-intro-card']")).toBeVisible();
    await expect(page.locator("[data-testid='debrief-voice-topic-list']")).toBeVisible();
    await expect(page.locator("[data-testid='debrief-voice-start']")).toBeVisible();
    const card = await rectOf(page, "[data-testid='debrief-voice-intro-card']");
    const workspace = await rectOf(page, "[data-testid='debrief-record-workspace']");
    expect(card.width).toBeLessThanOrEqual(540);
    if (testInfo.project.name === "mobile") {
      expect(card.left).toBeGreaterThanOrEqual(workspace.left - 1);
      expect(card.right).toBeLessThanOrEqual(workspace.right + 1);
    } else {
      expect(card.left).toBeGreaterThan(workspace.left + 80);
    }
  });

  test("voice start opens the prototype continuous conversation state", async ({
    page,
  }, testInfo) => {
    await goToDebrief(page);
    await page.locator("[data-testid='debrief-mode-toggle-voice']").click();
    await page.locator("[data-testid='debrief-voice-start']").click();
    await expect(page.locator("[data-testid='debrief-voice-chat']")).toBeVisible();
    await expect(page.locator("[data-testid='debrief-voice-intro-card']")).toBeHidden();
    await expect(page.locator("[data-testid='debrief-voice-status']")).toBeVisible();
    await expect(page.locator("[data-testid='debrief-voice-live-extract']")).toBeVisible();
    await expect(page.locator("[data-testid='debrief-voice-extracted-card']")).toHaveCount(3);
    await expect(page.locator("[data-testid='debrief-voice-end-review']")).toBeVisible();
    const thread = await rectOf(page, "[data-testid='debrief-voice-thread']");
    const extract = await rectOf(page, "[data-testid='debrief-voice-live-extract']");
    if (testInfo.project.name === "mobile") {
      expect(extract.top).toBeGreaterThan(thread.bottom - 4);
    } else {
      expect(extract.left).toBeGreaterThan(thread.right - 4);
    }
    await page.locator("[data-testid='debrief-voice-end-review']").click();
    await expect(page.locator("[data-testid='debrief-voice-review']")).toBeVisible();
    await page.locator("[data-testid='debrief-voice-save']").click();
    await expect(page.locator("[data-testid='debrief-voice-committed']")).toBeVisible();
    await expect(page.locator("[data-testid='debrief-record-summary-count']")).toHaveText("3");
    await expect(page.locator("[data-testid='debrief-chip-voice']")).toContainText("3");
  });

  test("primary debrief regions stay inside the viewport", async ({ page }) => {
    await goToDebrief(page);
    const viewport = page.viewportSize();
    expect(viewport).not.toBeNull();
    for (const selector of [
      "[data-testid='route-debrief']",
      "[data-testid='debrief-header']",
      "[data-testid='debrief-context-strip']",
      "[data-testid='debrief-stepper']",
      "[data-testid='debrief-step-panel-0']",
      "[data-testid='debrief-submit-bar']",
    ]) {
      const rect = await rectOf(page, selector);
      expect(rect.left, selector).toBeGreaterThanOrEqual(-1);
      expect(rect.right, selector).toBeLessThanOrEqual(viewport!.width + 1);
    }
  });

  test("mobile layout avoids horizontal overflow", async ({ page }, testInfo) => {
    test.skip(testInfo.project.name !== "mobile", "mobile-only responsive check");
    await goToDebrief(page);
    const overflow = await page.evaluate(
      () => document.documentElement.scrollWidth,
    );
    expect(overflow).toBeLessThanOrEqual(420);
  });

  test("dark mode and custom accent change debrief computed values", async ({
    page,
  }) => {
    await goToDebrief(page);
    const card = page.locator("[data-testid='debrief-context-card-targetJob']");
    const lightBg = await card.evaluate((node) => getComputedStyle(node).backgroundColor);
    await page.evaluate(() => {
      document.documentElement.setAttribute("data-mode", "dark");
    });
    const darkBg = await card.evaluate((node) => getComputedStyle(node).backgroundColor);
    expect(lightBg).not.toBe(darkBg);

    const button = page.locator("[data-testid='debrief-submit-btn']");
    const defaultAccent = await button.evaluate(
      (node) => getComputedStyle(node).backgroundColor,
    );
    await page.evaluate(() => {
      document.documentElement.setAttribute("data-custom-accent", "active");
      document.documentElement.style.setProperty(
        "--ei-color-accent",
        "oklch(55% 0.18 215)",
      );
      document.documentElement.style.setProperty(
        "--ei-color-accent-soft",
        "oklch(94% 0.04 215)",
      );
    });
    const customAccent = await button.evaluate(
      (node) => getComputedStyle(node).backgroundColor,
    );
    expect(defaultAccent).not.toBe(customAccent);
  });

  test("screenshot smoke is non-empty without a checked-in baseline", async ({
    page,
  }) => {
    await goToDebrief(page);
    const image = await page.screenshot({ fullPage: false });
    expect(image.length).toBeGreaterThan(10_000);
  });
});
