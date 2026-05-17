import { expect, test } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

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

const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";
const PLAN_ID = "01918fa0-0000-7000-8000-000000004000";
const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";
const RESUME_ID = "01918fa0-0000-7000-8000-000000001000";

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

async function mockPracticeApis(
  page: import("@playwright/test").Page,
  getScenario = "default",
) {
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    const path = url.pathname.replace(/^\/api\/v1/, "");
    if (path === "/runtime-config") {
      await fulfillFixture(route, "openapi/fixtures/Auth/getRuntimeConfig.json");
      return;
    }
    if (path === "/me") {
      await fulfillFixture(route, "openapi/fixtures/Auth/getMe.json", "authenticated");
      return;
    }
    if (/^\/practice\/sessions\/[^/]+$/.test(path)) {
      await fulfillFixture(
        route,
        "openapi/fixtures/PracticeSessions/getPracticeSession.json",
        getScenario,
      );
      return;
    }
    if (/^\/practice\/sessions\/[^/]+\/events$/.test(path)) {
      await fulfillFixture(
        route,
        "openapi/fixtures/PracticeSessions/appendSessionEvent.json",
      );
      return;
    }
    if (/^\/practice\/sessions\/[^/]+\/complete$/.test(path)) {
      await fulfillFixture(
        route,
        "openapi/fixtures/PracticeSessions/completePracticeSession.json",
      );
      return;
    }
    await route.fulfill({
      status: 404,
      headers: { "content-type": "application/json; charset=utf-8" },
      body: JSON.stringify({ error: { code: "NOT_FOUND", message: path } }),
    });
  });
}

async function goToPractice(
  page: import("@playwright/test").Page,
  params: Record<string, string> = {},
  getScenario = "default",
) {
  await mockPracticeApis(page, getScenario);
  await page.addInitScript((route) => {
    (
      window as Window & {
        __EASYINTERVIEW_INITIAL_ROUTE__?: {
          name: string;
          params: Record<string, string>;
        };
      }
    ).__EASYINTERVIEW_INITIAL_ROUTE__ = route;
  }, {
    name: "practice",
    params: {
      sessionId: SESSION_ID,
      planId: PLAN_ID,
      targetJobId: TARGET_JOB_ID,
      jdId: `jd-${TARGET_JOB_ID}`,
      resumeVersionId: RESUME_ID,
      roundId: "round-technical-1",
      mode: "text",
      modality: "text",
      practiceMode: "assisted",
      practiceGoal: "baseline",
      hintUsed: "false",
      hintCount: "0",
      ...params,
    },
  });
  await page.goto("/");
  await page.waitForSelector("[data-testid='practice-screen']");
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

test.describe("practice screen DOM and geometry parity", () => {
  test("renders the text-mode source-level anchor set", async ({ page }) => {
    await goToPractice(page);
    for (const id of [
      "practice-topbar",
      "practice-topbar-mode-segment",
      "practice-sessionmap",
      "practice-question",
      "practice-transcript",
      "practice-input",
      "practice-input-textarea",
      "practice-rightpanel",
      "practice-rightpanel-ai-transparency",
      "practice-rightpanel-cta-finish-wrap",
      "practice-rightpanel-cta-finish",
    ]) {
      await expect(page.locator(`[data-testid='${id}']`), id).toHaveCount(1);
    }
  });

  test("primary practice anchors stay inside the viewport", async ({ page }) => {
    await goToPractice(page);
    const viewport = page.viewportSize();
    expect(viewport).toBeTruthy();

    for (const selector of [
      "[data-testid='practice-topbar']",
      "[data-testid='practice-main']",
      "[data-testid='practice-center']",
      "[data-testid='practice-input']",
      "[data-testid='practice-rightpanel']",
    ]) {
      const rect = await rectOf(page, selector);
      expect(rect.left, selector).toBeGreaterThanOrEqual(-1);
      expect(rect.right, selector).toBeLessThanOrEqual(viewport!.width + 1);
    }
  });

  test("mobile layout folds the three columns without horizontal overflow", async ({ page }, testInfo) => {
    test.skip(testInfo.project.name !== "mobile", "mobile-only responsive check");
    await goToPractice(page);
    const viewport = page.viewportSize();
    expect(viewport).toBeTruthy();
    const main = await rectOf(page, "[data-testid='practice-main']");
    const right = await rectOf(page, "[data-testid='practice-rightpanel']");
    expect(main.right).toBeLessThanOrEqual(viewport!.width + 1);
    expect(right.right).toBeLessThanOrEqual(viewport!.width + 1);
  });

  test("voice mode renders ui-design voice surface anchors and keeps them in the viewport", async ({ page }) => {
    await goToPractice(page, { mode: "voice", modality: "voice" });
    await expect(page.locator("[data-testid='practice-voice-coming-soon']")).toHaveCount(0);
    for (const id of [
      "practice-voice-surface",
      "practice-voice-waveform",
      "practice-voice-annotated-waveform",
      "practice-voice-live-transcript",
      "practice-voice-expression-panel",
    ]) {
      await expect(page.locator(`[data-testid='${id}']`), id).toHaveCount(1);
    }

    const viewport = page.viewportSize();
    expect(viewport).toBeTruthy();
    for (const selector of [
      "[data-testid='practice-voice-waveform']",
      "[data-testid='practice-voice-annotated-waveform']",
      "[data-testid='practice-voice-expression-panel']",
    ]) {
      const rect = await rectOf(page, selector);
      expect(rect.left, selector).toBeGreaterThanOrEqual(-1);
      expect(rect.right, selector).toBeLessThanOrEqual(viewport!.width + 1);
      expect(rect.height, selector).toBeGreaterThan(24);
    }
  });

  test("dark mode and custom accent visibly change computed values", async ({ page }) => {
    await goToPractice(page);
    const lightBg = await page.locator("[data-testid='practice-screen']").evaluate(
      (node) => getComputedStyle(node).backgroundColor,
    );
    await page.evaluate(() => {
      document.documentElement.setAttribute("data-mode", "dark");
    });
    const darkBg = await page.locator("[data-testid='practice-screen']").evaluate(
      (node) => getComputedStyle(node).backgroundColor,
    );
    expect(lightBg).not.toBe(darkBg);

    const sendButton = page.locator("[data-testid='practice-input-send']");
    const defaultAccent = await sendButton.evaluate(
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
    const accentAfter = await sendButton.evaluate(
      (node) => getComputedStyle(node).backgroundColor,
    );
    const customAccent = await page.evaluate(() =>
      document.documentElement.style.getPropertyValue("--ei-color-accent"),
    );
    expect(defaultAccent).not.toBe(accentAfter);
    expect(customAccent).toContain("oklch");
  });

  test("screenshot smoke is non-empty without a checked-in baseline", async ({ page }) => {
    await goToPractice(page);
    const image = await page.screenshot({ fullPage: false });
    expect(image.length).toBeGreaterThan(10_000);
  });
});
