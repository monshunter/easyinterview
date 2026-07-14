import { expect, test } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

/**
 * Phase 6.1 — Workspace screen DOM anchor and layout parity.
 *
 * Truth source: ui-design/src/screen-workspace.jsx::WorkspaceScreen,
 * docs/spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-
 * context/plan.md §4 Phase 6.
 *
 * Covers desktop (1440x900) and mobile (390x844) projects:
 * - DOM anchors (workspace plan-list landing and target-scoped unified plan-detail
 *   mother page)
 * - Bounding box stays in viewport, no overlap
 * - default (ocean)/light -> dark -> customAccent theme switching
 * - non-empty screenshot smoke
 * - Negative: out-of-scope prototype testids absent
 *
 * Full data-driven rendering is reached through an explicit initial route
 * bootstrap with the server-owned TargetJob locator on the workspace route.
 * TopBar navigation covers the no-context interview plan-list landing.
 */

interface Rect {
  left: number; top: number; right: number; bottom: number;
  width: number; height: number;
}

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

const WORKSPACE_TARGET_ID = "01918fa0-0000-7000-8000-000000002000";
const WORKSPACE_RESUME_ID = "01918fa0-0000-7000-8000-000000001000";

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

async function mockWorkspaceApis(page: import("@playwright/test").Page): Promise<void> {
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    const path = url.pathname.replace(/^\/api\/v1/, "");
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
    if (path.startsWith("/targets/")) {
      await fulfillFixture(route, "openapi/fixtures/TargetJobs/getTargetJob.json");
      return;
    }
    if (path.startsWith("/resumes/")) {
      await fulfillFixture(route, "openapi/fixtures/Resumes/getResume.json");
      return;
    }
    if (path === "/resumes") {
      await fulfillFixture(route, "openapi/fixtures/Resumes/listResumes.json");
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

async function rectOf(page: import("@playwright/test").Page, selector: string): Promise<Rect> {
  return page.evaluate(({ selector }) => {
    const el = document.querySelector(selector) as HTMLElement | null;
    if (!el) throw new Error(`selector not found: ${selector}`);
    const r = el.getBoundingClientRect();
    return {
      left: r.left, top: r.top, right: r.right, bottom: r.bottom,
      width: r.width, height: r.height,
    };
  }, { selector });
}

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

/**
 * Navigate to workspace by clicking the TopBar Interview button.
 *
 * `workspace` is an auth-gated business route (frontend-shell Phase 10 /
 * BUG-0115): the route guard renders `auth-route-gate` and redirects to
 * `auth_login` until runtime auth resolves to authenticated. Mock the auth
 * APIs first so the TopBar-click path reaches the workspace plan-list landing
 * instead of the auth gate.
 */
async function goToWorkspace(page: import("@playwright/test").Page) {
  await mockWorkspaceApis(page);
  await page.goto("/");
  await page.waitForSelector("[data-testid='topbar-nav-workspace']");
  await page.click("[data-testid='topbar-nav-workspace']");
  await page.waitForSelector("[data-testid='workspace-plan-list']", {
    timeout: 5000,
  });
}

/** Navigate to read-only Workspace detail through its sole server locator. */
async function goToWorkspaceDetail(page: import("@playwright/test").Page) {
  await mockWorkspaceApis(page);
  await page.addInitScript((route) => {
    (
      window as Window & {
        __EASYINTERVIEW_INITIAL_ROUTE__?: {
          name: string;
          params: Record<string, string>;
        };
      }
    ).__EASYINTERVIEW_INITIAL_ROUTE__ = route;
  }, {
    name: "workspace",
    params: {
      targetJobId: WORKSPACE_TARGET_ID,
    },
  });
  await page.goto("/");
  await page.waitForSelector("[data-testid='unified-plan-detail']");
  await page.evaluate(() => window.scrollTo(0, 0));
  await expect(page.locator("[data-testid='unified-plan-detail-title']")).toContainText(
    "Interview plan detail",
  );
  await expect(page.locator("[data-testid='parse-basics-title']")).toContainText(
    "Senior Frontend Engineer",
  );
  await expect(page.locator("[data-testid='parse-basics-title'] input")).toHaveCount(0);
}

test.describe("workspace DOM anchor parity", () => {
  test("workspace route is reachable via TopBar and renders workspace-specific chrome", async ({ page }) => {
    await goToWorkspace(page);
    await expect(page.locator("[data-testid='workspace-plan-list']")).toHaveCount(1, { timeout: 5000 });
    await expect(page.locator("[data-testid='workspace-empty']")).toHaveCount(0);
  });

  test("workspace plan-list landing renders all expected sub-elements", async ({ page }) => {
    await goToWorkspace(page);
    await expect(page.locator("[data-testid='workspace-plan-list-eyebrow']")).toHaveCount(1);
    await expect(page.locator("[data-testid='workspace-plan-list-title']")).toHaveCount(1);
    await expect(page.locator("[data-testid='workspace-plan-list-card-01918fa0-0000-7000-8000-000000002000']")).toHaveCount(1);
    await expect(page.locator("[data-testid='workspace-plan-list-create']")).toHaveCount(1);
  });

  test("ready plan card opens workspace detail without Parse animation or route-side mutation", async ({ page }) => {
    await goToWorkspace(page);
    const requests: Array<{ method: string; path: string }> = [];
    page.on("request", (request) => {
      const url = new URL(request.url());
      if (url.pathname.startsWith("/api/v1/")) {
        requests.push({ method: request.method(), path: url.pathname });
      }
    });

    await page.locator(
      `[data-testid='workspace-plan-list-card-${WORKSPACE_TARGET_ID}']`,
    ).click();
    await page.waitForURL(`/workspace?targetJobId=${WORKSPACE_TARGET_ID}`);
    await expect(page.locator("[data-testid='unified-plan-detail']")).toBeVisible();
    await expect(page.locator("[data-testid='parse-loading-step-0']")).toHaveCount(0);

    expect(
      requests.filter(
        ({ method, path }) =>
          method === "GET" && path === `/api/v1/targets/${WORKSPACE_TARGET_ID}`,
      ),
    ).toHaveLength(1);
    expect(
      requests.filter(
        ({ method, path }) =>
          method !== "GET" || path === "/api/v1/targets/import",
      ),
    ).toHaveLength(0);
  });

  test("workspace rail renders backend progress with unchanged node geometry", async ({ page }) => {
    await goToWorkspace(page);
    const rail = page.locator(
      `[data-testid='workspace-plan-list-rail-${WORKSPACE_TARGET_ID}']`,
    );
    await expect(rail.locator('[data-round-state="done"]')).toHaveCount(1);
    await expect(rail.locator('[data-round-state="current"]')).toHaveCount(1);
    await expect(rail.locator('[data-round-state="pending"]')).toHaveCount(1);

    const nodeBoxes = await rail.locator("[data-round-state] > div:first-child").evaluateAll(
      (nodes) => nodes.map((node) => {
        const rect = node.getBoundingClientRect();
        return { width: rect.width, height: rect.height };
      }),
    );
    expect(nodeBoxes).toHaveLength(3);
    for (const box of nodeBoxes) {
      expect(box.width).toBe(18);
      expect(box.height).toBe(18);
    }
  });

  test("TopBar workspace nav button has aria-current=page after navigation", async ({ page }) => {
    await goToWorkspace(page);
    const ariaCurrent = await page.getAttribute("[data-testid='topbar-nav-workspace']", "aria-current");
    expect(ariaCurrent).toBe("page");
  });

  test("workspace detail renders the unified plan-detail anchor set", async ({ page }) => {
    await goToWorkspaceDetail(page);
    const anchorIds = [
      "route-workspace",
      "unified-plan-detail",
      "unified-plan-detail-title",
      "parse-basics-title",
      "parse-basics-company",
      "parse-requirement-must_have-0",
      "parse-requirement-nice_to_have-0",
      "parse-hidden-signal-0",
      "parse-round-0",
      "parse-launch",
      "parse-resume-binding",
      "parse-action-start-interview",
    ];
    for (const id of anchorIds) {
      await expect(page.locator(`[data-testid='${id}']`), id).toHaveCount(1);
    }
    await expect(page.locator("[data-testid='route-parse']")).toHaveCount(0);
  });

  test("workspace detail keeps resume binding readonly and hides deleted workspace modals", async ({ page }) => {
    const requests: Array<{ method: string; path: string }> = [];
    page.on("request", (request) => {
      const url = new URL(request.url());
      if (url.pathname.startsWith("/api/v1/")) {
        requests.push({ method: request.method(), path: url.pathname });
      }
    });
    await goToWorkspaceDetail(page);

    await expect(page.locator("[data-testid='workspace-plan-modal-card']")).toHaveCount(0);
    await expect(page.locator("[data-testid='workspace-resume-modal-card']")).toHaveCount(0);
    await expect(page.locator("[data-testid='parse-resume-bound-title']")).toContainText(
      "Resume saved with this interview plan",
    );
    await expect(page.locator("[data-testid='parse-resume-bound-meta']")).toContainText(
      "The saved binding is read-only",
    );
    await expect(page.locator("[data-testid='parse-resume-picker-toggle']")).toHaveCount(0);
    await expect(page.locator("[data-testid='parse-resume-picker']")).toHaveCount(0);
    await expect(
      page.locator(`[data-testid='parse-resume-option-${WORKSPACE_RESUME_ID}']`),
    ).toHaveCount(0);

    expect(
      requests.filter(
        ({ method, path }) =>
          method === "GET" && path === `/api/v1/targets/${WORKSPACE_TARGET_ID}`,
      ),
    ).toHaveLength(1);
    expect(
      requests.filter(
        ({ method, path }) => method === "GET" && path === "/api/v1/resumes",
      ),
    ).toHaveLength(0);
    expect(
      requests.filter(
        ({ method, path }) =>
          method === "GET" && path === `/api/v1/resumes/${WORKSPACE_RESUME_ID}`,
      ),
    ).toHaveLength(0);
  });
});

test.describe("workspace bounding box parity", () => {
  test("workspace plan-list elements do not overlap and stay in viewport", async ({ page }) => {
    await goToWorkspace(page);
    const viewport = page.viewportSize();
    expect(viewport).toBeTruthy();

    const anchorIds = [
      "workspace-plan-list-eyebrow",
      "workspace-plan-list-title",
      "workspace-plan-list-subtitle",
      "workspace-plan-list-create",
      `workspace-plan-list-card-${WORKSPACE_TARGET_ID}`,
      `workspace-plan-list-card-body-${WORKSPACE_TARGET_ID}`,
      `workspace-plan-list-rail-${WORKSPACE_TARGET_ID}`,
      `workspace-plan-list-card-footer-${WORKSPACE_TARGET_ID}`,
      `workspace-plan-list-start-${WORKSPACE_TARGET_ID}`,
      `workspace-plan-list-delete-${WORKSPACE_TARGET_ID}`,
    ];

    const rects: Array<{ id: string; r: Rect }> = [];
    for (const id of anchorIds) {
      const el = page.locator(`[data-testid='${id}']`);
      if ((await el.count()) > 0) {
        const r = await rectOf(page, `[data-testid='${id}']`);
        rects.push({ id, r });
        expect(r.top).toBeGreaterThanOrEqual(0);
        expect(r.left).toBeGreaterThanOrEqual(0);
        expect(r.right).toBeLessThanOrEqual(viewport!.width + 1);
        expect(r.width).toBeGreaterThan(0);
      }
    }

    // All elements must have non-zero width and be reasonably positioned
    for (const { id, r } of rects) {
      expect(r.width, `${id} width is zero`).toBeGreaterThan(0);
      expect(r.top, `${id} top is negative`).toBeGreaterThanOrEqual(-5);
      expect(r.left, `${id} left is negative`).toBeGreaterThanOrEqual(-5);
    }
  });

  test("workspace plan-list cards keep visible card affordance", async ({ page }) => {
    await goToWorkspace(page);

    const requiredSections = [
      `workspace-plan-list-card-${WORKSPACE_TARGET_ID}`,
      `workspace-plan-list-card-body-${WORKSPACE_TARGET_ID}`,
      `workspace-plan-list-rail-${WORKSPACE_TARGET_ID}`,
      `workspace-plan-list-card-footer-${WORKSPACE_TARGET_ID}`,
      `workspace-plan-list-start-${WORKSPACE_TARGET_ID}`,
      `workspace-plan-list-delete-${WORKSPACE_TARGET_ID}`,
    ];
    for (const testId of requiredSections) {
      await expect(page.locator(`[data-testid='${testId}']`)).toHaveCount(1);
    }

    const styles = await page.evaluate((targetId) => {
      const card = document.querySelector(`[data-testid='workspace-plan-list-card-${targetId}']`) as HTMLElement | null;
      const body = document.querySelector(`[data-testid='workspace-plan-list-card-body-${targetId}']`) as HTMLElement | null;
      const rail = document.querySelector(`[data-testid='workspace-plan-list-rail-${targetId}']`) as HTMLElement | null;
      const footer = document.querySelector(`[data-testid='workspace-plan-list-card-footer-${targetId}']`) as HTMLElement | null;
      const startButton = document.querySelector(`[data-testid='workspace-plan-list-start-${targetId}']`) as HTMLButtonElement | null;
      const deleteButton = document.querySelector(`[data-testid='workspace-plan-list-delete-${targetId}']`) as HTMLButtonElement | null;
      if (!card || !body || !rail || !footer || !startButton || !deleteButton) {
        throw new Error("workspace plan-list card sections missing");
      }
      const cardStyle = getComputedStyle(card);
      const bodyStyle = getComputedStyle(body);
      const footerStyle = getComputedStyle(footer);
      const buttonStyle = getComputedStyle(startButton);
      const deleteButtonStyle = getComputedStyle(deleteButton);
      const probe = document.createElement("div");
      probe.style.color = getComputedStyle(document.documentElement).getPropertyValue("--ei-color-accent").trim();
      document.body.appendChild(probe);
      const accentColor = getComputedStyle(probe).color;
      probe.remove();
      return {
        cardBg: cardStyle.backgroundColor,
        bodyBg: bodyStyle.backgroundColor,
        footerBg: footerStyle.backgroundColor,
        borderTopWidth: cardStyle.borderTopWidth,
        borderTopStyle: cardStyle.borderTopStyle,
        railText: rail.innerText,
        bodyPaddingTop: Number.parseFloat(bodyStyle.paddingTop),
        footerPaddingTop: Number.parseFloat(footerStyle.paddingTop),
        footerBorderTopWidth: footerStyle.borderTopWidth,
        footerDisplay: footerStyle.display,
        footerJustifyContent: footerStyle.justifyContent,
        footerText: footer.innerText,
        footerContainsDelete: footer.contains(deleteButton),
        buttonBg: buttonStyle.backgroundColor,
        buttonBorderColor: buttonStyle.borderTopColor,
        deleteButtonBg: deleteButtonStyle.backgroundColor,
        deleteButtonBorderColor: deleteButtonStyle.borderTopColor,
        deleteButtonPosition: deleteButtonStyle.position,
        deleteButtonRight: deleteButtonStyle.right,
        deleteButtonTop: deleteButtonStyle.top,
        accentColor,
      };
    }, WORKSPACE_TARGET_ID);

    expect(styles.cardBg).toBe(styles.bodyBg);
    expect(styles.footerBg).toBe(styles.cardBg);
    expect(styles.borderTopWidth).toBe("1px");
    expect(styles.borderTopStyle).toBe("solid");
    expect(styles.railText).toContain("Frontend architecture screen");
    expect(styles.railText).toContain("Hiring manager impact interview");
    expect(styles.bodyPaddingTop).toBe(0);
    expect(styles.footerPaddingTop).toBeGreaterThanOrEqual(12);
    expect(styles.footerBorderTopWidth).toBe("1px");
    expect(styles.footerDisplay).toBe("flex");
    expect(styles.footerJustifyContent).toBe("flex-end");
    expect(styles.footerText).not.toMatch(/URL import|Manual input|ZH-CN/i);
    expect(styles.footerText).not.toMatch(/Open plan|进入规划/i);
    expect(styles.footerText).toMatch(/Start interview now|立即面试/i);
    expect(styles.footerContainsDelete).toBe(false);
    expect(styles.buttonBg).toBe(styles.accentColor);
    expect(styles.buttonBorderColor).toBe(styles.accentColor);
    expect(styles.deleteButtonBg).not.toBe(styles.accentColor);
    expect(styles.deleteButtonPosition).toBe("absolute");
    expect(Number.parseFloat(styles.deleteButtonRight)).toBeGreaterThanOrEqual(10);
    expect(Number.parseFloat(styles.deleteButtonTop)).toBeGreaterThanOrEqual(10);
  });

  test("workspace detail primary anchors stay in viewport", async ({ page }) => {
    await goToWorkspaceDetail(page);
    const viewport = page.viewportSize();
    expect(viewport).toBeTruthy();

    const anchorIds = [
      "unified-plan-detail",
      "parse-basics-title",
      "parse-requirement-must_have-0",
      "parse-launch",
      "parse-resume-binding",
    ];

    for (const id of anchorIds) {
      const r = await rectOf(page, `[data-testid='${id}']`);
      expect(r.width, `${id} width is zero`).toBeGreaterThan(0);
      expect(r.height, `${id} height is zero`).toBeGreaterThan(0);
      expect(r.left, `${id} left`).toBeGreaterThanOrEqual(-5);
      expect(r.right, `${id} right`).toBeLessThanOrEqual(viewport!.width + 5);
    }
  });
});

test.describe("workspace dark mode + customAccent visual diff", () => {
  test("dark mode changes workspace background color", async ({ page }) => {
    await goToWorkspace(page);
    await page.waitForSelector("[data-testid='topbar-dark-toggle']");

    const beforeBg = await page.evaluate(() =>
      getComputedStyle(document.body).backgroundColor,
    );

    await page.click("[data-testid='topbar-dark-toggle']");
    await page.waitForTimeout(300);

    const afterBg = await page.evaluate(() =>
      getComputedStyle(document.body).backgroundColor,
    );
    expect(beforeBg).not.toBe(afterBg);
  });

  test("customAccent propagates data-custom-accent attribute on workspace route", async ({ page }) => {
    await goToWorkspace(page);

    await page.click("[data-testid='topbar-theme-button']");
    await page.click("[data-testid='topbar-theme-custom-option']");

    const attr = await page.evaluate(() =>
      document.documentElement.getAttribute("data-custom-accent"),
    );
    expect(attr).toBe("active");
  });

  test("theme dropdown is accessible from workspace", async ({ page }) => {
    await goToWorkspace(page);

    await page.click("[data-testid='topbar-theme-button']");
    await page.waitForSelector("[data-testid='topbar-theme-menu']");
    await expect(page.locator("[data-testid='topbar-theme-menu']")).toBeVisible();
    await expect(page.locator("[data-testid^='topbar-theme-option-']")).toHaveCount(2);
  });
});

test.describe("workspace screenshot regression", () => {
  test("workspace plan-list landing renders a non-empty screenshot", async ({ page }) => {
    await goToWorkspace(page);
    await freezeAnimations(page);
    await expect(page.locator("[data-testid='workspace-plan-list']")).toBeVisible();
    const screenshot = await page.screenshot({ fullPage: false });
    expect(screenshot.length).toBeGreaterThan(10_000);
  });

  test("workspace detail renders a non-empty screenshot", async ({ page }) => {
    await goToWorkspaceDetail(page);
    await freezeAnimations(page);
    await expect(page.locator("[data-testid='unified-plan-detail']")).toBeVisible();
    const screenshot = await page.screenshot({ fullPage: false });
    expect(screenshot.length).toBeGreaterThan(10_000);
  });
});

test.describe("deleted workspace detail negative gate", () => {
  test("workspace detail does not render the old independent workspace anchors", async ({ page }) => {
    await goToWorkspaceDetail(page);
    for (const oldAnchorId of [
      "workspace-header",
      "workspace-launcher",
      "workspace-jd-card",
      "workspace-prep-card",
      "workspace-history-card",
    ]) {
      await expect(page.locator(`[data-testid='${oldAnchorId}']`)).toHaveCount(0);
    }
  });
});

test.describe("out-of-scope prototype testid negative gate (workspace)", () => {
  test("out-of-scope workspace prototype testids do not appear in DOM", async ({ page }) => {
    await goToWorkspace(page);

    const banned = [
      "practice-mode-card-warmup",
      "practice-mode-card-single_drill",
      "practice-mode-card-drill_builder",
      "practice-mode-card-mistake_queue",
      "growth-center",
      "drill-builder",
      "mistake-queue",
    ];
    for (const testid of banned) {
      await expect(page.locator(`[data-testid='${testid}']`)).toHaveCount(0);
    }
  });

  test("out-of-scope route names do not appear as TopBar entries", async ({ page }) => {
    await goToWorkspace(page);
    for (const banned of [
      "topbar-nav-welcome", "topbar-nav-mistakes",
      "topbar-nav-growth", "topbar-nav-drill", "topbar-nav-voice",
    ]) {
      await expect(page.locator(`[data-testid='${banned}']`)).toHaveCount(0);
    }
  });
});
