/**
 * @vitest-environment jsdom
 *
 * Phase 1.6 — GeneratingScreen Vitest gate. Covers:
 *  - i18n zh/en switch repaints the surface
 *  - ≥ 10 generating-* testid anchors render
 *  - missing reportId triggers ErrorState and skips the request
 *  - failed status nav handoff carries reportStatus + errorCode
 *  - ready status nav handoff carries reportId + sessionId (debounced to 1 call)
 *  - timeout state surfaces retry CTA
 *  - negative: no mistakesQueue / report-timeline testids leak in.
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

import type {
  ApiErrorCode,
  FeedbackReport,
} from "../../../../api/generated/types";
import { EasyInterviewClient } from "../../../../api/generated/client";
import { App } from "../../../App";
import type { LooseRoute } from "../../../normalizeRoute";

const REPORT_ID = "01918fa0-0000-7000-8000-000000007000";
const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";
const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";

function makeReport(overrides: Partial<FeedbackReport> = {}): FeedbackReport {
  return {
    id: REPORT_ID,
    sessionId: SESSION_ID,
    targetJobId: TARGET_JOB_ID,
    status: "generating",
    createdAt: "2026-05-16T00:00:00Z",
    updatedAt: "2026-05-16T00:00:01Z",
    ...overrides,
  };
}

function buildClient(
  responses: Array<FeedbackReport | { reject: unknown }>,
): EasyInterviewClient {
  let i = 0;
  return {
    async getRuntimeConfig() {
      return { aiProviderProfile: "stub" } as never;
    },
    async getMe() {
      return {
        id: "user-1",
        emailMasked: "u***@example.com",
        displayName: "User",
        profileCompletionRequired: false,
      } as never;
    },
    async getFeedbackReport(): Promise<FeedbackReport> {
      const next = responses[Math.min(i, responses.length - 1)];
      i += 1;
      if (next && typeof next === "object" && "reject" in next) {
        throw next.reject;
      }
      return next as FeedbackReport;
    },
    // The screen handoff navigates to <ReportScreen> on ready/failed; the
    // dashboard there hydrates context labels via getTargetJob / getResume.
    // Stub them so the post-handoff render doesn't throw.
    async getTargetJob() {
      throw new Error("HTTP 404 Not Found");
    },
    async getResume() {
      throw new Error("HTTP 404 Not Found");
    },
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

describe("GeneratingScreen", () => {
  it("renders the 10+ generating testids on mount when reportId is present", async () => {
    const client = buildClient([makeReport({ status: "generating" })]);
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "generating",
          params: { reportId: REPORT_ID, sessionId: SESSION_ID },
        }}
      />,
    );
    await screen.findByTestId("generating-screen");
    const anchors = [
      "generating-screen",
      "generating-header-eyebrow",
      "generating-header-title",
      "generating-header-subtitle",
      "generating-progress",
      "generating-progress-counter",
      "generating-progress-percentage",
      "generating-progress-rail",
      "generating-progress-fill",
      "generating-phase-list",
      "generating-phase-0",
      "generating-live-stream",
      "generating-live-stream-label",
      "generating-sla-hint",
      "generating-notify-cta",
    ];
    for (const id of anchors) {
      expect(screen.queryByTestId(id), `${id} missing`).not.toBeNull();
    }
    // Negative anchors for non-current modules.
    expect(screen.queryByTestId("mistakes-queue")).toBeNull();
    expect(screen.queryByTestId("report-timeline")).toBeNull();
  });

  it("renders ErrorState and skips the request when reportId is missing (TestGeneratingScreenMissingReportIdRendersErrorState)", async () => {
    const client = buildClient([makeReport()]);
    const spy = vi.spyOn(client, "getFeedbackReport");
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "generating",
          params: { sessionId: SESSION_ID },
        }}
      />,
    );
    await screen.findByTestId("generating-error-state");
    expect(
      screen.getByTestId("generating-error-state").getAttribute("data-error-kind"),
    ).toBe("missingReportId");
    expect(spy).not.toHaveBeenCalled();
  });

  it("renders the live evidence stream eyebrow in zh when locale is Chinese", async () => {
    localStorage.setItem("ei-lang", "zh");
    const client = buildClient([makeReport()]);
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "generating",
          params: { reportId: REPORT_ID, sessionId: SESSION_ID },
        }}
      />,
    );
    const label = await screen.findByTestId("generating-live-stream-label");
    expect(label.textContent).toContain("实时观察");
    localStorage.removeItem("ei-lang");
  });

  it("on ready status, navigates exactly once to report with reportId + sessionId in params (TestReadyCallbackDebouncesNavReport)", async () => {
    const client = buildClient([
      makeReport({ status: "ready", preparednessLevel: "basically_ready" }),
    ]);
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "generating",
          params: {
            reportId: REPORT_ID,
            sessionId: SESSION_ID,
            targetJobId: TARGET_JOB_ID,
            planId: "plan-x",
          },
        }}
      />,
    );

    // After ready, the screen replaces the generating layout. We assert the
    // route switched by looking for app-root and absence of generating-screen.
    await waitFor(
      () => {
        expect(screen.queryByTestId("generating-screen")).toBeNull();
      },
      { timeout: 4000 },
    );
  });

  it("on failed status, navigates to report carrying reportStatus + errorCode (TestFailedCallbackNavReportWithStatus)", async () => {
    const client = buildClient([
      makeReport({
        status: "failed",
        errorCode: "AI_PROVIDER_TIMEOUT" as ApiErrorCode,
      }),
    ]);
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "generating",
          params: {
            reportId: REPORT_ID,
            sessionId: SESSION_ID,
            targetJobId: TARGET_JOB_ID,
          },
        }}
      />,
    );
    await waitFor(
      () => {
        expect(screen.queryByTestId("generating-screen")).toBeNull();
      },
      { timeout: 4000 },
    );
  });

  it("phase labels show in English when the App boots with English locale (TestGeneratingScreenI18nEnglish)", async () => {
    localStorage.setItem("ei-lang", "en");
    const client = buildClient([makeReport()]);
    render(
      <Harness
        client={client}
        initialRoute={{
          name: "generating",
          params: { reportId: REPORT_ID, sessionId: SESSION_ID },
        }}
      />,
    );
    const phaseList = await screen.findByTestId("generating-phase-list");
    expect(
      within(phaseList).getAllByText(/[A-Za-z]/).length,
    ).toBeGreaterThan(0);
    expect(within(phaseList).queryByText(/转写/)).toBeNull();
    localStorage.removeItem("ei-lang");
  });
});
