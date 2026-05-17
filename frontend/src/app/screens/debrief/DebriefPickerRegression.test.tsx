// @vitest-environment jsdom
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

import type { EasyInterviewClient } from "../../../api/generated/client";
import type {
  ResumeAsset,
  ResumeVersion,
  RuntimeConfig,
  TargetJob,
  UserContext,
} from "../../../api/generated/types";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { GuidedDebriefRecord } from "./components/GuidedDebriefRecord";
import { JDPicker } from "./components/JDPicker";
import { ResumePicker } from "./components/ResumePicker";

const runtimeConfig: RuntimeConfig = {
  analyticsEnabled: false,
  appVersion: "test",
  defaultUiLanguage: "en",
  featureFlags: {},
};

const user: UserContext = {
  displayName: "Candidate",
  emailMasked: "c***@example.com",
  id: "user-1",
  preferredPracticeLanguage: "en-US",
  uiLanguage: "en",
};

const pageInfo = { hasMore: false, nextCursor: null, pageSize: 20 };

const targetJob: TargetJob = {
  analysisStatus: "ready",
  companyName: "Acme",
  createdAt: "2026-05-17T00:00:00Z",
  id: "tj-1",
  openQuestionIssueCount: 0,
  requirements: [],
  sourceType: "manual_form",
  status: "applied",
  targetLanguage: "en-US",
  title: "Senior Backend Engineer",
  updatedAt: "2026-05-17T00:00:00Z",
};

const oldResumeAsset: ResumeAsset = {
  createdAt: "2026-05-17T00:00:00Z",
  id: "ra-old",
  language: "en-US",
  parseStatus: "ready",
  status: "active",
  title: "Old resume",
  updatedAt: "2026-05-17T00:00:00Z",
};

const newResumeAsset: ResumeAsset = {
  ...oldResumeAsset,
  id: "ra-2",
  title: "Backend resume",
};

const newResumeVersion: ResumeVersion = {
  createdAt: "2026-05-17T00:00:00Z",
  displayName: "Backend resume v4",
  id: "rv-4",
  provenance: {
    dataSourceVersion: "resume-v4",
    featureFlag: "resume",
    language: "en-US",
    modelId: "fixture",
    promptVersion: "p1",
    rubricVersion: "r1",
  },
  resumeAssetId: "ra-2",
  structuredProfile: {},
  suggestions: [],
  updatedAt: "2026-05-17T00:00:00Z",
  versionType: "targeted",
};

function renderWithRuntime(ui: ReactNode, client: Partial<EasyInterviewClient>) {
  const fullClient = {
    getRuntimeConfig: vi.fn().mockResolvedValue(runtimeConfig),
    getMe: vi.fn().mockResolvedValue(user),
    ...client,
  } as unknown as EasyInterviewClient;

  return render(
    <DisplayPreferencesProvider initial={{ lang: "en" }}>
      <AppRuntimeProvider client={fullClient}>{ui}</AppRuntimeProvider>
    </DisplayPreferencesProvider>,
  );
}

describe("Debrief pickers and guided entry regressions", () => {
  it("filters target jobs by analysisStatus=ready", async () => {
    const listTargetJobs = vi.fn().mockResolvedValue({
      items: [targetJob],
      pageInfo,
    });
    renderWithRuntime(
      <JDPicker
        selectedId={null}
        onClose={vi.fn()}
        onConfirm={vi.fn()}
      />,
      { listTargetJobs },
    );

    await waitFor(() => {
      expect(listTargetJobs).toHaveBeenCalledWith({
        query: { analysisStatus: "ready" },
      });
    });
  });

  it("reopens resume picker on the asset list and keeps the selected asset object", async () => {
    const listResumes = vi.fn().mockResolvedValue({
      items: [oldResumeAsset, newResumeAsset],
      pageInfo,
    });
    const listResumeVersions = vi.fn().mockResolvedValue({
      items: [newResumeVersion],
      pageInfo,
    });
    const onConfirm = vi.fn();

    renderWithRuntime(
      <ResumePicker
        selectedAssetId="ra-old"
        selectedVersionId="rv-old"
        onClose={vi.fn()}
        onConfirm={onConfirm}
      />,
      { listResumes, listResumeVersions },
    );

    fireEvent.click(await screen.findByTestId("debrief-picker-option-ra-2"));
    fireEvent.click(screen.getByTestId("debrief-picker-confirm"));

    await waitFor(() => {
      expect(listResumeVersions).toHaveBeenCalledWith("ra-2");
    });

    fireEvent.click(await screen.findByTestId("debrief-picker-option-rv-4"));
    fireEvent.click(screen.getByTestId("debrief-picker-confirm"));

    expect(onConfirm).toHaveBeenCalledWith({
      asset: newResumeAsset,
      version: newResumeVersion,
    });
  });

  it.each([
    { errorCode: "AI_PROVIDER_TIMEOUT", label: "failure", suggestions: null },
    { errorCode: null, label: "empty", suggestions: [] },
  ])("keeps manual entry available on $label suggestion state", ({ errorCode, suggestions }) => {
    const setEntries = vi.fn();
    render(
      <DisplayPreferencesProvider initial={{ lang: "en" }}>
        <GuidedDebriefRecord
          activeGuide={0}
          entries={[]}
          errorCode={errorCode}
          loading={false}
          onRegenerate={vi.fn()}
          setActiveGuide={vi.fn()}
          setEntries={setEntries}
          suggestions={suggestions}
        />
      </DisplayPreferencesProvider>,
    );

    fireEvent.click(screen.getByTestId("debrief-suggested-question-manual"));
    fireEvent.change(screen.getByTestId("debrief-guided-editor-input"), {
      target: { value: "What was the production incident timeline?" },
    });
    fireEvent.change(screen.getByTestId("debrief-guided-editor-answer"), {
      target: { value: "I summarized detection, mitigation, and follow-up." },
    });
    fireEvent.click(screen.getByTestId("debrief-guided-editor-save"));

    expect(setEntries).toHaveBeenCalledWith([
      expect.objectContaining({
        myAnswerSummary: "I summarized detection, mitigation, and follow-up.",
        questionText: "What was the production incident timeline?",
        source: "manual",
      }),
    ]);
  });
});
