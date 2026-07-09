/**
 * @vitest-environment jsdom
 *
 * Phase 4 — Replay CTA path A + path B wire. Verified through ReportHeader
 * inside the live dashboard so the test exercises the actual handoff:
 *  - authenticated → report owner creates a fresh practice session and lands
 *    on practice
 *  - unauthenticated → useRequestAuth (nav auth_login carrying replay_practice
 *    pending action for report recovery) and no direct nav practice
 *  - payload integrity: 9+ owner / display knob fields, no raw text
 *  - getFeedbackReport not re-invoked on click
 *  - listTargetJobReports never invoked from report scope
 *  - path B carries nextRoundId derived from the current roundId
 */

import {
  act,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { useEffect, type FC, type ReactNode } from "react";

import type { FeedbackReport } from "../../../../api/generated/types";
import { EasyInterviewClient } from "../../../../api/generated/client";
import { App } from "../../../App";
import { useNavigation } from "../../../navigation/NavigationProvider";
import type { LooseRoute } from "../../../normalizeRoute";

const REPORT_ID = "01918fa0-0000-7000-8000-000000007000";
const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";
const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";
const RESUME_VERSION_ID = "01918fa0-0000-7000-8000-000000004000";

function makeReport(): FeedbackReport {
  return {
    id: REPORT_ID,
    sessionId: SESSION_ID,
    targetJobId: TARGET_JOB_ID,
    status: "ready",
    preparednessLevel: "basically_ready",
    highlights: [],
    issues: [
      { dimension: "technical_depth", evidence: "missing metric", confidence: "medium" },
    ],
    nextActions: [{ type: "retry_current_round", label: "rerun" }],
    questionAssessments: [
      {
        turnId: "turn-1",
        questionIntent: "design.api.versioning",
        dimensionResults: {},
        reviewStatus: "queued_for_retry",
        includedInRetryPlan: true,
      },
    ],
    retryFocusTurnIds: ["turn-1", "turn-3"],
    provenance: {
      promptVersion: "feedback_report.v3",
      rubricVersion: "feedback_report.rubric.v2",
      modelId: "model-profile:contract.default",
      language: "zh-CN",
      featureFlag: "none",
      dataSourceVersion: "practice_session.v9",
    },
    createdAt: "2026-05-16T00:00:00Z",
    updatedAt: "2026-05-16T00:00:10Z",
  };
}

interface ClientOpts {
  authenticated: boolean;
}

function makeClient(opts: ClientOpts): EasyInterviewClient {
  const feedback = vi.fn(async (_: string, options?: { headers?: Record<string, string> }) => {
    if (options?.headers) {
      const k = Object.keys(options.headers).map((h) => h.toLowerCase());
      if (k.includes("idempotency-key")) throw new Error("read leaked idempotency");
    }
    return makeReport();
  });
  return {
    async getRuntimeConfig() {
      return { aiProviderProfile: "stub" } as never;
    },
    async getMe() {
      if (opts.authenticated) {
        return { id: "user-1", email: "u@example.com" } as never;
      }
      throw new Error("HTTP 401 Unauthorized");
    },
    getFeedbackReport: feedback,
    getTargetJob: vi.fn(async () => ({
      id: TARGET_JOB_ID,
      analysisStatus: "ready",
      title: "Senior Frontend Engineer",
      companyName: "Acme",
      locationText: "Remote",
      targetLanguage: "zh-CN",
      sourceType: "manual_text",
      sourceUrl: null,
      requirements: [],
      latestReportId: REPORT_ID,
      openQuestionIssueCount: 0,
      status: "ready",
      createdAt: "2026-05-16T00:00:00Z",
      updatedAt: "2026-05-16T00:00:00Z",
    })),
    listTargetJobs: vi.fn(async () => ({
      items: [
        {
          id: TARGET_JOB_ID,
          analysisStatus: "ready",
          title: "Senior Frontend Engineer",
          companyName: "Acme",
          locationText: "Remote",
          targetLanguage: "zh-CN",
          sourceType: "manual_text",
          sourceUrl: null,
          requirements: [],
          latestReportId: REPORT_ID,
          openQuestionIssueCount: 0,
          status: "ready",
          createdAt: "2026-05-16T00:00:00Z",
          updatedAt: "2026-05-16T00:00:00Z",
        },
      ],
      pageInfo: { nextCursor: null, pageSize: 12, hasMore: false },
    })),
    getResume: vi.fn(async () => ({
      id: RESUME_VERSION_ID,
      title: "Resume v3",
      parsedSummary: { headline: "Frontend lead" },
    })),
    createPracticePlan: vi.fn(async () => ({
      id: "01918fa0-0000-7000-8000-000000008000",
      status: "ready",
    })),
    startPracticeSession: vi.fn(async () => ({
      id: "01918fa0-0000-7000-8000-000000009000",
      status: "active",
    })),
    listTargetJobReports: vi.fn(async () => {
      throw new Error("must not be called");
    }),
    getPracticeSession: vi.fn(async () => {
      throw new Error("HTTP 404 Not Found");
    }),
    getPracticePlan: vi.fn(async () => {
      throw new Error("HTTP 404 Not Found");
    }),
  } as unknown as EasyInterviewClient;
}

function NavSpy({
  onRouteChange,
}: {
  onRouteChange: (name: string, params: Record<string, string>) => void;
}) {
  // Use a probe that overrides the navigate callback to record calls.
  return <NavRecorder onRouteChange={onRouteChange} />;
}

function NavRecorder({
  onRouteChange,
}: {
  onRouteChange: (name: string, params: Record<string, string>) => void;
}) {
  const { navigate } = useNavigation();
  useEffect(() => {
    const original = navigate;
    // Intercept by wrapping the navigation provider isn't possible here; instead
    // assert via DOM in the harness for `auth_login` route presence.
    void original;
    onRouteChange("__mount__", {});
  }, [navigate, onRouteChange]);
  return null;
}

const ROUTE_BASE: Record<string, string> = {
  sessionId: SESSION_ID,
  reportId: REPORT_ID,
  targetJobId: TARGET_JOB_ID,
  resumeId: RESUME_VERSION_ID,
  roundId: "round-tech-1",
  planId: "plan-1",
  jdId: "jd-1",
  mode: "text",
  modality: "text",
  practiceMode: "strict",
  hintUsed: "false",
  hintCount: "0",
};

const Harness: FC<{
  client: EasyInterviewClient;
  initialRoute: LooseRoute;
  children?: ReactNode;
}> = ({ client, initialRoute, children }) => (
  <App client={client} initialRoute={initialRoute}>
    {children}
  </App>
);

describe("Replay CTAs", () => {
  it("authenticated user clicking replay CTA creates a fresh practice session directly (TestReplayCtaPathA_AuthenticatedAutoStartPractice)", async () => {
    const client = makeClient({ authenticated: true });
    const startSpy = client.startPracticeSession as ReturnType<typeof vi.fn>;
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: ROUTE_BASE,
        }}
      />,
    );
    await screen.findByTestId("report-dashboard");
    await act(async () => {
      screen.getByTestId("report-replay-cta").click();
    });
    await waitFor(() => {
      expect(startSpy).toHaveBeenCalledTimes(1);
    });
    await waitFor(() => {
      expect(screen.queryByTestId("report-dashboard")).toBeNull();
    });
    expect(screen.queryByTestId("auth-login-screen")).toBeNull();
  });

  it("unauthenticated report route enters auth_login before mounting replay CTAs (TestReplayCtaPathA_UnauthenticatedUseRequestAuth)", async () => {
    const client = makeClient({ authenticated: false });
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: ROUTE_BASE,
        }}
      />,
    );
    await waitFor(() => {
      expect(screen.getByTestId("auth-login-email-form")).toBeInTheDocument();
    });
    expect(screen.getByTestId("auth-side-pending-action")).toBeInTheDocument();
    expect(screen.queryByTestId("report-dashboard")).toBeNull();
    expect(client.getFeedbackReport).not.toHaveBeenCalled();
    expect(client.startPracticeSession).not.toHaveBeenCalled();
  });

  it("path B (next round) CTA rotates roundId and auto-starts a fresh practice session (TestNextRoundCta_AutoStartPractice / TestNextRoundCta_NextRoundIdInference)", async () => {
    const client = makeClient({ authenticated: true });
    const createSpy = client.createPracticePlan as ReturnType<typeof vi.fn>;
    const startSpy = client.startPracticeSession as ReturnType<typeof vi.fn>;
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: ROUTE_BASE,
        }}
      />,
    );
    await screen.findByTestId("report-dashboard");
    await act(async () => {
      screen.getByTestId("report-next-cta").click();
    });
    await waitFor(() => {
      expect(startSpy).toHaveBeenCalledTimes(1);
    });
    expect(createSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        goal: "next_round",
        sourceReportId: REPORT_ID,
      }),
      expect.anything(),
    );
    await waitFor(() => {
      expect(screen.queryByTestId("report-dashboard")).toBeNull();
    });
  });
});

describe("Replay payload integrity", () => {
  it("buildReplayPayload includes the 9 owner / display knob fields and never raw text (TestReplayCtaPathA_PayloadIntegrity / NoRawText)", async () => {
    const { buildReplayPayload } = await import("../handoff");
    const payload = buildReplayPayload({
      route: {
        name: "report",
        params: ROUTE_BASE,
      },
      report: makeReport(),
      sessionId: SESSION_ID,
    });
    expect(payload).toMatchObject({
      sourceSessionId: SESSION_ID,
      replayItems: "turn-1,turn-3",
      evidenceGaps: "technical_depth",
      planId: "plan-1",
      targetJobId: TARGET_JOB_ID,
      jdId: "jd-1",
      resumeId: RESUME_VERSION_ID,
      sourceReportId: REPORT_ID,
      roundId: "round-tech-1",
      mode: "text",
      modality: "text",
      practiceMode: "strict",
      practiceGoal: "retry_current_round",
    });
    for (const value of Object.values(payload)) {
      expect(value).not.toMatch(/answerText/i);
      expect(value).not.toMatch(/questionText/i);
      expect(value).not.toMatch(/hint:/i);
      expect(value).not.toMatch(/promptHash/i);
      expect(value).not.toMatch(/modelId.*raw/i);
    }
  });

  it("buildNextRoundPayload rotates the roundId via inferNextRoundId (TestNextRoundCta_PayloadIntegrity)", async () => {
    const { buildNextRoundPayload, inferNextRoundId } = await import("../handoff");
    expect(inferNextRoundId("round-tech-1")).toBe("round-tech-2");
    expect(inferNextRoundId("round-tech-2")).toBe("round-manager");
    expect(inferNextRoundId("round-manager")).toBe("round-manager");
    expect(inferNextRoundId(undefined)).toBe("round-tech-2");

    const payload = buildNextRoundPayload({
      route: { name: "report", params: ROUTE_BASE },
      report: makeReport(),
      sessionId: SESSION_ID,
    });
    expect(payload.nextRoundId).toBe("round-tech-2");
    expect(payload.roundId).toBe("round-tech-2");
    expect(payload.practiceGoal).toBe("next_round");
    expect(payload.sourceReportId).toBe(REPORT_ID);
  });

  it("CTA click does not re-invoke getFeedbackReport or any listTargetJobReports call from report scope (TestReplayCtaPathA_NoReportReadCalls)", async () => {
    const client = makeClient({ authenticated: true });
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: ROUTE_BASE,
        }}
      />,
    );
    await screen.findByTestId("report-dashboard");
    const feedbackSpy = client.getFeedbackReport as ReturnType<typeof vi.fn>;
    const listSpy = client.listTargetJobReports as ReturnType<typeof vi.fn>;
    const callsBefore = feedbackSpy.mock.calls.length;
    await act(async () => {
      screen.getByTestId("report-replay-cta").click();
    });
    await waitFor(() => {
      expect(client.startPracticeSession).toHaveBeenCalled();
    });
    expect(feedbackSpy.mock.calls.length).toBe(callsBefore);
    expect(listSpy).not.toHaveBeenCalled();
  });
});
