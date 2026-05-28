/**
 * @vitest-environment jsdom
 *
 * Phase 2 — ReportScreen dispatch + dashboard data + privacy red lines.
 *  - reportStatus=failed → ReportFailureState
 *  - missing sessionId   → ReportMissingSessionState
 *  - happy path          → ReportDashboard (>=10 report-* testids)
 *  - cross-user 404      → not-found UI (separate from AI_* enum copy)
 *  - generated client receives no Idempotency-Key header
 *  - ContextStrip uses generated getTargetJob / getResumeVersion labels
 *  - listTargetJobReports is never invoked from report scope
 */

import {
  act,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { FC, ReactNode } from "react";

import type {
  ApiErrorCode,
  FeedbackReport,
  ResumeVersion,
  TargetJob,
} from "../../../../api/generated/types";
import { EasyInterviewClient } from "../../../../api/generated/client";
import { App } from "../../../App";
import type { LooseRoute } from "../../../normalizeRoute";

const REPORT_ID = "01918fa0-0000-7000-8000-000000007000";
const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";
const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";
const RESUME_VERSION_ID = "01918fa0-0000-7000-8000-000000004000";

function readyReport(overrides: Partial<FeedbackReport> = {}): FeedbackReport {
  return {
    id: REPORT_ID,
    sessionId: SESSION_ID,
    targetJobId: TARGET_JOB_ID,
    status: "ready",
    preparednessLevel: "basically_ready",
    highlights: [
      { dimension: "ownership", evidence: "strong evidence", confidence: "high" },
    ],
    issues: [
      {
        dimension: "technical_depth",
        evidence: "missing metrics",
        confidence: "medium",
      },
    ],
    nextActions: [{ type: "retry_current_round", label: "replay" }],
    questionAssessments: [
      {
        turnId: "turn-1",
        questionIntent: "design.api.versioning",
        dimensionResults: {
          ownership: { status: "strong", confidence: "high" },
          technical_depth: { status: "needs_work", confidence: "medium" },
        },
        reviewStatus: "queued_for_retry",
        includedInRetryPlan: true,
      },
      {
        turnId: "turn-2",
        questionIntent: "leadership.escalation",
        dimensionResults: {
          communication: { status: "meets_bar", confidence: "medium" },
        },
        reviewStatus: "resolved",
        includedInRetryPlan: false,
      },
    ],
    retryFocusTurnIds: ["turn-1"],
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
    ...overrides,
  };
}

interface ClientOptions {
  report?: FeedbackReport;
  reportReject?: unknown;
  targetJob?: TargetJob;
  targetJobReject?: unknown;
  resumeVersion?: ResumeVersion;
  resumeVersionReject?: unknown;
  authenticated?: boolean;
}

function makeClient(options: ClientOptions = {}): EasyInterviewClient {
  const targetJobFn = vi.fn(async () => {
    if (options.targetJobReject) throw options.targetJobReject;
    return (
      options.targetJob ??
      ({
        id: TARGET_JOB_ID,
        title: "Senior Frontend Engineer",
        companyName: "Acme",
        status: "ready",
      } as unknown as TargetJob)
    );
  });
  const resumeVersionFn = vi.fn(async () => {
    if (options.resumeVersionReject) throw options.resumeVersionReject;
    return (
      options.resumeVersion ??
      ({
        id: RESUME_VERSION_ID,
        displayName: "Resume v3",
      } as unknown as ResumeVersion)
    );
  });
  const feedbackReportFn = vi.fn(async (id: string, opts?: { headers?: Record<string, string> }) => {
    // The hook MUST NOT send Idempotency-Key — capture and assert in tests.
    if (opts?.headers && ("Idempotency-Key" in opts.headers || "idempotency-key" in opts.headers)) {
      throw new Error("read path leaked Idempotency-Key");
    }
    if (options.reportReject) throw options.reportReject;
    return options.report ?? readyReport();
  });
  const listTargetJobReportsFn = vi.fn(async () => {
    throw new Error("listTargetJobReports must not be invoked from report scope");
  });
  return {
    async getRuntimeConfig() {
      return { aiProviderProfile: "stub" } as never;
    },
    async getMe() {
      if (options.authenticated === false) {
        throw new Error("HTTP 401 Unauthorized");
      }
      return {
        id: "user-1",
        emailMasked: "u***@example.com",
        displayName: "User",
        profileCompletionRequired: false,
      } as never;
    },
    getFeedbackReport: feedbackReportFn,
    getTargetJob: targetJobFn,
    getResumeVersion: resumeVersionFn,
    listTargetJobReports: listTargetJobReportsFn,
  } as unknown as EasyInterviewClient;
}

function spies(client: EasyInterviewClient) {
  return {
    feedbackReport: client.getFeedbackReport as ReturnType<typeof vi.fn>,
    targetJob: client.getTargetJob as ReturnType<typeof vi.fn>,
    resumeVersion: client.getResumeVersion as ReturnType<typeof vi.fn>,
    listTargetJobReports: client.listTargetJobReports as ReturnType<typeof vi.fn>,
  };
}

const Harness: FC<{
  client: EasyInterviewClient;
  initialRoute: LooseRoute;
  children?: ReactNode;
}> = ({ client, initialRoute, children }) => (
  <App client={client} initialRoute={initialRoute}>
    {children}
  </App>
);

describe("ReportScreen dispatch", () => {
  it("dispatches ReportFailureState when reportStatus=failed (TestReportScreenDispatchesFailureState)", async () => {
    const client = makeClient();
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: {
            sessionId: SESSION_ID,
            reportId: REPORT_ID,
            reportStatus: "failed",
            errorCode: "AI_PROVIDER_TIMEOUT",
          },
        }}
      />,
    );
    expect(await screen.findByTestId("report-failure-state")).toBeInTheDocument();
    expect(screen.queryByTestId("report-dashboard")).toBeNull();
    expect(spies(client).feedbackReport).not.toHaveBeenCalled();
  });

  it("dispatches ReportMissingSessionState when sessionId is missing (TestReportScreenDispatchesMissingSession + TestReportMissingSessionNoApiCall)", async () => {
    const client = makeClient();
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: { reportId: REPORT_ID },
        }}
      />,
    );
    expect(await screen.findByTestId("report-missing-session")).toBeInTheDocument();
    expect(spies(client).feedbackReport).not.toHaveBeenCalled();
    expect(spies(client).listTargetJobReports).not.toHaveBeenCalled();
  });

  it("dispatches missing-report state when reportId is missing and never fetches (TestReportScreenDispatchesMissingReportId)", async () => {
    const client = makeClient();
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: { sessionId: SESSION_ID },
        }}
      />,
    );
    expect(await screen.findByTestId("report-missing-report")).toBeInTheDocument();
    expect(spies(client).feedbackReport).not.toHaveBeenCalled();
    expect(spies(client).listTargetJobReports).not.toHaveBeenCalled();
  });

  it("renders ReportDashboard on the happy path with 10+ report-* testids (TestReportScreenDispatchesDashboard)", async () => {
    const client = makeClient({ authenticated: true });
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: {
            sessionId: SESSION_ID,
            reportId: REPORT_ID,
            targetJobId: TARGET_JOB_ID,
            resumeVersionId: RESUME_VERSION_ID,
            roundId: "round-tech-1",
            roundName: "Round 1",
            planId: "plan-1",
            mode: "text",
            modality: "text",
            practiceMode: "strict",
            practiceGoal: "baseline",
            hintUsed: "false",
            hintCount: "0",
          },
        }}
      />,
    );
    await screen.findByTestId("report-dashboard");
    const required = [
      "report-dashboard",
      "report-back-button",
      "report-header",
      "report-header-title",
      "report-header-subtitle",
      "report-replay-cta",
      "report-next-cta",
      "report-context-strip",
      "report-context-session",
      "report-summary-cards",
      "report-summary-readiness",
      "report-summary-questions",
      "report-detail-surface",
      "report-detail-tab-questions",
      "report-detail-panel-questions",
    ];
    for (const id of required) {
      expect(screen.queryByTestId(id), `missing ${id}`).not.toBeNull();
    }
    await waitFor(() =>
      expect(screen.getByTestId("report-header-title")).toHaveTextContent(
        "Senior Frontend Engineer",
      ),
    );
    expect(screen.queryByTestId("route-report")).toBeNull();
    expect(screen.queryByTestId("mistakes-queue")).toBeNull();
    expect(screen.queryByTestId("report-timeline")).toBeNull();
  });

  it("getFeedbackReport is read-only and never leaks Idempotency-Key (TestUseFeedbackReportNoIdempotencyHeader)", async () => {
    const client = makeClient();
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: {
            sessionId: SESSION_ID,
            reportId: REPORT_ID,
          },
        }}
      />,
    );
    await screen.findByTestId("report-dashboard");
    expect(spies(client).feedbackReport).toHaveBeenCalled();
    // listTargetJobReports must NEVER be invoked from report scope (dashboard-only D-7).
    expect(spies(client).listTargetJobReports).not.toHaveBeenCalled();
  });

  it("HTTP 404 from getFeedbackReport surfaces the not-found UI with dedicated copy (TestUseFeedbackReportCrossUser404 + TestReportFailureStateRendersNotFoundCopy)", async () => {
    const client = makeClient({
      reportReject: new Error("HTTP 404 Not Found"),
    });
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: {
            sessionId: SESSION_ID,
            reportId: REPORT_ID,
          },
        }}
      />,
    );
    const notFound = await screen.findByTestId("report-failure-state");
    expect(notFound.getAttribute("data-not-found")).toBe("true");
    expect(screen.getByTestId("report-failure-state-not-found-title")).toBeInTheDocument();
    expect(
      screen.queryByText(/AI service timeout/i),
    ).toBeNull();
  });

  it("ContextStrip displays target / resume labels via generated client + falls back to ID when getTargetJob fails (TestReportContextDataLoadsTargetJobAndResumeVersion / TestReportContextDataFallsBackToIds)", async () => {
    const client = makeClient({
      targetJobReject: new Error("HTTP 500 Internal"),
    });
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: {
            sessionId: SESSION_ID,
            reportId: REPORT_ID,
            targetJobId: TARGET_JOB_ID,
            resumeVersionId: RESUME_VERSION_ID,
          },
        }}
      />,
    );
    await screen.findByTestId("report-dashboard");
    await waitFor(() => {
      const job = screen.getByTestId("report-context-job");
      expect(job.textContent).toContain(TARGET_JOB_ID);
    });
    await waitFor(() => {
      const resume = screen.getByTestId("report-context-resume");
      expect(resume.textContent).toContain("Resume v3");
    });
    expect(spies(client).targetJob).toHaveBeenCalled();
    expect(spies(client).resumeVersion).toHaveBeenCalled();
  });

  it("ContextStrip never reads raw resume / JD body fields (privacy red line)", async () => {
    const sensitiveResume = {
      id: RESUME_VERSION_ID,
      displayName: "Resume v3",
      originalText: "PRIVATE: do-not-leak resume body",
      parsedTextSnapshot: "PRIVATE: parsed snapshot",
    } as unknown as ResumeVersion;
    const sensitiveJob = {
      id: TARGET_JOB_ID,
      title: "Senior Frontend Engineer",
      companyName: "Acme",
      jdText: "PRIVATE: JD body",
      rawJd: "PRIVATE: raw JD",
    } as unknown as TargetJob;
    const client = makeClient({
      targetJob: sensitiveJob,
      resumeVersion: sensitiveResume,
    });
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: {
            sessionId: SESSION_ID,
            reportId: REPORT_ID,
            targetJobId: TARGET_JOB_ID,
            resumeVersionId: RESUME_VERSION_ID,
          },
        }}
      />,
    );
    const strip = await screen.findByTestId("report-context-strip");
    expect(strip.textContent ?? "").not.toContain("PRIVATE");
  });
});
