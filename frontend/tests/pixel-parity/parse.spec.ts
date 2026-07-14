import { expect, test } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import {
  configureDeterministicPage,
  expectFullPagePixelParity,
  expectPixelParity,
  expectSurfaceParity,
  normalizedText,
  pauseDeterministicClock,
  settleVisualSurface,
  surfaceSnapshot,
} from "./report-parity-helpers";

/**
 * Phase 6.2 — Parse screen DOM anchor and loading state parity.
 *
 * Truth source: ui-design/src/screens-p0-complete.jsx::ParseScreen,
 * docs/spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-
 * parse/plan.md §4 Phase 6.
 *
 * The parse screen requires a targetJobId param to load. In Playwright
 * we can only test the initial loading state when navigated to via the
 * home import flow (paste JD -> submit). Without mock transport, the
 * import call will fail; the test asserts the DOM anchors that are
 * reachable in the SPA flow.
 *
 * Full e2e with fixture-backed transport is deferred to the scenario
 * gate. These fixture-backed checks are UI parity tests, not E2E.
 */

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

async function mockParseReadyApis(
  page: import("@playwright/test").Page,
  targetAnalysisStatus: "ready" | "processing" = "ready",
): Promise<void> {
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
    if (method === "GET" && /^\/targets\/[^/]+$/.test(path)) {
      if (targetAnalysisStatus === "ready") {
        await fulfillFixture(route, "openapi/fixtures/TargetJobs/getTargetJob.json");
      } else {
        const response = fixtureResponse(
          "openapi/fixtures/TargetJobs/getTargetJob.json",
        );
        await route.fulfill({
          status: response.status,
          headers: {
            "content-type": "application/json; charset=utf-8",
            ...(response.headers ?? {}),
          },
          body: JSON.stringify({
            ...(response.body as Record<string, unknown>),
            analysisStatus: targetAnalysisStatus,
          }),
        });
      }
      return;
    }
    if (method === "GET" && path === "/resumes") {
      await fulfillFixture(route, "openapi/fixtures/Resumes/listResumes.json");
      return;
    }
    await route.fulfill({
      status: 404,
      headers: { "content-type": "application/json; charset=utf-8" },
      body: JSON.stringify({ error: { code: "NOT_FOUND", message: `No fixture for ${path}` } }),
    });
  });
}

async function mockParseConfirmApis(
  page: import("@playwright/test").Page,
  onUpdateTargetJob: (request: import("@playwright/test").Request) => Promise<void>,
): Promise<void> {
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
    if (path === "/targets") {
      await fulfillFixture(route, "openapi/fixtures/TargetJobs/listTargetJobs.json");
      return;
    }
    if (method === "GET" && /^\/targets\/[^/]+$/.test(path)) {
      await fulfillFixture(route, "openapi/fixtures/TargetJobs/getTargetJob.json");
      return;
    }
    if (method === "GET" && path === "/resumes") {
      await fulfillFixture(route, "openapi/fixtures/Resumes/listResumes.json");
      return;
    }
    if (method === "PATCH" && path.startsWith("/targets/")) {
      await onUpdateTargetJob(route.request());
      await fulfillFixture(route, "openapi/fixtures/TargetJobs/updateTargetJob.json");
      return;
    }
    if (method === "POST" && path === "/practice/plans") {
      await fulfillFixture(route, "openapi/fixtures/PracticePlans/createPracticePlan.json");
      return;
    }
    if (method === "POST" && path === "/practice/sessions") {
      await fulfillFixture(route, "openapi/fixtures/PracticeSessions/startPracticeSession.json");
      return;
    }
    if (method === "GET" && path.startsWith("/practice/sessions/")) {
      await fulfillFixture(route, "openapi/fixtures/PracticeSessions/getPracticeSession.json");
      return;
    }
    if (path.startsWith("/resumes/")) {
      await fulfillFixture(route, "openapi/fixtures/Resumes/getResume.json");
      return;
    }
    if (path.startsWith("/practice/plans/")) {
      await fulfillFixture(route, "openapi/fixtures/PracticePlans/getPracticePlan.json");
      return;
    }
    await route.fulfill({
      status: 404,
      headers: { "content-type": "application/json; charset=utf-8" },
      body: JSON.stringify({ error: { code: "NOT_FOUND", message: `No fixture for ${path}` } }),
    });
  });
}

async function freezeVisualAnimations(page: import("@playwright/test").Page): Promise<void> {
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

test.describe("parse screen DOM anchor parity", () => {
  test("home screen renders parse entry points (textarea + submit)", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-jd-textarea']");

    await expect(page.locator("[data-testid='home-jd-textarea']")).toBeEnabled();
    await expect(page.locator("[data-testid='home-jd-submit']")).toBeVisible();

    // Submit should be disabled when textarea is empty
    await expect(page.locator("[data-testid='home-jd-submit']")).toBeDisabled();
  });

  test("home jd textarea accepts input and submit enables", async ({
    page,
  }) => {
    await mockParseReadyApis(page);
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-jd-textarea']");

    await page.fill(
      "[data-testid='home-jd-textarea']",
      "Senior Frontend Engineer needed at Acme Corp",
    );

    await page.waitForSelector(
      "[data-testid='home-resume-option-01918fa0-0000-7000-8000-000000001000']",
      { state: "attached" },
    );
    await page.selectOption(
      "[data-testid='home-resume-select']",
      "01918fa0-0000-7000-8000-000000001000",
    );

    // Submit requires both JD text and an explicit bound resume.
    await expect(page.locator("[data-testid='home-jd-submit']")).toBeEnabled();
  });

  test("processing target job response keeps the loading demo free of internal metadata", async ({
    page,
  }, testInfo) => {
    await mockParseReadyApis(page, "processing");

    await page.goto("/parse?targetJobId=01918fa0-0000-7000-8000-000000002000");
    await page.waitForSelector("[data-testid='parse-loading-step-0']");
    await expect(page.locator("[data-testid='route-parse']")).toHaveCount(1);
    await expect(page.locator("[data-testid='parse-loading-step-0']")).toBeVisible();
    await expect(page.locator("[data-testid='parse-loading-step-1']")).toBeVisible();
    await expect(page.locator("[data-testid='parse-loading-step-2']")).toBeVisible();
    await expect(page.locator("[data-testid='parse-loading-step-3']")).toBeVisible();
    await expect(page.locator("[data-testid='parse-loading-footer']")).toHaveCount(0);
    await expect(page.locator("[data-testid='parse-basics-title']")).toHaveCount(0);

    const loadingSurface = page.locator("[data-testid='route-parse']");
    const loadingDomAudit = await loadingSurface.evaluate((root) => {
      const elements = [root, ...root.querySelectorAll("*")];
      return {
        text: root.textContent ?? "",
        attributes: elements.flatMap((element) =>
          Array.from(element.attributes, ({ name, value }) => `${name}=${value}`),
        ),
      };
    });
    const loadingAccessibilityAudit = await loadingSurface.ariaSnapshot();
    const internalMetadata =
      /\b(model|provider|rubric|prompt|provenance|version|hash|typical|latency)\b|模型|供应商|评分规则|提示词|溯源|版本|哈希|典型耗时/iu;
    expect(loadingDomAudit.text).not.toMatch(internalMetadata);
    expect(loadingDomAudit.attributes.join("\n")).not.toMatch(internalMetadata);
    expect(loadingAccessibilityAudit).not.toMatch(internalMetadata);

    await freezeVisualAnimations(page);
    const viewport = page.viewportSize();
    expect(viewport).not.toBeNull();
    const screenshotPath = testInfo.outputPath(
      `parse-processing-formal-${testInfo.project.name}.png`,
    );
    const screenshot = await page
      .locator("[data-testid='route-parse']")
      .screenshot({ path: screenshotPath });
    await testInfo.attach("parse-processing-response-loading-demo", {
      path: screenshotPath,
      contentType: "image/png",
    });
    expect(screenshot.length).toBeGreaterThan(10_000);
    console.log(
      `PIXEL_PARITY processing-response loading browser gate project=${testInfo.project.name} viewport=${viewport!.width}x${viewport!.height} screenshotBytes=${screenshot.length}`,
    );

    await page.waitForTimeout(1_000);
    await expect(page.locator("[data-testid='parse-loading-step-0']")).toBeVisible();
    await expect(page.locator("[data-testid='parse-basics-title']")).toHaveCount(0);

    await page.waitForTimeout(2_600);
    await expect(page.locator("[data-testid='parse-loading-step-0']")).toBeVisible();
    await expect(page.locator("[data-testid='parse-basics-title']")).toHaveCount(0);
  });

  test("parse loading matches the UI truth at desktop and mobile", async ({
    page,
    context,
  }, testInfo) => {
    const prototype = await context.newPage();
    await Promise.all([
      configureDeterministicPage(page, "zh"),
      configureDeterministicPage(prototype, "zh"),
      mockParseReadyApis(page, "processing"),
    ]);

    const formalRoot = "[data-testid='route-parse']";
    const prototypeRoot = "[data-screen-label='parse'] > div:last-child > .ei-fadein";
    // Hold both command-progress demos at their initial server-owned state.
    // The prototype's external font/script load can outlive its 3.2 s demo;
    // waiting for page.goto(load) before locating the root would observe the
    // subsequent Workspace route instead of Parse.
    // The formal Parse route is authenticated by mockParseReadyApis. Keep the
    // golden preview in the same auth state; on mobile the wider signed-in
    // user control changes the TopBar wrap height, which is part of the source
    // layout contract rather than Parse content.
    await Promise.all([
      pauseDeterministicClock(page),
      pauseDeterministicClock(prototype),
    ]);
    await Promise.all([
      prototype.goto(
        "/ui-design/#route=parse&lang=zh&signedIn=1&targetJobId=tj-import-pending",
      ),
      page.goto(
        "/parse?targetJobId=01918fa0-0000-7000-8000-000000002000",
      ),
    ]);
    await Promise.all([
      prototype.locator(prototypeRoot).waitFor(),
      page.locator("[data-testid='parse-loading-step-0']").waitFor(),
    ]);
    // A route navigation can preserve a page's prior scroll offset while its
    // sticky TopBar remains pinned. Reset both surfaces before comparing
    // viewport-relative boxes and full-page pixels.
    await Promise.all([
      page.evaluate(() => window.scrollTo(0, 0)),
      prototype.evaluate(() => window.scrollTo(0, 0)),
    ]);
    await Promise.all([
      freezeVisualAnimations(page),
      freezeVisualAnimations(prototype),
    ]);

    const internalMetadata =
      /\b(model|provider|rubric|prompt|provenance|version|hash|typical|latency)\b|模型|供应商|评分规则|提示词|溯源|版本|哈希|典型耗时/iu;
    for (const [surface, root] of [
      [page, formalRoot],
      [prototype, prototypeRoot],
    ] as const) {
      const audit = await surface.locator(root).evaluate((node) => {
        const elements = [node, ...node.querySelectorAll("*")];
        return {
          text: node.textContent ?? "",
          attributes: elements.flatMap((element) =>
            Array.from(element.attributes, ({ name, value }) => `${name}=${value}`),
          ),
        };
      });
      expect(audit.text).not.toMatch(internalMetadata);
      expect(audit.attributes.join("\n")).not.toMatch(internalMetadata);
      expect(await surface.locator(root).ariaSnapshot()).not.toMatch(internalMetadata);
    }

    const surfaces = [
      {
        label: "parse loading root",
        formal: formalRoot,
        prototype: prototypeRoot,
        properties: ["min-height", "display", "align-items", "justify-content", "padding"],
      },
      {
        label: "parse loading content",
        formal: `${formalRoot} > div`,
        prototype: `${prototypeRoot} > div`,
        properties: ["max-width", "width"],
      },
      {
        label: "parse loading label",
        formal: `${formalRoot} > div > .ei-label`,
        prototype: `${prototypeRoot} > div > .ei-label`,
        properties: ["color", "margin-bottom", "font-size", "font-family", "letter-spacing"],
      },
      {
        label: "parse loading title",
        formal: `${formalRoot} > div > .ei-serif`,
        prototype: `${prototypeRoot} > div > .ei-serif`,
        properties: ["font-size", "color", "letter-spacing", "line-height", "margin-bottom"],
      },
      {
        label: "parse loading steps",
        formal: `${formalRoot} > div > div:nth-child(3)`,
        prototype: `${prototypeRoot} > div > div:nth-child(3)`,
        properties: ["display", "flex-direction", "gap"],
      },
    ] as const;
    for (const surface of surfaces) {
      const [formal, golden] = await Promise.all([
        surfaceSnapshot(page, surface.formal, surface.properties),
        surfaceSnapshot(prototype, surface.prototype, surface.properties),
      ]);
      expectSurfaceParity(formal, golden, surface.label);
    }

    const viewport = page.viewportSize();
    expect(viewport).not.toBeNull();
    for (const surface of [page, prototype]) {
      expect(await surface.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
        viewport!.width,
      );
    }
    const formalScreenshotPath = testInfo.outputPath(
      `parse-loading-formal-parity-${testInfo.project.name}.png`,
    );
    const prototypeScreenshotPath = testInfo.outputPath(
      `parse-loading-prototype-parity-${testInfo.project.name}.png`,
    );
    const formalViewportScreenshotPath = testInfo.outputPath(
      `parse-loading-formal-viewport-${testInfo.project.name}.png`,
    );
    const [formalScreenshot, prototypeScreenshot] = await Promise.all([
      page.screenshot({ path: formalScreenshotPath, fullPage: true, animations: "disabled" }),
      prototype.screenshot({ path: prototypeScreenshotPath, fullPage: true, animations: "disabled" }),
    ]);
    await page.evaluate(() => window.scrollTo(0, 0));
    const formalViewportScreenshot = await page.screenshot({
      path: formalViewportScreenshotPath,
      fullPage: false,
      animations: "disabled",
    });
    await Promise.all([
      testInfo.attach(`parse-loading-formal-${testInfo.project.name}`, {
        path: formalScreenshotPath,
        contentType: "image/png",
      }),
      testInfo.attach(`parse-loading-prototype-${testInfo.project.name}`, {
        path: prototypeScreenshotPath,
        contentType: "image/png",
      }),
      testInfo.attach(`parse-loading-formal-viewport-${testInfo.project.name}`, {
        path: formalViewportScreenshotPath,
        contentType: "image/png",
      }),
    ]);
    expect(formalViewportScreenshot.length).toBeGreaterThan(10_000);
    const changedRatio = await expectFullPagePixelParity(
      page,
      prototype,
      testInfo,
      `parse-loading-${testInfo.project.name}`,
    );
    console.log(
      `PIXEL_PARITY parse loading browser gate project=${testInfo.project.name} viewport=${viewport!.width}x${viewport!.height} formalScreenshotBytes=${formalScreenshot.length} prototypeScreenshotBytes=${prototypeScreenshot.length} changedRatio=${changedRatio.toFixed(6)}`,
    );
    await prototype.close();
  });

  test("workspace detail exposes only direct start with bound resume context", async ({
    page,
  }, testInfo) => {
    const updateCalls: Array<{
      body: unknown;
      idempotencyKey: string | null;
    }> = [];
    await mockParseConfirmApis(page, async (request) => {
      updateCalls.push({
        body: request.postDataJSON(),
        idempotencyKey: request.headers()["idempotency-key"] ?? null,
      });
    });
    const resumeListRequests: string[] = [];
    page.on("request", (request) => {
      const url = new URL(request.url());
      if (request.method() === "GET" && url.pathname === "/api/v1/resumes") {
        resumeListRequests.push(request.url());
      }
    });

    await page.goto("/workspace?targetJobId=01918fa0-0000-7000-8000-000000002000");
    await page.waitForSelector("[data-testid='parse-basics-title']", { timeout: 5_000 });
    await expect(page.locator("[data-testid='parse-resume-bound-title']")).toContainText(
      "Resume saved with this interview plan",
    );
    await expect(page.locator("[data-testid='parse-resume-bound-meta']")).toContainText(
      "The saved binding is read-only",
    );
    expect(resumeListRequests).toHaveLength(0);
    await expect(page.locator("[data-testid='unified-plan-detail']")).toBeVisible();
    await expect(page.locator("[data-testid='parse-basics-title'] input")).toHaveCount(0);
    await expect(page.locator("[data-testid='parse-action-save-plan']")).toHaveCount(0);
    await expect(page.locator("[data-testid='parse-action-cancel']")).toHaveCount(0);
    await expect(page.locator("[data-testid='parse-action-reparse']")).toHaveCount(0);
    await expect(page.locator("[data-testid='parse-resume-picker']")).toHaveCount(0);
    await expect(page.locator("[data-testid='parse-resume-picker-toggle']")).toHaveCount(0);
    await expect(page.locator("[data-testid='parse-action-start-interview']")).toBeEnabled();
    const roundCards = page.locator("[data-testid='parse-rounds'] > [data-round-state]");
    await expect(roundCards).toHaveCount(3);
    const roundCount = await roundCards.count();
    expect(roundCount).toBeGreaterThanOrEqual(2);
    expect(roundCount).toBeLessThanOrEqual(5);
    await expect(roundCards.nth(0)).toContainText("Frontend architecture screen · 45m");
    await expect(roundCards.nth(0)).toContainText(
      "Probe scaling design systems across 10+ teams.",
    );
    await expect(roundCards.nth(1)).toContainText("Hiring manager impact interview · 50m");
    await expect(roundCards.nth(2)).toContainText("Collaboration and operating style · 40m");
    await expect(roundCards.nth(0)).toHaveAttribute("data-round-state", "done");
    await expect(roundCards.nth(0)).toContainText("Completed");
    await expect(roundCards.nth(1)).toHaveAttribute("data-round-state", "current");
    await expect(roundCards.nth(1)).toContainText("Up next");
    await expect(roundCards.nth(2)).toHaveAttribute("data-round-state", "pending");
    await expect(roundCards.nth(2)).toContainText("Not started");
    console.log(
      "PIXEL_PARITY structured-rounds browser gate count=3 first='Frontend architecture screen · 45m' second='Hiring manager impact interview · 50m' third='Collaboration and operating style · 40m'",
    );
    expect(updateCalls).toHaveLength(0);

    await freezeVisualAnimations(page);
    const screenshot = await page.locator("[data-testid='unified-plan-detail']").screenshot();
    await testInfo.attach("parse-readonly-detail-bound-resume", {
      body: screenshot,
      contentType: "image/png",
    });
    expect(screenshot.length).toBeGreaterThan(10_000);
    console.log(
      `PIXEL_PARITY workspace readonly-detail browser gate resumeId=01918fa0-0000-7000-8000-000000001000 resumeListRequests=0 parseAnimation=0 screenshotBytes=${screenshot.length}`,
    );
  });

  test("workspace detail round states match the UI truth at desktop and mobile", async ({
    page,
    context,
  }, testInfo) => {
    const prototype = await context.newPage();
    await Promise.all([
      configureDeterministicPage(page, "zh"),
      configureDeterministicPage(prototype, "zh"),
      mockParseConfirmApis(page, async () => undefined),
    ]);

    await Promise.all([
      page.goto(
        "/workspace?targetJobId=01918fa0-0000-7000-8000-000000002000",
      ),
      prototype.goto(
        "/ui-design/#route=workspace&lang=zh&signedIn=1&targetJobId=tj-1",
      ),
    ]);
    await Promise.all([
      page.locator("[data-testid='parse-rounds']").waitFor({ timeout: 8_000 }),
      prototype.locator("[data-testid='parse-rounds']").waitFor({ timeout: 8_000 }),
    ]);
    await Promise.all([settleVisualSurface(page), settleVisualSurface(prototype)]);
    await Promise.all([pauseDeterministicClock(page), pauseDeterministicClock(prototype)]);

    const formalSequence = await page
      .locator("[data-testid='parse-rounds'] > [data-round-state]")
      .evaluateAll((cards) => cards.map((card) => card.getAttribute("data-round-state")));
    expect(formalSequence).toEqual(["done", "current", "pending"]);
    const prototypeStates = new Set(
      await prototype
        .locator("[data-testid='parse-rounds'] > [data-round-state]")
        .evaluateAll((cards) => cards.map((card) => card.getAttribute("data-round-state"))),
    );
    expect(prototypeStates).toEqual(new Set(["done", "current", "pending"]));

    const cardProperties = [
      "background-color",
      "border-top-color",
      "border-top-style",
      "border-top-width",
      "border-radius",
      "padding-top",
      "padding-right",
      "padding-bottom",
      "padding-left",
      "position",
    ] as const;
    const labelProperties = [
      "font-family",
      "font-size",
      "color",
      "letter-spacing",
    ] as const;
    const backgrounds: string[] = [];
    const borders: string[] = [];
    for (const state of ["done", "current", "pending"] as const) {
      const selector = `:nth-match([data-testid='parse-rounds'] > [data-round-state='${state}'], 1)`;
      const [formalCard, prototypeCard] = await Promise.all([
        surfaceSnapshot(page, selector, cardProperties),
        surfaceSnapshot(prototype, selector, cardProperties),
      ]);
      expect(formalCard.style, `${state} card source style`).toEqual(
        prototypeCard.style,
      );
      backgrounds.push(formalCard.style["background-color"]);
      borders.push(formalCard.style["border-top-color"]);

      const [formalLabel, prototypeLabel] = await Promise.all([
        surfaceSnapshot(
          page,
          `${selector} [data-testid^='parse-round-state-']`,
          labelProperties,
        ),
        surfaceSnapshot(
          prototype,
          `${selector} [data-testid^='parse-round-state-']`,
          labelProperties,
        ),
      ]);
      expect(formalLabel.style, `${state} label source style`).toEqual(
        prototypeLabel.style,
      );
    }
    expect(new Set(backgrounds).size).toBe(3);
    expect(new Set(borders).size).toBe(3);

    const statePalette = async (surface: import("@playwright/test").Page) =>
      surface
        .locator("[data-testid='parse-rounds'] > [data-round-state]")
        .evaluateAll((cards) =>
          cards.map((card) => {
            const style = getComputedStyle(card);
            return `${style.backgroundColor}|${style.borderTopColor}`;
          }),
        );
    await Promise.all([
      page.getByTestId("topbar-dark-toggle").click(),
      prototype.locator(".ei-topbar-dark").click(),
    ]);
    await Promise.all([page.waitForTimeout(100), prototype.waitForTimeout(100)]);
    expect(new Set(await statePalette(page)).size).toBe(3);
    expect(new Set(await statePalette(prototype)).size).toBe(3);

    await page.getByTestId("topbar-theme-button").click();
    await prototype.locator("button[title='主题色']").click();
    await Promise.all([
      page.getByTestId("topbar-theme-custom-option").click(),
      prototype
        .locator(".ei-topbar-theme-menu button")
        .filter({ hasText: "自定义" })
        .click(),
    ]);
    await Promise.all([page.waitForTimeout(100), prototype.waitForTimeout(100)]);
    expect(new Set(await statePalette(page)).size).toBe(3);
    expect(new Set(await statePalette(prototype)).size).toBe(3);

    const viewport = page.viewportSize();
    expect(viewport).not.toBeNull();
    for (const surface of [page, prototype]) {
      const boxes = await surface
        .locator("[data-testid='parse-rounds'] > [data-round-state]")
        .evaluateAll((cards) =>
          cards.map((card) => {
            const box = card.getBoundingClientRect();
            return {
              left: box.left,
              right: box.right,
              width: box.width,
              height: box.height,
            };
          }),
        );
      expect(boxes.every((box) => box.left >= 0 && box.right <= viewport!.width)).toBe(true);
      expect(boxes.every((box) => box.width > 0 && box.height > 0)).toBe(true);
      expect(
        await surface.evaluate(() => document.documentElement.scrollWidth),
      ).toBeLessThanOrEqual(viewport!.width);
    }

    const formalScreenshotPath = testInfo.outputPath(
      `workspace-round-states-formal-${testInfo.project.name}.png`,
    );
    const prototypeScreenshotPath = testInfo.outputPath(
      `workspace-round-states-prototype-${testInfo.project.name}.png`,
    );
    const [formalScreenshot, prototypeScreenshot] = await Promise.all([
      page.locator("[data-testid='parse-rounds']").screenshot({
        path: formalScreenshotPath,
        animations: "disabled",
      }),
      prototype.locator("[data-testid='parse-rounds']").screenshot({
        path: prototypeScreenshotPath,
        animations: "disabled",
      }),
    ]);
    await Promise.all([
      testInfo.attach(`workspace-round-states-formal-${testInfo.project.name}`, {
        path: formalScreenshotPath,
        contentType: "image/png",
      }),
      testInfo.attach(`workspace-round-states-prototype-${testInfo.project.name}`, {
        path: prototypeScreenshotPath,
        contentType: "image/png",
      }),
    ]);
    expect(formalScreenshot.length).toBeGreaterThan(3_000);
    expect(prototypeScreenshot.length).toBeGreaterThan(3_000);
    console.log(
      `PIXEL_PARITY round-state project=${testInfo.project.name} sequence=${formalSequence.join(",")} distinctBackgrounds=3 distinctBorders=3 viewport=${viewport!.width}x${viewport!.height}`,
    );
    await prototype.close();
  });

  test("workspace plan-detail reports entry matches the UI truth and stays report-list-free", async ({
    page,
    context,
  }, testInfo) => {
    const prototype = await context.newPage();
    const reportRequests: string[] = [];
    page.on("request", (request) => {
      if (/\/api\/v1\/targets\/[^/]+\/reports(?:\?|$)/.test(request.url())) {
        reportRequests.push(request.url());
      }
    });
    await Promise.all([
      configureDeterministicPage(page, "zh"),
      configureDeterministicPage(prototype, "zh"),
      mockParseConfirmApis(page, async () => undefined),
    ]);

    const entry = "[data-testid='parse-reports-entry']";
    await Promise.all([
      page.goto(
        "/workspace?targetJobId=01918fa0-0000-7000-8000-000000002000&section=reports",
      ),
      prototype.goto(
        "/ui-design/#route=workspace&lang=zh&signedIn=1&targetJobId=tj-1&section=reports",
      ),
    ]);
    await Promise.all([
      page.locator(entry).waitFor({ timeout: 8_000 }),
      prototype.locator(entry).waitFor({ timeout: 8_000 }),
    ]);
    await Promise.all([settleVisualSurface(page), settleVisualSurface(prototype)]);
    await Promise.all([pauseDeterministicClock(page), pauseDeterministicClock(prototype)]);

    expect(await normalizedText(page, entry)).toBe(await normalizedText(prototype, entry));
    expect(reportRequests).toHaveLength(0);
    expect(new URL(page.url()).searchParams.has("section")).toBe(false);
    await expect(page.locator("[data-testid='parse-reports']")).toHaveCount(0);
    await expect(page.locator("[data-testid^='parse-report-round-']")).toHaveCount(0);
    await expect(prototype.locator("[data-testid='parse-reports']")).toHaveCount(0);
    await expect(prototype.locator("[data-testid^='parse-report-round-']")).toHaveCount(0);
    await expect(page.locator("[data-testid='topbar-nav-reports']")).toHaveCount(0);
    await expect(page.locator("[data-testid='topbar-primary-nav'] button")).toHaveCount(3);
    await expect(prototype.locator(".ei-topbar-nav button")).toHaveCount(3);
    await expect(prototype.locator(".ei-topbar-nav")).not.toContainText("面试报告");

    const surfaces = [
      {
        label: "plan-detail reports entry wrapper",
        selector: entry,
        properties: ["display", "flex-shrink", "font-family"],
      },
      {
        label: "plan-detail reports entry button",
        selector: `${entry} button`,
        properties: ["height", "padding", "font-size", "font-weight", "background-color", "color", "border", "border-radius", "font-family", "letter-spacing"],
      },
    ] as const;
    for (const surface of surfaces) {
      const [formal, golden] = await Promise.all([
        surfaceSnapshot(page, surface.selector, surface.properties),
        surfaceSnapshot(prototype, surface.selector, surface.properties),
      ]);
      expectSurfaceParity(formal, golden, surface.label);
    }

    const viewport = page.viewportSize();
    expect(viewport).not.toBeNull();
    for (const surface of [page, prototype]) {
      expect(
        await surface.evaluate(() => document.documentElement.scrollWidth),
      ).toBeLessThanOrEqual(viewport!.width);
    }
    const changedRatio = await expectPixelParity(
      page,
      prototype,
      entry,
      testInfo,
      `parse-reports-entry-${testInfo.project.name}`,
    );
    const formalScreenshotPath = testInfo.outputPath(
      `parse-reports-entry-formal-${testInfo.project.name}.png`,
    );
    const prototypeScreenshotPath = testInfo.outputPath(
      `parse-reports-entry-prototype-${testInfo.project.name}.png`,
    );
    await Promise.all([
      page.locator(entry).screenshot({ path: formalScreenshotPath, animations: "disabled" }),
      prototype.locator(entry).screenshot({ path: prototypeScreenshotPath, animations: "disabled" }),
    ]);
    await Promise.all([
      testInfo.attach(`parse-reports-entry-formal-${testInfo.project.name}`, {
        path: formalScreenshotPath,
        contentType: "image/png",
      }),
      testInfo.attach(`parse-reports-entry-prototype-${testInfo.project.name}`, {
        path: prototypeScreenshotPath,
        contentType: "image/png",
      }),
    ]);

    await page.locator(`${entry} button`).click();
    await page.waitForURL(/\/reports\?targetJobId=/);
    const destination = new URL(page.url());
    expect(destination.pathname).toBe("/reports");
    expect([...destination.searchParams.keys()]).toEqual(["targetJobId"]);
    expect(destination.searchParams.get("targetJobId")).toBe(
      "01918fa0-0000-7000-8000-000000002000",
    );
    console.log(
      `PIXEL_PARITY workspace plan-detail reports entry project=${testInfo.project.name} viewport=${viewport!.width}x${viewport!.height} reportListRequestsBeforeClick=0 topbarReportsEntry=0 embeddedReports=0 sectionReportsAccepted=false destination=/reports targetJobId=01918fa0-0000-7000-8000-000000002000 changedRatio=${changedRatio.toFixed(6)}`,
    );
    await prototype.close();
  });

  test("workspace start interview hands off directly to practice with bound resume", async ({
    page,
  }) => {
    const updateCalls: Array<{
      body: unknown;
      idempotencyKey: string | null;
    }> = [];
    await mockParseConfirmApis(page, async (request) => {
      updateCalls.push({
        body: request.postDataJSON(),
        idempotencyKey: request.headers()["idempotency-key"] ?? null,
      });
    });

    await page.goto("/workspace?targetJobId=01918fa0-0000-7000-8000-000000002000");
    await page.waitForSelector("[data-testid='parse-basics-title']", { timeout: 5_000 });
    await expect(
      page.locator("[data-testid='parse-action-start-interview']"),
    ).toBeEnabled();
    await expect(page.locator("[data-testid='parse-action-start-interview']")).toBeEnabled();
    await page.click("[data-testid='parse-action-start-interview']");

    await expect(page.locator("[data-testid='practice-screen']")).toBeVisible({
      timeout: 10_000,
    });
    expect(updateCalls).toHaveLength(0);

    const url = new URL(page.url());
    expect(url.pathname).toBe("/practice");
    expect(url.searchParams.get("targetJobId")).toBe(
      "01918fa0-0000-7000-8000-000000002000",
    );
    expect(url.searchParams.get("resumeId")).toBe(
      "01918fa0-0000-7000-8000-000000001000",
    );
    expect(url.search).not.toContain("resume-unbound");
    console.log(
      "PIXEL_PARITY workspace start-interview browser gate resumeId=01918fa0-0000-7000-8000-000000001000 route=practice noUpdateTargetJob=true",
    );
  });
});
