/**
 * @vitest-environment jsdom
 *
 * Phase 1.7 — useReportGenerationPoll: 7-state machine, exponential backoff
 * (1.5s × 1.5 capped at 8s, max attempts 49), visibility/focus pause-resume,
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
import { AppRuntimeContext } from "../../../runtime/AppRuntimeProvider";
import {
  REPORT_GENERATION_POLL_MAX_ATTEMPTS,
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
    errorCode: null,
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
      language: "en-US",
      hasNextRound: true,
    },
    dimensionAssessments: [],
    highlights: [],
    issues: [],
    nextActions: [],
    retryFocusDimensionCodes: [],
    provenance: null,
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
    async getFeedbackReport(): Promise<FeedbackReport> {
      return makeReport({ status: "generating" });
    },
  } as unknown as EasyInterviewClient;
}

function buildAbortAwareStuckClient() {
  let activeRequests = 0;
  let maxActiveRequests = 0;

  const client = {
    getFeedbackReport(
      _reportId: string,
      options?: { signal?: AbortSignal },
    ): Promise<FeedbackReport> {
      activeRequests += 1;
      maxActiveRequests = Math.max(maxActiveRequests, activeRequests);

      return new Promise((_resolve, reject) => {
        const onAbort = () => {
          activeRequests -= 1;
          const error = new Error("aborted");
          error.name = "AbortError";
          reject(error);
        };
        options?.signal?.addEventListener("abort", onAbort, { once: true });
      });
    },
  } as unknown as EasyInterviewClient;

  return {
    client,
    getActiveRequests: () => activeRequests,
    getMaxActiveRequests: () => maxActiveRequests,
  };
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
  return (
    <AppRuntimeContext.Provider
      value={{
        client,
        runtime: { status: "ready", config: {} as never },
        auth: { status: "unauthenticated" },
        refreshAuth: () => {},
      }}
    >
      {children}
    </AppRuntimeContext.Provider>
  );
}

describe("useReportGenerationPoll", () => {
  it("keeps the default status-check window open across three provider retry backoffs", () => {
    expect(REPORT_GENERATION_POLL_MAX_ATTEMPTS).toBe(49);

    const delays = Array.from(
      { length: REPORT_GENERATION_POLL_MAX_ATTEMPTS - 1 },
      (_, index) => Math.min(8000, 1500 * Math.pow(1.5, index)),
    );
    const totalWindowMs = delays.reduce((sum, delay) => sum + delay, 0);

    expect(totalWindowMs).toBeGreaterThanOrEqual(6 * 60 * 1000);
    expect(totalWindowMs).toBeLessThan(6 * 60 * 1000 + 10_000);
  });

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

  it("surfaces an exhausted network check separately from a resource timeout and allows checking again", async () => {
    const ready = makeReport({ status: "ready", preparednessLevel: "basically_ready" });
    const client = buildClientWithSequence([
      { reject: new Error("network unavailable") },
      ready,
    ]);
    const onReady = vi.fn();
    const { result } = renderHook(
      () => useReportGenerationPoll({
        reportId: REPORT_ID,
        maxAttempts: 1,
        onReady,
      }),
      {
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );

    await waitFor(() => expect(result.current.state).toBe("error"));
    act(() => result.current.retry());
    await waitFor(() => expect(result.current.state).toBe("ready"));
    expect(onReady).toHaveBeenCalledWith(ready);
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

  it("resume reuses the scheduled next attempt instead of firing an immediate duplicate request", async () => {
    const client = buildClientStuck();
    const spy = vi.spyOn(client, "getFeedbackReport");
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
    await waitFor(() => expect(spy).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(sched.pending.length).toBeGreaterThan(0));

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
    expect(spy).toHaveBeenCalledTimes(1);

    await act(async () => {
      sched.flushNext();
    });
    await waitFor(() => expect(spy).toHaveBeenCalledTimes(2));
    expect(result.current.attemptCount).toBe(2);
  });

  it("pausing an in-flight request resumes at n+1 after the preserved delay", async () => {
    const tracked = buildAbortAwareStuckClient();
    const spy = vi.spyOn(tracked.client, "getFeedbackReport");
    const sched = manualScheduler();
    const { result, unmount } = renderHook(
      () =>
        useReportGenerationPoll({
          reportId: REPORT_ID,
          scheduler: sched,
          initialDelayMs: 1500,
        }),
      {
        wrapper: ({ children }) => (
          <Wrapper client={tracked.client}>{children}</Wrapper>
        ),
      },
    );

    await waitFor(() => expect(spy).toHaveBeenCalledTimes(1));
    expect(result.current.attemptCount).toBe(1);
    expect(tracked.getActiveRequests()).toBe(1);

    await act(async () => {
      Object.defineProperty(document, "visibilityState", {
        configurable: true,
        get: () => "hidden",
      });
      document.dispatchEvent(new Event("visibilitychange"));
      window.dispatchEvent(new Event("blur"));
      window.dispatchEvent(new Event("blur"));
    });
    await waitFor(() => expect(result.current.state).toBe("paused"));
    await waitFor(() => expect(tracked.getActiveRequests()).toBe(0));

    await act(async () => {
      Object.defineProperty(document, "visibilityState", {
        configurable: true,
        get: () => "visible",
      });
      document.dispatchEvent(new Event("visibilitychange"));
      window.dispatchEvent(new Event("focus"));
      window.dispatchEvent(new Event("focus"));
    });
    await waitFor(() => expect(result.current.state).toBe("polling"));

    expect(spy).toHaveBeenCalledTimes(1);
    expect(
      sched.pending.filter((entry) => !entry.cancelled).map((entry) => entry.ms),
    ).toEqual([1500]);

    await act(async () => {
      sched.flushNext();
    });
    await waitFor(() => expect(spy).toHaveBeenCalledTimes(2));
    expect(result.current.attemptCount).toBe(2);
    expect(tracked.getMaxActiveRequests()).toBe(1);

    unmount();
    await waitFor(() => expect(tracked.getActiveRequests()).toBe(0));
  });

  it("repeated pause-resume during a saved wait is idempotent and does not restart attempt one", async () => {
    const client = buildClientStuck();
    const spy = vi.spyOn(client, "getFeedbackReport");
    const sched = manualScheduler();
    const { result } = renderHook(
      () =>
        useReportGenerationPoll({
          reportId: REPORT_ID,
          scheduler: sched,
          initialDelayMs: 1500,
        }),
      {
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );

    await waitFor(() => expect(spy).toHaveBeenCalledTimes(1));
    await waitFor(() =>
      expect(sched.pending.some((entry) => !entry.cancelled)).toBe(true),
    );

    for (let cycle = 0; cycle < 3; cycle += 1) {
      await act(async () => {
        window.dispatchEvent(new Event("blur"));
        window.dispatchEvent(new Event("blur"));
      });
      await waitFor(() => expect(result.current.state).toBe("paused"));

      await act(async () => {
        window.dispatchEvent(new Event("focus"));
        window.dispatchEvent(new Event("focus"));
      });
      await waitFor(() => expect(result.current.state).toBe("polling"));
      expect(spy).toHaveBeenCalledTimes(1);
      expect(
        sched.pending.filter((entry) => !entry.cancelled).map((entry) => entry.ms),
      ).toEqual([1500]);
    }

    await act(async () => {
      sched.flushNext();
    });
    await waitFor(() => expect(spy).toHaveBeenCalledTimes(2));
    expect(result.current.attemptCount).toBe(2);
  });

  it("fences a scheduled callback immediately when blur pauses the run", async () => {
    const client = buildClientStuck();
    const spy = vi.spyOn(client, "getFeedbackReport");
    const sched = manualScheduler();
    const { result } = renderHook(
      () =>
        useReportGenerationPoll({
          reportId: REPORT_ID,
          scheduler: sched,
          initialDelayMs: 1500,
        }),
      {
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );

    await waitFor(() => expect(spy).toHaveBeenCalledTimes(1));
    await waitFor(() =>
      expect(sched.pending.some((entry) => !entry.cancelled)).toBe(true),
    );

    act(() => {
      window.dispatchEvent(new Event("blur"));
      sched.flushNext();
    });

    await waitFor(() => expect(result.current.state).toBe("paused"));
    expect(spy).toHaveBeenCalledTimes(1);

    await act(async () => {
      window.dispatchEvent(new Event("focus"));
    });
    await waitFor(() => expect(result.current.state).toBe("polling"));
    expect(spy).toHaveBeenCalledTimes(1);
    expect(
      sched.pending.filter((entry) => !entry.cancelled).map((entry) => entry.ms),
    ).toEqual([1500]);
  });

  it("never starts more than 49 requests in one default poll run", async () => {
    const client = buildClientStuck();
    const spy = vi.spyOn(client, "getFeedbackReport");
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

    await waitFor(() => expect(spy).toHaveBeenCalledTimes(1));
    for (
      let expectedCalls = 2;
      expectedCalls <= REPORT_GENERATION_POLL_MAX_ATTEMPTS;
      expectedCalls += 1
    ) {
      await waitFor(() =>
        expect(sched.pending.some((entry) => !entry.cancelled)).toBe(true),
      );
      await act(async () => {
        sched.flushNext();
      });
      await waitFor(() => expect(spy).toHaveBeenCalledTimes(expectedCalls));
    }

    await waitFor(() => expect(result.current.state).toBe("timeout"));
    expect(result.current.attemptCount).toBe(
      REPORT_GENERATION_POLL_MAX_ATTEMPTS,
    );
    expect(spy).toHaveBeenCalledTimes(REPORT_GENERATION_POLL_MAX_ATTEMPTS);

    await act(async () => {
      sched.flushAll();
    });
    expect(spy).toHaveBeenCalledTimes(REPORT_GENERATION_POLL_MAX_ATTEMPTS);
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
