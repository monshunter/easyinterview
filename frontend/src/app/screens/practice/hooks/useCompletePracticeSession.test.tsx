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

import { describe, expect, it } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { useEffect, type ReactNode } from "react";

import { EasyInterviewClient } from "../../../../api/generated/client";
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
});
