// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import type { EasyInterviewClient } from "../../../../api/generated/client";
import type { RuntimeConfig } from "../../../../api/generated/types";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { AppRuntimeContext } from "../../../runtime/AppRuntimeProvider";
import { ResumeCreateFlow } from "./ResumeCreateFlow";

import getRuntimeConfigFixture from "../../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";

const TEST_RUNTIME_CONFIG = getRuntimeConfigFixture.scenarios.default.response
  .body as RuntimeConfig;

function renderCreateFlow(
  navigate: ReturnType<typeof vi.fn> = vi.fn(),
  initialMode?: "upload" | "paste",
) {
  return render(
    <DisplayPreferencesProvider>
      <AppRuntimeContext.Provider
        value={{
          client: {} as EasyInterviewClient,
          runtime: { status: "ready", config: TEST_RUNTIME_CONFIG },
          auth: { status: "authenticated", user: {} as never },
          refreshAuth: vi.fn(),
        }}
      >
        <NavigationProvider value={{ navigate }}>
          <ResumeCreateFlow initialMode={initialMode} />
        </NavigationProvider>
      </AppRuntimeContext.Provider>
    </DisplayPreferencesProvider>,
  );
}

function renderCreateFlowWithRuntime(
  client: Pick<EasyInterviewClient, "registerResume">,
  navigate: ReturnType<typeof vi.fn> = vi.fn(),
) {
  return {
    navigate,
    ...render(
      <DisplayPreferencesProvider>
        <AppRuntimeContext.Provider
          value={{
            client: client as EasyInterviewClient,
            runtime: { status: "ready", config: TEST_RUNTIME_CONFIG },
            auth: { status: "authenticated", user: {} as never },
            refreshAuth: vi.fn(),
          }}
        >
          <NavigationProvider value={{ navigate }}>
            <ResumeCreateFlow initialMode="paste" />
          </NavigationProvider>
        </AppRuntimeContext.Provider>
      </DisplayPreferencesProvider>,
    ),
  };
}

const REGISTER_RESULT = {
  resumeId: "01918fa0-0000-7000-8000-000000001999",
  job: {
    id: "01918fa0-0000-7000-8000-00000000b999",
    jobType: "resume_parse",
    status: "queued",
    resourceType: "resume_asset",
    resourceId: "01918fa0-0000-7000-8000-000000001999",
    errorCode: null,
    createdAt: "2026-07-07T00:00:00Z",
    updatedAt: "2026-07-07T00:00:00Z",
  },
} as const;

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

  it("does not render the redundant `WHAT GETS SAVED` or `WHAT HAPPENS NEXT` sidebar", () => {
    renderCreateFlow();
    expect(screen.queryByTestId("resume-create-sidebar")).not.toBeInTheDocument();
    expect(screen.queryByText(/会保存什么|WHAT GETS SAVED/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/接下来|WHAT HAPPENS NEXT/i)).not.toBeInTheDocument();
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

  it("renders only upload and paste tabs", () => {
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
    // Only two tabs remain; ArrowRight wraps back to upload.
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

  it("submits pasted content with a neutral source title and navigates directly to detail", async () => {
    const user = userEvent.setup();
    const registerResume = vi.fn().mockResolvedValue(REGISTER_RESULT);
    const { navigate } = renderCreateFlowWithRuntime({ registerResume });

    fireEvent.change(screen.getByTestId("resume-create-paste-textarea"), {
      target: {
        value:
          "张三 · 后端平台工程师\nFerry / reloadr / grayplan - GitOps CI/CD 与配置治理平台",
      },
    });
    await user.click(screen.getByTestId("resume-create-paste-submit"));

    expect(registerResume).toHaveBeenCalledTimes(1);
    expect(registerResume.mock.calls[0]![0]).toMatchObject({
      sourceType: "paste",
      title: "Pasted text",
    });
    expect(registerResume.mock.calls[0]![0].title).not.toContain(
      "张三 · 后端平台工程师",
    );
    expect(registerResume.mock.calls[0]![0].title).not.toBe("粘贴的简历");
    expect(navigate).toHaveBeenCalledWith({
      name: "resume_versions",
      params: { resumeId: REGISTER_RESULT.resumeId },
    });
    expect(screen.queryByTestId("resume-parse-flow")).not.toBeInTheDocument();
    expect(screen.queryByTestId("resume-preview-confirm")).not.toBeInTheDocument();
  });

});
