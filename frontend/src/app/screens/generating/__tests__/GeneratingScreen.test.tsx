/** @vitest-environment jsdom */
import { render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { EasyInterviewClient } from "../../../../api/generated/client";
import type { FeedbackReport } from "../../../../api/generated/types";
import { App } from "../../../App";

const REPORT_ID = "01918fa0-0000-7000-8000-000000007000";
const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";

function makeReport(overrides: Partial<FeedbackReport> = {}): FeedbackReport {
  return {
    id: REPORT_ID,
    sessionId: SESSION_ID,
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    status: "generating",
    errorCode: null,
    summary: null,
    preparednessLevel: null,
    context: {
      sourcePlanId: "01918fa0-0000-7000-8000-000000004000",
      targetJobTitle: "高级前端工程师",
      targetJobCompany: "星环科技",
      resumeId: "01918fa0-0000-7000-8000-000000001000",
      resumeDisplayName: "前端工程师简历",
      roundId: "round-2-technical",
      roundSequence: 2,
      roundName: "技术一面 · 45m",
      roundType: "technical",
      language: "zh-CN",
      hasNextRound: true,
    },
    dimensionAssessments: [],
    highlights: [],
    issues: [],
    nextActions: [],
    retryFocusDimensionCodes: [],
    provenance: null,
    createdAt: "2026-07-12T08:00:00Z",
    updatedAt: "2026-07-12T08:01:00Z",
    ...overrides,
  };
}

function readyReport(overrides: Partial<FeedbackReport> = {}): FeedbackReport {
  return makeReport({
    status: "ready",
    summary: "模型原文：证据清楚，但技术取舍仍需补强。",
    preparednessLevel: "needs_practice",
    dimensionAssessments: [
      { code: "technical_tradeoffs", label: "技术取舍", status: "needs_work", confidence: "high" },
    ],
    issues: [
      { dimensionCode: "technical_tradeoffs", evidence: "没有比较替代方案。", confidence: "high" },
    ],
    nextActions: [
      { type: "retry_current_round", label: "先补充替代方案比较。" },
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
    ...overrides,
  });
}

function buildClient(responses: FeedbackReport[]) {
  let index = 0;
  const getTargetJob = vi.fn(async () => { throw new Error("must not read mutable target"); });
  const getResume = vi.fn(async () => { throw new Error("must not read mutable resume"); });
  const client = {
    async getRuntimeConfig() { return { aiProviderProfile: "stub" } as never; },
    async getMe() {
      return {
        id: "user-1",
        emailMasked: "u***@example.com",
        displayName: "User",
        profileCompletionRequired: false,
      } as never;
    },
    async getFeedbackReport() {
      const value = responses[Math.min(index, responses.length - 1)]!;
      index += 1;
      return value;
    },
    getTargetJob,
    getResume,
  } as unknown as EasyInterviewClient;
  return { client, getTargetJob, getResume };
}

afterEach(() => localStorage.removeItem("ei-lang"));

describe("GeneratingScreen honest state projection", () => {
  it("shows only the real generating status without fake progress, observations, notify, or records promises", async () => {
    localStorage.setItem("ei-lang", "zh");
    const { client } = buildClient([makeReport({ status: "generating" })]);
    render(<App client={client} initialRoute={{ name: "generating", params: { reportId: REPORT_ID } }} />);

    await waitFor(() => expect(screen.getByTestId("generating-screen")).toHaveAttribute("data-report-status", "generating"));
    expect(screen.getByTestId("generating-header-title")).toHaveTextContent("正在生成证据化报告");
    expect(screen.queryByTestId("generating-progress")).not.toBeInTheDocument();
    expect(screen.queryByTestId("generating-phase-list")).not.toBeInTheDocument();
    expect(screen.queryByTestId("generating-live-stream")).not.toBeInTheDocument();
    expect(screen.queryByTestId("generating-notify-cta")).not.toBeInTheDocument();
    expect(document.body.textContent).not.toMatch(/实时观察|好了通知我|会话记录|\d+%/);
  });

  it("navigates ready reports with reportId only and renders frozen API context without mutable label fetches", async () => {
    localStorage.setItem("ei-lang", "en");
    const { client, getTargetJob, getResume } = buildClient([readyReport()]);
    render(<App client={client} initialRoute={{
      name: "generating",
      params: {
        reportId: REPORT_ID,
        sessionId: "route-session-must-drop",
        targetJobId: "route-target-must-drop",
        reportStatus: "failed",
      },
    }} />);

    expect(await screen.findByTestId("report-dashboard")).toBeInTheDocument();
    expect(screen.getByTestId("report-context-job")).toHaveTextContent("星环科技 · 高级前端工程师");
    expect(screen.getByText("模型原文：证据清楚，但技术取舍仍需补强。")).toBeInTheDocument();
    expect(getTargetJob).not.toHaveBeenCalled();
    expect(getResume).not.toHaveBeenCalled();
  });

  it("renders REPORT_CONTEXT_TOO_LARGE as a back-only terminal state with actionable shorter-input copy", async () => {
    localStorage.setItem("ei-lang", "zh");
    const { client } = buildClient([makeReport({
      status: "failed",
      errorCode: "REPORT_CONTEXT_TOO_LARGE",
    })]);
    render(<App client={client} initialRoute={{ name: "generating", params: { reportId: REPORT_ID } }} />);

    const surface = await screen.findByTestId("generating-screen");
    expect(surface).toHaveStyle({ minHeight: "100vh" });
    const state = screen.getByTestId("generating-error-state");
    expect(state).toHaveAttribute("data-error-kind", "contextTooLarge");
    expect(screen.getByTestId("generating-header-eyebrow")).toHaveTextContent("报告不可用");
    expect(screen.getByTestId("generating-header-title")).toHaveTextContent("本次材料与对话过长");
    expect(screen.getByTestId("generating-error-desc")).toHaveTextContent("缩短输入");
    expect(screen.queryByTestId("generating-error-retry")).not.toBeInTheDocument();
    expect(screen.getByTestId("generating-error-back-to-workspace")).toBeInTheDocument();
  });

  it("keeps backend failed and not-found reports terminal instead of re-GET-as-regenerate", async () => {
    const { client } = buildClient([makeReport({ status: "failed", errorCode: "AI_OUTPUT_INVALID" })]);
    render(<App client={client} initialRoute={{ name: "generating", params: { reportId: REPORT_ID } }} />);

    await waitFor(() => expect(screen.getByTestId("generating-error-state")).toHaveAttribute("data-error-kind", "invalidReport"));
    expect(screen.getByTestId("generating-screen")).toBeInTheDocument();
    expect(screen.queryByTestId("generating-error-retry")).not.toBeInTheDocument();
  });

  it("does not request a report when reportId is missing", async () => {
    const { client } = buildClient([makeReport()]);
    const request = vi.spyOn(client, "getFeedbackReport");
    render(<App client={client} initialRoute={{ name: "generating", params: {} }} />);

    expect(await screen.findByTestId("generating-error-state")).toHaveAttribute("data-error-kind", "missingReportId");
    expect(request).not.toHaveBeenCalled();
  });
});
