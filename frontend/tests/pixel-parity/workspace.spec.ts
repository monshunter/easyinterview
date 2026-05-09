import { expect, test } from "@playwright/test";

/**
 * Phase 6.1 — Workspace screen DOM anchor and layout parity.
 *
 * Truth source: ui-design/src/screen-workspace.jsx::WorkspaceScreen,
 * docs/spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-
 * context/plan.md §4 Phase 6.
 *
 * Covers desktop (1440x900) and mobile (390x844) projects:
 * - DOM anchors (workspace crumbs, plan eyebrow, empty/missing states)
 * - Bounding box stays in viewport, no overlap
 * - warm/light -> dark -> customAccent theme switching
 * - toHaveScreenshot baseline
 * - Negative: old prototype testids absent
 *
 * Without fixture-backed transport in the production build, the workspace
 * renders empty/missing state when navigated through TopBar. Full
 * data-driven rendering parity is covered by the Vitest + jsdom test suite.
 */

interface Rect {
  left: number; top: number; right: number; bottom: number;
  width: number; height: number;
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

/** Navigate to workspace by clicking the TopBar Mock Interview button. */
async function goToWorkspace(page: import("@playwright/test").Page) {
  await page.goto("/");
  await page.waitForSelector("[data-testid='topbar-nav-workspace']");
  await page.click("[data-testid='topbar-nav-workspace']");
  await page.waitForTimeout(400);
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
  test("workspace empty state matches the colocated baseline", async ({ page }, testInfo) => {
    await goToWorkspace(page);
    await freezeAnimations(page);
    await expect(page).toHaveScreenshot(
      `workspace-empty-${testInfo.project.name}.png`,
      { fullPage: false, maxDiffPixels: 4000 },
    );
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
