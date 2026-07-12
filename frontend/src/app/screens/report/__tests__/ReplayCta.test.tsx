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
import type { FC, ReactNode } from "react";

import type { FeedbackReport } from "../../../../api/generated/types";
import { EasyInterviewClient } from "../../../../api/generated/client";
import { App } from "../../../App";
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
    dimensionAssessments: [
      { dimension: "technical_depth", status: "needs_work", confidence: "medium" },
    ],
    retryFocusCompetencyCodes: ["technical_depth"],
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
  targetJob?: ReturnType<typeof makeTargetJob>;
}

function makeTargetJob(
  interviewRounds = [
    {
      sequence: 1,
      type: "technical",
      name: "Technical one",
      durationMinutes: 45,
      focus: "Coding",
    },
    {
      sequence: 2,
      type: "technical",
      name: "Technical two",
      durationMinutes: 60,
      focus: "Architecture",
    },
  ],
) {
  return {
    id: TARGET_JOB_ID,
    analysisStatus: "ready",
    title: "Senior Frontend Engineer",
    companyName: "Acme",
    locationText: "Remote",
    targetLanguage: "zh-CN",
    sourceType: "manual_text",
    sourceUrl: null,
    summary: {
      coreThemes: [],
      interviewRounds,
      provenance: {
        promptVersion: "target_job.v1",
        rubricVersion: "target_job.v1",
        modelId: "fixture",
        language: "en",
        featureFlag: "none",
        dataSourceVersion: "fixture",
      },
    },
    requirements: [],
    latestReportId: REPORT_ID,
    openQuestionIssueCount: 0,
    status: "ready",
    createdAt: "2026-05-16T00:00:00Z",
    updatedAt: "2026-05-16T00:00:00Z",
  } as const;
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
    getTargetJob: vi.fn(async () => opts.targetJob ?? makeTargetJob()),
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

const ROUTE_BASE: Record<string, string> = {
  sessionId: SESSION_ID,
  reportId: REPORT_ID,
  targetJobId: TARGET_JOB_ID,
  resumeId: RESUME_VERSION_ID,
  roundId: "round-1-technical",
  planId: "plan-1",
  jdId: "jd-1",
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
  it("authenticated user clicking replay CTA creates a fresh practice session directly (TestReplayCtaPathA_AuthenticatedDirectStartPractice)", async () => {
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

  it("path B (next round) CTA rotates roundId and directly starts a fresh practice session (TestNextRoundCta_DirectStartPractice / TestNextRoundCta_NextRoundIdInference)", async () => {
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
    await waitFor(() => {
      expect(screen.getByTestId("report-next-cta")).not.toBeDisabled();
    });
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

  it("keeps next round disabled while structured rounds are loading or failed", async () => {
    const client = makeClient({ authenticated: true });
    let rejectTarget!: (reason: unknown) => void;
    client.getTargetJob = vi.fn(() => new Promise<never>((_, reject) => {
      rejectTarget = reject;
    }));
    render(
      <Harness
        client={client}
        initialRoute={{ name: "report", params: ROUTE_BASE }}
      />,
    );
    await screen.findByTestId("report-dashboard");
    expect(screen.getByTestId("report-next-cta")).toBeDisabled();
    await act(async () => {
      rejectTarget(new Error("HTTP 500 Internal"));
    });
    await waitFor(() => {
      expect(client.getTargetJob).toHaveBeenCalled();
      expect(screen.getByTestId("report-next-cta")).toBeDisabled();
    });
    expect(client.createPracticePlan).not.toHaveBeenCalled();
    expect(client.startPracticeSession).not.toHaveBeenCalled();
  });

  it.each([
    ["final round", { ...ROUTE_BASE, roundId: "round-2-technical" }, makeTargetJob()],
    ["unknown round", { ...ROUTE_BASE, roundId: "round-99-technical" }, makeTargetJob()],
    [
      "duplicate derived round ids",
      ROUTE_BASE,
      makeTargetJob([
        { sequence: 1, type: "technical", name: "A", durationMinutes: 45, focus: "A" },
        { sequence: 1, type: "technical", name: "B", durationMinutes: 60, focus: "B" },
      ]),
    ],
  ])("fails closed for %s", async (_name, params, targetJob) => {
    const client = makeClient({ authenticated: true, targetJob });
    render(
      <Harness
        client={client}
        initialRoute={{ name: "report", params }}
      />,
    );
    await screen.findByTestId("report-dashboard");
    await waitFor(() => expect(client.getTargetJob).toHaveBeenCalled());
    const next = screen.getByTestId("report-next-cta");
    expect(next).toBeDisabled();
    next.click();
    expect(client.createPracticePlan).not.toHaveBeenCalled();
    expect(client.startPracticeSession).not.toHaveBeenCalled();
  });

  it("locks both CTAs synchronously and creates at most one plan/session for repeated clicks", async () => {
    const client = makeClient({ authenticated: true });
    let resolvePlan!: (value: { id: string; status: string }) => void;
    client.createPracticePlan = vi.fn(() => new Promise((resolve) => {
      resolvePlan = resolve;
    })) as never;
    const createSpy = client.createPracticePlan as ReturnType<typeof vi.fn>;
    render(
      <Harness
        client={client}
        initialRoute={{ name: "report", params: ROUTE_BASE }}
      />,
    );
    await screen.findByTestId("report-dashboard");
    await waitFor(() => {
      expect(screen.getByTestId("report-next-cta")).not.toBeDisabled();
    });
    const replay = screen.getByTestId("report-replay-cta");
    const next = screen.getByTestId("report-next-cta");
    await act(async () => {
      replay.click();
      replay.click();
    });
    await waitFor(() => expect(createSpy).toHaveBeenCalledTimes(1));
    expect(replay).toBeDisabled();
    expect(next).toBeDisabled();
    await act(async () => {
      resolvePlan({ id: "01918fa0-0000-7000-8000-000000008000", status: "ready" });
    });
    await waitFor(() => expect(client.startPracticeSession).toHaveBeenCalledTimes(1));
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
      focusCompetencyCodes: "technical_depth",
      evidenceGaps: "technical_depth",
      planId: "plan-1",
      targetJobId: TARGET_JOB_ID,
      jdId: "jd-1",
      resumeId: RESUME_VERSION_ID,
      sourceReportId: REPORT_ID,
      roundId: "round-1-technical",
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

  it("buildNextRoundPayload uses the resolved structured next round (TestNextRoundCta_PayloadIntegrity)", async () => {
    const { buildNextRoundPayload } = await import("../handoff");
    const payload = buildNextRoundPayload({
      route: { name: "report", params: ROUTE_BASE },
      report: makeReport(),
      sessionId: SESSION_ID,
    }, {
      id: "round-2-technical",
      name: "Technical two · 60m",
      focus: "Architecture",
      type: "technical",
      durationMinutes: 60,
    });
    expect(payload.nextRoundId).toBe("round-2-technical");
    expect(payload.roundId).toBe("round-2-technical");
    expect(payload.roundName).toBe("Technical two · 60m");
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
