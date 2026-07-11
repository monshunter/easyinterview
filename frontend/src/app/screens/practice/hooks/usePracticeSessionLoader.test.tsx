/**
 * @vitest-environment jsdom
 *
 * Item 1.3 — usePracticeSessionLoader: 5 states (idle / loading / data /
 * sessionLost / error), explicit refresh, auto refresh on visibility / focus
 * / online, MERGE_SESSION dispatch on success.
 */

import { describe, expect, it, vi } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import { useEffect, type ReactNode } from "react";

import { EasyInterviewClient } from "../../../../api/generated/client";
import type {
  PracticeSession,
  TargetJob,
} from "../../../../api/generated/types";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../api/mockTransport";
import {
  InterviewContextProvider,
  useInterviewContext,
} from "../../../interview-context/InterviewContext";
import {
  AppRuntimeContext,
  AppRuntimeProvider,
  type AppRuntimeValue,
} from "../../../runtime/AppRuntimeProvider";
import { usePracticeSessionLoader } from "./usePracticeSessionLoader";
import { usePracticeTargetDisplay } from "../usePracticeTargetDisplay";

import getPracticeSessionFixture from "../../../../../../openapi/fixtures/PracticeSessions/getPracticeSession.json";

const SESSION_A = "01918fa0-0000-7000-8000-000000005000";
const SESSION_B = "01918fa0-0000-7000-8000-000000005001";
const TARGET_A = "01918fa0-0000-7000-8000-000000002000";
const TARGET_B = "01918fa0-0000-7000-8000-000000002001";

function buildClient(scenario: string = "default") {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([getPracticeSessionFixture]),
      { scenario },
    ),
  });
}

interface WrapperProps {
  children: ReactNode;
  client: EasyInterviewClient;
  initialSessionId?: string;
}

function Wrapper({ children, client, initialSessionId }: WrapperProps) {
  return (
    <InterviewContextProvider>
      <AppRuntimeProvider client={client}>
        <HydrateContext sessionId={initialSessionId}>{children}</HydrateContext>
      </AppRuntimeProvider>
    </InterviewContextProvider>
  );
}

function HydrateContext({
  children,
  sessionId,
}: {
  children: ReactNode;
  sessionId?: string;
}) {
  const { dispatch } = useInterviewContext();
  useEffect(() => {
    if (sessionId) {
      dispatch({ type: "HYDRATE_FROM_ROUTE", params: { sessionId } });
    }
  }, [sessionId, dispatch]);
  return <>{children}</>;
}

describe("usePracticeSessionLoader", () => {
  it("returns sessionLost immediately when sessionId is missing (no fetch)", async () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "getPracticeSession");

    const { result } = renderHook(() => usePracticeSessionLoader(), {
      wrapper: ({ children }) => (
        <Wrapper client={client}>{children}</Wrapper>
      ),
    });

    await waitFor(() => {
      expect(result.current.state).toBe("sessionLost");
    });
    expect(spy).not.toHaveBeenCalled();
    expect(result.current.data).toBeNull();
  });

  it("transitions loading → data on success and dispatches MERGE_SESSION", async () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "getPracticeSession");

    const { result } = renderHook(() => usePracticeSessionLoader(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} initialSessionId={SESSION_A}>
          {children}
        </Wrapper>
      ),
    });

    expect(result.current.state).toBe("loading");

    await waitFor(() => {
      expect(result.current.state).toBe("data");
    });

    expect(spy).toHaveBeenCalledWith(SESSION_A);
    expect(result.current.data?.id).toBe(SESSION_A);
    expect(result.current.error).toBeNull();
  });

  it("transitions to sessionLost when getPracticeSession returns 404", async () => {
    const client = buildClient("missing-session");

    const { result } = renderHook(() => usePracticeSessionLoader(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} initialSessionId={SESSION_A}>
          {children}
        </Wrapper>
      ),
    });

    await waitFor(() => {
      expect(result.current.state).toBe("sessionLost");
    });
    expect(result.current.data).toBeNull();
  });

  it("transitions to error on 5xx (non-404)", async () => {
    const failingClient = new EasyInterviewClient({
      fetch: async () =>
        new Response(JSON.stringify({ error: { code: "INTERNAL", message: "boom" } }), {
          status: 500,
          headers: { "Content-Type": "application/json" },
        }),
    });

    const { result } = renderHook(() => usePracticeSessionLoader(), {
      wrapper: ({ children }) => (
        <Wrapper client={failingClient} initialSessionId={SESSION_A}>
          {children}
        </Wrapper>
      ),
    });

    await waitFor(() => {
      expect(result.current.state).toBe("error");
    });
    expect(result.current.error).toBeInstanceOf(Error);
  });

  it("refresh() re-fetches the session", async () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "getPracticeSession");

    const { result } = renderHook(() => usePracticeSessionLoader(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} initialSessionId={SESSION_A}>
          {children}
        </Wrapper>
      ),
    });

    await waitFor(() => {
      expect(result.current.state).toBe("data");
    });
    const before = spy.mock.calls.length;

    await act(async () => {
      result.current.refresh();
    });

    await waitFor(() => {
      expect(spy.mock.calls.length).toBeGreaterThan(before);
    });
  });

  it("auto-refreshes on window focus", async () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "getPracticeSession");

    renderHook(() => usePracticeSessionLoader(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} initialSessionId={SESSION_A}>
          {children}
        </Wrapper>
      ),
    });

    await waitFor(() => {
      expect(spy).toHaveBeenCalled();
    });
    const before = spy.mock.calls.length;

    await act(async () => {
      window.dispatchEvent(new Event("focus"));
    });

    await waitFor(() => {
      expect(spy.mock.calls.length).toBeGreaterThan(before);
    });
  });

  it("auto-refreshes when window goes online", async () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "getPracticeSession");

    renderHook(() => usePracticeSessionLoader(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} initialSessionId={SESSION_A}>
          {children}
        </Wrapper>
      ),
    });

    await waitFor(() => {
      expect(spy).toHaveBeenCalled();
    });
    const before = spy.mock.calls.length;

    await act(async () => {
      window.dispatchEvent(new Event("online"));
    });

    await waitFor(() => {
      expect(spy.mock.calls.length).toBeGreaterThan(before);
    });
  });

  it("auto-refreshes on document visibility change to visible", async () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "getPracticeSession");

    renderHook(() => usePracticeSessionLoader(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} initialSessionId={SESSION_A}>
          {children}
        </Wrapper>
      ),
    });

    await waitFor(() => {
      expect(spy).toHaveBeenCalled();
    });
    const before = spy.mock.calls.length;

    await act(async () => {
      Object.defineProperty(document, "visibilityState", {
        configurable: true,
        get: () => "visible",
      });
      document.dispatchEvent(new Event("visibilitychange"));
    });

    await waitFor(() => {
      expect(spy.mock.calls.length).toBeGreaterThan(before);
    });
  });

  it("MERGE_SESSION write keeps InterviewContext.sessionId populated after success", async () => {
    const client = buildClient();
    let probedSessionId: string | undefined;

    function SessionIdProbe() {
      const { ctx } = useInterviewContext();
      probedSessionId = ctx.sessionId;
      return null;
    }

    const { result } = renderHook(() => usePracticeSessionLoader(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} initialSessionId={SESSION_A}>
          <SessionIdProbe />
          {children}
        </Wrapper>
      ),
    });

    await waitFor(() => {
      expect(result.current.state).toBe("data");
    });
    expect(probedSessionId).toBe(SESSION_A);
  });

  it("adopts a matching mutation snapshot immediately and rejects a foreign session", async () => {
    const initial = practiceSession(SESSION_A, TARGET_A, "turn-a", 1);
    const client = clientWithSessionAndTarget({
      getPracticeSession: async () => initial,
    });

    const { result } = renderHook(() => usePracticeSessionLoader(SESSION_A), {
      wrapper: directRuntimeWrapper(client),
    });

    await waitFor(() => expect(result.current.data?.id).toBe(SESSION_A));
    const advanced = practiceSession(SESSION_A, TARGET_A, "turn-b", 2);

    act(() => {
      expect(result.current.adopt(advanced)).toBe(true);
    });
    expect(result.current.state).toBe("data");
    expect(result.current.data?.currentTurn?.id).toBe("turn-b");

    act(() => {
      expect(
        result.current.adopt(
          practiceSession(SESSION_B, TARGET_B, "foreign-turn", 1),
        ),
      ).toBe(false);
    });
    expect(result.current.data?.id).toBe(SESSION_A);
    expect(result.current.data?.currentTurn?.id).toBe("turn-b");
  });

  it("clears the A snapshot synchronously on session A to B and never exposes A target identity", async () => {
    const sessionBRequest = deferred<PracticeSession>();
    const getPracticeSession = vi.fn((sessionId: string) =>
      sessionId === SESSION_A
        ? Promise.resolve(practiceSession(SESSION_A, TARGET_A, "turn-a", 1))
        : sessionBRequest.promise,
    );
    const getTargetJob = vi.fn(async (targetJobId: string) =>
      targetJob(
        targetJobId,
        targetJobId === TARGET_A ? "Company A" : "Company B",
      ),
    );
    const client = clientWithSessionAndTarget({
      getPracticeSession,
      getTargetJob,
    });

    const { result, rerender } = renderHook(
      ({ sessionId, routeTargetJobId }) => {
        const loader = usePracticeSessionLoader(sessionId);
        const target = usePracticeTargetDisplay({
          session: loader.data
            ? { targetJobId: loader.data.targetJobId }
            : null,
          routeTargetJobId,
        });
        return { loader, target };
      },
      {
        initialProps: {
          sessionId: SESSION_A,
          routeTargetJobId: TARGET_A,
        },
        wrapper: directRuntimeWrapper(client),
      },
    );

    await waitFor(() =>
      expect(result.current.target.companyName).toBe("Company A"),
    );

    rerender({ sessionId: SESSION_B, routeTargetJobId: TARGET_B });

    expect(result.current.loader.data).toBeNull();
    expect(result.current.loader.state).toBe("loading");
    expect(result.current.target.targetJobId).toBe(TARGET_B);
    expect(result.current.target.companyName).toBeNull();

    await act(async () => {
      sessionBRequest.resolve(
        practiceSession(SESSION_B, TARGET_B, "turn-b", 1),
      );
      await sessionBRequest.promise;
    });
    await waitFor(() =>
      expect(result.current.target.companyName).toBe("Company B"),
    );
    expect(result.current.loader.data?.id).toBe(SESSION_B);
  });
});

function practiceSession(
  id: string,
  targetJobId: string,
  turnId: string,
  turnIndex: number,
): PracticeSession {
  return {
    id,
    planId: "plan-1",
    targetJobId,
    status: "running",
    language: "zh-CN",
    hintsEnabled: true,
    turnCount: turnIndex,
    currentTurn: {
      id: turnId,
      turnIndex,
      questionText: `Question ${turnIndex}`,
      status: "asked",
    },
    createdAt: "2026-07-11T00:00:00Z",
    updatedAt: "2026-07-11T00:00:00Z",
  };
}

function targetJob(id: string, companyName: string): TargetJob {
  return {
    id,
    companyName,
    title: `${companyName} Role`,
  } as TargetJob;
}

function clientWithSessionAndTarget(overrides: {
  getPracticeSession?: (sessionId: string) => Promise<PracticeSession>;
  getTargetJob?: (targetJobId: string) => Promise<TargetJob>;
}): EasyInterviewClient {
  return {
    getPracticeSession: overrides.getPracticeSession ?? vi.fn(),
    getTargetJob: overrides.getTargetJob ?? vi.fn(),
  } as unknown as EasyInterviewClient;
}

function directRuntimeWrapper(client: EasyInterviewClient) {
  const value: AppRuntimeValue = {
    client,
    runtime: { status: "loading" },
    auth: { status: "unauthenticated" },
    refreshAuth: vi.fn(),
  };
  return function DirectRuntimeWrapper({ children }: { children: ReactNode }) {
    return (
      <InterviewContextProvider>
        <AppRuntimeContext.Provider value={value}>
          {children}
        </AppRuntimeContext.Provider>
      </InterviewContextProvider>
    );
  };
}

interface Deferred<T> {
  promise: Promise<T>;
  resolve: (value: T) => void;
}

function deferred<T>(): Deferred<T> {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((onResolve) => {
    resolve = onResolve;
  });
  return { promise, resolve };
}
