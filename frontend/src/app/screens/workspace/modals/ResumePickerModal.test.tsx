/**
 * @vitest-environment jsdom
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";

import { InterviewContextProvider } from "../../../interview-context/InterviewContext";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import { ResumePickerModal } from "./ResumePickerModal";

function withProviders(ui: ReactNode) {
  const nav = vi.fn();
  return {
    nav,
    ...render(
      <InterviewContextProvider>
        <NavigationProvider value={{ navigate: nav }}>
          {ui}
        </NavigationProvider>
      </InterviewContextProvider>,
    ),
  };
}

describe("ResumePickerModal (Phase 3.3)", () => {
  it("renders when open with correct testids and i18n", () => {
    withProviders(
      <ResumePickerModal open onClose={vi.fn()} boundResumeId="rv-1" />,
    );
    expect(screen.getByTestId("workspace-resume-modal-overlay")).toBeDefined();
    expect(screen.getByTestId("workspace-resume-modal-card")).toBeDefined();
    expect(screen.getByTestId("workspace-resume-modal-card-rv-1")).toBeDefined();
    expect(screen.getByTestId("workspace-resume-modal-disabled-note")).toBeDefined();
    expect(screen.getByTestId("workspace-resume-modal-close")).toBeDefined();
    expect(screen.getByTestId("workspace-resume-modal-cancel")).toBeDefined();
    expect(screen.getByTestId("workspace-resume-modal-confirm")).toBeDefined();
  });

  it("does NOT render when open=false", () => {
    const { container } = withProviders(
      <ResumePickerModal open={false} onClose={vi.fn()} />,
    );
    expect(container.querySelector("[data-testid]")).toBeNull();
  });

  it("closes on Cancel button click", async () => {
    const onClose = vi.fn();
    withProviders(<ResumePickerModal open onClose={onClose} />);
    const user = userEvent.setup();
    await user.click(screen.getByTestId("workspace-resume-modal-cancel"));
    expect(onClose).toHaveBeenCalled();
  });

  it("closes on Confirm button click and calls onUseResume", async () => {
    const onClose = vi.fn();
    const onUseResume = vi.fn();
    withProviders(
      <ResumePickerModal
        open
        onClose={onClose}
        boundResumeId="rv-1"
        onUseResume={onUseResume}
      />,
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("workspace-resume-modal-confirm"));
    expect(onUseResume).toHaveBeenCalled();
    expect(onClose).toHaveBeenCalled();
  });

  it("has aria-modal attribute", () => {
    withProviders(<ResumePickerModal open onClose={vi.fn()} />);
    expect(screen.getByTestId("workspace-resume-modal-card")).toHaveAttribute(
      "aria-modal",
      "true",
    );
  });
});
