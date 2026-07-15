// @vitest-environment jsdom
/**
 * Code-level app shell visual-system regression.
 *
 * Truth source: docs/spec/frontend-shell/plans/002-app-shell-visual-system/
 *               bdd-plan.md + bdd-checklist.md.
 *
 * Coverage: DOM anchors, className wiring,
 *     data-attribute flips on theme / dark / customAccent, inline-style
 *     overrides for customAccent oklch swatch, out-of-scope-module negative
 *     assertions, i18n switch, and getComputedStyle for declared CSS
 *     variables (jsdom resolves `:root[data-theme=...][data-mode=...]`
 *     selectors and var() lookups against injected stylesheets).
 */
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
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

describe("app shell visual system", () => {
  it("renders the D2 ei-shell-topbar / ei-screen-shell scaffold with semantic classNames", async () => {
    const client = buildClient();
    const { unmount } = render(
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
      "workspace",
      "resume_versions",
    ]) {
      const navBtn = screen.getByTestId(`topbar-nav-${route}`);
      expect(navBtn.className).toMatch(/\bei-topbar-nav-button\b/);
      expect(navBtn.className).toMatch(/\bei-text-body\b/);
    }
    const user = userEvent.setup();
    await user.click(screen.getByTestId("topbar-lang-toggle"));
    await user.click(screen.getByTestId("topbar-lang-option-zh"));
    expect(screen.getByTestId("topbar-nav-home")).toHaveTextContent("首页");
    await user.click(screen.getByTestId("topbar-lang-toggle"));
    await user.click(screen.getByTestId("topbar-lang-option-en"));
    expect(screen.getByTestId("topbar-nav-home")).toHaveTextContent("Home");
    unmount();
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
    // product-scope D-21 (v2.1): ocean is the default theme.
    expect(root.getAttribute("data-theme")).toBe("ocean");
    expect(root.getAttribute("data-mode")).toBe("light");

    // ocean/light resolves the current formal frontend canvas token.
    expect(
      getComputedStyle(root).getPropertyValue("--ei-color-bg-canvas").trim(),
    ).toBe("#f8fafd");

    const user = userEvent.setup();
    await user.click(screen.getByTestId("topbar-dark-toggle"));
    expect(root.getAttribute("data-mode")).toBe("dark");
    expect(
      getComputedStyle(root).getPropertyValue("--ei-color-fg-primary").trim(),
    ).toBe("#e8edf6");

    await user.click(screen.getByTestId("topbar-theme-button"));
    expect(screen.queryByTestId("topbar-theme-option-warm")).toBeNull();
    expect(screen.queryByTestId("topbar-theme-option-forest")).toBeNull();
    await user.click(screen.getByTestId("topbar-theme-option-plum"));
    expect(root.getAttribute("data-theme")).toBe("plum");
    expect(
      getComputedStyle(root).getPropertyValue("--ei-color-bg-canvas").trim(),
    ).toBe("#15101a");
  });

  it("updates customAccent from hue/chroma and exits only through Ocean or Plum", async () => {
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
    await user.click(screen.getByTestId("topbar-theme-button"));
    await user.click(screen.getByTestId("topbar-theme-custom-option"));
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
    expect(screen.queryByTestId("topbar-custom-accent-clear")).toBeNull();
    expect(screen.queryByText(/恢复主题默认色|Reset to theme accent/)).toBeNull();

    fireEvent.change(screen.getByTestId("topbar-custom-accent-hue"), {
      target: { value: "120" },
    });
    expect(root.style.getPropertyValue("--ei-color-accent")).toBe(
      "oklch(58% 0.160 120.0)",
    );
    fireEvent.change(screen.getByTestId("topbar-custom-accent-chroma"), {
      target: { value: "0.205" },
    });
    expect(root.style.getPropertyValue("--ei-color-accent")).toBe(
      "oklch(58% 0.205 120.0)",
    );

    await user.click(screen.getByTestId("topbar-theme-option-ocean"));
    expect(root.getAttribute("data-theme")).toBe("ocean");
    expect(root.hasAttribute("data-custom-accent")).toBe(false);
    expect(root.style.getPropertyValue("--ei-color-accent")).toBe("");
    expect(root.style.getPropertyValue("--ei-color-accent-soft")).toBe("");

    await user.click(screen.getByTestId("topbar-theme-button"));
    await user.click(screen.getByTestId("topbar-theme-custom-option"));
    expect(root.getAttribute("data-custom-accent")).toBe("active");
    await user.click(screen.getByTestId("topbar-theme-option-plum"));
    expect(root.getAttribute("data-theme")).toBe("plum");
    expect(root.hasAttribute("data-custom-accent")).toBe(false);
    expect(root.style.getPropertyValue("--ei-color-accent")).toBe("");
    expect(root.style.getPropertyValue("--ei-color-accent-soft")).toBe("");
  });

  it("auth_login renders the ei-auth-shell card scaffold + D1 form testids when navigated", async () => {
    const client = buildClient();
    const { unmount } = render(
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
    unmount();
  });

  it("settings / route shell render the ei-screen-shell + ei-screen-card scaffold", async () => {
    const client = buildClient();
    const { unmount: unmountSettings } = render(
      <App
        client={client}
        initialRoute={{ name: "settings", params: {} }}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
      />,
    );
    expect((await screen.findByTestId("route-settings")).className).toMatch(
      /\bei-screen-shell\b/,
    );
    expect(screen.getByTestId("settings-account").className).toMatch(
      /\bei-screen-card\b/,
    );
    unmountSettings();

    render(
      <App
        client={client}
        initialRoute={{ name: "standalone_insight", params: { jobId: "tj-1" } }}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
      />,
    );
    expect(screen.queryByTestId("route-standalone_insight")).not.toBeInTheDocument();
    expect(await screen.findByTestId("route-home")).toBeInTheDocument();
  });

  it("out-of-scope entries (welcome / standalone voice / growth / mistakes / drill) do not flow back", async () => {
    const client = buildClient();
    const { unmount } = render(
      <App
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      />,
    );
    expect(screen.queryByTestId("route-welcome")).not.toBeInTheDocument();
    for (const outOfScope of ["welcome", "growth", "mistakes", "drill", "voice"]) {
      expect(
        screen.queryByTestId(`topbar-nav-${outOfScope}`),
      ).not.toBeInTheDocument();
      expect(screen.queryByTestId(`route-${outOfScope}`)).not.toBeInTheDocument();
    }
    // Settings must not surface out-of-scope module copy.
    const html = document.documentElement.outerHTML;
    expect(html).not.toMatch(/错题本|成长中心|经历库|目标角色|技能标签/);
    unmount();
  });

});
