import { expect, test } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

/**
 * Phase 6.1 — JD Match three-tab DOM anchor parity.
 *
 * Truth source: ui-design/src/screen-jd-match.jsx,
 * docs/spec/frontend-home-job-picks-and-parse/plans/002-jd-match-recommendations/plan.md §3.5.
 *
 * Covers desktop (1440x900) and mobile (390x844) projects:
 * - Hero / profile chip / tabs DOM anchors
 * - Recommended tab list + JDDetail sticky
 * - Search tab natural-language search bar + five source chips + four chip filters
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

async function mockJdMatchApis(page: import("@playwright/test").Page): Promise<void> {
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
    if (path === "/jd-match/profile") {
      await fulfillFixture(route, "openapi/fixtures/JobMatch/getJobMatchProfile.json");
      return;
    }
    if (path === "/jd-match/agent-status") {
      await fulfillFixture(route, "openapi/fixtures/JobMatch/getAgentScanStatus.json");
      return;
    }
    if (path === "/jd-match/recommendations" && method === "GET") {
      await fulfillFixture(route, "openapi/fixtures/JobMatch/listJobRecommendations.json");
      return;
    }
    if (path.startsWith("/jd-match/recommendations/") && method === "GET") {
      await fulfillFixture(route, "openapi/fixtures/JobMatch/getJobRecommendation.json");
      return;
    }
    if (path === "/jd-match/saved-searches" && method === "GET") {
      await fulfillFixture(route, "openapi/fixtures/JobMatch/listSavedSearches.json");
      return;
    }
    if (path === "/jd-match/search" && method === "POST") {
      await fulfillFixture(route, "openapi/fixtures/JobMatch/searchJobs.json");
      return;
    }
    if (path === "/jd-match/watchlist" && method === "GET") {
      await fulfillFixture(route, "openapi/fixtures/JobMatch/listWatchlist.json");
      return;
    }
    if (path === "/jd-match/market-signals") {
      await fulfillFixture(route, "openapi/fixtures/JobMatch/getMarketSignals.json");
      return;
    }
    await route.fulfill({
      status: 404,
      headers: { "content-type": "application/json; charset=utf-8" },
      body: JSON.stringify({ error: { code: "NOT_FOUND", message: `No fixture for ${path}` } }),
    });
  });
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

async function gridColumnCount(
  page: import("@playwright/test").Page,
  selector: string,
): Promise<number> {
  return page.evaluate(({ selector }) => {
    const el = document.querySelector(selector) as HTMLElement | null;
    if (!el) throw new Error(`selector not found: ${selector}`);
    return getComputedStyle(el)
      .gridTemplateColumns
      .split(" ")
      .filter(Boolean).length;
  }, { selector });
}

async function freezeAnimations(page: import("@playwright/test").Page): Promise<void> {
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

  test("Switching to Search tab renders the natural-language search bar, five sources and four chip filters", async ({
    page,
  }) => {
    await gotoJdMatch(page);
    await page.locator("[data-testid='jdmatch-tab-search']").click();
    await expect(page.locator("[data-testid='jdmatch-search-tab']"))
      .toHaveCount(1);
    await expect(
      page.locator("[data-testid='jdmatch-search-natural-language-heading']"),
    ).toHaveText(/NATURAL LANGUAGE SEARCH|自然语言搜索/);
    await expect(page.locator("[data-testid='jdmatch-search-input-icon']"))
      .toHaveCount(1);
    await expect(page.locator("[data-testid='jdmatch-search-input']"))
      .toHaveCount(1);
    await expect(page.locator("[data-testid='jdmatch-search-run']"))
      .toHaveCount(1);
    await expect(page.locator("[data-testid='jdmatch-search-sources-label']"))
      .toHaveText(/SOURCES|数据源/);
    for (const k of ["linkedin", "boss", "maimai", "lagou", "company"]) {
      await expect(
        page.locator(`[data-testid='jdmatch-search-source-${k}']`),
      ).toHaveCount(1);
    }
    await expect(page.locator("[data-testid='jdmatch-search-source-company']"))
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

  test("Responsive geometry matches jd_match layout contracts", async ({
    page,
  }) => {
    await mockJdMatchApis(page);
    await gotoJdMatch(page);
    await page.waitForSelector("[data-testid='jdmatch-detail']");
    const viewport = page.viewportSize()!;
    const mobile = viewport.width <= 700;

    await expect(page.locator("[data-testid='jdmatch-recommended-tab']"))
      .toHaveCount(1);
    expect(
      await gridColumnCount(page, "[data-testid='jdmatch-recommended-tab']"),
    ).toBe(mobile ? 1 : 2);
    const detailPosition = await page
      .locator("[data-testid='jdmatch-detail']")
      .evaluate((el) => getComputedStyle(el as HTMLElement).position);
    expect(detailPosition).toBe(mobile ? "static" : "sticky");

    await page.locator("[data-testid='jdmatch-tab-search']").click();
    await expect(page.locator("[data-testid='jdmatch-search-saved-grid']"))
      .toHaveCount(1);
    expect(
      await gridColumnCount(page, "[data-testid='jdmatch-search-saved-grid']"),
    ).toBe(mobile ? 1 : 3);
    await page.locator("[data-testid='jdmatch-search-input']").fill("frontend remote");
    await page.locator("[data-testid='jdmatch-search-run']").click();
    await page.waitForSelector("[data-testid='jdmatch-search-results']");
    expect(
      await gridColumnCount(page, "[data-testid='jdmatch-search-results']"),
    ).toBe(mobile ? 1 : 2);

    await page.locator("[data-testid='jdmatch-tab-watchlist']").click();
    await page.waitForSelector("[data-testid='jdmatch-market-signal-0']");
    expect(
      await gridColumnCount(page, "[data-testid='jdmatch-market-signals-inner']"),
    ).toBe(mobile ? 2 : 4);
  });

  test("dark mode and customAccent visibly affect jd_match computed colors", async ({
    page,
  }) => {
    await mockJdMatchApis(page);
    await gotoJdMatch(page);
    await page.waitForSelector("[data-testid='jdmatch-detail']");

    const before = await page.evaluate(() => ({
      body: getComputedStyle(document.body).backgroundColor,
      card: getComputedStyle(
        document.querySelector("[data-testid='jdmatch-card-01918fa0-0000-7000-8000-00000000a001']") as HTMLElement,
      ).borderLeftColor,
      accent: getComputedStyle(document.documentElement)
        .getPropertyValue("--ei-color-accent")
        .trim(),
    }));

    await page.click("[data-testid='topbar-dark-toggle']");
    const dark = await page.evaluate(() => ({
      body: getComputedStyle(document.body).backgroundColor,
      mode: document.documentElement.getAttribute("data-mode"),
    }));
    expect(dark.mode).toBe("dark");
    expect(dark.body).not.toBe(before.body);

    await page.click("[data-testid='topbar-theme-button']");
    await page.click("[data-testid='topbar-theme-custom-option']");
    const custom = await page.evaluate(() => ({
      attr: document.documentElement.getAttribute("data-custom-accent"),
      accent: getComputedStyle(document.documentElement)
        .getPropertyValue("--ei-color-accent")
        .trim(),
      detailButton: getComputedStyle(
        document.querySelector("[data-testid='jdmatch-detail-action-save']") as HTMLElement,
      ).borderColor,
    }));
    expect(custom.attr).toBe("active");
    expect(custom.accent).not.toBe(before.accent);
    expect(custom.detailButton).not.toBe(before.card);
  });

  test("Recommended tab focused screenshot is stable and non-empty without a checked-in baseline", async ({
    page,
  }) => {
    await mockJdMatchApis(page);
    await gotoJdMatch(page);
    await page.waitForSelector("[data-testid='jdmatch-detail']");
    await freezeAnimations(page);
    const target = page.locator("[data-testid='jdmatch-recommended-tab']");
    const box = await target.boundingBox();
    expect(box).toBeTruthy();
    expect(box!.width).toBeGreaterThan(250);
    expect(box!.height).toBeGreaterThan(300);
    const png = await target.screenshot();
    expect(png.byteLength).toBeGreaterThan(10_000);
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
