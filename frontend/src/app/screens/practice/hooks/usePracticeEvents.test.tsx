/**
 * @vitest-environment jsdom
 *
 * Item 2.1 — usePracticeEvents: 4 mutations (submitAnswer / requestHint /
 * pauseSession / resumeSession), UUIDv7 clientEventId, NO
 * Idempotency-Key header on appendSessionEvent (spec D-12), retry of the
 * same user action reuses clientEventId, fresh action mints a new id.
 */

import { describe, expect, it } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import { useEffect, type ReactNode } from "react";

import { EasyInterviewClient } from "../../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../api/mockTransport";
import { UUID_V7_REGEX } from "../../../../lib/ids";
import {
  InterviewContextProvider,
  useInterviewContext,
} from "../../../interview-context/InterviewContext";
import { AppRuntimeProvider } from "../../../runtime/AppRuntimeProvider";
import { usePracticeEvents } from "./usePracticeEvents";

import appendSessionEventFixture from "../../../../../../openapi/fixtures/PracticeSessions/appendSessionEvent.json";

const SESSION_A = "01918fa0-0000-7000-8000-000000005000";
const TURN_A = "01918fa0-0000-7000-8000-000000006000";

interface CapturedRequest {
  url: string;
  method: string;
  headers: Headers;
  bodyText: string | null;
}

function buildClientWithCapture(
  scenario: string = "default",
  forceFailFirstAppend: boolean = false,
): { client: EasyInterviewClient; calls: CapturedRequest[] } {
  const calls: CapturedRequest[] = [];
  const fixtureFetch = createFixtureBackedFetch(
    createFixtureRegistry([appendSessionEventFixture]),
    { scenario },
  );
  let appendAttempts = 0;
  const wrappedFetch: typeof fetch = async (input, init) => {
    const url =
      typeof input === "string"
        ? input
        : input instanceof URL
          ? input.href
          : input.url;
    const headers = new Headers(init?.headers ?? {});
    let bodyText: string | null = null;
    const body = init?.body;
    if (typeof body === "string") bodyText = body;
    calls.push({
      url,
      method: (init?.method ?? "GET").toUpperCase(),
      headers,
      bodyText,
    });
    const isAppendCall =
      (init?.method ?? "GET").toUpperCase() === "POST" &&
      /\/practice\/sessions\/[^/]+\/events$/.test(
        new URL(url, "http://x").pathname,
      );
    if (isAppendCall) {
      appendAttempts += 1;
      if (forceFailFirstAppend && appendAttempts === 1) {
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

interface WrapperProps {
  children: ReactNode;
  client: EasyInterviewClient;
}

function Wrapper({ children, client }: WrapperProps) {
  return (
    <InterviewContextProvider>
      <AppRuntimeProvider client={client}>
        <HydrateContext>{children}</HydrateContext>
      </AppRuntimeProvider>
    </InterviewContextProvider>
  );
}

function HydrateContext({ children }: { children: ReactNode }) {
  const { dispatch } = useInterviewContext();
  useEffect(() => {
    dispatch({
      type: "HYDRATE_FROM_ROUTE",
      params: { sessionId: SESSION_A },
    });
  }, [dispatch]);
  return <>{children}</>;
}

function readBody(call: CapturedRequest): Record<string, unknown> {
  if (!call.bodyText) throw new Error("expected JSON body");
  return JSON.parse(call.bodyText) as Record<string, unknown>;
}

function eventCalls(all: CapturedRequest[]): CapturedRequest[] {
  return all.filter(
    (c) => c.method === "POST" && /\/practice\/sessions\/[^/]+\/events$/.test(new URL(c.url, "http://x").pathname),
  );
}

describe("usePracticeEvents", () => {
  it("submitAnswer posts answer_submitted with clientEventId UUIDv7 and no Idempotency-Key", async () => {
    const { client, calls } = buildClientWithCapture();

    const { result } = renderHook(() => usePracticeEvents(), {
      wrapper: ({ children }) => (
        <Wrapper client={client}>{children}</Wrapper>
      ),
    });

    await act(async () => {
      await result.current.submitAnswer({ turnId: TURN_A, answerText: "hi" });
    });

    const last = calls.at(-1);
    expect(last).toBeDefined();
    expect(last!.method).toBe("POST");
    expect(last!.url).toContain(`/practice/sessions/${SESSION_A}/events`);
    expect(last!.headers.get("Idempotency-Key")).toBeNull();
    const body = readBody(last!);
    expect(body.kind).toBe("answer_submitted");
    expect(body.payload).toEqual({ turnId: TURN_A, answerText: "hi" });
    expect(typeof body.clientEventId).toBe("string");
    expect(UUID_V7_REGEX.test(body.clientEventId as string)).toBe(true);
    expect(typeof body.occurredAt).toBe("string");
  });

  it.each([
    ["requestHint", "hint_requested", { turnId: TURN_A }],
    ["pauseSession", "session_paused", {}],
    ["resumeSession", "session_resumed", {}],
  ] as const)(
    "%s posts %s with typed payload and no Idempotency-Key",
    async (mutation, expectedKind, expectedPayload) => {
      const { client, calls } = buildClientWithCapture();

      const { result } = renderHook(() => usePracticeEvents(), {
        wrapper: ({ children }) => (
          <Wrapper client={client}>{children}</Wrapper>
        ),
      });

      await act(async () => {
        if (mutation === "requestHint") await result.current.requestHint({ turnId: TURN_A });
        if (mutation === "pauseSession") await result.current.pauseSession();
        if (mutation === "resumeSession") await result.current.resumeSession();
      });

      const last = calls.at(-1)!;
      expect(last.headers.get("Idempotency-Key")).toBeNull();
      const body = readBody(last);
      expect(body.kind).toBe(expectedKind);
      expect(body.payload).toEqual(expectedPayload);
      expect(UUID_V7_REGEX.test(body.clientEventId as string)).toBe(true);
    },
  );

  it("retrying the same submitAnswer reuses the same clientEventId", async () => {
    const { client, calls } = buildClientWithCapture("default", true);

    const { result } = renderHook(() => usePracticeEvents(), {
      wrapper: ({ children }) => (
        <Wrapper client={client}>{children}</Wrapper>
      ),
    });

    // First call throws (simulated network failure)
    await act(async () => {
      try {
        await result.current.submitAnswer({ turnId: TURN_A, answerText: "x" });
      } catch (err) {
        expect(err).toBeInstanceOf(Error);
      }
    });

    // Retry the SAME action — should reuse the clientEventId
    await act(async () => {
      await result.current.submitAnswer({ turnId: TURN_A, answerText: "x" });
    });

    const events = eventCalls(calls);
    expect(events.length).toBeGreaterThanOrEqual(2);
    const first = readBody(events[0]!);
    const second = readBody(events[1]!);
    expect(second.clientEventId).toBe(first.clientEventId);
  });

  it("a fresh user action mints a new clientEventId", async () => {
    const { client, calls } = buildClientWithCapture();

    const { result } = renderHook(() => usePracticeEvents(), {
      wrapper: ({ children }) => (
        <Wrapper client={client}>{children}</Wrapper>
      ),
    });

    await act(async () => {
      await result.current.submitAnswer({ turnId: TURN_A, answerText: "first" });
    });
    await act(async () => {
      await result.current.submitAnswer({ turnId: TURN_A, answerText: "second" });
    });

    const events = eventCalls(calls);
    expect(events.length).toBe(2);
    const first = readBody(events[0]!);
    const second = readBody(events[1]!);
    expect(first.clientEventId).not.toBe(second.clientEventId);
  });

  it("returns the SessionEventResult from the server response", async () => {
    const { client } = buildClientWithCapture();

    const { result } = renderHook(() => usePracticeEvents(), {
      wrapper: ({ children }) => (
        <Wrapper client={client}>{children}</Wrapper>
      ),
    });

    let response;
    await act(async () => {
      response = await result.current.submitAnswer({
        turnId: TURN_A,
        answerText: "hi",
      });
    });

    expect(response).toBeDefined();
    expect(response!.acknowledged).toBe(true);
    expect(response!.assistantAction.type).toBe("ask_follow_up");
  });

  it("waits for InterviewContext.sessionId before allowing mutations", async () => {
    const { client, calls } = buildClientWithCapture();

    function NoSessionWrapper({ children }: { children: ReactNode }) {
      return (
        <InterviewContextProvider>
          <AppRuntimeProvider client={client}>{children}</AppRuntimeProvider>
        </InterviewContextProvider>
      );
    }

    const { result } = renderHook(() => usePracticeEvents(), {
      wrapper: NoSessionWrapper,
    });

    // sessionId missing → submitAnswer must reject with a clear error (not POST)
    await waitFor(() => expect(result.current.ready).toBe(false));
    await expect(
      result.current.submitAnswer({ turnId: TURN_A, answerText: "x" }),
    ).rejects.toThrow(/sessionId/);
    expect(eventCalls(calls).length).toBe(0);
  });
});
