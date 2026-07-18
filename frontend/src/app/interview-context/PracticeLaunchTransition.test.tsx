// @vitest-environment jsdom

import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { DisplayPreferencesProvider } from "../display/DisplayPreferencesProvider";
import { PracticeLaunchTransition } from "./PracticeLaunchTransition";

describe("PracticeLaunchTransition", () => {
  it("announces an honest busy state, blocks the background, and restores it on unmount", () => {
    const { container, unmount } = render(
      <DisplayPreferencesProvider>
        <PracticeLaunchTransition />
      </DisplayPreferencesProvider>,
    );

    const transition = screen.getByTestId("practice-launch-transition");
    expect(transition).toHaveAttribute("role", "status");
    expect(transition).toHaveAttribute("aria-live", "polite");
    expect(transition).toHaveAttribute("aria-busy", "true");
    expect(transition).toHaveTextContent("Preparing your interview");
    expect(transition).toHaveFocus();
    expect(container).toHaveAttribute("inert");
    expect(container).toHaveAttribute("aria-hidden", "true");
    expect(document.body.style.overflow).toBe("hidden");
    expect(transition).not.toHaveTextContent(/\d+%|opening message/i);

    unmount();

    expect(container).not.toHaveAttribute("inert");
    expect(container).not.toHaveAttribute("aria-hidden");
    expect(document.body.style.overflow).toBe("");
  });

  it("ships a fixed viewport layer and disables non-essential motion for reduced-motion users", () => {
    const css = readFileSync(resolve(__dirname, "..", "screens", "screens.css"), "utf8");

    expect(css).toMatch(/\.ei-practice-launch-transition\s*\{[^}]*position:\s*fixed[^}]*inset:\s*0/s);
    expect(css).toMatch(/@media\s*\(prefers-reduced-motion:\s*reduce\)[\s\S]*?\.ei-practice-launch-orbit[\s\S]*?animation:\s*none/);
  });
});
