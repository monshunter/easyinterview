/** @vitest-environment jsdom */

import { act, render, screen, waitFor } from "@testing-library/react";
import type { FC, ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

import type {
  CreatePracticePlanRequest,
  FeedbackReport,
  PracticeGoal,
  PracticePlan,
  PracticeSession,
} from "../../../../api/generated/types";
import { EasyInterviewClient } from "../../../../api/generated/client";
import { App } from "../../../App";
import type { LooseRoute } from "../../../normalizeRoute";

const REPORT_ID = "01918fa0-0000-7000-8000-000000007000";
const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";
const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";
const RESUME_ID = "01918fa0-0000-7000-8000-000000004000";
const PLAN_ID = "01918fa0-0000-7000-8000-000000008000";
const STARTED_SESSION_ID = "01918fa0-0000-7000-8000-000000009000";

const BASE_CONTEXT: FeedbackReport["context"] = {
  sourcePlanId: "01918fa0-0000-7000-8000-000000006000",
  targetJobTitle: "Senior Frontend Engineer",
  targetJobCompany: "Acme",
  resumeId: RESUME_ID,
  resumeDisplayName: "Frontend resume",
  roundId: "round-1-technical",
  roundSequence: 1,
  roundName: "Technical one",
  roundType: "technical",
  language: "zh-CN",
  hasNextRound: true,
};

function makeReport(overrides: Partial<FeedbackReport> = {}): FeedbackReport {
  return {
    id: REPORT_ID,
    sessionId: SESSION_ID,
    targetJobId: TARGET_JOB_ID,
    status: "ready",
    errorCode: null,
    summary: "回答结构清楚，但技术取舍仍需要可核验的对比证据。",
    context: BASE_CONTEXT,
    preparednessLevel: "needs_practice",
    dimensionAssessments: [
      { code: "technical_depth", label: "技术深度", status: "needs_work", confidence: "medium" },
    ],
    highlights: [],
    issues: [
      { dimensionCode: "technical_depth", evidence: "没有给出替代方案的成本和结果指标。", confidence: "medium" },
    ],
    nextActions: [
      { type: "retry_current_round", label: "补齐技术取舍证据后复练当前轮。" },
      { type: "next_round", label: "完成补强后进入下一轮。" },
    ],
    retryFocusDimensionCodes: ["technical_depth"],
    provenance: {
      promptVersion: "v0.2.0",
      rubricVersion: "v0.2.0",
      modelId: "model-profile:contract.default",
      language: "zh-CN",
      featureFlag: "none",
      dataSourceVersion: "report-context.v1",
    },
    createdAt: "2026-05-16T00:00:00Z",
    updatedAt: "2026-05-16T00:00:10Z",
    ...overrides,
  };
}

function planFor(goal: PracticeGoal): PracticePlan {
  const next = goal === "next_round";
  return {
    id: PLAN_ID,
    targetJobId: TARGET_JOB_ID,
    resumeId: RESUME_ID,
    goal,
    sourceReportId: REPORT_ID,
    interviewerPersona: "hiring_manager",
    difficulty: "standard",
    language: "zh-CN",
    timeBudgetMinutes: next ? 60 : 45,
    status: "ready",
    roundId: next ? "round-2-technical" : "round-1-technical",
    roundSequence: next ? 2 : 1,
    createdAt: "2026-05-16T00:01:00Z",
  };
}

function startedSession(): PracticeSession {
  return {
    id: STARTED_SESSION_ID,
    planId: PLAN_ID,
    targetJobId: TARGET_JOB_ID,
    language: "zh-CN",
    status: "running",
    messages: [],
    createdAt: "2026-05-16T00:01:10Z",
    updatedAt: "2026-05-16T00:01:10Z",
  };
}

function makeClient(options: {
  authenticated: boolean;
  report?: FeedbackReport;
}): EasyInterviewClient {
  const value = options.report ?? makeReport();
  return {
    async getRuntimeConfig() {
      return { aiProviderProfile: "stub" } as never;
    },
    async getMe() {
      if (options.authenticated) return { id: "user-1", email: "u@example.com" } as never;
      throw new Error("HTTP 401 Unauthorized");
    },
    getFeedbackReport: vi.fn(async () => value),
    getTargetJob: vi.fn(async () => { throw new Error("mutable target read is forbidden"); }),
    getResume: vi.fn(async () => { throw new Error("mutable resume read is forbidden"); }),
    listTargetJobReports: vi.fn(async () => { throw new Error("report list read is forbidden"); }),
    createPracticePlan: vi.fn(async (body: CreatePracticePlanRequest) => planFor(body.goal)),
    startPracticeSession: vi.fn(async () => startedSession()),
    getPracticePlan: vi.fn(async () => planFor("retry_current_round")),
    getPracticeSession: vi.fn(async () => startedSession()),
  } as unknown as EasyInterviewClient;
}

const Harness: FC<{
  client: EasyInterviewClient;
  initialRoute: LooseRoute;
  children?: ReactNode;
}> = ({ client, initialRoute, children }) => (
  <App client={client} initialRoute={initialRoute}>{children}</App>
);

describe("report-derived practice CTAs", () => {
  it("replay sends only goal + sourceReportId and starts the backend-derived plan", async () => {
    const client = makeClient({ authenticated: true });
    render(<Harness client={client} initialRoute={{ name: "report", params: { reportId: REPORT_ID } }} />);

    await screen.findByTestId("report-dashboard");
    expect(client.getTargetJob).not.toHaveBeenCalled();
    expect(client.getResume).not.toHaveBeenCalled();
    await act(async () => screen.getByTestId("report-replay-cta").click());

    await waitFor(() => expect(client.startPracticeSession).toHaveBeenCalledTimes(1));
    expect(client.createPracticePlan).toHaveBeenCalledWith(
      { goal: "retry_current_round", sourceReportId: REPORT_ID },
      expect.anything(),
    );
  });

  it("next-round sends only goal + sourceReportId when frozen context allows it", async () => {
    const client = makeClient({ authenticated: true });
    render(<Harness client={client} initialRoute={{ name: "report", params: { reportId: REPORT_ID } }} />);

    await screen.findByTestId("report-dashboard");
    expect(client.getTargetJob).not.toHaveBeenCalled();
    expect(screen.getByTestId("report-next-cta")).toBeEnabled();
    await act(async () => screen.getByTestId("report-next-cta").click());

    await waitFor(() => expect(client.startPracticeSession).toHaveBeenCalledTimes(1));
    expect(client.createPracticePlan).toHaveBeenCalledWith(
      { goal: "next_round", sourceReportId: REPORT_ID },
      expect.anything(),
    );
  });

  it("uses frozen hasNextRound=false as the accessible terminal gate", async () => {
    const report = makeReport({
      context: { ...BASE_CONTEXT, hasNextRound: false },
      nextActions: [{ type: "retry_current_round", label: "复练当前轮。" }],
    });
    const client = makeClient({ authenticated: true, report });
    render(<Harness client={client} initialRoute={{ name: "report", params: { reportId: REPORT_ID } }} />);

    await screen.findByTestId("report-dashboard");
    const next = screen.getByTestId("report-next-cta");
    expect(next).toBeDisabled();
    expect(next).toHaveAttribute("aria-describedby", "report-next-disabled-reason");
    expect(screen.getByTestId("report-next-disabled-reason")).not.toHaveTextContent("");
    next.click();
    expect(client.createPracticePlan).not.toHaveBeenCalled();
  });

  it("locks both CTAs synchronously and creates at most one plan", async () => {
    const client = makeClient({ authenticated: true });
    let resolvePlan!: (value: PracticePlan) => void;
    client.createPracticePlan = vi.fn(() => new Promise<PracticePlan>((resolve) => {
      resolvePlan = resolve;
    })) as never;
    render(<Harness client={client} initialRoute={{ name: "report", params: { reportId: REPORT_ID } }} />);

    await screen.findByTestId("report-dashboard");
    const replay = screen.getByTestId("report-replay-cta");
    await act(async () => {
      replay.click();
      replay.click();
    });
    await waitFor(() => expect(client.createPracticePlan).toHaveBeenCalledTimes(1));
    expect(replay).toBeDisabled();
    expect(screen.getByTestId("report-next-cta")).toBeDisabled();

    await act(async () => resolvePlan(planFor("retry_current_round")));
    await waitFor(() => expect(client.startPracticeSession).toHaveBeenCalledTimes(1));
  });

  it("unauthenticated entry stays behind auth and does not read or mutate report state", async () => {
    const client = makeClient({ authenticated: false });
    render(<Harness client={client} initialRoute={{ name: "report", params: { reportId: REPORT_ID } }} />);

    expect(await screen.findByTestId("auth-login-email-form")).toBeInTheDocument();
    expect(screen.getByTestId("auth-side-pending-action")).toBeInTheDocument();
    expect(client.getFeedbackReport).not.toHaveBeenCalled();
    expect(client.createPracticePlan).not.toHaveBeenCalled();
  });
});

describe("report-derived payload integrity", () => {
  it("builds exact report-owner requests with no route-selected context", async () => {
    const { buildNextRoundPayload, buildReplayPayload } = await import("../handoff");
    const report = makeReport();
    expect(buildReplayPayload({ report })).toEqual({
      goal: "retry_current_round",
      sourceReportId: REPORT_ID,
    });
    expect(buildNextRoundPayload({ report })).toEqual({
      goal: "next_round",
      sourceReportId: REPORT_ID,
    });
  });

  it("does not re-read report or mutable context when starting replay", async () => {
    const client = makeClient({ authenticated: true });
    render(<Harness client={client} initialRoute={{ name: "report", params: { reportId: REPORT_ID } }} />);

    await screen.findByTestId("report-dashboard");
    const readsBefore = vi.mocked(client.getFeedbackReport).mock.calls.length;
    expect(client.getTargetJob).not.toHaveBeenCalled();
    expect(client.getResume).not.toHaveBeenCalled();
    expect(client.listTargetJobReports).not.toHaveBeenCalled();
    await act(async () => screen.getByTestId("report-replay-cta").click());
    await waitFor(() => expect(client.startPracticeSession).toHaveBeenCalledTimes(1));

    expect(vi.mocked(client.getFeedbackReport).mock.calls).toHaveLength(readsBefore);
    expect(client.listTargetJobReports).not.toHaveBeenCalled();
  });
});
