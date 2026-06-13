import { expect, test } from "@playwright/test";

/**
 * Phase 3 — Layout + bounding box parity.
 *
 * Truth source: docs/spec/frontend-shell/plans/003-ui-design-pixel-parity-
 * gate/plan.md §4 Phase 3.
 *
 * Asserts that the production frontend dist lays out the TopBar and the
 * auth shell within their intended viewport without overlap, on both the
 * desktop (1440×900) and mobile (390×844) Playwright projects.
 */

interface Rect {
  left: number;
  top: number;
  right: number;
  bottom: number;
  width: number;
  height: number;
}

async function rectOf(page: import("@playwright/test").Page, selector: string): Promise<Rect> {
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

function intersects(a: Rect, b: Rect): boolean {
  return (
    a.left < b.right - 0.5 &&
    a.right > b.left + 0.5 &&
    a.top < b.bottom - 0.5 &&
    a.bottom > b.top + 0.5
  );
}

test.describe("frontend dist layout + bounding box parity", () => {
  test("TopBar header fits inside the documented viewport", async ({
    page,
  }, testInfo) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='app-shell-topbar']");
    const topbarRect = await rectOf(page, "[data-testid='app-shell-topbar']");
    const viewport = page.viewportSize();
    expect(viewport).toBeTruthy();

    expect(topbarRect.top).toBeCloseTo(0, 0);
    expect(topbarRect.left).toBeCloseTo(0, 0);
    if (testInfo.project.name === "desktop") {
      // Desktop height matches ui-design TopBar literal (58px) with 1px tolerance.
      expect(Math.abs(topbarRect.height - 58)).toBeLessThanOrEqual(1);
    } else {
      // Mobile intentionally wraps controls/nav into two rows to avoid
      // horizontal overflow while preserving all primary entries.
      expect(topbarRect.height).toBeGreaterThanOrEqual(58);
      expect(topbarRect.height).toBeLessThanOrEqual(150);
    }
    // TopBar must span full viewport width (1440 desktop / 390 mobile).
    expect(Math.abs(topbarRect.right - viewport!.width)).toBeLessThanOrEqual(1);
    // Sanity: bottom of TopBar inside viewport.
    expect(topbarRect.bottom).toBeLessThan(viewport!.height);
    testInfo.attach("topbar-rect", {
      body: JSON.stringify(topbarRect),
      contentType: "application/json",
    });
  });

  test("four primary nav buttons stay in viewport with no pairwise overlap (D-17)", async ({
    page,
  }) => {
    await page.goto("/");
    const navRects = await page.evaluate(() => {
      const buttons = Array.from(
        document.querySelectorAll(
          "[data-testid='topbar-primary-nav'] button[data-testid^='topbar-nav-']",
        ),
      ) as HTMLElement[];
      return buttons.map((el) => {
        const r = el.getBoundingClientRect();
        return {
          testid: el.getAttribute("data-testid"),
          left: r.left,
          top: r.top,
          right: r.right,
          bottom: r.bottom,
          width: r.width,
          height: r.height,
        };
      });
    });
    expect(navRects.length).toBe(4);
    const viewport = page.viewportSize()!;
    for (const r of navRects) {
      expect(r.width, `${r.testid} width`).toBeGreaterThan(0);
      expect(r.height, `${r.testid} height`).toBeGreaterThan(0);
      expect(r.top).toBeGreaterThanOrEqual(0);
      expect(r.bottom).toBeLessThanOrEqual(viewport.height);
    }
    for (let i = 0; i < navRects.length; i++) {
      for (let j = i + 1; j < navRects.length; j++) {
        expect(
          intersects(navRects[i] as Rect, navRects[j] as Rect),
          `${navRects[i]!.testid} and ${navRects[j]!.testid} must not overlap`,
        ).toBe(false);
      }
    }
  });

  test("display controls + user area do not overlap the primary nav", async ({
    page,
  }, testInfo) => {
    await page.goto("/");
    const navRect = await rectOf(page, "[data-testid='topbar-primary-nav']");
    const controlsRect = await rectOf(
      page,
      "[data-testid='topbar-display-controls']",
    );
    const userRect = await rectOf(page, "[data-testid='topbar-user-area']");
    expect(intersects(navRect, controlsRect)).toBe(false);
    expect(intersects(navRect, userRect)).toBe(false);
    expect(intersects(controlsRect, userRect)).toBe(false);
    const viewport = page.viewportSize()!;
    for (const r of [navRect, controlsRect, userRect]) {
      expect(r.left).toBeGreaterThanOrEqual(-1);
      expect(r.right).toBeLessThanOrEqual(viewport.width + 1);
    }
    if (testInfo.project.name === "desktop") {
      // Desktop keeps the source-level flex order: nav → controls → user.
      expect(navRect.right).toBeLessThanOrEqual(controlsRect.left + 1);
      expect(controlsRect.right).toBeLessThanOrEqual(userRect.left + 1);
    } else {
      // Mobile moves the nav into its own wrapped row.
      expect(navRect.top).toBeGreaterThanOrEqual(
        Math.min(controlsRect.bottom, userRect.bottom) - 1,
      );
    }
  });

  test("auth_login two-column layout fits the viewport", async ({
    page,
  }, testInfo) => {
    await page.goto("/");
    await page.click("[data-testid='topbar-login']");
    await page.waitForSelector("[data-testid='route-auth_login']");

    const shell = await rectOf(
      page,
      "[data-testid='route-auth_login']",
    );
    const side = await rectOf(
      page,
      "[data-testid='route-auth_login'] .ei-auth-side",
    );
    const card = await rectOf(
      page,
      "[data-testid='route-auth_login'] .ei-auth-card",
    );
    const viewport = page.viewportSize()!;

    // Card and side must both be inside the shell horizontally.
    expect(side.left).toBeGreaterThanOrEqual(shell.left - 1);
    expect(card.right).toBeLessThanOrEqual(shell.right + 1);

    // Card width must be positive (form actually renders).
    expect(card.width).toBeGreaterThan(80);

    // Both columns must stay within the viewport width.
    expect(side.right).toBeLessThanOrEqual(viewport.width + 1);
    expect(card.right).toBeLessThanOrEqual(viewport.width + 1);

    testInfo.attach(`auth-shell-${testInfo.project.name}`, {
      body: JSON.stringify({ shell, side, card, viewport }, null, 2),
      contentType: "application/json",
    });

    if (testInfo.project.name === "desktop") {
      // Two-column desktop layout: side ends before card begins.
      expect(side.right).toBeLessThanOrEqual(card.left + 1);
    }
    if (testInfo.project.name === "mobile") {
      // Mobile fold: AuthShell collapses to a single column under 768px
      // (auth.css media query). Side and card stack vertically rather
      // than appearing side-by-side, so we assert vertical stacking and
      // bottom-within-document instead of horizontal layout.
      expect(card.top).toBeGreaterThanOrEqual(side.top - 1);
      const scrollHeight = await page.evaluate(
        () => document.body.scrollHeight,
      );
      expect(card.bottom).toBeLessThanOrEqual(scrollHeight + 1);
    }
  });
});
