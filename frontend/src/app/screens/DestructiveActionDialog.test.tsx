// @vitest-environment jsdom
import { useRef } from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { DestructiveActionDialog } from "./DestructiveActionDialog";

function DialogHarness({
  pending = false,
  onCancel = vi.fn(),
  onConfirm = vi.fn(),
}: {
  pending?: boolean;
  onCancel?: () => void;
  onConfirm?: () => void;
}) {
  const triggerRef = useRef<HTMLButtonElement>(null);
  return (
    <>
      <button ref={triggerRef} type="button">
        Delete item
      </button>
      <DestructiveActionDialog
        eyebrow="Please be careful"
        title="Delete this item?"
        description="This item will leave the list."
        cancelLabel="Cancel"
        confirmLabel="Confirm deletion"
        pendingLabel="Deleting…"
        retryLabel="Try again"
        errorMessage={null}
        pending={pending}
        restoreFocusRef={triggerRef}
        onCancel={onCancel}
        onConfirm={onConfirm}
      />
    </>
  );
}

describe("DestructiveActionDialog", () => {
  it("starts on the safe action, traps focus, and restores the trigger after Escape", async () => {
    const onCancel = vi.fn();
    render(<DialogHarness onCancel={onCancel} />);
    const user = userEvent.setup();

    expect(screen.getByRole("button", { name: "Cancel" })).toHaveFocus();
    await user.keyboard("{Shift>}{Tab}{/Shift}");
    expect(screen.getByRole("button", { name: "Confirm deletion" })).toHaveFocus();
    await user.keyboard("{Tab}");
    expect(screen.getByRole("button", { name: "Cancel" })).toHaveFocus();
    await user.keyboard("{Escape}");

    expect(onCancel).toHaveBeenCalledTimes(1);
    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Delete item" })).toHaveFocus();
    });
  });

  it("treats backdrop click as cancel but locks every close path while pending", () => {
    const onCancel = vi.fn();
    const { rerender } = render(<DialogHarness onCancel={onCancel} />);
    const layer = screen.getByTestId("destructive-action-dialog-layer");
    fireEvent.mouseDown(layer);
    expect(onCancel).toHaveBeenCalledTimes(1);

    rerender(<DialogHarness pending onCancel={onCancel} />);
    fireEvent.mouseDown(screen.getByTestId("destructive-action-dialog-layer"));
    fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });
    expect(onCancel).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Deleting…" })).toBeDisabled();
  });
});
