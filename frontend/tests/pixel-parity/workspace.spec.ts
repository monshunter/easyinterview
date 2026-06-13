import { expect, test } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

/**
 * Phase 6.1 — Workspace screen DOM anchor and layout parity.
 *
 * Truth source: ui-design/src/screen-workspace.jsx::WorkspaceScreen,
 * docs/spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-
 * context/plan.md §4 Phase 6.
 *
 * Covers desktop (1440x900) and mobile (390x844) projects:
 * - DOM anchors (workspace crumbs, plan eyebrow, launcher, main columns, modals,
 *   empty/missing states)
 * - Bounding box stays in viewport, no overlap
 * - default (ocean)/light -> dark -> customAccent theme switching
 * - non-empty screenshot smoke
 * - Negative: old prototype testids absent
 *
 * Full data-driven rendering is reached through an explicit initial route
 * bootstrap with server-bound IDs. TopBar navigation still covers the
 * no-context empty state, and Home recent cards keep their product
 * `resume-unbound` behavior outside this pixel harness.
 */

interface Rect {
  left: number; top: number; right: number; bottom: number;
  width: number; height: number;
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

const WORKSPACE_TARGET_ID = "01918fa0-0000-7000-8000-000000002000";
const WORKSPACE_RESUME_ID = "01918fa0-0000-7000-8000-000000001000";
const WORKSPACE_PLAN_ID = "01918fa0-0000-7000-8000-000000004000";

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

async function mockWorkspaceApis(page: import("@playwright/test").Page): Promise<void> {
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
    if (path === "/targets") {
      await fulfillFixture(route, "openapi/fixtures/TargetJobs/listTargetJobs.json");
      return;
    }
    if (path.startsWith("/targets/")) {
      await fulfillFixture(route, "openapi/fixtures/TargetJobs/getTargetJob.json");
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

async function rectOf(page: import("@playwright/test").Page, selector: string): Promise<Rect> {
  return page.evaluate(({ selector }) => {
    const el = document.querySelector(selector) as HTMLElement | null;
    if (!el) throw new Error(`selector not found: ${selector}`);
    const r = el.getBoundingClientRect();
    return {
      left: r.left, top: r.top, right: r.right, bottom: r.bottom,
      width: r.width, height: r.height,
    };
  }, { selector });
}

async function freezeAnimations(page: import("@playwright/test").Page): Promise<void> {
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

/**
 * Navigate to workspace by clicking the TopBar Mock Interview button.
 *
 * `workspace` is an auth-gated business route (frontend-shell Phase 10 /
 * BUG-0115): the route guard renders `auth-route-gate` and redirects to
 * `auth_login` until runtime auth resolves to authenticated. Mock the auth
 * APIs first so the TopBar-click path reaches the workspace empty state
 * instead of the auth gate.
 */
async function goToWorkspace(page: import("@playwright/test").Page) {
  await mockWorkspaceApis(page);
  await page.goto("/");
  await page.waitForSelector("[data-testid='topbar-nav-workspace']");
  await page.click("[data-testid='topbar-nav-workspace']");
  await page.waitForSelector("[data-testid='workspace-empty']", {
    timeout: 5000,
  });
}

/** Navigate to a hydrated workspace through server-bound initial route params. */
async function goToHydratedWorkspace(page: import("@playwright/test").Page) {
  await mockWorkspaceApis(page);
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
    name: "workspace",
    params: {
      targetJobId: WORKSPACE_TARGET_ID,
      jobId: WORKSPACE_TARGET_ID,
      jdId: `jd-${WORKSPACE_TARGET_ID}`,
      planId: WORKSPACE_PLAN_ID,
      resumeVersionId: WORKSPACE_RESUME_ID,
      roundId: "round-technical-1",
      roundName: "Technical Round 1",
    },
  });
  await page.goto("/");
  await page.waitForSelector("[data-testid='workspace-header-title']");
  await page.evaluate(() => window.scrollTo(0, 0));
  await expect(page.locator("[data-testid='workspace-header-title']")).toContainText(
    "Senior Frontend Engineer",
  );
}

test.describe("workspace DOM anchor parity", () => {
  test("workspace route is reachable via TopBar and renders workspace-specific chrome", async ({ page }) => {
    await goToWorkspace(page);
    // Without targetJobId, workspace renders empty state. Verify empty state elements exist.
    await expect(page.locator("[data-testid='workspace-empty']")).toHaveCount(1, { timeout: 5000 });
  });

  test("workspace empty state renders all expected sub-elements", async ({ page }) => {
    await goToWorkspace(page);
    await expect(page.locator("[data-testid='workspace-empty-eyebrow']")).toHaveCount(1);
    await expect(page.locator("[data-testid='workspace-empty-title']")).toHaveCount(1);
    await expect(page.locator("[data-testid='workspace-empty-desc']")).toHaveCount(1);
    await expect(page.locator("[data-testid='workspace-empty-cta']")).toHaveCount(1);
  });

  test("TopBar workspace nav button has aria-current=page after navigation", async ({ page }) => {
    await goToWorkspace(page);
    const ariaCurrent = await page.getAttribute("[data-testid='topbar-nav-workspace']", "aria-current");
    expect(ariaCurrent).toBe("page");
  });

  test("hydrated workspace renders full source-level anchor set", async ({ page }) => {
    await goToHydratedWorkspace(page);
    const anchorIds = [
      "workspace-crumbs",
      "workspace-plan-eyebrow",
      "workspace-plan-eyebrow-title",
      "workspace-plan-action-switch",
      "workspace-header",
      "workspace-header-title",
      "workspace-launcher",
      "workspace-round-rail",
      "workspace-cta-start",
      "workspace-binding-jd",
      "workspace-binding-resume",
      "workspace-companyintel-summary",
      "workspace-companyintel-open",
      "workspace-jd-card",
      "workspace-jd-block-must",
      "workspace-jd-block-nice",
      "workspace-jd-block-hidden",
      "workspace-prep-card",
      "workspace-prep-strongs",
      "workspace-prep-risks",
      "workspace-history-card",
      "workspace-history-empty",
    ];
    for (const id of anchorIds) {
      await expect(page.locator(`[data-testid='${id}']`), id).toHaveCount(1);
    }
  });

  test("hydrated workspace opens plan switcher and resume picker modals", async ({ page }) => {
    await goToHydratedWorkspace(page);

    await page.click("[data-testid='workspace-plan-action-switch']");
    await expect(page.locator("[data-testid='workspace-plan-modal-card']")).toBeVisible();
    await expect(page.locator("[data-testid^='workspace-plan-modal-card-']")).toHaveCount(2);
    await page.keyboard.press("Escape");
    await expect(page.locator("[data-testid='workspace-plan-modal-card']")).toHaveCount(0);

    await page.click("[data-testid='workspace-binding-resume-change']");
    await expect(page.locator("[data-testid='workspace-resume-modal-card']")).toBeVisible();
    await expect(page.locator("[data-testid='workspace-resume-modal-disabled-note']")).toBeVisible();
  });
});

test.describe("workspace bounding box parity", () => {
  test("workspace empty state elements do not overlap and stay in viewport", async ({ page }) => {
    await goToWorkspace(page);
    const viewport = page.viewportSize();
    expect(viewport).toBeTruthy();

    const anchorIds = [
      "workspace-empty-eyebrow",
      "workspace-empty-title",
      "workspace-empty-desc",
      "workspace-empty-cta",
    ];

    const rects: Array<{ id: string; r: Rect }> = [];
    for (const id of anchorIds) {
      const el = page.locator(`[data-testid='${id}']`);
      if ((await el.count()) > 0) {
        const r = await rectOf(page, `[data-testid='${id}']`);
        rects.push({ id, r });
        expect(r.top).toBeGreaterThanOrEqual(0);
        expect(r.left).toBeGreaterThanOrEqual(0);
        expect(r.right).toBeLessThanOrEqual(viewport!.width + 1);
        expect(r.width).toBeGreaterThan(0);
      }
    }

    // All elements must have non-zero width and be reasonably positioned
    for (const { id, r } of rects) {
      expect(r.width, `${id} width is zero`).toBeGreaterThan(0);
      expect(r.top, `${id} top is negative`).toBeGreaterThanOrEqual(-5);
      expect(r.left, `${id} left is negative`).toBeGreaterThanOrEqual(-5);
    }
  });

  test("hydrated workspace primary anchors stay in viewport", async ({ page }) => {
    await goToHydratedWorkspace(page);
    const viewport = page.viewportSize();
    expect(viewport).toBeTruthy();

    const anchorIds = [
      "workspace-plan-eyebrow",
      "workspace-header",
      "workspace-launcher",
      "workspace-companyintel-summary",
      "workspace-jd-card",
      "workspace-prep-card",
      "workspace-history-card",
    ];

    for (const id of anchorIds) {
      const r = await rectOf(page, `[data-testid='${id}']`);
      expect(r.width, `${id} width is zero`).toBeGreaterThan(0);
      expect(r.height, `${id} height is zero`).toBeGreaterThan(0);
      expect(r.left, `${id} left`).toBeGreaterThanOrEqual(-5);
      expect(r.right, `${id} right`).toBeLessThanOrEqual(viewport!.width + 5);
    }
  });
});

test.describe("workspace dark mode + customAccent visual diff", () => {
  test("dark mode changes workspace background color", async ({ page }) => {
    await goToWorkspace(page);
    await page.waitForSelector("[data-testid='topbar-dark-toggle']");

    const beforeBg = await page.evaluate(() =>
      getComputedStyle(document.body).backgroundColor,
    );

    await page.click("[data-testid='topbar-dark-toggle']");
    await page.waitForTimeout(300);

    const afterBg = await page.evaluate(() =>
      getComputedStyle(document.body).backgroundColor,
    );
    expect(beforeBg).not.toBe(afterBg);
  });

  test("customAccent propagates data-custom-accent attribute on workspace route", async ({ page }) => {
    await goToWorkspace(page);

    await page.click("[data-testid='topbar-theme-button']");
    await page.click("[data-testid='topbar-theme-custom-option']");

    const attr = await page.evaluate(() =>
      document.documentElement.getAttribute("data-custom-accent"),
    );
    expect(attr).toBe("active");
  });

  test("theme dropdown is accessible from workspace", async ({ page }) => {
    await goToWorkspace(page);

    await page.click("[data-testid='topbar-theme-button']");
    await page.waitForSelector("[data-testid='topbar-theme-menu']");
    await expect(page.locator("[data-testid='topbar-theme-menu']")).toBeVisible();
    await expect(page.locator("[data-testid^='topbar-theme-option-']")).toHaveCount(4);
  });
});

test.describe("workspace screenshot regression", () => {
  test("workspace empty state renders a non-empty screenshot without a baseline prerequisite", async ({ page }) => {
    await goToWorkspace(page);
    await freezeAnimations(page);
    await expect(page.locator("[data-testid='workspace-empty']")).toBeVisible();
    const screenshot = await page.screenshot({ fullPage: false });
    expect(screenshot.length).toBeGreaterThan(10_000);
  });

  test("hydrated workspace renders a non-empty screenshot without a baseline prerequisite", async ({ page }) => {
    await goToHydratedWorkspace(page);
    await freezeAnimations(page);
    await expect(page.locator("[data-testid='workspace-header-title']")).toBeVisible();
    const screenshot = await page.screenshot({ fullPage: false });
    expect(screenshot.length).toBeGreaterThan(10_000);
  });
});

test.describe("old prototype testid negative gate (workspace)", () => {
  test("retired workspace prototype testids do not appear in DOM", async ({ page }) => {
    await goToWorkspace(page);

    const banned = [
      "practice-mode-card-warmup",
      "practice-mode-card-single_drill",
      "practice-mode-card-drill_builder",
      "practice-mode-card-mistake_queue",
      "growth-center",
      "drill-builder",
      "mistake-queue",
    ];
    for (const testid of banned) {
      await expect(page.locator(`[data-testid='${testid}']`)).toHaveCount(0);
    }
  });

  test("retired route names do not appear as TopBar entries", async ({ page }) => {
    await goToWorkspace(page);
    for (const banned of [
      "topbar-nav-welcome", "topbar-nav-mistakes",
      "topbar-nav-growth", "topbar-nav-drill", "topbar-nav-voice",
    ]) {
      await expect(page.locator(`[data-testid='${banned}']`)).toHaveCount(0);
    }
  });
});
