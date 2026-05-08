// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { JDAssistModal } from "./JDAssistModal";

function wrap(ui: React.ReactElement) {
  return <DisplayPreferencesProvider>{ui}</DisplayPreferencesProvider>;
}

describe("JDAssistModal — upload variant", () => {
  it("renders upload DOM with required testids", () => {
    render(
      wrap(
        <JDAssistModal
          type="upload"
          onClose={vi.fn()}
          onConfirm={vi.fn()}
        />,
      ),
    );

    expect(screen.getByTestId("home-modal-upload-dropzone")).toBeInTheDocument();
    expect(screen.getByTestId("home-modal-upload-continue")).toBeInTheDocument();
    expect(screen.getByTestId("home-modal-upload-cancel")).toBeInTheDocument();
    expect(screen.getByTestId("home-modal-upload-close")).toBeInTheDocument();
  });

  it("calls onClose when X button clicked", () => {
    const onClose = vi.fn();
    render(
      wrap(
        <JDAssistModal type="upload" onClose={onClose} onConfirm={vi.fn()} />,
      ),
    );

    fireEvent.click(screen.getByTestId("home-modal-upload-close"));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("calls onClose when Cancel button clicked", () => {
    const onClose = vi.fn();
    render(
      wrap(
        <JDAssistModal type="upload" onClose={onClose} onConfirm={vi.fn()} />,
      ),
    );

    fireEvent.click(screen.getByTestId("home-modal-upload-cancel"));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("calls onClose on ESC key", async () => {
    const onClose = vi.fn();
    render(
      wrap(
        <JDAssistModal type="upload" onClose={onClose} onConfirm={vi.fn()} />,
      ),
    );

    await userEvent.keyboard("{Escape}");
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("calls onClose when overlay backdrop clicked", () => {
    const onClose = vi.fn();
    render(
      wrap(
        <JDAssistModal type="upload" onClose={onClose} onConfirm={vi.fn()} />,
      ),
    );

    const overlay = screen.getByTestId("home-modal-upload-backdrop");
    fireEvent.click(overlay);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("does NOT close when inner dialog card clicked", () => {
    const onClose = vi.fn();
    render(
      wrap(
        <JDAssistModal type="upload" onClose={onClose} onConfirm={vi.fn()} />,
      ),
    );

    fireEvent.click(screen.getByTestId("home-modal-upload-dropzone"));
    expect(onClose).not.toHaveBeenCalled();
  });

  it("calls onConfirm with upload source when Continue button clicked", () => {
    const onConfirm = vi.fn();
    render(
      wrap(
        <JDAssistModal
          type="upload"
          onClose={vi.fn()}
          onConfirm={onConfirm}
        />,
      ),
    );

    fireEvent.click(screen.getByTestId("home-modal-upload-continue"));
    expect(onConfirm).toHaveBeenCalledTimes(1);
    expect(onConfirm).toHaveBeenCalledWith({ source: "upload" });
  });
});

describe("JDAssistModal — url variant", () => {
  it("renders url DOM with required testids", () => {
    render(
      wrap(
        <JDAssistModal type="url" onClose={vi.fn()} onConfirm={vi.fn()} />,
      ),
    );

    expect(screen.getByTestId("home-modal-url-input")).toBeInTheDocument();
    expect(screen.getByTestId("home-modal-url-continue")).toBeInTheDocument();
    expect(screen.getByTestId("home-modal-url-cancel")).toBeInTheDocument();
    expect(screen.getByTestId("home-modal-url-close")).toBeInTheDocument();
  });

  it("calls onClose when X button clicked", () => {
    const onClose = vi.fn();
    render(
      wrap(
        <JDAssistModal type="url" onClose={onClose} onConfirm={vi.fn()} />,
      ),
    );

    fireEvent.click(screen.getByTestId("home-modal-url-close"));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("calls onClose when Cancel button clicked", () => {
    const onClose = vi.fn();
    render(
      wrap(
        <JDAssistModal type="url" onClose={onClose} onConfirm={vi.fn()} />,
      ),
    );

    fireEvent.click(screen.getByTestId("home-modal-url-cancel"));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("calls onClose on ESC key", async () => {
    const onClose = vi.fn();
    render(
      wrap(
        <JDAssistModal type="url" onClose={onClose} onConfirm={vi.fn()} />,
      ),
    );

    await userEvent.keyboard("{Escape}");
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("calls onClose when overlay backdrop clicked", () => {
    const onClose = vi.fn();
    render(
      wrap(
        <JDAssistModal type="url" onClose={onClose} onConfirm={vi.fn()} />,
      ),
    );

    const overlay = screen.getByTestId("home-modal-url-backdrop");
    fireEvent.click(overlay);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("calls onConfirm with url source when Continue button clicked", async () => {
    const onConfirm = vi.fn();
    render(
      wrap(
        <JDAssistModal type="url" onClose={vi.fn()} onConfirm={onConfirm} />,
      ),
    );

    const input = screen.getByTestId("home-modal-url-input");
    await userEvent.type(input, "https://example.com/job");

    fireEvent.click(screen.getByTestId("home-modal-url-continue"));
    expect(onConfirm).toHaveBeenCalledTimes(1);
    expect(onConfirm).toHaveBeenCalledWith({
      source: "url",
      url: "https://example.com/job",
    });
  });
});
