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
import { useWorkspacePracticePlan } from "./useWorkspacePracticePlan";

import getPracticePlanFixture from "../../../../../../openapi/fixtures/PracticePlans/getPracticePlan.json";

function buildClient(scenario = "default") {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([getPracticePlanFixture]),
      { scenario },
    ),
  });
}

interface WrapperProps {
  children: ReactNode;
  client: EasyInterviewClient;
  planId?: string;
}

function Wrapper({ children, client, planId }: WrapperProps) {
  return (
    <InterviewContextProvider>
      <AppRuntimeProvider client={client}>
        <Hydrate planId={planId}>{children}</Hydrate>
      </AppRuntimeProvider>
    </InterviewContextProvider>
  );
}

function Hydrate({ children, planId }: { children: ReactNode; planId?: string }) {
  const { dispatch } = useInterviewContext();
  useEffect(() => {
    if (planId) {
      dispatch({ type: "HYDRATE_FROM_ROUTE", params: { targetJobId: "tj-1", planId } });
    }
  }, [planId, dispatch]);
  return <>{children}</>;
}

describe("useWorkspacePracticePlan", () => {
  it("calls getPracticePlan and dispatches MERGE_PRACTICE_PLAN on ready status", async () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "getPracticePlan");

    const { result } = renderHook(() => useWorkspacePracticePlan(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} planId="01918fa0-0000-7000-8000-000000004000">
          {children}
        </Wrapper>
      ),
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(spy).toHaveBeenCalledWith("01918fa0-0000-7000-8000-000000004000");
    expect(result.current.ready).toBe(true);
    expect(result.current.error).toBeNull();
  });

  it("returns ready=false for archived status", async () => {
    const client = buildClient("archived");

    const { result } = renderHook(() => useWorkspacePracticePlan(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} planId="01918fa0-0000-7000-8000-000000004001">
          {children}
        </Wrapper>
      ),
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.ready).toBe(false);
  });

  it("handles not-found (404) gracefully", async () => {
    const client = buildClient("not-found");

    const { result } = renderHook(() => useWorkspacePracticePlan(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} planId="rv-notfound">
          {children}
        </Wrapper>
      ),
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBeDefined();
    expect(result.current.ready).toBe(false);
  });

  it("returns empty state when planId is missing", () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "getPracticePlan");

    renderHook(() => useWorkspacePracticePlan(), {
      wrapper: ({ children }) => (
        <Wrapper client={client}>
          {children}
        </Wrapper>
      ),
    });

    expect(spy).not.toHaveBeenCalled();
  });
});
