import { expect, test } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

/**
 * Phase 5.1-5.4 — Resume Workshop screen DOM anchor + computed style +
 * bounding box + screenshot smoke pixel parity.
 *
 * Truth source: ui-design/src/screen-resume-workshop.jsx and
 * docs/spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-
 * readonly/plan.md, D-20 flat resume model.
 *
 * Covers desktop (1440x900) and mobile (390x844) projects through DOM anchors,
 * computed styles and non-empty screenshot smoke.
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

interface MockResumeWorkshopOptions {
  detailRenderer?: "markdown" | "pdf";
}

const RESUME_DETAIL_ID = "01918fa0-0000-7000-8000-000000001000";
const PDF_SOURCE_FIXTURE_BASE64 =
  "JVBERi0xLjMKJZOMi54gUmVwb3J0TGFiIEdlbmVyYXRlZCBQREYgZG9jdW1lbnQgKG9wZW5zb3VyY2UpCjEgMCBvYmoKPDwKL0YxIDIgMCBSIC9GMiAzIDAgUgo+PgplbmRvYmoKMiAwIG9iago8PAovQmFzZUZvbnQgL0hlbHZldGljYSAvRW5jb2RpbmcgL1dpbkFuc2lFbmNvZGluZyAvTmFtZSAvRjEgL1N1YnR5cGUgL1R5cGUxIC9UeXBlIC9Gb250Cj4+CmVuZG9iagozIDAgb2JqCjw8Ci9CYXNlRm9udCAvSGVsdmV0aWNhLUJvbGQgL0VuY29kaW5nIC9XaW5BbnNpRW5jb2RpbmcgL05hbWUgL0YyIC9TdWJ0eXBlIC9UeXBlMSAvVHlwZSAvRm9udAo+PgplbmRvYmoKNCAwIG9iago8PAovQ29udGVudHMgOSAwIFIgL01lZGlhQm94IFsgMCAwIDYxMiA3OTIgXSAvUGFyZW50IDggMCBSIC9SZXNvdXJjZXMgPDwKL0ZvbnQgMSAwIFIgL1Byb2NTZXQgWyAvUERGIC9UZXh0IC9JbWFnZUIgL0ltYWdlQyAvSW1hZ2VJIF0KPj4gL1JvdGF0ZSAwIC9UcmFucyA8PAoKPj4gCiAgL1R5cGUgL1BhZ2UKPj4KZW5kb2JqCjUgMCBvYmoKPDwKL0NvbnRlbnRzIDEwIDAgUiAvTWVkaWFCb3ggWyAwIDAgNjEyIDc5MiBdIC9QYXJlbnQgOCAwIFIgL1Jlc291cmNlcyA8PAovRm9udCAxIDAgUiAvUHJvY1NldCBbIC9QREYgL1RleHQgL0ltYWdlQiAvSW1hZ2VDIC9JbWFnZUkgXQo+PiAvUm90YXRlIDAgL1RyYW5zIDw8Cgo+PiAKICAvVHlwZSAvUGFnZQo+PgplbmRvYmoKNiAwIG9iago8PAovUGFnZU1vZGUgL1VzZU5vbmUgL1BhZ2VzIDggMCBSIC9UeXBlIC9DYXRhbG9nCj4+CmVuZG9iago3IDAgb2JqCjw8Ci9BdXRob3IgKGFub255bW91cykgL0NyZWF0aW9uRGF0ZSAoRDoyMDI2MDcwODAwMjYxOCswOCcwMCcpIC9DcmVhdG9yIChhbm9ueW1vdXMpIC9LZXl3b3JkcyAoKSAvTW9kRGF0ZSAoRDoyMDI2MDcwODAwMjYxOCswOCcwMCcpIC9Qcm9kdWNlciAoUmVwb3J0TGFiIFBERiBMaWJyYXJ5IC0gXChvcGVuc291cmNlXCkpIAogIC9TdWJqZWN0ICh1bnNwZWNpZmllZCkgL1RpdGxlICh1bnRpdGxlZCkgL1RyYXBwZWQgL0ZhbHNlCj4+CmVuZG9iago4IDAgb2JqCjw8Ci9Db3VudCAyIC9LaWRzIFsgNCAwIFIgNSAwIFIgXSAvVHlwZSAvUGFnZXMKPj4KZW5kb2JqCjkgMCBvYmoKPDwKL0ZpbHRlciBbIC9BU0NJSTg1RGVjb2RlIC9GbGF0ZURlY29kZSBdIC9MZW5ndGggMTgyCj4+CnN0cmVhbQpHYXJXMTVta0lfJjRRPVZgRiNYJ2cnTDJmQnFfXDdMdVxXMjouY2QnLSZKOygyY2MlbEE7OSg3IzJVXmFIImFwRSIkMjszSk5zOTdLQSpuajRKSjFtYlFVRStrK1xPQCdsNj1xcixsLS5YVmh1UzZiL1t0UzJvRkgwLVJOXFguU1ssNzNVRlA1LWdBUDhQZDBSckZlPDcwUmVdRyhbKlJLYCZZYks5Xkg0bzwxbk88PD5tZSJ+PmVuZHN0cmVhbQplbmRvYmoKMTAgMCBvYmoKPDwKL0ZpbHRlciBbIC9BU0NJSTg1RGVjb2RlIC9GbGF0ZURlY29kZSBdIC9MZW5ndGggMTgwCj4+CnN0cmVhbQpHYXJXcVltUz81JSolaE46W3NHMEUmTDRjXGVaYiROcSEuMUNGVV8hK19fSTdIdHNKYSs6L3FiX182W2opT2RrdUpWcF1xJyQsWi86Z2pfIykpWiZKXCJOXSM+O0c8RGR0Y1QvQmxlNlM8VzBGYyMpUVs7aSM/cTQtMnVAJFUtLSlGLllWYm44Z2VLZj5hKlcrPzBETCZmXGdkOkxvbjZzZUUiWkdATVhIVDRUU11zMjRzfj5lbmRzdHJlYW0KZW5kb2JqCnhyZWYKMCAxMQowMDAwMDAwMDAwIDY1NTM1IGYgCjAwMDAwMDAwNjEgMDAwMDAgbiAKMDAwMDAwMDEwMiAwMDAwMCBuIAowMDAwMDAwMjA5IDAwMDAwIG4gCjAwMDAwMDAzMjEgMDAwMDAgbiAKMDAwMDAwMDUxNCAwMDAwMCBuIAowMDAwMDAwNzA4IDAwMDAwIG4gCjAwMDAwMDA3NzYgMDAwMDAgbiAKMDAwMDAwMTAzNyAwMDAwMCBuIAowMDAwMDAxMTAyIDAwMDAwIG4gCjAwMDAwMDEzNzQgMDAwMDAgbiAKdHJhaWxlcgo8PAovSUQgCls8Nzg0NjVjYmQ5M2YwMjNiNmUxOTQwMTIwN2JjMmQ1NGU+PDc4NDY1Y2JkOTNmMDIzYjZlMTk0MDEyMDdiYzJkNTRlPl0KJSBSZXBvcnRMYWIgZ2VuZXJhdGVkIFBERiBkb2N1bWVudCAtLSBkaWdlc3QgKG9wZW5zb3VyY2UpCgovSW5mbyA3IDAgUgovUm9vdCA2IDAgUgovU2l6ZSAxMQo+PgpzdGFydHhyZWYKMTY0NQolJUVPRgo=";

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

async function fulfillPdfSource(route: import("@playwright/test").Route) {
  await route.fulfill({
    status: 200,
    headers: {
      "content-type": "application/pdf",
      "content-disposition": 'inline; filename="alice-example.pdf"',
    },
    body: Buffer.from(PDF_SOURCE_FIXTURE_BASE64, "base64"),
  });
}

async function mockResumeWorkshopApis(
  page: import("@playwright/test").Page,
  options: MockResumeWorkshopOptions = {},
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
    if (/^\/resumes\/[^/]+\/source$/.test(path)) {
      await fulfillPdfSource(route);
      return;
    }
    if (/^\/resumes\/[^/]+$/.test(path)) {
      const response = fixtureResponse("openapi/fixtures/Resumes/getResume.json");
      const body =
        options.detailRenderer === "pdf" &&
        typeof response.body === "object" &&
        response.body !== null
          ? {
              ...(response.body as Record<string, unknown>),
              id: RESUME_DETAIL_ID,
              title: "Alice Example Resume.pdf",
              displayName: "Alice Example — Senior Frontend Engineer",
              sourceType: "upload",
              fileObjectId: "01918fa0-0000-7000-8000-000000001100",
            }
          : response.body;
      await route.fulfill({
        status: response.status,
        headers: {
          "content-type": "application/json; charset=utf-8",
          ...(response.headers ?? {}),
        },
        body: JSON.stringify(body),
      });
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
  options: MockResumeWorkshopOptions = {},
): Promise<void> {
  await mockResumeWorkshopApis(page, options);
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
        resumeId: RESUME_DETAIL_ID,
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
    ]) {
      await expect(page.locator(`[data-testid='${anchor}']`)).toBeVisible();
    }
    await expect(page.locator("[data-testid='resume-workshop-upload-cta']")).toHaveCount(0);
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

  test("flat rows expose one-request Open buttons and no out-of-scope tree/view-switcher anchors", async ({ page }) => {
    const detailRequests: string[] = [];
    page.on("request", (request) => {
      const url = new URL(request.url());
      if (
        request.method() === "GET" &&
        url.pathname === `/api/v1/resumes/${RESUME_DETAIL_ID}`
      ) {
        detailRequests.push(request.url());
      }
    });
    await goToList(page);
    await freezeAnimations(page);
    const open = page.locator(
      `[data-testid='resume-list-open-${RESUME_DETAIL_ID}']`,
    );
    await expect(open).toBeVisible();
    expect((await open.evaluate((node) => node.tagName)).toLowerCase()).toBe("button");
    expect(detailRequests).toHaveLength(0);
    for (const outOfScope of [
      "resume-workshop-view-switcher-tree",
      "resume-workshop-view-switcher-flat",
      "resume-workshop-stats-originals",
      "resume-detail-branch-graph",
    ]) {
      await expect(page.locator(`[data-testid='${outOfScope}']`)).toHaveCount(0);
    }
    await expect(page.locator("[data-testid^='resume-tree-row-']")).toHaveCount(0);
    await expect(page.locator("[data-testid^='resume-flat-row-']")).toHaveCount(0);

    await open.click();
    await expect(page.locator("[data-testid='resume-detail-crumb']")).toBeVisible();
    expect(detailRequests).toHaveLength(1);
    console.log(
      "Resume Workshop browser transport PASS getResumeBeforeOpen=0 getResumeAfterOpen=1",
    );
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
    await expect(page.locator(".ei-resume-detail-preview-body")).toBeVisible();
    await expect(page.locator("[data-testid='resume-detail-markdown-page']")).toBeVisible();
    await expect(
      page.locator("[data-testid='resume-detail-markdown-page']"),
    ).not.toContainText("Alice Example — Senior Frontend Engineer");
    await expect(page.locator("[data-testid='resume-detail-pdf-preview']")).toHaveCount(0);

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
      ["width", "padding-top", "background-color", "box-shadow", "font-family"],
    );
    expect(cardStyle["padding-top"]).toBe(
      viewport!.width > 700 ? "28px" : "20px",
    );
    expect(cardStyle["background-color"]).toBe("rgb(246, 243, 238)");
    expect(cardStyle["width"]).not.toBe("auto");
    expect(cardStyle["box-shadow"]).toContain("rgba(30, 22, 15, 0.1)");
    expect(cardStyle["font-family"].toLowerCase()).toContain("georgia");

    const markdownPageStyle = await computedStyleOf(
      page,
      "[data-testid='resume-detail-markdown-page']",
      ["width", "padding-top", "background-color", "box-shadow"],
    );
    expect(markdownPageStyle["padding-top"]).toBe(
      viewport!.width > 700 ? "44px" : "32px",
    );
    expect(markdownPageStyle["background-color"]).toBe("rgb(255, 255, 255)");
    expect(markdownPageStyle["width"]).not.toBe("auto");
    expect(markdownPageStyle["box-shadow"]).toContain(
      "rgba(30, 22, 15, 0.08)",
    );

    const screenshot = await page.screenshot();
    expect(screenshot.length).toBeGreaterThan(0);
    await testInfo.attach("resume-workshop-detail", {
      body: screenshot,
      contentType: "image/png",
    });
  });

  test("upload-backed PDF detail uses a top-to-bottom PDF page stack", async ({ page }, testInfo) => {
    await goToDetail(page, { detailRenderer: "pdf" });
    await freezeAnimations(page);

    const preview = page.locator("[data-testid='resume-detail-pdf-preview-stack']");
    await expect(preview).toBeVisible();
    await expect(page.locator(".ei-resume-detail-preview-card")).not.toHaveClass(
      /ei-resume-detail-preview-card--pdf/,
    );
    await expect(preview).toHaveAttribute(
      "data-source-url",
      new RegExp(`/api/v1/resumes/${RESUME_DETAIL_ID}/source$`),
    );
    await expect(page.locator("object, iframe, embed")).toHaveCount(0);
    await expect(page.locator(".ei-resume-detail-preview-body")).toHaveCount(0);
    await expect(
      page.locator("[data-testid^='resume-detail-pdf-page-']"),
    ).toHaveCount(2);
    await expect(page.locator("[data-testid='resume-detail-pdf-page-1']")).toHaveAttribute(
      "data-render-state",
      "ready",
      { timeout: 10000 },
    );
    await expect(page.locator("[data-testid='resume-detail-pdf-page-2']")).toHaveAttribute(
      "data-render-state",
      "ready",
      { timeout: 10000 },
    );

    const previewBox = await rectOf(
      page,
      "[data-testid='resume-detail-pdf-preview-stack']",
    );
    expect(previewBox.width).toBeGreaterThan(280);
    expect(previewBox.height).toBeGreaterThanOrEqual(680);

    const firstPage = await rectOf(
      page,
      "[data-testid='resume-detail-pdf-page-1']",
    );
    expect(firstPage.width).toBeGreaterThan(280);
    expect(firstPage.height).toBeGreaterThan(360);

    const screenshot = await page
      .locator("[data-testid='resume-detail-preview-content']")
      .screenshot();
    expect(screenshot.length).toBeGreaterThan(0);
    await testInfo.attach("resume-workshop-detail-pdf-source", {
      body: screenshot,
      contentType: "image/png",
    });
  });
});
