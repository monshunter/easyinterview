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
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../api/mockTransport";
import {
  InterviewContextProvider,
  useInterviewContext,
} from "../../../interview-context/InterviewContext";
import { AppRuntimeProvider } from "../../../runtime/AppRuntimeProvider";
import { usePracticeSessionLoader } from "./usePracticeSessionLoader";

import getPracticeSessionFixture from "../../../../../../openapi/fixtures/PracticeSessions/getPracticeSession.json";

const SESSION_A = "01918fa0-0000-7000-8000-000000005000";

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
});
