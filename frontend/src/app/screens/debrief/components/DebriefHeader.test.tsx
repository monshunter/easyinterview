// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";

import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import type { TargetJob } from "../../../../api/generated/types";
import { DebriefHeader } from "./DebriefHeader";
import { EMPTY_SELECTED_CONTEXT } from "../types";

function withProviders(node: React.ReactNode) {
  return (
    <DisplayPreferencesProvider initial={{ lang: "zh" }}>
      {node}
    </DisplayPreferencesProvider>
  );
}

const TARGET_JOB: TargetJob = {
  id: "01918fa0-0000-7000-8000-000000002000",
  status: "interviewing",
  analysisStatus: "ready",
  title: "Senior Frontend Engineer",
  companyName: "Acme",
  locationText: "Shanghai",
  targetLanguage: "zh-CN",
  sourceType: "url",
  openQuestionIssueCount: 0,
  requirements: [],
  createdAt: "2026-05-15T13:00:00Z",
  updatedAt: "2026-05-15T14:12:00Z",
};

describe("DebriefHeader — TestDebriefHeader_RenderWithContext", () => {
  it("renders the eyebrow with company + title segments from the selected target job", () => {
    render(
      withProviders(
        <DebriefHeader
          selectedContext={{
            ...EMPTY_SELECTED_CONTEXT,
            targetJob: TARGET_JOB,
          }}
          onBack={() => undefined}
          capturedAtLabel="刚刚"
          interviewerLabel="张哲 · 技术负责人"
          modalityLabel="远程视频"
        />,
      ),
    );
    const eyebrow = screen.getByTestId("debrief-header-eyebrow");
    expect(eyebrow).toHaveTextContent("复盘");
    expect(eyebrow).toHaveTextContent("Acme");
    expect(eyebrow).toHaveTextContent("Senior Frontend Engineer");
    expect(screen.getByTestId("debrief-header-title")).toBeInTheDocument();
    expect(screen.getByTestId("debrief-header-subcopy")).toBeInTheDocument();
    const meta = screen.getByTestId("debrief-header-meta");
    expect(meta).toHaveTextContent("刚刚");
    expect(meta).toHaveTextContent("张哲");
    expect(meta).toHaveTextContent("远程视频");
  });

  it("invokes onBack when the back button is clicked", () => {
    const onBack = vi.fn();
    render(
      withProviders(
        <DebriefHeader
          selectedContext={EMPTY_SELECTED_CONTEXT}
          onBack={onBack}
        />,
      ),
    );
    fireEvent.click(screen.getByTestId("debrief-header-back"));
    expect(onBack).toHaveBeenCalledTimes(1);
  });
});

describe("DebriefHeader — TestDebriefHeader_FallbackOnMissingContext", () => {
  it("falls back to the unset eyebrow + meta copy when context is empty", () => {
    render(
      withProviders(
        <DebriefHeader
          selectedContext={EMPTY_SELECTED_CONTEXT}
          onBack={() => undefined}
        />,
      ),
    );
    const eyebrow = screen.getByTestId("debrief-header-eyebrow");
    expect(eyebrow).toHaveTextContent("待关联岗位");
    expect(eyebrow).toHaveTextContent("未指定面试轮次");
    const meta = screen.getByTestId("debrief-header-meta");
    // Three "未填写" placeholders, one per meta row.
    expect(meta.querySelectorAll("dd")).toHaveLength(3);
    expect(meta.textContent).toMatch(/未填写/);
  });
});
