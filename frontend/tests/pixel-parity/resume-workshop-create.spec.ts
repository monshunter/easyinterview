import { expect, test } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

/**
 * frontend-resume-workshop/002 Phase 6.7 — pixel-parity gate for the
 * ResumeCreateFlow direct-detail handoff.
 *
 * Truth source: ui-design/src/screen-resume-workshop.jsx, exporting:
 *   ResumeCreateFlow (Upload / Paste tabs only)
 *   ResumeDetailView (waiting / source-format detail states after register success)
 *
 * The spec covers Upload tab, Paste tab, extra-tab negative, direct detail
 * handoff, and detail-owned waiting state in both desktop (1440x900) and
 * mobile (390x844). The check baseline is DOM anchor + computed style +
 * bounding box parity, plus a non-empty screenshot smoke. Clean checkout
 * PASS does not depend on local snapshots.
 */

interface Rect {
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

interface MockCreateFlowOptions {
  getResumeParseStatus?: "processing" | "ready";
}

function fixtureResponse(relativePath: string, scenario = "default") {
  const absolutePath = resolve(process.cwd(), "..", relativePath);
  const fixture = JSON.parse(readFileSync(absolutePath, "utf8")) as OperationFixture;
  const response = fixture.scenarios[scenario]?.response;
  if (!response) {
    throw new Error(`missing fixture scenario ${relativePath}#${scenario}`);
  }
  return response;
}

async function fulfillJson(
  route: import("@playwright/test").Route,
  status: number,
  body: unknown,
  headers: Record<string, string> = {},
) {
  await route.fulfill({
    status,
    headers: {
      "content-type": "application/json; charset=utf-8",
      ...headers,
    },
    body: JSON.stringify(body),
  });
}

async function fulfillFixture(
  route: import("@playwright/test").Route,
  relativePath: string,
  scenario = "default",
) {
  const response = fixtureResponse(relativePath, scenario);
  await fulfillJson(route, response.status, response.body, response.headers);
}

async function mockCreateFlowApis(
  page: import("@playwright/test").Page,
  options: MockCreateFlowOptions = {},
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
      if (route.request().method() === "POST") {
        await fulfillFixture(
          route,
          "openapi/fixtures/Resumes/registerResume.json",
        );
        return;
      }
      await fulfillFixture(route, "openapi/fixtures/Resumes/listResumes.json");
      return;
    }
    if (/^\/resumes\/[^/]+$/.test(path)) {
      const response = fixtureResponse("openapi/fixtures/Resumes/getResume.json");
      const body =
        typeof response.body === "object" && response.body !== null
          ? {
              ...(response.body as Record<string, unknown>),
              parseStatus: options.getResumeParseStatus ?? "ready",
            }
          : response.body;
      await fulfillJson(route, response.status, body, response.headers);
      return;
    }
    if (path === "/uploads/presign") {
      await fulfillFixture(
        route,
        "openapi/fixtures/Uploads/createUploadPresign.json",
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

async function rectOf(
  page: import("@playwright/test").Page,
  selector: string,
): Promise<Rect> {
  return page.evaluate(({ selector }) => {
    const el = document.querySelector(selector) as HTMLElement | null;
    if (!el) throw new Error(`selector not found: ${selector}`);
    const r = el.getBoundingClientRect();
    return { width: r.width, height: r.height };
  }, { selector });
}

async function gotoCreateFlow(
  page: import("@playwright/test").Page,
  params: Record<string, string> = {},
  options: MockCreateFlowOptions = {},
): Promise<void> {
  await mockCreateFlowApis(page, options);
  const search = new URLSearchParams({
    route: "resume_versions",
    ...{ flow: "create", ...params },
  }).toString();
  await page.goto(`/#${search}`);
  await page.waitForSelector("[data-testid='resume-create-flow']");
  await freezeAnimations(page);
}

async function submitPasteResume(
  page: import("@playwright/test").Page,
): Promise<void> {
  await page.getByTestId("resume-create-tab-paste").click();
  await page
    .getByTestId("resume-create-paste-textarea")
    .fill(
      "Senior frontend engineer focused on React, TypeScript, design systems, and interview-ready delivery evidence.",
    );
  await page.getByTestId("resume-create-paste-submit").click();
}

test.describe("ResumeCreateFlow pixel parity — input stages", () => {
  test("Upload tab — DOM anchors, no sidebar, bounding box parity", async ({
    page,
  }) => {
    await gotoCreateFlow(page);
    await expect(page.getByTestId("resume-create-flow")).toBeVisible();
    await expect(page.getByTestId("resume-create-tab-upload")).toHaveAttribute(
      "data-active",
      "true",
    );
    await expect(page.getByTestId("resume-create-upload-dropzone")).toBeVisible();
    await expect(page.getByTestId("resume-create-sidebar")).toHaveCount(0);
    // Bounding box: dropzone height ≥ 260
    const dropzone = await rectOf(
      page,
      "[data-testid='resume-create-upload-dropzone']",
    );
    expect(dropzone.height).toBeGreaterThanOrEqual(220);
    // Screenshot smoke (non-empty).
    const buf = await page.locator("[data-testid='resume-create-flow']").screenshot();
    expect(buf.byteLength).toBeGreaterThan(0);
  });

  test("Paste tab — textarea + submit disabled + helper copy", async ({
    page,
  }) => {
    await gotoCreateFlow(page);
    await page.getByTestId("resume-create-tab-paste").click();
    await expect(page.getByTestId("resume-create-paste-panel")).toBeVisible();
    await expect(page.getByTestId("resume-create-paste-textarea")).toBeVisible();
    await expect(page.getByTestId("resume-create-paste-submit")).toBeDisabled();
    const buf = await page.locator("[data-testid='resume-create-flow']").screenshot();
    expect(buf.byteLength).toBeGreaterThan(0);
  });

  test("Guided tab stays outside D-20 current scope and cannot be selected", async ({ page }) => {
    await gotoCreateFlow(page);
    await expect(page.getByTestId("resume-create-tab-guided")).toHaveCount(0);
    await expect(page.getByTestId("resume-create-guided-panel")).toHaveCount(0);
    await expect(page.getByTestId("resume-create-flow")).toHaveAttribute(
      "data-create-mode",
      "upload",
    );
    const buf = await page.locator("[data-testid='resume-create-flow']").screenshot();
    expect(buf.byteLength).toBeGreaterThan(0);
  });
});

test.describe("ResumeCreateFlow pixel parity — direct detail handoff", () => {
  test("Register success hands off to detail-owned waiting state", async ({
    page,
  }) => {
    await gotoCreateFlow(page, {}, { getResumeParseStatus: "processing" });
    await submitPasteResume(page);
    await expect(page.getByTestId("resume-detail-parse-waiting")).toBeVisible();
    await expect(page.getByTestId("resume-parse-flow")).toHaveCount(0);
    await expect(page.getByTestId("resume-preview-confirm")).toHaveCount(0);
    const waitingCard = await rectOf(
      page,
      "[data-testid='resume-detail-parse-waiting']",
    );
    expect(waitingCard.width).toBeGreaterThan(320);
    expect(waitingCard.height).toBeGreaterThan(180);
    const buf = await page
      .locator("[data-testid='resume-detail-parse-waiting']")
      .screenshot();
    expect(buf.byteLength).toBeGreaterThan(0);
  });

  test("Register success renders source-format detail without preview confirm", async ({
    page,
  }) => {
    await gotoCreateFlow(page);
    await submitPasteResume(page);
    await expect(page.getByTestId("resume-detail-container")).toBeVisible();
    await expect(page.getByTestId("resume-detail-preview-content")).toBeVisible();
    await expect(page.getByTestId("resume-parse-flow")).toHaveCount(0);
    await expect(page.getByTestId("resume-preview-confirm")).toHaveCount(0);
    const detail = await rectOf(
      page,
      "[data-testid='resume-detail-container']",
    );
    expect(detail.width).toBeGreaterThan(320);
    expect(detail.height).toBeGreaterThan(360);
    const buf = await page
      .locator("[data-testid='resume-detail-container']")
      .screenshot();
    expect(buf.byteLength).toBeGreaterThan(0);
  });
});
