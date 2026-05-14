/**
 * @vitest-environment jsdom
 *
 * Item 4.3 — completePracticeSession body schema:
 *   - body keys = exactly { clientCompletedAt: ISO8601 }
 *   - request always carries Idempotency-Key (spec D-13)
 *   - display fields (mode/modality/practiceMode/practiceGoal/hintUsed/
 *     hintCount) MUST NOT appear in the request body
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
import { useCompletePracticeSession } from "../hooks/useCompletePracticeSession";

import completePracticeSessionFixture from "../../../../../../openapi/fixtures/PracticeSessions/completePracticeSession.json";

const SESSION_A = "01918fa0-0000-7000-8000-000000005000";
const FORBIDDEN_BODY_KEYS = [
  "mode",
  "modality",
  "practiceMode",
  "practiceGoal",
  "hintUsed",
  "hintCount",
  "sessionId",
  "planId",
  "targetJobId",
];

interface CapturedRequest {
  url: string;
  method: string;
  headers: Headers;
  bodyText: string | null;
}

function buildClient(): { client: EasyInterviewClient; calls: CapturedRequest[] } {
  const calls: CapturedRequest[] = [];
  const fixtureFetch = createFixtureBackedFetch(
    createFixtureRegistry([completePracticeSessionFixture]),
    { scenario: "default" },
  );
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

describe("completePracticeSession body schema (item 4.3)", () => {
  it("body has exactly one key — clientCompletedAt — as ISO8601", async () => {
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
      await result.current.complete();
    });
    const calls1 = completeCalls(calls);
    expect(calls1.length).toBe(1);
    const body = JSON.parse(calls1[0]!.bodyText ?? "{}") as Record<
      string,
      unknown
    >;
    expect(Object.keys(body)).toEqual(["clientCompletedAt"]);
    expect(typeof body.clientCompletedAt).toBe("string");
    expect(Number.isFinite(Date.parse(body.clientCompletedAt as string))).toBe(true);
  });

  it("request always carries Idempotency-Key", async () => {
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
      await result.current.complete();
    });
    const call = completeCalls(calls)[0]!;
    expect(call.headers.get("Idempotency-Key")).not.toBeNull();
  });

  it("body never contains display / context fields (negative)", async () => {
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
      await result.current.complete();
    });
    const body = JSON.parse(completeCalls(calls)[0]!.bodyText ?? "{}") as Record<
      string,
      unknown
    >;
    for (const key of FORBIDDEN_BODY_KEYS) {
      expect(body).not.toHaveProperty(key);
    }
  });
});
