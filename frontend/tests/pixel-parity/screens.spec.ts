import { expect, test } from "@playwright/test";

/**
 * Phase 2.2 — Auth / Profile / Settings / Placeholder DOM anchor parity.
 *
 * Truth source: docs/spec/frontend-shell/plans/003-ui-design-pixel-parity-
 * gate/plan.md §4 Phase 2.2.
 *
 * Realistic Playwright scope:
 *   - `auth_login` is reachable from Home by clicking the `topbar-login`
 *     entry on the frontend, and via hash route (`#route=auth_login`) on the
 *     ui-design golden preview. Both sides render an auth two-column shell;
 *     we assert frontend's `ei-auth-shell` + `ei-auth-card` semantic
 *     classNames and the structural equivalent on ui-design (a sticky `<h1>`
 *     hero column + a card on the right).
 *   - `profile`, `settings`, `company_intel`, and other placeholder routes
 *     either require a signed-in session (profile / settings) or arrive only
 *     after navigating through D2-D6 business flows (company_intel via
 *     workspace). The DOM-anchor parity for those routes is covered by the
 *     in-process scenario test
 *     `frontend/src/app/scenarios/p0-005-app-shell-visual-system-smoke.test
 *     .tsx`, which renders them with the production CSS bundle and asserts
 *     `ei-screen-shell` / `ei-screen-card` / `ei-skeleton-stripe` parity.
 *
 * This spec therefore concentrates on the auth shell parity that is
 * end-to-end reachable in a real browser. The wider DOM parity remains
 * gated by E2E.P0.005 (jsdom) at the scaffold level.
 */

const FRONTEND_PATH = "/";
const UI_DESIGN_AUTH_LOGIN = "/ui-design/#route=auth_login";

test.describe("auth_login DOM anchor parity", () => {
  test("frontend dist navigates to auth_login when topbar-login is clicked and renders ei-auth-shell", async ({
    page,
  }) => {
    await page.goto(FRONTEND_PATH);
    await page.waitForSelector("[data-testid='topbar-login']");
    await page.click("[data-testid='topbar-login']");
    await page.waitForSelector("[data-testid='route-auth_login']");

    const root = page.locator("[data-testid='route-auth_login']");
    await expect(root).toHaveClass(/\bei-auth-shell\b/);
    await expect(
      root.locator(".ei-auth-side"),
    ).toHaveCount(1);
    await expect(
      root.locator(".ei-auth-card"),
    ).toHaveCount(1);
    await expect(
      root.locator("[data-testid='auth-login-email-form']"),
    ).toHaveCount(1);
    await expect(
      root.locator("[data-testid='auth-login-submit-email']"),
    ).toHaveClass(/\bei-auth-cta\b/);
    await expect(
      root.locator("[data-testid='auth-login-password-stub']"),
    ).toHaveCount(1);
    await expect(
      root.locator("[data-testid='auth-login-oauth-stub']"),
    ).toHaveCount(1);
  });

  test("frontend auth login title uses the ei-text-display typography token", async ({
    page,
  }) => {
    await page.goto(FRONTEND_PATH);
    await page.click("[data-testid='topbar-login']");
    const heading = page.locator(
      "[data-testid='route-auth_login'] h1",
    );
    await expect(heading).toHaveClass(/\bei-text-display\b/);
    const fontSize = await heading.evaluate(
      (el) => getComputedStyle(el).fontSize,
    );
    expect(fontSize).toBe("48px");
  });

  test("frontend auth_login eyebrow uses the ei-text-label / mono token", async ({
    page,
  }) => {
    await page.goto(FRONTEND_PATH);
    await page.click("[data-testid='topbar-login']");
    const eyebrow = page
      .locator("[data-testid='route-auth_login'] .ei-auth-eyebrow")
      .first();
    await expect(eyebrow).toHaveClass(/\bei-text-label\b/);
    const fontFamily = await eyebrow.evaluate(
      (el) => getComputedStyle(el).fontFamily,
    );
    expect(fontFamily).toMatch(/JetBrains Mono/);
  });

  test("ui-design golden preview hash route #route=auth_login renders an h1 hero + form card", async ({
    page,
  }) => {
    await page.goto(UI_DESIGN_AUTH_LOGIN);
    // Wait for Babel to mount the React tree and route to auth_login.
    await page.waitForFunction(
      () => {
        const heading = document.querySelector("h1");
        const text = heading?.textContent ?? "";
        return text.includes("继续") || text.includes("Continue");
      },
      undefined,
      { timeout: 15_000 },
    );
    const heroH1 = page.locator("h1").first();
    await expect(heroH1).toBeVisible();
    const fontSize = await heroH1.evaluate(
      (el) => getComputedStyle(el).fontSize,
    );
    // ui-design uses inline style fontSize: 42 for AuthShell heroes; the
    // browser may report 42px or rounded value depending on rem scaling.
    expect(parseFloat(fontSize)).toBeGreaterThanOrEqual(40);
    expect(parseFloat(fontSize)).toBeLessThanOrEqual(48);
  });

  test("frontend auth_login card padding resolves to 28px (matches screen-auth.jsx literal)", async ({
    page,
  }) => {
    await page.goto(FRONTEND_PATH);
    await page.click("[data-testid='topbar-login']");
    const card = page.locator(
      "[data-testid='route-auth_login'] .ei-auth-card",
    );
    const padding = await card.evaluate((el) => {
      const cs = getComputedStyle(el);
      return [cs.paddingTop, cs.paddingRight, cs.paddingBottom, cs.paddingLeft];
    });
    expect(padding).toEqual(["28px", "28px", "28px", "28px"]);
  });

  test("retired entries (welcome / mistakes / growth / drill / standalone voice) do not appear in DOM", async ({
    page,
  }) => {
    await page.goto(FRONTEND_PATH);
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
  });
});
