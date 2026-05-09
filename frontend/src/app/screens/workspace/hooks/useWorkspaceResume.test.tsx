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
import { useWorkspaceResume } from "./useWorkspaceResume";

import getResumeFixture from "../../../../../../openapi/fixtures/Resumes/getResume.json";

function buildClient() {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([getResumeFixture]),
      { scenario: "default" },
    ),
  });
}

interface WrapperProps {
  children: ReactNode;
  client: EasyInterviewClient;
  resumeVersionId?: string;
}

function Wrapper({ children, client, resumeVersionId }: WrapperProps) {
  return (
    <InterviewContextProvider>
      <AppRuntimeProvider client={client}>
        <HydrateContext resumeVersionId={resumeVersionId}>
          {children}
        </HydrateContext>
      </AppRuntimeProvider>
    </InterviewContextProvider>
  );
}

function HydrateContext({
  children,
  resumeVersionId,
}: {
  children: ReactNode;
  resumeVersionId?: string;
}) {
  const { dispatch } = useInterviewContext();
  useEffect(() => {
    if (resumeVersionId) {
      dispatch({
        type: "HYDRATE_FROM_ROUTE",
        params: { targetJobId: "tj-1", resumeVersionId },
      });
    }
  }, [resumeVersionId, dispatch]);
  return <>{children}</>;
}

describe("useWorkspaceResume", () => {
  it("calls getResume once with correct resumeVersionId", async () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "getResume");

    const { result } = renderHook(() => useWorkspaceResume(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} resumeVersionId="01918fa0-0000-7000-8000-000000001000">
          {children}
        </Wrapper>
      ),
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(spy).toHaveBeenCalledWith(
      "01918fa0-0000-7000-8000-000000001000",
    );
    expect(result.current.data).toBeDefined();
    expect(result.current.data?.id).toBe("01918fa0-0000-7000-8000-000000001000");
    expect(result.current.error).toBeNull();
  });

  it("transitions through loading → data states", async () => {
    const client = buildClient();

    const { result } = renderHook(() => useWorkspaceResume(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} resumeVersionId="01918fa0-0000-7000-8000-000000001000">
          {children}
        </Wrapper>
      ),
    });

    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toBeDefined();
    expect(result.current.error).toBeNull();
  });

  it("returns empty state when resumeVersionId is missing", () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "getResume");

    const { result } = renderHook(() => useWorkspaceResume(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} resumeVersionId={undefined}>
          {children}
        </Wrapper>
      ),
    });

    expect(result.current.loading).toBe(false);
    expect(result.current.data).toBeNull();
    expect(spy).not.toHaveBeenCalled();
  });

  it("handles 404 error and sets error state", async () => {
    const errorFixture = [
      {
        operationId: "getResume" as const,
        scenarios: {
          default: {
            response: {
              status: 404,
              body: { error: { code: "NOT_FOUND", message: "not found" } },
            },
          },
        },
      },
    ];
    const fetch = createFixtureBackedFetch(
      createFixtureRegistry(errorFixture),
      { scenario: "default" },
    );
    const client = new EasyInterviewClient({ fetch });

    const { result } = renderHook(() => useWorkspaceResume(), {
      wrapper: ({ children }) => (
        <Wrapper client={client} resumeVersionId="rv-notfound">
          {children}
        </Wrapper>
      ),
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBeDefined();
    expect(result.current.data).toBeNull();
  });
});
