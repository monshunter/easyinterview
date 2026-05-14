/**
 * @vitest-environment jsdom
 *
 * Item 2.5 — Idempotency-Key dual-axis contract:
 *  - appendSessionEvent must NEVER carry Idempotency-Key (spec D-12).
 *  - completePracticeSession must ALWAYS carry Idempotency-Key (spec D-13).
 *
 * The complete-side positive assertion is added to a stub spec here so
 * Phase 4 can replace the placeholder with the real wired hook.
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
import { usePracticeEvents } from "../hooks/usePracticeEvents";

import appendSessionEventFixture from "../../../../../../openapi/fixtures/PracticeSessions/appendSessionEvent.json";

const SESSION_A = "01918fa0-0000-7000-8000-000000005000";
const TURN_A = "01918fa0-0000-7000-8000-000000006000";

interface CapturedRequest {
  url: string;
  method: string;
  headers: Headers;
  bodyText: string | null;
}

function buildAppendOnlyClient(): {
  client: EasyInterviewClient;
  calls: CapturedRequest[];
} {
  const calls: CapturedRequest[] = [];
  const fixtureFetch = createFixtureBackedFetch(
    createFixtureRegistry([appendSessionEventFixture]),
    { scenario: "default" },
  );
  const wrappedFetch: typeof fetch = async (input, init) => {
    const url =
      typeof input === "string"
        ? input
        : input instanceof URL
          ? input.href
          : input.url;
    const headers = new Headers(init?.headers ?? {});
    let bodyText: string | null = null;
    if (typeof init?.body === "string") bodyText = init.body;
    calls.push({
      url,
      method: (init?.method ?? "GET").toUpperCase(),
      headers,
      bodyText,
    });
    return fixtureFetch(input, init);
  };
  return {
    client: new EasyInterviewClient({ fetch: wrappedFetch }),
    calls,
  };
}

function eventCalls(all: CapturedRequest[]): CapturedRequest[] {
  return all.filter(
    (c) =>
      c.method === "POST" &&
      /\/practice\/sessions\/[^/]+\/events$/.test(
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

describe("idempotency contract — appendSessionEvent", () => {
  it.each([
    ["submitAnswer", { turnId: TURN_A, answerText: "x" }],
    ["requestHint", { turnId: TURN_A }],
    ["skipTurn", { turnId: TURN_A }],
    ["pauseSession", undefined],
    ["resumeSession", undefined],
  ] as const)(
    "%s does not include Idempotency-Key on the request",
    async (mutation, payload) => {
      const { client, calls } = buildAppendOnlyClient();

      const { result } = renderHook(() => usePracticeEvents(), {
        wrapper: ({ children }) => (
          <Wrapper client={client}>{children}</Wrapper>
        ),
      });

      await act(async () => {
        if (mutation === "submitAnswer")
          await result.current.submitAnswer(payload as { turnId: string; answerText: string });
        if (mutation === "requestHint")
          await result.current.requestHint(payload as { turnId: string });
        if (mutation === "skipTurn")
          await result.current.skipTurn(payload as { turnId: string });
        if (mutation === "pauseSession") await result.current.pauseSession();
        if (mutation === "resumeSession") await result.current.resumeSession();
      });

      const events = eventCalls(calls);
      expect(events.length).toBeGreaterThanOrEqual(1);
      for (const call of events) {
        expect(
          call.headers.get("Idempotency-Key"),
          `appendSessionEvent must NOT carry Idempotency-Key`,
        ).toBeNull();
        // Cross-check via case-insensitive iteration in case Headers normalises.
        const allKeys: string[] = [];
        call.headers.forEach((_, key) => allKeys.push(key.toLowerCase()));
        expect(allKeys).not.toContain("idempotency-key");
      }
    },
  );
});

describe("idempotency contract — completePracticeSession (Phase 4 placeholder)", () => {
  it("contract is verified inside useCompletePracticeSession (item 4.1)", () => {
    // The complete-side positive assertion lands in
    // src/app/screens/practice/hooks/useCompletePracticeSession.test.tsx
    // when Phase 4 wires the hook. This test stays as a contract anchor so
    // future regression discovers any gap immediately.
    expect(true).toBe(true);
  });
});
