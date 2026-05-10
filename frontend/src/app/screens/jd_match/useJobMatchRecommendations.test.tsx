// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";

import listJobRecommendationsFixture from "../../../../../openapi/fixtures/JobMatch/listJobRecommendations.json";

import { useJobMatchRecommendations } from "./useJobMatchRecommendations";

function buildClient(scenario: string) {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([listJobRecommendationsFixture]),
      { scenario },
    ),
  });
}

function wrapWithRuntime(client: EasyInterviewClient) {
  return ({ children }: { children: ReactNode }) => (
    <AppRuntimeProvider client={client}>{children}</AppRuntimeProvider>
  );
}

describe("useJobMatchRecommendations (item 3.2)", () => {
  it("starts in loading=true with empty items and null error", () => {
    const client = buildClient("default");
    const { result } = renderHook(() => useJobMatchRecommendations(), {
      wrapper: wrapWithRuntime(client),
    });
    expect(result.current.loading).toBe(true);
    expect(result.current.items).toEqual([]);
    expect(result.current.error).toBeNull();
  });

  it("calls listJobRecommendations exactly once on mount", async () => {
    const client = buildClient("default");
    const spy = vi.spyOn(client, "listJobRecommendations");
    const { result } = renderHook(() => useJobMatchRecommendations(), {
      wrapper: wrapWithRuntime(client),
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(spy).toHaveBeenCalledTimes(1);
  });

  it("variant=default → loads 3 recommendations", async () => {
    const client = buildClient("default");
    const { result } = renderHook(() => useJobMatchRecommendations(), {
      wrapper: wrapWithRuntime(client),
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.items).toHaveLength(3);
    expect(result.current.error).toBeNull();
  });

  it("variant=empty → loads 0 recommendations and no error", async () => {
    const client = buildClient("empty");
    const { result } = renderHook(() => useJobMatchRecommendations(), {
      wrapper: wrapWithRuntime(client),
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.items).toEqual([]);
    expect(result.current.error).toBeNull();
  });

  it("variant=one → loads exactly 1 recommendation", async () => {
    const client = buildClient("one");
    const { result } = renderHook(() => useJobMatchRecommendations(), {
      wrapper: wrapWithRuntime(client),
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.items).toHaveLength(1);
  });

  it("variant=many → loads 4 recommendations", async () => {
    const client = buildClient("many");
    const { result } = renderHook(() => useJobMatchRecommendations(), {
      wrapper: wrapWithRuntime(client),
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.items).toHaveLength(4);
  });

  it("variant=failed → surfaces error and keeps items empty", async () => {
    const client = buildClient("failed");
    const { result } = renderHook(() => useJobMatchRecommendations(), {
      wrapper: wrapWithRuntime(client),
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.items).toEqual([]);
    expect(result.current.error).not.toBeNull();
  });

  it("retry refetches and recovers from error", async () => {
    const client = buildClient("default");
    const spy = vi.spyOn(client, "listJobRecommendations");
    spy.mockRejectedValueOnce(new Error("HTTP 502 — boom"));
    const { result } = renderHook(() => useJobMatchRecommendations(), {
      wrapper: wrapWithRuntime(client),
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).not.toBeNull();
    spy.mockRestore();

    await act(async () => {
      result.current.retry();
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBeNull();
    expect(result.current.items.length).toBeGreaterThan(0);
  });

  it("returns inert state when no AppRuntimeProvider is mounted", () => {
    const { result } = renderHook(() => useJobMatchRecommendations());
    expect(result.current.loading).toBe(false);
    expect(result.current.items).toEqual([]);
    expect(result.current.error).toBeNull();
  });

  it("exposes pageInfo.hasMore for cursor pagination wiring", async () => {
    const client = buildClient("default");
    const { result } = renderHook(() => useJobMatchRecommendations(), {
      wrapper: wrapWithRuntime(client),
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.pageInfo).toMatchObject({ hasMore: false });
  });
});
