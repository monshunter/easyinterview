// @vitest-environment jsdom
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import {
  afterEach,
  beforeEach,
  describe,
  expect,
  it,
} from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { FC } from "react";

import {
  DisplayPreferencesProvider,
  useDisplayPreferences,
} from "./DisplayPreferencesProvider";

const THEMES_CSS = readFileSync(
  resolve(__dirname, "..", "theme", "themes.css"),
  "utf8",
);

let injectedStyle: HTMLStyleElement | null = null;

function injectThemesCss() {
  injectedStyle = document.createElement("style");
  injectedStyle.id = "ei-themes-css-test";
  injectedStyle.textContent = THEMES_CSS;
  document.head.appendChild(injectedStyle);
}

function removeThemesCss() {
  injectedStyle?.remove();
  injectedStyle = null;
}

const Probe: FC = () => {
  const prefs = useDisplayPreferences();
  return (
    <div>
      <button
        type="button"
        data-testid="probe-set-theme-plum"
        onClick={() => prefs.setTheme("plum")}
      />
      <button
        type="button"
        data-testid="probe-set-dark-on"
        onClick={() => prefs.setDark(true)}
      />
      <button
        type="button"
        data-testid="probe-set-dark-off"
        onClick={() => prefs.setDark(false)}
      />
      <button
        type="button"
        data-testid="probe-set-custom-accent"
        onClick={() => prefs.setCustomAccent({ h: 120, c: 0.18 })}
      />
      <button
        type="button"
        data-testid="probe-clear-custom-accent"
        onClick={() => prefs.setCustomAccent(null)}
      />
    </div>
  );
};

describe("DisplayPreferencesProvider root-element wiring (Phase 1.2)", () => {
  beforeEach(() => {
    injectThemesCss();
    document.documentElement.removeAttribute("data-theme");
    document.documentElement.removeAttribute("data-mode");
    document.documentElement.removeAttribute("data-custom-accent");
    document.documentElement.style.removeProperty("--ei-color-accent");
    document.documentElement.style.removeProperty("--ei-color-accent-soft");
  });

  afterEach(() => {
    removeThemesCss();
  });

  it("applies data-theme / data-mode / data-custom-accent attributes on mount", () => {
    render(
      <DisplayPreferencesProvider>
        <Probe />
      </DisplayPreferencesProvider>,
    );
    // product-scope D-21 (v2.1): ocean is the default theme.
    expect(document.documentElement.getAttribute("data-theme")).toBe("ocean");
    expect(document.documentElement.getAttribute("data-mode")).toBe("light");
    expect(document.documentElement.getAttribute("data-custom-accent")).toBe(
      null,
    );
  });

  it("flips data-theme attribute and computed --ei-color-bg-canvas when theme changes", async () => {
    render(
      <DisplayPreferencesProvider>
        <Probe />
      </DisplayPreferencesProvider>,
    );
    const before = getComputedStyle(document.documentElement)
      .getPropertyValue("--ei-color-bg-canvas")
      .trim();
    expect(before).toBe("#f8fafd");

    await userEvent.click(screen.getByTestId("probe-set-theme-plum"));
    expect(document.documentElement.getAttribute("data-theme")).toBe("plum");
    const after = getComputedStyle(document.documentElement)
      .getPropertyValue("--ei-color-bg-canvas")
      .trim();
    expect(after).toBe("#fcf8fa");
  });

  it("flips data-mode attribute and computed --ei-color-fg-primary when dark toggles", async () => {
    render(
      <DisplayPreferencesProvider>
        <Probe />
      </DisplayPreferencesProvider>,
    );
    const lightFg = getComputedStyle(document.documentElement)
      .getPropertyValue("--ei-color-fg-primary")
      .trim();
    expect(lightFg).toBe("#141821");

    await userEvent.click(screen.getByTestId("probe-set-dark-on"));
    expect(document.documentElement.getAttribute("data-mode")).toBe("dark");
    const darkFg = getComputedStyle(document.documentElement)
      .getPropertyValue("--ei-color-fg-primary")
      .trim();
    expect(darkFg).toBe("#e8edf6");

    await userEvent.click(screen.getByTestId("probe-set-dark-off"));
    expect(document.documentElement.getAttribute("data-mode")).toBe("light");
  });

  it("activates customAccent inline overrides limited to accent / accent-soft", async () => {
    render(
      <DisplayPreferencesProvider>
        <Probe />
      </DisplayPreferencesProvider>,
    );
    expect(
      document.documentElement.style.getPropertyValue("--ei-color-accent"),
    ).toBe("");

    await userEvent.click(screen.getByTestId("probe-set-custom-accent"));
    expect(
      document.documentElement.getAttribute("data-custom-accent"),
    ).toBe("active");
    const accent = document.documentElement.style.getPropertyValue(
      "--ei-color-accent",
    );
    const accentSoft = document.documentElement.style.getPropertyValue(
      "--ei-color-accent-soft",
    );
    expect(accent).toMatch(/^oklch\(58% 0\.180 120\.0\)/);
    expect(accentSoft).toMatch(/^oklch\(92%/);
    // Other base palette tokens must NOT be overridden inline.
    expect(
      document.documentElement.style.getPropertyValue(
        "--ei-color-bg-canvas",
      ),
    ).toBe("");
    expect(
      document.documentElement.style.getPropertyValue(
        "--ei-color-fg-primary",
      ),
    ).toBe("");

    await userEvent.click(screen.getByTestId("probe-clear-custom-accent"));
    expect(
      document.documentElement.getAttribute("data-custom-accent"),
    ).toBe(null);
    expect(
      document.documentElement.style.getPropertyValue("--ei-color-accent"),
    ).toBe("");
    expect(
      document.documentElement.style.getPropertyValue(
        "--ei-color-accent-soft",
      ),
    ).toBe("");
  });

  it("re-evaluates customAccent oklch lightness when dark mode toggles", async () => {
    render(
      <DisplayPreferencesProvider>
        <Probe />
      </DisplayPreferencesProvider>,
    );
    await userEvent.click(screen.getByTestId("probe-set-custom-accent"));
    expect(
      document.documentElement.style.getPropertyValue("--ei-color-accent"),
    ).toContain("58%");

    await userEvent.click(screen.getByTestId("probe-set-dark-on"));
    expect(
      document.documentElement.style.getPropertyValue("--ei-color-accent"),
    ).toContain("68%");
    expect(
      document.documentElement.style.getPropertyValue(
        "--ei-color-accent-soft",
      ),
    ).toContain("28%");
  });
});
