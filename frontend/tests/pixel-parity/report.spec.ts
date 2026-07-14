import { expect, test, type Page } from "@playwright/test";

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

const ROOT = "[data-testid='report-dashboard']";

test.use({ deviceScaleFactor: 1, locale: "zh-CN", timezoneId: "UTC" });

const DIRECT_CASES = [
  { label: "zh needs-practice", scenario: "prototype-baseline", lang: "zh" },
  { label: "en 24-word boundary", scenario: "ready-well-prepared", lang: "en", boundary: { actionIndex: 0, unit: "words", limit: 24 } },
  { label: "zh 64-code-point boundary", scenario: "long-content", lang: "zh", boundary: { actionIndex: 1, unit: "codePoints", limit: 64 } },
] as const;

async function expectInternalReportLocatorsAbsent(
  surface: Page,
  reportSentinel: string,
  sessionSentinel: string,
) {
  const root = surface.locator(ROOT);
  await expect(root.locator("[data-testid='report-context-session']")).toHaveCount(0);
  await expect(
    root.locator("[data-testid='report-context-strip'] > [data-testid^='report-context-']"),
  ).toHaveCount(3);

  const domAudit = await root.evaluate((node) => {
    const elements = [node, ...node.querySelectorAll("*")];
    return {
      text: node.textContent ?? "",
      attributes: elements.flatMap((element) =>
        Array.from(element.attributes, ({ name, value }) => `${name}=${value}`),
      ),
    };
  });
  const accessibilityAudit = await root.ariaSnapshot();

  for (const sentinel of [reportSentinel, sessionSentinel]) {
    expect(domAudit.text).not.toContain(sentinel);
    expect(domAudit.attributes.join("\n")).not.toContain(sentinel);
    expect(accessibilityAudit).not.toContain(sentinel);
  }
}

test.describe("report source, geometry, and screenshot parity", () => {
  for (const parityCase of DIRECT_CASES) test(`${parityCase.label} direct semantic report matches the UI truth at desktop and mobile`, async ({
    page, context,
  }, testInfo) => {
    const prototype = await context.newPage();
    const fixture = reportFixture(parityCase.scenario).body;
    const reportId = String(fixture.id);
    const sessionId = String(fixture.sessionId);
    expect(sessionId).not.toBe(reportId);

    await Promise.all([
      configureDeterministicPage(page, parityCase.lang),
      configureDeterministicPage(prototype, parityCase.lang),
    ]);
    const mutableContextReads = await mockFormalReportApis(page, parityCase.scenario);
    await injectPrototypeReportFixture(prototype, parityCase.scenario, "report");

    await Promise.all([
      page.goto(`/report?reportId=${reportId}&sessionId=route-session&reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT`),
      prototype.goto(`/ui-design/#route=report&reportId=${reportId}&lang=${parityCase.lang}&signedIn=1`),
    ]);
    await Promise.all([
      page.locator(ROOT).waitFor(),
      prototype.locator(ROOT).waitFor(),
    ]);
    await Promise.all([settleVisualSurface(page), settleVisualSurface(prototype)]);

    expect(mutableContextReads).toEqual([]);
    expect(new URL(page.url()).search).toBe(`?reportId=${reportId}`);
    await Promise.all([
      expectInternalReportLocatorsAbsent(page, reportId, sessionId),
      expectInternalReportLocatorsAbsent(prototype, reportId, sessionId),
    ]);
    expect(await normalizedText(page, ROOT)).toBe(await normalizedText(prototype, ROOT));

    if ("boundary" in parityCase) {
      const actions = fixture.nextActions as Array<{ label: string }>;
      const boundaryLabel = actions[parityCase.boundary.actionIndex]!.label;
      const count = parityCase.boundary.unit === "words"
        ? boundaryLabel.trim().split(/\s+/u).length
        : [...boundaryLabel].length;
      expect(count).toBe(parityCase.boundary.limit);

      for (const surface of [page, prototype]) {
        const actionCard = surface.locator("[data-testid='report-actions'] > div");
        const actionLabel = surface.locator(".ei-report-action-label").nth(parityCase.boundary.actionIndex);
        await expect(actionLabel).toHaveText(boundaryLabel);
        await expect(actionLabel).toBeVisible();

        const [cardBox, labelBox, overflow] = await Promise.all([
          actionCard.boundingBox(),
          actionLabel.boundingBox(),
          actionLabel.evaluate((node) => {
            const element = node as HTMLElement;
            const style = getComputedStyle(element);
            return {
              clientWidth: element.clientWidth,
              scrollWidth: element.scrollWidth,
              overflow: style.overflow,
              overflowWrap: style.overflowWrap,
              textOverflow: style.textOverflow,
              whiteSpace: style.whiteSpace,
              wordBreak: style.wordBreak,
            };
          }),
        ]);
        expect(cardBox).not.toBeNull();
        expect(labelBox).not.toBeNull();
        expect(labelBox!.x).toBeGreaterThanOrEqual(cardBox!.x);
        expect(labelBox!.x + labelBox!.width).toBeLessThanOrEqual(cardBox!.x + cardBox!.width + 1);
        expect(labelBox!.height).toBeGreaterThan(20);
        expect(overflow.scrollWidth).toBeLessThanOrEqual(overflow.clientWidth + 1);
        expect(overflow).toMatchObject({
          overflow: "visible",
          overflowWrap: "anywhere",
          textOverflow: "clip",
          whiteSpace: "normal",
          wordBreak: "normal",
        });
        expect(await surface.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
          testInfo.project.name === "mobile" ? 390 : 1440,
        );
      }
    }

    const [formalTopBar, goldenTopBar, formalTopBarNav, goldenTopBarNav] = await Promise.all([
      surfaceSnapshot(
        page,
        "[data-testid='app-shell-topbar']",
        ["display", "flex-wrap", "gap", "padding", "overflow-x", "border-bottom-width", "border-bottom-color"],
      ),
      surfaceSnapshot(
        prototype,
        ".ei-shell-topbar",
        ["display", "flex-wrap", "gap", "padding", "overflow-x", "border-bottom-width", "border-bottom-color"],
      ),
      surfaceSnapshot(
        page,
        "[data-testid='topbar-primary-nav']",
        ["order", "width", "margin-left", "overflow-x", "gap"],
      ),
      surfaceSnapshot(
        prototype,
        ".ei-topbar-nav",
        ["order", "width", "margin-left", "overflow-x", "gap"],
      ),
    ]);
    expectSurfaceParity(formalTopBar, goldenTopBar, "shared TopBar viewport contract");
    expectSurfaceParity(formalTopBarNav, goldenTopBarNav, "shared TopBar nav contract");

    const surfaces: ReadonlyArray<{
      label: string;
      selector: string;
      properties: readonly string[];
    }> = [
      {
        label: "report header",
        selector: "[data-testid='report-header']",
        properties: ["display", "align-items", "gap", "margin-bottom"],
      },
      {
        label: "frozen context strip",
        selector: "[data-testid='report-context-strip']",
        properties: ["display", "grid-template-columns", "gap", "border-top-width", "border-top-color", "margin-bottom"],
      },
      {
        label: "summary metrics",
        selector: "[data-testid='report-summary-cards']",
        properties: ["display", "grid-template-columns", "gap", "margin-bottom"],
      },
      {
        label: "detail grid",
        selector: "[data-testid='report-detail-grid']",
        properties: ["display", "grid-template-columns", "gap"],
      },
      {
        label: "dimension card",
        selector: "[data-testid='report-dimensions'] > div",
        properties: ["padding", "border-top-width", "border-top-color", "border-radius", "background-color"],
      },
      {
        label: "dimension row wrapping",
        selector: "[data-testid='report-dimensions'] .ei-report-dimension-row:nth-child(2)",
        properties: ["display", "flex-wrap", "gap", "align-items", "justify-content", "padding", "border-bottom-width"],
      },
      {
        label: "dimension label readability",
        selector: "[data-testid='report-dimensions'] .ei-report-dimension-row:nth-child(2) > .ei-report-dimension-label",
        properties: ["flex", "min-width", "overflow-wrap", "word-break"],
      },
      {
        label: "dimension status wrapping",
        selector: "[data-testid='report-dimensions'] .ei-report-dimension-row:nth-child(2) > .ei-report-dimension-status",
        properties: ["flex", "max-width", "overflow-wrap", "word-break", "text-align"],
      },
      {
        label: "action row wrapping",
        selector: "[data-testid='report-actions'] .ei-report-action-row:nth-child(2)",
        properties: ["display", "min-width", "gap", "overflow-wrap", "word-break"],
      },
      {
        label: "action label wrapping",
        selector: "[data-testid='report-actions'] .ei-report-action-row:nth-child(2) .ei-report-action-label",
        properties: ["min-width", "overflow-wrap", "word-break", "white-space", "overflow", "text-overflow"],
      },
    ];

    for (const surface of surfaces) {
      const [formal, golden] = await Promise.all([
        surfaceSnapshot(page, surface.selector, surface.properties, ROOT),
        surfaceSnapshot(prototype, surface.selector, surface.properties, ROOT),
      ]);
      expectSurfaceParity(formal, golden, surface.label);
    }

    const formalCards = page.locator("[data-testid='report-detail-grid'] > *");
    const first = await formalCards.nth(0).boundingBox();
    const second = await formalCards.nth(1).boundingBox();
    expect(first).not.toBeNull();
    expect(second).not.toBeNull();
    if (testInfo.project.name === "mobile") {
      expect(Math.abs(first!.x - second!.x)).toBeLessThanOrEqual(1);
      expect(second!.y).toBeGreaterThan(first!.y + first!.height - 1);
      expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390);
      expect(await prototype.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390);
      const expectedDimensionCount = await page.locator(".ei-report-dimension-label").count();
      expect(expectedDimensionCount).toBeGreaterThan(0);
      for (const surface of [page, prototype]) {
        const labels = await surface.locator(".ei-report-dimension-label").all();
        const statuses = await surface.locator(".ei-report-dimension-status").all();
        expect(labels).toHaveLength(expectedDimensionCount);
        expect(statuses).toHaveLength(expectedDimensionCount);
        const card = await surface.locator("[data-testid='report-dimensions'] > div").boundingBox();
        expect(card).not.toBeNull();
        for (const label of labels) {
          const box = await label.boundingBox();
          expect(box).not.toBeNull();
          expect(box!.width).toBeGreaterThanOrEqual(120);
          expect(box!.height).toBeLessThanOrEqual(48);
          expect(box!.width / box!.height).toBeGreaterThanOrEqual(2);
          expect(box!.x + box!.width).toBeLessThanOrEqual(card!.x + card!.width + 1);
        }
        for (const status of statuses) {
          const box = await status.boundingBox();
          expect(box).not.toBeNull();
          expect(box!.width).toBeGreaterThanOrEqual(90);
          expect(box!.height).toBeLessThanOrEqual(48);
          expect(box!.x + box!.width).toBeLessThanOrEqual(card!.x + card!.width + 1);
        }
      }
    } else {
      expect(second!.x).toBeGreaterThan(first!.x + first!.width - 1);
    }

    // Keep the viewport-relative root gate separate from the internal surface
    // comparisons above. A shared shell offset must stay visible as a real
    // parity failure; it must not be normalized away by report-local geometry.
    const [formalRoot, goldenRoot] = await Promise.all([
      surfaceSnapshot(
        page,
        ROOT,
        ["max-width", "padding-top", "padding-right", "padding-bottom", "padding-left"],
      ),
      surfaceSnapshot(
        prototype,
        ROOT,
        ["max-width", "padding-top", "padding-right", "padding-bottom", "padding-left"],
      ),
    ]);
    expectSurfaceParity(formalRoot, goldenRoot, "report root absolute viewport geometry");

    await expectPixelParity(
      page,
      prototype,
      ROOT,
      testInfo,
      `report-${parityCase.lang}-${testInfo.project.name}`,
    );
    await expectFullPagePixelParity(
      page,
      prototype,
      testInfo,
      `report-full-page-${parityCase.lang}-${testInfo.project.name}`,
    );
    const formalScreenshotPath = testInfo.outputPath(
      `report-formal-${parityCase.scenario}-${testInfo.project.name}.png`,
    );
    const formalScreenshot = await page.screenshot({
      path: formalScreenshotPath,
      fullPage: true,
      animations: "disabled",
    });
    await testInfo.attach(
      `report-formal-${parityCase.scenario}-${testInfo.project.name}`,
      { path: formalScreenshotPath, contentType: "image/png" },
    );
    expect(formalScreenshot.length).toBeGreaterThan(10_000);
    await page.locator("[data-testid='report-back-button']").click();
    await expect.poll(() => {
      const url = new URL(page.url());
      return url.pathname + url.search;
    }).toBe(`/reports?targetJobId=${String(fixture.targetJobId)}`);
    await prototype.close();
  });

  test("REPORT_CONTEXT_TOO_LARGE is a localized back-only API terminal state", async ({ page }) => {
    const fixture = reportFixture("failed-context-too-large").body;
    await configureDeterministicPage(page, "zh");
    await mockFormalReportApis(page, "failed-context-too-large");
    await page.goto(`/report?reportId=${String(fixture.id)}&reportStatus=ready`);

    const state = page.locator("[data-testid='report-failure-state']");
    await expect(state).toBeVisible();
    await expect(page.locator("[data-testid='report-failure-desc']")).toContainText("缩短输入");
    await expect(page.locator("[data-testid='report-failure-retry-cta']")).toHaveCount(0);
    await expect(page.locator("[data-testid='report-failure-back-to-workspace']")).toBeVisible();
    await expect(page.locator("body")).not.toContainText("好了通知我");
    await expect(page.locator("body")).not.toContainText("会话记录");
    await page.locator("[data-testid='report-failure-back-to-workspace']").click();
    await expect.poll(() => {
      const url = new URL(page.url());
      return url.pathname + url.search;
    }).toBe(`/reports?targetJobId=${String(fixture.targetJobId)}`);
  });
});
