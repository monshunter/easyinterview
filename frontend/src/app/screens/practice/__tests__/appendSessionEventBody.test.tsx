/**
 * @vitest-environment jsdom
 *
 * Item 2.6 — body of every appendSessionEvent request must conform to the
 * OpenAPI PracticeSessionEventRequest schema. Payload typing per kind:
 *   answer_submitted   → { turnId, answerText }
 *   hint_requested     → { turnId }
 *   session_paused     → {}
 *   session_resumed    → {}
 */

import { describe, expect, it } from "vitest";
import { act, renderHook } from "@testing-library/react";
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
import { usePracticeEvents } from "../hooks/usePracticeEvents";

import appendSessionEventFixture from "../../../../../../openapi/fixtures/PracticeSessions/appendSessionEvent.json";

const SESSION_A = "01918fa0-0000-7000-8000-000000005000";
const TURN_A = "01918fa0-0000-7000-8000-000000006000";

const REQUIRED_KEYS = ["clientEventId", "kind", "occurredAt", "payload"] as const;
const ALLOWED_KINDS = new Set([
  "answer_submitted",
  "hint_requested",
  "session_paused",
  "session_resumed",
]);

interface CapturedRequest {
  url: string;
  method: string;
  bodyText: string | null;
}

function buildClient(): { client: EasyInterviewClient; calls: CapturedRequest[] } {
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
    let bodyText: string | null = null;
    if (typeof init?.body === "string") bodyText = init.body;
    calls.push({ url, method: (init?.method ?? "GET").toUpperCase(), bodyText });
    return fixtureFetch(input, init);
  };
  return {
    client: new EasyInterviewClient({ fetch: wrappedFetch }),
    calls,
  };
}

function eventBodies(all: CapturedRequest[]): Record<string, unknown>[] {
  return all
    .filter(
      (c) =>
        c.method === "POST" &&
        /\/practice\/sessions\/[^/]+\/events$/.test(
          new URL(c.url, "http://x").pathname,
        ),
    )
    .map((c) => JSON.parse(c.bodyText ?? "{}") as Record<string, unknown>);
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

function assertSchemaShape(body: Record<string, unknown>): void {
  for (const k of REQUIRED_KEYS) {
    expect(body, `missing required key ${k}`).toHaveProperty(k);
  }
  expect(typeof body.clientEventId).toBe("string");
  expect(UUID_V7_REGEX.test(body.clientEventId as string)).toBe(true);
  expect(typeof body.occurredAt).toBe("string");
  // RFC 3339 / ISO 8601 sanity: parse must produce a finite timestamp.
  expect(Number.isFinite(Date.parse(body.occurredAt as string))).toBe(true);
  expect(typeof body.kind).toBe("string");
  expect(ALLOWED_KINDS.has(body.kind as string)).toBe(true);
  expect(typeof body.payload).toBe("object");
  // No out-of-scope / disallowed top-level keys.
  for (const key of Object.keys(body)) {
    expect(REQUIRED_KEYS).toContain(key as (typeof REQUIRED_KEYS)[number]);
  }
}

describe("appendSessionEvent body schema", () => {
  it("answer_submitted: payload = {turnId, answerText} only", async () => {
    const { client, calls } = buildClient();
    const { result } = renderHook(() => usePracticeEvents(), {
      wrapper: ({ children }) => (
        <Wrapper client={client}>{children}</Wrapper>
      ),
    });
    await act(async () => {
      await result.current.submitAnswer({
        turnId: TURN_A,
        answerText: "hello world",
      });
    });
    const body = eventBodies(calls).at(-1)!;
    assertSchemaShape(body);
    expect(body.kind).toBe("answer_submitted");
    expect(body.payload).toEqual({ turnId: TURN_A, answerText: "hello world" });
  });

  it("hint_requested: payload = {turnId} only", async () => {
    const { client, calls } = buildClient();
    const { result } = renderHook(() => usePracticeEvents(), {
      wrapper: ({ children }) => (
        <Wrapper client={client}>{children}</Wrapper>
      ),
    });
    await act(async () => {
      await result.current.requestHint({ turnId: TURN_A });
    });
    const body = eventBodies(calls).at(-1)!;
    assertSchemaShape(body);
    expect(body.kind).toBe("hint_requested");
    expect(body.payload).toEqual({ turnId: TURN_A });
  });

  it.each([
    ["session_paused", "pauseSession"],
    ["session_resumed", "resumeSession"],
  ] as const)("%s: payload = {} (empty object)", async (expectedKind, mutation) => {
    const { client, calls } = buildClient();
    const { result } = renderHook(() => usePracticeEvents(), {
      wrapper: ({ children }) => (
        <Wrapper client={client}>{children}</Wrapper>
      ),
    });
    await act(async () => {
      if (mutation === "pauseSession") await result.current.pauseSession();
      if (mutation === "resumeSession") await result.current.resumeSession();
    });
    const body = eventBodies(calls).at(-1)!;
    assertSchemaShape(body);
    expect(body.kind).toBe(expectedKind);
    expect(body.payload).toEqual({});
  });

});
