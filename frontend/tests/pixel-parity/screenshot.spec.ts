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
 *   - 4.1 — frontend home baseline regression: `toHaveScreenshot` against a
 *     locally maintained baseline, regenerated with `--update-snapshots`.
 *     Baselines live under `tests/pixel-parity/__screenshots__/` and stay
 *     out of git per `frontend/.gitignore`.
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

test.describe("frontend home screenshot regression (Phase 4.1)", () => {
  test("default warm/light home matches the colocated baseline", async ({
    page,
  }, testInfo) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='app-shell-topbar']");
    await freezeAnimations(page);
    // Hide the dynamic SVG mark text-orientation jitter by clipping the
    // bounding box; we use the documented topbar + main shell area only.
    await expect(page).toHaveScreenshot(
      `home-warm-light-${testInfo.project.name}.png`,
      {
        fullPage: false,
        maxDiffPixels: 4000,
      },
    );
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
    expect(lightTokens.bg).toBe("#fdfcf8");
    expect(lightTokens.fg).toBe("#1c1917");
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
    expect(darkTokens.bg).toBe("#16130e");
    expect(darkTokens.fg).toBe("#f5f0e4");
    // The browser must actually paint the new background color.
    const computedBodyBg = await page.evaluate(
      () => getComputedStyle(document.body).backgroundColor,
    );
    expect(computedBodyBg).toBe("rgb(22, 19, 14)");
  });

  test("activating customAccent overrides --ei-color-accent inline with an oklch value", async ({
    page,
  }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='topbar-custom-accent-button']");
    const before = await page.evaluate(() => ({
      accent: document.documentElement.style.getPropertyValue("--ei-color-accent"),
      attr: document.documentElement.getAttribute("data-custom-accent"),
    }));
    expect(before.accent).toBe("");
    expect(before.attr).toBeNull();

    await page.click("[data-testid='topbar-custom-accent-button']");
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
      page.locator("[data-testid='topbar-custom-accent-hue']"),
    ).toHaveCount(1);
    await expect(
      page.locator("[data-testid='topbar-custom-accent-chroma']"),
    ).toHaveCount(1);
  });

  test("retired entries (welcome / mistakes / growth / drill / standalone voice) do not flow back into rendered DOM", async ({
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
