import { expect, test } from "@playwright/test";

import {
  configureDeterministicPage,
  expectFullPagePixelParity,
  expectPixelParity,
  expectSurfaceParity,
  injectPrototypeReportFixture,
  mockFormalReportApis,
  normalizedText,
  reportFixture,
  settleVisualSurface,
  surfaceSnapshot,
} from "./report-parity-helpers";

const ROOT = "[data-testid='generating-screen']";

test.use({ deviceScaleFactor: 1, locale: "zh-CN", timezoneId: "UTC" });

test.describe("honest generating source, geometry, and screenshot parity", () => {
  test("API generating state matches the UI truth at desktop and mobile", async ({
    page,
    context,
  }, testInfo) => {
    const prototype = await context.newPage();
    const fixture = reportFixture("generating").body;
    const reportId = String(fixture.id);

    await Promise.all([
      configureDeterministicPage(page, "zh"),
      configureDeterministicPage(prototype, "zh"),
    ]);
    await mockFormalReportApis(page, "generating");
    await injectPrototypeReportFixture(prototype, "generating", "reportGeneration");

    await Promise.all([
      page.goto(`/generating?reportId=${reportId}&sessionId=route-session&reportStatus=ready`),
      prototype.goto(`/ui-design/#route=generating&reportId=${reportId}&lang=zh&nochrome=1`),
    ]);
    await Promise.all([
      page.locator(`${ROOT}[data-report-status='generating']`).waitFor(),
      prototype.locator(`${ROOT}[data-report-status='generating']`).waitFor(),
    ]);
    await Promise.all([settleVisualSurface(page), settleVisualSurface(prototype)]);

    expect(new URL(page.url()).search).toBe(`?reportId=${reportId}`);
    expect(await normalizedText(page, ROOT)).toBe(await normalizedText(prototype, ROOT));
    for (const selector of [
      "[data-testid='generating-progress']",
      "[data-testid='generating-phase-list']",
      "[data-testid='generating-live-stream']",
      "[data-testid='generating-notify-cta']",
    ]) {
      await expect(page.locator(selector)).toHaveCount(0);
      await expect(prototype.locator(selector)).toHaveCount(0);
    }
    await expect(page.locator("body")).not.toContainText(/实时观察|好了通知我|会话记录|\d+%/);

    const surfaces = [
      {
        label: "generating root",
        selector: ROOT,
        properties: ["min-height", "display", "align-items", "justify-content", "padding-top", "padding-right", "padding-bottom", "padding-left", "background-color"],
      },
      {
        label: "generating eyebrow",
        selector: "[data-testid='generating-header-eyebrow']",
        properties: ["font-family", "font-size", "letter-spacing", "margin-bottom", "color"],
      },
      {
        label: "generating title",
        selector: "[data-testid='generating-header-title']",
        properties: ["font-size", "font-weight", "letter-spacing", "line-height", "margin-bottom", "color"],
      },
      {
        label: "generating subtitle",
        selector: "[data-testid='generating-header-subtitle']",
        properties: ["max-width", "font-size", "line-height", "color"],
      },
    ] as const;
    for (const surface of surfaces) {
      const [formal, golden] = await Promise.all([
        surfaceSnapshot(page, surface.selector, surface.properties),
        surfaceSnapshot(prototype, surface.selector, surface.properties),
      ]);
      expectSurfaceParity(formal, golden, surface.label);
    }

    if (testInfo.project.name === "mobile") {
      expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390);
    }
    await expectPixelParity(page, prototype, ROOT, testInfo, `generating-${testInfo.project.name}`);
    await expectFullPagePixelParity(
      page,
      prototype,
      testInfo,
      `generating-full-page-${testInfo.project.name}`,
    );
    await prototype.close();
  });

  test("queued and context-too-large states are projected from the API action matrix", async ({
    context,
  }, testInfo) => {
    const queuedPage = await context.newPage();
    const failedPage = await context.newPage();
    const failedPrototype = await context.newPage();
    const queued = reportFixture("queued").body;
    const failed = reportFixture("failed-context-too-large").body;
    await Promise.all([
      configureDeterministicPage(queuedPage, "zh"),
      configureDeterministicPage(failedPage, "zh"),
      configureDeterministicPage(failedPrototype, "zh"),
      mockFormalReportApis(queuedPage, "queued"),
      mockFormalReportApis(failedPage, "failed-context-too-large"),
    ]);
    await injectPrototypeReportFixture(
      failedPrototype,
      "failed-context-too-large",
      "reportGeneration",
    );

    await Promise.all([
      queuedPage.goto(`/generating?reportId=${String(queued.id)}`),
      failedPage.goto(`/generating?reportId=${String(failed.id)}`),
      failedPrototype.goto(`/ui-design/#route=generating&reportId=${String(failed.id)}&lang=zh&nochrome=1`),
    ]);
    await expect(queuedPage.locator(`${ROOT}[data-report-status='queued']`)).toBeVisible();
    await expect(queuedPage.locator("body")).not.toContainText(/\d+%|实时观察|好了通知我|会话记录/);

    const terminal = failedPage.locator("[data-testid='generating-error-state']");
    await expect(terminal).toHaveAttribute("data-error-kind", "contextTooLarge");
    await expect(failedPage.locator("[data-testid='generating-error-desc']")).toContainText("缩短输入");
    await expect(failedPage.locator("[data-testid='generating-error-retry']")).toHaveCount(0);
    await expect(failedPage.locator("[data-testid='generating-error-back-to-workspace']")).toBeVisible();
    await expect(failedPage.locator("body")).not.toContainText(/好了通知我|会话记录/);
    await failedPrototype.locator("[data-testid='generating-screen']").waitFor();
    await Promise.all([
      settleVisualSurface(failedPage),
      settleVisualSurface(failedPrototype),
    ]);
    expect(await normalizedText(failedPage, ROOT)).toBe(
      await normalizedText(failedPrototype, ROOT),
    );
    for (const surface of [
      {
        label: "context-too-large root",
        selector: ROOT,
        properties: ["min-height", "display", "align-items", "justify-content", "padding-top", "padding-right", "padding-bottom", "padding-left", "background-color"],
      },
      {
        label: "context-too-large title",
        selector: "[data-testid='generating-header-title']",
        properties: ["font-size", "font-weight", "letter-spacing", "line-height", "margin-bottom", "color"],
      },
      {
        label: "context-too-large action divider",
        selector: "[data-testid='generating-error-state'] > div:last-child",
        prototypeSelector: "[data-testid='generating-screen'] > div > div:last-child",
        properties: ["display", "gap", "margin-top", "padding-top", "border-top-width", "border-top-color"],
      },
    ] as const) {
      const [formal, golden] = await Promise.all([
        surfaceSnapshot(failedPage, surface.selector, surface.properties),
        surfaceSnapshot(
          failedPrototype,
          "prototypeSelector" in surface ? surface.prototypeSelector : surface.selector,
          surface.properties,
        ),
      ]);
      expectSurfaceParity(formal, golden, surface.label);
    }
    await expectPixelParity(
      failedPage,
      failedPrototype,
      ROOT,
      testInfo,
      `generating-context-too-large-${testInfo.project.name}`,
    );

    await Promise.all([
      queuedPage.close(),
      failedPage.close(),
      failedPrototype.close(),
    ]);
  });
});
