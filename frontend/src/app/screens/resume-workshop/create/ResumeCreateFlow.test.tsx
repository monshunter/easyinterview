// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { NavigationProvider } from "../../../navigation/NavigationProvider";
import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { ResumeCreateFlow } from "./ResumeCreateFlow";

function renderCreateFlow(
  navigate: ReturnType<typeof vi.fn> = vi.fn(),
  initialMode?: "upload" | "paste",
) {
  return render(
    <DisplayPreferencesProvider>
      <NavigationProvider value={{ navigate }}>
        <ResumeCreateFlow initialMode={initialMode} />
      </NavigationProvider>
    </DisplayPreferencesProvider>,
  );
}

describe("ResumeCreateFlow container", () => {
  it("renders the create-flow shell with the upload tab active by default", () => {
    renderCreateFlow();
    const flow = screen.getByTestId("resume-create-flow");
    expect(flow).toHaveAttribute("data-stage", "input");
    expect(flow).toHaveAttribute("data-create-mode", "upload");
    expect(screen.getByTestId("resume-create-tab-upload")).toHaveAttribute(
      "data-active",
      "true",
    );
    expect(
      screen.getByTestId("resume-create-upload-panel"),
    ).toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-create-paste-panel"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-create-guided-panel"),
    ).not.toBeInTheDocument();
  });

  it("renders the sidebar `WHAT GETS SAVED` and `WHAT HAPPENS NEXT` cards", () => {
    renderCreateFlow();
    const sidebar = screen.getByTestId("resume-create-sidebar");
    expect(sidebar).toBeInTheDocument();
    expect(
      within(sidebar).getAllByText(/会保存什么|WHAT GETS SAVED/i).length,
    ).toBeGreaterThanOrEqual(1);
    expect(
      within(sidebar).getAllByText(/接下来|WHAT HAPPENS NEXT/i).length,
    ).toBeGreaterThanOrEqual(1);
  });

  it("switches to the paste tab when the user clicks it", async () => {
    const user = userEvent.setup();
    renderCreateFlow();
    await user.click(screen.getByTestId("resume-create-tab-paste"));
    expect(screen.getByTestId("resume-create-flow")).toHaveAttribute(
      "data-create-mode",
      "paste",
    );
    expect(screen.getByTestId("resume-create-paste-panel")).toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-create-upload-panel"),
    ).not.toBeInTheDocument();
  });

  it("does not expose a guided tab or guided panel (D-20 removed guided intake)", () => {
    renderCreateFlow();
    expect(
      screen.queryByTestId("resume-create-tab-guided"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-create-guided-panel"),
    ).not.toBeInTheDocument();
  });

  it("preserves paste raw text when switching tab away and back (input retained for cancel-from-parse path)", async () => {
    const user = userEvent.setup();
    renderCreateFlow();
    await user.click(screen.getByTestId("resume-create-tab-paste"));
    const textarea = screen.getByTestId(
      "resume-create-paste-textarea",
    ) as HTMLTextAreaElement;
    fireEvent.change(textarea, { target: { value: "alpha beta gamma" } });
    expect(textarea.value).toBe("alpha beta gamma");
    await user.click(screen.getByTestId("resume-create-tab-upload"));
    await user.click(screen.getByTestId("resume-create-tab-paste"));
    expect(
      (screen.getByTestId(
        "resume-create-paste-textarea",
      ) as HTMLTextAreaElement).value,
    ).toBe("alpha beta gamma");
  });

  it("disables the paste submit button when the textarea is empty (whitespace only)", async () => {
    const user = userEvent.setup();
    renderCreateFlow();
    await user.click(screen.getByTestId("resume-create-tab-paste"));
    const submit = screen.getByTestId(
      "resume-create-paste-submit",
    ) as HTMLButtonElement;
    expect(submit.disabled).toBe(true);
    fireEvent.change(screen.getByTestId("resume-create-paste-textarea"), {
      target: { value: "   \n  " },
    });
    expect(submit.disabled).toBe(true);
    fireEvent.change(screen.getByTestId("resume-create-paste-textarea"), {
      target: { value: "some real content" },
    });
    expect(submit.disabled).toBe(false);
  });

  it("clicking the back button navigates to resume_versions list", async () => {
    const user = userEvent.setup();
    const navigate = vi.fn();
    renderCreateFlow(navigate);
    await user.click(screen.getByTestId("resume-create-back"));
    expect(navigate).toHaveBeenCalledTimes(1);
    expect(navigate.mock.calls[0]![0]).toEqual({
      name: "resume_versions",
      params: {},
    });
  });

  it("honours initialMode=paste from the createMode route param", () => {
    renderCreateFlow(vi.fn(), "paste");
    expect(screen.getByTestId("resume-create-flow")).toHaveAttribute(
      "data-create-mode",
      "paste",
    );
    expect(screen.getByTestId("resume-create-paste-panel")).toBeInTheDocument();
  });

  it("ArrowRight/ArrowLeft on the tablist cycles between the two tabs (upload <-> paste)", () => {
    renderCreateFlow();
    const uploadTab = screen.getByTestId(
      "resume-create-tab-upload",
    ) as HTMLButtonElement;
    uploadTab.focus();
    fireEvent.keyDown(uploadTab, { key: "ArrowRight" });
    expect(screen.getByTestId("resume-create-flow")).toHaveAttribute(
      "data-create-mode",
      "paste",
    );
    const pasteTab = screen.getByTestId(
      "resume-create-tab-paste",
    ) as HTMLButtonElement;
    // Only two tabs remain (D-20 removed guided); ArrowRight wraps back to upload.
    fireEvent.keyDown(pasteTab, { key: "ArrowRight" });
    expect(screen.getByTestId("resume-create-flow")).toHaveAttribute(
      "data-create-mode",
      "upload",
    );
    fireEvent.keyDown(uploadTab, { key: "ArrowLeft" });
    expect(screen.getByTestId("resume-create-flow")).toHaveAttribute(
      "data-create-mode",
      "paste",
    );
  });
});
