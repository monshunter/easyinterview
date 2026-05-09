// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";

import getJobMatchProfileFixture from "../../../../../openapi/fixtures/JobMatch/getJobMatchProfile.json";

import { useJobMatchProfile } from "./useJobMatchProfile";

function buildClient(scenario?: string) {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([getJobMatchProfileFixture]),
      scenario ? { scenario } : undefined,
    ),
  });
}

function wrapWithRuntime(client: EasyInterviewClient) {
  return ({ children }: { children: ReactNode }) => (
    <AppRuntimeProvider client={client}>{children}</AppRuntimeProvider>
  );
}

describe("useJobMatchProfile", () => {
  it("starts in loading state with null data and null error", () => {
    const client = buildClient("default");
    const { result } = renderHook(() => useJobMatchProfile(), {
      wrapper: wrapWithRuntime(client),
    });

    expect(result.current.loading).toBe(true);
    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
  });

  it("calls getJobMatchProfile exactly once on mount and returns the profile", async () => {
    const client = buildClient("default");
    const spy = vi.spyOn(client, "getJobMatchProfile");

    const { result } = renderHook(() => useJobMatchProfile(), {
      wrapper: wrapWithRuntime(client),
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(spy).toHaveBeenCalledTimes(1);
    expect(result.current.data).not.toBeNull();
    expect(result.current.data?.displayName).toBe("Alice Example");
    expect(result.current.error).toBeNull();
  });

  it("surfaces error when getJobMatchProfile rejects", async () => {
    const client = buildClient("default");
    vi.spyOn(client, "getJobMatchProfile").mockRejectedValue(
      new Error("HTTP 500 — boom"),
    );

    const { result } = renderHook(() => useJobMatchProfile(), {
      wrapper: wrapWithRuntime(client),
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).not.toBeNull();
    expect(result.current.error?.message).toMatch(/HTTP 500/);
    expect(result.current.data).toBeNull();
  });

  it("returns empty/loading=false when no AppRuntimeProvider is mounted", () => {
    const { result } = renderHook(() => useJobMatchProfile());

    expect(result.current.loading).toBe(false);
    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
  });
});
