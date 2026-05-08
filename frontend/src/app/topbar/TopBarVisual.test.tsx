// @vitest-environment jsdom
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { afterEach, describe, expect, it } from "vitest";
import { render, screen, type RenderResult } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactElement } from "react";

import {
  DisplayPreferencesProvider,
  type CustomAccent,
  type Lang,
  type Theme,
} from "../display/DisplayPreferencesProvider";
import { TopBar } from "./TopBar";

const HERE = resolve(__dirname);
const FRONTEND_ROOT = resolve(HERE, "..", "..", "..");
const REPO_ROOT = resolve(HERE, "..", "..", "..", "..");
const TOPBAR_CSS = resolve(HERE, "topbar.css");
const APP_JSX = resolve(REPO_ROOT, "ui-design", "src", "app.jsx");

interface RenderOpts {
  signedIn?: boolean;
  initial?: {
    theme?: Theme;
    dark?: boolean;
    lang?: Lang;
    customAccent?: CustomAccent | null;
  };
}

function renderTopBar(opts: RenderOpts = {}): RenderResult {
  return render(
    <DisplayPreferencesProvider initial={opts.initial}>
      <TopBar
        activeRoute="home"
        onNavigate={() => {}}
        signedIn={opts.signedIn ?? false}
      />
    </DisplayPreferencesProvider>,
  );
}

afterEach(() => {
  document.documentElement.removeAttribute("data-theme");
  document.documentElement.removeAttribute("data-mode");
  document.documentElement.removeAttribute("data-custom-accent");
  document.documentElement.style.removeProperty("--ei-color-accent");
  document.documentElement.style.removeProperty("--ei-color-accent-soft");
});

describe("TopBar shell visual contract (Phase 3.1)", () => {
  it("root header carries ei-shell-topbar className and the documented testid", () => {
    renderTopBar();
    const root = screen.getByTestId("app-shell-topbar");
    expect(root.tagName.toLowerCase()).toBe("header");
    expect(root.className).toMatch(/\bei-shell-topbar\b/);
  });

  it("topbar-primary-nav and topbar-display-controls and topbar-user-area carry semantic classNames", () => {
    renderTopBar();
    expect(screen.getByTestId("topbar-primary-nav").className).toMatch(
      /\bei-topbar-nav\b/,
    );
    expect(screen.getByTestId("topbar-display-controls").className).toMatch(
      /\bei-topbar-controls\b/,
    );
    expect(screen.getByTestId("topbar-user-area").className).toMatch(
      /\bei-topbar-user\b/,
    );
  });

  it("topbar.css defines the ui-design TopBar rhythm (height 58, padding 32, gap 28)", () => {
    const css = readFileSync(TOPBAR_CSS, "utf8");
    expect(css).toMatch(/\.ei-shell-topbar\s*\{[^}]*height:\s*58px/);
    expect(css).toMatch(
      /\.ei-shell-topbar\s*\{[^}]*padding:\s*0\s+var\(--ei-space-8\)/,
    );
    expect(css).toMatch(/\.ei-shell-topbar\s*\{[^}]*gap:\s*var\(--ei-space-7\)/);
    expect(css).toMatch(
      /\.ei-shell-topbar\s*\{[^}]*border-bottom:\s*1px solid var\(--ei-color-rule-strong\)/,
    );
    expect(css).toMatch(
      /\.ei-shell-topbar\s*\{[^}]*background:\s*var\(--ei-color-bg-canvas\)/,
    );
    expect(css).toMatch(/\.ei-shell-topbar\s*\{[^}]*position:\s*sticky/);
    expect(css).toMatch(/\.ei-shell-topbar\s*\{[^}]*top:\s*0/);
    expect(css).toMatch(/\.ei-shell-topbar\s*\{[^}]*z-index:\s*30/);
    expect(css).toMatch(/\.ei-shell-topbar\s*\{[^}]*display:\s*flex/);
    expect(css).toMatch(/\.ei-shell-topbar\s*\{[^}]*align-items:\s*center/);
  });

  it("topbar.css values can be traced back to ui-design/src/app.jsx TopBar literal", () => {
    const app = readFileSync(APP_JSX, "utf8");
    // Extract the TopBar wrapper inline style block from app.jsx.
    expect(app).toContain('borderBottom: `1px solid ${T.rule}`');
    expect(app).toContain("position: \"sticky\"");
    expect(app).toContain("height: 58");
    expect(app).toContain('padding: "0 32px"');
    expect(app).toContain("gap: 28");
    expect(app).toContain("zIndex: 30");
  });

  it("regression: D1 testids and aria-current/aria-pressed contract intact", () => {
    renderTopBar();
    expect(screen.getByTestId("topbar-primary-nav")).toBeInTheDocument();
    expect(screen.getByTestId("topbar-display-controls")).toBeInTheDocument();
    expect(screen.getByTestId("topbar-user-area")).toBeInTheDocument();
    expect(screen.getByTestId("topbar-nav-home")).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(screen.getByTestId("topbar-dark-toggle")).toHaveAttribute(
      "aria-pressed",
      "false",
    );
  });
});

describe("TopBar five-entry + display controls visual (Phase 3.2)", () => {
  it("each primary nav button carries semantic className and ei-text-body typography", () => {
    renderTopBar();
    for (const route of [
      "home",
      "jd_match",
      "workspace",
      "resume_versions",
      "debrief",
    ]) {
      const btn = screen.getByTestId(`topbar-nav-${route}`);
      expect(btn.className).toMatch(/\bei-topbar-nav-button\b/);
      expect(btn.className).toMatch(/\bei-text-body\b/);
    }
  });

  it("active nav button gets [aria-current=page] which CSS styles via ei-topbar-nav-button[aria-current=page]", () => {
    const css = readFileSync(TOPBAR_CSS, "utf8");
    expect(css).toMatch(
      /\.ei-topbar-nav-button\[aria-current="page"\]\s*\{[^}]*background:\s*var\(--ei-color-bg-soft\)/,
    );
    expect(css).toMatch(
      /\.ei-topbar-nav-button\[aria-current="page"\]\s*\{[^}]*color:\s*var\(--ei-color-fg-primary\)/,
    );
  });

  it("display controls replicate ui-design dropdown controls instead of native selects", async () => {
    renderTopBar();
    const user = userEvent.setup();

    expect(screen.queryByTestId("topbar-theme-select")).not.toBeInTheDocument();
    expect(screen.queryByTestId("topbar-lang-select")).not.toBeInTheDocument();
    expect(screen.getByText("EasyInterview")).toBeInTheDocument();
    expect(screen.queryByTestId("topbar-brand-subtitle")).not.toBeInTheDocument();
    expect(screen.getAllByTestId(/^topbar-nav-icon-/)).toHaveLength(5);

    const themeButton = screen.getByTestId("topbar-theme-button");
    expect(themeButton).toHaveClass("ei-topbar-control");
    expect(themeButton).toHaveAttribute("title", "Theme");
    await user.click(themeButton);
    expect(screen.getByTestId("topbar-theme-menu")).toBeInTheDocument();
    expect(screen.getAllByTestId(/^topbar-theme-option-/)).toHaveLength(4);
    expect(screen.getByTestId("topbar-theme-custom-option")).toHaveTextContent(
      "Custom",
    );

    expect(screen.getByTestId("topbar-dark-toggle").className).toMatch(
      /\bei-topbar-dark\b/,
    );
    expect(screen.getByTestId("topbar-dark-toggle")).toHaveTextContent("");
    expect(screen.getByTestId("topbar-lang-toggle").className).toMatch(
      /\bei-topbar-lang\b/,
    );
    expect(screen.getByTestId("topbar-lang-toggle")).toHaveTextContent(
      "English",
    );
    expect(screen.getByTestId("topbar-lang-toggle")).toHaveAttribute(
      "aria-expanded",
      "false",
    );
    await user.click(screen.getByTestId("topbar-lang-toggle"));
    expect(screen.getByTestId("topbar-lang-menu")).toBeInTheDocument();
    expect(screen.getByTestId("topbar-lang-option-en")).toHaveAttribute(
      "aria-pressed",
      "true",
    );
    expect(screen.getByTestId("topbar-lang-option-zh")).toHaveTextContent(
      "中文",
    );
    await user.click(screen.getByTestId("topbar-lang-option-zh"));
    expect(screen.getByTestId("topbar-lang-toggle")).toHaveTextContent(
      "中文",
    );
    expect(screen.getByTestId("topbar-login").className).toMatch(
      /\bei-topbar-auth-login\b/,
    );
    expect(screen.getByTestId("topbar-register").className).toMatch(
      /\bei-topbar-auth-register\b/,
    );
  });

  it("custom accent picker is nested in the ui-design theme menu", async () => {
    renderTopBar();
    const user = userEvent.setup();
    await user.click(screen.getByTestId("topbar-theme-button"));
    const customOption = screen.getByTestId("topbar-theme-custom-option");
    expect(customOption.className).toMatch(/\bei-topbar-theme-option\b/);
    await user.click(customOption);

    expect(document.documentElement).toHaveAttribute(
      "data-custom-accent",
      "active",
    );
    expect(screen.getByTestId("topbar-custom-accent-hue")).toBeInTheDocument();
    expect(
      screen.getByTestId("topbar-custom-accent-chroma"),
    ).toBeInTheDocument();
  });

  it("active customAccent renders the TopBar swatch with oklch inline value", async () => {
    renderTopBar({
      initial: {
        customAccent: { h: 200, c: 0.18 },
      },
    });
    // ui-design only renders the custom row while the theme menu is open.
    const user = userEvent.setup();
    await user.click(screen.getByTestId("topbar-theme-button"));
    const swatch = screen.getByTestId("topbar-custom-accent-swatch");
    expect(swatch.style.background).toMatch(/^oklch\(/);
    expect(
      screen.getByTestId("topbar-theme-custom-option"),
    ).toHaveAttribute("aria-pressed", "true");
  });
});

describe("TopBar i18n regression after visual parity (Phase 3.2)", () => {
  it("switches navigation copy to English when lang=en", async () => {
    render(
      <DisplayPreferencesProvider initial={{ lang: "zh" }}>
        <TopBar activeRoute="home" onNavigate={() => {}} />
      </DisplayPreferencesProvider>,
    );
    const user = userEvent.setup();
    expect(screen.getByTestId("topbar-nav-home")).toHaveTextContent("首页");
    expect(screen.getByTestId("topbar-nav-jd_match")).toHaveTextContent(
      "岗位推荐",
    );

    await user.click(screen.getByTestId("topbar-lang-toggle"));
    await user.click(screen.getByTestId("topbar-lang-option-en"));
    expect(screen.getByTestId("topbar-nav-home")).toHaveTextContent("Home");
    expect(screen.getByTestId("topbar-nav-jd_match")).toHaveTextContent(
      "Job Picks",
    );
  });
});

describe("TopBar inline-style budget (Phase 3.2)", () => {
  const tsx = readFileSync(
    resolve(FRONTEND_ROOT, "src", "app", "topbar", "TopBar.tsx"),
    "utf8",
  );

  it("does not declare inline width / height / padding / gap px literals", () => {
    expect(tsx).not.toMatch(/style=\{\{[^}]*\bpadding:\s*\d+(px|\s)/);
    expect(tsx).not.toMatch(/style=\{\{[^}]*\bgap:\s*\d/);
    expect(tsx).not.toMatch(/style=\{\{[^}]*\bheight:\s*\d/);
  });

  it("renders language options from the i18n locale catalog instead of a two-language TopBar constant", () => {
    expect(tsx).toContain("SUPPORTED_LOCALES.map");
    expect(tsx).not.toMatch(/const\s+LANG_OPTIONS/);
    expect(tsx).not.toMatch(/const\s+LANG_LABELS/);
  });
});
