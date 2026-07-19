// @vitest-environment jsdom

import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { AsyncTransitionScene } from "./AsyncTransitionScene";

describe("AsyncTransitionScene", () => {
  it.each(["brand", "resume", "report", "job"] as const)(
    "renders the %s code-native illustration inside one honest busy surface",
    (variant) => {
      render(
        <AsyncTransitionScene
          variant={variant}
          testId={`scene-${variant}`}
          eyebrow="Observed status"
          title="Waiting for the server"
          body="This can take a moment."
          hint="Keep this page open."
        />,
      );

      const scene = screen.getByTestId(`scene-${variant}`);
      expect(scene).toHaveAttribute("role", "status");
      expect(scene).toHaveAttribute("aria-live", "polite");
      expect(scene).toHaveAttribute("aria-busy", "true");
      expect(scene).toHaveAttribute("data-transition-variant", variant);
      expect(screen.getByTestId(`transition-illustration-${variant}`).tagName).toBe("svg");
      expect(scene).not.toHaveTextContent(/\d+%|estimated|provider|prompt/i);
    },
  );

  it("renders optional steps and a real action without turning the visual rule into determinate progress", async () => {
    const onAction = vi.fn();
    render(
      <AsyncTransitionScene
        variant="job"
        testId="scene-with-steps"
        eyebrow="Step 2 of 4"
        title="Reading the JD"
        body="Building interview context."
        showProgress
        steps={[
          { label: "Read title", state: "done", testId: "step-0" },
          { label: "Read requirements", state: "current", testId: "step-1", statusLabel: "Working" },
          { label: "Build interview", state: "pending", testId: "step-2" },
        ]}
        action={{ label: "Back", onClick: onAction, testId: "transition-back" }}
      />,
    );

    expect(screen.getByTestId("step-0")).toHaveAttribute("data-step-state", "done");
    expect(screen.getByTestId("step-1")).toHaveAttribute("aria-current", "step");
    expect(
      screen.getByTestId("step-1").querySelector(".ei-transition-scene__step-marker"),
    ).toHaveTextContent("2");
    expect(
      screen.getByTestId("step-2").querySelector(".ei-transition-scene__step-marker"),
    ).toHaveTextContent("3");
    expect(screen.getByTestId("transition-progress")).not.toHaveAttribute("aria-valuenow");
    await userEvent.setup().click(screen.getByTestId("transition-back"));
    expect(onAction).toHaveBeenCalledOnce();
  });

  it("ships one responsive canvas and disables decorative motion for reduced-motion users", () => {
    const source = readFileSync(resolve(__dirname, "AsyncTransitionScene.tsx"), "utf8");
    const css = readFileSync(resolve(__dirname, "..", "screens", "screens.css"), "utf8");

    expect(source).not.toMatch(/\bcard\??:/);
    expect(source).not.toContain("ei-transition-scene--card");
    expect(css).toMatch(/\.ei-transition-scene\s*\{[^}]*min-height:\s*calc\(100dvh\s*-\s*76px\)/s);
    expect(css).toMatch(/\.ei-transition-scene__content\s*\{[^}]*width:\s*min\(100%,\s*1090px\)/s);
    expect(css).not.toContain(".ei-transition-scene--card");
    expect(css).toMatch(/\.ei-transition-scene--resume\s*\{[^}]*width:\s*100vw/s);
    expect(css).toMatch(/@media\s*\(max-width:\s*720px\)[\s\S]*?\.ei-transition-scene/s);
    expect(css).toMatch(/@media\s*\(prefers-reduced-motion:\s*reduce\)[\s\S]*?\.ei-transition-scene__illustration[^}]*animation:\s*none/s);
  });
});
