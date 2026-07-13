/** @vitest-environment jsdom */
import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { EasyInterviewClient } from "../../../../api/generated/client";
import type { FeedbackReport } from "../../../../api/generated/types";
import { App } from "../../../App";

const REPORT_ID = "01918fa0-0000-7000-8000-000000007000";

function report(overrides: Partial<FeedbackReport> = {}): FeedbackReport {
  return {
    id: REPORT_ID,
    sessionId: "01918fa0-0000-7000-8000-000000005000",
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
          emailMasked: "t***@example.com",
          profileCompletionRequired: false,
        } as never;
      },
      async getFeedbackReport() { return value; },
      getTargetJob,
      getResume,
    } as unknown as EasyInterviewClient,
    getTargetJob,
    getResume,
  };
}

afterEach(() => localStorage.removeItem("ei-lang"));

describe("grounded direct-semantic feedback report", () => {
  it("uses frozen API context, localizes enums, and preserves model prose under a different UI locale", async () => {
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
    expect(screen.getByTestId("report-context-session")).toHaveTextContent("01918fa0-0000-7000-8000-000000005000");
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
