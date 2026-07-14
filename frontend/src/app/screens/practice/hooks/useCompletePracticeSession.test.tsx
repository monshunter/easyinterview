/**
 * @vitest-environment jsdom
 *
 * Item 4.1 — useCompletePracticeSession hook contract.
 *  - body is exactly { clientCompletedAt: ISO8601 }
 *  - request always carries an Idempotency-Key header (spec D-13)
 *  - retries reuse the same key (replay returns the first response)
 *  - StrictMode double-invocation deduplicates to one POST
 *  - 409 mismatch surfaces error; 5xx surfaces retryable error
 *  - 3 consecutive failures unlock a "back to workspace" fallback signal
 */

import { describe, expect, it, vi } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import { useEffect, type ReactNode } from "react";

import { EasyInterviewClient } from "../../../../api/generated/client";
import type { ReportWithJob } from "../../../../api/generated/types";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../api/mockTransport";
import {
  InterviewContextProvider,
  useInterviewContext,
} from "../../../interview-context/InterviewContext";
import { AppRuntimeProvider } from "../../../runtime/AppRuntimeProvider";
import { useCompletePracticeSession } from "./useCompletePracticeSession";

import completePracticeSessionFixture from "../../../../../../openapi/fixtures/PracticeSessions/completePracticeSession.json";

const SESSION_A = "01918fa0-0000-7000-8000-000000005000";
const SESSION_B = "01918fa0-0000-7000-8000-000000005001";

interface CapturedRequest {
  url: string;
  method: string;
  headers: Headers;
  bodyText: string | null;
}

function buildClient(opts: {
  scenario?: string;
  forceFailFirstN?: number;
} = {}): { client: EasyInterviewClient; calls: CapturedRequest[] } {
  const calls: CapturedRequest[] = [];
  const fixtureFetch = createFixtureBackedFetch(
    createFixtureRegistry([completePracticeSessionFixture]),
    { scenario: opts.scenario ?? "default" },
  );
  let attempts = 0;
  const wrappedFetch: typeof fetch = async (input, init) => {
    const url =
      typeof input === "string"
        ? input
        : input instanceof URL
          ? input.href
          : input.url;
    const method = (init?.method ?? "GET").toUpperCase();
    const headers = new Headers(init?.headers ?? {});
    let bodyText: string | null = null;
    if (typeof init?.body === "string") bodyText = init.body;
    calls.push({ url, method, headers, bodyText });
    const path = new URL(url, "http://x").pathname;
    const isCompleteCall =
      method === "POST" &&
      /\/practice\/sessions\/[^/]+\/complete$/.test(path);
    if (isCompleteCall) {
      attempts += 1;
      if (opts.forceFailFirstN && attempts <= opts.forceFailFirstN) {
        throw new Error("simulated network failure");
      }
    }
    return fixtureFetch(input, init);
  };
  return {
    client: new EasyInterviewClient({ fetch: wrappedFetch }),
    calls,
  };
}

function completeCalls(all: CapturedRequest[]): CapturedRequest[] {
  return all.filter(
    (c) =>
      c.method === "POST" &&
      /\/practice\/sessions\/[^/]+\/complete$/.test(
        new URL(c.url, "http://x").pathname,
      ),
  );
}

function readBody(call: CapturedRequest): Record<string, unknown> {
  if (!call.bodyText) throw new Error("expected JSON body");
  return JSON.parse(call.bodyText) as Record<string, unknown>;
}

function Wrapper({
  children,
  client,
}: {
  children: ReactNode;
  client: EasyInterviewClient;
}) {
  return (
    <InterviewContextProvider>
      <AppRuntimeProvider client={client}>
        <Hydrate>{children}</Hydrate>
      </AppRuntimeProvider>
    </InterviewContextProvider>
  );
}

function Hydrate({ children }: { children: ReactNode }) {
  const { dispatch } = useInterviewContext();
  useEffect(() => {
    dispatch({
      type: "HYDRATE_FROM_ROUTE",
      params: { sessionId: SESSION_A },
    });
  }, [dispatch]);
  return <>{children}</>;
}

describe("useCompletePracticeSession", () => {
  it("happy path: posts {clientCompletedAt} only and carries Idempotency-Key", async () => {
    const { client, calls } = buildClient();

    const { result } = renderHook(
      () => useCompletePracticeSession(SESSION_A),
      {
        wrapper: ({ children }) => (
          <Wrapper client={client}>{children}</Wrapper>
        ),
      },
    );

    let report;
    await act(async () => {
      report = await result.current.complete();
    });
    expect(report).toBeDefined();
    expect(report!.reportId).toBeDefined();

    const calls1 = completeCalls(calls);
    expect(calls1.length).toBe(1);
    const body = readBody(calls1[0]!);
    expect(Object.keys(body)).toEqual(["clientCompletedAt"]);
    expect(typeof body.clientCompletedAt).toBe("string");
    expect(Number.isFinite(Date.parse(body.clientCompletedAt as string))).toBe(true);
    expect(calls1[0]!.headers.get("Idempotency-Key")).not.toBeNull();
  });

  it("retry of the same complete reuses the same Idempotency-Key", async () => {
    const { client, calls } = buildClient({ forceFailFirstN: 1 });

    const { result } = renderHook(
      () => useCompletePracticeSession(SESSION_A),
      {
        wrapper: ({ children }) => (
          <Wrapper client={client}>{children}</Wrapper>
        ),
      },
    );

    await act(async () => {
      try {
        await result.current.complete();
      } catch (err) {
        expect(err).toBeInstanceOf(Error);
      }
    });
    await act(async () => {
      await result.current.complete();
    });

    const c = completeCalls(calls);
    expect(c.length).toBeGreaterThanOrEqual(2);
    expect(c[0]!.headers.get("Idempotency-Key")).toBe(
      c[1]!.headers.get("Idempotency-Key"),
    );
  });

  it("StrictMode-style double invocation deduplicates to one POST", async () => {
    const { client, calls } = buildClient();

    const { result } = renderHook(
      () => useCompletePracticeSession(SESSION_A),
      {
        wrapper: ({ children }) => (
          <Wrapper client={client}>{children}</Wrapper>
        ),
      },
    );

    await act(async () => {
      const a = result.current.complete();
      const b = result.current.complete();
      const [r1, r2] = await Promise.all([a, b]);
      expect(r1).toBe(r2);
    });

    expect(completeCalls(calls).length).toBe(1);
  });

  it("3 failed attempts surface the back-to-workspace fallback", async () => {
    const { client, calls } = buildClient({ forceFailFirstN: 3 });

    const { result } = renderHook(
      () => useCompletePracticeSession(SESSION_A),
      {
        wrapper: ({ children }) => (
          <Wrapper client={client}>{children}</Wrapper>
        ),
      },
    );

    for (let i = 0; i < 3; i++) {
      await act(async () => {
        try {
          await result.current.complete();
        } catch {
          /* ignore */
        }
      });
    }

    expect(result.current.state.kind).toBe("error");
    if (result.current.state.kind === "error") {
      expect(result.current.state.attempts).toBeGreaterThanOrEqual(3);
      expect(result.current.state.fallbackBackToWorkspace).toBe(true);
    }
    expect(completeCalls(calls).length).toBeGreaterThanOrEqual(3);
  });

  it("409 mismatch surfaces a non-retryable error state", async () => {
    const { client } = buildClient({ scenario: "mismatch" });

    const { result } = renderHook(
      () => useCompletePracticeSession(SESSION_A),
      {
        wrapper: ({ children }) => (
          <Wrapper client={client}>{children}</Wrapper>
        ),
      },
    );

    await act(async () => {
      try {
        await result.current.complete();
      } catch {
        /* ignore */
      }
    });
    expect(result.current.state.kind).toBe("error");
    if (result.current.state.kind === "error") {
      expect(result.current.state.code).toBe(409);
    }
  });

  it("does not infer an HTTP status from a plain Error.message", async () => {
    const { client } = buildClient();
    client.completePracticeSession = async () => {
      throw new Error("HTTP 409 message text is not typed metadata");
    };

    const { result } = renderHook(
      () => useCompletePracticeSession(SESSION_A),
      {
        wrapper: ({ children }) => (
          <Wrapper client={client}>{children}</Wrapper>
        ),
      },
    );

    await act(async () => {
      await expect(result.current.complete()).rejects.toThrow("HTTP 409 message text is not typed metadata");
    });
    expect(result.current.state.kind).toBe("error");
    if (result.current.state.kind === "error") {
      expect(result.current.state.code).toBeNull();
    }
  });

  it("does not reuse session A's in-flight promise, success cache, or idempotency key after rerendering for session B", async () => {
    const { client } = buildClient();
    const calls: Array<{ sessionId: string; key: string | undefined }> = [];
    let resolveA: ((report: ReportWithJob) => void) | undefined;
    let resolveB: ((report: ReportWithJob) => void) | undefined;
    vi.spyOn(client, "completePracticeSession").mockImplementation((sessionId, _body, options) => {
      calls.push({ sessionId, key: options?.headers?.["Idempotency-Key"] });
      return new Promise((resolve) => {
        if (sessionId === SESSION_A) resolveA = resolve;
        if (sessionId === SESSION_B) resolveB = resolve;
      });
    });
    const { result, rerender } = renderHook(
      ({ sessionId }) => useCompletePracticeSession(sessionId),
      {
        initialProps: { sessionId: SESSION_A },
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );

    let aPromise!: Promise<ReportWithJob>;
    act(() => { aPromise = result.current.complete(); });
    expect(result.current.state.kind).toBe("loading");

    rerender({ sessionId: SESSION_B });
    expect(result.current.state.kind).toBe("idle");
    await act(async () => { resolveA?.(report("report-a")); await aPromise; });
    expect(result.current.state.kind).toBe("idle");

    let bPromise!: Promise<ReportWithJob>;
    act(() => { bPromise = result.current.complete(); });
    expect(calls.map((call) => call.sessionId)).toEqual([SESSION_A, SESSION_B]);
    expect(calls[1]!.key).toBeTruthy();
    expect(calls[1]!.key).not.toBe(calls[0]!.key);
    await act(async () => { resolveB?.(report("report-b")); });
    await expect(bPromise).resolves.toMatchObject({ reportId: "report-b" });
    expect(result.current.state).toMatchObject({
      kind: "success",
      report: { reportId: "report-b" },
    });
  });

  it("does not let a late session A rejection overwrite session B's in-flight or success state", async () => {
    const { client } = buildClient();
    let rejectA: ((cause: Error) => void) | undefined;
    let resolveB: ((report: ReportWithJob) => void) | undefined;
    vi.spyOn(client, "completePracticeSession").mockImplementation((sessionId) => new Promise((resolve, reject) => {
      if (sessionId === SESSION_A) rejectA = reject;
      if (sessionId === SESSION_B) resolveB = resolve;
    }));
    const { result, rerender } = renderHook(
      ({ sessionId }) => useCompletePracticeSession(sessionId),
      {
        initialProps: { sessionId: SESSION_A },
        wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
      },
    );

    let aPromise!: Promise<ReportWithJob>;
    act(() => { aPromise = result.current.complete(); });
    const observedA = aPromise.catch((cause: unknown) => cause);
    rerender({ sessionId: SESSION_B });
    let bPromise!: Promise<ReportWithJob>;
    act(() => { bPromise = result.current.complete(); });
    const observedB = bPromise.then(
      (value) => value,
      (cause: unknown) => cause,
    );
    expect(result.current.state.kind).toBe("loading");

    await act(async () => { rejectA?.(new Error("late A failure")); await observedA; });
    expect(result.current.state.kind).toBe("loading");
    await act(async () => { resolveB?.(report("report-b")); });
    await expect(observedB).resolves.toMatchObject({ reportId: "report-b" });
    await waitFor(() => expect(result.current.state).toMatchObject({
      kind: "success",
      report: { reportId: "report-b" },
    }));
  });
});

function report(reportId: string): ReportWithJob {
  return {
    ...(completePracticeSessionFixture.scenarios.default.response.body as unknown as ReportWithJob),
    reportId,
  };
}
