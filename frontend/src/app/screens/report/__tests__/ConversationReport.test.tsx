/** @vitest-environment jsdom */
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { EasyInterviewClient } from "../../../../api/generated/client";
import type { FeedbackReport } from "../../../../api/generated/types";
import { App } from "../../../App";

const REPORT_ID = "01918fa0-0000-7000-8000-000000007000";
const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";

function report(overrides: Partial<FeedbackReport> = {}): FeedbackReport {
  return {
    id: REPORT_ID,
    sessionId: SESSION_ID,
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    status: "ready",
    errorCode: null,
    summary: "模型原文：你的案例有清楚证据，但技术取舍需要补充。",
    preparednessLevel: "needs_practice",
    context: {
      sourcePlanId: "01918fa0-0000-7000-8000-000000004000",
      targetJobTitle: "很长的高级前端工程师岗位名称用于验证完整换行",
      targetJobCompany: "星环科技",
      resumeId: "01918fa0-0000-7000-8000-000000001000",
      resumeDisplayName: "前端工程师简历 · 包含完整项目证据的长名称",
      roundId: "round-2-technical",
      roundSequence: 2,
      roundName: "技术一面 · 系统设计与工程取舍",
      roundType: "technical",
      language: "zh-CN",
      hasNextRound: true,
    },
    dimensionAssessments: [
      { code: "answer_structure", label: "回答结构", status: "strong", confidence: "high" },
      { code: "technical_tradeoffs", label: "技术取舍", status: "needs_work", confidence: "medium" },
    ],
    highlights: [
      { dimensionCode: "answer_structure", evidence: "按背景、行动、结果说明了跨团队推进路径。", confidence: "high" },
    ],
    issues: [
      { dimensionCode: "technical_tradeoffs", evidence: "没有比较替代方案的成本与风险。", confidence: "medium" },
    ],
    nextActions: [
      { type: "next_round", label: "模型原文：补齐一个取舍案例后进入下一轮。" },
      { type: "retry_current_round", label: "模型原文：也可以先复练当前轮。" },
    ],
    retryFocusDimensionCodes: ["technical_tradeoffs"],
    provenance: {
      promptVersion: "v0.2.0",
      rubricVersion: "v0.2.0",
      modelId: "model-profile:contract.default",
      language: "zh-CN",
      featureFlag: "none",
      dataSourceVersion: "report-context.v1",
    },
    createdAt: "2026-07-12T08:30:00Z",
    updatedAt: "2026-07-12T08:31:00Z",
    ...overrides,
  };
}

function failedReport(overrides: Partial<FeedbackReport> = {}): FeedbackReport {
  return report({
    status: "failed",
    errorCode: "AI_PROVIDER_TIMEOUT",
    summary: null,
    preparednessLevel: null,
    dimensionAssessments: [],
    highlights: [],
    issues: [],
    nextActions: [],
    retryFocusDimensionCodes: [],
    provenance: null,
    ...overrides,
  });
}

function clientFor(value: FeedbackReport) {
  const getTargetJob = vi.fn(async () => { throw new Error("mutable target must not load"); });
  const getResume = vi.fn(async () => { throw new Error("mutable resume must not load"); });
  return {
    client: {
      async getRuntimeConfig() { return { aiProviderProfile: "stub" } as never; },
      async getMe() {
        return {
          id: "user-1",
          displayName: "Tester",
          email: "test@example.com",
          profileCompletionRequired: false,
        } as never;
      },
      async getFeedbackReport() { return value; },
      async listTargetJobs() {
        return {
          items: [],
          pageInfo: { hasNextPage: false, nextCursor: null },
        } as never;
      },
      async listTargetJobReports() {
        throw new Error("reports page data is outside this report-return test");
      },
      getTargetJob,
      getResume,
    } as unknown as EasyInterviewClient,
    getTargetJob,
    getResume,
  };
}

afterEach(() => {
  localStorage.removeItem("ei-lang");
  window.history.replaceState(null, "", "/");
});

describe("grounded direct-semantic feedback report", () => {
  it("returns a ready report to the API-trusted reports page", async () => {
    const trustedTargetJobId = "01918fa0-0000-7000-8000-000000002000";
    const { client } = clientFor(report({ targetJobId: trustedTargetJobId }));
    window.history.replaceState(
      null,
      "",
      `/report?reportId=${REPORT_ID}&targetJobId=route-target-must-be-ignored`,
    );

    render(<App client={client} />);

    fireEvent.click(await screen.findByTestId("report-back-button"));
    await waitFor(() => {
      expect(window.location.pathname + window.location.search).toBe(
        `/reports?targetJobId=${trustedTargetJobId}`,
      );
    });
  });

  it("returns a valid failed report to the same API-trusted reports page", async () => {
    const trustedTargetJobId = "01918fa0-0000-7000-8000-000000002000";
    const { client } = clientFor(failedReport({ targetJobId: trustedTargetJobId }));
    window.history.replaceState(null, "", `/report?reportId=${REPORT_ID}`);

    render(<App client={client} />);

    fireEvent.click(await screen.findByTestId("report-failure-back-to-workspace"));
    await waitFor(() => {
      expect(window.location.pathname + window.location.search).toBe(
        `/reports?targetJobId=${trustedTargetJobId}`,
      );
    });
  });

  it("keeps the trusted reports return visible while a report is pending", async () => {
    const trustedTargetJobId = "01918fa0-0000-7000-8000-000000002000";
    const { client } = clientFor(
      failedReport({
        status: "generating",
        errorCode: null,
        targetJobId: trustedTargetJobId,
      }),
    );
    window.history.replaceState(null, "", `/report?reportId=${REPORT_ID}`);

    render(<App client={client} />);

    fireEvent.click(await screen.findByTestId("report-pending-back-button"));
    await waitFor(() => {
      expect(window.location.pathname + window.location.search).toBe(
        `/reports?targetJobId=${trustedTargetJobId}`,
      );
    });
  });

  it.each([
    ["malformed queued", { ...failedReport({ status: "queued", errorCode: null }), context: null }],
    ["malformed generating", { ...failedReport({ status: "generating", errorCode: null }), routeTarget: "must-not-survive" }],
    ["unknown status", { ...failedReport(), status: "unknown" }],
    ["invalid ready", { ...report(), summary: null }],
    ["failed unknown error", { ...failedReport(), errorCode: "REPORT_UNKNOWN_FAILURE" }],
  ])("renders typed invalid terminal for %s without trusting its target", async (_label, value) => {
    const { client } = clientFor(value as unknown as FeedbackReport);
    window.history.replaceState(null, "", `/report?reportId=${REPORT_ID}`);

    render(<App client={client} />);

    const failure = await screen.findByTestId("report-failure-state");
    expect(failure).toHaveAttribute("data-contract-invalid", "true");
    expect(screen.queryByTestId("report-pending-state")).not.toBeInTheDocument();
    expect(screen.queryByTestId("report-dashboard")).not.toBeInTheDocument();
    fireEvent.click(screen.getByTestId("report-failure-back-to-workspace"));
    await waitFor(() => {
      expect(window.location.pathname + window.location.search).toBe("/workspace");
    });
  });

  it.each([
    ["first-load network failure", new Error("network unavailable")],
    ["not found", new Error("HTTP 404: REPORT_NOT_FOUND")],
  ])("falls back to workspace after %s with no trusted response", async (_label, failure) => {
    const { client } = clientFor(report());
    vi.spyOn(client, "getFeedbackReport").mockRejectedValue(failure);
    window.history.replaceState(null, "", `/report?reportId=${REPORT_ID}`);

    render(<App client={client} />);

    fireEvent.click(await screen.findByTestId("report-failure-back-to-workspace"));
    await waitFor(() => {
      expect(window.location.pathname + window.location.search).toBe("/workspace");
    });
  });

  it("renders four peer frozen context items with canonical resume and interview-record URLs", async () => {
    localStorage.setItem("ei-lang", "en");
    const { client, getTargetJob, getResume } = clientFor(report());
    render(<App client={client} initialRoute={{
      name: "report",
      params: {
        reportId: REPORT_ID,
        sessionId: "route-session-must-be-ignored",
        targetJobId: "route-target-must-be-ignored",
        resumeId: "route-resume-must-be-ignored",
        roundName: "route-round-must-be-ignored",
        reportStatus: "failed",
      },
    }} />);

    expect(await screen.findByTestId("report-dashboard")).toBeInTheDocument();
    expect(screen.getByTestId("report-header-title")).toHaveTextContent("星环科技 · 很长的高级前端工程师岗位名称");
    expect(screen.queryByTestId("report-context-session")).not.toBeInTheDocument();
    const contextStrip = screen.getByTestId("report-context-strip");
    expect(contextStrip.children).toHaveLength(4);
    expect(screen.getByTestId("report-context-resume-link")).toHaveAttribute(
      "href",
      "/resume-versions?resumeId=01918fa0-0000-7000-8000-000000001000",
    );
    expect(screen.getByTestId("report-context-strip")).toContainElement(
      screen.getByTestId("report-context-conversation-action"),
    );
    expect(screen.queryByTestId("report-conversation-entry")).not.toBeInTheDocument();
    const dashboard = screen.getByTestId("report-dashboard");
    const dashboardAttributes = [dashboard, ...dashboard.querySelectorAll("*")]
      .flatMap((element) =>
        Array.from(element.attributes, ({ name, value }) => `${name}=${value}`),
      )
      .join("\n");
    for (const sentinel of [REPORT_ID, SESSION_ID]) {
      expect(dashboard).not.toHaveTextContent(sentinel);
      expect(dashboardAttributes).not.toContain(sentinel);
    }
    expect(screen.getByTestId("report-context-round")).toHaveTextContent("技术一面 · 系统设计与工程取舍");
    expect(screen.getByText("模型原文：你的案例有清楚证据，但技术取舍需要补充。")).toBeInTheDocument();
    const actionLabel = screen.getByText("模型原文：补齐一个取舍案例后进入下一轮。");
    expect(actionLabel).toBeInTheDocument();
    expect(actionLabel).toHaveClass("ei-report-action-label");
    expect(actionLabel).toHaveStyle({
      minWidth: 0,
      overflowWrap: "anywhere",
      wordBreak: "normal",
    });
    expect(screen.getByTestId("report-dimensions")).toHaveTextContent("Strong · High confidence");
    expect(screen.getByTestId("report-dimensions")).toHaveTextContent("Needs work · Medium confidence");
    expect(document.body.textContent).not.toMatch(/answer_structure|technical_tradeoffs|needs_work|\bhigh\b|\bmedium\b/);
    expect(getTargetJob).not.toHaveBeenCalled();
    expect(getResume).not.toHaveBeenCalled();
  });

  it("renders the ready report in 4/2/2/2/1 order with one bottom full-width interview summary", async () => {
    localStorage.setItem("ei-lang", "zh");
    const value = report();
    const { client } = clientFor(value);
    render(<App client={client} initialRoute={{ name: "report", params: { reportId: REPORT_ID } }} />);

    const dashboard = await screen.findByTestId("report-dashboard");
    expect(screen.getByTestId("report-context-strip").children).toHaveLength(4);
    expect(screen.getByTestId("report-summary-cards").children).toHaveLength(2);
    expect(screen.getByTestId("report-detail-grid").children).toHaveLength(5);

    const overall = screen.getByTestId("report-overall-summary");
    expect(overall).toHaveStyle({ gridColumn: "1 / -1" });
    expect(overall).toHaveTextContent("面试总评");
    expect(overall).toHaveTextContent("建议再练");
    expect(overall).toHaveTextContent(value.summary ?? "");
    expect(dashboard.querySelectorAll('[data-testid="report-overall-summary"]')).toHaveLength(1);
    expect(dashboard.textContent?.split(value.summary ?? "")).toHaveLength(2);

    const readyGroups = Array.from(dashboard.querySelectorAll(
      '[data-testid="report-context-strip"], [data-testid="report-summary-cards"], [data-testid="report-detail-grid"], [data-testid="report-overall-summary"]',
    ));
    expect(readyGroups).toEqual([
      screen.getByTestId("report-context-strip"),
      screen.getByTestId("report-summary-cards"),
      screen.getByTestId("report-detail-grid"),
      overall,
    ]);
  });

  it("uses the first action only for CTA visual priority and keeps an empty replay focus valid", async () => {
    const value = report({
      nextActions: [{ type: "retry_current_round", label: "通用同轮复练。" }],
      retryFocusDimensionCodes: [],
    });
    const { client } = clientFor(value);
    render(<App client={client} initialRoute={{ name: "report", params: { reportId: REPORT_ID } }} />);

    expect(await screen.findByTestId("report-replay-cta")).toHaveAttribute("data-variant", "accent");
    expect(screen.getByTestId("report-next-cta")).toHaveAttribute("data-variant", "secondary");
    expect(screen.getByTestId("report-replay-cta")).toBeEnabled();
  });

  it("keeps long English dimension labels readable when the status wraps on mobile", async () => {
    localStorage.setItem("ei-lang", "en");
    const value = report({
      dimensionAssessments: [
        { code: "answer_structure", label: "System design", status: "strong", confidence: "high" },
        { code: "technical_tradeoffs", label: "Technical leadership", status: "needs_work", confidence: "medium" },
      ],
    });
    const { client } = clientFor(value);
    render(<App client={client} initialRoute={{ name: "report", params: { reportId: REPORT_ID } }} />);

    const matchingLabels = await screen.findAllByText("Technical leadership");
    const label = matchingLabels.find((element) => element.tagName === "SPAN");
    expect(label).toBeDefined();
    if (!label) throw new Error("dimension label span is missing");
    expect(label).toHaveClass("ei-report-dimension-label");
    expect(label).toHaveStyle({
      flex: "1 1 160px",
      overflowWrap: "break-word",
      wordBreak: "normal",
    });
    const row = label.parentElement;
    expect(row).not.toBeNull();
    expect(row).toHaveClass("ei-report-dimension-row");
    expect(row).toHaveStyle({ flexWrap: "wrap", gap: "8px 16px" });
    const status = row?.querySelector(".ei-report-dimension-status");
    expect(status).not.toBeNull();
    expect(status).toHaveStyle({
      flex: "0 1 auto",
      maxWidth: "100%",
      overflowWrap: "break-word",
      wordBreak: "normal",
    });
  });

  it("fails closed when non-empty replay focus is not backed by a needs-work dimension and same-code issue", async () => {
    const value = report({ retryFocusDimensionCodes: ["answer_structure"] });
    const { client } = clientFor(value);
    render(<App client={client} initialRoute={{ name: "report", params: { reportId: REPORT_ID } }} />);

    expect(await screen.findByTestId("report-failure-state")).toHaveAttribute("data-contract-invalid", "true");
    expect(screen.queryByTestId("report-failure-retry-cta")).not.toBeInTheDocument();
    expect(screen.queryByTestId("report-dashboard")).not.toBeInTheDocument();
  });

  it("renders the typed contract failure instead of dereferencing malformed nested API data", async () => {
    const value = report({ context: null as unknown as FeedbackReport["context"] });
    const { client } = clientFor(value);
    render(<App client={client} initialRoute={{ name: "report", params: { reportId: REPORT_ID } }} />);

    expect(await screen.findByTestId("report-failure-state")).toHaveAttribute(
      "data-contract-invalid",
      "true",
    );
    expect(screen.queryByTestId("report-dashboard")).not.toBeInTheDocument();
  });

  it("fails closed without exposing a malformed 201-code-point action label", async () => {
    const malformed = `private-${"x".repeat(193)}`;
    const value = report({
      nextActions: [{ type: "retry_current_round", label: malformed }],
    });
    const { client } = clientFor(value);
    render(<App client={client} initialRoute={{ name: "report", params: { reportId: REPORT_ID } }} />);

    expect(await screen.findByTestId("report-failure-state")).toHaveAttribute(
      "data-contract-invalid",
      "true",
    );
    expect(screen.queryByTestId("report-dashboard")).not.toBeInTheDocument();
    expect(document.body).not.toHaveTextContent(malformed);
  });

  it.each([
    {
      label: "25-word English action",
      language: "en" as const,
      malformed: Array.from({ length: 25 }, (_, index) => `private${index + 1}`).join(" "),
    },
    {
      label: "65-code-point Chinese action",
      language: "zh-CN" as const,
      malformed: "私".repeat(65),
    },
  ])("fails closed without exposing a $label", async ({ language, malformed }) => {
    const base = report();
    const value = report({
      context: { ...base.context, language },
      provenance: { ...base.provenance!, language },
      nextActions: [{ type: "retry_current_round", label: malformed }],
    });
    const { client } = clientFor(value);
    render(<App client={client} initialRoute={{ name: "report", params: { reportId: REPORT_ID } }} />);

    expect(await screen.findByTestId("report-failure-state")).toHaveAttribute(
      "data-contract-invalid",
      "true",
    );
    expect(screen.queryByTestId("report-dashboard")).not.toBeInTheDocument();
    expect(document.body).not.toHaveTextContent(malformed);
  });

  it("keeps a legal unbroken English action token recoverable for wrapping", async () => {
    const base = report();
    const unbroken = "ArchitectureTradeoffEvidence".repeat(3);
    const value = report({
      context: { ...base.context, language: "en" },
      provenance: { ...base.provenance!, language: "en" },
      nextActions: [{ type: "retry_current_round", label: unbroken }],
    });
    const { client } = clientFor(value);
    render(<App client={client} initialRoute={{ name: "report", params: { reportId: REPORT_ID } }} />);

    const label = await screen.findByText(unbroken);
    expect(label).toHaveClass("ei-report-action-label");
    expect(label).toHaveStyle({
      minWidth: 0,
      overflowWrap: "anywhere",
      wordBreak: "normal",
    });
    expect(label.parentElement).toHaveClass("ei-report-action-row");
  });
});
