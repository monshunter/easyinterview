/**
 * Shared test harness for PracticeScreen interaction tests (Phase 3+).
 * Mounts PracticeScreen with a fixture-backed client, navigation spy,
 * and a probe for inspecting InterviewContext.
 */

import { vi } from "vitest";
import { render, type RenderResult } from "@testing-library/react";
import { useEffect, type FC, type ReactNode } from "react";

import { EasyInterviewClient } from "../../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../api/mockTransport";
import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import {
  InterviewContextProvider,
  useInterviewContext,
} from "../../../interview-context/InterviewContext";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import { AppRuntimeProvider } from "../../../runtime/AppRuntimeProvider";
import type { Route } from "../../../routes";

import getPracticeSessionFixture from "../../../../../../openapi/fixtures/PracticeSessions/getPracticeSession.json";
import appendSessionEventFixture from "../../../../../../openapi/fixtures/PracticeSessions/appendSessionEvent.json";
import completePracticeSessionFixture from "../../../../../../openapi/fixtures/PracticeSessions/completePracticeSession.json";
import createPracticeVoiceTurnFixture from "../../../../../../openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json";
import { PracticeScreen } from "../PracticeScreen";

export const SESSION_A = "01918fa0-0000-7000-8000-000000005000";
export const TURN_A = "01918fa0-0000-7000-8000-000000006000";
export const PLAN_A = "01918fa0-0000-7000-8000-000000004000";
export const TARGET_JOB_A = "01918fa0-0000-7000-8000-000000002000";
export const RESUME_A = "01918fa0-0000-7000-8000-000000001000";

export interface CapturedRequest {
  url: string;
  method: string;
  headers: Headers;
  bodyText: string | null;
}

export interface BuildFixtureClientOptions {
  scenario?: string;
  /** Per-operationId scenario override (takes precedence over `scenario`). */
  scenarioByOp?: Partial<
    Record<
      | "getPracticeSession"
      | "appendSessionEvent"
      | "completePracticeSession"
      | "createPracticeVoiceTurn",
      string
    >
  >;
  forceAppendFailFirstN?: number;
  forceCompleteFailFirstN?: number;
}

export function buildPracticeClient(
  opts: BuildFixtureClientOptions = {},
): { client: EasyInterviewClient; calls: CapturedRequest[] } {
  const calls: CapturedRequest[] = [];
  const fixtureFetch = createFixtureBackedFetch(
    createFixtureRegistry([
      getPracticeSessionFixture,
      appendSessionEventFixture,
      completePracticeSessionFixture,
      createPracticeVoiceTurnFixture,
    ]),
    { scenario: opts.scenario ?? "default" },
  );
  let appendAttempts = 0;
  let completeAttempts = 0;
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
    let scenarioOverride: string | undefined;
    if (/\/practice\/sessions\/[^/]+$/.test(path) && method === "GET") {
      scenarioOverride = opts.scenarioByOp?.getPracticeSession;
    } else if (
      /\/practice\/sessions\/[^/]+\/events$/.test(path) &&
      method === "POST"
    ) {
      appendAttempts += 1;
      if (
        opts.forceAppendFailFirstN &&
        appendAttempts <= opts.forceAppendFailFirstN
      ) {
        throw new Error("simulated network failure");
      }
      scenarioOverride = opts.scenarioByOp?.appendSessionEvent;
    } else if (
      /\/practice\/sessions\/[^/]+\/complete$/.test(path) &&
      method === "POST"
    ) {
      completeAttempts += 1;
      if (
        opts.forceCompleteFailFirstN &&
        completeAttempts <= opts.forceCompleteFailFirstN
      ) {
        throw new Error("simulated network failure");
      }
      scenarioOverride = opts.scenarioByOp?.completePracticeSession;
    } else if (
      /\/practice\/sessions\/[^/]+\/voice-turns$/.test(path) &&
      method === "POST"
    ) {
      scenarioOverride = opts.scenarioByOp?.createPracticeVoiceTurn;
    }
    if (scenarioOverride) {
      const merged = new Headers(init?.headers ?? {});
      merged.set("Prefer", `example=${scenarioOverride}`);
      return fixtureFetch(input, { ...init, headers: merged });
    }
    return fixtureFetch(input, init);
  };
  return {
    client: new EasyInterviewClient({ fetch: wrappedFetch }),
    calls,
  };
}

export function eventCalls(all: CapturedRequest[]): CapturedRequest[] {
  return all.filter(
    (c) =>
      c.method === "POST" &&
      /\/practice\/sessions\/[^/]+\/events$/.test(
        new URL(c.url, "http://x").pathname,
      ),
  );
}

export function getSessionCalls(all: CapturedRequest[]): CapturedRequest[] {
  return all.filter(
    (c) =>
      c.method === "GET" &&
      /\/practice\/sessions\/[^/]+$/.test(
        new URL(c.url, "http://x").pathname,
      ),
  );
}

export function completeCalls(all: CapturedRequest[]): CapturedRequest[] {
  return all.filter(
    (c) =>
      c.method === "POST" &&
      /\/practice\/sessions\/[^/]+\/complete$/.test(
        new URL(c.url, "http://x").pathname,
      ),
  );
}

export function voiceTurnCalls(all: CapturedRequest[]): CapturedRequest[] {
  return all.filter(
    (c) =>
      c.method === "POST" &&
      /\/practice\/sessions\/[^/]+\/voice-turns$/.test(
        new URL(c.url, "http://x").pathname,
      ),
  );
}

export function readBody(call: CapturedRequest): Record<string, unknown> {
  if (!call.bodyText) throw new Error("expected JSON body");
  return JSON.parse(call.bodyText) as Record<string, unknown>;
}

export interface MountPracticeOptions {
  client?: EasyInterviewClient;
  routeParams?: Partial<Route["params"]>;
}

export interface MountPracticeResult extends RenderResult {
  nav: ReturnType<typeof vi.fn>;
}

export function defaultRoute(overrides: Partial<Route["params"]> = {}): Route {
  return {
    name: "practice",
    params: {
      sessionId: SESSION_A,
      planId: PLAN_A,
      targetJobId: TARGET_JOB_A,
      jdId: "jd-1",
      resumeId: RESUME_A,
      roundId: "round-tech1",
      mode: "text",
      modality: "text",
      practiceMode: "assisted",
      practiceGoal: "baseline",
      hintUsed: "false",
      hintCount: "0",
      ...overrides,
    },
  };
}

const HydrateContext: FC<{ params: Route["params"]; children: ReactNode }> = ({
  params,
  children,
}) => {
  const { dispatch } = useInterviewContext();
  useEffect(() => {
    dispatch({ type: "HYDRATE_FROM_ROUTE", params });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
  return <>{children}</>;
};

export function mountPracticeScreen(
  options: MountPracticeOptions = {},
): MountPracticeResult {
  const nav = vi.fn();
  const route = defaultRoute(options.routeParams);
  const client = options.client ?? buildPracticeClient().client;
  const result = render(
    <DisplayPreferencesProvider>
      <InterviewContextProvider>
        <AppRuntimeProvider client={client}>
          <NavigationProvider value={{ navigate: nav }}>
            <HydrateContext params={route.params}>
              <PracticeScreen route={route} />
            </HydrateContext>
          </NavigationProvider>
        </AppRuntimeProvider>
      </InterviewContextProvider>
    </DisplayPreferencesProvider>,
  );
  return { ...result, nav };
}
