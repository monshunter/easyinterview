/** @vitest-environment jsdom */

import { useLayoutEffect, type FC } from "react";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { EasyInterviewClient } from "../../../../api/generated/client";
import type { FeedbackReport } from "../../../../api/generated/types";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import { AppRuntimeContext } from "../../../runtime/AppRuntimeProvider";
import { ReportDashboard } from "../components/ReportDashboard";

const REPORT_ID = "01918fa0-0000-7000-8000-000000007000";

function readyReport(): FeedbackReport {
  return {
    id: REPORT_ID,
    sessionId: "01918fa0-0000-7000-8000-000000005000",
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    status: "ready",
    errorCode: null,
    summary: "Grounded summary.",
    preparednessLevel: "needs_practice",
    context: {
      sourcePlanId: "01918fa0-0000-7000-8000-000000004000",
      targetJobTitle: "Platform Engineer",
      targetJobCompany: "Acme",
      resumeId: "01918fa0-0000-7000-8000-000000001000",
      resumeDisplayName: "Platform resume",
      roundId: "round-1-technical",
      roundSequence: 1,
      roundName: "Technical interview",
      roundType: "technical",
      language: "en",
      hasNextRound: true,
    },
    dimensionAssessments: [
      {
        code: "technical_depth",
        label: "Technical depth",
        status: "needs_work",
        confidence: "medium",
      },
    ],
    highlights: [],
    issues: [
      {
        dimensionCode: "technical_depth",
        evidence: "The answer needs a measurable result.",
        confidence: "medium",
      },
    ],
    nextActions: [
      {
        type: "retry_current_round",
        label: "Practice this round with a measurable result",
      },
    ],
    retryFocusDimensionCodes: ["technical_depth"],
    provenance: {
      promptVersion: "v0.2.0",
      rubricVersion: "v0.2.0",
      modelId: "fixture",
      language: "en",
      featureFlag: "none",
      dataSourceVersion: "fixture.v1",
    },
    createdAt: "2026-07-14T08:00:00Z",
    updatedAt: "2026-07-14T08:01:00Z",
  };
}

const SwitchCommitProbe: FC<{
  active: boolean;
  observations: boolean[];
}> = ({ active, observations }) => {
  useLayoutEffect(() => {
    if (!active) return;
    const staleBack = screen.queryByTestId("report-back-button");
    observations.push(staleBack !== null);
    staleBack?.click();
  }, [active, observations]);
  return null;
};

const Harness: FC<{
  client: EasyInterviewClient;
  inspectSwitch: boolean;
  navigate: ReturnType<typeof vi.fn>;
  observations: boolean[];
}> = ({ client, inspectSwitch, navigate, observations }) => (
  <AppRuntimeContext.Provider
    value={{
      client,
      runtime: { status: "ready", config: {} as never },
      auth: { status: "unauthenticated" },
      refreshAuth: () => undefined,
    }}
  >
    <NavigationProvider value={{ navigate }}>
      <ReportDashboard reportId={REPORT_ID} />
      <SwitchCommitProbe active={inspectSwitch} observations={observations} />
    </NavigationProvider>
  </AppRuntimeContext.Provider>
);

describe("ReportDashboard client identity isolation", () => {
  it("does not expose the previous client's report or Back target on the switch commit", async () => {
    const firstClient = {
      getFeedbackReport: vi.fn(async () => readyReport()),
    } as unknown as EasyInterviewClient;
    const secondClient = {
      getFeedbackReport: vi.fn(() => new Promise<FeedbackReport>(() => undefined)),
    } as unknown as EasyInterviewClient;
    const navigate = vi.fn();
    const observations: boolean[] = [];
    const { rerender } = render(
      <Harness
        client={firstClient}
        inspectSwitch={false}
        navigate={navigate}
        observations={observations}
      />,
    );

    expect(await screen.findByTestId("report-dashboard")).toBeInTheDocument();
    rerender(
      <Harness
        client={secondClient}
        inspectSwitch
        navigate={navigate}
        observations={observations}
      />,
    );

    expect(observations).toEqual([false]);
    expect(navigate).not.toHaveBeenCalled();
    expect(screen.queryByTestId("report-back-button")).not.toBeInTheDocument();
    expect(screen.getByTestId("report-dashboard-loading")).toBeInTheDocument();
  });
});
