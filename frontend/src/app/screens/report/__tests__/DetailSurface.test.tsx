/**
 * @vitest-environment jsdom
 *
 * Phase 3 — DetailSurface 5-tab gate. Switching tabs swaps the panel,
 * default tab is `questions`, ARIA tablist semantics hold, dashboard body
 * sections render (dim row, top priority, next practice, perq, issues,
 * highlights).
 */

import {
  act,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { FC, ReactNode } from "react";

import type { FeedbackReport } from "../../../../api/generated/types";
import { EasyInterviewClient } from "../../../../api/generated/client";
import { App } from "../../../App";
import type { LooseRoute } from "../../../normalizeRoute";

const REPORT_ID = "01918fa0-0000-7000-8000-000000007000";
const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";

const FULL_REPORT: FeedbackReport = {
  id: REPORT_ID,
  sessionId: SESSION_ID,
  targetJobId: "01918fa0-0000-7000-8000-000000002000",
  status: "ready",
  preparednessLevel: "basically_ready",
  highlights: [
    { dimension: "ownership", evidence: "ownership detail", confidence: "high" },
    { dimension: "narrative", evidence: "narrative detail", confidence: "high" },
  ],
  issues: [
    { dimension: "technical_depth", evidence: "missing metrics", confidence: "medium" },
    { dimension: "conflict_resolution", evidence: "depth gap", confidence: "low" },
  ],
  nextActions: [
    { type: "retry_current_round", label: "rerun current round" },
    { type: "review_evidence", label: "review evidence" },
    { type: "next_round", label: "advance to next round" },
  ],
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
      reviewStatus: "open",
      includedInRetryPlan: false,
    },
    {
      turnId: "turn-3",
      questionIntent: "behavioral.ownership",
      dimensionResults: {
        ownership: { status: "strong", confidence: "high" },
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
};

function makeClient(report = FULL_REPORT): EasyInterviewClient {
  return {
    async getRuntimeConfig() {
      return { aiProviderProfile: "stub" } as never;
    },
    async getMe() {
      throw new Error("HTTP 401 Unauthorized");
    },
    getFeedbackReport: vi.fn(async () => report),
    getTargetJob: vi.fn(async () => ({
      id: "01918fa0-0000-7000-8000-000000002000",
      title: "Senior Frontend Engineer",
      companyName: "Acme",
    })),
    getResumeVersion: vi.fn(async () => ({
      id: "01918fa0-0000-7000-8000-000000004000",
      displayName: "Resume v3",
    })),
    listTargetJobReports: vi.fn(async () => {
      throw new Error("must not call from report scope");
    }),
  } as unknown as EasyInterviewClient;
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

describe("DetailSurface", () => {
  it("renders ARIA tablist with the questions panel active by default (TestDetailSurfaceDefaultQuestions + TestDetailSurfaceAriaTablist)", async () => {
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
    const tablist = screen.getByRole("tablist");
    expect(tablist).toBeInTheDocument();
    const questionsTab = screen.getByTestId("report-detail-tab-questions");
    expect(questionsTab.getAttribute("aria-selected")).toBe("true");
    expect(screen.getByTestId("report-detail-panel-questions")).toBeInTheDocument();
  });

  it("switches all 5 tabs in order and only the active panel renders content (TestDetailSurfaceSwitches5Tabs)", async () => {
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

    const tabs = ["readiness", "dimensions", "questions", "evidence", "next"] as const;
    for (const key of tabs) {
      const button = screen.getByTestId(`report-detail-tab-${key}`);
      await act(async () => {
        button.click();
      });
      const panel = screen.getByTestId(`report-detail-panel-${key}`);
      expect(panel.hasAttribute("hidden")).toBe(false);
      const other = tabs.filter((other) => other !== key);
      for (const sibling of other) {
        expect(
          screen.getByTestId(`report-detail-panel-${sibling}`).hasAttribute("hidden"),
        ).toBe(true);
      }
    }
  });

  it("renders readiness dial + 3 detail rows when readiness is the active tab (TestReadinessTabDial)", async () => {
    const client = makeClient();
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: { sessionId: SESSION_ID, reportId: REPORT_ID },
        }}
      />,
    );
    await screen.findByTestId("report-dashboard");
    await act(async () => {
      screen.getByTestId("report-detail-tab-readiness").click();
    });
    expect(screen.getByTestId("report-readiness-dial")).toBeInTheDocument();
    expect(screen.getByTestId("report-readiness-jd-align")).toBeInTheDocument();
    expect(screen.getByTestId("report-readiness-evidence-density")).toBeInTheDocument();
    expect(screen.getByTestId("report-readiness-next-threshold")).toBeInTheDocument();
  });

  it("renders dimensions grid + at least one DimRow on the dimensions tab (TestDimensionsTabGrid + TestDimensionStatus3StatesMapping)", async () => {
    const client = makeClient();
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: { sessionId: SESSION_ID, reportId: REPORT_ID },
        }}
      />,
    );
    await screen.findByTestId("report-dashboard");
    await act(async () => {
      screen.getByTestId("report-detail-tab-dimensions").click();
    });
    const grid = await screen.findByTestId("report-dimensions-grid");
    const cards = within(grid).getAllByTestId(/report-dim-card-\d+/);
    expect(cards.length).toBeGreaterThan(0);
    const statuses = cards
      .map((card) => card.getAttribute("data-dim-status"))
      .filter((value): value is string => Boolean(value));
    expect(statuses.length).toBeGreaterThan(0);
    for (const status of statuses) {
      expect(["strong", "meets_bar", "needs_work", "unknown"]).toContain(status);
    }
  });

  it("renders questions list + active detail with add-to-replay CTA (TestQuestionsTabListAndDetail)", async () => {
    const client = makeClient();
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: { sessionId: SESSION_ID, reportId: REPORT_ID },
        }}
      />,
    );
    await screen.findByTestId("report-dashboard");
    expect(screen.getByTestId("report-questions-list")).toBeInTheDocument();
    expect(screen.getByTestId("report-questions-detail-topic")).toBeInTheDocument();
    expect(screen.getByTestId("report-questions-add-to-replay")).toBeInTheDocument();
  });

  it("renders evidence risk + highlight columns (TestEvidenceTabRiskAndHighlight)", async () => {
    const client = makeClient();
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: { sessionId: SESSION_ID, reportId: REPORT_ID },
        }}
      />,
    );
    await screen.findByTestId("report-dashboard");
    await act(async () => {
      screen.getByTestId("report-detail-tab-evidence").click();
    });
    expect(screen.getByTestId("report-evidence-risk-0")).toBeInTheDocument();
    expect(screen.getByTestId("report-evidence-highlight-0")).toBeInTheDocument();
  });

  it("next tab renders both path A and path B columns with CTAs (TestNextTabPathAAndB)", async () => {
    const client = makeClient();
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: { sessionId: SESSION_ID, reportId: REPORT_ID },
        }}
      />,
    );
    await screen.findByTestId("report-dashboard");
    await act(async () => {
      screen.getByTestId("report-detail-tab-next").click();
    });
    expect(screen.getByTestId("report-next-path-a")).toBeInTheDocument();
    expect(screen.getByTestId("report-next-path-b")).toBeInTheDocument();
    expect(screen.getByTestId("report-next-cta-a")).toBeInTheDocument();
    expect(screen.getByTestId("report-next-cta-b")).toBeInTheDocument();
  });

  it("dashboard body renders dim row / top priority / next practice / perq / issue / highlight rows (TestDashboardDimensionsCardRow + TestDashboardQuestionRecap + TestDashboardIssuesAndHighlights)", async () => {
    const client = makeClient();
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: { sessionId: SESSION_ID, reportId: REPORT_ID },
        }}
      />,
    );
    await screen.findByTestId("report-dashboard");
    expect(screen.getByTestId("report-dim-row-0")).toBeInTheDocument();
    expect(screen.getByTestId("report-top-priority")).toBeInTheDocument();
    expect(screen.getByTestId("report-next-practice-0")).toBeInTheDocument();
    expect(screen.getByTestId("report-perq-0")).toBeInTheDocument();
    expect(screen.getByTestId("report-issue-0")).toBeInTheDocument();
    expect(screen.getByTestId("report-highlight-0")).toBeInTheDocument();
  });

  it("EmptyHint surfaces when issues / highlights / questions are empty (TestEvidenceTabRiskAndHighlight empty branch)", async () => {
    const emptyReport: FeedbackReport = {
      ...FULL_REPORT,
      questionAssessments: [],
      issues: [],
      highlights: [],
      nextActions: [],
      retryFocusTurnIds: [],
    };
    const client = makeClient(emptyReport);
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "report",
          params: { sessionId: SESSION_ID, reportId: REPORT_ID },
        }}
      />,
    );
    await screen.findByTestId("report-dashboard");
    expect(screen.getByTestId("report-body-questions-empty")).toBeInTheDocument();
    expect(screen.getByTestId("report-body-issues-empty")).toBeInTheDocument();
    expect(screen.getByTestId("report-body-highlights-empty")).toBeInTheDocument();
    await act(async () => {
      screen.getByTestId("report-detail-tab-evidence").click();
    });
    expect(screen.getByTestId("report-evidence-risk-empty")).toBeInTheDocument();
    expect(screen.getByTestId("report-evidence-highlight-empty")).toBeInTheDocument();
  });
});
