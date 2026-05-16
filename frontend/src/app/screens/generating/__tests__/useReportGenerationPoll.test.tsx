/**
 * @vitest-environment jsdom
 *
 * Phase 1.7 — useReportGenerationPoll: 7-state machine, exponential backoff
 * (1.5s × 1.5 capped at 8s, max attempts 30), visibility/focus pause-resume,
 * onReady / onFailed callbacks, unmount-cancel, no Idempotency-Key, cross-user
 * 404 → REPORT_NOT_FOUND failure path.
 */

import { describe, expect, it, vi } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";

import type {
  ApiErrorCode,
  FeedbackReport,
} from "../../../../api/generated/types";
import { EasyInterviewClient } from "../../../../api/generated/client";
import { AppRuntimeProvider } from "../../../runtime/AppRuntimeProvider";
import {
  useReportGenerationPoll,
  type PollScheduler,
} from "../hooks/useReportGenerationPoll";

const REPORT_ID = "01918fa0-0000-7000-8000-000000007000";

function makeReport(overrides: Partial<FeedbackReport>): FeedbackReport {
  return {
    id: REPORT_ID,
    sessionId: "01918fa0-0000-7000-8000-000000005000",
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    status: "generating",
    createdAt: "2026-05-16T00:00:00Z",
    updatedAt: "2026-05-16T00:00:01Z",
    ...overrides,
  };
}

function buildClientWithSequence(
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
    async getFeedbackReport(): Promise<FeedbackReport> {
      const next = responses[Math.min(i, responses.length - 1)];
      i += 1;
      if (next && typeof next === "object" && "reject" in next) {
        throw next.reject;
      }
      return next as FeedbackReport;
    },
  } as unknown as EasyInterviewClient;
}

function buildClientStuck(): EasyInterviewClient {
  return {
    async getRuntimeConfig() {
      return { aiProviderProfile: "stub" } as never;
    },
    async getMe() {
      throw new Error("HTTP 401 Unauthorized");
    },
    async getFeedbackReport(): Promise<FeedbackReport> {
      return makeReport({ status: "generating" });
    },
  } as unknown as EasyInterviewClient;
}

interface ManualScheduler extends PollScheduler {
  pending: Array<{ ms: number; cb: () => void; cancelled: boolean }>;
  flushNext: () => void;
  flushAll: () => void;
}

function manualScheduler(): ManualScheduler {
  const pending: ManualScheduler["pending"] = [];
  return {
    pending,
    schedule(ms, cb) {
      const entry = { ms, cb, cancelled: false };
      pending.push(entry);
      return () => {
        entry.cancelled = true;
      };
    },
    flushNext() {
      while (pending.length > 0) {
        const next = pending.shift()!;
        if (!next.cancelled) {
          next.cb();
          return;
        }
      }
    },
    flushAll() {
      let safety = 200;
      while (pending.length > 0 && safety-- > 0) {
        const next = pending.shift()!;
        if (!next.cancelled) next.cb();
      }
    },
  };
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

describe("useReportGenerationPoll", () => {
  it("renders error state immediately when reportId is missing (TestGeneratingScreenMissingReportIdRendersErrorState)", () => {
    const client = buildClientStuck();
    const spy = vi.spyOn(client, "getFeedbackReport");
    const { result } = renderHook(
      () => useReportGenerationPoll({ reportId: "" }),
      {
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );
    expect(result.current.state).toBe("error");
    expect(spy).not.toHaveBeenCalled();
  });

  it("transitions polling → ready and fires onReady (TestOnReadyCallbackNavReport)", async () => {
    const client = buildClientWithSequence([
      makeReport({ status: "generating" }),
      makeReport({ status: "generating" }),
      makeReport({ status: "ready", preparednessLevel: "basically_ready" }),
    ]);
    const onReady = vi.fn();
    const sched = manualScheduler();

    const { result } = renderHook(
      () =>
        useReportGenerationPoll({
          reportId: REPORT_ID,
          onReady,
          scheduler: sched,
        }),
      {
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );

    await waitFor(() => expect(result.current.attemptCount).toBe(1));
    // After the first attempt the hook schedules the next backoff; flush it.
    await act(async () => {
      sched.flushNext();
    });
    await waitFor(() => expect(result.current.attemptCount).toBe(2));
    await act(async () => {
      sched.flushNext();
    });
    await waitFor(() => expect(result.current.state).toBe("ready"));
    expect(onReady).toHaveBeenCalledTimes(1);
    expect(onReady.mock.calls[0]![0].status).toBe("ready");
  });

  it("transitions polling → failed and fires onFailed with errorCode (TestOnFailedCallbackNavReportWithReportStatus)", async () => {
    const client = buildClientWithSequence([
      makeReport({
        status: "failed",
        errorCode: "AI_PROVIDER_TIMEOUT" as ApiErrorCode,
      }),
    ]);
    const onFailed = vi.fn();
    const sched = manualScheduler();
    const { result } = renderHook(
      () =>
        useReportGenerationPoll({
          reportId: REPORT_ID,
          onFailed,
          scheduler: sched,
        }),
      {
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );
    await waitFor(() => expect(result.current.state).toBe("failed"));
    expect(result.current.errorCode).toBe("AI_PROVIDER_TIMEOUT");
    expect(onFailed).toHaveBeenCalledTimes(1);
    expect(onFailed.mock.calls[0]![0]).toBe("AI_PROVIDER_TIMEOUT");
  });

  it("HTTP 404 maps to failed + REPORT_NOT_FOUND (TestUseReportGenerationPollCrossUser404)", async () => {
    const client = buildClientWithSequence([
      { reject: new Error("HTTP 404 Not Found") },
    ]);
    const onFailed = vi.fn();
    const sched = manualScheduler();
    const { result } = renderHook(
      () =>
        useReportGenerationPoll({
          reportId: REPORT_ID,
          onFailed,
          scheduler: sched,
        }),
      {
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );
    await waitFor(() => expect(result.current.state).toBe("failed"));
    expect(result.current.errorCode).toBe("REPORT_NOT_FOUND");
    expect(onFailed).toHaveBeenCalledWith("REPORT_NOT_FOUND");
  });

  it("hits timeout after maxAttempts on persistently generating responses (TestGeneratingScreenTimeoutStateShowsRetryCta core)", async () => {
    const client = buildClientStuck();
    const sched = manualScheduler();
    const { result } = renderHook(
      () =>
        useReportGenerationPoll({
          reportId: REPORT_ID,
          scheduler: sched,
          maxAttempts: 3,
        }),
      {
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );
    await waitFor(() => expect(result.current.attemptCount).toBe(1));
    await act(async () => {
      sched.flushNext();
    });
    await waitFor(() => expect(result.current.attemptCount).toBe(2));
    await act(async () => {
      sched.flushNext();
    });
    await waitFor(() => expect(result.current.attemptCount).toBe(3));
    await waitFor(() => expect(result.current.state).toBe("timeout"));
  });

  it("uses exponential backoff capped at maxDelay (TestUseReportGenerationPollExponentialBackoff)", async () => {
    const client = buildClientStuck();
    const sched = manualScheduler();
    renderHook(
      () =>
        useReportGenerationPoll({
          reportId: REPORT_ID,
          scheduler: sched,
          initialDelayMs: 1500,
          backoffFactor: 1.5,
          maxDelayMs: 8000,
          maxAttempts: 30,
        }),
      {
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );
    // First attempt fires immediately; first scheduled delay is the backoff
    // from attempt 1 → attempt 2, which is initialDelayMs * 1.5^0 = 1500.
    await waitFor(() => expect(sched.pending.length).toBeGreaterThan(0));
    expect(sched.pending[0]!.ms).toBe(1500);
    await act(async () => {
      sched.flushNext();
    });
    await waitFor(() => expect(sched.pending.length).toBeGreaterThan(0));
    expect(sched.pending[0]!.ms).toBeCloseTo(2250); // 1500 * 1.5
    await act(async () => {
      sched.flushNext();
    });
    await waitFor(() => expect(sched.pending.length).toBeGreaterThan(0));
    expect(sched.pending[0]!.ms).toBeCloseTo(3375); // 1500 * 1.5^2
    // Run forward enough attempts to hit the maxDelay cap.
    for (let i = 0; i < 6; i += 1) {
      await act(async () => {
        sched.flushNext();
      });
      await waitFor(() => expect(sched.pending.length).toBeGreaterThan(0));
    }
    expect(sched.pending[0]!.ms).toBe(8000);
  });

  it("pauses on visibilitychange + blur and resumes on visible + focus (TestUseReportGenerationPollVisibilityPauseResume / FocusEvents)", async () => {
    const client = buildClientStuck();
    const sched = manualScheduler();
    const { result } = renderHook(
      () =>
        useReportGenerationPoll({
          reportId: REPORT_ID,
          scheduler: sched,
        }),
      {
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );
    await waitFor(() => expect(result.current.state).toBe("polling"));

    await act(async () => {
      Object.defineProperty(document, "visibilityState", {
        configurable: true,
        get: () => "hidden",
      });
      document.dispatchEvent(new Event("visibilitychange"));
    });
    await waitFor(() => expect(result.current.state).toBe("paused"));

    await act(async () => {
      Object.defineProperty(document, "visibilityState", {
        configurable: true,
        get: () => "visible",
      });
      document.dispatchEvent(new Event("visibilitychange"));
    });
    await waitFor(() => expect(result.current.state).toBe("polling"));

    await act(async () => {
      window.dispatchEvent(new Event("blur"));
    });
    await waitFor(() => expect(result.current.state).toBe("paused"));

    await act(async () => {
      window.dispatchEvent(new Event("focus"));
    });
    await waitFor(() => expect(result.current.state).toBe("polling"));
  });

  it("getFeedbackReport requests do not carry Idempotency-Key (TestUseReportGenerationPollNoIdempotencyHeader)", async () => {
    // The hook calls the generated client method without RequestOptions
    // carrying mutation headers. This is enforced by signature: getFeedbackReport
    // only accepts {signal, headers, query}. Assert via spy:
    const client = buildClientWithSequence([
      makeReport({ status: "ready", preparednessLevel: "basically_ready" }),
    ]);
    const spy = vi.spyOn(client, "getFeedbackReport");
    const sched = manualScheduler();
    renderHook(
      () =>
        useReportGenerationPoll({
          reportId: REPORT_ID,
          scheduler: sched,
        }),
      {
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );
    await waitFor(() => expect(spy).toHaveBeenCalled());
    const lastCallOpts = spy.mock.calls[spy.mock.calls.length - 1]![1] ?? {};
    const headers = (lastCallOpts as { headers?: Record<string, string> })
      .headers;
    if (headers) {
      expect(headers).not.toHaveProperty("Idempotency-Key");
      expect(headers).not.toHaveProperty("idempotency-key");
    }
  });

  it("unmount cancels inflight (TestUseReportGenerationPollUnmountCancels)", async () => {
    const client = buildClientStuck();
    const sched = manualScheduler();
    const { result, unmount } = renderHook(
      () =>
        useReportGenerationPoll({
          reportId: REPORT_ID,
          scheduler: sched,
        }),
      {
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );
    await waitFor(() => expect(result.current.state).toBe("polling"));
    unmount();
    // After unmount the scheduler MUST not keep firing. Manually flush to make
    // sure no act-on-unmounted occurs.
    await act(async () => {
      sched.flushAll();
    });
  });

  it("does not invoke onReady more than once even if downstream re-renders (TestReadyCallbackDebouncesNavReport)", async () => {
    const ready = makeReport({
      status: "ready",
      preparednessLevel: "basically_ready",
    });
    const client = buildClientWithSequence([ready, ready, ready]);
    const onReady = vi.fn();
    const sched = manualScheduler();
    renderHook(
      () =>
        useReportGenerationPoll({
          reportId: REPORT_ID,
          onReady,
          scheduler: sched,
        }),
      {
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );
    await waitFor(() => expect(onReady).toHaveBeenCalledTimes(1));
    // Even if scheduler is flushed nothing else should fire since state moved
    // to 'ready' which stops scheduling.
    await act(async () => {
      sched.flushAll();
    });
    expect(onReady).toHaveBeenCalledTimes(1);
  });
});
