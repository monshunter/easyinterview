// @vitest-environment jsdom
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";
import { render } from "@testing-library/react";

import { DisplayPreferencesProvider } from "../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../navigation/NavigationProvider";
import { RouteShellScreen } from "./RouteShellScreen";
import { SettingsScreen } from "./SettingsScreen";

const HERE = resolve(__dirname);
const SCREENS_CSS = resolve(HERE, "screens.css");
const FRONTEND_README = resolve(HERE, "..", "..", "..", "README.md");

function withProvider(node: React.ReactElement) {
  return (
    <DisplayPreferencesProvider>
      <NavigationProvider value={{ navigate: () => {}, replaceRoute: () => {} }}>
        {node}
      </NavigationProvider>
    </DisplayPreferencesProvider>
  );
}

describe("Settings shell visual contract (Phase 5.1 / Phase 12.2)", () => {
  it("renders one account and privacy page without a tab rail or static prototype cards", () => {
    const { container } = render(
      withProvider(<SettingsScreen route={{ name: "settings", params: {} }} />),
    );
    const root = container.querySelector("[data-testid='route-settings']");
    expect(root).toBeTruthy();
    expect(root!.className).toMatch(/\bei-screen-shell\b/);
    const heading = root!.querySelector("h1");
    expect(heading?.className).toMatch(/\bei-text-display\b/);

    expect(container.querySelectorAll("[data-testid^='settings-tab-']")).toHaveLength(0);
    for (const sectionId of ["settings-account", "settings-privacy"]) {
      const section = container.querySelector(`[data-testid='${sectionId}']`);
      expect(section, `${sectionId} missing`).toBeTruthy();
      expect(section!.className).toMatch(/\bei-screen-card\b/);
    }
    for (const removed of [
      "settings-login-security",
      "settings-font-preset",
      "settings-app-info",
    ]) {
      expect(container.querySelector(`[data-testid='${removed}']`)).toBeFalsy();
    }
  });

  it("renders the screenshot-aligned header and three horizontal function cards", () => {
    const { container } = render(
      withProvider(<SettingsScreen route={{ name: "settings", params: {} }} />),
    );
    const root = container.querySelector("[data-testid='route-settings']")!;
    expect(root.querySelector(".ei-settings-header-copy")).toBeTruthy();
    const headerArt = root.querySelector("[data-testid='settings-header-art']");
    expect(headerArt).toHaveClass("ei-settings-header-art");
    expect(headerArt).toHaveAttribute("aria-hidden", "true");
    expect(headerArt).toHaveAttribute("data-settings-art", "security-profile");
    for (const layer of ["window", "profile", "chart", "lock", "shield"]) {
      expect(
        headerArt?.querySelector(`[data-settings-art-layer='${layer}']`),
        `settings Header art is missing the ${layer} layer`,
      ).toBeTruthy();
    }
    expect(
      headerArt?.querySelectorAll("[data-settings-art-layer='sparkle']"),
    ).toHaveLength(2);
    expect(headerArt?.querySelector("[data-settings-art-layer='mountains']")).toBeFalsy();
    for (const sectionId of [
      "settings-appearance",
      "settings-account",
      "settings-privacy",
    ]) {
      const section = root.querySelector(`[data-testid='${sectionId}']`)!;
      expect(section.className).toMatch(/\bei-settings-card\b/);
      expect(section.querySelector(".ei-settings-card-icon")).toBeTruthy();
      expect(section.querySelector(".ei-settings-card-body")).toBeTruthy();
    }
  });

  it("rejects out-of-scope Growth / Experiences / Mistakes / Drill / 独立 voice copy and testid", () => {
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
        `out-of-scope testid settings-${banned} must not appear`,
      ).toBe(false);
    }
    expect(html).not.toMatch(/错题本|成长中心|经历库|目标角色|技能标签/);
  });
});

describe("RouteShellScreen card skeleton (Phase 5.2)", () => {
  it("renders a card skeleton (title + description + route stripe) for retained route-shell routes", () => {
    const { container } = render(
      withProvider(
        <RouteShellScreen
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
    const stripe = card!.querySelector(".ei-skeleton-stripe");
    expect(stripe).toBeTruthy();
    expect(stripe!.textContent).toBe("route shell");
    expect(card!.textContent).not.toContain(["D2", "D6"].join("-"));
  });

  it("falls back to routeShell.default copy for retained route-shell routes", () => {
    const { container } = render(
      withProvider(
        <RouteShellScreen
          route={{ name: "parse", params: {} }}
        />,
      ),
    );
    expect(container.querySelector("h1")?.textContent).toMatch(/Page|页面壳/);
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

  it("ei-skeleton-stripe defines striped skeleton pattern", () => {
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

  it("does not keep or document a screen-card grid without a DOM consumer", () => {
    expect(css).not.toMatch(/\.ei-screen-card-grid\s*\{/);
    expect(readFileSync(FRONTEND_README, "utf8")).not.toContain(
      "ei-screen-card-grid",
    );
  });

  it("stacks settings values and privacy actions at the phone breakpoint", () => {
    expect(css).toMatch(
      /@media \(max-width: 600px\)[\s\S]*\.ei-settings-value-row,[\s\S]*\.ei-settings-privacy-row\s*\{[^}]*grid-template-columns:\s*minmax\(0, 1fr\)/,
    );
    expect(css).toMatch(
      /@media \(max-width: 600px\)[\s\S]*\.ei-settings-value-row dd\s*\{[^}]*text-align:\s*left/,
    );
  });

  it("defines the 1372px Settings page and horizontal card composition", () => {
    expect(css).toMatch(/\.ei-settings-screen\s*\{[^}]*max-width:\s*1372px/);
    expect(css).toMatch(/\.ei-settings-screen\s*>\s*\.ei-settings-header\s*\{[^}]*display:\s*grid/);
    expect(css).toMatch(/\.ei-settings-card\s*\{[^}]*display:\s*grid/);
    expect(css).toMatch(/\.ei-settings-card\s*\{[^}]*grid-template-columns:\s*64px\s+minmax\(0,\s*1fr\)/);
  });

  it("paints the Settings security illustration as layered theme-aware surfaces", () => {
    expect(css).toMatch(
      /\.ei-settings-header-art__window-frame\s*\{[^}]*fill:\s*color-mix\([^}]*var\(--ei-color-accent-soft\)/,
    );
    expect(css).toMatch(
      /\.ei-settings-header-art__lock-tile\s*\{[^}]*fill:\s*color-mix\([^}]*var\(--ei-color-accent\)/,
    );
    expect(css).toMatch(
      /\.ei-settings-header-art__shield-body\s*\{[^}]*fill:\s*color-mix\([^}]*var\(--ei-color-accent\)/,
    );
    expect(css).toMatch(
      /\.ei-settings-header-art__(?:window-frame|lock-tile)\s*\{[^}]*filter:\s*drop-shadow/,
    );
    expect(css).not.toMatch(
      /\.ei-settings-header-art\s*\{[^}]*opacity:\s*0\.22/,
    );
    expect(css).toMatch(
      /@media \(max-width: 720px\)[\s\S]*?\.ei-settings-header-art\s*\{[^}]*display:\s*none/,
    );
  });

  it("stacks the custom theme editor below persistent choices with informative color tracks", () => {
    expect(css).toMatch(
      /\[data-testid="settings-appearance"\]\s+\.ei-settings-theme-editor\s*\{[^}]*grid-column:\s*2\s*\/\s*4[^}]*grid-row:\s*1\s*\/\s*span\s*2/,
    );
    expect(css).toMatch(/\.ei-settings-theme-editor\s*\{[^}]*display:\s*grid[^}]*gap:/);
    expect(css).toMatch(
      /\.ei-settings-theme-primary-row\s*\{[^}]*display:\s*flex[^}]*align-items:\s*center[^}]*justify-content:\s*space-between/,
    );
    expect(css).not.toMatch(
      /\[data-testid="settings-appearance"\]\s+\.ei-settings-actions\s*\{/,
    );
    expect(css).not.toMatch(
      /\.ei-settings-theme-options,\s*\n\[data-testid="settings-appearance"\]\s+\.ei-settings-custom-accent/,
    );
    expect(css).toMatch(
      /\.ei-settings-accent-range--hue\s*\{[^}]*linear-gradient\([^}]*#f45b69[^}]*#946bff[^}]*#f45b69/,
    );
    expect(css).toMatch(
      /\.ei-settings-accent-range--chroma\s*\{[^}]*linear-gradient\([^}]*oklch\([^}]*var\(--ei-settings-accent-hue\)[^}]*oklch\([^}]*var\(--ei-settings-accent-hue\)/,
    );
  });
});
