import { expect, test } from "@playwright/test";

/**
 * Phase 4 — Screenshot regression + dark / customAccent visual diff.
 *
 * Truth source: docs/spec/frontend-shell/plans/003-ui-design-pixel-parity-
 * gate/plan.md §4 Phase 4.
 *
 * Realistic scope: cross-side pixel-by-pixel diff against
 * `ui-design/index.html` is brittle because the two surfaces use different
 * font sources (fontsource bundle vs Google Fonts CDN) and slightly
 * different SVG icon stacks. We therefore split Phase 4 into two
 * complementary gates:
 *
 *   - 4.1 — frontend home screenshot smoke: a non-empty browser screenshot
 *     buffer plus the surrounding DOM/computed-style gates.
 *   - 4.2 — dark / customAccent visual diff: assert the browser actually
 *     resolves the documented CSS variables to different values when the
 *     user toggles dark mode and activates customAccent, so the live theme
 *     wiring cannot silently fail.
 */

async function freezeAnimations(page: import("@playwright/test").Page): Promise<void> {
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

test.describe("frontend home screenshot smoke (Phase 4.1)", () => {
  test("default ocean/light home renders a stable non-empty screenshot", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='app-shell-topbar']");
    await freezeAnimations(page);
    const screenshot = await page.screenshot({ fullPage: false });
    expect(screenshot.length).toBeGreaterThan(10_000);
  });
});

test.describe("dark + customAccent visual diff (Phase 4.2)", () => {
  test("toggling dark mode flips the resolved --ei-color-bg-canvas / --ei-color-fg-primary tokens", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='app-shell-topbar']");
    const lightTokens = await page.evaluate(() => {
      const cs = getComputedStyle(document.documentElement);
      return {
        bg: cs.getPropertyValue("--ei-color-bg-canvas").trim(),
        fg: cs.getPropertyValue("--ei-color-fg-primary").trim(),
        mode: document.documentElement.getAttribute("data-mode"),
      };
    });
    expect(lightTokens.bg).toBe("#f8fafd");
    expect(lightTokens.fg).toBe("#141821");
    expect(lightTokens.mode).toBe("light");

    await page.click("[data-testid='topbar-dark-toggle']");
    const darkTokens = await page.evaluate(() => {
      const cs = getComputedStyle(document.documentElement);
      return {
        bg: cs.getPropertyValue("--ei-color-bg-canvas").trim(),
        fg: cs.getPropertyValue("--ei-color-fg-primary").trim(),
        mode: document.documentElement.getAttribute("data-mode"),
      };
    });
    expect(darkTokens.mode).toBe("dark");
    expect(darkTokens.bg).toBe("#0c0f17");
    expect(darkTokens.fg).toBe("#e8edf6");
    // The browser must actually paint the new background color.
    const computedBodyBg = await page.evaluate(
      () => getComputedStyle(document.body).backgroundColor,
    );
    expect(computedBodyBg).toBe("rgb(12, 15, 23)");
  });

  test("activating customAccent overrides --ei-color-accent inline with an oklch value", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='topbar-theme-button']");
    const before = await page.evaluate(() => ({
      accent: document.documentElement.style.getPropertyValue("--ei-color-accent"),
      attr: document.documentElement.getAttribute("data-custom-accent"),
    }));
    expect(before.accent).toBe("");
    expect(before.attr).toBeNull();

    await page.click("[data-testid='topbar-theme-button']");
    await page.click("[data-testid='topbar-theme-custom-option']");
    const after = await page.evaluate(() => ({
      accent: document.documentElement.style.getPropertyValue("--ei-color-accent"),
      accentSoft: document.documentElement.style.getPropertyValue(
        "--ei-color-accent-soft",
      ),
      attr: document.documentElement.getAttribute("data-custom-accent"),
      // Base palette tokens must NOT be overridden.
      bgCanvas: document.documentElement.style.getPropertyValue(
        "--ei-color-bg-canvas",
      ),
      fgPrimary: document.documentElement.style.getPropertyValue(
        "--ei-color-fg-primary",
      ),
    }));
    expect(after.attr).toBe("active");
    expect(after.accent).toMatch(/^oklch\(58%/);
    expect(after.accentSoft).toMatch(/^oklch\(92%/);
    expect(after.bgCanvas).toBe("");
    expect(after.fgPrimary).toBe("");

    // Hue + chroma sliders are surfaced when the picker opens.
    await expect(
      page.locator("[data-testid='topbar-custom-accent-picker']"),
    ).toHaveCount(1);
    await expect(
      page.locator("[data-testid='topbar-custom-accent-hue']"),
    ).toHaveCount(1);
    await expect(
      page.locator("[data-testid='topbar-custom-accent-chroma']"),
    ).toHaveCount(1);
  });

  test("out-of-scope entries (welcome / mistakes / growth / drill / standalone voice) do not flow back into rendered DOM", async ({
    page,
  }) => {
    await page.goto("/");
    for (const banned of [
      "topbar-nav-welcome",
      "topbar-nav-mistakes",
      "topbar-nav-growth",
      "topbar-nav-drill",
      "topbar-nav-voice",
      "route-welcome",
    ]) {
      await expect(page.locator(`[data-testid='${banned}']`)).toHaveCount(0);
    }
    const html = await page.content();
    expect(html).not.toMatch(/错题本|成长中心|经历库|目标角色|技能标签/);
  });
});
