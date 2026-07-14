// @vitest-environment jsdom
import { StrictMode, type PropsWithChildren } from "react";

import { renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import listTargetJobsFixture from "../../../../../../openapi/fixtures/TargetJobs/listTargetJobs.json";
import { EasyInterviewClient } from "../../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../api/mockTransport";
import {
  AppRuntimeContext,
  type AppRuntimeValue,
} from "../../../runtime/AppRuntimeProvider";
import { useWorkspaceTargetJobs } from "./useWorkspaceTargetJobs";

function authenticatedRuntime(client: EasyInterviewClient): AppRuntimeValue {
  return {
    client,
    runtime: { status: "loading" },
    auth: {
      status: "authenticated",
      user: {
        id: "01918fa0-0000-7000-8000-00000000feed",
        emailMasked: "te***@example.com",
        displayName: "Test User",
        uiLanguage: "en",
        preferredPracticeLanguage: "en",
        profileCompletionRequired: false,
      },
    },
    refreshAuth: vi.fn(),
  };
}

describe("useWorkspaceTargetJobs read effect", () => {
  it("issues one transport GET under StrictMode and stable semantic runtime rerenders", async () => {
    const urls: string[] = [];
    const fixtureFetch = createFixtureBackedFetch(
      createFixtureRegistry([listTargetJobsFixture]),
    );
    const client = new EasyInterviewClient({
      fetch: async (input, init) => {
        const url =
          typeof input === "string"
            ? input
            : input instanceof URL
              ? input.href
              : input.url;
        urls.push(url);
        return fixtureFetch(input, init);
      },
    });
    const wrapper = ({ children }: PropsWithChildren) => (
      <StrictMode>
        <AppRuntimeContext.Provider value={authenticatedRuntime(client)}>
          {children}
        </AppRuntimeContext.Provider>
      </StrictMode>
    );
    const { result, rerender } = renderHook(() => useWorkspaceTargetJobs(), {
      wrapper,
    });

    await waitFor(() => expect(result.current.jobs.length).toBeGreaterThan(0));
    rerender();
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(urls.filter((url) => url.includes("/targets?"))).toHaveLength(1);
  });
});
