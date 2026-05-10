// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../api/generated/client";
import type { JobMatchRecommendation } from "../../../api/generated/types";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";

import getJobRecommendationFixture from "../../../../../openapi/fixtures/JobMatch/getJobRecommendation.json";

import { useJobRecommendation } from "./useJobRecommendation";

function buildClient(scenario = "default") {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([getJobRecommendationFixture]),
      { scenario },
    ),
  });
}

function wrapWithRuntime(client: EasyInterviewClient) {
  return ({ children }: { children: ReactNode }) => (
    <AppRuntimeProvider client={client}>{children}</AppRuntimeProvider>
  );
}

describe("useJobRecommendation (item 3.11)", () => {
  it("calls getJobRecommendation once for the selected id", async () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "getJobRecommendation");
    const { result } = renderHook(
      () => useJobRecommendation("01918fa0-0000-7000-8000-00000000a001"),
      { wrapper: wrapWithRuntime(client) },
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(spy).toHaveBeenCalledTimes(1);
    expect(spy).toHaveBeenCalledWith(
      "01918fa0-0000-7000-8000-00000000a001",
    );
    expect(result.current.data?.id).toBe(
      "01918fa0-0000-7000-8000-00000000a001",
    );
    expect(result.current.error).toBeNull();
  });

  it("does not fetch when selected id is null", async () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "getJobRecommendation");
    const { result } = renderHook(() => useJobRecommendation(null), {
      wrapper: wrapWithRuntime(client),
    });

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.loading).toBe(false);
    expect(result.current.data).toBeNull();
    expect(spy).not.toHaveBeenCalled();
  });

  it("surfaces the failed fixture as an error without stale detail data", async () => {
    const client = buildClient("failed");
    const { result } = renderHook(
      () => useJobRecommendation("01918fa0-0000-7000-8000-00000000a001"),
      { wrapper: wrapWithRuntime(client) },
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.data).toBeNull();
    expect(result.current.error?.message).toContain("HTTP 502");
    expect(result.current.error?.message).toContain(
      "Recommendation detail pipeline timed out",
    );
  });

  it("retries after a detail failure and replaces the error with fresh data", async () => {
    const client = buildClient();
    const defaultDetail = getJobRecommendationFixture.scenarios.default.response
      .body as JobMatchRecommendation;
    const spy = vi
      .spyOn(client, "getJobRecommendation")
      .mockRejectedValueOnce(new Error("detail unavailable"))
      .mockResolvedValueOnce({
        ...defaultDetail,
        title: "Retry detail title",
      });

    const { result } = renderHook(
      () => useJobRecommendation("01918fa0-0000-7000-8000-00000000a001"),
      { wrapper: wrapWithRuntime(client) },
    );

    await waitFor(() =>
      expect(result.current.error?.message).toBe("detail unavailable"),
    );
    expect(result.current.data).toBeNull();

    act(() => result.current.retry());

    await waitFor(() =>
      expect(result.current.data?.title).toBe("Retry detail title"),
    );
    expect(result.current.error).toBeNull();
    expect(spy).toHaveBeenCalledTimes(2);
  });
});
