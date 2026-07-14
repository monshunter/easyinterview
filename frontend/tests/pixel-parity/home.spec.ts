import { expect, test } from "@playwright/test";

import {
  configureDeterministicPage,
  expectFullPagePixelParity,
  expectSurfaceParity,
  settleVisualSurface,
  surfaceSnapshot,
} from "./report-parity-helpers";

/**
 * Phase 6.1 — Home screen DOM anchor and layout parity.
 *
 * Truth source: ui-design/src/screen-home.jsx::HomeScreen,
 * docs/spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-
 * parse/plan.md §4 Phase 6.
 *
 * Covers desktop (1440x900) and mobile (390x844) projects:
 * - DOM anchors (hero, paste-only textarea, resume picker, out-of-scope source/aux-card negatives)
 * - Bounding box stays in viewport, no overlap
 * - default (ocean)/light -> dark -> customAccent theme switching
 * - Mobile: textarea card not overflowing
 */

interface Rect {
  left: number;
  top: number;
  right: number;
  bottom: number;
  width: number;
  height: number;
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

test.describe("home screen DOM anchor parity", () => {
  test("home route renders hero, textarea, and aux card testids", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-hero-label']");

    await expect(page.locator("[data-testid='home-hero-label']")).toHaveCount(1);
    await expect(page.locator("[data-testid='home-hero-title']")).toHaveCount(1);
    await expect(page.locator("[data-testid='home-hero-sub']")).toHaveCount(0);
    await expect(page.locator("[data-testid='home-source-layout']")).toHaveCount(
      0,
    );
    await expect(page.locator("[data-testid='home-jd-paste-panel']")).toHaveCount(
      0,
    );
    await expect(
      page.locator("[data-testid='home-upload-source-panel']"),
    ).toHaveCount(0);
    await expect(page.locator("[data-testid='home-jd-input-card']")).toHaveCount(
      1,
    );
    await expect(
      page.locator("[data-testid='home-jd-source-controls']"),
    ).toHaveCount(0);
    await expect(page.locator("[data-testid='home-upload-trigger']")).toHaveCount(
      0,
    );
    await expect(page.locator("[data-testid='home-url-trigger']")).toHaveCount(
      0,
    );
    await expect(page.locator("[data-testid='home-jd-textarea']")).toHaveCount(
      1,
    );
    await expect(page.locator("[data-testid='home-jd-submit']")).toHaveCount(1);
    await expect(page.locator("[data-testid='home-jd-submit']")).toContainText(
      /立即面试|Start interview now/,
    );
    await expect(page.locator("[data-testid='home-resume-select']")).toHaveCount(
      1,
    );
    await expect(page.locator("[data-testid='home-resume-row']")).toHaveCount(1);
    await expect(page.locator("[data-testid='home-submit-row']")).toHaveCount(1);
    await expect(page.locator("[data-testid='home-resume-select']")).toHaveJSProperty(
      "tagName",
      "SELECT",
    );
    // product-scope D-17/D-22 keep these aux cards outside the current Home UI.
    await expect(page.locator("[data-testid='home-aux-jobpicks']")).toHaveCount(
      0,
    );
    await expect(page.locator("[data-testid='home-aux-debrief']")).toHaveCount(0);
  });

  test("home import layout keeps paste-only input and submit below resume", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-jd-input-card']");

    const placement = await page.evaluate(() => {
      const inputCard = document.querySelector(
        "[data-testid='home-jd-input-card']",
      ) as HTMLElement | null;
      const submitButton = document.querySelector(
        "[data-testid='home-jd-submit']",
      ) as HTMLElement | null;
      const resumeRow = document.querySelector(
        "[data-testid='home-resume-row']",
      ) as HTMLElement | null;
      const submitRow = document.querySelector(
        "[data-testid='home-submit-row']",
      ) as HTMLElement | null;
      const resumeSelect = document.querySelector(
        "[data-testid='home-resume-select']",
      ) as HTMLElement | null;
      const createCta = document.querySelector(
        "[data-testid='home-resume-create']",
      ) as HTMLElement | null;

      if (
        !inputCard ||
        !submitButton ||
        !resumeRow ||
        !submitRow ||
        !resumeSelect ||
        !createCta
      ) {
        throw new Error("missing Home import layout anchor");
      }

      const selectRect = resumeSelect.getBoundingClientRect();
      const createRect = createCta.getBoundingClientRect();
      return {
        viewportWidth: window.innerWidth,
        submitOutsideInput: !inputCard.contains(submitButton),
        submitAfterResume: Boolean(
          resumeRow.compareDocumentPosition(submitRow) &
            Node.DOCUMENT_POSITION_FOLLOWING,
        ),
        selectWidth: selectRect.width,
        createTopDelta: Math.abs(selectRect.top - createRect.top),
      };
    });

    expect(placement.submitOutsideInput).toBe(true);
    expect(placement.submitAfterResume).toBe(true);
    expect(placement.selectWidth).toBeLessThanOrEqual(362);
    if (placement.viewportWidth >= 900) {
      expect(placement.createTopDelta).toBeLessThanOrEqual(2);
    }
  });

  test("home textarea card stays inside viewport (desktop)", async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-jd-textarea']");

    const viewport = page.viewportSize();
    expect(viewport).toBeTruthy();

    const textareaRect = await rectOf(
      page,
      "[data-testid='home-jd-textarea']",
    );
    expect(textareaRect.top).toBeGreaterThanOrEqual(0);
    expect(textareaRect.left).toBeGreaterThanOrEqual(0);
    expect(textareaRect.right).toBeLessThanOrEqual(viewport!.width + 1);
  });

  test("home out-of-scope aux cards stay absent", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-jd-textarea']");

    await expect(
      page.locator("[data-testid='home-aux-jobpicks']"),
    ).toHaveCount(0);
    await expect(
      page.locator("[data-testid='home-aux-debrief']"),
    ).toHaveCount(0);
  });

  test("dark mode toggle changes computed background color", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='topbar-dark-toggle']");

    const lightBg = await page.evaluate(
      () => getComputedStyle(document.body).backgroundColor,
    );

    await page.click("[data-testid='topbar-dark-toggle']");
    await page.waitForTimeout(300);

    const darkBg = await page.evaluate(
      () => getComputedStyle(document.body).backgroundColor,
    );

    expect(lightBg).not.toBe(darkBg);
  });

  test("paste-only Home matches the UI truth and captures desktop/mobile evidence", async ({
    page,
    context,
  }, testInfo) => {
    const prototype = await context.newPage();
    await Promise.all([
      configureDeterministicPage(page, "zh"),
      configureDeterministicPage(prototype, "zh"),
    ]);

    // Signed-out formal Home intentionally has no business Resume rows. Mirror
    // that data state in the golden preview while leaving its source layout and
    // styling untouched, so the parity gate compares like-for-like UI states.
    await prototype.route("**/ui-design/src/screen-workspace.jsx*", async (route) => {
      const response = await route.fetch();
      const source = await response.text();
      await route.fulfill({
        response,
        contentType: "application/javascript; charset=utf-8",
        body: `${source}\nwindow.getWorkspaceResumeOptions = () => [];`,
      });
    });

    await Promise.all([
      page.goto("/"),
      prototype.goto("/ui-design/#route=home&lang=zh"),
    ]);
    await Promise.all([
      page.locator("[data-testid='home-jd-textarea']").waitFor(),
      prototype.locator("[data-testid='home-jd-textarea']").waitFor(),
    ]);
    await Promise.all([settleVisualSurface(page), settleVisualSurface(prototype)]);

    for (const surface of [page, prototype]) {
      await expect(surface.locator("[data-testid='home-jd-input-card']")).toHaveCount(1);
      await expect(surface.locator("[data-testid='home-jd-textarea']")).toHaveCount(1);
      await expect(surface.locator("[data-testid='home-resume-select']")).toHaveCount(1);
      await expect(surface.locator("[data-testid='home-submit-row']")).toHaveCount(1);
      await expect(surface.locator("[data-testid='home-jd-source-controls']")).toHaveCount(0);
      await expect(surface.locator("[data-testid='home-upload-trigger']")).toHaveCount(0);
      await expect(surface.locator("[data-testid='home-url-trigger']")).toHaveCount(0);
    }

    const surfaces = [
      {
        label: "JD input card",
        formal: "[data-testid='home-jd-input-card']",
        prototype: "[data-testid='home-jd-input-card']",
        properties: ["background-color", "border-top-width", "border-top-color", "border-radius", "padding"],
      },
      {
        label: "JD textarea",
        formal: "[data-testid='home-jd-textarea']",
        prototype: "[data-testid='home-jd-textarea']",
        properties: ["width", "min-height", "font-size", "line-height", "color", "background-color"],
      },
      {
        label: "Resume row",
        formal: "[data-testid='home-resume-row']",
        prototype: "[data-testid='home-resume-row']",
        properties: ["display", "align-items", "gap", "flex-wrap"],
      },
      {
        label: "Resume select",
        formal: "[data-testid='home-resume-select']",
        prototype: "[data-testid='home-resume-select']",
        properties: ["width", "max-width", "min-height", "font-size", "padding", "border-radius"],
      },
      {
        label: "Submit row",
        formal: "[data-testid='home-submit-row']",
        prototype: "[data-testid='home-submit-row']",
        properties: ["display", "margin-top"],
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

    const formalScreenshotPath = testInfo.outputPath(`home-formal-${testInfo.project.name}.png`);
    const prototypeScreenshotPath = testInfo.outputPath(`home-prototype-${testInfo.project.name}.png`);
    const formalViewportScreenshotPath = testInfo.outputPath(
      `home-formal-viewport-${testInfo.project.name}.png`,
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
      testInfo.attach(`home-formal-${testInfo.project.name}`, {
        path: formalScreenshotPath,
        contentType: "image/png",
      }),
      testInfo.attach(`home-prototype-${testInfo.project.name}`, {
        path: prototypeScreenshotPath,
        contentType: "image/png",
      }),
      testInfo.attach(`home-formal-viewport-${testInfo.project.name}`, {
        path: formalViewportScreenshotPath,
        contentType: "image/png",
      }),
    ]);
    expect(formalScreenshot.length).toBeGreaterThan(10_000);
    expect(prototypeScreenshot.length).toBeGreaterThan(10_000);
    expect(formalViewportScreenshot.length).toBeGreaterThan(10_000);
    const changedRatio = await expectFullPagePixelParity(
      page,
      prototype,
      testInfo,
      `home-paste-only-${testInfo.project.name}`,
    );
    console.log(
      `PIXEL_PARITY home paste-only browser gate project=${testInfo.project.name} viewport=${viewport!.width}x${viewport!.height} formalScreenshotBytes=${formalScreenshot.length} prototypeScreenshotBytes=${prototypeScreenshot.length} changedRatio=${changedRatio.toFixed(6)}`,
    );
    await prototype.close();
  });
});
