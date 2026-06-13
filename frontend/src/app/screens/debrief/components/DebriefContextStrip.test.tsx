// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";

import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import type {
  PracticeSession,
  Resume,
  TargetJob,
} from "../../../../api/generated/types";
import { DebriefContextStrip } from "./DebriefContextStrip";
import { EMPTY_SELECTED_CONTEXT } from "../types";

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

const PRACTICE_SESSION: PracticeSession = {
  id: "01918fa0-0000-7000-8000-000000005000",
  planId: "01918fa0-0000-7000-8000-000000004000",
  targetJobId: TARGET_JOB.id,
  status: "completed",
  language: "zh-CN",
  hintsEnabled: true,
  turnCount: 8,
  currentTurn: null,
  createdAt: "2026-05-15T13:00:00Z",
  updatedAt: "2026-05-15T14:12:00Z",
};

const RESUME: Resume = {
  id: "01918fa0-0000-7000-8000-000000001000",
  title: "Alice — Senior Frontend Engineer",
  displayName: "Alice — Senior Frontend Engineer",
  language: "zh-CN",
  parseStatus: "ready",
  status: "active",
  sourceType: "upload",
  fileObjectId: "01918fa0-0000-7000-8000-000000001100",
  originalText: null,
  parsedTextSnapshot: null,
  parsedSummary: null,
  createdAt: "2026-04-22T09:30:00Z",
  updatedAt: "2026-05-12T08:00:00Z",
  deletedAt: null,
};

function setup(selected = EMPTY_SELECTED_CONTEXT) {
  const onOpenPicker = vi.fn();
  render(
    <DisplayPreferencesProvider initial={{ lang: "zh" }}>
      <DebriefContextStrip
        selectedContext={selected}
        onOpenPicker={onOpenPicker}
      />
    </DisplayPreferencesProvider>,
  );
  return { onOpenPicker };
}

describe("DebriefContextStrip — TestContextStrip_OpenPicker", () => {
  it("invokes onOpenPicker with the card kind when the action button is clicked", () => {
    const { onOpenPicker } = setup();
    fireEvent.click(screen.getByTestId("debrief-context-card-targetJob-open"));
    fireEvent.click(
      screen.getByTestId("debrief-context-card-mockSession-open"),
    );
    fireEvent.click(screen.getByTestId("debrief-context-card-resume-open"));
    expect(onOpenPicker).toHaveBeenNthCalledWith(1, "targetJob");
    expect(onOpenPicker).toHaveBeenNthCalledWith(2, "mockSession");
    expect(onOpenPicker).toHaveBeenNthCalledWith(3, "resume");
  });
});

describe("DebriefContextStrip — TestContextStrip_DisplayNameFetch", () => {
  it("renders the selected target job / mock session / resume titles when context is populated", () => {
    setup({
      targetJob: TARGET_JOB,
      mockSession: PRACTICE_SESSION,
      resume: RESUME,
    });
    const tj = screen.getByTestId("debrief-context-card-targetJob-title");
    expect(tj).toHaveTextContent("Acme · Senior Frontend Engineer");
    const ms = screen.getByTestId("debrief-context-card-mockSession-title");
    expect(ms).toHaveTextContent(PRACTICE_SESSION.id);
    const rs = screen.getByTestId("debrief-context-card-resume-title");
    expect(rs).toHaveTextContent(RESUME.displayName);
  });
});

describe("DebriefContextStrip — TestContextStrip_FallbackOnAPIError", () => {
  it("falls back to the unset copy when no selection is present", () => {
    setup();
    expect(
      screen.getByTestId("debrief-context-card-targetJob-title"),
    ).toHaveTextContent("未选择");
    expect(
      screen.getByTestId("debrief-context-card-mockSession-title"),
    ).toHaveTextContent("未选择");
    expect(
      screen.getByTestId("debrief-context-card-resume-title"),
    ).toHaveTextContent("未选择");
  });
});
