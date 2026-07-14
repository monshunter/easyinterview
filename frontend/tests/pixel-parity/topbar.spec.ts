import { expect, test } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

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
 *   - Three primary nav entries by visible label (English by default when the
 *     browser locale is unsupported or English).
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
const REPO_ROOT = resolve(process.cwd(), "..");

const PRIMARY_NAV_LABELS_EN = [
  "Home",
  "Interview",
  "Resume",
] as const;

interface OperationFixture {
  scenarios: Record<
    string,
    {
      response: {
        status: number;
        headers?: Record<string, string>;
        body?: unknown;
      };
    }
  >;
}

function fixtureResponse(relativePath: string, scenario = "default") {
  const absolutePath = resolve(REPO_ROOT, relativePath);
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
      ...(response.body === undefined ? {} : { "content-type": "application/json; charset=utf-8" }),
      ...(response.headers ?? {}),
    },
    body: response.body === undefined ? undefined : JSON.stringify(response.body),
  });
}

async function mockStatefulAuthApis(page: import("@playwright/test").Page) {
  let signedIn = false;
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    const path = url.pathname.replace(/^\/api\/v1/, "");
    if (path === "/runtime-config") {
      await fulfillFixture(route, "openapi/fixtures/Auth/getRuntimeConfig.json");
      return;
    }
    if (path === "/me") {
      await fulfillFixture(
        route,
        "openapi/fixtures/Auth/getMe.json",
        signedIn ? "authenticated" : "unauthenticated",
      );
      return;
    }
    if (path === "/auth/email/start") {
      await fulfillFixture(route, "openapi/fixtures/Auth/startAuthEmailChallenge.json");
      return;
    }
    if (path === "/auth/email/verify") {
      signedIn = true;
      await fulfillFixture(route, "openapi/fixtures/Auth/verifyAuthEmailChallenge.json");
      return;
    }
    if (path === "/auth/logout") {
      signedIn = false;
      await fulfillFixture(route, "openapi/fixtures/Auth/logout.json");
      return;
    }
    await route.fulfill({
      status: 404,
      headers: { "content-type": "application/json; charset=utf-8" },
      body: JSON.stringify({ error: { code: "NOT_FOUND", message: `No fixture for ${path}` } }),
    });
  });
}

async function gotoUiDesign(page: import("@playwright/test").Page) {
  await page.goto(UI_DESIGN_PATH, { waitUntil: "domcontentloaded" });
  await expect(page.locator("nav button").first()).toBeVisible({ timeout: 30_000 });
}

function assertUiDesignUserMenuSourceLiterals() {
  const source = readFileSync(resolve(REPO_ROOT, "ui-design/src/app.jsx"), "utf8");
  expect(source).toContain("minWidth: 220");
  expect(source).toContain('top: "calc(100% + 6px)"');
  expect(source).toContain('padding: "3px 10px 3px 3px"');
  expect(source).not.toContain('labelZh: "用户画像"');
  expect(source).toContain('labelZh: "设置与隐私"');
  expect(source).toContain('name="logout"');
}

test.describe("TopBar DOM + computed style parity", () => {
  test("frontend dist renders three primary nav testids with the documented English labels (D-22)", async ({
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
          (button.textContent ?? "").replace(/[▾▴]/g, "").replace(/\s+/g, " ").trim(),
        ),
        themeTitle: topbar
          .querySelector("[data-testid='topbar-theme-button']")
          ?.getAttribute("title"),
        langText: topbar
          .querySelector("[data-testid='topbar-lang-toggle']")
          ?.textContent?.replace(/[▾▴]/g, "").replace(/\s+/g, " ").trim(),
      };
    });

    expect(summary.brand).toBe("EEasyInterview");
    expect(summary.selectCount).toBe(0);
    expect(summary.navIconCount).toBe(3);
    expect(summary.themeTitle).toBe("Theme");
    expect(summary.langText).toBe("English");
    expect(summary.buttonTexts).toContain("English");
  });

  test("frontend language dropdown exposes the ui-design locale list", async ({
    page,
  }) => {
    await page.goto(FRONTEND_PATH);
    await page.waitForSelector("[data-testid='topbar-lang-toggle']");
    await page.click("[data-testid='topbar-lang-toggle']");
    await page.waitForSelector("[data-testid='topbar-lang-menu']");

    await expect(page.locator("[data-testid='topbar-lang-menu']")).toBeVisible();
    await expect(page.locator("[data-testid='topbar-lang-option-zh']")).toHaveText(/中文/);
    await expect(page.locator("[data-testid='topbar-lang-option-en']")).toHaveText(/English/);
    await expect(page.locator("[data-testid='topbar-lang-option-en']")).toHaveAttribute("aria-pressed", "true");

    await page.click("[data-testid='topbar-lang-option-zh']");
    await expect(page.locator("[data-testid='topbar-nav-home']")).toHaveText(/首页/);
  });

  test("frontend theme menu exposes the minimal ui-design custom accent picker", async ({
    page,
  }, testInfo) => {
    const prototypeSource = readFileSync(resolve(REPO_ROOT, "ui-design/src/app.jsx"), "utf8");
    expect(prototypeSource).toContain(
      "const AccentPicker = ({ T, lang, dark, value, onChange }) =>",
    );
    expect(prototypeSource).not.toMatch(
      /Reset to theme accent|恢复主题默认色|Drag to apply|拖动应用/,
    );

    await page.goto(FRONTEND_PATH);
    await page.waitForSelector("[data-testid='topbar-theme-button']");
    await page.click("[data-testid='topbar-theme-button']");
    await page.waitForSelector("[data-testid='topbar-theme-menu']");

    const menu = page.locator("[data-testid='topbar-theme-menu']");
    await expect(menu).toBeVisible();
    await expect(page.locator("[data-testid^='topbar-theme-option-']")).toHaveCount(2);
    await expect(page.locator("[data-testid='topbar-theme-option-ocean']")).toBeVisible();
    await expect(page.locator("[data-testid='topbar-theme-option-plum']")).toBeVisible();
    await expect(page.locator("[data-testid='topbar-theme-option-warm']")).toHaveCount(0);
    await expect(page.locator("[data-testid='topbar-theme-option-forest']")).toHaveCount(0);
    await expect(page.locator("[data-testid='topbar-theme-custom-option']")).toHaveText(/Custom/);

    await page.click("[data-testid='topbar-theme-custom-option']");
    await expect(page.locator("[data-testid='topbar-custom-accent-picker']")).toHaveCount(1);
    await expect(page.locator("[data-testid='topbar-custom-accent-hue']")).toHaveCount(1);
    await expect(page.locator("[data-testid='topbar-custom-accent-chroma']")).toHaveCount(1);
    await expect(page.locator("[data-testid='topbar-custom-accent-clear']")).toHaveCount(0);
    await expect(page.locator(".ei-topbar-custom-accent-preview")).toHaveCount(0);
    await expect(page.locator(".ei-topbar-custom-accent-value")).toHaveCount(0);
    await expect(page.getByText(/Reset to theme accent|恢复主题默认色/)).toHaveCount(0);

    const picker = page.locator("[data-testid='topbar-custom-accent-picker']");
    const frontendPicker = await picker.evaluate((element) => {
      const node = element as HTMLElement;
      const style = getComputedStyle(node);
      const rect = node.getBoundingClientRect();
      const menu = node.closest("[data-testid='topbar-theme-menu']") as HTMLElement | null;
      const chroma = node.querySelector(
        "[data-testid='topbar-custom-accent-chroma']",
      )?.parentElement?.parentElement as HTMLElement | null;
      if (!menu || !chroma) throw new Error("minimal frontend picker geometry anchors missing");
      const menuRect = menu.getBoundingClientRect();
      const chromaRect = chroma.getBoundingClientRect();
      return {
        childCount: node.children.length,
        rangeCount: node.querySelectorAll("input[type='range']").length,
        text: node.textContent?.replace(/\s+/g, " ").trim(),
        style: {
          padding: style.padding,
          marginTop: style.marginTop,
          borderTopWidth: style.borderTopWidth,
          borderTopStyle: style.borderTopStyle,
        },
        rect: {
          left: rect.left,
          top: rect.top,
          right: rect.right,
          bottom: rect.bottom,
          width: rect.width,
          height: rect.height,
        },
        menu: {
          left: menuRect.left,
          right: menuRect.right,
          bottom: menuRect.bottom,
        },
        trailingSpace: rect.bottom - chromaRect.bottom,
        pickerOverflow: node.scrollWidth - node.clientWidth,
        documentOverflow: document.documentElement.scrollWidth - window.innerWidth,
        viewport: { width: window.innerWidth, height: window.innerHeight },
      };
    });
    expect(frontendPicker.childCount).toBe(2);
    expect(frontendPicker.rangeCount).toBe(2);
    expect(frontendPicker.text).toMatch(/Hue.*Chroma/);
    expect(frontendPicker.style).toEqual({
      padding: "10px 10px 12px",
      marginTop: "4px",
      borderTopWidth: "1px",
      borderTopStyle: "dotted",
    });
    expect(frontendPicker.trailingSpace).toBeGreaterThanOrEqual(11);
    expect(frontendPicker.trailingSpace).toBeLessThanOrEqual(13);
    expect(frontendPicker.pickerOverflow).toBeLessThanOrEqual(0);
    expect(frontendPicker.documentOverflow).toBeLessThanOrEqual(1);
    expect(frontendPicker.rect.left).toBeGreaterThanOrEqual(frontendPicker.menu.left - 1);
    expect(frontendPicker.rect.right).toBeLessThanOrEqual(frontendPicker.menu.right + 1);
    expect(frontendPicker.rect.bottom).toBeLessThanOrEqual(frontendPicker.menu.bottom + 1);
    expect(frontendPicker.menu.left).toBeGreaterThanOrEqual(-1);
    expect(frontendPicker.menu.right).toBeLessThanOrEqual(frontendPicker.viewport.width + 1);
    expect(frontendPicker.menu.bottom).toBeLessThanOrEqual(frontendPicker.viewport.height + 1);

    const frontendPng = await picker.screenshot();
    expect(frontendPng.length).toBeGreaterThan(500);
    await testInfo.attach(`minimal-accent-picker-frontend-${testInfo.project.name}`, {
      body: frontendPng,
      contentType: "image/png",
    });

    test.setTimeout(45_000);
    await gotoUiDesign(page);
    await page.locator("button[title='Theme']").click();
    await page.locator("button").filter({ hasText: /^Custom/ }).click();
    const prototypeHueLabel = page.locator(".ei-label").filter({ hasText: /^Hue$/ }).last();
    await expect(prototypeHueLabel).toBeVisible();
    const prototypePicker = prototypeHueLabel.locator("..").locator("..");
    const uiDesignPicker = await prototypePicker.evaluate((element) => {
      const node = element as HTMLElement;
      const style = getComputedStyle(node);
      const rect = node.getBoundingClientRect();
      const menu = node.parentElement as HTMLElement | null;
      const chroma = Array.from(node.children).at(-1) as HTMLElement | undefined;
      if (!menu || !chroma) throw new Error("minimal ui-design picker geometry anchors missing");
      const menuRect = menu.getBoundingClientRect();
      const chromaRect = chroma.getBoundingClientRect();
      return {
        childCount: node.children.length,
        rangeCount: node.querySelectorAll("input[type='range']").length,
        text: node.textContent?.replace(/\s+/g, " ").trim(),
        style: {
          padding: style.padding,
          marginTop: style.marginTop,
          borderTopWidth: style.borderTopWidth,
          borderTopStyle: style.borderTopStyle,
        },
        rect: {
          left: rect.left,
          right: rect.right,
          bottom: rect.bottom,
          width: rect.width,
          height: rect.height,
        },
        menu: {
          left: menuRect.left,
          right: menuRect.right,
          bottom: menuRect.bottom,
        },
        trailingSpace: rect.bottom - chromaRect.bottom,
        pickerOverflow: node.scrollWidth - node.clientWidth,
        documentOverflow: document.documentElement.scrollWidth - window.innerWidth,
        viewport: { width: window.innerWidth, height: window.innerHeight },
      };
    });
    expect(uiDesignPicker.childCount).toBe(2);
    expect(uiDesignPicker.rangeCount).toBe(2);
    expect(uiDesignPicker.text).toMatch(/Hue.*Chroma/);
    expect(uiDesignPicker.text).not.toMatch(/oklch|Reset to theme accent/);
    expect(uiDesignPicker.style).toEqual(frontendPicker.style);
    expect(Math.abs(uiDesignPicker.rect.width - frontendPicker.rect.width)).toBeLessThanOrEqual(1);
    expect(Math.abs(uiDesignPicker.rect.height - frontendPicker.rect.height)).toBeLessThanOrEqual(8);
    expect(uiDesignPicker.trailingSpace).toBeGreaterThanOrEqual(11);
    expect(uiDesignPicker.trailingSpace).toBeLessThanOrEqual(13);
    expect(uiDesignPicker.pickerOverflow).toBeLessThanOrEqual(0);
    expect(uiDesignPicker.documentOverflow).toBeLessThanOrEqual(1);
    expect(uiDesignPicker.rect.left).toBeGreaterThanOrEqual(uiDesignPicker.menu.left - 1);
    expect(uiDesignPicker.rect.right).toBeLessThanOrEqual(uiDesignPicker.menu.right + 1);
    expect(uiDesignPicker.rect.bottom).toBeLessThanOrEqual(uiDesignPicker.menu.bottom + 1);
    expect(uiDesignPicker.menu.left).toBeGreaterThanOrEqual(-1);
    expect(uiDesignPicker.menu.right).toBeLessThanOrEqual(uiDesignPicker.viewport.width + 1);
    expect(uiDesignPicker.menu.bottom).toBeLessThanOrEqual(uiDesignPicker.viewport.height + 1);

    const prototypePng = await prototypePicker.screenshot();
    expect(prototypePng.length).toBeGreaterThan(500);
    await testInfo.attach(`minimal-accent-picker-ui-design-${testInfo.project.name}`, {
      body: prototypePng,
      contentType: "image/png",
    });
  });

  test("frontend authenticated user menu matches ui-design dropdown geometry and logout flow", async ({
    page,
  }, testInfo) => {
    assertUiDesignUserMenuSourceLiterals();
    await mockStatefulAuthApis(page);
    await page.goto(FRONTEND_PATH);
    await expect(page.locator("[data-testid='topbar-user-area']")).toHaveAttribute(
      "data-signed-in",
      "false",
    );
    await expect(page.locator("[data-testid='topbar-login']")).toBeVisible();
    await expect(page.locator("[data-testid='topbar-register']")).toHaveCount(0);

    await page.click("[data-testid='topbar-login']");
    await page.fill("[data-testid='auth-login-email']", "alice@example.com");
    await page.click("[data-testid='auth-login-submit-email']");
    await expect(page.locator("[data-testid='route-auth_verify']")).toBeVisible();
    await page.fill("[data-testid='auth-verify-code']", "654321");
    await page.click("[data-testid='auth-verify-submit']");

    await expect(page.locator("[data-testid='topbar-user-area']")).toHaveAttribute(
      "data-signed-in",
      "true",
    );
    const chip = page.locator("[data-testid='topbar-user-chip']");
    await expect(chip).toBeVisible();
    await expect(chip).toHaveText(/Alice Example/);
    await expect(page.locator("[data-testid='topbar-user-avatar']")).toHaveText("AE");
    await expect(page.locator("[data-testid='topbar-user-menu']")).toHaveCount(0);

    await chip.click();
    const menu = page.locator("[data-testid='topbar-user-menu']");
    await expect(menu).toBeVisible();
    await expect(page.locator("[data-testid='topbar-user-menu-header']")).toContainText("Alice Example");
    await expect(page.locator("[data-testid='topbar-user-email']")).toHaveText("ali***@example.com");
    await expect(page.locator("[data-testid='topbar-user-profile']")).toHaveCount(0);
    await expect(page.locator("[data-testid='topbar-user-settings']")).toHaveText(/Settings & privacy/);
    await expect(page.locator("[data-testid='topbar-user-logout']")).toHaveText(/Sign out/);

    const styles = await menu.evaluate((el) => {
      const cs = getComputedStyle(el);
      return {
        position: cs.position,
        minWidth: cs.minWidth,
        padding: cs.padding,
        borderRadius: cs.borderRadius,
        zIndex: cs.zIndex,
        boxShadow: cs.boxShadow,
      };
    });
    expect(styles).toEqual({
      position: "absolute",
      minWidth: "220px",
      padding: "6px",
      borderRadius: "3px",
      zIndex: "40",
      boxShadow: "rgba(20, 15, 10, 0.16) 0px 12px 36px 0px",
    });

    const geometry = await page.evaluate(() => {
      const chipEl = document.querySelector("[data-testid='topbar-user-chip']") as HTMLElement | null;
      const menuEl = document.querySelector("[data-testid='topbar-user-menu']") as HTMLElement | null;
      const controlsEl = document.querySelector("[data-testid='topbar-display-controls']") as HTMLElement | null;
      if (!chipEl || !menuEl || !controlsEl) throw new Error("missing authenticated TopBar geometry anchor");
      const chipRect = chipEl.getBoundingClientRect();
      const menuRect = menuEl.getBoundingClientRect();
      const controlsRect = controlsEl.getBoundingClientRect();
      return {
        chip: {
          left: chipRect.left,
          right: chipRect.right,
          bottom: chipRect.bottom,
          height: chipRect.height,
        },
        menu: {
          left: menuRect.left,
          top: menuRect.top,
          right: menuRect.right,
          bottom: menuRect.bottom,
          width: menuRect.width,
          height: menuRect.height,
        },
        controls: {
          left: controlsRect.left,
          right: controlsRect.right,
          top: controlsRect.top,
          bottom: controlsRect.bottom,
        },
        viewport: {
          width: window.innerWidth,
          height: window.innerHeight,
        },
      };
    });
    expect(Math.abs(geometry.chip.height - 34)).toBeLessThanOrEqual(1);
    expect(Math.abs(geometry.menu.top - geometry.chip.bottom - 6)).toBeLessThanOrEqual(2);
    if (testInfo.project.name === "desktop") {
      expect(Math.abs(geometry.menu.right - geometry.chip.right)).toBeLessThanOrEqual(2);
    } else {
      expect(Math.abs(geometry.menu.left - geometry.chip.left)).toBeLessThanOrEqual(2);
    }
    expect(geometry.menu.width).toBeGreaterThanOrEqual(220);
    expect(geometry.menu.left).toBeGreaterThanOrEqual(-1);
    expect(geometry.menu.right).toBeLessThanOrEqual(geometry.viewport.width + 1);
    expect(geometry.menu.bottom).toBeLessThanOrEqual(geometry.viewport.height + 1);
    if (testInfo.project.name === "desktop") {
      expect(geometry.controls.right).toBeLessThanOrEqual(geometry.chip.left + 1);
    }

    const menuPng = await menu.screenshot();
    expect(menuPng.length).toBeGreaterThan(1000);
    await testInfo.attach(`authenticated-user-menu-${testInfo.project.name}`, {
      body: menuPng,
      contentType: "image/png",
    });

    await page.keyboard.press("Escape");
    await expect(page.locator("[data-testid='topbar-user-menu']")).toHaveCount(0);

    await chip.click();
    await page.click("[data-testid='topbar-user-logout']");
    await expect(page.locator("[data-testid='topbar-user-menu']")).toHaveCount(0);
    await expect(page.locator("[data-testid='route-auth_logout']")).toBeVisible();
    await page.click("[data-testid='auth-logout-confirm']");
    await expect(page.locator("[data-testid='topbar-user-area']")).toHaveAttribute(
      "data-signed-in",
      "false",
    );
    await expect(page.locator("[data-testid='topbar-login']")).toBeVisible();
    await expect(page.locator("[data-testid='topbar-register']")).toHaveCount(0);
  });

  test("ui-design golden preview renders three primary nav buttons with browser-default English labels (D-22)", async ({
    page,
  }) => {
    test.setTimeout(45_000);
    await gotoUiDesign(page);
    const navTexts = await page.$$eval(
      "nav button",
      (els) => els.map((el) => (el.textContent ?? "").replace(/\s+/g, " ").trim()),
    );
    // ui-design renders an `<Icon />` SVG followed by a label; the resulting
    // textContent should end with the label string. We assert each expected
    // label appears as a suffix of one nav button.
    for (const label of PRIMARY_NAV_LABELS_EN) {
      const matched = navTexts.some((text) => text.endsWith(label));
      expect(matched, `ui-design nav must contain a button ending with ${label} (got ${JSON.stringify(navTexts)})`).toBe(true);
    }
  });

  test("frontend TopBar shell height matches desktop source and mobile responsive contract", async ({
    page,
  }, testInfo) => {
    await page.goto(FRONTEND_PATH);
    const frontendHeight = await page.evaluate(() => {
      const el = document.querySelector(
        "[data-testid='app-shell-topbar']",
      ) as HTMLElement | null;
      if (!el) throw new Error("frontend topbar missing");
      return el.getBoundingClientRect().height;
    });

    test.setTimeout(45_000);
    await gotoUiDesign(page);
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

    if (testInfo.project.name === "desktop") {
      // Desktop targets 58px height per ui-design/src/app.jsx TopBar literal.
      expect(uiDesignHeight).toBeCloseTo(58, 0);
      expect(frontendHeight).toBeCloseTo(58, 0);
      expect(Math.abs(frontendHeight - uiDesignHeight)).toBeLessThanOrEqual(1);
    } else {
      // Mobile truth source and formal frontend both use content-driven rows;
      // neither preserves the overflowing fixed-height desktop bar.
      expect(uiDesignHeight).toBeGreaterThanOrEqual(58);
      expect(uiDesignHeight).toBeLessThanOrEqual(150);
      expect(frontendHeight).toBeGreaterThanOrEqual(58);
      expect(frontendHeight).toBeLessThanOrEqual(150);
      expect(Math.abs(frontendHeight - uiDesignHeight)).toBeLessThanOrEqual(1);
    }
  });

  test("frontend TopBar padding-left / padding-right follows desktop and mobile contracts", async ({
    page,
  }, testInfo) => {
    await page.goto(FRONTEND_PATH);
    const padding = await page.evaluate(() => {
      const el = document.querySelector(
        "[data-testid='app-shell-topbar']",
      ) as HTMLElement | null;
      if (!el) throw new Error("frontend topbar missing");
      const cs = getComputedStyle(el);
      return { left: cs.paddingLeft, right: cs.paddingRight };
    });
    if (testInfo.project.name === "desktop") {
      expect(padding.left).toBe("32px");
      expect(padding.right).toBe("32px");
    } else {
      expect(padding.left).toBe("14px");
      expect(padding.right).toBe("14px");
    }
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
        // border-bottom-color resolves to rgb(221, 226, 236) for the ocean/light
        // default (EI_THEMES.ocean.light.rule = #dde2ec, product-scope D-21).
        color: cs.borderBottomColor,
      };
    });
    expect(border.width).toBe("1px");
    expect(border.style).toBe("solid");
    expect(border.color).toBe("rgb(221, 226, 236)");
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
      "workspace",
      "resume_versions",
    ]) {
      const value = await page.getAttribute(
        `[data-testid='topbar-nav-${route}']`,
        "aria-current",
      );
      expect(value).toBeNull();
    }
    await expect(page.locator("[data-testid='topbar-nav-debrief']")).toHaveCount(0);
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
