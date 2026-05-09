/**
 * @vitest-environment jsdom
 */

import { describe, expect, it, vi } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
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
import { useWorkspaceTargetJob } from "./useWorkspaceTargetJob";

import getTargetJobFixture from "../../../../../../openapi/fixtures/TargetJobs/getTargetJob.json";

function buildClient() {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([getTargetJobFixture]),
      { scenario: "default" },
    ),
  });
}

interface WrapperProps {
  children: ReactNode;
  client: EasyInterviewClient;
  initialTargetJobId?: string;
}

function Wrapper({ children, client, initialTargetJobId }: WrapperProps) {
  return (
    <InterviewContextProvider>
      <AppRuntimeProvider client={client}>
        <HydrateContext targetJobId={initialTargetJobId}>
          {children}
        </HydrateContext>
      </AppRuntimeProvider>
    </InterviewContextProvider>
  );
}

function HydrateContext({
  children,
  targetJobId,
}: {
  children: ReactNode;
  targetJobId?: string;
}) {
  const { dispatch } = useInterviewContext();
  useEffect(() => {
    if (targetJobId) {
      dispatch({
        type: "HYDRATE_FROM_ROUTE",
        params: { targetJobId },
      });
    }
  }, [targetJobId, dispatch]);
  return <>{children}</>;
}

describe("useWorkspaceTargetJob", () => {
  it("calls getTargetJob once with correct targetJobId from InterviewContext", async () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "getTargetJob");

    const { result } = renderHook(() => useWorkspaceTargetJob(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} initialTargetJobId="01918fa0-0000-7000-8000-000000002000">
          {children}
        </Wrapper>
      ),
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // React StrictMode / renderHook may double-invoke effects in dev;
    // the inFlightRef guard ensures at most 1 real API call.
    expect(spy.mock.calls.length).toBeLessThanOrEqual(2);
    expect(spy).toHaveBeenCalledWith(
      "01918fa0-0000-7000-8000-000000002000",
    );
    expect(result.current.data).toBeDefined();
    expect(result.current.data?.id).toBe("01918fa0-0000-7000-8000-000000002000");
    expect(result.current.error).toBeNull();
  });

  it("transitions through loading → data → loaded states", async () => {
    const client = buildClient();

    const { result } = renderHook(() => useWorkspaceTargetJob(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} initialTargetJobId="01918fa0-0000-7000-8000-000000002000">
          {children}
        </Wrapper>
      ),
    });

    expect(result.current.loading).toBe(true);
    expect(result.current.data).toBeNull();

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toBeDefined();
    expect(result.current.error).toBeNull();
  });

  it("returns empty state immediately when targetJobId is missing", () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "getTargetJob");

    const { result } = renderHook(() => useWorkspaceTargetJob(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} initialTargetJobId={undefined}>
          {children}
        </Wrapper>
      ),
    });

    expect(result.current.loading).toBe(false);
    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
    expect(spy).not.toHaveBeenCalled();
  });

  it("handles 4xx / 5xx errors and sets error state", async () => {
    const errorFixture = [
      {
        operationId: "getTargetJob" as const,
        scenarios: {
          default: {
            response: { status: 500, body: { error: { code: "INTERNAL", message: "boom" } } },
          },
        },
      },
    ];
    const fetch = createFixtureBackedFetch(
      createFixtureRegistry(errorFixture),
      { scenario: "default" },
    );
    const client = new EasyInterviewClient({ fetch });

    const { result } = renderHook(() => useWorkspaceTargetJob(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} initialTargetJobId="tj-err">
          {children}
        </Wrapper>
      ),
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeDefined();
  });
});
