// @vitest-environment jsdom
/**
 * E2E.P0.005 — App shell visual system smoke + ui-design parity gate.
 *
 * Truth source: docs/spec/frontend-shell/plans/002-app-shell-visual-system/
 *               bdd-plan.md + bdd-checklist.md.
 *
 * Coverage classes (delegated):
 *   - **Verified here (vitest + jsdom)**: DOM anchors, className wiring,
 *     data-attribute flips on theme / dark / customAccent, inline-style
 *     overrides for customAccent oklch swatch, retired-module negative
 *     assertions, i18n switch, and getComputedStyle for declared CSS
 *     variables (jsdom resolves `:root[data-theme=...][data-mode=...]`
 *     selectors and var() lookups against injected stylesheets).
 *   - **Deferred to a Playwright follow-up**: bounding-box diff, viewport
 *     overlap detection, and screenshot diff against `ui-design/index.html`.
 *     The scenario README documents the gap and recommended scaffold; the
 *     scope of this scenario is bounded to what jsdom can verify so a
 *     phase-commit is possible without browser binaries.
 */
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import getMeFixture from "../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import { EasyInterviewClient } from "../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../api/mockTransport";
import { App } from "../App";

const HERE = resolve(__dirname);
const THEME_DIR = resolve(HERE, "..", "theme");
const TOPBAR_CSS_PATH = resolve(HERE, "..", "topbar", "topbar.css");
const AUTH_CSS_PATH = resolve(HERE, "..", "auth", "auth.css");
const SCREENS_CSS_PATH = resolve(HERE, "..", "screens", "screens.css");

const STYLES: string[] = [
  readFileSync(resolve(THEME_DIR, "themes.css"), "utf8"),
  readFileSync(resolve(THEME_DIR, "typography.css"), "utf8"),
  readFileSync(TOPBAR_CSS_PATH, "utf8"),
  readFileSync(AUTH_CSS_PATH, "utf8"),
  readFileSync(SCREENS_CSS_PATH, "utf8"),
];

let injected: HTMLStyleElement[] = [];

function buildClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([getRuntimeConfigFixture, getMeFixture]),
    ),
  });
}

beforeEach(() => {
  injected = STYLES.map((css) => {
    const node = document.createElement("style");
    node.textContent = css;
    document.head.appendChild(node);
    return node;
  });
});

afterEach(() => {
  for (const node of injected) {
    node.remove();
  }
  injected = [];
  document.documentElement.removeAttribute("data-theme");
  document.documentElement.removeAttribute("data-mode");
  document.documentElement.removeAttribute("data-custom-accent");
  document.documentElement.style.removeProperty("--ei-color-accent");
  document.documentElement.style.removeProperty("--ei-color-accent-soft");
});

describe("E2E.P0.005 app shell visual system smoke", () => {
  it("renders the D2 ei-shell-topbar / ei-screen-shell scaffold with ui-design native classNames", async () => {
    const client = buildClient();
    render(
      <App
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      />,
    );

    // TopBar shell anchors + className wiring.
    const topbar = screen.getByTestId("app-shell-topbar");
    expect(topbar.tagName.toLowerCase()).toBe("header");
    expect(topbar.className).toMatch(/\bei-shell-topbar\b/);
    expect(screen.getByTestId("topbar-primary-nav").className).toMatch(
      /\bei-topbar-nav\b/,
    );
    expect(screen.getByTestId("topbar-display-controls").className).toMatch(
      /\bei-topbar-controls\b/,
    );
    expect(screen.getByTestId("topbar-user-area").className).toMatch(
      /\bei-topbar-user\b/,
    );
    for (const route of [
      "home",
      "jd_match",
      "workspace",
      "resume_versions",
      "debrief",
    ]) {
      const navBtn = screen.getByTestId(`topbar-nav-${route}`);
      expect(navBtn.className).toMatch(/\bei-topbar-nav-button\b/);
      expect(navBtn.className).toMatch(/\bei-text-body\b/);
    }
  });

  it("flips :root[data-theme][data-mode] and resolves CSS variables on theme + dark switch", async () => {
    const client = buildClient();
    render(
      <App
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      />,
    );
    const root = document.documentElement;
    expect(root.getAttribute("data-theme")).toBe("warm");
    expect(root.getAttribute("data-mode")).toBe("light");

    // warm/light → bg-canvas resolves to ui-design EI_THEMES.warm.light.bg.
    expect(
      getComputedStyle(root).getPropertyValue("--ei-color-bg-canvas").trim(),
    ).toBe("#fdfcf8");

    const user = userEvent.setup();
    await user.click(screen.getByTestId("topbar-dark-toggle"));
    expect(root.getAttribute("data-mode")).toBe("dark");
    expect(
      getComputedStyle(root).getPropertyValue("--ei-color-fg-primary").trim(),
    ).toBe("#f5f0e4");

    await user.selectOptions(screen.getByTestId("topbar-theme-select"), "ocean");
    expect(root.getAttribute("data-theme")).toBe("ocean");
    expect(
      getComputedStyle(root).getPropertyValue("--ei-color-bg-canvas").trim(),
    ).toBe("#0c0f17");
  });

  it("activates customAccent overlay → only --ei-color-accent / -soft are inline-overridden", async () => {
    const client = buildClient();
    render(
      <App
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      />,
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("topbar-custom-accent-button"));
    const root = document.documentElement;
    expect(root.getAttribute("data-custom-accent")).toBe("active");

    const accent = root.style.getPropertyValue("--ei-color-accent");
    const accentSoft = root.style.getPropertyValue("--ei-color-accent-soft");
    expect(accent).toMatch(/^oklch\(58%/);
    expect(accentSoft).toMatch(/^oklch\(92%/);

    // Base palette tokens MUST NOT be overridden by the custom accent overlay.
    expect(root.style.getPropertyValue("--ei-color-bg-canvas")).toBe("");
    expect(root.style.getPropertyValue("--ei-color-fg-primary")).toBe("");

    expect(screen.getByTestId("topbar-custom-accent-hue")).toBeInTheDocument();
    expect(
      screen.getByTestId("topbar-custom-accent-chroma"),
    ).toBeInTheDocument();
  });

  it("auth_login renders the ei-auth-shell card scaffold + D1 form testids when navigated", async () => {
    const client = buildClient();
    render(
      <App
        client={client}
        initialRoute={{ name: "auth_login", params: {} }}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      />,
    );
    const root = screen.getByTestId("route-auth_login");
    expect(root.className).toMatch(/\bei-auth-shell\b/);
    expect(root.querySelector(".ei-auth-side")).toBeTruthy();
    expect(root.querySelector(".ei-auth-card")).toBeTruthy();
    expect(screen.getByTestId("auth-login-email-form").className).toMatch(
      /\bei-auth-form\b/,
    );
    expect(screen.getByTestId("auth-login-submit-email").className).toMatch(
      /\bei-auth-cta\b/,
    );
    expect(
      screen.getByTestId("auth-login-password-stub"),
    ).toBeInTheDocument();
    expect(screen.getByTestId("auth-login-oauth-stub")).toBeInTheDocument();
  });

  it("profile / settings / placeholder render the ei-screen-shell + ei-screen-card scaffold", async () => {
    const client = buildClient();
    const { unmount: unmountProfile } = render(
      <App
        client={client}
        initialRoute={{ name: "profile", params: {} }}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      />,
    );
    expect(screen.getByTestId("route-profile").className).toMatch(
      /\bei-screen-shell\b/,
    );
    expect(
      screen.getByTestId("profile-identity-summary").className,
    ).toMatch(/\bei-screen-card\b/);
    unmountProfile();

    const { unmount: unmountSettings } = render(
      <App
        client={client}
        initialRoute={{ name: "settings", params: {} }}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      />,
    );
    expect(screen.getByTestId("route-settings").className).toMatch(
      /\bei-screen-shell\b/,
    );
    expect(screen.getByTestId("settings-account").className).toMatch(
      /\bei-screen-card\b/,
    );
    unmountSettings();

    render(
      <App
        client={client}
        initialRoute={{ name: "company_intel", params: { jobId: "tj-1" } }}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      />,
    );
    expect(screen.getByTestId("route-company_intel").className).toMatch(
      /\bei-screen-shell\b/,
    );
    expect(
      screen.getByTestId("route-company_intel").querySelector(
        ".ei-screen-card",
      ),
    ).toBeTruthy();
    expect(
      screen.getByTestId("route-company_intel").querySelector(
        ".ei-skeleton-stripe",
      ),
    ).toBeTruthy();
  });

  it("legacy entries (welcome / standalone voice / growth / mistakes / drill) do not flow back", async () => {
    const client = buildClient();
    render(
      <App
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      />,
    );
    expect(screen.queryByTestId("route-welcome")).not.toBeInTheDocument();
    for (const legacy of ["welcome", "growth", "mistakes", "drill", "voice"]) {
      expect(
        screen.queryByTestId(`topbar-nav-${legacy}`),
      ).not.toBeInTheDocument();
      expect(screen.queryByTestId(`route-${legacy}`)).not.toBeInTheDocument();
    }
    // Settings / profile must not surface retired-module copy.
    const html = document.documentElement.outerHTML;
    expect(html).not.toMatch(/错题本|成长中心|经历库|目标角色|技能标签/);
  });

  it("ui-design source files (app.jsx + screen-auth.jsx + screen-profile.jsx) carry the literal values transcribed into D2 CSS", () => {
    const repoRoot = resolve(HERE, "..", "..", "..", "..");
    const appJsx = readFileSync(
      resolve(repoRoot, "ui-design", "src", "app.jsx"),
      "utf8",
    );
    const authJsx = readFileSync(
      resolve(repoRoot, "ui-design", "src", "screen-auth.jsx"),
      "utf8",
    );
    const profileJsx = readFileSync(
      resolve(repoRoot, "ui-design", "src", "screen-profile.jsx"),
      "utf8",
    );

    // TopBar literals
    expect(appJsx).toContain("height: 58");
    expect(appJsx).toContain('padding: "0 32px"');
    expect(appJsx).toContain("gap: 28");
    expect(appJsx).toContain("zIndex: 30");
    // Auth shell literals
    expect(authJsx).toContain("maxWidth: 1160");
    expect(authJsx).toContain('padding: "54px 48px 96px"');
    expect(authJsx).toContain('gridTemplateColumns: "0.88fr 1.12fr"');
    expect(authJsx).toContain("gap: 44");
    // Profile screen literals
    expect(profileJsx).toContain("padding: 28");
  });
});
