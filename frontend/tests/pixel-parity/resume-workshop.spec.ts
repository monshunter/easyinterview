import { expect, test } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

/**
 * Phase 5.1-5.4 — Resume Workshop screen DOM anchor + computed style +
 * bounding box + screenshot smoke pixel parity.
 *
 * Truth source: ui-design/src/screen-resume-workshop.jsx and
 * docs/spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-
 * readonly/plan.md, D-20 flat resume asset model.
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
    if (/^\/resumes\/[^/]+$/.test(path)) {
      await fulfillFixture(route, "openapi/fixtures/Resumes/getResume.json");
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

async function computedStyleOf(
  page: import("@playwright/test").Page,
  selector: string,
  properties: string[],
): Promise<Record<string, string>> {
  return page.evaluate(({ selector, properties }) => {
    const el = document.querySelector(selector) as HTMLElement | null;
    if (!el) throw new Error(`selector not found: ${selector}`);
    const styles = window.getComputedStyle(el);
    return Object.fromEntries(
      properties.map((property) => [
        property,
        styles.getPropertyValue(property),
      ]),
    );
  }, { selector, properties });
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
  await page.waitForSelector("[data-testid='resume-workshop-table']");
}

async function goToDetail(
  page: import("@playwright/test").Page,
): Promise<void> {
  await mockResumeWorkshopApis(page);
  const resumeId = "01918fa0-0000-7000-8000-000000001000";
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
      params: {
        resumeId,
        tab: "rewrites",
        tailorRunId: "01918fa0-0000-7000-8000-000000009000",
      },
    },
  );
  await page.goto("/");
  await page.waitForSelector("[data-testid='resume-detail-crumb']");
}

test.describe("Resume Workshop list DOM anchors", () => {
  test("flat list table renders and stays inside the viewport", async ({ page }, testInfo) => {
    await goToList(page);
    await freezeAnimations(page);

    for (const anchor of [
      "resume-workshop-list",
      "resume-workshop-table",
      "resume-workshop-create",
      "resume-workshop-upload-cta",
    ]) {
      await expect(page.locator(`[data-testid='${anchor}']`)).toBeVisible();
    }
    await expect(
      page.locator("[data-testid^='resume-list-row-'][role='row']"),
    ).toHaveCount(2);

    const viewport = page.viewportSize();
    expect(viewport).not.toBeNull();
    const table = await rectOf(
      page,
      "[data-testid='resume-workshop-table']",
    );
    expect(table.width).toBeGreaterThan(0);
    expect(table.height).toBeGreaterThan(0);
    expect(table.left).toBeGreaterThanOrEqual(0);
    expect(table.right).toBeLessThanOrEqual(viewport!.width + 1);

    const shellStyle = await computedStyleOf(
      page,
      "[data-testid='resume-workshop-screen']",
      ["max-width", "padding-top", "padding-right"],
    );
    expect(shellStyle["max-width"]).toBe("1320px");
    if (viewport!.width > 900) {
      expect(shellStyle["padding-top"]).toBe("40px");
      expect(shellStyle["padding-right"]).toBe("48px");
    } else {
      expect(shellStyle["padding-top"]).toBe("28px");
      expect(shellStyle["padding-right"]).toBe("18px");
    }

    const tableStyle = await computedStyleOf(
      page,
      "[data-testid='resume-workshop-table']",
      ["border-radius", "border-top-width", "overflow"],
    );
    expect(tableStyle["border-radius"]).toBe("3px");
    expect(tableStyle["border-top-width"]).toBe("1px");
    expect(tableStyle.overflow).toBe("hidden");

    const screenshot = await page.screenshot();
    expect(screenshot.length).toBeGreaterThan(0);
    await testInfo.attach("resume-workshop-list", {
      body: screenshot,
      contentType: "image/png",
    });
  });

  test("flat rows expose real Open buttons and no non-current tree/view-switcher anchors", async ({ page }) => {
    await goToList(page);
    await freezeAnimations(page);
    const open = page.locator("[data-testid^='resume-list-open-']").first();
    await expect(open).toBeVisible();
    expect((await open.evaluate((node) => node.tagName)).toLowerCase()).toBe("button");
    for (const nonCurrent of [
      "resume-workshop-view-switcher-tree",
      "resume-workshop-view-switcher-flat",
      "resume-workshop-stats-originals",
      "resume-detail-branch-graph",
    ]) {
      await expect(page.locator(`[data-testid='${nonCurrent}']`)).toHaveCount(0);
    }
    await expect(page.locator("[data-testid^='resume-tree-row-']")).toHaveCount(0);
    await expect(page.locator("[data-testid^='resume-flat-row-']")).toHaveCount(0);
  });
});

test.describe("Resume Workshop detail DOM anchors", () => {
  test("crumb + read-only resume body render without secondary actions", async ({ page }, testInfo) => {
    await goToDetail(page);
    await freezeAnimations(page);

    await expect(
      page.locator("[data-testid='resume-detail-crumb']"),
    ).toBeVisible();
    await expect(page.locator("[data-testid='resume-detail-branch-graph']")).toHaveCount(0);
    for (const removed of [
      "resume-detail-tab-preview",
      "resume-detail-tab-rewrites",
      "resume-detail-tab-edit",
      "resume-detail-export-pdf",
      "resume-detail-copy-text",
      "resume-detail-view-original",
      "resume-rewrites-tab",
      "resume-edit-tab",
    ]) {
      await expect(page.locator(`[data-testid='${removed}']`)).toHaveCount(0);
    }

    const viewport = page.viewportSize();
    expect(viewport).not.toBeNull();
    const previewStyle = await computedStyleOf(
      page,
      "[data-testid='resume-detail-preview-content']",
      ["display", "justify-content", "align-items"],
    );
    expect(previewStyle["display"]).toBe("flex");
    expect(previewStyle["justify-content"]).toBe("center");
    expect(previewStyle["align-items"]).toBe("flex-start");

    const cardStyle = await computedStyleOf(
      page,
      ".ei-resume-detail-preview-card",
      ["width", "padding-top", "box-shadow", "font-family"],
    );
    expect(cardStyle["padding-top"]).toBe(
      viewport!.width > 700 ? "44px" : "32px",
    );
    expect(cardStyle["width"]).not.toBe("auto");
    expect(cardStyle["box-shadow"]).toContain("rgba(30, 22, 15, 0.1)");
    expect(cardStyle["font-family"].toLowerCase()).toContain("georgia");

    const screenshot = await page.screenshot();
    expect(screenshot.length).toBeGreaterThan(0);
    await testInfo.attach("resume-workshop-detail", {
      body: screenshot,
      contentType: "image/png",
    });
  });
});
