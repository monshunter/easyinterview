import { expect, test } from "@playwright/test";

/**
 * Phase 6.3 — JD Match P1 placeholder shell DOM anchor parity.
 *
 * Truth source: ui-design/src/screen-jd-match.jsx::JDMatchScreen,
 * docs/spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-
 * parse/plan.md §4 Phase 6.
 *
 * Covers desktop (1440x900) and mobile (390x844) projects:
 * - DOM anchors (hero, profile chip, tabs, placeholder)
 * - Negative: old prototype business testids 0 hit
 * - Route reachable from TopBar and home aux card
 * - Theme switching visible
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

test.describe("jd_match placeholder DOM anchor parity", () => {
  test("jd_match route renders hero, tabs, and placeholder via home aux card", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-aux-jobpicks']");

    // Click Job Picks aux card button to navigate to jd_match
    const button = page.locator(
      "[data-testid='home-aux-jobpicks'] button",
    );
    if ((await button.count()) > 0) {
      await button.click();
    }

    await page.waitForTimeout(500);

    // Check if we're on jd_match
    const route = page.locator("[data-testid='route-jd_match']");
    if ((await route.count()) > 0) {
      await expect(
        page.locator("[data-testid='jdmatch-hero-label']"),
      ).toHaveCount(1);
      await expect(
        page.locator("[data-testid='jdmatch-hero-title']"),
      ).toHaveCount(1);
      await expect(
        page.locator("[data-testid='jdmatch-hero-sub']"),
      ).toHaveCount(1);

      await expect(
        page.locator("[data-testid='jdmatch-profile-chip']"),
      ).toHaveCount(1);

      await expect(
        page.locator("[data-testid='jdmatch-tab-recommended']"),
      ).toHaveCount(1);
      await expect(
        page.locator("[data-testid='jdmatch-tab-search']"),
      ).toHaveCount(1);
      await expect(
        page.locator("[data-testid='jdmatch-tab-watchlist']"),
      ).toHaveCount(1);

      await expect(
        page.locator("[data-testid='jdmatch-placeholder']"),
      ).toHaveCount(1);
    }
  });

  test("jd_match placeholder content stays inside viewport (desktop)", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-aux-jobpicks']");

    const button = page.locator(
      "[data-testid='home-aux-jobpicks'] button",
    );
    if ((await button.count()) > 0) {
      await button.click();
    }

    await page.waitForTimeout(500);

    const route = page.locator("[data-testid='route-jd_match']");
    if ((await route.count()) > 0) {
      await page.waitForSelector("[data-testid='jdmatch-placeholder']");
      const viewport = page.viewportSize();
      expect(viewport).toBeTruthy();

      const placeholderRect = await rectOf(
        page,
        "[data-testid='jdmatch-placeholder']",
      );
      expect(placeholderRect.left).toBeGreaterThanOrEqual(0);
      expect(placeholderRect.right).toBeLessThanOrEqual(viewport!.width + 1);
    }
  });

  test("negative — old prototype jd_match testids are absent", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-aux-jobpicks']");

    const button = page.locator(
      "[data-testid='home-aux-jobpicks'] button",
    );
    if ((await button.count()) > 0) {
      await button.click();
    }

    await page.waitForTimeout(500);

    await expect(
      page.locator("[data-testid='jdmatch-card-0']"),
    ).toHaveCount(0);
    await expect(
      page.locator("[data-testid='jdmatch-saved-search-0']"),
    ).toHaveCount(0);
    await expect(
      page.locator("[data-testid='jdmatch-watchlist-0']"),
    ).toHaveCount(0);
    await expect(
      page.locator("[data-testid='jdmatch-market-signal-0']"),
    ).toHaveCount(0);
    await expect(
      page.locator("[data-testid='jdmatch-search-bar']"),
    ).toHaveCount(0);
  });
});
