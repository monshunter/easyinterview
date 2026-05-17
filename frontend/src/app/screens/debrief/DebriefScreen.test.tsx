// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";

import type { EasyInterviewClient } from "../../../api/generated/client";
import type {
  Debrief,
  Job,
  PracticeSession,
  ResumeVersion,
  RuntimeConfig,
  TargetJob,
  UserContext,
} from "../../../api/generated/types";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { InterviewContextProvider } from "../../interview-context/InterviewContext";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { DebriefScreen } from "./DebriefScreen";

function renderDebriefScreen(
  navigate = vi.fn(),
  options: {
    client?: EasyInterviewClient;
    lang?: "zh" | "en";
    params?: Record<string, string>;
  } = {},
) {
  const screen = (
    <DebriefScreen
      route={{ name: "debrief", params: options.params ?? {} }}
    />
  );
  const wrapped = options.client ? (
    <AppRuntimeProvider client={options.client}>{screen}</AppRuntimeProvider>
  ) : (
    screen
  );
  return {
    navigate,
    ...render(
      <DisplayPreferencesProvider initial={{ lang: options.lang ?? "zh" }}>
        <InterviewContextProvider>
          <NavigationProvider value={{ navigate }}>{wrapped}</NavigationProvider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

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

const resumeVersion: ResumeVersion = {
  createdAt: "2026-05-17T00:00:00Z",
  displayName: "Backend resume v3",
  id: "rv-3",
  provenance: {
    dataSourceVersion: "resume-v3",
    featureFlag: "resume",
    language: "en-US",
    modelId: "fixture",
    promptVersion: "p1",
    rubricVersion: "r1",
  },
  resumeAssetId: "ra-1",
  structuredProfile: {},
  suggestions: [],
  updatedAt: "2026-05-17T00:00:00Z",
  versionType: "targeted",
};

const practiceSession: PracticeSession = {
  createdAt: "2026-05-17T00:00:00Z",
  hintsEnabled: true,
  id: "ps-1",
  language: "en-US",
  planId: "plan-1",
  status: "completed",
  targetJobId: "tj-1",
  turnCount: 4,
  updatedAt: "2026-05-17T00:00:00Z",
};

const debriefJob: Job = {
  createdAt: "2026-05-17T00:00:00Z",
  id: "job-debrief-1",
  jobType: "debrief_generate",
  resourceId: "deb-1",
  resourceType: "debrief",
  status: "succeeded",
  updatedAt: "2026-05-17T00:00:00Z",
};

const completedDebrief: Debrief = {
  createdAt: "2026-05-17T00:00:00Z",
  id: "deb-1",
  provenance: {
    dataSourceVersion: "debrief-fixture",
    featureFlag: "debrief",
    language: "en-US",
    modelId: "fixture-model",
    promptVersion: "prompt-1",
    rubricVersion: "rubric-1",
  },
  questions: [
    {
      aiAnalysis: "System design depth was clear.",
      myAnswerSummary: "Discussed queue recovery.",
      questionText: "How would you make the queue reliable?",
    },
  ],
  riskItems: [{ label: "Clarify backpressure tradeoffs", severity: "medium" }],
  roundType: "technical",
  status: "completed",
  targetJobId: "tj-1",
  updatedAt: "2026-05-17T00:00:00Z",
};

function createDebriefClient(overrides: Partial<Record<keyof EasyInterviewClient, unknown>> = {}) {
  const client = {
    getRuntimeConfig: vi.fn().mockResolvedValue(runtimeConfig),
    getMe: vi.fn().mockResolvedValue(user),
    getTargetJob: vi.fn().mockResolvedValue(targetJob),
    getResumeVersion: vi.fn().mockResolvedValue(resumeVersion),
    getPracticeSession: vi.fn().mockResolvedValue(practiceSession),
    suggestDebriefQuestions: vi.fn().mockResolvedValue({
      suggestions: [
        {
          questionText: "How would you make the queue reliable?",
          whyLikelyAsked: "It maps to the JD reliability focus.",
          source: "jd_requirement",
          stage: "system_design",
        },
      ],
    }),
    createDebrief: vi.fn().mockResolvedValue({
      debriefId: "deb-1",
      job: debriefJob,
    }),
    getJob: vi.fn().mockResolvedValue(debriefJob),
    getDebrief: vi.fn().mockResolvedValue(completedDebrief),
    ...overrides,
  };
  return client as unknown as EasyInterviewClient;
}

describe("DebriefScreen — TestDebriefScreen_DefaultRender", () => {
  it("renders the route-debrief shell with Header + ContextStrip + Stepper + step panel", () => {
    renderDebriefScreen();
    const shell = screen.getByTestId("route-debrief");
    expect(shell).toBeInTheDocument();
    expect(shell).toHaveAttribute("data-route-name", "debrief");
    expect(shell).toHaveAttribute("data-step", "0");
    expect(shell).toHaveAttribute("data-input-mode", "text");
    expect(shell).toHaveAttribute("data-picker-kind", "none");
    expect(screen.getByTestId("debrief-header")).toBeInTheDocument();
    expect(screen.getByTestId("debrief-context-strip")).toBeInTheDocument();
    expect(screen.getByTestId("debrief-stepper")).toBeInTheDocument();
    expect(screen.getByTestId("debrief-step-panel-0")).toBeInTheDocument();
  });

  it("invokes navigate({name:'home'}) when the header back control is clicked", () => {
    const { navigate } = renderDebriefScreen();
    fireEvent.click(screen.getByTestId("debrief-header-back"));
    expect(navigate).toHaveBeenCalledWith({ name: "home" });
  });

  it("flips data-picker-kind when a context-strip card is opened", () => {
    renderDebriefScreen();
    fireEvent.click(screen.getByTestId("debrief-context-card-targetJob-open"));
    const shell = screen.getByTestId("route-debrief");
    expect(shell).toHaveAttribute("data-picker-kind", "targetJob");
  });

  it("hydrates route context into picker state and suggests questions with the active UI language", async () => {
    const client = createDebriefClient();
    renderDebriefScreen(vi.fn(), {
      client,
      lang: "en",
      params: {
        targetJobId: "tj-1",
        sessionId: "ps-1",
        resumeVersionId: "rv-3",
      },
    });

    expect(
      await screen.findByText("Acme · Senior Backend Engineer"),
    ).toBeInTheDocument();
    await waitFor(() => {
      expect(client.suggestDebriefQuestions).toHaveBeenCalledWith({
        count: 6,
        language: "en-US",
        resumeVersionId: "rv-3",
        sessionId: "ps-1",
        targetJobId: "tj-1",
      });
    });
  });

  it("moves to Step 1 immediately after createDebrief succeeds and sends en-US", async () => {
    const client = createDebriefClient({
      getJob: vi.fn().mockResolvedValue({ ...debriefJob, status: "running" }),
    });
    renderDebriefScreen(vi.fn(), {
      client,
      lang: "en",
      params: {
        targetJobId: "tj-1",
        resumeVersionId: "rv-3",
      },
    });

    await screen.findByText("Acme · Senior Backend Engineer");
    await screen.findByTestId("debrief-guided-current");
    fireEvent.click(screen.getByTestId("debrief-suggested-question-occurred"));
    fireEvent.click(screen.getByTestId("debrief-submit-btn"));

    await waitFor(() => {
      expect(client.createDebrief).toHaveBeenCalledWith(
        expect.objectContaining({
          language: "en-US",
          questions: [
            expect.objectContaining({
              questionText: "How would you make the queue reliable?",
            }),
          ],
        }),
        expect.objectContaining({
          headers: expect.objectContaining({
            "Idempotency-Key": expect.any(String),
          }),
        }),
      );
      expect(screen.getByTestId("route-debrief")).toHaveAttribute(
        "data-step",
        "1",
      );
    });
  });

  it("includes language in the debrief replay handoff payload", async () => {
    const navigate = vi.fn();
    const client = createDebriefClient();
    renderDebriefScreen(navigate, {
      client,
      lang: "en",
      params: {
        targetJobId: "tj-1",
        sessionId: "ps-1",
        resumeVersionId: "rv-3",
      },
    });

    await screen.findByTestId("debrief-guided-current");
    fireEvent.click(screen.getByTestId("debrief-suggested-question-occurred"));
    fireEvent.click(screen.getByTestId("debrief-submit-btn"));
    await screen.findByTestId("debrief-analysis-step");
    fireEvent.click(screen.getByTestId("debrief-analysis-advance"));
    fireEvent.click(screen.getByTestId("debrief-start-interview-btn"));

    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith({
        name: "practice",
        params: expect.objectContaining({
          debriefId: "deb-1",
          language: "en-US",
          modality: "text",
          mode: "text",
          practiceGoal: "debrief",
          resumeVersionId: "rv-3",
          sessionId: "ps-1",
          targetJobId: "tj-1",
        }),
      });
    });
  });
});
