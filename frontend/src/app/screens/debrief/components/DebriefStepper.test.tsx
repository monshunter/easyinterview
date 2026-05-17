// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";

import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { DebriefStepper } from "./DebriefStepper";
import type { DebriefStep } from "../types";

function setup(step: DebriefStep, maxVisited: DebriefStep) {
  const onStep = vi.fn();
  render(
    <DisplayPreferencesProvider>
      <DebriefStepper step={step} maxVisited={maxVisited} onStep={onStep} />
    </DisplayPreferencesProvider>,
  );
  return { onStep };
}

describe("DebriefStepper — TestStepper_NavigationLogic", () => {
  it("highlights the current step and marks future steps unreachable", () => {
    setup(0, 0);
    expect(screen.getByTestId("debrief-stepper-step-0")).toHaveAttribute(
      "data-active",
      "true",
    );
    expect(screen.getByTestId("debrief-stepper-step-1")).toHaveAttribute(
      "data-reachable",
      "false",
    );
    expect(screen.getByTestId("debrief-stepper-step-2")).toHaveAttribute(
      "data-reachable",
      "false",
    );
  });

  it("allows clicking back to a visited step but blocks forward jumps", () => {
    const { onStep } = setup(2, 2);
    fireEvent.click(screen.getByTestId("debrief-stepper-step-0"));
    expect(onStep).toHaveBeenCalledWith(0);
    fireEvent.click(screen.getByTestId("debrief-stepper-step-1"));
    expect(onStep).toHaveBeenCalledWith(1);
  });

  it("does not invoke onStep when an unreachable forward step is clicked", () => {
    const { onStep } = setup(0, 0);
    fireEvent.click(screen.getByTestId("debrief-stepper-step-1"));
    expect(onStep).not.toHaveBeenCalled();
  });
});
