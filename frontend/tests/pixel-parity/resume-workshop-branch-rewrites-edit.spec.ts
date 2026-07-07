import AxeBuilder from "@axe-core/playwright";
import { expect, test, type Page, type Route } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

/**
 * frontend-resume-workshop/003 D-20 remediation — pixel-parity and axe gate for:
 *   non-current BranchFlow route fallback
 *   ResumeRewritesTab
 *   ResumeEditTab
 *
 * Truth source: ui-design/src/screen-resume-workshop.jsx
 * (current flat Resume Workshop / Rewrites / Edit surfaces) and
 * docs/ui-design/resume-module.md.
 *
 * Clean checkout PASS does not depend on local screenshot snapshots. The gate
 * asserts DOM anchors, computed style, viewport-safe bounding boxes,
 * non-empty screenshot buffers, scoped axe checks, Rewrites/Edit header export,
 * and Preview-tab Copy Text behavior.
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

const RESUME_ID = "01918fa0-0000-7000-8000-000000001000";
const TAILOR_RUN_ID = "01918fa0-0000-7000-8000-000000009000";
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
    if (/^\/resumes\/[^/]+$/.test(path) && route.request().method() === "GET") {
      await fulfillFixture(route, "openapi/fixtures/Resumes/getResume.json");
      return;
    }
    if (/^\/resume\/tailor-runs\/[^/]+$/.test(path)) {
      await fulfillFixture(route, "openapi/fixtures/ResumeTailor/getResumeTailorRun.json");
      return;
    }
    if (/^\/resumes\/[^/]+\/exports$/.test(path)) {
      calls.exportHeaders.push(route.request().headers());
      await fulfillFixture(
        route,
        "openapi/fixtures/Resumes/exportResume.json",
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
  const search = new URLSearchParams({
    route: "resume_versions",
    ...params,
    pixelRoute: String(routeNonce),
  });
  await page.goto(`/#${search.toString()}`);
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
    .disableRules(["color-contrast", "aria-required-children", "aria-required-parent"])
    .analyze();
  expect(
    results.violations.map((violation) => ({
      id: violation.id,
      impact: violation.impact,
      nodes: violation.nodes.map((node) => node.target),
    })),
  ).toEqual([]);
}

async function assertNonCurrentBranchFlowFallback(page: Page): Promise<void> {
  await gotoHashRoute(page, {
    flow: "branch",
    branchOriginalId: RESUME_ID,
  });
  await page.waitForSelector("[data-testid='resume-workshop-list']");
  for (const id of [
    "resume-workshop-list",
    "resume-workshop-table",
    `resume-list-row-${RESUME_ID}`,
    `resume-list-open-${RESUME_ID}`,
    "resume-workshop-create",
  ]) {
    await expect(page.getByTestId(id), id).toBeVisible();
  }
  await expect(page.getByTestId("resume-branch-flow")).toHaveCount(0);
  const createStyle = await computedStyleOf(
    page,
    "[data-testid='resume-workshop-create']",
    ["border-radius", "font-family"],
  );
  expect(createStyle["border-radius"]).toBe("2px");
  expect(createStyle["font-family"]).toContain("Inter");
  await expectInViewport(page, "[data-testid='resume-workshop-list']");
  await expectScreenshotSmoke(page, "[data-testid='resume-workshop-list']");
  await expectAxeClean(page, "[data-testid='resume-workshop-list']");
}

async function assertRewritesTab(page: Page): Promise<void> {
  await gotoHashRoute(page, {
    resumeId: RESUME_ID,
    tab: "rewrites",
    tailorRunId: TAILOR_RUN_ID,
  });
  await page.waitForSelector("[data-testid='resume-rewrites-tab']");
  for (const id of [
    "resume-detail-header-actions",
    "resume-detail-export-pdf",
    "resume-rewrites-scope-banner",
    "resume-rewrites-bullet-list",
    "resume-rewrites-diff-card",
    "resume-rewrites-action-accept",
  ]) {
    await expect(page.getByTestId(id), id).toBeVisible();
  }
  await expect(page.getByTestId("resume-rewrites-rerun-tailor")).toHaveCount(0);
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
    resumeId: RESUME_ID,
    tab: "edit",
  });
  await page.waitForSelector("[data-testid='resume-edit-tab']");
  for (const id of [
    "resume-detail-header-actions",
    "resume-detail-export-pdf",
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
  expect(inputStyle["font-family"]).toContain("Noto Serif SC");
  await expectInViewport(page, "[data-testid='resume-edit-tab']");
  await expectScreenshotSmoke(page, "[data-testid='resume-edit-tab']");
  await expectAxeClean(page, "[data-testid='resume-workshop-detail']");
}

test.describe("resume workshop non-current branch / rewrites / edit pixel parity", () => {
  test("renders the D-20 flat workshop surfaces with DOM, style, bounding-box, screenshot, and axe coverage", async ({
    page,
  }) => {
    await mockResumeWorkshopApis(page, { exportHeaders: [] });
    await assertNonCurrentBranchFlowFallback(page);
    await assertRewritesTab(page);
    await assertEditTab(page);
  });

  test("keeps Export PDF available on Rewrites/Edit and Copy Text wired from Preview", async ({
    page,
  }) => {
    const calls: MockCalls = { exportHeaders: [] };
    await mockResumeWorkshopApis(page, calls);

    await gotoHashRoute(page, {
      resumeId: RESUME_ID,
      tab: "rewrites",
    });
    await page.waitForSelector("[data-testid='resume-rewrites-tab']");
    await expect(page.getByTestId("resume-detail-copy-text")).toHaveCount(0);

    await page.getByTestId("resume-detail-export-pdf").click();
    await expect.poll(() => calls.exportHeaders.length).toBe(1);
    expect(calls.exportHeaders[0]?.["idempotency-key"]).toMatch(/^v1\.\d+\./);

    await gotoHashRoute(page, {
      resumeId: RESUME_ID,
      tab: "preview",
    });
    await page.waitForSelector("[data-testid='resume-detail-preview-content']");
    await page.getByTestId("resume-detail-copy-text").click();
    await expect
      .poll(() => page.evaluate(() => window.__EI_COPIED_TEXT__ ?? ""))
      .toContain("Senior frontend engineer");

    await gotoHashRoute(page, {
      resumeId: RESUME_ID,
      tab: "edit",
    });
    await page.waitForSelector("[data-testid='resume-edit-tab']");
    await expect(page.getByTestId("resume-detail-copy-text")).toHaveCount(0);
    await expect(page.getByTestId("resume-detail-export-pdf")).toBeVisible();
  });
});
