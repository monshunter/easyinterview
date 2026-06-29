// @vitest-environment jsdom
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";
import { fireEvent, render } from "@testing-library/react";

import { DisplayPreferencesProvider } from "../display/DisplayPreferencesProvider";
import { PlaceholderScreen } from "./PlaceholderScreen";
import { SettingsScreen } from "./SettingsScreen";

const HERE = resolve(__dirname);
const SCREENS_CSS = resolve(HERE, "screens.css");

function withProvider(node: React.ReactElement) {
  return <DisplayPreferencesProvider>{node}</DisplayPreferencesProvider>;
}

describe("Settings shell visual contract (Phase 5.1 / Phase 12.2)", () => {
  it("renders ei-screen-shell with a two-tab rail and profile-tab ei-screen-card sections", () => {
    const { container } = render(
      withProvider(<SettingsScreen route={{ name: "settings", params: {} }} />),
    );
    const root = container.querySelector("[data-testid='route-settings']");
    expect(root).toBeTruthy();
    expect(root!.className).toMatch(/\bei-screen-shell\b/);
    const heading = root!.querySelector("h1");
    expect(heading?.className).toMatch(/\bei-text-display\b/);

    // D-21: exactly two tabs — profile and privacy & data.
    expect(
      container.querySelector("[data-testid='settings-tab-profile']"),
    ).toBeTruthy();
    expect(
      container.querySelector("[data-testid='settings-tab-privacy']"),
    ).toBeTruthy();
    expect(
      container.querySelectorAll("[data-testid^='settings-tab-']"),
    ).toHaveLength(2);

    // Default profile tab carries account / login-security / font preset /
    // product info cards; the privacy card lives on the privacy tab.
    for (const sectionId of [
      "settings-account",
      "settings-login-security",
      "settings-font-preset",
      "settings-app-info",
    ]) {
      const section = container.querySelector(`[data-testid='${sectionId}']`);
      expect(section, `${sectionId} missing`).toBeTruthy();
      expect(section!.className).toMatch(/\bei-screen-card\b/);
    }
    expect(
      container.querySelector("[data-testid='settings-privacy']"),
    ).toBeFalsy();

    // D-21: P1 placeholder tabs are removed for good.
    expect(
      container.querySelector(
        "[data-testid='settings-notifications-placeholder']",
      ),
    ).toBeFalsy();
    expect(
      container.querySelector(
        "[data-testid='settings-subscription-placeholder']",
      ),
    ).toBeFalsy();

    // D-16: login security only states the passwordless method.
    const security = container.querySelector(
      "[data-testid='settings-login-security']",
    );
    expect(security!.textContent).toMatch(
      /邮箱验证码 · 无密码|Email sign-in code · no password/,
    );
    expect(security!.textContent).not.toMatch(
      /密码（|两步验证|Two-step verification/,
    );
  });

  it("shows the privacy & data card when the privacy tab is selected", () => {
    const { container } = render(
      withProvider(<SettingsScreen route={{ name: "settings", params: {} }} />),
    );
    fireEvent.click(
      container.querySelector("[data-testid='settings-tab-privacy']")!,
    );
    const privacy = container.querySelector(
      "[data-testid='settings-privacy']",
    );
    expect(privacy).toBeTruthy();
    expect(privacy!.className).toMatch(/\bei-screen-card\b/);
    expect(
      container.querySelector("[data-testid='settings-account']"),
    ).toBeFalsy();
  });

  it("rejects retired Growth / Experiences / Mistakes / Drill / 独立 voice copy and testid", () => {
    const { container } = render(
      withProvider(<SettingsScreen route={{ name: "settings", params: {} }} />),
    );
    const html = container.innerHTML;
    for (const banned of [
      "growth",
      "experiences",
      "mistakes",
      "drill",
      "voice",
    ]) {
      expect(
        new RegExp(`data-testid=["']settings-${banned}["']`).test(html),
        `legacy testid settings-${banned} must not appear`,
      ).toBe(false);
    }
    expect(html).not.toMatch(/错题本|成长中心|经历库|目标角色|技能标签/);
  });
});

describe("PlaceholderScreen card skeleton (Phase 5.2)", () => {
  it("renders a card skeleton (title + description + skeleton stripes) for D2-D6 routes", () => {
    const { container } = render(
      withProvider(
        <PlaceholderScreen
          route={{ name: "workspace", params: { jobId: "tj-1" } }}
        />,
      ),
    );
    const root = container.querySelector("[data-testid='route-workspace']");
    expect(root).toBeTruthy();
    expect(root!.className).toMatch(/\bei-screen-shell\b/);
    expect(root!.getAttribute("data-route-name")).toBe("workspace");
    expect(root!.getAttribute("data-route-params")).toBe(
      JSON.stringify({ jobId: "tj-1" }),
    );

    const heading = root!.querySelector("h1");
    expect(heading?.className).toMatch(/\bei-text-display\b/);
    const card = root!.querySelector(".ei-screen-card");
    expect(card).toBeTruthy();
    expect(card!.querySelector(".ei-skeleton-stripe")).toBeTruthy();
  });

  it("falls back to placeholder.default copy for retained placeholder routes", () => {
    const { container } = render(
      withProvider(
        <PlaceholderScreen
          route={{ name: "parse", params: {} }}
        />,
      ),
    );
    expect(container.querySelector("h1")?.textContent).toMatch(/Route shell|页面壳/);
  });
});

describe("screens.css visual rhythm (Phase 5.1 + 5.2)", () => {
  const css = readFileSync(SCREENS_CSS, "utf8");

  it("declares ei-screen-shell layout (max-width / padding / gap)", () => {
    expect(css).toMatch(/\.ei-screen-shell\s*\{[^}]*max-width:\s*1160px/);
    expect(css).toMatch(/\.ei-screen-shell\s*\{[^}]*padding:\s*54px\s+48px\s+96px/);
    expect(css).toMatch(/\.ei-screen-shell\s*\{[^}]*display:\s*flex/);
    expect(css).toMatch(
      /\.ei-screen-shell\s*\{[^}]*flex-direction:\s*column/,
    );
    expect(css).toMatch(/\.ei-screen-shell\s*\{[^}]*gap:\s*var\(--ei-space-7\)/);
  });

  it("ei-screen-card carries bg-card + rule + radius and section padding", () => {
    expect(css).toMatch(
      /\.ei-screen-card\s*\{[^}]*background:\s*var\(--ei-color-bg-card\)/,
    );
    expect(css).toMatch(
      /\.ei-screen-card\s*\{[^}]*border:\s*1px solid var\(--ei-color-rule-strong\)/,
    );
    expect(css).toMatch(
      /\.ei-screen-card\s*\{[^}]*border-radius:\s*var\(--ei-radius-md\)/,
    );
    expect(css).toMatch(/\.ei-screen-card\s*\{[^}]*padding:\s*28px/);
  });

  it("ei-skeleton-stripe defines striped placeholder background", () => {
    expect(css).toMatch(
      /\.ei-skeleton-stripe\s*\{[^}]*background:\s*repeating-linear-gradient/,
    );
  });

  it("global.css imports screens.css", () => {
    const globalCss = readFileSync(
      resolve(HERE, "..", "theme", "global.css"),
      "utf8",
    );
    expect(globalCss).toMatch(/@import\s+["']\.\.\/screens\/screens\.css["'];/);
  });
});
