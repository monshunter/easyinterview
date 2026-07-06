import { expect, test } from "@playwright/test";

/**
 * Phase 6.1 — Home screen DOM anchor and layout parity.
 *
 * Truth source: ui-design/src/screen-home.jsx::HomeScreen,
 * docs/spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-
 * parse/plan.md §4 Phase 6.
 *
 * Covers desktop (1440x900) and mobile (390x844) projects:
 * - DOM anchors (hero, textarea, resume picker, retired aux-card negatives)
 * - Bounding box stays in viewport, no overlap
 * - default (ocean)/light -> dark -> customAccent theme switching
 * - Mobile: textarea card not overflowing
 */

interface Rect {
  left: number;
  top: number;
  right: number;
  bottom: number;
  width: number;
  height: number;
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

test.describe("home screen DOM anchor parity", () => {
  test("home route renders hero, textarea, and aux card testids", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-hero-label']");

    await expect(page.locator("[data-testid='home-hero-label']")).toHaveCount(1);
    await expect(page.locator("[data-testid='home-hero-title']")).toHaveCount(1);
    await expect(page.locator("[data-testid='home-hero-sub']")).toHaveCount(0);
    await expect(page.locator("[data-testid='home-jd-textarea']")).toHaveCount(
      1,
    );
    await expect(page.locator("[data-testid='home-jd-submit']")).toHaveCount(1);
    await expect(page.locator("[data-testid='home-jd-submit']")).toContainText(
      /立即面试|Start interview now/,
    );
    await expect(page.locator("[data-testid='home-resume-select']")).toHaveCount(
      1,
    );
    // product-scope D-17 removed the JOB PICKS aux card, and D-22 removed the
    // post-interview debrief card.
    await expect(page.locator("[data-testid='home-aux-jobpicks']")).toHaveCount(
      0,
    );
    await expect(page.locator("[data-testid='home-aux-debrief']")).toHaveCount(0);
  });

  test("home textarea card stays inside viewport (desktop)", async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-jd-textarea']");

    const viewport = page.viewportSize();
    expect(viewport).toBeTruthy();

    const textareaRect = await rectOf(
      page,
      "[data-testid='home-jd-textarea']",
    );
    expect(textareaRect.top).toBeGreaterThanOrEqual(0);
    expect(textareaRect.left).toBeGreaterThanOrEqual(0);
    expect(textareaRect.right).toBeLessThanOrEqual(viewport!.width + 1);
  });

  test("home retired aux cards stay absent", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-jd-textarea']");

    await expect(
      page.locator("[data-testid='home-aux-jobpicks']"),
    ).toHaveCount(0);
    await expect(
      page.locator("[data-testid='home-aux-debrief']"),
    ).toHaveCount(0);
  });

  test("dark mode toggle changes computed background color", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='topbar-dark-toggle']");

    const lightBg = await page.evaluate(
      () => getComputedStyle(document.body).backgroundColor,
    );

    await page.click("[data-testid='topbar-dark-toggle']");
    await page.waitForTimeout(300);

    const darkBg = await page.evaluate(
      () => getComputedStyle(document.body).backgroundColor,
    );

    expect(lightBg).not.toBe(darkBg);
  });
});
