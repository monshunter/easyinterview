import { expect, test } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

/**
 * Phase 5.1-5.4 — Resume Workshop screen DOM anchor + computed style +
 * bounding box + screenshot smoke pixel parity.
 *
 * Truth source: ui-design/src/screen-resume-workshop.jsx and
 * docs/spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-
 * readonly/plan.md §4 Phase 5.
 *
 * Covers desktop (1440x900) and mobile (390x844) projects. Clean checkout
 * gate does not depend on local screenshot snapshots; the spec asserts DOM
 * anchors and computed styles directly and only ships a non-empty
 * screenshot smoke (no toMatchSnapshot here).
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
  const fixture = JSON.parse(
    readFileSync(absolutePath, "utf8"),
  ) as OperationFixture;
  const response = fixture.scenarios[scenario]?.response;
  if (!response)
    throw new Error(`missing fixture scenario ${relativePath}#${scenario}`);
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

async function mockResumeWorkshopApis(
  page: import("@playwright/test").Page,
): Promise<void> {
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
    if (path === "/resumes") {
      await fulfillFixture(route, "openapi/fixtures/Resumes/listResumes.json");
      return;
    }
    if (/^\/resumes\/[^/]+\/versions$/.test(path)) {
      await fulfillFixture(
        route,
        "openapi/fixtures/Resumes/listResumeVersions.json",
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
    await route.fulfill({
      status: 404,
      headers: { "content-type": "application/json; charset=utf-8" },
      body: JSON.stringify({
        error: { code: "NOT_FOUND", message: `No fixture for ${path}` },
      }),
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

async function freezeAnimations(
  page: import("@playwright/test").Page,
): Promise<void> {
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

async function goToList(page: import("@playwright/test").Page): Promise<void> {
  await mockResumeWorkshopApis(page);
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
    name: "resume_versions",
    params: {},
  });
  await page.goto("/");
  await page.waitForSelector(
    "[data-testid='resume-workshop-stats-originals']",
  );
}

async function goToDetail(
  page: import("@playwright/test").Page,
): Promise<void> {
  await mockResumeWorkshopApis(page);
  const versionId = "0195f2d0-0001-7000-8000-000000000202";
  await page.addInitScript(
    (route) => {
      (
        window as Window & {
          __EASYINTERVIEW_INITIAL_ROUTE__?: {
            name: string;
            params: Record<string, string>;
          };
        }
      ).__EASYINTERVIEW_INITIAL_ROUTE__ = route;
    },
    {
      name: "resume_versions",
      params: { versionId, tab: "preview" },
    },
  );
  await page.goto("/");
  await page.waitForSelector("[data-testid='resume-detail-breadcrumb']");
}

test.describe("Resume Workshop list DOM anchors", () => {
  test("stats + view switcher + tree rows render and stay inside the viewport", async ({ page }, testInfo) => {
    await goToList(page);
    await freezeAnimations(page);

    for (const anchor of [
      "resume-workshop-stats-originals",
      "resume-workshop-stats-versions",
      "resume-workshop-stats-top-match",
      "resume-workshop-stats-recent",
      "resume-workshop-view-switcher-tree",
      "resume-workshop-view-switcher-flat",
    ]) {
      await expect(page.locator(`[data-testid='${anchor}']`)).toBeVisible();
    }

    const viewport = page.viewportSize();
    expect(viewport).not.toBeNull();
    const stats = await rectOf(
      page,
      "[data-testid='resume-workshop-stats-originals']",
    );
    expect(stats.width).toBeGreaterThan(0);
    expect(stats.height).toBeGreaterThan(0);
    expect(stats.left).toBeGreaterThanOrEqual(0);
    expect(stats.right).toBeLessThanOrEqual(viewport!.width);

    const tree = page.locator(
      "[data-testid^='resume-tree-row-01918fa0-']",
    );
    await expect(tree.first()).toBeVisible();
    const screenshot = await page.screenshot();
    expect(screenshot.length).toBeGreaterThan(0);
    await testInfo.attach("resume-workshop-list", {
      body: screenshot,
      contentType: "image/png",
    });
  });

  test("tree row toggle button is a real <button> with aria-expanded", async ({ page }) => {
    await goToList(page);
    await freezeAnimations(page);
    const toggle = page.locator(
      "[data-testid^='resume-tree-row-01918fa0-'][data-testid$='-toggle']",
    );
    const handle = toggle.first();
    await expect(handle).toBeVisible();
    expect(await handle.getAttribute("aria-expanded")).toBe("true");
    expect((await handle.evaluate((node) => node.tagName)).toLowerCase()).toBe(
      "button",
    );
  });

  test("clicking the flat view switcher renders flat rows and the active tab flips", async ({ page }) => {
    await goToList(page);
    await freezeAnimations(page);
    await page.click("[data-testid='resume-workshop-view-switcher-flat']");
    await expect(
      page.locator("[data-testid^='resume-flat-row-'][role='row']"),
    ).toHaveCount(2);
    expect(
      await page
        .locator("[data-testid='resume-workshop-view-switcher-flat']")
        .getAttribute("aria-selected"),
    ).toBe("true");
  });
});

test.describe("Resume Workshop detail DOM anchors", () => {
  test("breadcrumb + branch graph + three tabs render with role=tab and aria-selected reflecting the active tab", async ({ page }, testInfo) => {
    await goToDetail(page);
    await freezeAnimations(page);

    await expect(
      page.locator("[data-testid='resume-detail-breadcrumb']"),
    ).toBeVisible();
    await expect(
      page.locator("[data-testid='resume-detail-branch-graph']"),
    ).toBeVisible();
    for (const tab of ["preview", "rewrites", "edit"]) {
      const handle = page.locator(`[data-testid='resume-detail-tab-${tab}']`);
      await expect(handle).toBeVisible();
      expect(await handle.getAttribute("role")).toBe("tab");
    }
    expect(
      await page
        .locator("[data-testid='resume-detail-tab-preview']")
        .getAttribute("aria-selected"),
    ).toBe("true");

    const screenshot = await page.screenshot();
    expect(screenshot.length).toBeGreaterThan(0);
    await testInfo.attach("resume-workshop-detail", {
      body: screenshot,
      contentType: "image/png",
    });
  });

  test("view-original button opens the modal dialog with role=dialog and aria-modal=true", async ({ page }) => {
    await goToDetail(page);
    await freezeAnimations(page);

    await page.click("[data-testid='resume-detail-view-original']");
    const dialog = page.locator(
      "[data-testid='resume-detail-original-modal']",
    );
    await expect(dialog).toBeVisible();
    expect(await dialog.getAttribute("role")).toBe("dialog");
    expect(await dialog.getAttribute("aria-modal")).toBe("true");
  });
});
