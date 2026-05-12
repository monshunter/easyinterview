// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { OriginalResumePreviewModal } from "./OriginalResumePreviewModal";

function renderModal(onClose = vi.fn()) {
  return render(
    <DisplayPreferencesProvider>
      <button data-testid="anchor">anchor</button>
      <OriginalResumePreviewModal
        open
        onClose={onClose}
        originalText={["Liu Zhe", "Senior frontend engineer"]}
        title="Master CN"
      />
    </DisplayPreferencesProvider>,
  );
}

describe("OriginalResumePreviewModal a11y", () => {
  it("renders role=dialog with aria-labelledby and aria-modal=true", () => {
    renderModal();
    const dialog = screen.getByTestId("resume-detail-original-modal");
    expect(dialog).toHaveAttribute("role", "dialog");
    expect(dialog).toHaveAttribute("aria-modal", "true");
    expect(dialog).toHaveAttribute("aria-labelledby");
    expect(dialog).toHaveAttribute("aria-describedby");
  });

  it("closes when ESC key is pressed", async () => {
    const onClose = vi.fn();
    renderModal(onClose);
    await userEvent.setup().keyboard("{Escape}");
    expect(onClose).toHaveBeenCalled();
  });

  it("closes when the outer overlay is clicked", async () => {
    const onClose = vi.fn();
    renderModal(onClose);
    const overlay = screen.getByTestId("resume-detail-original-modal-overlay");
    await userEvent.setup().click(overlay);
    expect(onClose).toHaveBeenCalled();
  });

  it("closes when the X (close) button is clicked", async () => {
    const onClose = vi.fn();
    renderModal(onClose);
    const closeBtn = screen.getByTestId("resume-detail-original-modal-close");
    await userEvent.setup().click(closeBtn);
    expect(onClose).toHaveBeenCalled();
  });

  it("does not close when clicking inside the dialog content", async () => {
    const onClose = vi.fn();
    renderModal(onClose);
    const content = screen.getByTestId("resume-detail-original-modal-content");
    await userEvent.setup().click(content);
    expect(onClose).not.toHaveBeenCalled();
  });

  it("renders the original text lines as paragraphs", () => {
    renderModal();
    const content = screen.getByTestId("resume-detail-original-modal-content");
    expect(content).toHaveTextContent("Liu Zhe");
    expect(content).toHaveTextContent("Senior frontend engineer");
  });
});
