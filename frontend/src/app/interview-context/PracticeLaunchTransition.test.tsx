// @vitest-environment jsdom

import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { DisplayPreferencesProvider } from "../display/DisplayPreferencesProvider";
import { PracticeLaunchTransition } from "./PracticeLaunchTransition";

describe("PracticeLaunchTransition", () => {
  it("announces an honest busy state, blocks the background, and restores it on unmount", () => {
    const { unmount } = render(
      <DisplayPreferencesProvider>
        <div data-testid="app-root">
          <header>
            <button type="button" data-testid="topbar-control">
              Settings
            </button>
          </header>
          <main data-testid="app-main">
            <PracticeLaunchTransition />
          </main>
        </div>
      </DisplayPreferencesProvider>,
    );

    const transition = screen.getByTestId("practice-launch-transition");
    const appRoot = screen.getByTestId("app-root");
    expect(transition).toHaveAttribute("role", "status");
    expect(transition).toHaveAttribute("aria-live", "polite");
    expect(transition).toHaveAttribute("aria-busy", "true");
    expect(transition).toHaveTextContent("Preparing your interview");
    expect(transition).toHaveAttribute("data-transition-variant", "brand");
    expect(screen.getByTestId("transition-illustration-brand")).toBeInTheDocument();
    expect(transition).toHaveFocus();
    expect(screen.getByTestId("topbar-control")).toBeInTheDocument();
    expect(appRoot).toHaveAttribute("inert");
    expect(appRoot).toHaveAttribute("aria-hidden", "true");
    expect(screen.getByTestId("app-main")).not.toHaveAttribute("inert");
    expect(document.body.style.overflow).toBe("hidden");
    expect(transition).not.toHaveTextContent(/\d+%|opening message/i);

    unmount();

    expect(appRoot).not.toHaveAttribute("inert");
    expect(appRoot).not.toHaveAttribute("aria-hidden");
    expect(document.body.style.overflow).toBe("");
  });

  it("ships a fixed viewport layer and disables non-essential motion for reduced-motion users", () => {
    const css = readFileSync(resolve(__dirname, "..", "screens", "screens.css"), "utf8");

    expect(css).toMatch(/\.ei-practice-launch-transition\s*\{[^}]*position:\s*fixed[^}]*inset:\s*0[^}]*z-index:\s*20/s);
    expect(css).toMatch(/@media\s*\(prefers-reduced-motion:\s*reduce\)[\s\S]*?\.ei-transition-scene[^}]*animation:\s*none/s);
  });
});
