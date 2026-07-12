/** @vitest-environment jsdom */
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { EasyInterviewClient } from "../../../../api/generated/client";
import { App } from "../../../App";

describe("conversation-level feedback report", () => {
  it("renders dimensions and evidence without question review structures", async () => {
    const client = {
      async getRuntimeConfig() { return { aiProviderProfile: "stub" }; },
      async getMe() { return { id: "user-1", displayName: "Tester", emailMasked: "t***@example.com", profileCompletionRequired: false }; },
      async getFeedbackReport() { return {
        id: "01918fa0-0000-7000-8000-000000007000", sessionId: "01918fa0-0000-7000-8000-000000005000", targetJobId: "01918fa0-0000-7000-8000-000000002000", status: "ready",
        preparednessLevel: "basically_ready", dimensionAssessments: [{ dimension: "technical_depth", status: "needs_work", confidence: "medium" }],
        highlights: [{ dimension: "ownership", evidence: "明确说明了跨团队推进路径。", confidence: "high" }], issues: [{ dimension: "technical_depth", evidence: "缺少灰度策略证据。", confidence: "medium" }],
        nextActions: [{ type: "retry_current_round", label: "复练灰度策略" }], retryFocusCompetencyCodes: ["technical_depth"], createdAt: "2026-07-12T08:30:00Z", updatedAt: "2026-07-12T08:31:00Z",
      }; },
      async getTargetJob() { return { id: "01918fa0-0000-7000-8000-000000002000", title: "Senior Engineer", companyName: "Acme", requirements: [] }; },
      async getResume() { return { id: "01918fa0-0000-7000-8000-000000004000", displayName: "Resume" }; },
    } as unknown as EasyInterviewClient;
    render(<App client={client} initialRoute={{ name: "report", params: {
      reportId: "01918fa0-0000-7000-8000-000000007000",
      sessionId: "01918fa0-0000-7000-8000-000000005000",
      targetJobId: "01918fa0-0000-7000-8000-000000002000",
      resumeId: "01918fa0-0000-7000-8000-000000004000",
    } }} />);

    expect(await screen.findByTestId("report-dashboard")).toBeInTheDocument();
    expect(screen.getByTestId("report-dimensions")).toHaveTextContent("technical_depth");
    expect(screen.getByTestId("report-highlights")).toHaveTextContent("跨团队推进路径");
    expect(screen.getByTestId("report-issues")).toHaveTextContent("灰度策略");
    expect(screen.queryByTestId("report-detail-tab-questions")).not.toBeInTheDocument();
    expect(screen.queryByText("题目回顾")).not.toBeInTheDocument();
  });
});
