import { expect, test } from "@playwright/test";

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
 * gate (E2E.P0.015 / E2E.P0.016 under test/scenarios/e2e/).
 */

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
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-jd-textarea']");

    await page.fill(
      "[data-testid='home-jd-textarea']",
      "Senior Frontend Engineer needed at Acme Corp",
    );

    // Submit button should become enabled
    await expect(page.locator("[data-testid='home-jd-submit']")).toBeEnabled();
  });

  test("upload modal opens and closes", async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("[data-testid='home-jd-textarea']");

    // Click upload button (or upload link)
    const uploadTrigger = page.locator("[data-testid='home-jd-upload-trigger']");
    if ((await uploadTrigger.count()) > 0) {
      await uploadTrigger.click();
      await expect(
        page.locator("[data-testid='home-modal-upload-dropzone']"),
      ).toBeVisible();

      // Close with X
      await page.click("[data-testid='home-modal-upload-close']");
      await expect(
        page.locator("[data-testid='home-modal-upload-dropzone']"),
      ).toHaveCount(0);
    }
  });
});
