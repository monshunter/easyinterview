import { expect, test } from "@playwright/test";

/**
 * Phase 6.1 — JD Match three-tab DOM anchor parity.
 *
 * Truth source: ui-design/src/screen-jd-match.jsx,
 * docs/spec/frontend-home-job-picks-and-parse/plans/002-jd-match-recommendations/plan.md §3.5.
 *
 * Covers desktop (1440x900) and mobile (390x844) projects:
 * - Hero / profile chip / tabs DOM anchors
 * - Recommended tab list + JDDetail sticky
 * - Search tab natural-language search bar + four chip filters
 * - Watchlist tab list + market signals 4-card grid + refresh footer
 * - Negative: legacy plan-001 placeholder testid stays absent
 * - Negative: prototype numeric-index testids stay absent
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

async function gotoJdMatch(page: import("@playwright/test").Page) {
  await page.goto("/");
  await page.waitForSelector("[data-testid='home-aux-jobpicks']");
  const button = page.locator("[data-testid='home-aux-jobpicks'] button");
  if ((await button.count()) > 0) {
    await button.click();
  }
  await page.waitForSelector("[data-testid='route-jd_match']");
}

test.describe("jd_match three-tab DOM anchor parity", () => {
  test("Hero, profile chip and three tab labels are present after navigation", async ({
    page,
  }) => {
    await gotoJdMatch(page);

    await expect(page.locator("[data-testid='jdmatch-hero-label']"),)
      .toHaveCount(1);
    await expect(page.locator("[data-testid='jdmatch-hero-title']"),)
      .toHaveCount(1);
    await expect(page.locator("[data-testid='jdmatch-hero-sub']"),)
      .toHaveCount(1);

    await expect(page.locator("[data-testid='jdmatch-profile-chip']"))
      .toHaveCount(1);
    await expect(
      page.locator("[data-testid='jdmatch-profile-chip-avatar']"),
    ).toHaveCount(1);
    await expect(
      page.locator("[data-testid='jdmatch-profile-chip-skills']"),
    ).toHaveCount(1);
    await expect(
      page.locator("[data-testid='jdmatch-profile-chip-sources']"),
    ).toHaveCount(1);

    await expect(page.locator("[data-testid='jdmatch-tab-recommended']"))
      .toHaveCount(1);
    await expect(page.locator("[data-testid='jdmatch-tab-search']"))
      .toHaveCount(1);
    await expect(page.locator("[data-testid='jdmatch-tab-watchlist']"))
      .toHaveCount(1);

    await expect(page.locator("[data-testid='jdmatch-agent-status-badge']"))
      .toHaveCount(1);
  });

  test("Recommended tab is the default body and renders either detail-or-empty state", async ({
    page,
  }) => {
    await gotoJdMatch(page);
    await expect(page.locator("[data-testid='jdmatch-recommended-tab']"))
      .toHaveCount(1);
    // Without a backend wired into the dev server the recommendations may
    // arrive as an empty list / loading / error. We accept any of the
    // documented body states; Vitest specs already exercise the data path
    // with fixture-backed mock transport.
    const detail = page.locator("[data-testid='jdmatch-detail']");
    const empty = page.locator("[data-testid='jdmatch-recommended-empty']");
    const loading = page.locator(
      "[data-testid='jdmatch-recommended-loading']",
    );
    const error = page.locator("[data-testid='jdmatch-recommended-error']");
    const total =
      (await detail.count()) +
      (await empty.count()) +
      (await loading.count()) +
      (await error.count());
    expect(total).toBeGreaterThanOrEqual(1);

    // When JDDetail renders, the four action buttons must all be present.
    if ((await detail.count()) > 0) {
      await expect(
        page.locator("[data-testid='jdmatch-detail-action-confirm']"),
      ).toHaveCount(1);
      await expect(
        page.locator("[data-testid='jdmatch-detail-action-save']"),
      ).toHaveCount(1);
      await expect(
        page.locator("[data-testid='jdmatch-detail-action-source']"),
      ).toHaveCount(1);
      await expect(
        page.locator("[data-testid='jdmatch-detail-action-dismiss']"),
      ).toHaveCount(1);
    }
  });

  test("Switching to Search tab renders the natural-language search bar and four chip filters", async ({
    page,
  }) => {
    await gotoJdMatch(page);
    await page.locator("[data-testid='jdmatch-tab-search']").click();
    await expect(page.locator("[data-testid='jdmatch-search-tab']"))
      .toHaveCount(1);
    await expect(page.locator("[data-testid='jdmatch-search-input']"))
      .toHaveCount(1);
    await expect(page.locator("[data-testid='jdmatch-search-run']"))
      .toHaveCount(1);
    for (const k of ["all", "strong", "remote", "unseen"]) {
      await expect(
        page.locator(`[data-testid='jdmatch-search-filter-${k}']`),
      ).toHaveCount(1);
    }
  });

  test("Switching to Watchlist tab renders the watchlist + market signals + refresh footer", async ({
    page,
  }) => {
    await gotoJdMatch(page);
    await page.locator("[data-testid='jdmatch-tab-watchlist']").click();
    await expect(page.locator("[data-testid='jdmatch-watchlist-tab']"))
      .toHaveCount(1);
    await expect(
      page.locator("[data-testid='jdmatch-market-signals-grid']"),
    ).toHaveCount(1);
    await expect(
      page.locator("[data-testid='jdmatch-watchlist-refresh-footer']"),
    ).toHaveCount(1);
  });

  test("Hero block stays inside viewport on desktop", async ({ page }) => {
    await gotoJdMatch(page);
    const viewport = page.viewportSize();
    expect(viewport).toBeTruthy();
    const heroRect = await rectOf(
      page,
      "[data-testid='jdmatch-hero-title']",
    );
    expect(heroRect.left).toBeGreaterThanOrEqual(0);
    expect(heroRect.right).toBeLessThanOrEqual(viewport!.width + 1);
  });

  test("Negative — legacy plan-001 placeholder testid stays absent", async ({
    page,
  }) => {
    await gotoJdMatch(page);
    await expect(page.locator("[data-testid='jdmatch-placeholder']"))
      .toHaveCount(0);
    await expect(page.locator("[data-testid='jdmatch-placeholder-cta']"))
      .toHaveCount(0);
  });

  test("Negative — prototype numeric-index testids stay absent", async ({
    page,
  }) => {
    await gotoJdMatch(page);
    // The new convention uses UUID ids (jdmatch-card-${uuid},
    // jdmatch-watchlist-item-${uuid}); legacy prototype numeric-index
    // testids must never be re-introduced.
    await expect(page.locator("[data-testid='jdmatch-card-0']")).toHaveCount(0);
    await expect(
      page.locator("[data-testid='jdmatch-saved-search-0']"),
    ).toHaveCount(0);
    await expect(page.locator("[data-testid='jdmatch-watchlist-0']")).toHaveCount(
      0,
    );
    await expect(page.locator("[data-testid='jdmatch-search-bar']")).toHaveCount(
      0,
    );
  });
});
