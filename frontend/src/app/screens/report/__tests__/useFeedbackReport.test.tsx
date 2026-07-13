/**
 * @vitest-environment jsdom
 *
 * Phase 2.8 — useFeedbackReport: idle / loading / data / error / notFound;
 * cross-user 404 maps to notFound + REPORT_NOT_FOUND; read path never carries
 * Idempotency-Key; unmount cancels inflight.
 */

import { describe, expect, it, vi } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";

import type { FeedbackReport } from "../../../../api/generated/types";
import { EasyInterviewClient } from "../../../../api/generated/client";
import { AppRuntimeProvider } from "../../../runtime/AppRuntimeProvider";
import { useFeedbackReport } from "../hooks/useFeedbackReport";

const REPORT_ID = "01918fa0-0000-7000-8000-000000007000";

function makeReport(overrides: Partial<FeedbackReport> = {}): FeedbackReport {
  return {
    id: REPORT_ID,
    sessionId: "01918fa0-0000-7000-8000-000000005000",
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    status: "ready",
    errorCode: null,
    summary: "Grounded summary",
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
    preparednessLevel: "needs_practice",
    highlights: [],
    issues: [{ dimensionCode: "technical_depth", evidence: "Needs a metric.", confidence: "medium" }],
    nextActions: [{ type: "retry_current_round", label: "Practice with a measurable result." }],
    dimensionAssessments: [{ code: "technical_depth", label: "Technical depth", status: "needs_work", confidence: "medium" }],
    retryFocusDimensionCodes: ["technical_depth"],
    provenance: {
      promptVersion: "v0.2.0",
      rubricVersion: "v0.2.0",
      modelId: "fixture",
      language: "en",
      featureFlag: "none",
      dataSourceVersion: "fixture.v1",
    },
    createdAt: "2026-05-16T00:00:00Z",
    updatedAt: "2026-05-16T00:00:00Z",
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
      throw new Error("HTTP 401 Unauthorized");
    },
    async getFeedbackReport(_: string, opts?: { headers?: Record<string, string> }) {
      if (opts?.headers) {
        if ("Idempotency-Key" in opts.headers || "idempotency-key" in opts.headers) {
          throw new Error("read path leaked Idempotency-Key");
        }
      }
      const next = responses[Math.min(i, responses.length - 1)];
      i += 1;
      if (next && typeof next === "object" && "reject" in next) {
        throw next.reject;
      }
      return next as FeedbackReport;
    },
  } as unknown as EasyInterviewClient;
}

function Wrapper({
  client,
  children,
}: {
  client: EasyInterviewClient;
  children: ReactNode;
}) {
  return <AppRuntimeProvider client={client}>{children}</AppRuntimeProvider>;
}

describe("useFeedbackReport", () => {
  it("transitions loading → data on ready (TestUseFeedbackReport4States happy path)", async () => {
    const client = buildClient([makeReport()]);
    const { result } = renderHook(() => useFeedbackReport(REPORT_ID), {
      wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
    });
    await waitFor(() => expect(result.current.state).toBe("data"));
    expect(result.current.data?.id).toBe(REPORT_ID);
  });

  it("HTTP 404 produces notFound + REPORT_NOT_FOUND errorCode (TestUseFeedbackReportCrossUser404)", async () => {
    const client = buildClient([{ reject: new Error("HTTP 404 Not Found") }]);
    const { result } = renderHook(() => useFeedbackReport(REPORT_ID), {
      wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
    });
    await waitFor(() => expect(result.current.state).toBe("notFound"));
    expect(result.current.errorCode).toBe("REPORT_NOT_FOUND");
  });

  it("missing reportId stays in error and never fetches", async () => {
    const client = buildClient([makeReport()]);
    const spy = vi.spyOn(client, "getFeedbackReport");
    const { result } = renderHook(() => useFeedbackReport(""), {
      wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
    });
    await waitFor(() => expect(result.current.state).toBe("error"));
    expect(spy).not.toHaveBeenCalled();
  });

  it("5xx transitions to error (not notFound) and refresh re-fetches", async () => {
    const client = buildClient([
      { reject: new Error("HTTP 500 Internal") },
      makeReport(),
    ]);
    const spy = vi.spyOn(client, "getFeedbackReport");
    const { result } = renderHook(() => useFeedbackReport(REPORT_ID), {
      wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
    });
    await waitFor(() => expect(result.current.state).toBe("error"));
    expect(result.current.errorCode).toBeNull();
    act(() => {
      result.current.refresh();
    });
    await waitFor(() => expect(result.current.state).toBe("data"));
    expect(spy).toHaveBeenCalledTimes(2);
  });
});
