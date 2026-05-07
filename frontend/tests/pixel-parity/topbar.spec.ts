import { expect, test } from "@playwright/test";

/**
 * Phase 2.1 — TopBar DOM + computed style parity.
 *
 * Truth source: docs/spec/frontend-shell/plans/003-ui-design-pixel-parity-
 * gate/plan.md §4 Phase 2.1.
 *
 * The frontend dist mounts the production React shell at `/`, while the
 * ui-design golden preview is mounted at `/ui-design/`. Both default to the
 * Home route and render the TopBar. We compare:
 *
 *   - Five primary nav entries by visible label (English by default for
 *     frontend, Chinese-default for ui-design).
 *   - TopBar shell computed style (height, padding-left, padding-right,
 *     border-bottom-width, background-color) within a small tolerance.
 *
 * The frontend uses semantic data-testid attributes from D2; the ui-design
 * preview uses inline-style structural anchors. We therefore use a header /
 * nav structural selector on the ui-design side and the frontend's testids on
 * the frontend side, then compare the surfaces by content + computed style.
 */

const FRONTEND_PATH = "/";
const UI_DESIGN_PATH = "/ui-design/";

const PRIMARY_NAV_LABELS_ZH = [
  "首页",
  "岗位推荐",
  "模拟面试",
  "简历",
  "复盘",
] as const;
const PRIMARY_NAV_LABELS_EN = [
  "Home",
  "Job Picks",
  "Mock Interview",
  "Resume",
  "Debrief",
] as const;

test.describe("TopBar DOM + computed style parity", () => {
  test("frontend dist renders five primary nav testids with the documented English labels", async ({
    page,
  }) => {
    await page.goto(FRONTEND_PATH);
    await page.waitForSelector("[data-testid='app-shell-topbar']");
    const labels = await page.$$eval(
      "[data-testid='topbar-primary-nav'] button[data-testid^='topbar-nav-']",
      (els) => els.map((el) => el.textContent?.trim()),
    );
    expect(labels).toEqual([...PRIMARY_NAV_LABELS_EN]);
  });

  test("frontend TopBar visible structure matches ui-design source-level controls", async ({
    page,
  }) => {
    await page.goto(FRONTEND_PATH);
    await page.waitForSelector("[data-testid='app-shell-topbar']");
    const summary = await page.evaluate(() => {
      const topbar = document.querySelector(
        "[data-testid='app-shell-topbar']",
      ) as HTMLElement | null;
      if (!topbar) throw new Error("frontend topbar missing");
      return {
        brand: topbar.querySelector(".ei-topbar-brand")?.textContent?.replace(/\s+/g, " ").trim(),
        selectCount: topbar.querySelectorAll("select").length,
        navIconCount: topbar.querySelectorAll("[data-testid^='topbar-nav-icon-']").length,
        buttonTexts: Array.from(topbar.querySelectorAll("button")).map((button) =>
          (button.textContent ?? "").replace(/\s+/g, " ").trim(),
        ),
        themeTitle: topbar
          .querySelector("[data-testid='topbar-theme-button']")
          ?.getAttribute("title"),
        langText: topbar
          .querySelector("[data-testid='topbar-lang-toggle']")
          ?.textContent?.replace(/\s+/g, " ").trim(),
      };
    });

    expect(summary.brand).toBe("EEasyInterview面试训练器 · v1.0");
    expect(summary.selectCount).toBe(0);
    expect(summary.navIconCount).toBe(5);
    expect(summary.themeTitle).toBe("Theme");
    expect(summary.langText).toBe("EN · 中");
    expect(summary.buttonTexts).toContain("EN · 中");
  });

  test("frontend theme menu exposes the ui-design theme list and custom accent picker", async ({
    page,
  }) => {
    await page.goto(FRONTEND_PATH);
    await page.waitForSelector("[data-testid='topbar-theme-button']");
    await page.click("[data-testid='topbar-theme-button']");
    await page.waitForSelector("[data-testid='topbar-theme-menu']");

    const menu = page.locator("[data-testid='topbar-theme-menu']");
    await expect(menu).toBeVisible();
    await expect(page.locator("[data-testid^='topbar-theme-option-']")).toHaveCount(4);
    await expect(page.locator("[data-testid='topbar-theme-custom-option']")).toHaveText(/Custom/);

    await page.click("[data-testid='topbar-theme-custom-option']");
    await expect(page.locator("[data-testid='topbar-custom-accent-picker']")).toHaveCount(1);
    await expect(page.locator("[data-testid='topbar-custom-accent-hue']")).toHaveCount(1);
    await expect(page.locator("[data-testid='topbar-custom-accent-chroma']")).toHaveCount(1);
    await expect(page.locator("[data-testid='topbar-custom-accent-clear']")).toHaveCount(1);
  });

  test("ui-design golden preview renders five primary nav buttons with Chinese labels", async ({
    page,
  }) => {
    await page.goto(UI_DESIGN_PATH);
    // Wait for Babel-transpiled scripts to mount the React tree.
    await page.waitForFunction(
      () => document.querySelector("nav button") !== null,
      undefined,
      { timeout: 15_000 },
    );
    const navTexts = await page.$$eval(
      "nav button",
      (els) => els.map((el) => (el.textContent ?? "").replace(/\s+/g, " ").trim()),
    );
    // ui-design renders an `<Icon />` SVG followed by a label; the resulting
    // textContent should end with the label string. We assert each expected
    // label appears as a suffix of one nav button.
    for (const label of PRIMARY_NAV_LABELS_ZH) {
      const matched = navTexts.some((text) => text.endsWith(label));
      expect(matched, `ui-design nav must contain a button ending with ${label} (got ${JSON.stringify(navTexts)})`).toBe(true);
    }
  });

  test("frontend TopBar shell height matches ui-design TopBar shell height (58px)", async ({
    page,
  }) => {
    await page.goto(FRONTEND_PATH);
    const frontendHeight = await page.evaluate(() => {
      const el = document.querySelector(
        "[data-testid='app-shell-topbar']",
      ) as HTMLElement | null;
      if (!el) throw new Error("frontend topbar missing");
      return el.getBoundingClientRect().height;
    });

    await page.goto(UI_DESIGN_PATH);
    await page.waitForFunction(
      () => document.querySelector("nav button") !== null,
      undefined,
      { timeout: 15_000 },
    );
    const uiDesignHeight = await page.evaluate(() => {
      // ui-design TopBar is the first `<div>` whose direct child is a
      // sticky-positioned header with the brand mark. Use the parent of the
      // <nav> as a structural anchor.
      const nav = document.querySelector("nav");
      if (!nav) throw new Error("ui-design nav missing");
      const header = nav.parentElement as HTMLElement | null;
      if (!header) throw new Error("ui-design header missing");
      return header.getBoundingClientRect().height;
    });

    // Both sides target 58px height per ui-design/src/app.jsx TopBar literal.
    expect(frontendHeight).toBeCloseTo(58, 0);
    expect(uiDesignHeight).toBeCloseTo(58, 0);
    // Pairwise tolerance: 1px (ui-design rounds to integer; frontend uses
    // semantic CSS variable that resolves to 58px exactly).
    expect(Math.abs(frontendHeight - uiDesignHeight)).toBeLessThanOrEqual(1);
  });

  test("frontend TopBar padding-left / padding-right honours --ei-space-8 (32px)", async ({
    page,
  }) => {
    await page.goto(FRONTEND_PATH);
    const padding = await page.evaluate(() => {
      const el = document.querySelector(
        "[data-testid='app-shell-topbar']",
      ) as HTMLElement | null;
      if (!el) throw new Error("frontend topbar missing");
      const cs = getComputedStyle(el);
      return { left: cs.paddingLeft, right: cs.paddingRight };
    });
    expect(padding.left).toBe("32px");
    expect(padding.right).toBe("32px");
  });

  test("frontend TopBar border-bottom resolves to 1px solid via --ei-color-rule-strong", async ({
    page,
  }) => {
    await page.goto(FRONTEND_PATH);
    const border = await page.evaluate(() => {
      const el = document.querySelector(
        "[data-testid='app-shell-topbar']",
      ) as HTMLElement | null;
      if (!el) throw new Error("frontend topbar missing");
      const cs = getComputedStyle(el);
      return {
        width: cs.borderBottomWidth,
        style: cs.borderBottomStyle,
        // border-bottom-color resolves to rgb(231, 226, 214) for warm/light.
        color: cs.borderBottomColor,
      };
    });
    expect(border.width).toBe("1px");
    expect(border.style).toBe("solid");
    expect(border.color).toBe("rgb(231, 226, 214)");
  });

  test("frontend default home renders aria-current=page on the Home nav button", async ({
    page,
  }) => {
    await page.goto(FRONTEND_PATH);
    const ariaCurrent = await page.getAttribute(
      "[data-testid='topbar-nav-home']",
      "aria-current",
    );
    expect(ariaCurrent).toBe("page");
    for (const route of [
      "jd_match",
      "workspace",
      "resume_versions",
      "debrief",
    ]) {
      const value = await page.getAttribute(
        `[data-testid='topbar-nav-${route}']`,
        "aria-current",
      );
      expect(value).toBeNull();
    }
  });

  test("frontend topbar-dark-toggle defaults to aria-pressed=false", async ({
    page,
  }) => {
    await page.goto(FRONTEND_PATH);
    const pressed = await page.getAttribute(
      "[data-testid='topbar-dark-toggle']",
      "aria-pressed",
    );
    expect(pressed).toBe("false");
  });
});
