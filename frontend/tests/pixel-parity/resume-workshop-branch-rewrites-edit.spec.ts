import AxeBuilder from "@axe-core/playwright";
import { expect, test, type Page, type Route } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

/**
 * frontend-resume-workshop/003 Phase 7.4 — pixel-parity and axe gate for:
 *   ResumeBranchFlow
 *   ResumeRewritesTab
 *   ResumeEditTab
 *
 * Truth source: ui-design/src/screen-resume-workshop.jsx
 * (ResumeBranchFlow L1018-1195, ResumeRewritesTab L784-940,
 * ResumeEditTab L943-1012) and docs/ui-design/resume-module.md.
 *
 * Clean checkout PASS does not depend on local screenshot snapshots. The gate
 * asserts DOM anchors, computed style, viewport-safe bounding boxes,
 * non-empty screenshot buffers, scoped axe checks, and the Rewrites/Edit
 * header actions for Export PDF / Copy Text.
 */

declare global {
  interface Window {
    __EI_COPIED_TEXT__?: string;
  }
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

interface MockCalls {
  exportHeaders: Array<Record<string, string>>;
}

const RESUME_ASSET_ID = "01918fa0-0000-7000-8000-000000001000";
const TARGETED_VERSION_ID = "0195f2d0-0001-7000-8000-000000000202";
let routeNonce = 0;

function fixtureResponse(relativePath: string, scenario = "default") {
  const absolutePath = resolve(process.cwd(), "..", relativePath);
  const fixture = JSON.parse(readFileSync(absolutePath, "utf8")) as OperationFixture;
  const response = fixture.scenarios[scenario]?.response;
  if (!response) {
    throw new Error(`missing fixture scenario ${relativePath}#${scenario}`);
  }
  return response;
}

async function fulfillFixture(
  route: Route,
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

async function mockResumeWorkshopApis(page: Page, calls: MockCalls): Promise<void> {
  await page.addInitScript(() => {
    Object.defineProperty(navigator, "clipboard", {
      configurable: true,
      value: {
        writeText: async (text: string) => {
          window.__EI_COPIED_TEXT__ = String(text);
        },
      },
    });
  });

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
    if (path === "/resumes") {
      await fulfillFixture(route, "openapi/fixtures/Resumes/listResumes.json");
      return;
    }
    if (/^\/resumes\/[^/]+\/versions$/.test(path)) {
      await fulfillFixture(route, "openapi/fixtures/Resumes/listResumeVersions.json");
      return;
    }
    if (/^\/resume-versions\/[^/]+$/.test(path) && route.request().method() === "GET") {
      await fulfillFixture(
        route,
        "openapi/fixtures/Resumes/getResumeVersion.json",
        "targeted-with-suggestions",
      );
      return;
    }
    if (/^\/resume-versions\/[^/]+\/exports$/.test(path)) {
      calls.exportHeaders.push(route.request().headers());
      await fulfillFixture(
        route,
        "openapi/fixtures/Resumes/exportResumeVersion.json",
        "p0-501-not-available",
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

async function freezeAnimations(page: Page): Promise<void> {
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

async function gotoHashRoute(
  page: Page,
  params: Record<string, string>,
): Promise<void> {
  routeNonce += 1;
  const search = new URLSearchParams({ route: "resume_versions", ...params });
  await page.goto(`/?pixelRoute=${routeNonce}#${search.toString()}`);
  await freezeAnimations(page);
}

async function computedStyleOf(
  page: Page,
  selector: string,
  properties: string[],
): Promise<Record<string, string>> {
  return page.evaluate(({ selector, properties }) => {
    const el = document.querySelector(selector) as HTMLElement | null;
    if (!el) throw new Error(`selector not found: ${selector}`);
    const styles = window.getComputedStyle(el);
    return Object.fromEntries(
      properties.map((property) => [property, styles.getPropertyValue(property)]),
    );
  }, { selector, properties });
}

async function expectInViewport(page: Page, selector: string): Promise<void> {
  const viewport = page.viewportSize();
  expect(viewport).not.toBeNull();
  const box = await page.locator(selector).boundingBox();
  expect(box, `${selector} bounding box`).not.toBeNull();
  expect(box!.width, `${selector} width`).toBeGreaterThan(0);
  expect(box!.height, `${selector} height`).toBeGreaterThan(0);
  expect(box!.x, `${selector} left`).toBeGreaterThanOrEqual(-1);
  expect(box!.x + box!.width, `${selector} right`).toBeLessThanOrEqual(
    viewport!.width + 1,
  );
}

async function expectScreenshotSmoke(page: Page, selector: string): Promise<void> {
  const buf = await page.locator(selector).screenshot();
  expect(buf.byteLength, `${selector} screenshot bytes`).toBeGreaterThan(1000);
}

async function expectAxeClean(page: Page, selector: string): Promise<void> {
  const results = await new AxeBuilder({ page })
    .include(selector)
    .disableRules(["color-contrast"])
    .analyze();
  expect(
    results.violations.map((violation) => ({
      id: violation.id,
      impact: violation.impact,
      nodes: violation.nodes.map((node) => node.target),
    })),
  ).toEqual([]);
}

async function assertBranchFlow(page: Page): Promise<void> {
  await gotoHashRoute(page, {
    flow: "branch",
    branchOriginalId: RESUME_ASSET_ID,
  });
  await page.waitForSelector("[data-testid='resume-branch-flow-form']");
  for (const id of [
    "resume-branch-flow",
    "resume-branch-from-card",
    "resume-branch-field-name",
    "resume-branch-field-target",
    "resume-branch-focus-chip-platform",
    "resume-branch-seed-card-copy_master",
    "resume-branch-seed-card-blank",
    "resume-branch-seed-card-ai_select",
    "resume-branch-submit",
  ]) {
    await expect(page.getByTestId(id), id).toBeVisible();
  }
  await page.getByTestId("resume-branch-field-name").fill("Northstar Systems frontend target");
  await page.getByTestId("resume-branch-field-target").fill("Northstar · Frontend Platform");
  await expect(page.getByTestId("resume-branch-submit")).toBeEnabled();
  const submitStyle = await computedStyleOf(
    page,
    "[data-testid='resume-branch-submit']",
    ["border-radius", "font-family"],
  );
  expect(submitStyle["border-radius"]).toBe("2px");
  expect(submitStyle["font-family"]).toContain("Inter");
  await expectInViewport(page, "[data-testid='resume-branch-flow']");
  await expectScreenshotSmoke(page, "[data-testid='resume-branch-flow']");
  await expectAxeClean(page, "[data-testid='resume-branch-flow']");
}

async function assertRewritesTab(page: Page): Promise<void> {
  await gotoHashRoute(page, {
    versionId: TARGETED_VERSION_ID,
    tab: "rewrites",
  });
  await page.waitForSelector("[data-testid='resume-rewrites-tab']");
  for (const id of [
    "resume-detail-header-actions",
    "resume-detail-export-pdf",
    "resume-detail-copy-text",
    "resume-rewrites-scope-banner",
    "resume-rewrites-bullet-list",
    "resume-rewrites-diff-card",
    "resume-rewrites-action-reject",
    "resume-rewrites-action-edit",
    "resume-rewrites-action-accept",
    "resume-rewrites-rerun-tailor",
  ]) {
    await expect(page.getByTestId(id), id).toBeVisible();
  }
  await expect(page.getByTestId("resume-rewrites-tab")).toHaveAttribute(
    "data-bullet-count",
    "1",
  );
  const bannerStyle = await computedStyleOf(
    page,
    "[data-testid='resume-rewrites-scope-banner']",
    ["border-radius", "display", "background-color"],
  );
  expect(bannerStyle["border-radius"]).toBe("2px");
  expect(bannerStyle.display).toBe("flex");
  expect(bannerStyle["background-color"]).not.toBe("rgba(0, 0, 0, 0)");
  await expectInViewport(page, "[data-testid='resume-rewrites-tab']");
  await expectScreenshotSmoke(page, "[data-testid='resume-rewrites-tab']");
  await expectAxeClean(page, "[data-testid='resume-workshop-detail']");
}

async function assertEditTab(page: Page): Promise<void> {
  await gotoHashRoute(page, {
    versionId: TARGETED_VERSION_ID,
    tab: "edit",
  });
  await page.waitForSelector("[data-testid='resume-edit-tab']");
  for (const id of [
    "resume-detail-header-actions",
    "resume-detail-export-pdf",
    "resume-detail-copy-text",
    "resume-edit-scope-banner",
    "resume-edit-headline-input",
    "resume-edit-summary-textarea",
    "resume-edit-section-experience",
    "resume-edit-section-skills",
    "resume-edit-save-button",
  ]) {
    await expect(page.getByTestId(id), id).toBeVisible();
  }
  const inputStyle = await computedStyleOf(
    page,
    "[data-testid='resume-edit-headline-input']",
    ["border-radius", "font-family"],
  );
  expect(inputStyle["border-radius"]).toBe("2px");
  expect(inputStyle["font-family"]).toContain("Georgia");
  await expectInViewport(page, "[data-testid='resume-edit-tab']");
  await expectScreenshotSmoke(page, "[data-testid='resume-edit-tab']");
  await expectAxeClean(page, "[data-testid='resume-workshop-detail']");
}

test.describe("resume workshop branch / rewrites / edit pixel parity", () => {
  test("renders the three plan 003 screens with DOM, style, bounding-box, screenshot, and axe coverage", async ({
    page,
  }) => {
    await mockResumeWorkshopApis(page, { exportHeaders: [] });
    await assertBranchFlow(page);
    await assertRewritesTab(page);
    await assertEditTab(page);
  });

  test("keeps Export PDF and Copy Text usable from Rewrites and Edit tabs", async ({
    page,
  }) => {
    const calls: MockCalls = { exportHeaders: [] };
    await mockResumeWorkshopApis(page, calls);

    await gotoHashRoute(page, {
      versionId: TARGETED_VERSION_ID,
      tab: "rewrites",
    });
    await page.waitForSelector("[data-testid='resume-rewrites-tab']");
    await page.getByTestId("resume-detail-copy-text").click();
    await expect
      .poll(() => page.evaluate(() => window.__EI_COPIED_TEXT__ ?? ""))
      .toContain("Senior frontend engineer");

    await page.getByTestId("resume-detail-export-pdf").click();
    await expect.poll(() => calls.exportHeaders.length).toBe(1);
    expect(calls.exportHeaders[0]?.["idempotency-key"]).toMatch(/^v1\.\d+\./);

    await gotoHashRoute(page, {
      versionId: TARGETED_VERSION_ID,
      tab: "edit",
    });
    await page.waitForSelector("[data-testid='resume-edit-tab']");
    await expect(page.getByTestId("resume-detail-copy-text")).toBeVisible();
    await expect(page.getByTestId("resume-detail-export-pdf")).toBeVisible();
  });
});
