/** @vitest-environment jsdom */

import { cleanup, fireEvent, render, screen, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { FeedbackReport } from "../../../../api/generated/types";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import { GeneratingScreen } from "../GeneratingScreen";

const pollMock = vi.hoisted(() => ({
  current: {} as Record<string, unknown>,
}));

vi.mock("../hooks/useReportGenerationPoll", () => ({
  useReportGenerationPoll: () => pollMock.current,
}));

const REPORT_ID = "01918fa0-0000-7000-8000-000000007000";
const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";

function trustedReport(status: "queued" | "generating" | "failed"): FeedbackReport {
  return {
    id: REPORT_ID,
    sessionId: "01918fa0-0000-7000-8000-000000005000",
    targetJobId: TARGET_JOB_ID,
    status,
    errorCode: status === "failed" ? "AI_PROVIDER_TIMEOUT" : null,
    summary: null,
    preparednessLevel: null,
    context: {
      sourcePlanId: "01918fa0-0000-7000-8000-000000004000",
      targetJobTitle: "Senior Engineer",
      targetJobCompany: "Acme",
      resumeId: "01918fa0-0000-7000-8000-000000001000",
      resumeDisplayName: "Resume",
      roundId: "round-2-technical",
      roundSequence: 2,
      roundName: "Technical",
      roundType: "technical",
      language: "en",
      hasNextRound: true,
    },
    dimensionAssessments: [],
    highlights: [],
    issues: [],
    nextActions: [],
    retryFocusDimensionCodes: [],
    provenance: null,
    createdAt: "2026-07-14T00:00:00Z",
    updatedAt: "2026-07-14T00:00:01Z",
  };
}

function renderBack(
  state: "timeout" | "error" | "failed" | "invalid" | "ready",
  report: FeedbackReport | null,
  reportId = REPORT_ID,
) {
  const navigate = vi.fn();
  pollMock.current = {
    state,
    attemptCount: 2,
    report,
    errorCode: state === "failed" ? report?.errorCode ?? "REPORT_NOT_FOUND" : null,
    retry: vi.fn(),
  };
  render(
    <NavigationProvider value={{ navigate }}>
      <GeneratingScreen route={{ name: "generating", params: reportId ? { reportId } : {} }} />
    </NavigationProvider>,
  );
  fireEvent.click(screen.getByTestId("generating-error-back-to-workspace"));
  return navigate;
}

afterEach(() => cleanup());

describe("Generating Back destination", () => {
  it.each([
    ["timeout with the last queued response", "timeout", trustedReport("queued")],
    ["network exhaustion with the last generating response", "error", trustedReport("generating")],
    ["terminal failure with its current response", "failed", trustedReport("failed")],
    ["invalid terminal with the last generating response", "invalid", trustedReport("generating")],
  ] as const)("uses the trusted target after %s", (_label, state, report) => {
    const navigate = renderBack(state, report);

    expect(screen.getByTestId("generating-error-back-to-workspace")).toHaveTextContent("Back to interview reports");
    expect(navigate).toHaveBeenCalledWith({
      name: "reports",
      params: { targetJobId: TARGET_JOB_ID },
    });
  });

  it("keeps the trusted reports return visible while generation is in progress", () => {
    const navigate = vi.fn();
    pollMock.current = {
      state: "polling",
      attemptCount: 1,
      report: trustedReport("generating"),
      errorCode: null,
      retry: vi.fn(),
    };
    render(
      <NavigationProvider value={{ navigate }}>
        <GeneratingScreen
          route={{ name: "generating", params: { reportId: REPORT_ID } }}
        />
      </NavigationProvider>,
    );

    fireEvent.click(
      within(screen.getByTestId("generating-back-button")).getByRole("button"),
    );
    expect(
      within(screen.getByTestId("generating-back-button")).getByRole("button"),
    ).toHaveTextContent("Back to interview reports");
    expect(navigate).toHaveBeenCalledWith({
      name: "reports",
      params: { targetJobId: TARGET_JOB_ID },
    });
  });

  it.each([
    ["first-load network exhaustion", "error", null],
    ["404 without a response", "failed", null],
    [
      "invalid response target",
      "invalid",
      { ...trustedReport("generating"), targetJobId: "route-target-must-not-be-trusted" },
    ],
  ] as const)("falls back to workspace after %s", (_label, state, report) => {
    const navigate = renderBack(state, report as FeedbackReport | null);

    expect(screen.getByTestId("generating-error-back-to-workspace")).toHaveTextContent("Back");
    expect(screen.getByTestId("generating-error-back-to-workspace")).not.toHaveTextContent("Back to interview reports");
    expect(navigate).toHaveBeenCalledWith({ name: "workspace", params: {} });
  });

  it("falls back to workspace when reportId is missing even if stale report data exists", () => {
    const navigate = renderBack("error", trustedReport("generating"), "");

    expect(navigate).toHaveBeenCalledWith({ name: "workspace", params: {} });
  });
});
